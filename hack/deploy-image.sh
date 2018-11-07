#!/bin/bash

# is this even needed for e2e tests?
exit 0

set -euxo pipefail

source "$(dirname $0)/common"

IMAGE_TAG=${IMAGE_TAG:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}

registry_port=$(oc get service docker-registry -n default -o jsonpath={.spec.ports[0].port})
tag="127.0.0.1:${registry_port}/openshift/cluster-logging-operator:latest"

${IMAGE_TAG} quay.io/openshift/cluster-logging-operator:latest ${tag}

echo "Setting up port-forwarding to remote docker-registry..."
oc port-forward service/docker-registry -n default ${LOCAL_PORT}:${registry_port} &
forwarding_pid=$!

trap "kill -15 ${forwarding_pid}" EXIT
sleep 1.0

echo "Pushing image ${tag}..."
# where/how is the oc login done to use the correct creds here?
docker login 127.0.0.1:${LOCAL_PORT} -u ${ADMIN_USER} -p $(oc whoami -t)
docker push ${tag}
