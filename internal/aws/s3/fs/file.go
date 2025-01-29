package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hown3d/s3-csi/internal/aws/s3"
	"k8s.io/klog/v2"
)

type s3File struct {
	mu        sync.Mutex
	cacheFile *os.File
	obj       *s3.Object
}

func newS3FileHandler(obj *s3.Object) (*s3File, error) {
	fileNameFullPath := filepath.Join(os.TempDir(), obj.Bucket.String(), obj.String(), "cache")
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
	return s.obj.ReadAt(dest, off)
}

func (s *s3File) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	klog.Infof("writing object with key: %s at offset %d", s.obj, off)
	s.mu.Lock()
	defer s.mu.Unlock()

	n, err := s.cacheFile.WriteAt(data, off)
	if err != nil {
		printError("Write", fmt.Errorf("error writing new data to cache file: %w", err))
		return 0, syscall.EIO
	}
	_, err = s.cacheFile.Seek(0, io.SeekStart)
	if err != nil {
		printError("Write", fmt.Errorf("error reseting cache file offset: %w", err))
		return 0, syscall.EIO
	}

	return uint32(n), 0
}

func (s *s3File) Flush(ctx context.Context) syscall.Errno {
	klog.Infof("removing cachefile for object %s", s.obj)
	_, err := io.Copy(s.obj, s.cacheFile)
	if err != nil {
		printError("Write", fmt.Errorf("writting to : %w", err))
		return syscall.EIO
	}
	s.cacheFile.Close()
	err = os.Remove(s.cacheFile.Name())
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
