package fusePodManager

import (
    "context"
    "fmt"
    pb "github.com/hown3d/s3-csi/proto/gen/fuse_pod_manager/v1alpha1"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/fields"
    corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
    "k8s.io/klog/v2"
    "k8s.io/utils/pointer"
)

type fusePodManagerService struct {
    podClient corev1Client.PodInterface
}

const (
    bucketAnnotation   string = "hown3d.s3-csi.bucket"
    volumeIdAnnotation string = "hown3d.s3-csi.volumeId"
    containerName      string = "fuse-fs"
    containerMntPath   string = "/tmp/s3-fs-mnt"

    hostVolumeName = "fs"
)

var (
    labels = fields.Set{
        "hown3d.s3-csi": "fuse-pod",
    }
    labelSelector = fields.SelectorFromSet(labels).String()
)

func (s fusePodManagerService) ListFusePods(ctx context.Context, _ *pb.ListFusePodsRequest) (*pb.ListFusePodsResponse, error) {
    opts := metav1.ListOptions{LabelSelector: labelSelector}
    podList, err := s.podClient.List(ctx, opts)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "getting pods from kubernetes api: %s", err)
    }

    fusePods := make([]*pb.FusePod, 0, len(podList.Items))
    for _, pod := range podList.Items {
        fusePod := parsePodToFusePodMessage(&pod)
        fusePods = append(fusePods, fusePod)
    }
    return &pb.ListFusePodsResponse{
        Pods: fusePods,
    }, nil
}

func parsePodToFusePodMessage(pod *corev1.Pod) *pb.FusePod {
    fusePod := &pb.FusePod{
        Name: pod.Name,
    }

    bucket, ok := pod.Annotations[bucketAnnotation]
    if ok {
        fusePod.Bucket = bucket
    }

    volumeId, ok := pod.Annotations[volumeIdAnnotation]
    if ok {
        fusePod.VolumeId = volumeId
    }

    for _, vol := range pod.Spec.Volumes {
        if vol.Name != hostVolumeName {
            continue
        }
        if vol.HostPath == nil {
            klog.Warningf("hostPath for host volume is nil, skipping")
        }
        fusePod.HostMountPath = vol.HostPath.Path
    }
    return fusePod
}

func (s fusePodManagerService) CreateFusePod(ctx context.Context, request *pb.CreateFusePodRequest) (*pb.CreateFusePodResponse, error) {
    volumeId := request.VolumeId
    if volumeId == "" {
        return nil, status.Errorf(codes.InvalidArgument, "volumeId cant be empty")
    }
    podImage := request.Image
    if podImage == "" {
        return nil, status.Errorf(codes.InvalidArgument, "podImage cant be empty")
    }

    bucket := request.Bucket
    if bucket == "" {
        return nil, status.Errorf(codes.InvalidArgument, "bucket cant be empty")
    }

    hostMntPath := request.HostMountPath
    if hostMntPath == "" {
        return nil, status.Errorf(codes.InvalidArgument, "hostMntPath cant be empty")
    }

    podName := fmt.Sprintf("fuse-fs-%s", volumeId)

    config := &podConfig{
        podName:     podName,
        podImage:    podImage,
        volumeId:    volumeId,
        bucket:      bucket,
        hostMntPath: hostMntPath,
    }
    pod := generatePod(config)

    _, err := s.podClient.Create(ctx, pod, metav1.CreateOptions{})
    if err != nil {
        return nil, status.Errorf(codes.Internal, "creating pod: %s", err)
    }
    return &pb.CreateFusePodResponse{
        Name: podName,
    }, nil
}

type podConfig struct {
    podName     string
    podImage    string
    volumeId    string
    bucket      string
    hostMntPath string
}

func generatePod(config *podConfig) *corev1.Pod {
    // store as variable to allow pointer reference
    var (
        mountPropBidirectional = corev1.MountPropagationBidirectional
        hostPathDirOrCreate    = corev1.HostPathDirectoryOrCreate
    )

    return &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: config.podName,
            Annotations: map[string]string{
                volumeIdAnnotation: config.volumeId,
                bucketAnnotation:   config.bucket,
            },
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:  containerName,
                    Image: config.podImage,
                    Args: []string{
                        fmt.Sprintf("-s3-bucket=%s", config.bucket),
                        fmt.Sprintf("-mount-dir=%s", containerMntPath),
                    },
                    SecurityContext: &corev1.SecurityContext{
                        Privileged: pointer.Bool(true),
                    },
                    VolumeMounts: []corev1.VolumeMount{
                        {
                            Name:      "fuse",
                            MountPath: "/dev/fuse",
                        },
                        {
                            Name:             "fs",
                            MountPath:        containerMntPath,
                            MountPropagation: &mountPropBidirectional,
                        },
                    },
                },
            },
            Volumes: []corev1.Volume{
                {
                    Name: "fuse",
                    VolumeSource: corev1.VolumeSource{
                        HostPath: &corev1.HostPathVolumeSource{
                            Path: "/dev/fuse",
                        },
                    },
                },

                {
                    Name: "fs",
                    VolumeSource: corev1.VolumeSource{
                        HostPath: &corev1.HostPathVolumeSource{
                            Path: config.hostMntPath,
                            Type: &hostPathDirOrCreate,
                        },
                    },
                },
            },
        },
    }
}

func (s fusePodManagerService) DeleteFusePod(ctx context.Context, request *pb.DeleteFusePodRequest) (*pb.DeleteFusePodResponse, error) {
    //TODO implement me
    panic("implement me")
}
