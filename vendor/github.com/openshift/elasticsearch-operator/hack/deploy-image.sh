#!/bin/bash

if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi
set -euo pipefail

if [ "${REMOTE_REGISTRY:-true}" = false ] ; then
    exit 0
fi

source "$(dirname $0)/common"

if [ "${USE_IMAGE_STREAM:-false}" = true ] ; then
    oc process \
        -p ES_OP_GITHUB_URL=https://github.com/${ES_OP_GITHUB_REPO:-openshift}/elasticsearch-operator \
        -p ES_OP_GITHUB_BRANCH=${ES_OP_GITHUB_BRANCH:-master} \
        -f hack/image-stream-build-config-template.yaml | \
      oc -n openshift create -f -
    # wait for is and bc
    for ii in $(seq 1 60) ; do
        if oc -n openshift get bc elasticsearch-operator > /dev/null 2>&1 && \
           oc -n openshift get is origin-elasticsearch-operator > /dev/null 2>&1 ; then
            break
        fi
        sleep 1
    done
    if [ $ii = 60 ] ; then
        echo ERROR: timeout waiting for elasticsearch-operator buildconfig and imagestream to be available
        exit 1
    fi
    # wait
    oc -n openshift start-build --follow bc/elasticsearch-operator
    exit 0
fi

IMAGE_TAGGER=${IMAGE_TAGGER:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}

image_tag=$( echo "$IMAGE_TAG" | sed -e 's,quay.io/,,' )
tag=${tag:-"127.0.0.1:${LOCAL_PORT}/$image_tag"}

${IMAGE_TAGGER} ${IMAGE_TAG} ${tag}

echo "Setting up port-forwarding to remote registry..."
oc --loglevel=9 -n openshift-image-registry port-forward service/image-registry ${LOCAL_PORT}:5000 > pf.log 2>&1 &
forwarding_pid=$!

trap "kill -15 ${forwarding_pid}" EXIT
for ii in $(seq 1 10) ; do
    if [ "$(curl -sk -w '%{response_code}\n' https://localhost:5000 || :)" = 200 ] ; then
        break
    fi
    sleep 1
done
if [ $ii = 10 ] ; then
    echo ERROR: timeout waiting for port-forward to be available
    exit 1
fi

login_to_registry "127.0.0.1:${LOCAL_PORT}"
echo "Pushing image ${IMAGE_TAG} to ${tag} ..."
rc=0
for ii in $( seq 1 5 ) ; do
    if push_image ${IMAGE_TAG} ${tag} ; then
        rc=0
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
