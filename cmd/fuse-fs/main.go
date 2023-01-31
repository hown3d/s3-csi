package main

import (
    "context"
    "flag"
    "github.com/hown3d/s3-csi/internal/aws"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "github.com/hown3d/s3-csi/internal/aws/s3/fs"
    "log"
)

var (
    mntDir     = flag.String("mount-dir", "/tmp/s3-fuse-mnt", "directory at which the filesystem will be mounted")
    bucketName = flag.String("s3-bucket", "", "directory at which the filesystem will be mounted")
    debug      = flag.Bool("debug", false, "enable debug logging of fuse server")
)

func main() {
    flag.Parse()
    if *bucketName == "" {
        log.Fatalf("s3-bucket flag must be set")
    }

    awsCfg, err := aws.NewConfig(context.Background())
    if err != nil {
        log.Fatalf("error creating aws config: %s", err)
    }
    cfg := &fs.Config{
        MountDir:   *mntDir,
        S3Client:   s3.NewClient(awsCfg),
        BucketName: *bucketName,
        Debug:      *debug,
    }
    server, err := fs.NewServer(cfg)
    if err != nil {
        log.Fatalf("error creating fs server: %s", err)
    }
    // start serving the file system
    if err := server.Run(); err != nil {
        log.Fatalf("error running fs server: %s", err)
    }
}
