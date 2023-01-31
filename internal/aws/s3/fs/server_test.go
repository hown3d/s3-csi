package fs

import (
    "bytes"
    "context"
    "fmt"
    aws_internal "github.com/hown3d/s3-csi/internal/aws"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "github.com/hown3d/s3-csi/test/aws/localstack"
    "github.com/stretchr/testify/assert"
    "io"
    "os"
    "path/filepath"
    "testing"
)

func setupServer(t *testing.T, cfg *Config) {

    server, err := NewServer(cfg)
    assert.NoError(t, err)

    err = server.createAndMount()
    assert.NoError(t, err)

    go server.serve()

    err = server.waitMount()
    assert.NoError(t, err)

    t.Cleanup(func() {
        umountErr := server.fuseServer.Unmount()
        if umountErr != nil {
            t.Logf("error unmounting fuse server: %s", umountErr)
        }
    })
}

func startLocalstack(t *testing.T) {
    con, err := localstack.New()
    if err != nil {
        t.Fatal(err)
    }
    if err := con.SetAWSEnvVariables(); err != nil {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        con.Delete()
    })
}

// setupEnvironment setups the environment and fills the created config in cfg
func setupEnvironment(t *testing.T, cfg *Config) *s3_internal.Bucket {
    startLocalstack(t)

    awsCfg, err := aws_internal.NewConfig(context.Background())
    assert.NoError(t, err)

    dir, err := os.MkdirTemp("", t.Name())
    assert.NoError(t, err)
    t.Cleanup(func() {
        os.RemoveAll(dir)
    })

    bucketName := "testbucket"

    cfg.MountDir = dir
    cfg.S3Client = s3_internal.NewClient(awsCfg)
    cfg.BucketName = bucketName

    bucket, err := cfg.S3Client.CreateBucket(context.Background(), bucketName, "")
    if !assert.NoError(t, err) {
        t.Fatal(err)
    }
    setupServer(t, cfg)
    return bucket
}

func TestCreate(t *testing.T) {
    cfg := new(Config)
    bucket := setupEnvironment(t, cfg)

    key := "testfile"
    name := filepath.Join(cfg.MountDir, key)
    f, err := os.Create(name)
    if !assert.NoError(t, err) {
        t.Fatal(err)
    }

    t.Cleanup(func() {
        f.Close()
    })

    ctx := context.Background()
    obj, err := bucket.GetObject(ctx, key)
    assert.NoError(t, err)

    r, err := obj.Read(ctx, nil)
    assert.NoError(t, err)

    data, err := io.ReadAll(r)
    assert.NoError(t, err)

    assert.Equal(t, 0, len(data))
}

func TestWrite(t *testing.T) {
    cfg := new(Config)
    bucket := setupEnvironment(t, cfg)

    key := "testfile"
    name := filepath.Join(cfg.MountDir, key)

    data := []byte("hello-world")
    err := os.WriteFile(name, data, 0755)
    assert.NoError(t, err)

    ctx := context.Background()
    obj, err := bucket.GetObject(ctx, key)
    assert.NoError(t, err)

    r, err := obj.Read(ctx, nil)
    assert.NoError(t, err)

    actualData, err := io.ReadAll(r)
    assert.NoError(t, err)

    assert.Equal(t, len(data), len(actualData))
    assert.Equal(t, data, actualData)
}

func TestRead(t *testing.T) {
    cfg := new(Config)
    cfg.Debug = true
    bucket := setupEnvironment(t, cfg)

    ctx := context.Background()
    key := "testfile"
    data := []byte("hello-world")

    obj, err := bucket.CreateObject(ctx, key, nil)
    assert.NoError(t, err)

    err = obj.Write(ctx, bytes.NewReader(data), nil)
    assert.NoError(t, err)

    filename := filepath.Join(cfg.MountDir, key)
    actualData, err := os.ReadFile(filename)
    if assert.NoError(t, err) {
        assert.Equal(t, data, actualData)
    }
}

func TestMkdir(t *testing.T) {
    // broken atm, because mkdir somehow returns EINVAL on macos but the fuse impl returns 0 (Status OK)
    t.SkipNow()

    cfg := new(Config)
    bucket := setupEnvironment(t, cfg)

    dirName := "testdir"
    fullDirName := filepath.Join(cfg.MountDir, dirName)
    err := os.Mkdir(fullDirName, 0755)
    assert.NoError(t, err)

    ctx := context.Background()
    obj, err := bucket.GetObject(ctx, fmt.Sprintf("%s/", dirName))
    assert.NoError(t, err)

    r, err := obj.Read(ctx, nil)
    assert.NoError(t, err)

    data, err := io.ReadAll(r)
    assert.NoError(t, err)

    assert.Equal(t, 0, len(data))
}
