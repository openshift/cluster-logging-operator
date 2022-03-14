#!/bin/bash

set -euo pipefail

if oc -n "openshift-operators-redhat" get deployment loki-operator-controller-manager -o name > /dev/null 2>&1 ; then
    echo loki-operator already deployed
  exit 0
fi

registry_org=$1
version=$2

pushd ../../grafana/loki/operator
  make olm-deploy REGISTRY_ORG=$registry_org VERSION=$version
popd
