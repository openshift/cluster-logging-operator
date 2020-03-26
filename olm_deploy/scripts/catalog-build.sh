#!/bin/sh
set -eou pipefail

echo "Building operator registry image ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
docker build -f olm_deploy/operatorregistry/Dockerfile -t ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY} .

echo "Pushing operator registry image ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
docker push ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}
