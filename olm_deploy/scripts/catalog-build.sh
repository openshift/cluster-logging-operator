#!/usr/bin/env bash
set -eo pipefail

IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY:-$LOCAL_IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}
echo "Building operator registry image ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
podman build --platform linux/amd64 -f olm_deploy/operatorregistry/Dockerfile -t ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY} .

if [ -n "${LOCAL_IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}" ] ; then
    coproc oc -n openshift-image-registry port-forward --address 0.0.0.0 service/image-registry 5000:5000
    trap "kill -15 $COPROC_PID" EXIT
    read PORT_FORWARD_STDOUT <&"${COPROC[0]}"
    if [[ "$PORT_FORWARD_STDOUT" =~ ^Forwarding.*5000$ ]] ; then
        user=$(oc whoami | sed s/://)
        token=$(oc whoami -t)
        podman login --tls-verify=false -u ${user} -p ${token} 127.0.0.1:5000
        if [[ "$(uname -s)" == "Darwin" ]]; then
            # Copy auth entry so push inside the VM can authenticate via host.containers.internal
            AUTH_FILE="${XDG_RUNTIME_DIR:-${HOME}/.config}/containers/auth.json"
            if [[ -f "${AUTH_FILE}" ]]; then
                jq '.auths["host.containers.internal:5000"] = .auths["127.0.0.1:5000"]' "${AUTH_FILE}" > "${AUTH_FILE}.tmp" && mv "${AUTH_FILE}.tmp" "${AUTH_FILE}"
            fi
            PUSH_TAG="${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY/127.0.0.1/host.containers.internal}"
            podman tag ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY} ${PUSH_TAG}
            IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY="${PUSH_TAG}"
        fi
    else
        echo "Unexpected message from oc port-forward: $PORT_FORWARD_STDOUT"
    fi
fi
echo "Pushing image ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
podman push --tls-verify=false ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}
