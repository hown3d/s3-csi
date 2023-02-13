package s3

import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
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
