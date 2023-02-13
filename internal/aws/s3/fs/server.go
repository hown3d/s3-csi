package fs

import (
    "context"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "k8s.io/klog/v2"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type Config struct {
    MountDir   string
    S3Client   *s3.Client
    BucketName string
    Debug      bool
}

type Server struct {
    fuseServer *fuse.Server
    killSigs   chan<- os.Signal
}

const ROOT_KEY = "root/"

func NewServer(cfg *Config) (*Server, error) {
    one := time.Second
    logger := log.Default()
    logger.SetPrefix("s3-fuse: ")

    options := &fs.Options{
        EntryTimeout: &one,
        AttrTimeout:  &one,
        MountOptions: fuse.MountOptions{
            // Set to true to see how the file system works.
            Debug:       cfg.Debug,
            Name:        "s3-fs",
            FsName:      "s3-fs",
            DirectMount: true,
        },
        Logger: logger,
    }
    bucket := cfg.S3Client.NewBucket(cfg.BucketName)
    if !bucket.Exists(context.Background()) {
        return nil, fmt.Errorf("bucket %s does not exist", cfg.BucketName)
    }

    rootObj := cfg.S3Client.NewObject(cfg.BucketName, ROOT_KEY)
    err := rootObj.Create(context.Background(), nil)
    if err != nil {
        return nil, fmt.Errorf("creating root object: %w", err)
    }

    root := &s3Node{
        bucketName: bucket.Name,
        client:     cfg.S3Client,
    }
    rawFS := fs.NewNodeFS(root, options)

    err = os.MkdirAll(cfg.MountDir, 0700)
    if !os.IsExist(err) && err != nil {
        return nil, fmt.Errorf("mountdir could not be created: %w", err)
    }
    fuseServer, err := fuse.NewServer(rawFS, cfg.MountDir, &options.MountOptions)
    if err != nil {
        return nil, err
    }

    killSigs := make(chan os.Signal)
    signal.Notify(killSigs, syscall.SIGTERM, syscall.SIGKILL)
    go func() {
        // Block until a signal is received.
        s := <-killSigs
        klog.Infof("Got signal: %s, unmounting filesystem", s)
        err := fuseServer.Unmount()
        if err != nil {
            klog.Error("error unmounting fs: %s", err)
        }
    }()
    return &Server{
        fuseServer: fuseServer,
    }, nil
}

func (s *Server) start() error {
    s.fuseServer.Serve()
    return nil
}

func (s *Server) waitMount() error {
    if err := s.fuseServer.WaitMount(); err != nil {
        return err
    }
    return nil
}

func (s *Server) wait() {
    s.fuseServer.Wait()
}

func (s *Server) Run() error {
    if err := s.start(); err != nil {
        return fmt.Errorf("starting fuse server: %w", err)
    }
    if err := s.waitMount(); err != nil {
        return fmt.Errorf("waiting for mount of fuse server: %w", err)
    }
    s.wait()
    return nil
}
