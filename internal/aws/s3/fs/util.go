package fs

import (
    "github.com/hanwen/go-fuse/v2/fuse"
    "hash/fnv"
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
