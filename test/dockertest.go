package test

import (
    "context"
    "fmt"
    "github.com/ory/dockertest/v3"
    "github.com/ory/dockertest/v3/docker"
    "io"
)

type Container struct {
    pool     *dockertest.Pool
    resource *dockertest.Resource
}

func NewContainer(opts *dockertest.RunOptions, healthCheckFunc func(resource *dockertest.Resource) error) (*Container, error) {
    // uses a sensible default on windows (tcp/http) and linux/osx (socket)
    p, err := dockertest.NewPool("")
    if err != nil {
        return nil, err
    }

    r, err := p.RunWithOptions(opts)
    if err != nil {
        return nil, err
    }

    if healthCheckFunc != nil {
        if err := p.Retry(func() error {
            return healthCheckFunc(r)
        }); err != nil {
            return nil, fmt.Errorf("container didn't pass health check function: %w", err)
        }
    }

    return &Container{
        pool:     p,
        resource: r,
    }, nil
}

func (c *Container) TailLogs(ctx context.Context, wr io.Writer, follow bool) error {
    opts := docker.LogsOptions{
        Context: ctx,

        Stderr:      true,
        Stdout:      true,
        Follow:      follow,
        Timestamps:  true,
        RawTerminal: true,

        Container: c.resource.Container.ID,

        OutputStream: wr,
    }

    return c.pool.Client.Logs(opts)
}

func (c *Container) Delete() error {
    return c.pool.Purge(c.resource)
}
