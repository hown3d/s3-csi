package s3

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Client struct {
    s3Client *s3.Client
}

func NewClient(cfg aws.Config) *Client {
    return &Client{
        s3Client: s3.NewFromConfig(cfg),
    }
}

func (c *Client) CreateBucket(ctx context.Context, name string, location string) error {
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
            return fmt.Errorf("location %s is not a valid BucketLocation", location)
        }

        input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
            LocationConstraint: types.BucketLocationConstraint(location),
        }
    }

    _, err := c.s3Client.CreateBucket(ctx, input)
    if err != nil {
        return fmt.Errorf("error creating bucket %s: %w", name, err)
    }
    return nil
}

func (c *Client) DeleteBucket(ctx context.Context, name string) error {
    input := &s3.DeleteBucketInput{
        Bucket: &name,
    }
    // clean bucket because only empty buckets can be deleted
    err := c.deleteAllObjects(ctx, name)
    if err != nil {
        return fmt.Errorf("error cleaning bucket %s: %w", name, err)
    }
    _, err = c.s3Client.DeleteBucket(ctx, input)
    if err != nil {
        return fmt.Errorf("error deleting bucket %s: %w", name, err)
    }
    return nil
}

func (c *Client) DeleteObject(ctx context.Context, bucketName string, objName string) error {
    input := &s3.DeleteObjectInput{
        Bucket: &bucketName,
        Key:    &objName,
    }
    _, err := c.s3Client.DeleteObject(ctx, input)
    if err != nil {
        return fmt.Errorf("error deleting object %s in bucket %s: %w", objName, bucketName, err)
    }
    return nil
}

func (c *Client) DeleteObjects(ctx context.Context, bucketName string, objs []string) error {
    objIdents := make([]types.ObjectIdentifier, len(objs))
    for _, objName := range objs {
        objIdents = append(objIdents, types.ObjectIdentifier{
            Key: &objName,
        })
    }
    input := &s3.DeleteObjectsInput{
        Bucket: &bucketName,
        Delete: &types.Delete{
            Objects: objIdents,
        },
    }
    _, err := c.s3Client.DeleteObjects(ctx, input)
    if err != nil {
        return fmt.Errorf("error deleting objects from bucket %s: %w", bucketName, err)
    }
    return nil
}

func (c *Client) ListObjects(ctx context.Context, bucketName string) ([]types.Object, error) {
    input := &s3.ListObjectsV2Input{
        Bucket: &bucketName,
    }
    out, err := c.s3Client.ListObjectsV2(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("error listing objects in bucket %s: %w", bucketName, err)
    }
    return out.Contents, nil
}

func (c *Client) deleteAllObjects(ctx context.Context, bucketName string) error {
    objs, err := c.ListObjects(ctx, bucketName)
    if err != nil {
        return err
    }
    objNames := make([]string, len(objs))
    for _, obj := range objs {
        objNames = append(objNames, *obj.Key)
    }
    return c.DeleteObjects(ctx, bucketName, objNames)
}
