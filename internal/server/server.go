package server

import (
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "google.golang.org/grpc"
    "log"
    "net"
)

type Config struct {
    UnixSocketPath string
}

func ListenAndServe(cfg *Config, ns csipb.NodeServer, cs csipb.ControllerServer, is csipb.IdentityServer) error {
    lis, err := net.Listen("unix", cfg.UnixSocketPath)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    csipb.RegisterNodeServer(grpcServer, ns)
    csipb.RegisterControllerServer(grpcServer, cs)
    csipb.RegisterIdentityServer(grpcServer, is)
    return grpcServer.Serve(lis)
}
