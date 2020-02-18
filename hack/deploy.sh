#!/bin/bash

set -euo pipefail
if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi

source "$(dirname $0)/common"

IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
deploy_clusterlogging_operator
