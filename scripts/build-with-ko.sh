#!/usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"

REPO=hown3d/s3-csi
export KO_DOCKER_REPO=${KO_DOCKER_REPO_PREFIX}/$REPO
for img in $IMAGES; do
  MAIN="$DIR/../cmd/$img"
  TAG=$(ko build $MAIN)
  pushd manifests/overlays/minikube
  kustomize edit set image ghcr.io/$REPO/$img=$(echo $TAG | sed "s/$(minikube ip)/localhost/")
  popd
done
