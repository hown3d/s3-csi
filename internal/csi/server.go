package csi

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/sts"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "github.com/hown3d/s3-csi/internal/aws"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    internal_sts "github.com/hown3d/s3-csi/internal/aws/sts"
    internal_grpc "github.com/hown3d/s3-csi/internal/grpc"
    "google.golang.org/grpc"
)

type Config struct {
    UnixSocketPath string
    Protocol       string
}

func NewServer(cfg *Config) (*internal_grpc.NonBlockingGrpcServer, error) {
    awsCfg, err := aws.NewConfig(context.Background())
    if err != nil {
        return nil, fmt.Errorf("creating aws config: %w", err)
    }

    ns := &NodeService{}
    cs := NewControllerServer(internal_sts.NewAssumer(awsCfg), s3.NewClient(awsCfg))
    is := NewIdentityService(sts.NewFromConfig(awsCfg))

    if cfg.Protocol == "" {
        cfg.Protocol = "unix"
    }

    return NewGRPCServer(cfg, ns, cs, is), nil
}

func NewGRPCServer(cfg *Config, ns csipb.NodeServer, cs csipb.ControllerServer, is csipb.IdentityServer, opts ...grpc.ServerOption) *internal_grpc.NonBlockingGrpcServer {
    grpcServer := grpc.NewServer(opts...)
    csipb.RegisterNodeServer(grpcServer, ns)
    csipb.RegisterControllerServer(grpcServer, cs)
    csipb.RegisterIdentityServer(grpcServer, is)
    return internal_grpc.NewNonBlockingGrpcServer(cfg.Protocol, cfg.UnixSocketPath, grpcServer)
}
