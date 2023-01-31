package main

import (
    "context"
    "errors"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "log"
    "path/filepath"
    "strings"
    "sync"
    "syscall"
)

type S3Node struct {
    fs.Inode

    // When file systems are mutable, all access must use
    // synchronization.
    mu sync.RWMutex
}

func (s *S3Node) Rmdir(ctx context.Context, name string) syscall.Errno {
    folderKey := s.key(name) + "/"
    objs, err := s3Client.ListObjects(ctx, bucketName, s3_internal.ListObjectsWithPrefix(folderKey))
    if err != nil {
        return syscall.EINVAL
    }
    if len(objs) == 0 {
        return syscall.ENOENT
    }

    objKeys := make([]string, len(objs))
    for i, obj := range objs {
        objKeys[i] = *obj.Key
    }
    if err := s3Client.DeleteObjects(ctx, bucketName, objKeys); err != nil {
        return syscall.EINVAL
    }
    return 0
}

func (s *S3Node) newNode(ctx context.Context, name string) *fs.Inode {
    node := &S3Node{}
    stable := fs.StableAttr{
        Mode: objNameToFileMode(name),
        Ino:  uniqueInode(name),
    }

    newInode := s.NewInode(ctx, node, stable)
    return newInode
}

func (s *S3Node) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    attrs := &s3_internal.ObjectAttrs{}
    cTime, set := in.GetCTime()
    if set {
        attrs.ChangeTime = &cTime
    }
    key := s.key("")
    if err := s3Client.SetObjectAttrs(ctx, bucketName, key, attrs); err != nil {
        log.Printf("setting object attr: %s", err)
        return syscall.EINVAL
    }
    return 0
}

func (s *S3Node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    key := s.key("")
    s3File, err := NewS3FileHandler(key)
    if err != nil {
        log.Printf("error creating s3 file handler: %s", err)
        return nil, 0, syscall.EIO
    }
    return s3File, 0, 0
}

func (s *S3Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    key := s.key("")
    objs, err := s3Client.ListObjects(ctx, bucketName,
        s3_internal.ListObjectsWithPrefix(key),
    )
    if err != nil {
        log.Printf("error: %#v", err)
        // TODO: handle errors correctly
        return nil, syscall.ENOENT
    }

    dirs := make([]fuse.DirEntry, 0, len(objs))
    for _, obj := range objs {
        numSeperators := strings.Count(*obj.Key, string(filepath.Separator))
        // more than 1 seperator is not in our dir
        if numSeperators > 1 {
            continue
        }
        if numSeperators == 1 {
            // seperators is 1, but not suffixed with folder seperator, so skip
            if !strings.HasSuffix(*obj.Key, string(filepath.Separator)) {
                continue
            }
        }
        d := fuse.DirEntry{
            Name: *obj.Key,
            Ino:  uniqueInode(*obj.Key),
            Mode: objNameToFileMode(*obj.Key),
        }
        dirs = append(dirs, d)
    }
    return fs.NewListDirStream(dirs), 0
}

func (s *S3Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    key := s.key(name)
    _, err := s3Client.GetObject(ctx, bucketName, key)
    if err != nil {
        var notFoundErr *s3_internal.ErrNotFound
        if !errors.As(err, &notFoundErr) {
            log.Println(err)
            return nil, syscall.EIO
        }
        // since fuse removes the slashes on the name, retry lookup with slash suffixed to make sure we find folders
        key = key + "/"
        _, err = s3Client.GetObject(ctx, bucketName, key)
        // clear error of previous invocation
        notFoundErr = nil
        if errors.As(err, &notFoundErr) {
            return nil, syscall.ENOENT
        }
    }

    stable := fs.StableAttr{
        Mode: objNameToFileMode(key),
        Ino:  uniqueInode(key),
    }

    setAttrs(ctx, key, &out.Attr)
    node := &S3Node{}
    newInode := s.NewInode(ctx, node, stable)
    if success := s.AddChild(key, newInode, true); !success {
        log.Printf("adding new inode as child to %#v was not successfull", s)
    }
    return newInode, 0
}

func (s *S3Node) Access(ctx context.Context, mask uint32) syscall.Errno {
    return 0
}

func (s *S3Node) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // root node has key set to ""
    if s.EmbeddedInode().IsRoot() {
        out = rootDirAttrs()
        return 0
    }

    key := s.key("")
    out.Attr.Mode = objNameToFileMode(key)
    out.Ino = uniqueInode(key)

    setAttrs(ctx, key, &out.Attr)
    return 0
}

func setAttrs(ctx context.Context, objKey string, out *fuse.Attr) {
    attrs := getS3Attrs(ctx, objKey)
    if attrs != nil {
        out.Size = attrs.Size
        out.SetTimes(attrs.AccessTime, attrs.ModifyTime, attrs.ChangeTime)
    }
}

// path returns the full path to the file in the underlying file
// system.
func (s *S3Node) key(childName string) string {
    nodePath := s.Path(s.Root())
    return filepath.Join(nodePath, childName)
}

func (s *S3Node) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    childKey := s.key(name)
    err := createEmptyObject(ctx, childKey)
    if err != nil {
        log.Printf("error creating object: %s", err)
        return nil, nil, 0, syscall.EIO
    }
    node = s.newNode(ctx, name)
    s3File, err := NewS3FileHandler(childKey)
    if err != nil {
        log.Printf("error s3 file handler: %s", err)
        return nil, nil, 0, syscall.EIO
    }
    return node, s3File, 0, 0
}

func (s *S3Node) Unlink(ctx context.Context, name string) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    childKey := s.key(name)
    if err := s3Client.DeleteObject(ctx, bucketName, childKey); err != nil {
        return syscall.EIO
    }
    return 0
}

func (s *S3Node) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    oldChildKey := s.key(name)
    newChildKey := s.key(newName)
    if err := s3Client.CopyObject(ctx, bucketName, oldChildKey, newChildKey); err != nil {
        return syscall.EINVAL
    }
    if err := s3Client.DeleteObject(ctx, bucketName, oldChildKey); err != nil {
        return syscall.EIO
    }
    success := s.MvChild(name, newParent.EmbeddedInode(), newName, true)
    if !success {
        log.Println("moving child was not successful")
    }
    return 0
}

func (s *S3Node) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // fuse removes the trailing slash
    // must be readded for server to recognize name as folder
    childKey := s.key(name) + "/"

    if err := createEmptyObject(ctx, childKey); err != nil {
        log.Printf("error creating empty object: %s", err)
        return nil, syscall.EIO
    }

    return s.newNode(ctx, name), 0
}

var (
    // Node types must be InodeEmbedders
    _ fs.InodeEmbedder = (*S3Node)(nil)
    _ fs.NodeLookuper  = (*S3Node)(nil)
    _ fs.NodeReaddirer = (*S3Node)(nil)

    _ fs.NodeSetattrer = (*S3Node)(nil)
    _ fs.NodeGetattrer = (*S3Node)(nil)

    _ fs.NodeAccesser = (*S3Node)(nil)

    _ fs.NodeMkdirer = (*S3Node)(nil)

    _ fs.NodeUnlinker = (*S3Node)(nil)
    _ fs.NodeRenamer  = (*S3Node)(nil)

    _ fs.NodeCreater = (*S3Node)(nil)

    // Implement (handleless) Open
    _ fs.NodeOpener = (*S3Node)(nil)

    _ fs.NodeRmdirer = (*S3Node)(nil)
)
