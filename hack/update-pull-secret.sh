#!/bin/bash
set -e

PULL_SECRET=$1
test -n "$PULL_SECRET" || { echo "Specify a pull secret file"; exit 1; }
BASE64=$(cat $PULL_SECRET | tr -d '[:space:]' | base64 -w0)

secret_yaml() {
    cat <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: pull-secret
  namespace: openshift-config
data:
  .dockerconfigjson: $BASE64
EOF
}

secret_yaml | oc delete --ignore-not-found -f -
secret_yaml | oc create -f -
