#!/bin/bash

set -euxo pipefail

if [ "${REMOTE_REGISTRY:-false}" = false ] ; then
    exit 0
fi

source "$(dirname $0)/common"

IMAGE_TAG_CMD=${IMAGE_TAG_CMD:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}
image_tag=$( echo "$IMAGE_TAG" | sed -e 's,quay.io/,,' )
tag=${tag:-"127.0.0.1:${registry_port}/$image_tag"}

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

${IMAGE_TAG_CMD} $IMAGE_TAG ${tag}

echo "Setting up port-forwarding to remote $registry_svc ..."
oc --loglevel=9 port-forward $port_fwd_obj -n $registry_namespace ${LOCAL_PORT}:${registry_port} > pf.log 2>&1 &
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

echo "Pushing image ${tag}..."
docker login --tls-verify=false 127.0.0.1:${LOCAL_PORT} -u ${ADMIN_USER} -p $(oc whoami -t)
docker push --tls-verify=false ${tag}
