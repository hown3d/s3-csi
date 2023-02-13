package s3

import (
    "context"
    "fmt"
    "io"
    "strconv"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// for testing override
var TimeNowFunc = time.Now

// Object represents an S3 Object
type Object struct {
    Bucket string
    Key    string

    client *Client
}

// NewObject creates an instance of the Object struct.
// This does not imply that the object exists in S3.
// To create the object inside S3, see the Create method.
func (c *Client) NewObject(bucketName string, key string) *Object {
    return &Object{
        Bucket: bucketName,
        Key:    key,
        client: c,
    }
}

func (o *Object) Exists(ctx context.Context) bool {
    _, err := o.head(ctx)
    return err == nil
}

func (o *Object) Create(ctx context.Context, metadata Metadata) error {
    defaultMetadata := defaultObjectMetadata()
    metadata = mergeMetadata(metadata, defaultMetadata)
    input := &s3.PutObjectInput{
        Bucket:   &o.Bucket,
        Key:      &o.Key,
        Metadata: metadata,
    }
    _, err := o.client.s3Client.PutObject(ctx, input)
    if err != nil {
        return wrapError(
            fmt.Errorf("creating object %s in bucket %s: %w", o.Key, o.Bucket, err),
        )
    }
    return nil
}

func (o *Object) Delete(ctx context.Context) error {
    input := &s3.DeleteObjectInput{
        Bucket: &o.Bucket,
        Key:    &o.Key,
    }
    _, err := o.client.s3Client.DeleteObject(ctx, input)
    if err != nil {
        return wrapError(
            fmt.Errorf("error deleting object %s in bucket %s: %w", o.Key, o.Bucket, err),
            objectErrCheckFuncs(o)...,
        )
    }
    return nil

}

type ObjectAttrs struct {
    Size       uint64
    ModifyTime *time.Time
    AccessTime *time.Time
    ChangeTime *time.Time
}

func (a *ObjectAttrs) toTagSlice() []types.Tag {
    var tags []types.Tag
    if a.ChangeTime != nil {
        nsec := strconv.Itoa(a.ChangeTime.Nanosecond())
        sec := strconv.Itoa(int(a.ChangeTime.Unix()))
        timeTags := []types.Tag{
            {
                Key:   aws.String(CHANGE_TIME_NSEC_METADATA_KEY),
                Value: aws.String(nsec),
            },
            {
                Key:   aws.String(CHANGE_TIME_SEC_METADATA_KEY),
                Value: aws.String(sec),
            },
        }
        tags = append(tags, timeTags...)
    }
    return tags
}

func (o *Object) GetAttrs(ctx context.Context) (*ObjectAttrs, error) {
    headOutput, err := o.head(ctx)
    if err != nil {
        return nil, err
    }

    attrs := &ObjectAttrs{
        Size:       uint64(headOutput.ContentLength),
        ModifyTime: headOutput.LastModified,
        AccessTime: aws.Time(TimeNowFunc()),
    }

    changeTimeNsecStr, okNsec := headOutput.Metadata[CHANGE_TIME_NSEC_METADATA_KEY]
    changeTimeSecStr, okSec := headOutput.Metadata[CHANGE_TIME_SEC_METADATA_KEY]
    if okNsec && okSec {
        // changeTimeStr should be a unix time integer
        changeTimeNsec, err := strconv.Atoi(changeTimeNsecStr)
        if err != nil {
            return nil, err
        }
        changeTimeSec, err := strconv.Atoi(changeTimeSecStr)
        if err != nil {
            return nil, err
        }
        changeTime := time.Unix(int64(changeTimeSec), int64(changeTimeNsec))
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
    _, err := o.client.s3Client.PutObjectTagging(ctx, input)
    if err != nil {
        return wrapError(
            fmt.Errorf("error updating object tags: %w", err),
            objectErrCheckFuncs(o)...,
        )
    }
    return nil
}

func (o *Object) Write(ctx context.Context, data io.Reader, metadata Metadata) error {
    defaultMetadata := defaultObjectMetadata()
    metadata = mergeMetadata(metadata, defaultMetadata)

    uploader := manager.NewUploader(o.client.s3Client)
    _, err := uploader.Upload(ctx, &s3.PutObjectInput{
        Bucket:   &o.Bucket,
        Key:      &o.Key,
        Body:     data,
        Metadata: metadata,
    })
    return wrapError(err, objectErrCheckFuncs(o)...)
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

    out, err := o.client.s3Client.GetObject(ctx, input)
    if err != nil {
        return nil, wrapError(
            fmt.Errorf("getting object: %w", err),
            objectErrCheckFuncs(o)...,
        )
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
    _, err := o.client.s3Client.CopyObject(ctx, input)
    if err != nil {
        return nil, wrapError(
            fmt.Errorf("error copying object %s to %s: %w", o.Key, newKey, err),
            objectErrCheckFuncs(o)...,
        )
    }

    return o.client.NewObject(o.Bucket, newKey), nil
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

func (o *Object) head(ctx context.Context) (*s3.HeadObjectOutput, error) {
    input := &s3.HeadObjectInput{
        Bucket: &o.Bucket,
        Key:    &o.Key,
    }
    out, err := o.client.s3Client.HeadObject(ctx, input)
    if err != nil {
        return nil, wrapError(err, objectErrCheckFuncs(o)...)
    }
    return out, nil
}
