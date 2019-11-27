#!/bin/bash

set -euo pipefail

source "$(dirname $0)/common"

IMAGE_ELASTICSEARCH_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-elasticsearch-operator:latest \
deploy_elasticsearch_operator