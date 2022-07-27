#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] flowcontrol"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/flowcontrol/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

reset_logging(){
    for r in "clusterlogging/instance" "clusterlogforwarder/instance"; do
      oc delete -n ${LOGGING_NS} $r --ignore-not-found --force --grace-period=0||:
      os::cmd::try_until_failure "oc -n ${LOGGING_NS} get $r" "$((1 * $minute))"
    done
}

cleanup(){
  local return_code="$?"

  os::test::junit::declare_suite_end

  set +e
  if [ "${DO_CLEANUP:-false}" == "true" ] ; then
    os::log::info "Running cleanup"
    if [ "$return_code" != "0" ] ; then
      gather_logging_resources ${LOGGING_NS} $test_artifactdir

      mkdir -p $ARTIFACT_DIR/$test_name
      oc -n $LOGGING_NS get configmap collector -o jsonpath={.data} --ignore-not-found > $ARTIFACT_DIR/$test_name/collector-configmap.log ||:
    fi
    end_seconds=$(date +%s)
    runtime="$(($end_seconds - $start_seconds))s"
  fi
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
reset_logging
for dir in $(ls -d $TEST_DIR); do
  os::log::info "=========================================================="
  os::log::info "Starting test of flowcontrol: $dir"
  os::log::info "=========================================================="
  artifact_dir=$ARTIFACT_DIR/flowcontrol/$(basename $dir)
  mkdir -p $artifact_dir
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/flowcontrol/cleanup.sh $artifact_dir" \
    artifact_dir=$artifact_dir \
    go test -count=1 -parallel=1 -timeout=60m "$dir" -ginkgo.noColor -ginkgo.trace -ginkgo.slowSpecThreshold=300.0 | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Flowcontrol $dir passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Flowcontrol $dir failed"
    os::log::info "======================================================="
  fi
  reset_logging
done
exit $failed
