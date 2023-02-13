package s3

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Bucket struct {
    Name string

    client *Client
}

// NewBucket creates an instance of the Bucket struct.
// This does not imply that the object exists in S3.
// To create the bucket inside S3, see the Create method.
func (c *Client) NewBucket(name string) *Bucket {
    return &Bucket{
        Name:   name,
        client: c,
    }
}

func (b *Bucket) Exists(ctx context.Context) bool {
    _, err := b.client.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
        Bucket: &b.Name,
    })
    return err == nil
}

// Create creates the bucket in s3
func (b *Bucket) Create(ctx context.Context, location string) error {
    input := &s3.CreateBucketInput{
        Bucket:                     &b.Name,
        ACL:                        types.BucketCannedACLPrivate,
        ObjectLockEnabledForBucket: true,
    }

    if location != "" {
        var (
            valid      bool
            constraint types.BucketLocationConstraint
        )
        for _, val := range constraint.Values() {
            if string(val) == location {
                valid = true
            }
        }
        if !valid {
            return fmt.Errorf("location %s is not a valid BucketLocation", location)
        }

        input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
            LocationConstraint: types.BucketLocationConstraint(location),
        }
    }

    _, err := b.client.s3Client.CreateBucket(ctx, input)
    if err != nil {
        return wrapError(fmt.Errorf("creating bucket %s: %w", b.Name, err), accessDeniedErrCheckFunc)
    }
    return nil
}
func (b *Bucket) Delete(ctx context.Context) error {
    input := &s3.DeleteBucketInput{
        Bucket: &b.Name,
    }
    // clean bucket because only Empty buckets can be deleted
    err := b.Empty(ctx)
    if err != nil {
        return fmt.Errorf("cleaning bucket %s: %w", b.Name, err)
    }
    _, err = b.client.s3Client.DeleteBucket(ctx, input)
    if err != nil {
        return wrapError(
            fmt.Errorf("deleting bucket %s: %w", b.Name, err),
            bucketErrCheckFuncs(b)...,
        )
    }
    return nil
}

func (b *Bucket) AddMetadataKey(ctx context.Context, key string, value string) error {
    metadata, err := b.GetMetadata(ctx)
    if err != nil {
        return fmt.Errorf("getting existing metadata: %w", err)
    }
    metadata[key] = value
    return b.persistMetadata(ctx, metadata)
}

func (b *Bucket) AddMetadata(ctx context.Context, addMetadata Metadata) error {
    metadata, err := b.GetMetadata(ctx)
    if err != nil {
        return fmt.Errorf("getting existing metadata: %w", err)
    }
    metadata = mergeMetadata(metadata, addMetadata)
    return b.persistMetadata(ctx, metadata)
}

// persistMetadata calls the s3 api and puts metadata as tags
func (b *Bucket) persistMetadata(ctx context.Context, metadata Metadata) error {
    input := &s3.PutBucketTaggingInput{
        Bucket: &b.Name,
        Tagging: &types.Tagging{
            TagSet: metadata.toTagSlice(),
        },
    }
    _, err := b.client.s3Client.PutBucketTagging(ctx, input)
    if err != nil {
        return wrapError(
            fmt.Errorf("put tags to bucket: %w", err),
            bucketErrCheckFuncs(b)...,
        )
    }
    return nil
}

func (b *Bucket) GetMetadata(ctx context.Context) (Metadata, error) {
    input := &s3.GetBucketTaggingInput{
        Bucket: &b.Name,
    }
    out, err := b.client.s3Client.GetBucketTagging(ctx, input)
    if err != nil {
        return nil, wrapError(
            fmt.Errorf("getting bucket tags: %w", err),
            bucketErrCheckFuncs(b)...,
        )
    }
    return metadataFromTagSlice(out.TagSet), nil
}

func (b *Bucket) Empty(ctx context.Context) error {
    objs, err := b.client.ListObjects(ctx, b.Name, nil)
    if err != nil {
        return err
    }
    return b.client.DeleteObjects(ctx, b.Name, objs)
}
