#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"

GOPKG=$(go env GOPATH)/pkg
docker run --rm \
  -v $DIR/..:/$DIR/.. \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $GOPKG:/go/pkg \
  -w $DIR/.. \
  golang:1.19 \
  go test -v ./...