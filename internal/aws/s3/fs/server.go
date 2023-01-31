package fs

import (
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "log"
    "time"
)

type Config struct {
    MountDir   string
    S3Client   *s3_internal.Client
    BucketName string
}

type Server struct {
    fs         fuse.RawFileSystem
    mountDir   string
    fuseServer *fuse.Server
    opts       *fs.Options
}

func NewServer(cfg *Config) (*Server, error) {
    one := time.Second
    logger := log.Default()
    logger.SetPrefix("s3-fuse")

    options := &fs.Options{
        EntryTimeout: &one,
        AttrTimeout:  &one,
        MountOptions: fuse.MountOptions{
            // Set to true to see how the file system works.
            Debug:  true,
            Name:   "s3-fs",
            FsName: "s3-fs",
        },
        Logger: logger,
    }

    root := &s3Node{
        s3Embedded: &s3Embedded{
            s3Client:   cfg.S3Client,
            bucketName: cfg.BucketName,
        },
    }
    rawFS := fs.NewNodeFS(root, options)

    return &Server{
        fs:       rawFS,
        mountDir: cfg.MountDir,
        opts:     options,
    }, nil
}

func (s *Server) createAndMount() error {
    server, err := fuse.NewServer(s.fs, s.mountDir, &s.opts.MountOptions)
    if err != nil {
        return err
    }
    s.fuseServer = server
    return nil
}

func (s *Server) serve() {
    s.fuseServer.Serve()
}

func (s *Server) start() error {
    if err := s.createAndMount(); err != nil {
        return fmt.Errorf("creating fuse server and mounting to %s: %w", s.mountDir, err)
    }
    s.serve()
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
