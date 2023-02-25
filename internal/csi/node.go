package csi

import (
    "context"
    "fmt"
    "log"
    "os"

    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

const NODE_ID_ENV_KEY = "KUBE_NODE_NAME"

type NodeService struct {
    s3Client *s3.Client
}

var nodeCapTypes = capMap[csipb.NodeServiceCapability_RPC_Type]{
    csipb.NodeServiceCapability_RPC_VOLUME_CONDITION:         true,
    csipb.NodeServiceCapability_RPC_GET_VOLUME_STATS:         true,
    csipb.NodeServiceCapability_RPC_EXPAND_VOLUME:            false,
    csipb.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER: true,
    csipb.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME:     true,
    csipb.NodeServiceCapability_RPC_VOLUME_MOUNT_GROUP:       true,
}

var _ csipb.NodeServer = (*NodeService)(nil)

func (n *NodeService) NodeStageVolume(ctx context.Context, request *csipb.NodeStageVolumeRequest) (*csipb.NodeStageVolumeResponse, error) {
    if !nodeCapTypes.isSupported(csipb.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME) {
        return nil, unimplementedError()
    }
    panic(fmt.Sprintf("%#v is supported, but not implemented", csipb.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME))
}

func (n *NodeService) NodeUnstageVolume(ctx context.Context, request *csipb.NodeUnstageVolumeRequest) (*csipb.NodeUnstageVolumeResponse, error) {
    if !nodeCapTypes.isSupported(csipb.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME) {
        return nil, unimplementedError()
    }
    panic(fmt.Sprintf("%#v is supported, but not implemented", csipb.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME))
}

func (n *NodeService) NodePublishVolume(ctx context.Context, req *csipb.NodePublishVolumeRequest) (*csipb.NodePublishVolumeResponse, error) {
    // Check arguments
    if req.GetVolumeCapability() == nil {
        return nil, status.Error(codes.InvalidArgument, "Volume capType missing in request")
    }
    if len(req.GetVolumeId()) == 0 {
        return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
    }
    if len(req.GetStagingTargetPath()) == 0 {
        return nil, status.Error(codes.InvalidArgument, "Staging Target path missing in request")
    }
    if len(req.GetTargetPath()) == 0 {
        return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
    }

    name := sanitizeVolumeID(req.VolumeId)
    bucket := n.s3Client.NewBucket(name)
    if !bucket.Exists(ctx) {
        return nil, status.Errorf(codes.NotFound, "bucket %s for volume %s does not exist", name, req.VolumeId)
    }

    panic("implement me")
}

func (n *NodeService) NodeUnpublishVolume(ctx context.Context, request *csipb.NodeUnpublishVolumeRequest) (*csipb.NodeUnpublishVolumeResponse, error) {
    panic("implement me")
}

func (n *NodeService) NodeGetVolumeStats(ctx context.Context, request *csipb.NodeGetVolumeStatsRequest) (*csipb.NodeGetVolumeStatsResponse, error) {
    if !nodeCapTypes.isSupported(csipb.NodeServiceCapability_RPC_GET_VOLUME_STATS) {
        return nil, unimplementedError()
    }
    log.Printf("%#v is supported, but not implemented", csipb.NodeServiceCapability_RPC_GET_VOLUME_STATS)
    return nil, unimplementedError()
}

func (n *NodeService) NodeExpandVolume(ctx context.Context, request *csipb.NodeExpandVolumeRequest) (*csipb.NodeExpandVolumeResponse, error) {
    if !nodeCapTypes.isSupported(csipb.NodeServiceCapability_RPC_EXPAND_VOLUME) {
        return nil, unimplementedError()
    }
    log.Printf("%#v is supported, but not implemented", csipb.NodeServiceCapability_RPC_EXPAND_VOLUME)
    return nil, unimplementedError()
}

func (n *NodeService) NodeGetCapabilities(ctx context.Context, request *csipb.NodeGetCapabilitiesRequest) (*csipb.NodeGetCapabilitiesResponse, error) {
    return &csipb.NodeGetCapabilitiesResponse{Capabilities: getNodeCapabilities()}, nil
}

func (n *NodeService) NodeGetInfo(ctx context.Context, request *csipb.NodeGetInfoRequest) (*csipb.NodeGetInfoResponse, error) {
    nodeId := os.Getenv(NODE_ID_ENV_KEY)
    if nodeId == "" {
        msg := fmt.Sprintf("%s needs to be set", NODE_ID_ENV_KEY)
        return nil, status.Error(codes.FailedPrecondition, msg)
    }
    return &csipb.NodeGetInfoResponse{
        NodeId: nodeId,
    }, nil
}

func getNodeCapabilities() []*csipb.NodeServiceCapability {
    var caps []*csipb.NodeServiceCapability

    for capType, supported := range nodeCapTypes {
        if !supported {
            continue
        }
        caps = append(caps, &csipb.NodeServiceCapability{
            Type: &csipb.NodeServiceCapability_Rpc{
                Rpc: &csipb.NodeServiceCapability_RPC{
                    Type: capType,
                },
            },
        })
    }
    return caps
}
