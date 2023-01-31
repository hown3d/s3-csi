package server

import (
    "context"
    "errors"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/aws/smithy-go"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    aws_internal "github.com/hown3d/s3-csi/internal/aws"
    s3_internal "github.com/hown3d/s3-csi/internal/aws/s3"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "strings"
)

var capTypes = map[csipb.ControllerServiceCapability_RPC_Type]bool{
    csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME:   true,
    csipb.ControllerServiceCapability_RPC_GET_VOLUME:             true,
    csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT: true,
    csipb.ControllerServiceCapability_RPC_LIST_SNAPSHOTS:         true,
    csipb.ControllerServiceCapability_RPC_LIST_VOLUMES:           true,
    // publish unpublish not needed, because we don't need to do anything to the node on that step
    csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME:     false,
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

type ControllerServer struct {
    assumer  aws_internal.Assumer
    s3Client *s3_internal.Client
}

var _ csipb.ControllerServer = (*ControllerServer)(nil)

func NewControllerServer(assumer aws_internal.Assumer, client *s3_internal.Client) *ControllerServer {
    return &ControllerServer{
        assumer:  assumer,
        s3Client: client,
    }
}

func (c *ControllerServer) CreateVolume(ctx context.Context, req *csipb.CreateVolumeRequest) (*csipb.CreateVolumeResponse, error) {
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
    location := params[LOCATION_KEY]

    name := strings.ToLower(req.GetName())
    err := c.s3Client.CreateBucket(ctx, name, location)
    var existsErr *types.BucketAlreadyExists
    if !errors.As(err, &existsErr) && err != nil {
        // TODO: move all of this into own function for reusability
        var apiErr smithy.APIError
        if errors.As(err, &apiErr) {
            if apiErr.ErrorCode() == "AccessDenied" {
                return nil, status.Error(codes.Unauthenticated, apiErr.Error())
            }
        }
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &csipb.CreateVolumeResponse{
        Volume: &csipb.Volume{
            CapacityBytes: 0,
            VolumeId:      name,
        },
    }, nil
}

func (c *ControllerServer) DeleteVolume(ctx context.Context, req *csipb.DeleteVolumeRequest) (*csipb.DeleteVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
        return nil, unimplementedError()
    }

    // TODO: check volume usage (if there are fuse-servers running that use that bucket)
    err := c.s3Client.DeleteBucket(ctx, req.GetVolumeId())
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &csipb.DeleteVolumeResponse{}, nil
}

func (c *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csipb.ControllerPublishVolumeRequest) (*csipb.ControllerPublishVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerUnpublishVolume(ctx context.Context, request *csipb.ControllerUnpublishVolumeRequest) (*csipb.ControllerUnpublishVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, request *csipb.ValidateVolumeCapabilitiesRequest) (*csipb.ValidateVolumeCapabilitiesResponse, error) {
    // TODO: implement
    panic("TODO: implement")
}

func (c *ControllerServer) ListVolumes(ctx context.Context, request *csipb.ListVolumesRequest) (*csipb.ListVolumesResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_LIST_VOLUMES) {
        return nil, unimplementedError()
    }
    panic("Implement")
}

func (c *ControllerServer) GetCapacity(ctx context.Context, request *csipb.GetCapacityRequest) (*csipb.GetCapacityResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_GET_CAPACITY) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerGetCapabilities(ctx context.Context, request *csipb.ControllerGetCapabilitiesRequest) (*csipb.ControllerGetCapabilitiesResponse, error) {
    return &csipb.ControllerGetCapabilitiesResponse{
        Capabilities: getControllerCapabilties(),
    }, nil
}

func (c *ControllerServer) CreateSnapshot(ctx context.Context, request *csipb.CreateSnapshotRequest) (*csipb.CreateSnapshotResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) DeleteSnapshot(ctx context.Context, request *csipb.DeleteSnapshotRequest) (*csipb.DeleteSnapshotResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ListSnapshots(ctx context.Context, request *csipb.ListSnapshotsRequest) (*csipb.ListSnapshotsResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_LIST_SNAPSHOTS) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerExpandVolume(ctx context.Context, request *csipb.ControllerExpandVolumeRequest) (*csipb.ControllerExpandVolumeResponse, error) {
    if !capabilityIsSupported(csipb.ControllerServiceCapability_RPC_EXPAND_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerGetVolume(ctx context.Context, request *csipb.ControllerGetVolumeRequest) (*csipb.ControllerGetVolumeResponse, error) {
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

func (c *ControllerServer) s3ClientForIamRole(role, sessionName string) *s3_internal.Client {
    awsCfg := aws_internal.NewConfigWithRoleAssumer(c.assumer, role, sessionName)
    return s3_internal.NewClient(awsCfg)
}
