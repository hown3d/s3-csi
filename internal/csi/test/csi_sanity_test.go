package test

import (
	"path/filepath"

	"github.com/hown3d/s3-csi/internal/csi"
	"github.com/hown3d/s3-csi/test/aws/localstack"
	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
	g "github.com/onsi/ginkgo/v2"
)

var (
	localstackCon *localstack.Container
	err           error
)

var _ = g.BeforeSuite(func() {
	localstackCon, err = localstack.New()
})

var _ = g.Describe("CSI Sanity Test", func() {
	var (
		config sanity.TestConfig
		s      *csi.Server
	)

	// Setup the full driver and its environment
	g.BeforeEach(func() {
		t := g.GinkgoT()

		addr := "0.0.0.0:9090"
		// use tcp connection here because it's more stable
		serverConf := csi.Config{
			UnixSocketPath: addr,
			Protocol:       "tcp",
		}
		// awsCfg, err := aws.NewConfig(context.Background())
		// if err != nil {
		// 	t.Fatal(err)
		// }
		s, err = csi.NewServer(&serverConf)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			defer g.GinkgoRecover()
			if err := s.Run(); err != nil {
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
		s.ForceStop()
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
