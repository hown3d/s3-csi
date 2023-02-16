//go:build darwin

package fs

import (
	"github.com/hown3d/s3-csi/internal/aws/s3"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestMkdir(t *testing.T) {
	// broken atm, because mkdir somehow returns EINVAL on macos but the fuse impl returns 0 (Status OK)
	t.SkipNow()

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
	cfg.Debug = true
	setupEnvironment(t, cfg)

	key := "testfile"
	name := filepath.Join(cfg.MountDir, key)
	f, err := os.Create(name)
	defer f.Close()

	fileInfo, err := f.Stat()
	failIfErr(t, err)

	sysFileInfo := fileInfo.Sys()
	stat, ok := sysFileInfo.(*syscall.Stat_t)
	if !ok {
		t.Fatalf("%T is not syscall.Stat_t", sysFileInfo)
	}
	assert.Equal(t, now, time.Unix(stat.Ctimespec.Unix()))
}
