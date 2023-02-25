package fusePodManager

import (
    "fmt"
    pb "github.com/hown3d/s3-csi/proto/gen/fuse_pod_manager/v1alpha1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "reflect"
    "testing"
)

func Test_parsePodToFusePodMessage(t *testing.T) {
    type args struct {
        pod *corev1.Pod
    }
    tests := []struct {
        name string
        args args
        want *pb.FusePod
    }{
        {
            name: "missing annotations",
            args: args{
                pod: generateTestPod("test123", "", "", "test-pod"),
            },
            want: &pb.FusePod{Name: "test-pod", MountPath: "test123"},
        },
        {
            name: "happy path",
            args: args{
                pod: generateTestPod("test123", "test", "volume123", "test-pod"),
            },
            want: &pb.FusePod{
                Name:      "test-pod",
                VolumeId:  "volume123",
                Bucket:    "test",
                MountPath: "test123",
            },
        },

        {
            name: "missing mountPath",
            args: args{
                pod: generateTestPod("", "test", "volume123", "test-pod"),
            },
            want: &pb.FusePod{
                Name:     "test-pod",
                VolumeId: "volume123",
                Bucket:   "test",
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := parsePodToFusePodMessage(tt.args.pod); !reflect.DeepEqual(got, tt.want) {
                t.Errorf("parsePodToFusePodMessage() = %v, want %v", got, tt.want)
            }
        })
    }
}

func generateTestPod(mountDir string, bucketAnnotationValue string, volumeIdAnnotationValue string, podName string) *corev1.Pod {
    annotations := map[string]string{}
    if bucketAnnotationValue != "" {
        annotations[bucketAnnotation] = bucketAnnotationValue
    }
    if volumeIdAnnotationValue != "" {
        annotations[volumeIdAnnotation] = volumeIdAnnotationValue
    }

    var mountDirArg string
    if mountDir != "" {
        mountDirArg = fmt.Sprintf("-mount-dir=%s", mountDir)
    }
    return &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:        podName,
            Annotations: annotations,
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name: containerName,
                    Args: []string{mountDirArg},
                },
            },
        },
    }
}
