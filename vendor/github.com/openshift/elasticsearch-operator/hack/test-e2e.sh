#!/bin/bash
set -euo pipefail

if [ -n "${DEBUG:-}" ]; then
    set -x
fi

repo_dir="$(dirname $0)/.."

manifest=$(mktemp)
files="01-service-account.yaml 02-role.yaml 03-role-bindings.yaml 05-deployment.yaml"
pushd manifests;
  for f in ${files}; do
     cat ${f} >> ${manifest};
  done;
popd

sudo sysctl -w vm.max_map_count=262144

oc create -n openshift-logging -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheusrule.crd.yaml
oc create -n openshift-logging -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/servicemonitor.crd.yaml

operator-sdk test local \
  --namespace openshift-logging \
  --namespaced-manifest ${manifest} \
  --global-manifest manifests/04-crd.yaml \
  ./test/e2e
