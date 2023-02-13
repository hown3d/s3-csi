#!/usr/bin/env bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && /bin/pwd)"
set -e

rm -rf $DIR/../packages

if [ ! -f ${DIR}/../melange.rsa ]; then
  docker run --rm -v "${DIR}/../":/work distroless.dev/melange keygen
fi
docker run --privileged --rm -v "${DIR}/../":/work \
  distroless.dev/melange build melange.fuse-fs.yaml \
  --arch amd64 \
  --repository-append packages \
  --signing-key melange.rsa

# Your GitHub username
GITHUB_USERNAME="hown3d"
REF="ghcr.io/${GITHUB_USERNAME}/s3-csi/fuse-fs"

docker run --rm -w /work -v "${DIR}/../":/work \
  distroless.dev/apko build apko.fuse-fs.yaml \
  "${REF}" /work/packages/fuse-fs-output.tar -k  melange.rsa.pub \
  --build-arch amd64