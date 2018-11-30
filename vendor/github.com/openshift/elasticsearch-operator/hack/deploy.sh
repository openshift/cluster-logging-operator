#!/bin/bash

if [ "${REMOTE_REGISTRY:-false}" = false ] ; then
    exit 0
fi

set -euxo pipefail

source "$(dirname $0)/common"

registry_ip=$(oc get service docker-registry -n default -o jsonpath={.spec.clusterIP})
cat manifests/05-deployment.yaml | \
    sed -e "s,${IMAGE_TAG},${registry_ip}:5000/openshift/elasticsearch-operator:latest," | \
	oc create -n ${NAMESPACE} -f -
