package s3

import (
    "context"
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/aws/smithy-go"
)

type Bucket struct {
    Name string

    s3Client *s3.Client
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
    _, err = b.s3Client.DeleteBucket(ctx, input)
    if err != nil {
        return fmt.Errorf("deleting bucket %s: %w", b.Name, err)
    }
    return nil
}

func (b *Bucket) Empty(ctx context.Context) error {
    objs, err := b.ListObjects(ctx, nil)
    if err != nil {
        return err
    }
    for _, obj := range objs {
        err := obj.Delete(ctx)
        if err != nil {
            return err
        }
    }

    return nil
}

type ListOpts struct {
    Prefix string
}

func (b *Bucket) ListObjects(ctx context.Context, opts *ListOpts) ([]*Object, error) {
    if opts == nil {
        opts = &ListOpts{}
    }
    input := &s3.ListObjectsV2Input{
        Bucket: &b.Name,
    }
    if opts.Prefix != "" {
        input.Prefix = &opts.Prefix
    }

    out, err := b.s3Client.ListObjectsV2(ctx, input)
    if err != nil {
        return nil, err
    }
    awsObjs := out.Contents

    objs := make([]*Object, 0, len(awsObjs))
    for _, awsObj := range awsObjs {
        newObj := &Object{
            Key:      *awsObj.Key,
            Bucket:   b.Name,
            s3Client: b.s3Client,
        }
        objs = append(objs, newObj)
    }
    return objs, nil
}

func (b *Bucket) DeleteObjects(ctx context.Context, objs []string) error {
    objIdents := make([]types.ObjectIdentifier, len(objs))
    for i, objName := range objs {
        objIdents[i] = types.ObjectIdentifier{
            Key: &objName,
        }
    }
    input := &s3.DeleteObjectsInput{
        Bucket: &b.Name,
        Delete: &types.Delete{
            Objects: objIdents,
        },
    }
    _, err := b.s3Client.DeleteObjects(ctx, input)
    if err != nil {
        return fmt.Errorf("deleting objects from bucket %s: %w", b.Name, err)
    }
    return nil
}

func (b *Bucket) GetObject(ctx context.Context, key string) (*Object, error) {
    input := &s3.HeadObjectInput{
        Bucket: &b.Name,
        Key:    &key,
    }
    _, err := b.s3Client.HeadObject(ctx, input)
    if err != nil {
        var apiErr smithy.APIError
        if errors.As(err, &apiErr) {
            if apiErr.ErrorCode() == "NotFound" {
                return nil, NewErrObjectNotFound(b.Name, key)
            }
        }
        return nil, fmt.Errorf("getting object %s in bucket %s: %w", key, b.Name, err)
    }
    return &Object{
        Bucket:   b.Name,
        Key:      key,
        s3Client: b.s3Client,
    }, nil
}

func (b *Bucket) CreateObject(ctx context.Context, key string, metadata Metadata) (*Object, error) {
    defaultsMetadata := defaultObjectMetadata()
    defaultsMetadata.merge(metadata)
    input := &s3.PutObjectInput{
        Bucket:   &b.Name,
        Key:      &key,
        Metadata: metadata,
    }
    _, err := b.s3Client.PutObject(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("creating object %s in bucket %s: %w", key, b.Name, err)
    }
    return &Object{
        Bucket:   b.Name,
        Key:      key,
        s3Client: b.s3Client,
    }, nil
}
