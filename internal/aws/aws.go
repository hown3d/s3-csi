package aws

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/smithy-go/logging"
    "log"
    "net/http"
    "os"
)

func NewConfig(ctx context.Context) (aws.Config, error) {
    return config.LoadDefaultConfig(ctx, defaultOptions()...)
}

// NewConfigWithRoleAssumer uses the assumer to assume roleArn with the sessionName.
// The credentials of the returned config will be of the assumed role
func NewConfigWithRoleAssumer(assumer Assumer, roleArn string, sessionName string) aws.Config {
    cfg := aws.Config{
        Credentials: &assumedCredsReciever{
            assumer:     assumer,
            sessionName: sessionName,
            roleArn:     roleArn,
        },
        HTTPClient: http.DefaultClient,
        Logger:     logging.StandardLogger{Logger: log.Default()},
    }
    if awsEndpoint := os.Getenv("AWS_ENDPOINT"); awsEndpoint != "" {
        cfg.EndpointResolverWithOptions = envVariableEndpointResolverFunc(awsEndpoint)
    }
    return cfg
}

func envVariableEndpointResolverFunc(endpoint string) aws.EndpointResolverWithOptionsFunc {
    return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
        return aws.Endpoint{URL: endpoint}, nil
    }
}

func defaultOptions() []func(*config.LoadOptions) error {
    var opts []func(options *config.LoadOptions) error
    if awsEndpoint := os.Getenv("AWS_ENDPOINT"); awsEndpoint != "" {
        opts = append(opts, config.WithEndpointResolverWithOptions(envVariableEndpointResolverFunc(awsEndpoint)))
    }
    return opts
}

type assumedCredsReciever struct {
    assumer     Assumer
    roleArn     string
    sessionName string
}

var _ aws.CredentialsProvider = (*assumedCredsReciever)(nil)

func (s *assumedCredsReciever) Retrieve(ctx context.Context) (aws.Credentials, error) {
    creds, err := s.assumer.AssumeRole(ctx, s.roleArn, s.sessionName)
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
