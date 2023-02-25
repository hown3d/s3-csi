package main

import (
    "flag"
    "github.com/hown3d/s3-csi/internal/csi"
    "log"
)

func main() {
    cfg := csi.Config{}
    flag.StringVar(&cfg.UnixSocketPath, "socket", "/tmp/csi.sock", "unix socket to bind")

    s, err := csi.NewServer(&cfg)
    if err != nil {
        log.Fatal(err)
    }
    if err := s.Run(); err != nil {
        log.Fatal(err)
    }
}
