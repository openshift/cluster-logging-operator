#! /bin/bash

set -euxo pipefail

cp /tmp/test/config /tmp/kubeconfig
export KUBECONFIG=/tmp/kubeconfig

manifest=$(mktemp)
files="01-service-account.yaml 02-role.yaml 03-role-bindings.yaml 05-deployment.yaml"
pushd manifests;
  for f in ${files}; do
     cat ${f} >> ${manifest};
  done;
popd
# update the manifest with the image built by ci
sed -i "s,quay.io/openshift/origin-elasticsearch-operator:latest,${IMAGE_ELASTICSEARCH_OPERATOR}," ${manifest}
sed -i "s/namespace: openshift-logging/namespace: ${TEST_NAMESPACE}/g" ${manifest}

go test ./test/e2e/... \
  -root=$(pwd) \
  -kubeconfig=/tmp/kubeconfig \
  -globalMan manifests/04-crd.yaml \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace \
  -timeout 900s
