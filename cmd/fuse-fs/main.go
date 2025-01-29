package main

import (
	"flag"

	"github.com/hown3d/s3-csi/internal/aws/s3"
	"github.com/hown3d/s3-csi/internal/aws/s3/fs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog/v2"
)

var (
	mntDir     = flag.String("mount-dir", "/tmp/s3-fuse-mnt", "directory at which the filesystem will be mounted")
	bucketName = flag.String("s3-bucket", "", "directory at which the filesystem will be mounted")
	endpoint   = flag.String("s3-endpoint", "", "s3 endpoint to connect to")
	debug      = flag.Bool("debug", false, "enable debug logging of fuse server")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	if *bucketName == "" {
		klog.Fatalf("s3-bucket flag must be set")
	}

	if *endpoint == "" {
		klog.Fatalf("s3-endpoint flag must be set")
	}
	s3Client, err := s3.NewClient(*endpoint, &minio.Options{
		Creds: credentials.NewEnvAWS(),
	})
	if err != nil {
		klog.Fatalf("creating s3 client: %v", err)
	}

	cfg := &fs.Config{
		MountDir:   *mntDir,
		S3Client:   s3Client,
		BucketName: *bucketName,
		Debug:      *debug,
	}
	server, err := fs.NewServer(cfg)
	if err != nil {
		klog.Fatalf("error creating fs server: %s", err)
	}
	// start serving the file system
	if err := server.Run(); err != nil {
		klog.Fatalf("error running fs server: %s", err)
	}
}
