package s3

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type PageOpts struct {
    Start int
    // Size can be used to set the Size of pages
    // if 0, will be interpreted to return everything till the end
    Size int
}

func slice[T any](slice []T, p PageOpts) ([]T, error) {
    if p.Start > len(slice) {
        return nil, ErrPageOutOfBounds
    }
    if p.Size != 0 {
        end := p.Start + p.Size
        if end <= len(slice) {
            return slice[p.Start:end], nil
        }
    }
    return slice[p.Start:], nil
}
func (c *Client) ListBuckets(ctx context.Context, pageOpts *PageOpts) ([]*Bucket, error) {
    out, err := c.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
    if err != nil {
        return nil, wrapError(err, accessDeniedErrCheckFunc)
    }
    allBuckets := out.Buckets
    if pageOpts != nil {
        allBuckets, err = slice(allBuckets, *pageOpts)
        if err != nil {
            return nil, fmt.Errorf("error slicing buckets with PageOpts: %#v: %w", pageOpts, err)
        }
    }

    buckets := make([]*Bucket, 0, len(allBuckets))
    for _, bucket := range allBuckets {
        newBucket := c.NewBucket(*bucket.Name)
        buckets = append(buckets, newBucket)
    }
    return buckets, nil
}
