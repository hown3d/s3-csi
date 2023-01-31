package aws

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type Assumer interface {
    AssumeRole(ctx context.Context, roleArn string, sessionName string) (*types.Credentials, error)
}
