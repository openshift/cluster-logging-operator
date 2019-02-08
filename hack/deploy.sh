#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

if [ $REMOTE_REGISTRY = false ] ; then
    oc create -n ${NAMESPACE} -f manifests/05-deployment.yaml
else
    if [ "${USE_IMAGE_STREAM_FOR_LOGGING:-false}" = false ] ; then
        fix_images() { cat ; }
    else
        fix_images() {
            sed -e "s,docker.io/openshift/origin-logging,$registry_host:5000/openshift/logging,"
                -e "s,quay.io/openshift/origin-logging,$registry_host:5000/openshift/logging,"
        }
    fi
    image_tag=$( echo "$IMAGE_TAG" | sed -e 's,quay.io/,,' )
    cat manifests/05-deployment.yaml | \
        sed -e "s,${IMAGE_TAG},${registry_host}:5000/${image_tag}," | \
        fix_images | \
	    oc create -n ${NAMESPACE} -f -
fi

if [ "${NO_BUILD:-false}" = true ] ; then
    CREATE_ES_SECRET=false NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} deploy-no-build
else
    CREATE_ES_SECRET=false NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} deploy
fi
