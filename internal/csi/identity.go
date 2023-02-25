package server

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/sts"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type IdentityService struct {
    stsClient *sts.Client
}

func NewIdentityService(stsClient *sts.Client) *IdentityService {
    return &IdentityService{
        stsClient: stsClient,
    }
}

var _ csipb.IdentityServer = (*IdentityService)(nil)

func (i *IdentityService) GetPluginInfo(ctx context.Context, request *csipb.GetPluginInfoRequest) (*csipb.GetPluginInfoResponse, error) {
    return &csipb.GetPluginInfoResponse{
        Name:          "hown3d.s3-csi",
        VendorVersion: "v1alpha1",
    }, nil
}

func (i *IdentityService) GetPluginCapabilities(ctx context.Context, request *csipb.GetPluginCapabilitiesRequest) (*csipb.GetPluginCapabilitiesResponse, error) {
    caps := []*csipb.PluginCapability{
        {
            Type: &csipb.PluginCapability_Service_{
                Service: &csipb.PluginCapability_Service{
                    Type: csipb.PluginCapability_Service_CONTROLLER_SERVICE,
                },
            },
        },
    }
    return &csipb.GetPluginCapabilitiesResponse{
        Capabilities: caps,
    }, nil
}

func (i *IdentityService) Probe(ctx context.Context, request *csipb.ProbeRequest) (*csipb.ProbeResponse, error) {
    if i.isHealthy(ctx) {
        return &csipb.ProbeResponse{}, nil
    }
    return nil, status.Error(codes.FailedPrecondition, "driver is unhealthy")
}

func (i *IdentityService) isHealthy(ctx context.Context) bool {
    _, err := i.stsClient.GetCallerIdentity(ctx, nil)
    if err != nil {
        return false
    }
    return true
}
