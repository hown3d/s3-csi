package fusePodManager

import (
    "fmt"
    internal_grpc "github.com/hown3d/s3-csi/internal/grpc"
    "github.com/hown3d/s3-csi/internal/k8s"
    pb "github.com/hown3d/s3-csi/proto/gen/fuse_pod_manager/v1alpha1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

func NewServer(port int, namespace string, opts ...grpc.ServerOption) (*internal_grpc.NonBlockingGrpcServer, error) {
    grpcServer := grpc.NewServer(opts...)

    clientset, err := k8s.NewClientSet("")
    if err != nil {
        return nil, err
    }

    podClient := clientset.CoreV1().Pods(namespace)

    service := &fusePodManagerService{podClient: podClient}
    pb.RegisterFusePodManagerServiceServer(grpcServer, service)
    // TODO: enable reflection based on configuration
    reflection.Register(grpcServer)
    addr := fmt.Sprintf("0.0.0.0:%d", port)
    return internal_grpc.NewNonBlockingGrpcServer("tcp", addr, grpcServer), nil
}
