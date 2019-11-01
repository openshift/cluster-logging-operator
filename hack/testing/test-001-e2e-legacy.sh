#!/bin/bash

# This is the legacy e2e test.
# Given a CVO like manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.
# Given the CLO is provided with new operand images, verify that it
# generates new operand resources that reference the upgrade images

set -e

repo_dir="$( cd "$(dirname "$0")/../.." ; pwd -P )"
source "${repo_dir}/hack/lib/init.sh"

set -x
NAMESPACE=openshift-logging

os::test::junit::declare_suite_start "${BASH_SOURCE[0]}"

ARTIFACT_DIR=${ARTIFACT_DIR:-_output}
OCP_VERSION=${OCP_VERSION:-4.3}
manifest=$(mktemp)
global_manifest=$(mktemp)
CSV_FILE="${CSV_FILE:-$repo_dir/manifests/$OCP_VERSION/cluster-logging.v$OCP_VERSION.0.clusterserviceversion.yaml}"
EO_CSV_FILE="${EO_CSV_FILE:-$repo_dir/vendor/github.com/openshift/elasticsearch-operator/manifests/$OCP_VERSION/elasticsearch-operator.v$OCP_VERSION.0.clusterserviceversion.yaml}"

cleanup(){
  local return_code="$?"
  set +e

  oc logs deployment/cluster-logging-operator > $ARTIFACT_DIR/operator.logs
  cat $manifest > $ARTIFACT_DIR/manifest
  cat $global_manifest > $ARTIFACT_DIR/global_manifest

  oc delete -f $manifest --ignore-not-found --wait=true
  oc delete -f $global_manifest --ignore-not-found --wait=true

  oc::cmd::try_until_failure "oc delete project $NAMESPACE"

  os::cleanup::all "${return_code}"
  
  exit $return_code
}
trap cleanup exit

if [ -n "${IMAGE_CLUSTER_LOGGING_OPERATOR:-}" ] ; then
  source "$repo_dir/hack/common"
fi
if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_CLUSTER_LOGGING_OPERATOR=$(sed -e "s,\${component},cluster-logging-operator," <(echo $IMAGE_FORMAT))
else
  IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-registry.svc.ci.openshift.org/origin/${OCP_VERSION}:cluster-logging-operator}
fi

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}


$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'ns' | oc create -f - || :

$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} >> ${manifest}
sed -i "s,quay.io/openshift/origin-cluster-logging-operator:latest,${IMAGE_CLUSTER_LOGGING_OPERATOR}," ${manifest}
if [ -n "${IMAGE_FORMAT:-}" ] ; then
  for comp in logging-curator5 logging-elasticsearch5 logging-fluentd logging-kibana5 logging-oauth-proxy ; do
    img=$(sed -e "s,\${component},$comp," <(echo $IMAGE_FORMAT))
    sed -i "s,quay.io/openshift/origin-${comp}:latest,${img}," ${manifest}
  done
fi

$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'crd' >> ${global_manifest}
$repo_dir/hack/gen-olm-artifacts.sh ${EO_CSV_FILE} ${NAMESPACE} 'crd' >> ${global_manifest}
$repo_dir/hack/gen-olm-artifacts.sh ${EO_CSV_FILE} ${NAMESPACE} >> ${manifest}

export LOG_LEVEL=debug

TEST_NAMESPACE=${NAMESPACE} go test ./test/e2e/ \
  -root=$(pwd) \
  -kubeconfig=${KUBECONFIG} \
  -globalMan ${global_manifest} \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace | tee -a $ARTIFACT_DIR/test.log
