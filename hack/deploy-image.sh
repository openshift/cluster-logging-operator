#!/bin/bash

set -euxo pipefail

if [ "${REMOTE_REGISTRY:-false}" = false ] ; then
    exit 0
fi

source "$(dirname $0)/common"

IMAGE_TAG_CMD=${IMAGE_TAG_CMD:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}

tag="127.0.0.1:${registry_port}/$IMAGE_TAG"

${IMAGE_TAG_CMD} $IMAGE_TAG ${tag}

echo "Setting up port-forwarding to remote $registry_svc ..."
oc port-forward $port_fwd_obj -n $registry_namespace ${LOCAL_PORT}:${registry_port} &
forwarding_pid=$!

trap "kill -15 ${forwarding_pid}" EXIT
sleep 1.0

echo "Pushing image ${tag}..."
docker login 127.0.0.1:${LOCAL_PORT} -u ${ADMIN_USER} -p $(oc whoami -t)
docker push ${tag}
