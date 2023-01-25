package main

import (
    "flag"
    "github.com/hown3d/s3-csi/internal/server"
    "log"
)

func main() {
    cfg := server.Config{}
    flag.StringVar(&cfg.UnixSocketPath, "socket", "/tmp/csi.sock", "unix socket to bind")

    if err := server.ListenAndServe(&cfg, &server.NodeService{}, &server.ControllerService{}, &server.IdentityService{}); err != nil {
        log.Fatal(err)
    }
}
