package s3

import "fmt"

type ErrObjectNotFound struct {
    Key    string
    Bucket string
}

func NewErrObjectNotFound(bucket, objName string) *ErrObjectNotFound {
    return &ErrObjectNotFound{
        Key:    objName,
        Bucket: bucket,
    }
}

func (e *ErrObjectNotFound) Error() string {
    return fmt.Sprintf("Object %s not found in bucket %s", e.Key, e.Bucket)
}

var _ error = &ErrObjectNotFound{}

type ErrBucketNotFound struct {
    Bucket string
}

func NewErrBucketNotFound(bucket string) *ErrBucketNotFound {
    return &ErrBucketNotFound{
        Bucket: bucket,
    }
}

func (e *ErrBucketNotFound) Error() string {
    return fmt.Sprintf("Bucket %s not found", e.Bucket)
}

var _ error = &ErrBucketNotFound{}
