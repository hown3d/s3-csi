package fs

import (
    "context"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "hash/fnv"
    "log"
    "strings"
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

func setFuseAttr(ctx context.Context, obj *s3.Object, out *fuse.Attr) error {
    key := obj.Key
    objAttrs, err := obj.GetAttrs(ctx)
    if err != nil {
        return err
    }
    out.Ino = uniqueInode(key)
    out.Mode = keyToFileMode(key)
    out.Size = objAttrs.Size
    out.SetTimes(objAttrs.AccessTime, objAttrs.ModifyTime, objAttrs.ChangeTime)
    return nil
}

func printError(method string, err error) {
    logger := log.Default()
    logger.SetPrefix(fmt.Sprintf("%s: ", method))
    logger.Println(err)
}
