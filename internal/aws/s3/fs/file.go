package fs

import (
    "context"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "io"
    "k8s.io/klog/v2"
    "os"
    "path/filepath"
    "strconv"
    "sync"
    "syscall"
)

type s3File struct {
    mu        sync.Mutex
    cacheFile *os.File
    obj       *s3.Object
}

func newS3FileHandler(obj *s3.Object) (*s3File, error) {
    fileNameFullPath := filepath.Join(os.TempDir(), strconv.FormatUint(uniqueInode(obj.Key), 10), "cache")
    err := os.MkdirAll(filepath.Dir(fileNameFullPath), 0700)
    if err != nil {
        return nil, fmt.Errorf("error creating cache dir: %w", err)
    }
    f, err := os.OpenFile(fileNameFullPath, os.O_RDWR|os.O_CREATE, 0600)
    if err != nil {
        return nil, fmt.Errorf("error opening cache file: %w", err)
    }
    s3f := &s3File{
        cacheFile: f,
        obj:       obj,
    }
    return s3f, nil
}

func (s *s3File) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()

    _, err := s.read(ctx, dest, off)
    if err != nil {
        printError("Read", err)
        return nil, syscall.EINVAL
    }

    return fuse.ReadResultData(dest), 0
}

func (s *s3File) read(ctx context.Context, dest []byte, off int64) (read int, err error) {
    end := int(off) + len(dest)
    readRange := &s3.ReadRange{
        Start: off,
        End:   int64(end),
    }
    klog.V(5).Infof("reading from object %s from start %d until end %d", s.obj.Key, off, end)
    r, err := s.obj.Read(ctx, readRange)
    if err != nil {
        return 0, fmt.Errorf("error getting object from s3: %s", err)
    }

    for {
        n, err := r.Read(dest)
        if err != nil {
            if err == io.EOF {
                break
            }
            return 0, fmt.Errorf("error reading from s3: %s", err)
        }
        read += n
    }
    return read, nil
}

func (s *s3File) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
    klog.V(5).Infof("writing object with key: %s at offset %d", s.obj.Key, off)
    s.mu.Lock()
    defer s.mu.Unlock()

    n, err := s.cacheFile.WriteAt(data, off)
    if err != nil {
        printError("Flush", fmt.Errorf("error writing new data to cache file: %w", err))
        return 0, syscall.EIO
    }
    _, err = s.cacheFile.Seek(0, io.SeekStart)
    if err != nil {
        printError("Flush", fmt.Errorf("error reseting cache file offset: %w", err))
        return 0, syscall.EIO
    }

    if err := s.obj.Write(ctx, s.cacheFile, nil); err != nil {
        return 0, syscall.EIO
    }

    return uint32(n), 0
}

func (s *s3File) Flush(ctx context.Context) syscall.Errno {
    klog.V(5).Infof("removing cachefile for object %s", s.obj.Key)
    err := os.Remove(s.cacheFile.Name())
    // flush could be called twice, so if the cache file is already deleted, return ok
    if err != nil && !os.IsNotExist(err) {
        printError("Flush", fmt.Errorf("error deleting cacheFile: %w", err))
        return syscall.EINVAL
    }
    return 0
}

// Interface compliance
var (
    _ fs.FileFlusher = (*s3File)(nil)
    _ fs.FileWriter  = (*s3File)(nil)

    _ fs.FileReader = (*s3File)(nil)
)
