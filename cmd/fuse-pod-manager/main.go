package main

import (
    "github.com/hown3d/s3-csi/internal/fusePodManager"
    "log"
    "os"
)

func main() {
    log.Println("starting server on port 9090")
    namespace := os.Getenv("K8S_NAMESPACE")
    if namespace == "" {
        log.Println("K8S_NAMESPACE is empty, using current context")
    }
    srv, err := fusePodManager.NewServer(9090, namespace)
    if err != nil {
        log.Fatalf("creating fuse-pod-manager server: %s", err)
    }
    if err := srv.Run(); err != nil {
        log.Fatal(err)
    }
}
