#!/bin/bash

set -e

source "$(dirname $0)/../common"

start_seconds=$(date +%s)

LOG_LEVEL=${LOG_LEVEL:-}
TEST_DIR=${TEST_DIR:-'./test/e2e/collection/*/'}
test_name=$(basename ${BASH_SOURCE[0]} .sh)
test_artifact_dir="${ARTIFACT_DIR:-"$repo_dir/_output"}/$test_name"
if [ ! -d $test_artifact_dir ] ; then
  mkdir -p $test_artifact_dir
fi

cleanup(){
  local return_code="$?"
  set +e
  log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  
  for item in "ns/openshift-logging" "ns/openshift-operators-redhat"; do
    oc delete $item --wait=true --ignore-not-found --force --grace-period=0
  done
  for item in "ns/openshift-logging" "ns/openshift-operators-redhat"; do
    try_until_failure "oc get ${item}" "$((1 * $minute))"
  done
  
  set -e
  exit ${return_code}
}

if [ "${DO_CLEANUP:-true}" == "true" ] ; then
  trap cleanup exit
fi

failed=0
for dir in $(ls -d $TEST_DIR); do
  log::info "=========================================================="
  log::info "Starting $test_name: $dir"
  log::info "=========================================================="
  
  if [ "${DO_SETUP:-true}" == "true" ] ; then
    log::info "Deploying elasticsearch-operator"
    deploy_elasticsearch_operator
    
    log::info "Deploying cluster-logging-operator"
    deploy_clusterlogging_operator
  fi
  artifact_dir=$test_artifact_dir/$(basename $dir)

  mkdir -p $artifact_dir
  GENERATOR_NS="clo-test-$RANDOM"
  log::info Using ns/$GENERATOR_NS for log generator
  # cleanup for individual golang test
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/logforwarding/cleanup.sh $artifact_dir $GENERATOR_NS"
  fi

  log::info "Running 'CLEANUP_CMD=${CLEANUP_CMD:-} artifact_dir=$artifact_dir  GENERATOR_NS=$GENERATOR_NS  LOG_LEVEL=${LOG_LEVEL:-}  go test -count=1 -parallel=1 $dir  | tee -a $artifact_dir/test.log'"
  if CLEANUP_CMD="${CLEANUP_CMD:-}" \
    artifact_dir=$artifact_dir \
    GENERATOR_NS=$GENERATOR_NS \
    LOG_LEVEL=${LOG_LEVEL:-} \
    go test -count=1 -parallel=1 $dir  | tee -a $artifact_dir/test.log ; then
    log::info "======================================================="
    log::info "$test_name: $dir passed"
    log::info "======================================================="
  else
    failed=1
    log::info "======================================================="
    log::info "$test_name: $dir failed"
    log::info "======================================================="
  fi
  #cleanup test namespaces
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    for ns in "ns/$GENERATOR_NS ns/openshift-logging"; do
      oc delete $ns --ignore-not-found --force --grace-period=0||:
      try_until_failure "oc get $ns" "$((1 * $minute))"
    done
  fi
done
exit $failed
