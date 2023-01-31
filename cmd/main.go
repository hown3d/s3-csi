package main

import (
    "flag"
    "github.com/hown3d/s3-csi/internal/server"
    "log"
)

func main() {
    cfg := server.Config{}
    flag.StringVar(&cfg.UnixSocketPath, "socket", "/tmp/csi.sock", "unix socket to bind")

    s, err := server.NewServer(&cfg)
    if err != nil {
        log.Fatal(err)
    }
    if err := s.Run(); err != nil {
        log.Fatal(err)
    }
}
