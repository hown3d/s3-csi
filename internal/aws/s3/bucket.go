package s3

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

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
