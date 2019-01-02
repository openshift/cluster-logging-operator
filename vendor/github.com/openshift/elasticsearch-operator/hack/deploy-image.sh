#!/bin/bash

set -euxo pipefail

if [ "${REMOTE_REGISTRY:-false}" = false ] ; then
    exit 0
fi

source "$(dirname $0)/common"

IMAGE_TAGGER=${IMAGE_TAGGER:-docker tag}
LOCAL_PORT=${LOCAL_PORT:-5000}

tag=${tag:-"127.0.0.1:${registry_port}/$IMAGE_TAG"}

${IMAGE_TAGGER} ${IMAGE_TAG} ${tag}

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
docker login 127.0.0.1:${LOCAL_PORT} -u ${ADMIN_USER} -p $(oc whoami -t)
docker push ${tag}
