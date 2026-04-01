#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Operator"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/operator/*/'}
INCLUDES="*"
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
for dir in $(ls -d $TEST_DIR| grep -E "${INCLUDES}"); do
  os::log::info "=========================================================="
  os::log::info "Starting test of operator $dir"
  os::log::info "=========================================================="
  pushd $dir
  artifact_dir=$ARTIFACT_DIR/operator/$(basename $dir)
  mkdir -p $artifact_dir
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/operator/cleanup.sh $artifact_dir" \
    artifact_dir=$artifact_dir \
    go test --ginkgo.v --ginkgo.no-color \
      --ginkgo.trace  --ginkgo.poll-progress-after=300s --ginkgo.poll-progress-interval=30s \
      --ginkgo.timeout=60m "$dir" | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "operator $dir passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "operator $dir failed"
    os::log::info "======================================================="
  fi
  popd
done
exit $failed
