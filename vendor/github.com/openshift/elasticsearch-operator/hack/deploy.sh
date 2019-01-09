#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

if [ $REMOTE_REGISTRY = false ] ; then
    oc create -n ${NAMESPACE} -f manifests/05-deployment.yaml
else
    cat manifests/05-deployment.yaml | \
        sed -e "s,${IMAGE_TAG},${registry_host}:5000/${IMAGE_TAG}," | \
	    oc create -n ${NAMESPACE} -f -
fi
