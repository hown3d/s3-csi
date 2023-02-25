package fusePodManager

import (
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
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := parsePodToFusePodMessage(tt.args.pod); !reflect.DeepEqual(got, tt.want) {
                t.Errorf("parsePodToFusePodMessage() = %v, want %v", got, tt.want)
            }
        })
    }
}
