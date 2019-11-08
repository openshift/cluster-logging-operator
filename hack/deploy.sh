#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"
os::test::junit::declare_suite_start "deploy-elasticsearch-operator"

cleanup(){
	local return_code="$?"	
	os::cleanup::all "${return_code}"
  	exit ${return_code}
}
trap cleanup exit 
IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
deploy_clusterlogging_operator
