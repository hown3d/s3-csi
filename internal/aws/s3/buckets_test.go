package s3

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestPageOpts_Slice(t *testing.T) {
    type testCase struct {
        name      string
        p         PageOpts
        input     []string
        expected  []string
        shouldErr bool
    }
    tests := []testCase{
        {
            name: "Start is 0 and Size is 1 when slice is len(5) should return slice with only first element",
            p: PageOpts{
                Start: 0,
                Size:  1,
            },
            input:    []string{"1", "2", "3", "4", "5"},
            expected: []string{"1"},
        },

        {
            name: "Size is bigger then slice should return full slice",
            p: PageOpts{
                Start: 0,
                Size:  10,
            },
            input:    []string{"1", "2", "3", "4", "5"},
            expected: []string{"1", "2", "3", "4", "5"},
        },
        {
            name: "Start is 0 and Size is exactly len(slice) should return full slice",
            p: PageOpts{
                Start: 0,
                Size:  5,
            },
            input:    []string{"1", "2", "3", "4", "5"},
            expected: []string{"1", "2", "3", "4", "5"},
        },
        {
            name: "Start is 0 and Size is nil should return full slice",
            p: PageOpts{
                Start: 0,
                Size:  5,
            },
            input:    []string{"1", "2", "3", "4", "5"},
            expected: []string{"1", "2", "3", "4", "5"},
        },
        {
            name: "Start is bigger then slice should return an error",
            p: PageOpts{
                Start: 7,
            },
            input:     []string{"1", "2", "3", "4", "5"},
            shouldErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual, err := slice(tt.input, tt.p)
            if tt.shouldErr {
                assert.Error(t, err)
            }
            assert.Equal(t, tt.expected, actual)
        })
    }
}
