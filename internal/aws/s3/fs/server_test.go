package fs

import (
    "bytes"
    "context"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/hown3d/s3-csi/internal/aws"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "github.com/hown3d/s3-csi/test/aws/localstack"
    "github.com/stretchr/testify/assert"
    "io"
    "log"
    "os"
    "path/filepath"
    "testing"
)

func setupServer(t *testing.T, cfg *Config) {
    server, err := NewServer(cfg)
    assert.NoError(t, err)

    go server.start()

    err = server.waitMount()
    assert.NoError(t, err)

    t.Cleanup(func() {
        umountErr := server.fuseServer.Unmount()
        if umountErr != nil {
            t.Logf("error unmounting fuse server: %s", umountErr)
        }
    })
}

func TestMain(m *testing.M) {
    cleanup, err := startLocalstack()
    if err != nil {
        cleanup()
        log.Print(err)
        return
    }
    exitCode := m.Run()
    cleanup()
    os.Exit(exitCode)
}

func startLocalstack() (cleanup func(), err error) {
    con, err := localstack.New()
    if err != nil {
        return func() {}, err
    }
    cleanup = func() {
        con.Delete()
    }
    if err := con.SetAWSEnvVariables(); err != nil {
        return cleanup, err
    }
    return cleanup, err
}

// setupEnvironment setups the environment and fills the created config in cfg
func setupEnvironment(t *testing.T, cfg *Config) {
    awsCfg, err := aws.NewConfig(context.Background())
    assert.NoError(t, err)

    dir, err := os.MkdirTemp("", t.Name())
    assert.NoError(t, err)
    t.Cleanup(func() {
        os.RemoveAll(dir)
    })

    bucketName := "testbucket"

    cfg.MountDir = dir
    cfg.S3Client = s3.NewClient(awsCfg)
    cfg.BucketName = bucketName

    bucket := cfg.S3Client.NewBucket(bucketName)
    err = bucket.Create(context.Background(), "")
    if err != nil {
        var alreadyExists *types.BucketAlreadyExists
        if !assert.ErrorAs(t, err, &alreadyExists) {
            t.Fatal(err)
        }
    }
    t.Cleanup(func() {
        bucket.Empty(context.Background())
    })
    setupServer(t, cfg)
}

func TestCreate(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    filename := "testfile"
    name := filepath.Join(cfg.MountDir, filename)
    f, err := os.Create(name)
    if !assert.NoError(t, err) {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        f.Close()
    })

    ctx := context.Background()
    obj := cfg.S3Client.NewObject(cfg.BucketName, objKey(filename))
    assert.True(t, obj.Exists(ctx))

    r, err := obj.Read(ctx, nil)
    assert.NoError(t, err)

    data, err := io.ReadAll(r)
    assert.NoError(t, err)

    assert.Equal(t, 0, len(data))
}

func TestWrite(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    filename := "testfile"
    name := filepath.Join(cfg.MountDir, filename)

    data := []byte("hello-world")
    err := os.WriteFile(name, data, 0755)
    assert.NoError(t, err)

    ctx := context.Background()
    obj := cfg.S3Client.NewObject(cfg.BucketName, objKey(filename))
    assert.True(t, obj.Exists(ctx))

    r, err := obj.Read(ctx, nil)
    assert.NoError(t, err)

    actualData, err := io.ReadAll(r)
    assert.NoError(t, err)

    assert.Equal(t, len(data), len(actualData))
    assert.Equal(t, data, actualData)
}

func TestRead(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    ctx := context.Background()
    key := objKey("testfile")
    data := []byte("hello-world")

    obj := cfg.S3Client.NewObject(cfg.BucketName, key)
    err := obj.Create(ctx, nil)
    assert.NoError(t, err)

    err = obj.Write(ctx, bytes.NewReader(data), nil)
    assert.NoError(t, err)

    filename := filepath.Join(cfg.MountDir, key)
    actualData, err := os.ReadFile(filename)
    if assert.NoError(t, err) {
        assert.Equal(t, data, actualData)
    }
}

func TestReadDir(t *testing.T) {
    cfg := new(Config)
    setupEnvironment(t, cfg)

    folderName := objKey("testFolder")
    dirName := filepath.Join(cfg.MountDir, folderName)
    files := []string{
        filepath.Join(dirName, "test1"),
        filepath.Join(dirName, "test2"),
    }
    for _, f := range files {
        fd, err := os.Create(f)
        assert.NoError(t, err)
        fd.Close()
    }

    entries, err := os.ReadDir(dirName)
    assert.NoError(t, err)
    for _, entry := range entries {
        assert.Contains(t, files, entry.Name())
    }
}

func objKey(filename string) string {
    return "/" + filename
}
