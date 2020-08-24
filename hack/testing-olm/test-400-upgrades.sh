#!/bin/bash

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Custom Resource Upgrades"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/upgrades/*/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

cleanup(){
  local return_code="$?"

  os::test::junit::declare_suite_end

  set +e
  log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  
  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
    ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
    ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
  fi

  for n in $(oc get nodes -o name); do
    oc debug $n -- bash -c 'chroot /host && rm -rf /var/lib/fluentd/** || rm /var/log/*.pos || rm /var/log/journal_pos.json || rm /var/log/openshift-apiserver/audit.log.pos || rm /var/log/kube-apiserver/audit.log.pos'
  done
  
  set -e
  exit ${return_code}
}
trap cleanup exit

failed=0
for dir in $(ls -d $TEST_DIR); do
  os::log::info "=========================================================="
  os::log::info "Starting test of upgrades $dir"
  os::log::info "=========================================================="
  os::log::info "Deploying cluster-logging-operator"
  "${repo_dir}"/olm_deploy/scripts/catalog-deploy.sh
  "${repo_dir}"/olm_deploy/scripts/operator-install.sh

  artifact_dir="$ARTIFACT_DIR/upgrades/$(basename $dir)"
  mkdir -p "$artifact_dir"

  GENERATOR_NS="clo-test-$RANDOM"
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/upgrades/cleanup.sh $artifact_dir $GENERATOR_NS" \
    artifact_dir="$artifact_dir" \
    GENERATOR_NS="$GENERATOR_NS" \
    go test -count=1 -parallel=1 -timeout=60m "$dir" -ginkgo.noColor | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Upgrades $dir passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Upgrades $dir failed"
    os::log::info "======================================================="
  fi
  for ns in "ns/$GENERATOR_NS ns/openshift-logging"; do
    oc delete $ns --ignore-not-found --force --grace-period=0||:
    oc::cmd::try_until_failure "oc get $ns" "$((1 * $minute))"
  done
done
exit $failed
