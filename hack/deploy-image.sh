#!/bin/bash

set -euo pipefail

echo "Setting up port-forwarding to remote registry ..."
oc -n openshift-image-registry port-forward service/image-registry 5000:5000 > pf.log 2>&1 &
forwarding_pid=$!
trap "kill -15 ${forwarding_pid}" EXIT

user=$(oc whoami | sed s/://)
echo "Login to registry..."
sleep 2
podman login --tls-verify=false -u ${user} -p $(oc whoami -t) 127.0.0.1:5000 

echo "Pushing image ${IMAGE_TAG} ..."
if podman push --tls-verify=false ${IMAGE_TAG} ; then
    oc -n openshift get imagestreams | grep cluster-logging-operator
fi
