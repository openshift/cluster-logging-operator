#!/bin/bash

if [ "${REMOTE_REGISTRY:-false}" = false ] ; then
    exit 0
fi

set -euxo pipefail

source "$(dirname $0)/common"

IMAGE_TAGGER=${IMAGE_TAGGER:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}

registry_port=$(oc get service docker-registry -n default -o jsonpath={.spec.ports[0].port})
tag="127.0.0.1:${registry_port}/openshift/elasticsearch-operator:latest"

${IMAGE_TAGGER} ${IMAGE_TAG} ${tag}

echo "Setting up port-forwarding to remote docker-registry..."
oc port-forward service/docker-registry -n default ${LOCAL_PORT}:${registry_port} &
forwarding_pid=$!

trap "kill -15 ${forwarding_pid}" EXIT
sleep 1.0

echo "Pushing image ${tag}..."
docker login 127.0.0.1:${LOCAL_PORT} -u ${ADMIN_USER} -p $(oc whoami -t)
docker push ${tag}
