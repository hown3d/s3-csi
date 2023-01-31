package s3

import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "time"
)

type Client struct {
    s3Client *s3.Client
}

func NewClient(cfg aws.Config, opts ...func(o *s3.Options)) *Client {
    opts = append(opts, func(o *s3.Options) {
        o.UsePathStyle = true
    })
    return &Client{
        s3Client: s3.NewFromConfig(cfg, opts...),
    }
}

type Metadata map[string]string

func defaultObjectMetadata() Metadata {
    return map[string]string{
        CHANGE_TIME_METADATA_KEY: time.Now().Format(time.UnixDate),
    }
}

func (m Metadata) merge(p Metadata) {
    if p == nil {
        return
    }
    for key, val := range p {
        m[key] = val
    }
}
