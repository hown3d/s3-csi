package csi

import (
	csipb "github.com/container-storage-interface/spec/lib/go/csi"
	internal_grpc "github.com/hown3d/s3-csi/internal/grpc"
	"google.golang.org/grpc"
)

type Server struct {
	*internal_grpc.NonBlockingGrpcServer
}

type Config struct {
	UnixSocketPath string
	Protocol       string
}

func NewServer(cfg *Config) (*Server, error) {
	ns := &NodeService{}
	cs := &ControllerServer{}
	is := &IdentityService{}

	if cfg.Protocol == "" {
		cfg.Protocol = "unix"
	}

	return &Server{
		NonBlockingGrpcServer: newGRPCServer(cfg, ns, cs, is),
	}, nil
}

func newGRPCServer(cfg *Config, ns csipb.NodeServer, cs csipb.ControllerServer, is csipb.IdentityServer, opts ...grpc.ServerOption) *internal_grpc.NonBlockingGrpcServer {
	grpcServer := grpc.NewServer(opts...)
	csipb.RegisterNodeServer(grpcServer, ns)
	csipb.RegisterControllerServer(grpcServer, cs)
	csipb.RegisterIdentityServer(grpcServer, is)
	return internal_grpc.NewNonBlockingGrpcServer(cfg.Protocol, cfg.UnixSocketPath, grpcServer)
}
