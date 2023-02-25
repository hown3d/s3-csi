package main

import (
    "github.com/hown3d/s3-csi/internal/fusePodManager"
    "log"
)

func main() {
    log.Println("starting server on port 9090")
    srv := fusePodManager.NewServer(9090)
    if err := srv.Run(); err != nil {
        log.Fatal(err)
    }
}
