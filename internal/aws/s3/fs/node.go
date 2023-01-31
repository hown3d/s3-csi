package fs

import (
    "context"
    "errors"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "log"
    "path/filepath"
    "strings"
    "sync"
    "syscall"
)

type s3Node struct {
    fs.Inode
    *s3Embedded

    // When file systems are mutable, all access must use
    // synchronization.
    mu sync.RWMutex
}

func (s *s3Node) Rmdir(ctx context.Context, name string) syscall.Errno {
    folderKey := s.key(name) + "/"
    objs, err := s.s3Client.ListObjects(ctx, s.bucketName, s3_internal.ListObjectsWithPrefix(folderKey))
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
    if err := s.s3Client.DeleteObjects(ctx, s.bucketName, objKeys); err != nil {
        return syscall.EINVAL
    }
    return 0
}

func (s *s3Node) newInode(ctx context.Context, key string) *fs.Inode {
    node := &s3Node{
        s3Embedded: s.s3Embedded,
    }
    stable := fs.StableAttr{
        Mode: keyToFileMode(key),
        Ino:  uniqueInode(key),
    }

    newInode := s.NewInode(ctx, node, stable)
    return newInode
}

func (s *s3Node) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    attrs := &s3_internal.ObjectAttrs{}
    cTime, set := in.GetCTime()
    if set {
        attrs.ChangeTime = &cTime
    }
    key := s.key("")
    if err := s.s3Client.SetObjectAttrs(ctx, s.bucketName, key, attrs); err != nil {
        log.Printf("setting object attr: %s", err)
        return syscall.EINVAL
    }
    return 0
}

func (s *s3Node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    key := s.key("")
    s3File, err := newS3FileHandler(s.s3Embedded, key)
    if err != nil {
        log.Printf("error creating s3 file handler: %s", err)
        return nil, 0, syscall.EIO
    }
    return s3File, 0, 0
}

func (s *s3Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    key := s.key("")
    objs, err := s.s3Client.ListObjects(ctx, s.bucketName,
        s3_internal.ListObjectsWithPrefix(key),
    )
    if err != nil {
        log.Printf("error: %#v", err)
        // TODO: handle errors correctly
        return nil, syscall.ENOENT
    }

    dirs := make([]fuse.DirEntry, 0, len(objs))
    for _, obj := range objs {
        dirKey := fmt.Sprintf("%s/", key)
        if strings.HasPrefix(*obj.Key, dirKey) && *obj.Key != dirKey {
            d := fuse.DirEntry{
                Name: *obj.Key,
                Ino:  uniqueInode(*obj.Key),
                Mode: keyToFileMode(*obj.Key),
            }
            dirs = append(dirs, d)
        }
    }
    return fs.NewListDirStream(dirs), 0
}

func (s *s3Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    key := s.key(name)
    _, err := s.s3Client.GetObject(ctx, s.bucketName, key)
    if err != nil {
        var notFoundErr *s3_internal.ErrNotFound
        if !errors.As(err, &notFoundErr) {
            log.Println(err)
            return nil, syscall.EIO
        }
        // since fuse removes the slashes on the name, retry lookup with slash suffixed to make sure we find folders
        key = key + "/"
        _, err = s.s3Client.GetObject(ctx, s.bucketName, key)
        // clear error of previous invocation
        notFoundErr = nil
        if errors.As(err, &notFoundErr) {
            return nil, syscall.ENOENT
        }
    }

    newInode := s.newInode(ctx, key)
    s.setAttrs(ctx, key, &out.Attr)
    if success := s.AddChild(name, newInode, true); !success {
        log.Printf("adding new inode as child to %#v was not successfull", s)
    }
    return newInode, 0
}

func (s *s3Node) Access(ctx context.Context, mask uint32) syscall.Errno {
    return 0
}

func (s *s3Node) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // root node has key set to ""
    if s.EmbeddedInode().IsRoot() {
        out = rootDirAttrs()
        return 0
    }

    key := s.key("")
    s.setAttrs(ctx, key, &out.Attr)
    return 0
}

func (s *s3Node) setAttrs(ctx context.Context, key string, out *fuse.Attr) {
    out.Ino = uniqueInode(key)
    out.Mode = keyToFileMode(key)
    attrs := s.getS3Attrs(ctx, key)
    if attrs != nil {
        out.Size = attrs.Size
        out.SetTimes(attrs.AccessTime, attrs.ModifyTime, attrs.ChangeTime)
    }
}

// key returns the filepath to the root node to use as the object name in s3
func (s *s3Node) key(childName string) string {
    nodePath := s.Path(s.Root())
    return filepath.Join(nodePath, childName)
}

func (s *s3Node) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    childKey := s.key(name)
    err := s.createEmptyObject(ctx, childKey)
    if err != nil {
        log.Printf("error creating object: %s", err)
        return nil, nil, 0, syscall.EIO
    }
    node = s.newInode(ctx, childKey)
    s3File, err := newS3FileHandler(s.s3Embedded, childKey)
    if err != nil {
        log.Printf("error s3 file handler: %s", err)
        return nil, nil, 0, syscall.EIO
    }
    s.setAttrs(ctx, childKey, &out.Attr)
    return node, s3File, 0, 0
}

func (s *s3Node) Unlink(ctx context.Context, name string) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    childKey := s.key(name)
    if err := s.s3Client.DeleteObject(ctx, s.bucketName, childKey); err != nil {
        return syscall.EIO
    }
    return 0
}

func (s *s3Node) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    oldChildKey := s.key(name)
    newChildKey := s.key(newName)
    if err := s.s3Copy(ctx, oldChildKey, newChildKey); err != nil {
        log.Println(err)
        return syscall.EIO
    }

    success := s.MvChild(name, newParent.EmbeddedInode(), newName, true)
    if !success {
        log.Println("moving child was not successful")
    }
    return 0
}

func (s *s3Node) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // fuse removes the trailing slash
    // must be readded for server to recognize name as folder
    childKey := s.key(name) + "/"

    if err := s.createEmptyObject(ctx, childKey); err != nil {
        log.Printf("error creating empty object: %s", err)
        return nil, syscall.EIO
    }
    s.setAttrs(ctx, childKey, &out.Attr)

    return s.newInode(ctx, name), 0
}

var (
    // Node types must be InodeEmbedders
    _ fs.InodeEmbedder = (*s3Node)(nil)
    _ fs.NodeLookuper  = (*s3Node)(nil)
    _ fs.NodeReaddirer = (*s3Node)(nil)

    _ fs.NodeSetattrer = (*s3Node)(nil)
    _ fs.NodeGetattrer = (*s3Node)(nil)

    _ fs.NodeAccesser = (*s3Node)(nil)

    _ fs.NodeMkdirer = (*s3Node)(nil)

    _ fs.NodeUnlinker = (*s3Node)(nil)
    _ fs.NodeRenamer  = (*s3Node)(nil)

    _ fs.NodeCreater = (*s3Node)(nil)

    // Implement (handleless) Open
    _ fs.NodeOpener = (*s3Node)(nil)

    _ fs.NodeRmdirer = (*s3Node)(nil)
)
