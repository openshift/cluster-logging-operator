#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogForwarder] Collection Input Selection"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/input_selection/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

cleanup(){
  local return_code="$?"

  os::test::junit::declare_suite_end

  set +e
  os::log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"

  set -e
  exit ${return_code}
}
trap cleanup exit

if [ "${DO_SETUP:-false}" == "true" ] ; then
  os::log::info "Deploying cluster-logging-operator"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh
fi

failed=0
for dir in $(ls -d $TEST_DIR); do
  os::log::info "=========================================================="
  os::log::info "Starting test of collection input selection '$dir'"
  os::log::info "=========================================================="
  artifact_dir=$ARTIFACT_DIR/$(basename $dir)
  mkdir -p $artifact_dir
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/input_selection/cleanup.sh $artifact_dir" \
    artifact_dir=$artifact_dir \
    ginkgo -p -procs=1 -v --no-color --trace --poll-progress-after=300s --poll-progress-interval=30s  --timeout=60m "$dir" | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Collection Input Selection '$dir' passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Collection Input Selection '$dir' failed"
    os::log::info "======================================================="
  fi
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    for obj in $(oc get clusterlogforwarder --all-namespaces -o name); do
      oc delete $obj --ignore-not-found --force --grace-period=0||:
    done
  fi
done
exit $failed
