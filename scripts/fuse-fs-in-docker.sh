#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"

IMAGE_NAME="fuse-fs"
. $DIR/fuse-fs/docker-build.sh

docker run \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN \
  -e AWS_REGION=eu-central-1 \
   --privileged \
   --device=/dev/fuse \
   --mount type=bind,source=/tmp/lima/mydir,target=/tmp/s3-fuse-mnt,bind-propagation=shared  \
   $IMAGE_NAME \
   -s3-bucket=s3-fuse-test \
   -debug
