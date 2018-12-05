#!/bin/bash

set -euxo pipefail

if [ "${REMOTE_REGISTRY:-false}" = false ] ; then
    exit 0
fi

source "$(dirname $0)/common"

IMAGE_TAG_CMD=${IMAGE_TAG_CMD:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}

registry_port=$(oc get service docker-registry -n default -o jsonpath={.spec.ports[0].port})
tag="127.0.0.1:${registry_port}/openshift/cluster-logging-operator:latest"

${IMAGE_TAG_CMD} openshift/cluster-logging-operator:latest ${tag}

echo "Setting up port-forwarding to remote docker-registry..."
oc port-forward service/docker-registry -n default ${LOCAL_PORT}:${registry_port} &
forwarding_pid=$!

trap "kill -15 ${forwarding_pid}" EXIT
sleep 1.0

echo "Pushing image ${tag}..."
docker login 127.0.0.1:${LOCAL_PORT} -u ${ADMIN_USER} -p $(oc whoami -t)
docker push ${tag}
