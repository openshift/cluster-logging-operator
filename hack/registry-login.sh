#!/bin/bash
# Log in to the clusters internal registry. Enable the registry if required.
set -e

# Try to login first.
oc registry login && exit 0

# Need to be logged in with a token to use the registry.
oc whoami -t || { echo 1>&2 "error: you must be logged in with a token, usually as kube:admin"; exit 1; }

# Try to enable the registry.
REGISTRY=configs.imageregistry.operator.openshift.io/cluster

# Make sure the registry has a default route.
if [ "$(oc get $REGISTRY -o=template='{{.spec.defaultRoute}}')" != "true" ]; then
  oc patch $REGISTRY --type merge -p '{"spec":{"defaultRoute":true}}'
fi

# Make sure the registry is managed (SNO clusters start with registry Deleted)
if [ "$(oc get $REGISTRY -o=template='{{.spec.managementState}}')" != "Managed" ]; then
    oc patch $REGISTRY --type merge -p '{"spec":{"managementState":"Managed"}}'
fi

# Make sure the registry has storage (SNO clusters start with no registry storage)
if [ "$(oc get $REGISTRY -o=template='{{.spec.storage}}')" = '<no value>' ]; then
    oc patch $REGISTRY --type merge -p '{"spec":{storage":{"emptyDir":{}}}}'
fi

# Log in to registry, retry till available.
RETRY=0
until oc registry login; do
    [ $(( RETRY++ )) = 50 ] && { echo "error: cannot log in to registry" 1>&2; exit 1; }
    echo "info: retry registry login..." 1>&2
    sleep 2
done
