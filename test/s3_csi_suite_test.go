package test_test

import (
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestS3Csi(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "S3 CSI Suite")
}
