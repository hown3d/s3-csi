package server

import (
    "context"
    "fmt"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "github.com/hown3d/s3-csi/internal/aws"
    "github.com/hown3d/s3-csi/internal/aws/sts"
    "golang.org/x/sync/errgroup"
    "google.golang.org/grpc"
    "log"
    "net"
)

type Config struct {
    UnixSocketPath string
    Protocol       string
}

type NonBlockingGrpcServer struct {
    grpcServer *grpc.Server
    addr       string
    protocol   string
    errGroup   *errgroup.Group
}

func NewServer(cfg *Config) (*NonBlockingGrpcServer, error) {
    awsCfg, err := aws.NewConfig(context.Background())
    if err != nil {
        return nil, fmt.Errorf("creating aws config: %w", err)
    }

    ns := &NodeService{}
    cs, err := NewControllerServer(sts.NewAssumer(awsCfg))
    if err != nil {
        return nil, fmt.Errorf("creating controller server: %w", err)
    }
    is := &IdentityService{}

    return NewServerWithCustomServiceImpls(cfg, ns, cs, is)
}

func NewServerWithCustomServiceImpls(cfg *Config, ns csipb.NodeServer, cs csipb.ControllerServer, is csipb.IdentityServer, opts ...grpc.ServerOption) (*NonBlockingGrpcServer, error) {
    grpcServer := grpc.NewServer(opts...)
    csipb.RegisterNodeServer(grpcServer, ns)
    csipb.RegisterControllerServer(grpcServer, cs)
    csipb.RegisterIdentityServer(grpcServer, is)
    s := &NonBlockingGrpcServer{
        grpcServer: grpcServer,
        errGroup:   new(errgroup.Group),
        addr:       cfg.UnixSocketPath,
        protocol:   cfg.Protocol,
    }
    if cfg.Protocol == "" {
        s.protocol = "unix"
    }
    return s, nil
}

func (s *NonBlockingGrpcServer) serve() error {
    lis, err := net.Listen(s.protocol, s.addr)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    return s.grpcServer.Serve(lis)
}

func (s *NonBlockingGrpcServer) Start() {
    s.errGroup.Go(s.serve)
}

func (s *NonBlockingGrpcServer) Wait() error {
    return s.errGroup.Wait()
}

func (s *NonBlockingGrpcServer) Stop() {
    s.grpcServer.GracefulStop()
}

func (s *NonBlockingGrpcServer) Run() error {
    s.Start()
    return s.Wait()
}
