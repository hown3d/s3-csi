package fusePodManager

import (
    "context"
    pb "github.com/hown3d/s3-csi/proto/gen/fuse_pod_manager/v1alpha1"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/fields"
    corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
    "k8s.io/klog/v2"
    "strings"
)

type fusePodManagerService struct {
    podClient corev1Client.PodInterface
}

const (
    bucketAnnotation   string = "hown3d.s3-csi.bucket"
    volumeIdAnnotation string = "hown3d.s3-csi.volumeId"
    containerName      string = "fuse-fs"
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

    for _, container := range pod.Spec.Containers {
        if container.Name != containerName {
            continue
        }
        for _, arg := range container.Args {
            if strings.Contains(arg, "mount-dir") {
                mountFlagSplit := strings.Split(arg, "=")
                if len(mountFlagSplit) != 2 {
                    klog.Warningf("container fuse-fs does not contain valid mount-dir flag")
                    break
                }
                fusePod.MountPath = mountFlagSplit[1]
            }
        }
    }
    return fusePod
}

func (s fusePodManagerService) CreateFusePod(ctx context.Context, request *pb.CreateFusePodRequest) (*pb.CreateFusePodResponse, error) {
    //TODO implement me
    panic("implement me")
}

func (s fusePodManagerService) DeleteFusePod(ctx context.Context, request *pb.DeleteFusePodRequest) (*pb.DeleteFusePodResponse, error) {
    //TODO implement me
    panic("implement me")
}
