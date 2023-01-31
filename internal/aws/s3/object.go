package s3

import (
    "context"
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/aws/smithy-go"
    "io"
    "time"
)

type Object struct {
    Bucket string
    Key    string

    s3Client *s3.Client
}

func (o *Object) Delete(ctx context.Context) error {
    input := &s3.DeleteObjectInput{
        Bucket: &o.Bucket,
        Key:    &o.Key,
    }
    _, err := o.s3Client.DeleteObject(ctx, input)
    if err != nil {
        return fmt.Errorf("error deleting object %s in bucket %s: %w", o.Key, o.Bucket, err)
    }
    return nil

}

func (o *Object) GetAttrs(ctx context.Context) (*ObjectAttrs, error) {
    input := &s3.HeadObjectInput{
        Bucket: &o.Bucket,
        Key:    &o.Key,
    }
    out, err := o.s3Client.HeadObject(ctx, input)
    if err != nil {
        var smithyErr *smithy.GenericAPIError
        if errors.As(err, &smithyErr) {
            if smithyErr.Code == "NotFound" {
                return nil, NewErrObjectNotFound(o.Bucket, o.Key)
            }
        }
        return nil, err
    }

    attrs := &ObjectAttrs{
        Size:       uint64(out.ContentLength),
        ModifyTime: out.LastModified,
        AccessTime: aws.Time(time.Now()),
    }

    changeTimeStr, ok := out.Metadata[CHANGE_TIME_METADATA_KEY]
    if ok {
        changeTime, err := time.Parse(time.UnixDate, changeTimeStr)
        if err != nil {
            return nil, fmt.Errorf("error parsing change-time from object metadata: %w", err)
        }
        attrs.ChangeTime = &changeTime
    }

    return attrs, nil
}

func (o *Object) SetAttrs(ctx context.Context, attrs *ObjectAttrs) error {
    tagSet := attrs.toTagSlice()
    if len(tagSet) == 0 {
        return nil
    }
    input := &s3.PutObjectTaggingInput{
        Bucket: &o.Bucket,
        Key:    &o.Key,
        Tagging: &types.Tagging{
            TagSet: tagSet,
        },
    }
    _, err := o.s3Client.PutObjectTagging(ctx, input)
    if err != nil {
        return fmt.Errorf("error updating object tags: %w", err)
    }
    return nil
}

func (o *Object) Write(ctx context.Context, data io.Reader, metadata Metadata) error {
    defaultMetadata := defaultObjectMetadata()
    defaultMetadata.merge(metadata)

    uploader := manager.NewUploader(o.s3Client)
    _, err := uploader.Upload(ctx, &s3.PutObjectInput{
        Bucket:   &o.Bucket,
        Key:      &o.Key,
        Body:     data,
        Metadata: metadata,
    })
    return err
}

type ReadRange struct {
    Start int64
    End   int64
}

func (o *Object) Read(ctx context.Context, readRange *ReadRange) (io.Reader, error) {
    input := &s3.GetObjectInput{
        Bucket: &o.Bucket,
        Key:    &o.Key,
    }
    if readRange != nil {
        byteRange := fmt.Sprintf("bytes=%d-%d", readRange.Start, readRange.End)
        input.Range = &byteRange
    }

    out, err := o.s3Client.GetObject(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("getting object: %w", err)
    }
    return out.Body, nil
}

func (o *Object) Copy(ctx context.Context, newKey string) (*Object, error) {
    input := &s3.CopyObjectInput{
        Bucket:            &o.Bucket,
        CopySource:        &o.Key,
        Key:               &newKey,
        MetadataDirective: types.MetadataDirectiveCopy,
    }
    _, err := o.s3Client.CopyObject(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("error copying object %s to %s: %w", o.Key, newKey, err)
    }

    return &Object{
        Key:      newKey,
        Bucket:   o.Bucket,
        s3Client: o.s3Client,
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
