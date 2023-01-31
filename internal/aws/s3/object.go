package s3

import (
    "context"
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/aws/smithy-go"
    "io"
    "time"
)

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

func ListObjectsWithPrefix(prefix string) func(in *s3.ListObjectsV2Input) {
    return func(in *s3.ListObjectsV2Input) {
        in.Prefix = &prefix
    }
}

func ListObjectsWithDelimiter(delimiter string) func(in *s3.ListObjectsV2Input) {
    return func(in *s3.ListObjectsV2Input) {
        in.Delimiter = &delimiter
    }
}

func (c *Client) ListObjects(ctx context.Context, bucketName string, opts ...func(in *s3.ListObjectsV2Input)) ([]types.Object, error) {
    input := &s3.ListObjectsV2Input{
        Bucket: &bucketName,
    }
    for _, opt := range opts {
        opt(input)
    }
    out, err := c.s3Client.ListObjectsV2(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("error listing objects in bucket %s: %w", bucketName, err)
    }
    return out.Contents, nil
}

func (c *Client) GetObject(ctx context.Context, bucketName, objName string, opts ...func(in *s3.GetObjectInput)) (*s3.GetObjectOutput, error) {
    input := &s3.GetObjectInput{
        Bucket: &bucketName,
        Key:    &objName,
    }
    for _, opt := range opts {
        opt(input)
    }
    out, err := c.s3Client.GetObject(ctx, input)
    if err != nil {
        var noSuchKeyErr *types.NoSuchKey
        if errors.As(err, &noSuchKeyErr) {
            return nil, NewErrNotFound(bucketName, objName)
        }
        return nil, fmt.Errorf("error getting object %s in bucket %s: %w", objName, bucketName, err)
    }
    return out, err
}

func (c *Client) GetObjectAttrs(ctx context.Context, bucketName, objName string) (*ObjectAttrs, error) {
    input := &s3.HeadObjectInput{
        Bucket: &bucketName,
        Key:    &objName,
    }
    out, err := c.s3Client.HeadObject(ctx, input)
    if err != nil {
        var smithyErr *smithy.GenericAPIError
        if errors.As(err, &smithyErr) {
            if smithyErr.Code == "NotFound" {
                return nil, NewErrNotFound(bucketName, objName)
            }
        }
        return nil, err
    }

    changeTime, err := time.Parse(time.UnixDate, out.Metadata[CHANGE_TIME_METADATA_KEY])
    if err != nil {
        return nil, fmt.Errorf("error parsing change-time from object metadata: %w", err)
    }

    now := time.Now()

    return &ObjectAttrs{
        Size:       uint64(out.ContentLength),
        ModifyTime: out.LastModified,
        AccessTime: &now,
        ChangeTime: &changeTime,
    }, nil
}

func (c *Client) SetObjectAttrs(ctx context.Context, bucketName, objName string, attrs *ObjectAttrs) error {
    tagSet := attrs.toTagSlice()
    if len(tagSet) == 0 {
        return nil
    }
    input := &s3.PutObjectTaggingInput{
        Bucket: &bucketName,
        Key:    &objName,
        Tagging: &types.Tagging{
            TagSet: tagSet,
        },
    }
    _, err := c.s3Client.PutObjectTagging(ctx, input)
    if err != nil {
        return fmt.Errorf("error updating object tags: %w", err)
    }
    return nil
}

func (c *Client) WriteObject(ctx context.Context, bucketName string, objName string, data io.Reader) error {
    uploader := manager.NewUploader(c.s3Client)
    metadata, err := createObjectMetadata()
    if err != nil {
        return fmt.Errorf("error creating metadata: %w", err)
    }
    _, err = uploader.Upload(ctx, &s3.PutObjectInput{
        Bucket:   &bucketName,
        Key:      &objName,
        Body:     data,
        Metadata: metadata,
    })
    return err
}

func (c *Client) CopyObject(ctx context.Context, bucketName string, oldObjName, newObjName string) error {
    input := &s3.CopyObjectInput{
        Bucket:            &bucketName,
        CopySource:        &oldObjName,
        Key:               &newObjName,
        MetadataDirective: types.MetadataDirectiveCopy,
    }
    _, err := c.s3Client.CopyObject(ctx, input)
    if err != nil {
        return fmt.Errorf("error copying object %s to %s: %w", oldObjName, newObjName, err)
    }
    return nil
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

func createObjectMetadata() (map[string]string, error) {
    return map[string]string{
        CHANGE_TIME_METADATA_KEY: time.Now().Format(time.UnixDate),
    }, nil
}
