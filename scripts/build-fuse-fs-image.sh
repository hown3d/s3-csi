#!/usr/bin/env bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && /bin/pwd)"
set -e

if [ ! -f ${DIR}/../melange.rsa ]; then
  docker run --rm -v "${DIR}/../":/work distroless.dev/melange keygen
fi

for image in $IMAGES; do
  MELANGE_CONFIG="melange.$image.yaml"
  echo "building package ${MELANGE_CONFIG}"
  docker run --privileged --rm -v "${DIR}/../":/work \
    distroless.dev/melange build ${MELANGE_CONFIG} \
    --arch amd64 \
    --repository-append packages \
    --signing-key melange.rsa
done


# Your GitHub username
GITHUB_USERNAME="hown3d"

for image in $IMAGES; do
  REF="ghcr.io/${GITHUB_USERNAME}/s3-csi/${image}"
  echo "building image $image to ref $REF"
  docker run --rm -w /work -v "${DIR}/../":/work \
    distroless.dev/apko build "apko.${image}.yaml" \
    "${REF}" /work/packages/${image}.tar -k melange.rsa.pub \
    --build-arch amd64
done
