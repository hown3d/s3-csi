package fs

import (
    "context"
    "errors"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "log"
    "path/filepath"
    "strings"
    "sync"
    "syscall"
)

type s3Node struct {
    fs.Inode

    bucket *s3.Bucket
    // When file systems are mutable, all access must use
    // synchronization.
    mu sync.RWMutex
}

func (s *s3Node) Rmdir(ctx context.Context, name string) syscall.Errno {
    folderKey := s.key(name) + "/"
    listOpts := s3.ListOpts{
        Prefix: folderKey,
    }
    objs, err := s.bucket.ListObjects(ctx, &listOpts)
    if err != nil {
        return syscall.EINVAL
    }
    if len(objs) == 0 {
        return syscall.ENOENT
    }

    objKeys := make([]string, 0, len(objs))
    for _, obj := range objs {
        objKeys = append(objKeys, obj.Key)
    }
    if err := s.bucket.DeleteObjects(ctx, objKeys); err != nil {
        return syscall.EINVAL
    }
    return 0
}

func (s *s3Node) newInode(ctx context.Context, key string) *fs.Inode {
    node := &s3Node{
        bucket: s.bucket,
    }
    stable := fs.StableAttr{
        Mode: keyToFileMode(key),
        Ino:  uniqueInode(key),
    }

    newInode := s.NewInode(ctx, node, stable)
    return newInode
}

func (s *s3Node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    key := s.key("")
    obj, err := s.bucket.GetObject(ctx, key)
    if err != nil {
        log.Printf("error getting object: %s", err)
        return nil, 0, syscall.EIO
    }
    s3F, err := newS3FileHandler(obj)
    if err != nil {
        log.Printf("error creating s3 file handler: %s", err)
        return nil, 0, syscall.EIO
    }
    return s3F, 0, 0
}

func (s *s3Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    key := s.key("")
    listOpts := s3.ListOpts{
        Prefix: key,
    }
    objs, err := s.bucket.ListObjects(ctx, &listOpts)
    if err != nil {
        log.Printf("error: %#v", err)
        // TODO: handle errors correctly
        return nil, syscall.ENOENT
    }

    dirs := make([]fuse.DirEntry, 0, len(objs))
    for _, obj := range objs {
        dirKey := fmt.Sprintf("%s/", key)
        if strings.HasPrefix(obj.Key, dirKey) && obj.Key != dirKey {
            d := fuse.DirEntry{
                Name: obj.Key,
                Ino:  uniqueInode(obj.Key),
                Mode: keyToFileMode(obj.Key),
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
    var (
        obj *s3.Object
        err error
    )
    obj, err = s.bucket.GetObject(ctx, key)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if !errors.As(err, &notFoundErr) {
            printError("Lookup", err)
            return nil, syscall.EIO
        }
        // since fuse removes the slashes on the name, retry lookup with slash suffixed to make sure we find folders
        key = key + "/"
        obj, err = s.bucket.GetObject(ctx, key)
        // clear error of previous invocation
        notFoundErr = nil
        if errors.As(err, &notFoundErr) {
            return nil, syscall.ENOENT
        }
    }

    newInode := s.newInode(ctx, key)
    err = setFuseAttr(ctx, obj, &out.Attr)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return nil, syscall.ENOENT
        }
    }

    if success := s.AddChild(name, newInode, true); !success {
        printError("Lookup", fmt.Errorf("adding new inode as child to %#v was not successfull", s))
    }
    return newInode, 0
}

func (s *s3Node) Access(ctx context.Context, mask uint32) syscall.Errno {
    return 0
}

// key returns the filepath to the root node to use as the object name in s3
func (s *s3Node) key(childName string) string {
    nodePath := s.Path(s.Root())
    return filepath.Join(nodePath, childName)
}

func (s *s3Node) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()

    childKey := s.key(name)
    obj, err := s.bucket.CreateObject(ctx, childKey, nil)
    if err != nil {
        printError("Create", fmt.Errorf("error creating object: %w", err))
        return nil, nil, 0, syscall.EIO
    }
    s3F, err := newS3FileHandler(obj)
    if err != nil {
        printError("Create", fmt.Errorf("error creating s3 file handler: %w", err))
        return nil, nil, 0, syscall.EIO
    }
    node = s.newInode(ctx, childKey)

    if err := setFuseAttr(ctx, obj, &out.Attr); err != nil {
        printError("Create", fmt.Errorf("error setting fuse attrs: %w", err))
        return nil, nil, 0, syscall.EIO
    }
    return node, s3F, 0, 0
}

func (s *s3Node) Unlink(ctx context.Context, name string) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    childKey := s.key(name)
    obj, err := s.bucket.GetObject(ctx, childKey)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return syscall.ENOENT
        }
        printError("Unlink", fmt.Errorf("error getting object: %w", err))
        return syscall.EIO
    }
    err = obj.Delete(ctx)
    if err != nil {
        printError("Unlink", fmt.Errorf("error deleting object"))
        return syscall.EIO
    }
    return 0
}

func (s *s3Node) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    oldChildKey := s.key(name)
    newChildKey := s.key(newName)
    obj, err := s.bucket.GetObject(ctx, oldChildKey)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return syscall.ENOENT
        }
        printError("Rename", fmt.Errorf("error getting object: %w", err))
        return syscall.EIO
    }
    if err := obj.Move(ctx, newChildKey); err != nil {
        printError("Rename", fmt.Errorf("error moving object: %w", err))
        return syscall.EIO
    }
    success := s.MvChild(name, newParent.EmbeddedInode(), newName, true)
    if !success {
        printError("Rename", errors.New("moving child was not successful"))
    }
    return 0
}

func (s *s3Node) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // fuse removes the trailing slash
    // must be readded for server to recognize name as folder
    childKey := s.key(name) + "/"

    obj, err := s.bucket.CreateObject(ctx, childKey, nil)
    if err != nil {
        printError("Mkdir", fmt.Errorf("error creating object: %w", err))
        return nil, syscall.EIO
    }
    if err := setFuseAttr(ctx, obj, &out.Attr); err != nil {
        printError("Mkdir", fmt.Errorf("error setting fuse attrs: %w", err))
        return nil, syscall.EIO
    }
    return s.newInode(ctx, name), 0
}

var (
    // Node types must be InodeEmbedders
    _ fs.InodeEmbedder = (*s3Node)(nil)
    _ fs.NodeLookuper  = (*s3Node)(nil)
    _ fs.NodeReaddirer = (*s3Node)(nil)

    _ fs.NodeAccesser = (*s3Node)(nil)

    _ fs.NodeMkdirer = (*s3Node)(nil)

    _ fs.NodeUnlinker = (*s3Node)(nil)
    _ fs.NodeRenamer  = (*s3Node)(nil)

    _ fs.NodeCreater = (*s3Node)(nil)

    // Implement (handleless) Open
    _ fs.NodeOpener = (*s3Node)(nil)

    _ fs.NodeRmdirer = (*s3Node)(nil)
)
