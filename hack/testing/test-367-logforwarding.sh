#!/bin/bash
# Jira LOG-367 - Log forwarding

set -e

source "$(dirname $0)/../common"

start_seconds=$(date +%s)

TEST_DIR=${TEST_DIR:-'./test/e2e/logforwarding/*/'}
test_artifact_dir="${ARTIFACT_DIR:-"$repo_dir/_output"}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifact_dir ] ; then
  mkdir -p $test_artifact_dir
fi

cleanup(){
  local return_code="$?"
  set +e
  log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  
  if [ "${SKIP_CLEANUP:-false}" == "false" ] ; then
    for item in "crd/elasticsearches.logging.openshift.io" "crd/clusterloggings.logging.openshift.io" "ns/openshift-logging" "ns/openshift-operators-redhat"; do
      oc delete $item --wait=true --ignore-not-found --force --grace-period=0
    done
    for item in "ns/openshift-logging" "ns/openshift-operators-redhat"; do
      try_until_failure "oc get ${item}" "$((1 * $minute))"
    done
  fi
  
  set -e
  exit ${return_code}
}
trap cleanup exit

# deploy_elasticsearch-operator from marketplace under the assumption it is current and CLO does not depend on any changes
if [ -n "${DEPLOY_FROM_MARKETPLACE:-}" ] ; then
  log::info "Deploying elasticsearch-operator from the OLM marketplace"
  deploy_marketplace_operator "openshift-operators-redhat" "elasticsearch-operator"
else
  log::info "Deploying elasticsearch-operator from the vendored manifest"
  deploy_elasticsearch_operator
fi

failed=0
for dir in $(ls -d $TEST_DIR); do
  log::info "=========================================================="
  log::info "Starting test of logforwarding $dir"
  log::info "=========================================================="
  log::info "Deploying cluster-logging-operator"
  deploy_clusterlogging_operator
  artifact_dir=$test_artifact_dir/$(basename $dir)

  tempdir=$(mktemp -d /tmp/elasticsearch-operator-XXXXXXXX)
  get_operator_files $tempdir elasticsearch-operator ${EO_REPO:-openshift} ${EO_BRANCH:-master}
  eo_manifest=${tempdir}/manifests

  mkdir -p $artifact_dir
  GENERATOR_NS="clo-test-$RANDOM"
  if CLEANUP_CMD="$( cd $( dirname ${BASH_SOURCE[0]} ) >/dev/null 2>&1 && pwd )/../../test/e2e/logforwarding/cleanup.sh $artifact_dir $GENERATOR_NS" \
    artifact_dir=$artifact_dir \
    GENERATOR_NS=$GENERATOR_NS \
    ELASTICSEARCH_IMAGE="$(format_elasticsearch_image $eo_manifest)" \
    go test -count=1 -parallel=1 $dir  | tee -a $artifact_dir/test.log ; then
    log::info "======================================================="
    log::info "Logforwarding $dir passed"
    log::info "======================================================="
  else
    failed=1
    log::info "======================================================="
    log::info "Logforwarding $dir failed"
    log::info "======================================================="
  fi
  for ns in "ns/$GENERATOR_NS ns/openshift-logging"; do
    oc delete $ns --ignore-not-found --force --grace-period=0||:
    try_until_failure "oc get $ns" "$((1 * $minute))"
  done
done
exit $failed
