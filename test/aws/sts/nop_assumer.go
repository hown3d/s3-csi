package sts

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sts/types"
    "time"
)

type NopAssumer struct{}

func (n NopAssumer) AssumeRole(ctx context.Context, roleArn string, sessionName string) (*types.Credentials, error) {
    expire := time.Now().Add(time.Hour * 1)
    return &types.Credentials{
        AccessKeyId:     aws.String("nop"),
        Expiration:      &expire,
        SecretAccessKey: aws.String("nop"),
        SessionToken:    aws.String("nop"),
    }, nil
}
