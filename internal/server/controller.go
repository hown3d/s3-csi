package server

import (
    "context"
    "errors"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    aws_internal "github.com/hown3d/s3-csi/internal/aws"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

var capTypes = map[csipb.ControllerServiceCapability_RPC_Type]bool{
    csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME:         true,
    csipb.ControllerServiceCapability_RPC_GET_VOLUME:                   true,
    csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT:       true,
    csipb.ControllerServiceCapability_RPC_LIST_SNAPSHOTS:               true,
    csipb.ControllerServiceCapability_RPC_LIST_VOLUMES:                 true,
    csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME:     true,
    csipb.ControllerServiceCapability_RPC_VOLUME_CONDITION:             false,
    csipb.ControllerServiceCapability_RPC_CLONE_VOLUME:                 false,
    csipb.ControllerServiceCapability_RPC_EXPAND_VOLUME:                false,
    csipb.ControllerServiceCapability_RPC_GET_CAPACITY:                 false,
    csipb.ControllerServiceCapability_RPC_LIST_VOLUMES_PUBLISHED_NODES: false,
    csipb.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER:     false,
    csipb.ControllerServiceCapability_RPC_PUBLISH_READONLY:             false,
}

const (
    LOCATION_KEY = "csi.storage.k8s.io/bucket-location"
    IAM_ROLE_KEY = "csi.storage.k8s.io/iam-role"
)

type ControllerService struct {
    parentConfig aws.Config
}

var _ csipb.ControllerServer = (*ControllerService)(nil)

func (c ControllerService) CreateVolume(ctx context.Context, req *csipb.CreateVolumeRequest) (*csipb.CreateVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
        return nil, unimplementedError()
    }
    // Check arguments
    if len(req.GetName()) == 0 {
        return nil, status.Error(codes.InvalidArgument, "Name missing in request")
    }

    // TODO: validate capabilties
    caps := req.GetVolumeCapabilities()
    if caps == nil {
        return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
    }

    params := req.GetParameters()
    iamRole, ok := params[IAM_ROLE_KEY]
    if !ok {
        return nil, status.Error(codes.InvalidArgument, "IAM Role parameter is missing in request")
    }
    location := params[LOCATION_KEY]

    awsCfg := aws_internal.NewConfigWithRoleAssumer(c.parentConfig, iamRole, req.GetName())
    s3Client := s3_internal.NewClient(awsCfg)

    err := s3Client.CreateBucket(ctx, req.GetName(), location)
    var existsErr *types.BucketAlreadyExists
    if !errors.As(err, &existsErr) {
        // TODO: check for correct error codes
        return nil, err
    }
    return &csipb.CreateVolumeResponse{
        Volume: &csipb.Volume{
            CapacityBytes: 0,
            VolumeId:      req.GetName(),
        },
    }, nil
}

func (c ControllerService) DeleteVolume(ctx context.Context, request *csipb.DeleteVolumeRequest) (*csipb.DeleteVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ControllerPublishVolume(ctx context.Context, request *csipb.ControllerPublishVolumeRequest) (*csipb.ControllerPublishVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ControllerUnpublishVolume(ctx context.Context, request *csipb.ControllerUnpublishVolumeRequest) (*csipb.ControllerUnpublishVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ValidateVolumeCapabilities(ctx context.Context, request *csipb.ValidateVolumeCapabilitiesRequest) (*csipb.ValidateVolumeCapabilitiesResponse, error) {
    // TODO: implement
    panic("TODO: implement")
}

func (c ControllerService) ListVolumes(ctx context.Context, request *csipb.ListVolumesRequest) (*csipb.ListVolumesResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_LIST_VOLUMES) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) GetCapacity(ctx context.Context, request *csipb.GetCapacityRequest) (*csipb.GetCapacityResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_GET_CAPACITY) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ControllerGetCapabilities(ctx context.Context, request *csipb.ControllerGetCapabilitiesRequest) (*csipb.ControllerGetCapabilitiesResponse, error) {
    return &csipb.ControllerGetCapabilitiesResponse{
        Capabilities: getControllerCapabilties(),
    }, nil
}

func (c ControllerService) CreateSnapshot(ctx context.Context, request *csipb.CreateSnapshotRequest) (*csipb.CreateSnapshotResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) DeleteSnapshot(ctx context.Context, request *csipb.DeleteSnapshotRequest) (*csipb.DeleteSnapshotResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ListSnapshots(ctx context.Context, request *csipb.ListSnapshotsRequest) (*csipb.ListSnapshotsResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_LIST_SNAPSHOTS) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ControllerExpandVolume(ctx context.Context, request *csipb.ControllerExpandVolumeRequest) (*csipb.ControllerExpandVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_EXPAND_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c ControllerService) ControllerGetVolume(ctx context.Context, request *csipb.ControllerGetVolumeRequest) (*csipb.ControllerGetVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_GET_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func getControllerCapabilties() []*csipb.ControllerServiceCapability {
    var caps []*csipb.ControllerServiceCapability

    for capType, supported := range capTypes {
        if !supported {
            continue
        }
        caps = append(caps, &csipb.ControllerServiceCapability{
            Type: &csipb.ControllerServiceCapability_Rpc{
                Rpc: &csipb.ControllerServiceCapability_RPC{
                    Type: capType,
                },
            },
        })
    }
    return caps
}

func capabilityIsSupported(cap csipb.ControllerServiceCapability_RPC_Type) bool {
    supported, ok := capTypes[cap]
    if !ok {
        return false
    }
    return supported
}

func unimplementedError() error {
    return status.Error(codes.Unimplemented, "unimplemented")
}
