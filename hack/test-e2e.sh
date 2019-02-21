#!/bin/bash -x
set -e

if [ -n "${IMAGE_CLUSTER_LOGGING_OPERATOR:-}" ] ; then
  source "$(dirname $0)/common"
fi

STABLE_IMAGE_CLUSTER_LOGGING_OPERATOR=$(echo $IMAGE_FORMAT | sed 's,${component},cluster-logging-operator,')
IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-quay.io/openshift/origin-cluster-logging-operator:latest}
IMAGE_MANIFEST_CLUSTER_LOGGING_OPERATOR=${STABLE_IMAGE_CLUSTER_LOGGING_OPERATOR:-$IMAGE_CLUSTER_LOGGING_OPERATOR}

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
# update the manifest with the image built by ci
sed -i "s,quay.io/openshift/origin-cluster-logging-operator:latest,${IMAGE_MANIFEST_CLUSTER_LOGGING_OPERATOR}," ${manifest}

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

TEST_NAMESPACE=${NAMESPACE} go test ./test/e2e/... \
  -root=$(pwd) \
  -kubeconfig=${KUBECONFIG} \
  -globalMan ${global_manifest} \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace
