#!/bin/sh
set -eou pipefail

IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY:-$LOCAL_IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}
echo "Building operator registry image ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
podman build -f olm_deploy/operatorregistry/Dockerfile -t ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY} .

if [ -n ${LOCAL_IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY} ] ; then
    oc -n openshift-image-registry port-forward service/image-registry 5000:5000 > /dev/null 2>&1 &
    forwarding_pid=$!
    trap "kill -15 ${forwarding_pid}" EXIT
    user=$(oc whoami | sed s/://)
    sleep 2
    podman login --tls-verify=false -u ${user} -p $(oc whoami -t) 127.0.0.1:5000 
fi
echo "Pushing image ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
podman push  --tls-verify=false ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}
