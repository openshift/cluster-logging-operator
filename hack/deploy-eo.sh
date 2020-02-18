#!/bin/bash

set -euo pipefail
if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi

source "$(dirname $0)/common"

if oc -n "openshift-operators-redhat" get deployment elasticsearch-operator -o name > /dev/null 2>&1 ; then
  exit 0
fi
deploy_elasticsearch_operator
