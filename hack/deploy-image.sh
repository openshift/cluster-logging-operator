#!/usr/bin/env bash

set -euo pipefail

echo "Setting up port-forwarding to remote registry ..."
coproc oc -n openshift-image-registry port-forward --address 0.0.0.0 service/image-registry 5000:5000
trap "kill -15 $COPROC_PID" EXIT
read PORT_FORWARD_STDOUT <&"${COPROC[0]}"
if [[ "$PORT_FORWARD_STDOUT" =~ ^Forwarding.*5000$ ]] ; then
    # On macOS, podman runs in a VM and can't reach host's 127.0.0.1.
    # Use host.containers.internal to reach the host from the VM.
    if [[ "$(uname -s)" == "Darwin" ]]; then
        REGISTRY_HOST="host.containers.internal"
    else
        REGISTRY_HOST="127.0.0.1"
    fi

    user=$(oc whoami | sed s/://)
    token=$(oc whoami -t)
    echo "Login to registry..."
    podman login --tls-verify=false -u ${user} -p ${token} 127.0.0.1:5000
    if [[ "$(uname -s)" == "Darwin" ]]; then
        # Copy auth entry so push inside the VM can authenticate via host.containers.internal
        AUTH_FILE="${XDG_RUNTIME_DIR:-${HOME}/.config}/containers/auth.json"
        if [[ -f "${AUTH_FILE}" ]]; then
            jq '.auths["host.containers.internal:5000"] = .auths["127.0.0.1:5000"]' "${AUTH_FILE}" > "${AUTH_FILE}.tmp" && mv "${AUTH_FILE}.tmp" "${AUTH_FILE}"
        fi
    fi

    PUSH_TAG="${IMAGE_TAG/127.0.0.1/${REGISTRY_HOST}}"
    echo "Pushing image ${IMAGE_TAG} to ${PUSH_TAG} ..."
    if podman push --tls-verify=false ${IMAGE_TAG} ${PUSH_TAG} ; then
        oc -n openshift get imagestreams | grep cluster-logging-operator
    fi
else
    echo "Unexpected message from oc port-forward: $PORT_FORWARD_STDOUT"
fi
