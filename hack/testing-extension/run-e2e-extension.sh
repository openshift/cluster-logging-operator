#!/bin/bash
# Acceptance testing

set -euo pipefail
source "$(dirname $0)/env.sh"
source "$(dirname $0)/../common"
source "$(dirname $0)/assertions"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] ${TEST_SUITE}"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-"${repo_dir}/test/test-extension"}
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

failed=0
os::log::info "=========================================================="
os::log::info "Starting test ${TEST_SUITE}"
os::log::info "=========================================================="
pushd ${TEST_DIR}
if go run cmd/main.go  run-suite "${TEST_SUITE}"  |tee ${ARTIFACT_DIR}/test.log; then
  os::log::info "======================================================="
  os::log::info "Test ${TEST_SUITE} end"
  os::log::info "======================================================="
  exit 0
else
  os::log::info "======================================================="
  os::log::info "Test ${TEST_SUITE} end"
  os::log::info "======================================================="
  exit $failed
fi
