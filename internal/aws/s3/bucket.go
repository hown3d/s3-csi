package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
)

type Bucket struct {
	Name string

	client *Client
}

type Tag struct {
	Key   string
	Value string
}

func (c *Client) NewBucket(ctx context.Context, region string, name string) (*Bucket, error) {
	b := &Bucket{
		Name:   name,
		client: c,
	}
	ok, err := b.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		err := c.s3Client.MakeBucket(ctx, b.Name, minio.MakeBucketOptions{
			Region: region,
		})
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (b *Bucket) AddMetadata(ctx context.Context, tags ...Tag) error {
	bucketTags, err := b.client.s3Client.GetBucketTagging(ctx, b.Name)
	if err != nil {
		return err
	}
	for _, t := range tags {
		bucketTags.Set(t.Key, t.Value)
	}
	return b.client.s3Client.SetBucketTagging(ctx, b.Name, bucketTags)
}

func (b *Bucket) Exists(ctx context.Context) (bool, error) {
	return b.client.s3Client.BucketExists(ctx, b.Name)
}

func (b *Bucket) Objects(ctx context.Context, prefix string) []*Object {
	objs := []*Object{}
	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    prefix,
	}
	for obj := range b.client.s3Client.ListObjects(ctx, b.Name, opts) {
		o := &Object{
			Bucket:       b,
			key:          obj.Key,
			size:         obj.Size,
			lastModified: obj.LastModified,
			client:       b.client,
		}
		objs = append(objs, o)
	}
	return objs
}

func (b *Bucket) Delete(ctx context.Context) error {
	// clean bucket because only Empty buckets can be deleted
	err := b.empty(ctx)
	if err != nil {
		return fmt.Errorf("cleaning bucket %s: %w", b.Name, err)
	}
	return b.client.s3Client.RemoveBucket(ctx, b.Name)
}

func (b *Bucket) empty(ctx context.Context) error {
	objs := b.Objects(ctx, "")
	var err error
	for _, o := range objs {
		err = errors.Join(err, o.Delete(ctx))
	}
	return err
}

func (b *Bucket) String() string {
	return b.Name
}
