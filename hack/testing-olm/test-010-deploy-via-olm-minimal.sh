#!/bin/bash
# Given an OLM manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.

set -eou pipefail

source "$(dirname "${BASH_SOURCE[0]}" )/../lib/init.sh"
source "$(dirname $0)/assertions"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Deploy via OLM minimal"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi

export CLUSTER_LOGGING_OPERATOR_NAMESPACE="openshift-logging"
NAMESPACE="${CLUSTER_LOGGING_OPERATOR_NAMESPACE}"
repo_dir="$(dirname $0)/../.."
manifest=${repo_dir}/manifests
version=$(basename $(find $manifest -type d | sort -r | head -n 1))

cleanup(){
  local return_code="$?"

  os::test::junit::declare_suite_end
  
  set +e
  if [ "$return_code" != "0" ] ; then 
    gather_logging_resources ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} $test_artifactdir
  fi

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

# test the validation of an invalid cr
os::cmd::expect_failure_and_text "oc -n $NAMESPACE create -f ${repo_dir}/hack/cr_invalid.yaml" "invalid: metadata.name: Unsupported value"

# deploy cluster logging
os::cmd::expect_success "oc -n $NAMESPACE create -f ${repo_dir}/hack/cr.yaml"

# assert deployment
assert_resources_exist
# assert kibana instance exists
assert_kibana_instance_exists

# delete cluster logging
os::cmd::expect_success "oc -n $NAMESPACE delete -f ${repo_dir}/hack/cr.yaml"

# deploy cluster logging with unmanaged state
os::cmd::expect_success "oc -n $NAMESPACE create -f ${repo_dir}/hack/cr-unmanaged.yaml"

# wait few seconds
sleep 10
# assert does not exist
assert_resources_does_not_exist
# assert kibana instance does not exists
assert_kibana_instance_does_not_exists
