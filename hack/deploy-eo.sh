#!/bin/bash

set -euo pipefail

if oc -n "openshift-operators-redhat" get deployment elasticsearch-operator -o name > /dev/null 2>&1 ; then
    echo elasticsearch-operator already deployed
  exit 0
fi

pushd ../elasticsearch-operator
  LOCAL_IMAGE_ELASTICSEARCH_OPERATOR_REGISTRY=127.0.0.1:5000/openshift/elasticsearch-operator-registry \
  make elasticsearch-catalog-deploy
  IMAGE_ELASTICSEARCH_OPERATOR_REGISTRY=image-registry.openshift-image-registry.svc:5000/openshift/elasticsearch-operator-registry \
  make -C ../elasticsearch-operator elasticsearch-operator-install
popd
