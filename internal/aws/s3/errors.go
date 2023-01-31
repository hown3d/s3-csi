package s3

import "fmt"

type ErrNotFound struct {
    Key    string
    Bucket string
}

func NewErrNotFound(bucket, objName string) *ErrNotFound {
    return &ErrNotFound{
        Key:    objName,
        Bucket: bucket,
    }
}

func (e *ErrNotFound) Error() string {
    return fmt.Sprintf("Object %s not found in bucket %s", e.Key, e.Bucket)
}

var _ error = &ErrNotFound{}
