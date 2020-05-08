#!/bin/bash

set -euo pipefail

function login_to_registry() {
    local savectx=$( oc config current-context )
    local token=""
    local username=""
    if [ -n "${PUSH_USER:-}" -a -n "${PUSH_PASSWORD:-}" ] ; then
        username=$PUSH_USER
        if [ "$username" = "kube:admin" ] ; then
            username=kubeadmin
        fi
        oc login -u "$username" -p "$PUSH_PASSWORD" > /dev/null
        token=$( oc whoami -t 2> /dev/null || : )
        oc config use-context "$savectx"
    else
        # see if current context has a token
        token=$( oc whoami -t 2> /dev/null || : )
        if [ -n "$token" ] ; then
            username=$( oc whoami )
        else
            # get the first user with a token
            token=$( oc config view -o go-template='{{ range .users }}{{ if .user.token }}{{ print .user.token }}{{ end }}{{ end }}' )
            if [ -n "$token" ] ; then
                username=$( oc config view -o go-template='{{ range .users }}{{ if .user.token }}{{ print .name }}{{ end }}{{ end }}' )
                # username is in form username/cluster - strip off the cluster part
                username=$( echo "$username" | sed 's,/.*$,,' )
            fi
        fi
        if [ -z "$token" ] ; then
            echo ERROR: could not determine token to use to login to "$1"
            echo please do `oc login -u username -p password` to create a context with a token
            echo OR
            echo set \$PUSH_USER and \$PUSH_PASSWORD and run this script again
            return 1
        fi
        if [ "$username" = "kube:admin" ] ; then
            username=kubeadmin
        fi
    fi
    podman login --tls-verify=false -u "$username" -p "$token" "$1" > /dev/null
}

echo "Setting up port-forwarding to remote registry ..."
oc -n openshift-image-registry port-forward service/image-registry 5000:5000 > pf.log 2>&1 &
forwarding_pid=$!
trap "kill -15 ${forwarding_pid}" EXIT

echo "Login to registry..."
sleep 2
login_to_registry 127.0.0.1:5000 

echo "Pushing image ${IMAGE_TAG} ..."
if podman push --tls-verify=false ${IMAGE_TAG} ; then
    oc -n openshift get imagestreams | grep cluster-logging-operator
fi
