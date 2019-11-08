#!/bin/bash

set -euo pipefail

source "$(dirname $0)/common"

# hack to use some of the comment utils
os::test::junit::declare_suite_start "deploye-elasticsearch-operator"

cleanup(){
	local return_code="$?"	
	os::cleanup::all "${return_code}"
  	exit ${return_code}
}
trap cleanup exit 

if oc -n "openshift-operators-redhat" get deployment elasticsearch-operator -o name > /dev/null 2>&1 ; then
  exit 0
fi
deploy_elasticsearch_operator
