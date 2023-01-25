package server

import (
    "context"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type IdentityService struct{}

var _ csipb.IdentityServer = (*IdentityService)(nil)

func (i IdentityService) GetPluginInfo(ctx context.Context, request *csipb.GetPluginInfoRequest) (*csipb.GetPluginInfoResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (i IdentityService) GetPluginCapabilities(ctx context.Context, request *csipb.GetPluginCapabilitiesRequest) (*csipb.GetPluginCapabilitiesResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (i IdentityService) Probe(ctx context.Context, request *csipb.ProbeRequest) (*csipb.ProbeResponse, error) {
    if isHealthy() {
        return &csipb.ProbeResponse{}, nil
    }
    return nil, status.Error(codes.FailedPrecondition, "driver is unhealthy")
}

func isHealthy() bool {
    return true
}
