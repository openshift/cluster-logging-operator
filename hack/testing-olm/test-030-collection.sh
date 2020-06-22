#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/collection/*/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

cleanup(){
  local return_code="$?"
  set +e
  log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
    ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
  fi
  
  set -e
  exit ${return_code}
}
trap cleanup exit

failed=0
for dir in $(ls -d $TEST_DIR); do
  log::info "=========================================================="
  log::info "Starting test of collection $dir"
  log::info "=========================================================="
  log::info "Deploying cluster-logging-operator"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh
  artifact_dir=$ARTIFACT_DIR/$(basename $dir)

  mkdir -p $artifact_dir
  GENERATOR_NS="clo-test-$RANDOM"
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/collection/cleanup.sh $artifact_dir $GENERATOR_NS" \
    artifact_dir=$artifact_dir \
    GENERATOR_NS=$GENERATOR_NS \
    go test -v -count=1 -parallel=1 -timeout=60m $dir -ginkgo.v  | tee -a $artifact_dir/test.log ; then
    log::info "======================================================="
    log::info "Collection $dir passed"
    log::info "======================================================="
  else
    failed=1
    log::info "======================================================="
    log::info "Collection $dir failed"
    log::info "======================================================="
  fi
  for ns in "ns/$GENERATOR_NS ns/openshift-logging"; do
    oc delete $ns --ignore-not-found --force --grace-period=0||:
    try_until_failure "oc get $ns" "$((1 * $minute))"
  done
done
exit $failed
