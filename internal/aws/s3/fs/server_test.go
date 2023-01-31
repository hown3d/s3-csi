package fs

import (
    "bytes"
    "context"
    "fmt"
    aws_internal "github.com/hown3d/s3-csi/internal/aws"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    internal_dockertest "github.com/hown3d/s3-csi/test/dockertest"
    "github.com/ory/dockertest/v3"
    "github.com/stretchr/testify/assert"
    "io"
    "os"
    "path/filepath"
    "testing"
)

func setupServer(t *testing.T, cfg *Config) (cleanup func()) {

    server, err := NewServer(cfg)
    assert.NoError(t, err)

    err = server.createAndMount()
    assert.NoError(t, err)

    go server.serve()

    err = server.waitMount()
    assert.NoError(t, err)

    return func() {
        umountErr := server.fuseServer.Unmount()
        if umountErr != nil {
            t.Logf("error unmounting fuse server: %s", umountErr)
        }
    }
}

var (
    minioPorts = []string{
        "9000/tcp",
        "9001/tcp",
    }
    minioBucket = "test-bucket"
)

func startMinioContainer(t *testing.T) (cleanup func()) {
    var (
        accessKey       = "minio-access-key"
        secretAccessKey = "minio-secret-access-key"
    )

    opts := &dockertest.RunOptions{
        Repository: "bitnami/minio",
        Tag:        "2023",
        Env: []string{
            fmt.Sprintf("MINIO_ROOT_USER=%s", accessKey),
            fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", secretAccessKey),
            fmt.Sprintf("MINIO_DEFAULT_BUCKETS=%s", minioBucket),
        },
        ExposedPorts: minioPorts,
    }

    // minio container reports healthy on http endpoint right after start, but bitnami restarts the container to create the buckets.
    // that's why we rely on log line health checking
    healthCheckFunc := internal_dockertest.LogHealthcheck("Status:         1 Online, 0 Offline. ")
    con, err := internal_dockertest.NewContainer(opts, healthCheckFunc)
    assert.NoError(t, err)

    port := con.GetPort("9000/tcp")

    err = os.Setenv("AWS_ENDPOINT", fmt.Sprintf("http://localhost:%s", port))
    assert.NoError(t, err)

    err = os.Setenv("AWS_REGION", "eu-central-1")
    assert.NoError(t, err)

    err = os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
    assert.NoError(t, err)

    err = os.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
    assert.NoError(t, err)

    return func() {
        con.Delete()
    }
}

// setupEnvironment setups the environment and fills the created config in cfg
func setupEnvironment(t *testing.T, cfg *Config) {
    conCleanup := startMinioContainer(t)
    t.Cleanup(conCleanup)

    awsCfg, err := aws_internal.NewConfig(context.Background())
    assert.NoError(t, err)

    dir, err := os.MkdirTemp("", t.Name())
    assert.NoError(t, err)
    t.Cleanup(func() {
        os.RemoveAll(dir)
    })

    cfg.MountDir = dir
    cfg.S3Client = s3_internal.NewClient(awsCfg)
    cfg.BucketName = minioBucket

    serverCleanup := setupServer(t, cfg)
    t.Cleanup(serverCleanup)
}

func TestCreate(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    key := "testfile"
    name := filepath.Join(cfg.MountDir, key)
    f, err := os.Create(name)
    if !assert.NoError(t, err) {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        f.Close()
    })

    obj, err := cfg.S3Client.GetObject(context.Background(), minioBucket, key)
    if assert.NoError(t, err) {
        assert.Equal(t, int64(0), obj.ContentLength)
    }
}

func TestWrite(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    key := "testfile"
    name := filepath.Join(cfg.MountDir, key)

    data := []byte("hello-world")
    err := os.WriteFile(name, data, 0755)
    assert.NoError(t, err)

    obj, err := cfg.S3Client.GetObject(context.Background(), minioBucket, key)
    if assert.NoError(t, err) {
        assert.Equal(t, int64(len(data)), obj.ContentLength)
        actualData, err := io.ReadAll(obj.Body)
        if assert.NoError(t, err) {
            assert.Equal(t, data, actualData)
        }
    }
}

func TestRead(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    key := "testfile"
    data := []byte("hello-world")
    err := cfg.S3Client.WriteObject(context.Background(), cfg.BucketName, key, bytes.NewReader(data))
    assert.NoError(t, err)

    filename := filepath.Join(cfg.MountDir, key)
    actualData, err := os.ReadFile(filename)
    if assert.NoError(t, err) {
        assert.Equal(t, data, actualData)
    }
}

func TestMkdir(t *testing.T) {
    // broken atm, because mkdir somehow returns EINVAL on macos but the fuse impl returns 0 (Status OK)
    //t.SkipNow()

    cfg := new(Config)
    setupEnvironment(t, cfg)

    dirName := "testdir"
    fullDirName := filepath.Join(cfg.MountDir, dirName)
    err := os.Mkdir(fullDirName, 0755)
    assert.NoError(t, err)

    obj, err := cfg.S3Client.GetObject(context.Background(), minioBucket, fmt.Sprintf("%s/", dirName))
    if assert.NoError(t, err) {
        assert.Equal(t, int64(0), obj.ContentLength)
    }
}
