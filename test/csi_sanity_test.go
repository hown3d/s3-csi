package test

import (
    "github.com/hown3d/s3-csi/internal/server"
    "github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
    "log"
    "path/filepath"
    "testing"
)

func TestDriver(t *testing.T) {
    // Setup the full driver and its environment
    addr := filepath.Join(t.TempDir(), "csi.sock")
    serverConf := server.Config{
        UnixSocketPath: addr,
    }

    go func() {
        if err := server.ListenAndServe(&serverConf, &server.NodeService{}, &server.ControllerService{}, &server.IdentityService{}); err != nil {
            log.Fatal(err)
        }
    }()

    config := sanity.NewTestConfig()
    // Set configuration options as needed
    config.Address = addr
    config.TestVolumeParameters = map[string]string{
        server.IAM_ROLE_KEY: "test123",
    }

    // Now call the test suite
    sanity.Test(t, config)
}
