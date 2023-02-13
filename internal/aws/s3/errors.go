package s3

import (
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/aws/smithy-go"
)

type ErrObjectNotFound struct {
    Key    string
    Bucket string
}

func (e *ErrObjectNotFound) Error() string {
    return fmt.Sprintf("Object %s not found in bucket %s", e.Key, e.Bucket)
}

type ErrBucketNotFound struct {
    Bucket string
}

func (e *ErrBucketNotFound) Error() string {
    return fmt.Sprintf("Bucket %s not found", e.Bucket)
}

type ErrAccessDenied struct {
    Message string
}

func (e *ErrAccessDenied) Error() string {
    return fmt.Sprintf("access denied: %s", e.Message)
}

var ErrPageOutOfBounds = errors.New("Start is bigger then slice length")

// wrapError checks if the provided error matches one of the errCheck funcs
// and override the error type is something matches
func wrapError(err error, wrapFuncs ...ErrCheckFunc) error {
    for _, f := range wrapFuncs {
        err = f(err)
    }
    return err
}

// ErrCheckFunc provides functionality to check if error matches a certain condition
// If error matches, a custom error should be returned. If not, the provided error is returned
type ErrCheckFunc func(err error) error

func objectErrCheckFuncs(o *Object) []ErrCheckFunc {
    return []ErrCheckFunc{
        accessDeniedErrCheckFunc,
        bucketNotFoundErrCheckFunc(o.Bucket),
        objectNotFoundErrCheckFunc(o.Bucket, o.Key),
    }
}

func bucketErrCheckFuncs(b *Bucket) []ErrCheckFunc {
    return []ErrCheckFunc{
        accessDeniedErrCheckFunc,
        bucketNotFoundErrCheckFunc(b.Name),
    }
}

func accessDeniedErrCheckFunc(err error) error {
    var apiErr smithy.APIError
    if errors.As(err, &apiErr) {
        if apiErr.ErrorCode() == "AccessDenied" {
            return &ErrAccessDenied{
                Message: apiErr.ErrorMessage(),
            }
        }
    }
    return err
}

func bucketNotFoundErrCheckFunc(bucketName string) func(err error) error {
    return func(err error) error {
        var noSuchBucketErr *types.NoSuchBucket
        if errors.As(err, &noSuchBucketErr) {
            return &ErrBucketNotFound{Bucket: bucketName}
        }

        // only HeadBucket implements deserialization into types.NotFound
        var apiErr *types.NotFound
        if errors.As(err, &apiErr) {
            return &ErrBucketNotFound{Bucket: bucketName}
        }
        return err
    }
}

func objectNotFoundErrCheckFunc(bucketName string, key string) func(err error) error {
    return func(err error) error {
        var noSuchKeyErr *types.NoSuchKey
        if errors.As(err, &noSuchKeyErr) {
            return &ErrObjectNotFound{
                Bucket: bucketName,
                Key:    key,
            }
        }
        var apiErr smithy.APIError
        if errors.As(err, &apiErr) {
            if apiErr.ErrorCode() == "NotFound" {
                return &ErrObjectNotFound{
                    Bucket: bucketName,
                    Key:    key,
                }
            }
        }
        return err
    }
}
