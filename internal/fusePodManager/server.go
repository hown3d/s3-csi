package fusePodManager

import (
    "fmt"
    internal_grpc "github.com/hown3d/s3-csi/internal/grpc"
    fuse_pod_managerv1alpha1 "github.com/hown3d/s3-csi/proto/gen/fuse_pod_manager/v1alpha1"
    "google.golang.org/grpc"
)

func NewServer(port int, opts ...grpc.ServerOption) *internal_grpc.NonBlockingGrpcServer {
    grpcServer := grpc.NewServer(opts...)
    fuse_pod_managerv1alpha1.RegisterFusePodManagerServiceServer(grpcServer, &fusePodManagerService{})
    addr := fmt.Sprintf("0.0.0.0:%d", port)
    return internal_grpc.NewNonBlockingGrpcServer("tcp", addr, grpcServer)
}
