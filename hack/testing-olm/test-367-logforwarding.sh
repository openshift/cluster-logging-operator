#!/bin/bash
# Jira LOG-367 - Log forwarding

set -euo pipefail

source "$(dirname $0)/../common"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Log Forwarding"

start_seconds=$(date +%s)

GINKGO_OPTS=${GINKGO_OPTS:-""}
TEST_DIR=${TEST_DIR:-'./test/e2e/logforwarding/*/'}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi
CLF_INCLUDES=${CLF_INCLUDES:-}

cleanup(){
  local return_code="$?"

  os::test::junit::declare_suite_end

  set +e
  os::log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"

  if [ "${DO_CLEANUP:-false}" == "true" ] ; then
    ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
    ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
  fi

  set -e
  exit ${return_code}
}
trap cleanup exit

if [ "${DO_SETUP:-false}" == "true" ] ; then
  if [ "${DO_EO_SETUP:-true}" == "true" ] ; then
      pushd ../../elasticsearch-operator
      # install the catalog containing the elasticsearch operator csv
      ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/catalog-deploy.sh
      # install the elasticsearch operator from that catalog
      ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/operator-install.sh
      popd
  fi

  os::log::info "Deploying cluster-logging-operator"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh
fi

reset_logging(){
    oc delete --ignore-not-found --force --grace-period=0 "ns/$GENERATOR_NS" "clusterlogging/instance" "clusterlogforwarder/instance"||:
    oc wait --for=delete "clusterlogging/instance" --timeout=30s
    oc wait --for=delete "clusterlogforwarder/instance" --timeout=30s
}

failed=0
for dir in $(eval echo $TEST_DIR); do
  if [ -n "${CLF_INCLUDES}" ] ; then
    if ! basename $dir | grep -P -q "${CLF_INCLUDES}" ; then
      os::log::info "==============================================================="
	    os::log::info "excluding logforwarding $dir "
	    os::log::info "==============================================================="
      continue
    fi
  fi
  os::log::info "=========================================================="
  os::log::info "Starting test of logforwarding $dir"
  os::log::info "=========================================================="
  artifact_dir=$ARTIFACT_DIR/logforwarding/$(basename $dir)

  mkdir -p /tmp/artifacts/junit
  mkdir -p $artifact_dir
  GENERATOR_NS="clo-test-$RANDOM"
  reset_logging
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/logforwarding/cleanup.sh $artifact_dir $GENERATOR_NS" \
    artifact_dir=$artifact_dir \
    GENERATOR_NS=$GENERATOR_NS \
    SUCCESS_TIMEOUT=10m \
    go test -count=1 -parallel=1 -timeout=90m "$dir" -ginkgo.noColor -ginkgo.trace -ginkgo.slowSpecThreshold=600.0 ${GINKGO_OPTS} | tee -a "$artifact_dir/test.log" ; then
    os::log::info "======================================================="
    os::log::info "Logforwarding $dir passed"
    os::log::info "======================================================="
  else
    failed=1
    os::log::info "======================================================="
    os::log::info "Logforwarding $dir failed"
    os::log::info "======================================================="
  fi
  reset_logging
done
exit $failed
