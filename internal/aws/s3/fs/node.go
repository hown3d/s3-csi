package fs

import (
    "context"
    "errors"
    "fmt"
    "github.com/hanwen/go-fuse/v2/fs"
    "github.com/hanwen/go-fuse/v2/fuse"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "k8s.io/klog/v2"
    "log"
    "path/filepath"
    "strings"
    "sync"
    "syscall"
)

type s3Node struct {
    fs.Inode

    client     *s3.Client
    bucketName string
    // When file systems are mutable, all access must use
    // synchronization.
    mu sync.RWMutex
}

func (s *s3Node) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

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

    key := s.key("")
    obj := s.client.NewObject(s.bucketName, key)
    if err := obj.SetAttrs(ctx, attrs); err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return syscall.ENOENT
        }
        printError("Setattr", err)
        return syscall.EIO
    }
    return 0
}

func (s *s3Node) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    key := s.key("")
    obj := s.client.NewObject(s.bucketName, key)
    err := setFuseAttr(ctx, obj, &out.Attr)
    if err != 0 {
        printError("Getattr", err)
        return err
    }
    return 0
}

// folderObjs returns the list of objects in this directory
func (s *s3Node) folderObjs(ctx context.Context, folderKey string) ([]*s3.Object, syscall.Errno) {
    listOpts := s3.ListOpts{
        Prefix: folderKey,
    }
    klog.V(5).Infof("listing objects with prefix: %s", folderKey)
    objs, err := s.client.ListObjects(ctx, s.bucketName, &listOpts)
    if err != nil {
        return nil, syscall.EINVAL
    }
    if len(objs) == 0 {
        return nil, syscall.ENOENT
    }
    return objs, 0
}

func (s *s3Node) Rmdir(ctx context.Context, name string) syscall.Errno {
    folderKey := s.folderKey(name)
    objs, err := s.folderObjs(ctx, folderKey)
    if err != 0 {
        return err
    }

    if err := s.client.DeleteObjects(ctx, s.bucketName, objs); err != nil {
        return syscall.EINVAL
    }
    return 0
}

func (s *s3Node) newInode(ctx context.Context, key string) *fs.Inode {
    node := &s3Node{
        bucketName: s.bucketName,
        client:     s.client,
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
    obj := s.client.NewObject(s.bucketName, key)
    if !obj.Exists(ctx) {
        return nil, 0, syscall.ENOENT
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

    folderKey := s.folderKey("")
    objs, err := s.folderObjs(ctx, folderKey)
    if err != 0 {
        return nil, err
    }

    dirs := make([]fuse.DirEntry, 0, len(objs))
    for _, obj := range objs {
        if obj.Key == ROOT_KEY {
            klog.V(5).Infof("skipping root object: %s", ROOT_KEY)
            continue
        }
        d := fuse.DirEntry{
            Name: strings.ReplaceAll(obj.Key, ROOT_KEY, ""),
            Ino:  uniqueInode(obj.Key),
            Mode: keyToFileMode(obj.Key),
        }
        klog.V(5).Infof("adding dir entry: %s", d)
        dirs = append(dirs, d)
    }
    return fs.NewListDirStream(dirs), 0
}

func (s *s3Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    key := s.key(name)
    obj := s.client.NewObject(s.bucketName, key)
    if !obj.Exists(ctx) {
        // since fuse removes the slashes on the name, retry lookup with slash suffixed to make sure we find folders
        key = s.folderKey(name)
        obj.Key = key
        if !obj.Exists(ctx) {
            return nil, syscall.ENOENT
        }
    }

    newInode := s.newInode(ctx, key)
    err := setFuseAttr(ctx, obj, &out.Attr)
    if err != 0 {
        return nil, err
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
    // can't use ROOT_KEY inside filepath.Join
    // because if nodePath and childName are empty,
    // the trailing / of ROOT_KEY would be removed
    return ROOT_KEY + filepath.Join(nodePath, childName)
}

// folderKey retrieves the key for a s3 object that is used as a folder.
// fuse removes the trailing slash
// must be readded for server to recognize name as folder
func (s *s3Node) folderKey(childName string) string {
    key := s.key(childName)
    if !strings.HasSuffix(key, "/") {
        key = key + "/"
    }
    return key
}

func (s *s3Node) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
    s.mu.Lock()
    defer s.mu.Unlock()

    childKey := s.key(name)
    obj := s.client.NewObject(s.bucketName, childKey)
    err := obj.Create(ctx, nil)
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

    if err := setFuseAttr(ctx, obj, &out.Attr); err != 0 {
        printError("Create", fmt.Errorf("error setting fuse attrs: %w", err))
        return nil, nil, 0, err
    }
    return node, s3F, 0, 0
}

func (s *s3Node) Unlink(ctx context.Context, name string) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    childKey := s.key(name)
    obj := s.client.NewObject(s.bucketName, childKey)
    klog.V(5).Infof("deleting object %s", childKey)
    err := obj.Delete(ctx)
    if err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return syscall.ENOENT
        }
        printError("Unlink", fmt.Errorf("error deleting object: %w", err))
        return syscall.EIO
    }
    return 0
}

func (s *s3Node) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
    s.mu.Lock()
    defer s.mu.Unlock()

    oldChildKey := s.key(name)
    newChildKey := s.key(newName)
    klog.V(5).Infof("renaming object %s to %s", oldChildKey, newChildKey)
    obj := s.client.NewObject(s.bucketName, oldChildKey)
    if err := obj.Move(ctx, newChildKey); err != nil {
        var notFoundErr *s3.ErrObjectNotFound
        if errors.As(err, &notFoundErr) {
            return syscall.ENOENT
        }
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
    folderKey := s.folderKey(name)

    obj := s.client.NewObject(s.bucketName, folderKey)
    err := obj.Create(ctx, nil)
    if err != nil {
        printError("Mkdir", fmt.Errorf("error creating object: %w", err))
        return nil, syscall.EIO
    }
    if err := setFuseAttr(ctx, obj, &out.Attr); err != 0 {
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

    _ fs.NodeSetattrer = (*s3Node)(nil)

    _ fs.NodeMkdirer = (*s3Node)(nil)

    _ fs.NodeSetattrer = (*s3Node)(nil)
    _ fs.NodeGetattrer = (*s3Node)(nil)

    _ fs.NodeUnlinker = (*s3Node)(nil)
    _ fs.NodeRenamer  = (*s3Node)(nil)

    _ fs.NodeCreater = (*s3Node)(nil)

    // Implement (handleless) Open
    _ fs.NodeOpener = (*s3Node)(nil)

    _ fs.NodeRmdirer = (*s3Node)(nil)
)
