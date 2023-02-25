package fusePodManager

import (
    pb "github.com/hown3d/s3-csi/proto/gen/fuse_pod_manager/v1alpha1"
    corev1 "k8s.io/api/core/v1"
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
                pod: generatePod(&podConfig{
                    podName:     "test-pod",
                    hostMntPath: "test123",
                }),
            },
            want: &pb.FusePod{
                Name:          "test-pod",
                HostMountPath: "test123",
            },
        },
        {
            name: "happy path",
            args: args{
                pod: generatePod(&podConfig{
                    podName:     "test-pod",
                    volumeId:    "volume123",
                    bucket:      "test",
                    hostMntPath: "test123",
                }),
            },
            want: &pb.FusePod{
                Name:          "test-pod",
                VolumeId:      "volume123",
                Bucket:        "test",
                HostMountPath: "test123",
            },
        },

        {
            name: "missing mountPath",
            args: args{
                pod: generatePod(&podConfig{
                    podName:  "test123",
                    volumeId: "volume123",
                    bucket:   "test",
                }),
            },
            want: &pb.FusePod{
                Name:     "test123",
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
