package main

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    aws_internal "github.com/hown3d/s3-csi/internal/aws"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "hash/fnv"
    "log"
    "strings"
    "time"
)

var (
    s3Client   *s3_internal.Client
    bucketName = "s3-fuse-test"

    root = &S3Node{}
)

func main() {
    awsCfg, err := aws_internal.NewConfig(context.Background())
    if err != nil {
        log.Fatalf("error creating aws config: %s", err)
    }
    s3Client = s3_internal.NewClient(awsCfg)

    mntDir := "/tmp/s3-fuse-mnt"
    oneSec := time.Second
    server, err := fs.Mount(mntDir, root, &fs.Options{

        EntryTimeout: &oneSec,
        AttrTimeout:  &oneSec,
        MountOptions: fuse.MountOptions{
            // Set to true to see how the file system works.
            Debug: true,
        },
        Logger: log.Default(),
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Unmount by calling 'umount %s'", mntDir)

    // start serving the file system
    server.Wait()
}

func objNameToFileMode(objName string) uint32 {
    leaf := isLeafObject(objName)
    if leaf {
        // signs that obj is a file
        return fuse.S_IFREG
    }
    // signs that obj is a directory
    return fuse.S_IFDIR
}

func isLeafObject(objName string) bool {
    return !strings.HasSuffix(objName, "/")
}

func uniqueInode(s string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(s))
    return h.Sum64()
}

func createEmptyObject(ctx context.Context, name string) error {
    if err := s3Client.WriteObject(ctx, bucketName, name, new(bytes.Buffer)); err != nil {
        return fmt.Errorf("error creating folder in s3: %s", err)
    }
    return nil
}

func getS3Attrs(ctx context.Context, key string) *s3_internal.ObjectAttrs {
    attrs, err := s3Client.GetObjectAttrs(ctx, bucketName, key)
    if err != nil {
        var notFound *s3_internal.ErrNotFound
        if !errors.As(err, &notFound) {
            log.Printf("error retrieving metadata for node: %s", err)
        }
        return nil
    }
    return attrs
}

func rootDirAttrs() *fuse.AttrOut {
    return &fuse.AttrOut{
        Attr: fuse.Attr{
            Mode: fuse.S_IFDIR,
        },
    }
}
