#!/bin/bash
set -e

source "$(dirname $0)/common"

repo_dir="$(dirname $0)/.."
if ! oc get project openshift-logging > /dev/null 2>&1 ; then
    oc create -f manifests/01-namespace.yaml
fi

manifest=$(mktemp)
files="02-service-account.yaml 03-role.yaml 04-role-binding.yaml 05-deployment.yaml"
pushd manifests; 
  for f in ${files}; do 
     cat ${f} >> ${manifest}; 
  done; 
popd

oc adm policy add-scc-to-user privileged -z fluentd -n openshift-logging

operator-sdk test local \
  --namespace openshift-logging \
  ./test/e2e \
  --namespaced-manifest ${manifest} \
  --global-manifest  ${repo_dir}/manifests/05-crd.yaml
