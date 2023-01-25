package server

import (
    "context"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
)

type NodeService struct{}

var _ csipb.NodeServer = (*NodeService)(nil)

func (n *NodeService) NodeStageVolume(ctx context.Context, request *csipb.NodeStageVolumeRequest) (*csipb.NodeStageVolumeResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodeUnstageVolume(ctx context.Context, request *csipb.NodeUnstageVolumeRequest) (*csipb.NodeUnstageVolumeResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodePublishVolume(ctx context.Context, request *csipb.NodePublishVolumeRequest) (*csipb.NodePublishVolumeResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodeUnpublishVolume(ctx context.Context, request *csipb.NodeUnpublishVolumeRequest) (*csipb.NodeUnpublishVolumeResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodeGetVolumeStats(ctx context.Context, request *csipb.NodeGetVolumeStatsRequest) (*csipb.NodeGetVolumeStatsResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodeExpandVolume(ctx context.Context, request *csipb.NodeExpandVolumeRequest) (*csipb.NodeExpandVolumeResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodeGetCapabilities(ctx context.Context, request *csipb.NodeGetCapabilitiesRequest) (*csipb.NodeGetCapabilitiesResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (n *NodeService) NodeGetInfo(ctx context.Context, request *csipb.NodeGetInfoRequest) (*csipb.NodeGetInfoResponse, error) {
    //TODO implement me
    panic("implement me")
}
