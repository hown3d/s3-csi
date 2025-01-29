package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

// for testing override
var TimeNowFunc = time.Now

// Object represents an S3 Object
type Object struct {
	Bucket       *Bucket
	key          string
	size         int64
	lastModified time.Time
	client       *Client
	obj          *minio.Object
}

func (b *Bucket) NewObject(ctx context.Context, key string) *Object {
	o := &Object{
		Bucket: b,
		key:    key,
		client: b.client,
	}

	return o
}

func (o *Object) Write(p []byte) (n int, err error) {
	r := bytes.NewReader(p)
	info, err := o.client.s3Client.PutObject(context.TODO(), o.Bucket.Name, o.key, r, int64(len(p)), minio.PutObjectOptions{})
	o.size = info.Size
	o.lastModified = o.lastModified
	return int(info.Size), nil
}

// Close implements io.Closer.
func (o *Object) Close() error {
	if o.obj == nil {
		return nil
	}
	err := o.obj.Close()
	if err != nil {
		return err
	}
	o.obj = nil
	return nil
}

// Read implements io.Reader.
func (o *Object) Read(p []byte) (n int, err error) {
	err = o.fetch()
	if err != nil {
		return 0, err
	}
	return o.obj.Read(p)
}

// Seek implements io.Seeker.
func (o *Object) Seek(offset int64, whence int) (int64, error) {
	err := o.fetch()
	if err != nil {
		return 0, err
	}
	return o.obj.Seek(offset, whence)
}

// ReadAt implements io.ReaderAt.
func (o *Object) ReadAt(p []byte, off int64) (n int, err error) {
	if o.obj == nil {
		err = o.fetch()
		if err != nil {
			return 0, err
		}
	}
	return o.obj.ReadAt(p, off)
}

var (
	_ io.Reader   = (*Object)(nil)
	_ io.ReaderAt = (*Object)(nil)
	_ io.Seeker   = (*Object)(nil)
	_ io.Closer   = (*Object)(nil)
)

func (o *Object) Exists(ctx context.Context) (bool, error) {
	if _, err := o.client.s3Client.StatObject(ctx, o.Bucket.Name, o.key, minio.GetObjectOptions{}); err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.Code == "NoSuchKey" || resp.Code == "NoSuchBucket" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (o *Object) Delete(ctx context.Context) error {
	err := o.client.s3Client.RemoveObject(ctx, o.Bucket.Name, o.key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("error deleting object %s in bucket %s: %w", o.key, o.Bucket, err)
	}
	return nil
}

func (o *Object) Copy(ctx context.Context, newKey string) (*Object, error) {
	src := minio.CopySrcOptions{
		Bucket: o.Bucket.Name,
		Object: o.key,
	}
	dst := minio.CopyDestOptions{
		Bucket: o.Bucket.Name,
		Object: newKey,
	}
	info, err := o.client.s3Client.CopyObject(ctx, dst, src)
	if err != nil {
		return nil, fmt.Errorf("error copying object %s to %s: %w", o.key, newKey, err)
	}

	return &Object{
		Bucket:       o.Bucket,
		key:          newKey,
		size:         info.Size,
		lastModified: info.LastModified,
		client:       o.client,
	}, nil
}

func (o *Object) Move(ctx context.Context, newKey string) error {
	newObj, err := o.Copy(ctx, newKey)
	if err != nil {
		return err
	}
	if err = o.Delete(ctx); err != nil {
		return err
	}
	*o = *newObj
	return nil
}

func (o *Object) fetch() error {
	if o.obj != nil {
		return nil
	}
	obj, err := o.client.s3Client.GetObject(context.TODO(), o.Bucket.Name, o.key, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	info, err := obj.Stat()
	if err != nil {
		return err
	}
	o.obj = obj
	o.lastModified = info.LastModified
	o.size = info.Size
	return nil
}

func (o *Object) String() string {
	return o.key
}
