package grpc

import (
    "golang.org/x/sync/errgroup"
    "google.golang.org/grpc"
    "log"
    "net"
)

type NonBlockingGrpcServer struct {
    grpcServer *grpc.Server
    addr       string
    protocol   string
    errGroup   *errgroup.Group
}

func NewNonBlockingGrpcServer(protocol string, addr string, grpcServer *grpc.Server) *NonBlockingGrpcServer {
    return &NonBlockingGrpcServer{
        grpcServer: grpcServer,
        addr:       addr,
        protocol:   protocol,
        errGroup:   new(errgroup.Group),
    }
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

func (s *NonBlockingGrpcServer) ForceStop() {
    s.grpcServer.Stop()
}

func (s *NonBlockingGrpcServer) Run() error {
    s.Start()
    return s.Wait()
}
