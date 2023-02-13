#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"
IMAGE_NAME="github.com/hown3d/s3-csi/fuse"
. $DIR/fuse-fs-docker-build.sh

helm repo add localstack-charts https://localstack.github.io/helm-charts
helm upgrade -i --wait localstack-fuse-fs localstack-charts/localstack \
  --set service.type="ClusterIP" \
  --set extraEnvVars[0].name="PROVIDER_OVERRIDE_S3" \
  --set extraEnvVars[0].value="asf" \
  --set startupScriptContent="awslocal s3api create-bucket --bucket=test" \
  --set enableStartupScripts=true \
  --set debug=true

kind load docker-image $IMAGE_NAME

kubectl delete pod s3-fuse || true
kubectl apply -f $DIR/fuse-pod.yaml