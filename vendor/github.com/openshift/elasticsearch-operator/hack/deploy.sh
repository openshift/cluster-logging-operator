#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

if [ $REMOTE_REGISTRY = false ] ; then
    oc create -n ${NAMESPACE} -f manifests/05-deployment.yaml
else
    if [ -n "${IMAGE_OVERRIDE:-}" ] ; then
        replace_image() {
            sed -e "s, image:.*\$, image: ${IMAGE_OVERRIDE},"
        }
    else
        replace_image() {
            sed -e "s,${IMAGE_TAG},${registry_host}:5000/${image_tag},"
        }
    fi
    image_tag=$( echo "$IMAGE_TAG" | sed -e 's,quay.io/,,' )
    cat manifests/05-deployment.yaml | \
        replace_image | \
	    oc create -n ${NAMESPACE} -f -
fi
