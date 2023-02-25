package fs

import (
    "context"
    "errors"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "hash/fnv"
    "k8s.io/klog/v2"
    "strings"
    "syscall"
)

func keyToFileMode(key string) uint32 {
    isDir := keyIsDir(key)
    if !isDir {
        // signs that obj is a file
        return fuse.S_IFREG
    }
    // signs that obj is a directory
    return fuse.S_IFDIR
}

func keyIsDir(key string) bool {
    return strings.HasSuffix(key, "/")
}

func uniqueInode(s string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(s))
    return h.Sum64()
}

func rootDirAttrs() *fuse.AttrOut {
    return &fuse.AttrOut{
        Attr: fuse.Attr{
            Mode: fuse.S_IFDIR,
        },
    }
}

func setFuseAttr(ctx context.Context, obj *s3.Object, out *fuse.Attr) syscall.Errno {
    key := obj.Key
    objAttrs, err := obj.GetAttrs(ctx)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            klog.Infof("%s: %s", "setFuseAttr", notFoundErr)
            return syscall.ENOENT
        }
        printError("setFuseAttr", err)
        return syscall.EIO
    }
    out.Ino = uniqueInode(key)
    out.Mode = keyToFileMode(key)
    out.Size = objAttrs.Size
    out.SetTimes(objAttrs.AccessTime, objAttrs.ModifyTime, objAttrs.ChangeTime)
    return 0
}

func printError(method string, err error) {
    klog.Errorf("%s: %s", method, err)
}
