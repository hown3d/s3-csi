package s3

import (
	"github.com/minio/minio-go/v7"
)

type Client struct {
	s3Client *minio.Client
}

func NewClient(endpoint string, opts *minio.Options) (*Client, error) {
	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}
	return &Client{
		s3Client: client,
	}, nil
}
