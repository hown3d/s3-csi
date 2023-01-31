package dockertest

import (
    "bufio"
    "bytes"
    "context"
    "errors"
    "fmt"
    "github.com/ory/dockertest/v3"
    "github.com/ory/dockertest/v3/docker"
    "io"
    "net/http"
)

type Container struct {
    *dockertest.Resource

    pool *dockertest.Pool
}

type HealthCheckFunc func(resource *dockertest.Resource, pool *dockertest.Pool) error

func LogHealthcheck(logLine string) HealthCheckFunc {
    return func(resource *dockertest.Resource, pool *dockertest.Pool) error {
        buf := new(bytes.Buffer)
        scanner := bufio.NewScanner(buf)

        logOpts := docker.LogsOptions{
            Context:      context.Background(),
            Container:    resource.Container.ID,
            OutputStream: buf,
            ErrorStream:  buf,
            Stdout:       true,
            Stderr:       true,
        }
        err := pool.Client.Logs(logOpts)
        if err != nil {
            return err
        }
        for scanner.Scan() {
            if scanner.Text() == logLine {
                return nil
            }
        }
        return errors.New("log line not discovered")
    }
}

func HttpHealthcheck(portIdentifier string, endpoint string, healthyStatusCode int) HealthCheckFunc {
    return func(resource *dockertest.Resource, pool *dockertest.Pool) error {
        resp, err := http.Get(fmt.Sprintf("http://localhost:%s/%s", resource.GetPort(portIdentifier), endpoint))
        if err != nil {
            return err
        }
        if resp.StatusCode != healthyStatusCode {
            return errors.New("not healthy")
        }
        return nil
    }
}

func NewContainer(opts *dockertest.RunOptions, healthCheckFunc HealthCheckFunc) (*Container, error) {
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
            return healthCheckFunc(r, p)
        }); err != nil {
            return nil, fmt.Errorf("container didn't pass health check function: %w", err)
        }
    }

    return &Container{r, p}, nil
}

func (c *Container) TailLogs(ctx context.Context, wr io.Writer, follow bool) error {
    opts := docker.LogsOptions{
        Context: ctx,

        Stderr:      true,
        Stdout:      true,
        Follow:      follow,
        Timestamps:  true,
        RawTerminal: true,

        Container: c.Container.ID,

        OutputStream: wr,
    }

    return c.pool.Client.Logs(opts)
}

func (c *Container) Delete() error {
    return c.pool.Purge(c.Resource)
}
