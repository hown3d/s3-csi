package test

import (
	"context"
	"github.com/hown3d/s3-csi/internal/aws"
	"github.com/hown3d/s3-csi/internal/aws/s3"
	"github.com/hown3d/s3-csi/internal/server"
	"github.com/hown3d/s3-csi/test/aws/localstack"
	"github.com/hown3d/s3-csi/test/aws/sts"
	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
	g "github.com/onsi/ginkgo/v2"
	"path/filepath"
)

var localstackCon *localstack.Container
var err error

var _ = g.BeforeSuite(func() {
	localstackCon, err = localstack.New()
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
		awsCfg, err := aws.NewConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		cs := server.NewControllerServer(sts.NopAssumer{}, s3.NewClient(awsCfg))
		grpcServer, err = server.NewServerWithCustomServiceImpls(&serverConf, &server.NodeService{}, cs, &server.IdentityService{})
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			defer g.GinkgoRecover()
			if err := grpcServer.Run(); err != nil {
				t.Error(err)
			}
		}()

		config = sanity.NewTestConfig()
		// Set configuration options as needed
		config.Address = addr
		config.TargetPath = filepath.Join(t.TempDir(), "csi-mount")
		config.StagingPath = filepath.Join(t.TempDir(), "csi-staging")

	})

	g.AfterEach(func() {
		grpcServer.ForceStop()
	})

	g.Describe("CSI sanity", func() {
		sanity.GinkgoTest(&config)
	})
})

var _ = g.AfterSuite(func() {
	err := localstackCon.Delete()
	if err != nil {
		g.GinkgoT().Logf("error cleaning up localstack container: %s", err)
	}
})
