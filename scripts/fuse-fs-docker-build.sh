#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"

docker build -f $DIR/../deploy/fuse-fs/Dockerfile -t $IMAGE_NAME $DIR/../
