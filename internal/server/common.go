package server

import (
    "crypto/sha1"
    "encoding/hex"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "io"
    "strings"
)

func sanitizeVolumeID(volumeID string) string {
    volumeID = strings.ToLower(volumeID)
    if len(volumeID) > 63 {
        h := sha1.New()
        io.WriteString(h, volumeID)
        volumeID = hex.EncodeToString(h.Sum(nil))
    }
    return volumeID
}

func unimplementedError() error {
    return status.Error(codes.Unimplemented, "unimplemented")
}
