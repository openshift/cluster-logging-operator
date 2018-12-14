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

global_manifest=$(mktemp)
global_files="05-crd.yaml"
elasticsearch_files="04-crd.yaml"
pushd manifests;
  for f in ${global_files}; do
    cat ${f} >> ${global_manifest}
  done;
popd
pushd vendor/github.com/openshift/elasticsearch-operator/manifests;
  for f in ${elasticsearch_files}; do
    cat ${f} >> ${global_manifest}
  done;
popd

# allows fluentd to have root access to node
oc adm policy add-scc-to-user privileged -z fluentd -n openshift-logging
# allows fluentd to query k8s api for all namespace/pod info
oc adm policy add-cluster-role-to-user cluster-reader -z fluentd -n openshift-logging
# allows rsyslog to have root access to node
oc adm policy add-scc-to-user privileged -z rsyslog -n openshift-logging
# allows rsyslog to query k8s api for all namespace/pod info
oc adm policy add-cluster-role-to-user cluster-reader -z rsyslog -n openshift-logging

TEST_NAMESPACE=openshift-logging go test ./test/e2e/... \
  -root=$(pwd) \
  -kubeconfig=$HOME/.kube/config \
  -globalMan ${global_manifest} \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace
