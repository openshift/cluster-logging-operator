#!/bin/bash

source "$(dirname $0)/common"

for csv in ${CSV_FILE} ${EO_CSV_FILE}; do
  $repo_dir/hack/gen-olm-artifacts.sh ${csv} ${NAMESPACE} 'all' \
  | oc delete -f - --ignore-not-found
done

oc delete -n openshift is origin-cluster-logging-operator || :
oc delete -n openshift bc cluster-logging-operator || :
