#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

if [ "${REMOTE_REGISTRY:-false}" = true ] ; then
    registry_ip=$(oc get service docker-registry -n default -o jsonpath={.spec.clusterIP})
    cat manifests/05-deployment.yaml | \
        sed -e "s,${IMAGE_TAG},${registry_ip}:5000/openshift/elasticsearch-operator:latest," | \
	    oc create -f -
else
    oc create -f manifests/05-deployment.yaml
fi

CREATE_ES_SECRET=false NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} deploy
