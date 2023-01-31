package fs

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "io"
    "log"
    "syscall"
)

type s3Embedded struct {
    s3Client   *s3_internal.Client
    bucketName string
}

func (e *s3Embedded) createEmptyObject(ctx context.Context, name string) error {
    if err := e.s3Client.WriteObject(ctx, e.bucketName, name, new(bytes.Buffer)); err != nil {
        return fmt.Errorf("error creating folder in s3: %e", err)
    }
    return nil
}

func (e *s3Embedded) getS3Attrs(ctx context.Context, key string) *s3_internal.ObjectAttrs {
    attrs, err := e.s3Client.GetObjectAttrs(ctx, e.bucketName, key)
    if err != nil {
        var notFound *s3_internal.ErrNotFound
        if !errors.As(err, &notFound) {
            log.Printf("error retrieving metadata for node: %e", err)
        }
        return nil
    }
    return attrs
}

func (e *s3Embedded) s3Copy(ctx context.Context, oldKey, newKey string) error {
    if err := e.s3Client.CopyObject(ctx, e.bucketName, oldKey, newKey); err != nil {
        log.Printf("error copying object: %e", err)
        return syscall.EIO
    }
    if err := e.s3Client.DeleteObject(ctx, e.bucketName, oldKey); err != nil {
        return fmt.Errorf("error deleting old object after s3Copy: %w", err)
    }
    return nil
}

func (e *s3Embedded) s3Write(ctx context.Context, key string, data io.Reader) error {
    err := e.s3Client.WriteObject(ctx, e.bucketName, key, data)
    if err != nil {
        return err
    }
    return nil
}
