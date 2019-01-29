#!/bin/bash

set -euxo pipefail

source "$(dirname $0)/common"

for repo in ${repo_dir}; do
  oc delete -f ${repo}/manifests --ignore-not-found
done

oc delete -n openshift is elasticsearch-operator || :
oc delete -n openshift bc elasticsearch-operator || :

oc delete namespace ${NAMESPACE} ||:
