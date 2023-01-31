package fs

import (
    "context"
    "errors"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "io"
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

func (s *s3File) Setattr(ctx context.Context, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    //root node has key set to "", so it doesnt map to a key in s3
    //if s.EmbeddedInode().IsRoot() {
    //    out = rootDirAttrs()
    //    return 0
    //}

    attrs := &s3.ObjectAttrs{}
    modifyTime, valid := in.SetAttrInCommon.GetMTime()
    if valid {
        attrs.ModifyTime = &modifyTime
    }

    accessTime, valid := in.SetAttrInCommon.GetATime()
    if valid {
        attrs.AccessTime = &accessTime
    }

    changeTime, valid := in.SetAttrInCommon.GetCTime()
    if valid {
        attrs.ChangeTime = &changeTime
    }

    if err := s.obj.SetAttrs(ctx, attrs); err != nil {
        printError("Setattr", err)
        return syscall.EIO
    }
    return 0

}

func (s *s3File) Getattr(ctx context.Context, out *fuse.AttrOut) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    //if s.EmbeddedInode().IsRoot() {
    //    out = rootDirAttrs()
    //    return 0
    //}

    err := setFuseAttr(ctx, s.obj, &out.Attr)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return syscall.ENOENT
        }
        printError("Gettattr", err)
        return syscall.EIO
    }
    return 0
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
    r, err := s.obj.Read(ctx, readRange)
    if err != nil {
        return 0, fmt.Errorf("error getting object from s3: %s", err)
    }

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
    err := os.Remove(s.cacheFile.Name())
    if err != nil {
        printError("Flush", fmt.Errorf("error deleting cacheFile: %w", err))
        return syscall.EINVAL
    }
    return 0
}

// Interface compliance
var (
    _ fs.FileSetattrer = (*s3File)(nil)
    _ fs.FileGetattrer = (*s3File)(nil)
    _ fs.FileFlusher   = (*s3File)(nil)
    _ fs.FileWriter    = (*s3File)(nil)

    _ fs.FileReader = (*s3File)(nil)
)
