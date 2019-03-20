#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

if [ $REMOTE_REGISTRY = false ] ; then
    oc create -n ${NAMESPACE} -f manifests/05-deployment.yaml
else
    image_tag=$( echo "$IMAGE_TAG" | sed -e 's,quay.io/,,' )
    cat manifests/05-deployment.yaml | \
        sed -e "s,${IMAGE_TAG},${registry_host}:5000/${image_tag}," | \
	    oc create -n ${NAMESPACE} -f -
fi
