//go:build linux

package fs

import (
	"context"
	"github.com/hown3d/s3-csi/internal/aws/s3"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestMkdir(t *testing.T) {
	cfg := new(Config)
	setupEnvironment(t, cfg)

	dirName := "testdir"
	fullDirName := filepath.Join(cfg.MountDir, dirName)
	err := os.Mkdir(fullDirName, 0755)
	failIfErr(t, err)

	ctx := context.Background()
	obj := cfg.S3Client.NewObject(cfg.BucketName, dirName+"/")
	assert.True(t, obj.Exists(ctx))

	r, err := obj.Read(ctx, nil)
	failIfErr(t, err)

	data, err := io.ReadAll(r)
	failIfErr(t, err)

	assert.Equal(t, 0, len(data))
}

func TestSetAttrs(t *testing.T) {
	now := time.Date(2000, 1, 1, 1, 1, 1, 1, time.Local)
	originalTimeNow := s3.TimeNowFunc
	s3.TimeNowFunc = func() time.Time {
		return now
	}
	defer func() {
		s3.TimeNowFunc = originalTimeNow
	}()

	cfg := new(Config)
	setupEnvironment(t, cfg)

	key := "testfile"
	name := filepath.Join(cfg.MountDir, key)
	f, err := os.Create(name)
	defer f.Close()
	assert.NoError(t, err)
	fileInfo, err := f.Stat()
	assert.NoError(t, err)

	sysFileInfo := fileInfo.Sys()
	stat, ok := sysFileInfo.(*syscall.Stat_t)
	if !ok {
		t.Fatalf("%T is not syscall.Stat_t", sysFileInfo)
	}
	assert.Equal(t, now, time.Unix(stat.Ctim.Unix()))
}
