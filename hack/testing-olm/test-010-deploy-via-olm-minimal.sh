#!/bin/bash
# Given an OLM manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.

set -eou pipefail

source "$(dirname "${BASH_SOURCE[0]}" )/../lib/init.sh"
source "$(dirname $0)/assertions"

os::test::junit::declare_suite_start "${BASH_SOURCE[0]}"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi

export CLUSTER_LOGGING_OPERATOR_NAMESPACE="openshift-logging"
repo_dir="$(dirname $0)/../.."
manifest=${repo_dir}/manifests
version=$(basename $(find $manifest -type d | sort -r | head -n 1))

cleanup(){
  local return_code="$?"
  set +e
  gather_logging_resources ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} $test_artifactdir

  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
      ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
      os::cleanup::all "${return_code}"
  fi
  
  set -e
  exit ${return_code}
}
trap cleanup exit

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
${repo_dir}/olm_deploy/scripts/operator-install.sh

TIMEOUT_MIN=$((2 * $minute))

##verify metrics rbac
# extra resources not support for ConfigMap based catelogs for now.
#os::cmd::expect_success "oc get clusterrole clusterlogging-collector-metrics"
#os::cmd::expect_success "oc get clusterrolebinding clusterlogging-collector-metrics"

# wait for operator to be ready
os::cmd::try_until_text "oc -n $NAMESPACE get deployment cluster-logging-operator -o jsonpath={.status.availableReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# deploy cluster logging
os::cmd::expect_success "oc -n $NAMESPACE create -f ${repo_dir}/hack/cr.yaml"

# assert deployment
assert_resources_exist
