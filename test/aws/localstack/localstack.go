package localstack

import (
    "fmt"
    internal_dockertest "github.com/hown3d/s3-csi/test/dockertest"
    "github.com/ory/dockertest/v3"
    "net/http"
    "os"
)

const LOCALSTACK_PORT string = "4566/tcp"

type Container struct {
    con *internal_dockertest.Container
}

func New() (*Container, error) {
    healthCheckFunc := internal_dockertest.HttpHealthcheck(LOCALSTACK_PORT, "_localstack/health", http.StatusOK)
    runOpts := dockertest.RunOptions{
        Repository:   "localstack/localstack",
        Tag:          "latest",
        ExposedPorts: []string{LOCALSTACK_PORT},
        Env:          []string{"PROVIDER_OVERRIDE_S3=asf", "DEBUG=1"},
    }
    con, err := internal_dockertest.NewContainer(&runOpts, healthCheckFunc)
    if err != nil {
        return nil, err
    }

    return &Container{con: con}, err
}

func (c *Container) SetAWSEnvVariables() error {
    port := c.con.GetPort(LOCALSTACK_PORT)
    if err := os.Setenv("AWS_ENDPOINT", fmt.Sprintf("http://localhost:%s", port)); err != nil {
        return err
    }
    if err := os.Setenv("AWS_REGION", "eu-central-1"); err != nil {
        return err
    }
    if err := os.Setenv("AWS_ACCESS_KEY_ID", "test"); err != nil {
        return err
    }
    if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "test"); err != nil {
        return err
    }
    return nil
}

func (c *Container) Delete() error {
    return c.con.Delete()
}
