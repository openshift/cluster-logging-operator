#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

if [ $REMOTE_REGISTRY = false ] ; then
    $repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'dep' \
    |  oc create -f -
else
    if [ "${USE_IMAGE_STREAM_FOR_LOGGING:-false}" = false ] ; then
        fix_images() { cat ; }
    else
        fix_images() {
            sed -e "s,docker.io/openshift/origin-logging,$registry_host:5000/openshift/logging," \
                -e "s,quay.io/openshift/origin-logging,$registry_host:5000/openshift/logging,"
        }
    fi
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
    $repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'dep' | \
        replace_image | \
        fix_images | \
	    oc -n $NAMESPACE create -f -
fi

if [ "${NO_BUILD:-false}" = true ] ; then
    CREATE_ES_SECRET=false NAMESPACE=openshift-logging IMAGE_OVERRIDE=${EO_IMAGE_OVERRIDE:-} make -C ${ELASTICSEARCH_OP_REPO} deploy-no-build
else
    CREATE_ES_SECRET=false NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} deploy
fi
