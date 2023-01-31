package fs

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    "io"
    "log"
    "os"
    "path/filepath"
    "sync"
    "syscall"
)

type s3File struct {
    *s3Embedded
    mu        sync.Mutex
    cacheFile *os.File
    key       string
}

func newS3FileHandler(embedded *s3Embedded, key string) (*s3File, error) {
    cacheFileName := filepath.Join(os.TempDir(), fmt.Sprintf("%d", uniqueInode(key)), "cache")
    err := os.MkdirAll(filepath.Dir(cacheFileName), 0700)
    if err != nil {
        return nil, fmt.Errorf("error creating cache dir: %w", err)
    }
    f, err := os.OpenFile(cacheFileName, os.O_RDWR|os.O_CREATE, 0600)
    if err != nil {
        return nil, fmt.Errorf("error opening cache file: %w", err)
    }
    return &s3File{
        s3Embedded: embedded,
        cacheFile:  f,
        key:        key,
    }, nil
}

func (s *s3File) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()

    _, err := s.read(ctx, dest, off)
    if err != nil {
        log.Printf("error reading: %s", err)
        return nil, syscall.EINVAL
    }

    return fuse.ReadResultData(dest), 0
}

func (s *s3File) read(ctx context.Context, dest []byte, off int64) (read int, err error) {
    end := int(off) + len(dest)
    byteRange := fmt.Sprintf("bytes=%d-%d", off, end)
    out, err := s.s3Client.GetObject(ctx, s.bucketName, s.key, func(input *s3.GetObjectInput) {
        input.Range = &byteRange
    })
    if err != nil {
        return 0, fmt.Errorf("error getting object from s3: %s", err)
    }
    r := out.Body

    n, err := io.ReadFull(r, dest)
    if err != nil {
        return 0, fmt.Errorf("error reading from s3: %s", err)
    }
    return n, nil
}

func (s *s3File) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()

    n, err := s.cacheFile.WriteAt(data, off)
    if err != nil {
        log.Printf("error writing new data to cache file: %s", err)
        return 0, syscall.EIO
    }
    _, err = s.cacheFile.Seek(0, io.SeekStart)
    if err != nil {
        log.Printf("error reseting cache file offset: %s", err)
        return 0, syscall.EIO
    }

    if err := s.s3Write(ctx, s.key, s.cacheFile); err != nil {
        return 0, syscall.EIO
    }

    return uint32(n), 0
}

func (s *s3File) Flush(ctx context.Context) syscall.Errno {
    err := os.Remove(s.cacheFile.Name())
    if err != nil {
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
