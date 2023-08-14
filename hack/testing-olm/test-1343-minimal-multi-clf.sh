#!/bin/bash
# Given an OLM manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.

set -eou pipefail

source "$(dirname "${BASH_SOURCE[0]}" )/../lib/init.sh"
source "$(dirname $0)/assertions"

LOGGING_NS=test-logging-$RANDOM

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogForwarder] Minimal test of multi-ClusterLogForwarder"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi

export CLUSTER_LOGGING_OPERATOR_NAMESPACE="openshift-logging"
repo_dir="$(dirname $0)/../.."

cleanup(){
  local return_code="$?"

  set +e
  if [ "$return_code" != "0" ] ; then 
    gather_logging_resources ${LOGGING_NS} $test_artifactdir
  fi

  if [ "${DO_TEST_CLEANUP:-true}" == "true" ] ; then
      oc delete ns $LOGGING_NS --ignore-not-found --force --grace-period=0
  fi

  os::test::junit::declare_suite_end

  set -e
  exit ${return_code}
}
trap cleanup exit

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

if [ "${DO_SETUP:-false}" == "true" ] ; then
  os::log::info "Deploying cluster-logging-operator"
  ${repo_dir}/make deploy
fi

TIMEOUT_MIN=$((2 * $minute))

# wait for operator to be ready
os::cmd::try_until_text "oc -n $CLUSTER_LOGGING_OPERATOR_NAMESPACE get deployment cluster-logging-operator -o jsonpath={.status.availableReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

os::cmd::expect_success "oc create ns $LOGGING_NS"

# deploy ClusterLogForwarder
os::cmd::expect_success "oc -n $LOGGING_NS create sa mine"
os::cmd::expect_success "oc -n $LOGGING_NS create clusterrolebinding collect-application-logs --clusterrole=collect-application-logs --serviceaccount=$LOGGING_NS:mine"
os::cmd::expect_success "oc -n $LOGGING_NS create -f ${repo_dir}/hack/clusterlogforwarder/cr.yaml"

assert_collector_exist my-collector

os::cmd::expect_success "oc -n $LOGGING_NS wait -lcomponent=collector --for=jsonpath='{.status.phase}'=Running pod"