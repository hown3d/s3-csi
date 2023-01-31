package s3

import (
    "context"
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (c *Client) GetBucket(ctx context.Context, name string) (*Bucket, error) {
    in := &s3.HeadBucketInput{
        Bucket: &name,
    }
    _, err := c.s3Client.HeadBucket(ctx, in)
    if err != nil {
        var noSuchKeyErr *types.NoSuchBucket
        if errors.As(err, &noSuchKeyErr) {
            return nil, NewErrBucketNotFound(name)
        }
        return nil, fmt.Errorf("getting bucket %s: %w", name, err)
    }
    return &Bucket{
        Name:     name,
        s3Client: c.s3Client,
    }, nil
}

func (c *Client) ListBuckets(ctx context.Context) ([]*Bucket, error) {
    out, err := c.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
    if err != nil {
        return nil, err
    }
    buckets := make([]*Bucket, 0, len(out.Buckets))
    for _, bucket := range out.Buckets {
        b := &Bucket{
            Name:     *bucket.Name,
            s3Client: c.s3Client,
        }
        buckets = append(buckets, b)
    }
    return buckets, nil
}

func (c *Client) CreateBucket(ctx context.Context, name string, location string) (*Bucket, error) {
    input := &s3.CreateBucketInput{
        Bucket:                     &name,
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
            return nil, fmt.Errorf("location %s is not a valid BucketLocation", location)
        }

        input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
            LocationConstraint: types.BucketLocationConstraint(location),
        }
    }

    _, err := c.s3Client.CreateBucket(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("creating bucket %s: %w", name, err)
    }
    return &Bucket{
        Name:     name,
        s3Client: c.s3Client,
    }, nil
}
