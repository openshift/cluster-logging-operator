#!/bin/bash

if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi
set -euo pipefail

source "$(dirname $0)/common"

IMAGE_TAG=$1
IMAGE_BUILDER=${2:-imagebuilder}
IMAGE_BUILDER_OPTS=${3:-}

workdir=${WORKDIR:-$( mktemp --tmpdir -d elasticsearch-operator-build-XXXXXXXXXX )}
if [ -z "${WORKDIR:-}" ] ; then
    trap "rm -rf $workdir" EXIT
fi

if image_is_ubi Dockerfile ; then
    pull_ubi_if_needed
fi

if image_needs_private_repo Dockerfile ; then
    repodir=$( get_private_repo_dir $workdir )
    mountarg="-mount $repodir:/etc/yum.repos.d/"
else
    mountarg=""
fi

echo building image $IMAGE_TAG - this may take a few minutes until you see any output . . .
$IMAGE_BUILDER $IMAGE_BUILDER_OPTS $mountarg -t $IMAGE_TAG .
