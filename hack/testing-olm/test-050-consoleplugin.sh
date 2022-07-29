#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Consoleplugin"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/consoleplugin/*/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

cleanup(){
  local return_code="$?"

  # Disable console plugin
  oc patch consoles.operator.openshift.io cluster --type=merge --patch '{ "spec": { "plugins": [] } }'

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

# Enable console plugin
oc patch consoles.operator.openshift.io cluster  --type=merge --patch '{ "spec": { "plugins": ["logging-view-plugin"] } }'

failed=0
for dir in $(ls -d $TEST_DIR); do
  os::log::info "=========================================================="
  os::log::info "Starting test of console-plugin: $dir"
  os::log::info "=========================================================="
  artifact_dir=$ARTIFACT_DIR/consoleplugin/$(basename $dir)
  mkdir -p $artifact_dir
  GENERATOR_NS="clo-test-$RANDOM"
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/consoleplugin/cleanup.sh $artifact_dir $GENERATOR_NS" \
    artifact_dir=$artifact_dir \
    GENERATOR_NS=$GENERATOR_NS \
    go test -count=1 -parallel=1 -timeout=60m "$dir" -ginkgo.noColor -ginkgo.trace -ginkgo.slowSpecThreshold=300.0 | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Consoleplugin $dir passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Consoleplugin $dir failed"
    os::log::info "======================================================="
  fi
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    for ns in "ns/$GENERATOR_NS" "clusterlogging/instance" "clusterlogforwarder/instance"; do
      oc delete $ns --ignore-not-found --force --grace-period=0||:
      os::cmd::try_until_failure "oc get $ns" "$((1 * $minute))"
    done
  fi
done
exit $failed
