package aws

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sts"
    "github.com/aws/aws-sdk-go-v2/service/sts/types"
    "github.com/aws/smithy-go/logging"
    "log"
    "net/http"
)

func NewConfig(ctx context.Context) (aws.Config, error) {
    return config.LoadDefaultConfig(ctx)
}

// NewConfigWithRoleAssumer uses parentConfig to assume roleArn with the sessionName.
// The credentials of the returned config will be of the assumed role
func NewConfigWithRoleAssumer(parentConfig aws.Config, roleArn string, sessionName string) aws.Config {
    return aws.Config{
        Credentials: &assumedCredsReciever{
            assumer:     newAssumer(parentConfig),
            sessionName: sessionName,
            roleArn:     roleArn,
        },
        HTTPClient: http.DefaultClient,
        Logger:     logging.StandardLogger{Logger: log.Default()},
    }
}

type assumedCredsReciever struct {
    assumer     *assumer
    roleArn     string
    sessionName string
}

var _ aws.CredentialsProvider = (*assumedCredsReciever)(nil)

func (s *assumedCredsReciever) Retrieve(ctx context.Context) (aws.Credentials, error) {
    creds, err := s.assumer.assumeRole(ctx, s.roleArn, s.sessionName)
    if err != nil {
        return aws.Credentials{}, err
    }
    return aws.Credentials{
        AccessKeyID:     *creds.AccessKeyId,
        SecretAccessKey: *creds.SecretAccessKey,
        SessionToken:    *creds.SessionToken,
        CanExpire:       true,
        Expires:         *creds.Expiration,
    }, nil
}

type assumer struct {
    stsClient *sts.Client
}

func newAssumer(cfg aws.Config) *assumer {
    return &assumer{
        stsClient: sts.NewFromConfig(cfg),
    }
}

func (a *assumer) assumeRole(ctx context.Context, roleArn string, sessionName string) (*types.Credentials, error) {
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
