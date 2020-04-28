#!/bin/bash

if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi
set -euo pipefail

source "$(dirname $0)/common"

IMAGE_TAG=$1
IMAGE_BUILDER=${2:-imagebuilder}
IMAGE_BUILDER_OPTS=${3:-}

echo building image $IMAGE_TAG - this may take a few minutes until you see any output . . .
podman build $IMAGE_BUILDER_OPTS -t $IMAGE_TAG .
