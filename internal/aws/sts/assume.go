package sts

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sts"
    "github.com/aws/aws-sdk-go-v2/service/sts/types"
    internal_aws "github.com/hown3d/s3-csi/internal/aws"
)

type Assumer struct {
    stsClient *sts.Client
}

var _ internal_aws.Assumer = (*Assumer)(nil)

func NewAssumer(cfg aws.Config) *Assumer {
    return &Assumer{
        stsClient: sts.NewFromConfig(cfg),
    }
}

func (a *Assumer) AssumeRole(ctx context.Context, roleArn string, sessionName string) (*types.Credentials, error) {
    input := &sts.AssumeRoleInput{
        RoleArn:         &roleArn,
        RoleSessionName: &sessionName,
    }
    out, err := a.stsClient.AssumeRole(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("error assuming role: %s: %w", roleArn, err)
    }
    return out.Credentials, nil
}
