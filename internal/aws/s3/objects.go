package s3

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type ListOpts struct {
    Prefix string
}

func (c *Client) ListObjects(ctx context.Context, bucketName string, opts *ListOpts) ([]*Object, error) {
    if opts == nil {
        opts = &ListOpts{}
    }
    input := &s3.ListObjectsV2Input{
        Bucket: &bucketName,
    }
    if opts.Prefix != "" {
        input.Prefix = &opts.Prefix
    }

    out, err := c.s3Client.ListObjectsV2(ctx, input)
    if err != nil {
        return nil, wrapError(err, accessDeniedErrCheckFunc, bucketNotFoundErrCheckFunc(bucketName))
    }
    awsObjs := out.Contents

    objs := make([]*Object, 0, len(awsObjs))
    for _, awsObj := range awsObjs {
        newObj := c.NewObject(bucketName, *awsObj.Key)
        objs = append(objs, newObj)
    }
    return objs, nil
}

func (c *Client) DeleteObjects(ctx context.Context, bucketName string, objs []*Object) error {
    objIdents := make([]types.ObjectIdentifier, len(objs))
    for i, obj := range objs {
        objIdents[i] = types.ObjectIdentifier{
            Key: &obj.Key,
        }
    }
    input := &s3.DeleteObjectsInput{
        Bucket: &bucketName,
        Delete: &types.Delete{
            Objects: objIdents,
        },
    }
    _, err := c.s3Client.DeleteObjects(ctx, input)
    if err != nil {
        // deleteObjects will not fail if the provided objKey doesn't exist. So no need to worry about ErrObjectNotFound
        return wrapError(
            fmt.Errorf("deleting objects from bucket %s: %w", bucketName, err),
            accessDeniedErrCheckFunc,
            bucketNotFoundErrCheckFunc(bucketName),
        )
    }
    return nil
}
