#!/bin/bash

current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/lib/init.sh"
source "${current_dir}/lib/util/logs.sh"
source "${current_dir}/testing-olm/utils"

get_setup_artifacts=true
CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}
TEST_DIR=${TEST_DIR:-'./test/upgrade/*/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
repo_dir="$(dirname $0)/.."
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi

cleanup(){
  local return_code="$?"
  
  set +e
  if [ "$return_code" != "0" -a $get_setup_artifacts ] ; then 
    oc get all -n openshift-operators-redhat > $test_artifactdir/openshift-operators-redhat_all.txt
    gather_logging_resources ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} $test_artifactdir
  fi

  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      os::log::info "Deleting operator namespace: ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}"
      oc delete namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} --force
  fi
  os::cleanup::all "${return_code}"
  set -e
  exit ${return_code}
}
trap cleanup exit

if [ "${DO_SETUP:-true}" == "true" ] ; then
  os::log::info "Creating operator namespace: ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}"
  oc create namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
  os::log::info "Deploying cluster-logging-operator catalog"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
fi

get_setup_artifacts=false
export JUNIT_REPORT_OUTPUT="/tmp/artifacts/junit/test-upgrade"
failed=0
os::log::info "Running tests from directory: $TEST_DIR"
for dir in $(ls -d $TEST_DIR); do
  os::log::info "=========================================================="
  os::log::info "Starting test of upgrade $dir"
  os::log::info "=========================================================="
  artifact_dir=$ARTIFACT_DIR/upgrade/$(basename $dir)
  mkdir -p $artifact_dir
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../test/e2e/collection/cleanup.sh $artifact_dir" \
    go test -count=1 -parallel=1 -timeout=15m "$dir" -ginkgo.noColor -ginkgo.trace -ginkgo.slowSpecThreshold=300.0 | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Upgrade $dir passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Upgrade $dir failed"
    os::log::info "======================================================="
  fi
done

ARTIFACT_DIR="/tmp/artifacts/junit/" os::test::junit::generate_report ||:

if [[ -n "${failed:-}" ]]; then
    exit 1
fi
