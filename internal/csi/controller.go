package server

import (
    "context"
    "errors"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    csipb "github.com/container-storage-interface/spec/lib/go/csi"
    "github.com/hown3d/s3-csi/internal/aws"
    "github.com/hown3d/s3-csi/internal/aws/s3"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

var controllerCaps = capMap[csipb.ControllerServiceCapability_RPC_Type]{
    csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME: true,
    csipb.ControllerServiceCapability_RPC_GET_VOLUME:           true,
    // not implemented for now
    // could be implemented by copying buckets
    // would need to implement a good and fast copy strategy
    csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT:   false,
    csipb.ControllerServiceCapability_RPC_LIST_SNAPSHOTS:           true,
    csipb.ControllerServiceCapability_RPC_LIST_VOLUMES:             true,
    csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME: true,
    csipb.ControllerServiceCapability_RPC_VOLUME_CONDITION:         false,
    // could be implemented by copying buckets
    // would need to implement a good strategy
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

    // VOLUME_IDENTIFIER_TAG_KEY_PREFIX is used on s3 buckets to identify
    BUCKET_CONTROLLED_BY_CSI_KEY   = "csi.storage.k8s.io/controlled-by"
    BUCKET_CONTROLLED_BY_CSI_VALUE = "s3-csi"
    // BUCKET_VOLUME_NAME_KEY is used as a key in the volumecontext add the bucket name
    BUCKET_VOLUME_NAME_KEY = "csi.storage.k8s.io/bucket-name"

    // BUCKET_VOLUME_ID_KEY is used in the bucket metadata to persit the volumeid
    BUCKET_VOLUME_ID_KEY = "csi.storage.k8s.io/volume-id"
)

type ControllerServer struct {
    assumer aws.Assumer
    //s3Client *s3.Client
}

var _ csipb.ControllerServer = (*ControllerServer)(nil)

func NewControllerServer(assumer aws.Assumer, client *s3.Client) *ControllerServer {
    return &ControllerServer{
        assumer: assumer,
        //s3Client: client,
    }
}

func (c *ControllerServer) CreateVolume(ctx context.Context, req *csipb.CreateVolumeRequest) (*csipb.CreateVolumeResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
        return nil, unimplementedError()
    }
    // Check arguments
    if len(req.Name) == 0 {
        return nil, status.Error(codes.InvalidArgument, "Name missing in request")
    }

    // TODO: validate capabilties
    caps := req.GetVolumeCapabilities()
    if caps == nil {
        return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
    }

    secrets := req.GetSecrets()
    s3Client, err := c.s3ClientFromSecrets(secrets, fmt.Sprintf("CreateVolume-%s", req.Name))
    if err != nil {
        return nil, err
    }

    name := sanitizeVolumeID(req.Name)
    bucket := s3Client.NewBucket(name)

    params := req.GetParameters()
    location := params[LOCATION_KEY]
    err = bucket.Create(ctx, location)
    var existsErr *types.BucketAlreadyExists
    if !errors.As(err, &existsErr) && err != nil {
        var errAccessDenied *s3.ErrAccessDenied
        if errors.As(err, &errAccessDenied) {
            return nil, status.Errorf(codes.Unauthenticated, "Access Denied. Please ensure you have the right AWS permissions: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "Failed create bucket %s for volume id %s: %s", bucket.Name, req.Name, err)
    }

    metadata := s3.Metadata{
        BUCKET_CONTROLLED_BY_CSI_KEY: BUCKET_CONTROLLED_BY_CSI_VALUE,
        BUCKET_VOLUME_ID_KEY:         req.Name,
    }

    err = bucket.AddMetadata(ctx, metadata)
    if err != nil {
        var errAccessDenied *s3.ErrAccessDenied
        if errors.As(err, &errAccessDenied) {
            return nil, status.Errorf(codes.Unauthenticated, "Access Denied. Please ensure you have the right AWS permissions: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "Failed to add metadata to bucket %s for volume id %s: %s", bucket.Name, req.Name, err)
    }

    return &csipb.CreateVolumeResponse{
        Volume: &csipb.Volume{
            CapacityBytes: 0,
            VolumeId:      name,
        },
    }, nil
}

func (c *ControllerServer) DeleteVolume(ctx context.Context, req *csipb.DeleteVolumeRequest) (*csipb.DeleteVolumeResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME) {
        return nil, unimplementedError()
    }

    name := sanitizeVolumeID(req.VolumeId)

    s3Client, err := c.s3ClientFromSecrets(req.Secrets, fmt.Sprintf("DeleteVolume-%s", req.VolumeId))
    if err != nil {
        return nil, err
    }

    bucket := s3Client.NewBucket(name)

    // TODO: check volume usage (if there are fuse-servers running that use that bucket)
    /*
       metadata, err := bucket.GetMetadata(ctx)
       if err != nil {
           return nil, status.Error(codes.Internal, err.Error())
       }
       _, inUse :=
    */

    if err := bucket.Delete(ctx); err != nil {
        var errAccessDenied *s3.ErrAccessDenied
        if errors.As(err, &errAccessDenied) {
            return nil, status.Errorf(codes.Unauthenticated, "Access Denied. Please ensure you have the right AWS permissions: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "Failed to delete bucket %s for volume id %s: %s", bucket.Name, req.VolumeId, err)
    }

    return &csipb.DeleteVolumeResponse{}, nil
}

func (c *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csipb.ControllerPublishVolumeRequest) (*csipb.ControllerPublishVolumeResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerUnpublishVolume(ctx context.Context, request *csipb.ControllerUnpublishVolumeRequest) (*csipb.ControllerUnpublishVolumeResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, request *csipb.ValidateVolumeCapabilitiesRequest) (*csipb.ValidateVolumeCapabilitiesResponse, error) {
    // TODO: implement
    panic("TODO: implement")
}

func (c *ControllerServer) ListVolumes(ctx context.Context, req *csipb.ListVolumesRequest) (*csipb.ListVolumesResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_LIST_VOLUMES) {
        return nil, unimplementedError()
    }
    // TODO: this should use the grpc side car container that publishes fuse daemons to query the status
    panic("implement!")

    //var pageOpts *s3.PageOpts
    //
    //resp := &csipb.ListVolumesResponse{}
    //
    //if req.MaxEntries != 0 && req.StartingToken != "" {
    //    pageOpts = new(s3.PageOpts)
    //    start, err := strconv.Atoi(req.StartingToken)
    //    if err != nil {
    //        return nil, status.Error(codes.Aborted, "starting token is not a valid integer to use for pagination")
    //    }
    //    pageOpts.Start = start
    //
    //    size := int(req.MaxEntries)
    //    pageOpts.Size = size
    //    if req.MaxEntries != 0 {
    //        resp.NextToken = strconv.Itoa(start + size)
    //    }
    //}
    //
    //s3Client, err := c.s3ClientFromSecrets(req.Secrets, fmt.Sprintf("ListVolumes-%s", req.VolumeId))
    //if err != nil {
    //    return nil, err
    //}
    //
    //allBuckets, err := c.s3Client.ListBuckets(ctx, pageOpts)
    //if err != nil {
    //    if errors.Is(err, s3.ErrPageOutOfBounds) {
    //        return nil, status.Errorf(codes.Aborted, "starting_token is not valid: %s", err)
    //    }
    //    var errAccessDenied *s3.ErrAccessDenied
    //    if errors.As(err, &errAccessDenied) {
    //        return nil, status.Errorf(codes.Unauthenticated, "Access Denied. Please ensure you have the right AWS permissions: %v", err)
    //    }
    //    return nil, status.Errorf(codes.Internal, "Failed to listBuckets: %s", err)
    //}
    //
    //var volumes []*csipb.ListVolumesResponse_Entry
    //for _, b := range allBuckets {
    //    metadata, err := b.GetMetadata(ctx)
    //    if err != nil {
    //        var errAccessDenied *s3.ErrAccessDenied
    //        if errors.As(err, &errAccessDenied) {
    //            return nil, status.Errorf(codes.Unauthenticated, "Access Denied. Please ensure you have the right AWS permissions: %v", err)
    //        }
    //        return nil, status.Errorf(codes.Internal, "Failed to get bucket metadata for %s: %s", b.Name, err)
    //    }
    //
    //    pluginName, ok := metadata[BUCKET_CONTROLLED_BY_CSI_KEY]
    //    if ok && pluginName == BUCKET_CONTROLLED_BY_CSI_VALUE {
    //        volume := &csipb.ListVolumesResponse_Entry{
    //            Volume: volumeFromBucket(b, metadata),
    //        }
    //        volumes = append(volumes, volume)
    //    }
    //
    //    // TODO: check for volume status
    //}
    //
    //resp.Entries = volumes
    //
    //return resp, nil

}

func (c *ControllerServer) GetCapacity(ctx context.Context, request *csipb.GetCapacityRequest) (*csipb.GetCapacityResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_GET_CAPACITY) {
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
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) DeleteSnapshot(ctx context.Context, request *csipb.DeleteSnapshotRequest) (*csipb.DeleteSnapshotResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ListSnapshots(ctx context.Context, request *csipb.ListSnapshotsRequest) (*csipb.ListSnapshotsResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_LIST_SNAPSHOTS) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerExpandVolume(ctx context.Context, request *csipb.ControllerExpandVolumeRequest) (*csipb.ControllerExpandVolumeResponse, error) {
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_EXPAND_VOLUME) {
        return nil, unimplementedError()
    }
    panic("is supported, but not implemented")
}

func (c *ControllerServer) ControllerGetVolume(ctx context.Context, req *csipb.ControllerGetVolumeRequest) (*csipb.ControllerGetVolumeResponse, error) {
    // TODO: this should use the grpc side car container that publishes fuse daemons to query the status
    if !controllerCaps.isSupported(csipb.ControllerServiceCapability_RPC_GET_VOLUME) {
        return nil, unimplementedError()
    }
    panic("implement")
    //klog.Infof("recieved GetVolume RPC for Request: %s", req)
    //
    //name := sanitizeVolumeID(req.VolumeId)
    //klog.Infof("using sanitized name: %s for bucket", name)
    //
    //bucket := c.s3Client.NewBucket(name)
    //if !bucket.Exists(ctx) {
    //    return nil, status.Error(codes.NotFound, fmt.Sprintf("bucket for volume %s does not exist", name))
    //}
    //
    //metadata, err := bucket.GetMetadata(ctx)
    //if err != nil {
    //    var errAccessDenied *s3.ErrAccessDenied
    //    if errors.As(err, &errAccessDenied) {
    //        return nil, status.Errorf(codes.Unauthenticated, "Access Denied. Please ensure you have the right AWS permissions: %v", err)
    //    }
    //    return nil, status.Errorf(codes.Internal, "Failed to get metadata of bucket %s for volume id %s: %s", bucket.Name, req.VolumeId, err)
    //}
    //return &csipb.ControllerGetVolumeResponse{
    //    Volume: volumeFromBucket(bucket, metadata),
    //    Status: nil,
    //}, nil

}

func getControllerCapabilties() []*csipb.ControllerServiceCapability {
    var caps []*csipb.ControllerServiceCapability

    for capType, supported := range controllerCaps {
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

func volumeFromBucket(bucket *s3.Bucket, metadata s3.Metadata) *csipb.Volume {
    volume := &csipb.Volume{
        VolumeContext: map[string]string{
            BUCKET_VOLUME_NAME_KEY: bucket.Name,
        },
    }
    volumeId, ok := metadata[BUCKET_VOLUME_ID_KEY]
    if ok {
        volume.VolumeId = volumeId
    }
    return volume
}

func (c *ControllerServer) s3ClientFromSecrets(secrets map[string]string, sessionName string) (*s3.Client, error) {
    iamRole, ok := secrets[IAM_ROLE_KEY]
    if !ok {
        return nil, status.Errorf(codes.InvalidArgument, "IAM Role Key %s to assume is missing in secret map", IAM_ROLE_KEY)
    }
    awsCfg := aws.NewConfigWithRoleAssumer(c.assumer, iamRole, sessionName)
    return s3.NewClient(awsCfg), nil
}
