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

if oc get project openshift-logging > /dev/null 2>&1 ; then
  echo using existing project openshift-logging
else
  oc create namespace openshift-logging
fi

oc create -n openshift-logging -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheusrule.crd.yaml || :
oc create -n openshift-logging -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/servicemonitor.crd.yaml || :

TEST_NAMESPACE=openshift-logging go test ./test/e2e/... \
  -root=$(pwd) \
  -kubeconfig=$HOME/.kube/config \
  -globalMan manifests/04-crd.yaml \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace
