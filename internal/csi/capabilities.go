package server

import csipb "github.com/container-storage-interface/spec/lib/go/csi"

type capMap[T capType] map[T]bool

type capType interface {
    csipb.NodeServiceCapability_RPC_Type | csipb.ControllerServiceCapability_RPC_Type
}

func (m capMap[T]) isSupported(cap T) bool {
    supported, ok := m[cap]
    if !ok {
        return false
    }
    return supported
}
