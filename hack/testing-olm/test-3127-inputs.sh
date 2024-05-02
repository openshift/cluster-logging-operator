#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogForwarder] Collection Syslog Inputs"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/inputs/syslog'}
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
  os::log::info "Starting test of collection Syslog Inputs '$dir'"
  os::log::info "=========================================================="
  artifact_dir=$ARTIFACT_DIR/$(basename $dir)
  mkdir -p $artifact_dir
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/inputs/cleanup.sh $artifact_dir" \
    artifact_dir=$artifact_dir \
    go test -count=1 -parallel=1 -timeout=60m "$dir" -ginkgo.noColor -ginkgo.trace -ginkgo.slowSpecThreshold=300.0 | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Collection Syslog Inputs '$dir' passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Collection Syslog Inputs '$dir' failed"
    os::log::info "======================================================="
  fi
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    for obj in $(oc get clusterlogforwarder --all-namespaces -o name); do
      oc delete $obj --ignore-not-found --force --grace-period=0||:
    done
  fi
done
exit $failed
