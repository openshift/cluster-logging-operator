#!/bin/bash

set -e

source "$(dirname $0)/common"

registry_ip=$(oc get service docker-registry -n default -o jsonpath={.spec.clusterIP})
cat manifests/05-deployment.yaml | \
    sed -e "s,quay.io/openshift/cluster-logging-operator,${registry_ip}:5000/openshift/cluster-logging-operator," | \
	oc create -f -
