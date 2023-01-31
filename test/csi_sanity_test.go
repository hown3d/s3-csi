package test

import (
    "errors"
    "fmt"
    "github.com/hown3d/s3-csi/internal/server"
    "github.com/hown3d/s3-csi/test/aws/sts"
    "github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
    g "github.com/onsi/ginkgo/v2"
    "github.com/ory/dockertest/v3"
    "net/http"
    "os"
    "path/filepath"
)

const LOCALSTACK_PORT string = "4566/tcp"

func startLocalStackContainer(t g.GinkgoTInterface) (container *Container) {
    healthCheckFunc := func(resource *dockertest.Resource) error {
        var err error
        resp, err := http.Get(fmt.Sprintf("http://localhost:%s/_localstack/health", resource.GetPort(LOCALSTACK_PORT)))
        if err != nil {
            return err
        }
        if resp.StatusCode != http.StatusOK {
            return errors.New("not healthy")
        }
        return nil
    }
    runOpts := dockertest.RunOptions{
        Repository:   "localstack/localstack",
        Tag:          "latest",
        ExposedPorts: []string{LOCALSTACK_PORT},
        Env:          []string{"PROVIDER_OVERRIDE_S3=asf", "DEBUG=1"},
    }
    con, err := NewContainer(&runOpts, healthCheckFunc)
    if err != nil {
        t.Fatal(err)
    }

    t.Log("localstack is running")
    return con
}

var localstackContainer *Container

var _ = g.BeforeSuite(func() {
    t := g.GinkgoT()
    localstackContainer = startLocalStackContainer(t)
    port := localstackContainer.resource.GetPort(LOCALSTACK_PORT)
    if err := os.Setenv("AWS_ENDPOINT", fmt.Sprintf("http://localhost:%s", port)); err != nil {
        t.Fatal(err)
    }
    if err := os.Setenv("AWS_REGION", "eu-central-1"); err != nil {
        t.Fatal(err)
    }
    if err := os.Setenv("AWS_ACCESS_KEY_ID", "test"); err != nil {
        t.Fatal(err)
    }
    if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "test"); err != nil {
        t.Fatal(err)
    }
})

var _ = g.Describe("CSI Sanity Test", func() {
    var (
        config     sanity.TestConfig
        grpcServer *server.NonBlockingGrpcServer
    )

    // Setup the full driver and its environment
    g.BeforeEach(func() {
        t := g.GinkgoT()

        addr := "0.0.0.0:9090"
        // use tcp connection here because it's more stable
        serverConf := server.Config{
            UnixSocketPath: addr,
            Protocol:       "tcp",
        }
        cs, err := server.NewControllerServer(sts.NopAssumer{})
        grpcServer, err = server.NewServerWithCustomServiceImpls(&serverConf, &server.NodeService{}, cs, &server.IdentityService{})
        if err != nil {
            t.Fatal(err)
        }
        go func() {
            if err := grpcServer.Run(); err != nil {
                t.Error(err)
            }
        }()

        config = sanity.NewTestConfig()
        // Set configuration options as needed
        config.Address = addr
        config.TestVolumeParameters = map[string]string{
            server.IAM_ROLE_KEY: "test123",
        }
        config.TargetPath = filepath.Join(t.TempDir(), "csi-mount")
        config.StagingPath = filepath.Join(t.TempDir(), "csi-staging")

    })

    g.AfterEach(func() {
        grpcServer.Stop()
    })

    g.Describe("CSI sanity", func() {
        sanity.GinkgoTest(&config)
    })
})

var _ = g.AfterSuite(func() {
    err := localstackContainer.Delete()
    if err != nil {
        g.GinkgoT().Logf("error cleaning up localstack container: %s", err)
    }
})
