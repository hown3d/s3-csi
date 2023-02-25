#!/usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"

IMAGES=fuse-pod-manager KO_DOCKER_REPO_PREFIX=$(minikube ip):5000 $DIR/../build-with-ko.sh

kustomize build $DIR/../../manifests/overlays/minikube/ | kubectl apply -f -
