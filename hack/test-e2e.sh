#!/bin/bash -x
set -e

ARTIFACT_DIR=${ARTIFACT_DIR:-_output}

rm -rf $ARTIFACT_DIR
mkdir $ARTIFACT_DIR

manifest=$(mktemp)
global_manifest=$(mktemp)

cleanup(){
  local return_code="$?"
  set +e

  cat $manifest > $ARTIFACT_DIR/manifest
  cat $global_manifest > $ARTIFACT_DIR/global_manifest
  
  exit $return_code
}
trap cleanup exit

if [ -n "${IMAGE_CLUSTER_LOGGING_OPERATOR:-}" ] ; then
  source "$(dirname $0)/common"
fi
if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_CLUSTER_LOGGING_OPERATOR=$(sed -e "s,\${component},cluster-logging-operator," <(echo $IMAGE_FORMAT))
fi
IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-quay.io/openshift/origin-cluster-logging-operator:latest}

repo_dir="$(dirname $0)/.."
$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'ns' | oc create -f -

$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} >> ${manifest}
sed -i "s,quay.io/openshift/origin-cluster-logging-operator:latest,${IMAGE_CLUSTER_LOGGING_OPERATOR}," ${manifest}

$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'crd' >> ${global_manifest}
$repo_dir/hack/gen-olm-artifacts.sh ${EO_CSV_FILE} ${NAMESPACE} 'crd' >> ${global_manifest}
$repo_dir/hack/gen-olm-artifacts.sh ${EO_CSV_FILE} ${NAMESPACE} >> ${manifest}

TEST_NAMESPACE=${NAMESPACE} go test ./test/e2e/... \
  -root=$(pwd) \
  -kubeconfig=${KUBECONFIG} \
  -globalMan ${global_manifest} \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace
