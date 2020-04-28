#!/bin/bash

set -euo pipefail

if [ "${REMOTE_REGISTRY:-true}" = false ] ; then
    exit 0
fi

source "$(dirname $0)/common"

IMAGE_TAG_CMD=${IMAGE_TAG_CMD:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}
image_tag=$( echo "$IMAGE_TAG" | sed -e 's,quay.io/,,' )
tag=${tag:-"127.0.0.1:${LOCAL_PORT}/$image_tag"}

if [ "${USE_IMAGE_STREAM:-false}" = true ] ; then
    oc process \
        -p CL_OP_GITHUB_URL=https://github.com/${CL_OP_GITHUB_REPO:-openshift}/cluster-logging-operator \
        -p CL_OP_GITHUB_BRANCH=${CL_OP_GITHUB_BRANCH:-master} \
        -f hack/image-stream-build-config-template.yaml | \
      oc -n openshift create -f -
    # wait for is and bc
    for ii in $(seq 1 60) ; do
        if oc -n openshift get bc cluster-logging-operator > /dev/null 2>&1 && \
           oc -n openshift get is origin-cluster-logging-operator > /dev/null 2>&1 ; then
            break
        fi
        sleep 1
    done
    if [ $ii = 60 ] ; then
        echo ERROR: timeout waiting for cluster-logging-operator buildconfig and imagestream to be available
        exit 1
    fi
    # build and wait
    oc -n openshift start-build --follow bc/cluster-logging-operator
    exit 0
fi

podman tag $IMAGE_TAG ${tag}

echo "Setting up port-forwarding to remote registry ..."
oc --loglevel=9 -n openshift-image-registry port-forward service/image-registry ${LOCAL_PORT}:5000 > pf.log 2>&1 &
forwarding_pid=$!

trap "kill -15 ${forwarding_pid}" EXIT
for ii in $(seq 1 30) ; do
    if [ "$(curl -sk -w '%{response_code}\n' https://localhost:5000 || :)" = 200 ] ; then
        break
    fi
    sleep 1
done
if [ $ii = 30 ] ; then
    echo ERROR: timeout waiting for port-forward to be available
    exit 1
fi

login_to_registry "127.0.0.1:${LOCAL_PORT}"
echo "Pushing image ${IMAGE_TAG} to ${tag} ..."
rc=0
for ii in $( seq 1 5 ) ; do
    if push_image ${IMAGE_TAG} ${tag} ; then
        rc=0
        oc -n openshift get imagestreams | grep cluster-logging-operator
        oc -n openshift get imagestreamtags | grep cluster-logging-operator
        break
    fi
    echo push failed - retrying
    rc=1
    sleep 1
done
if [ $rc = 1 -a $ii = 5 ] ; then
    echo ERROR: giving up push of ${IMAGE_TAG} to ${tag} after 5 tries
    exit 1
fi
