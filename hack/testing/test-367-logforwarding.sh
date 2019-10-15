#!/bin/bash
# Jira LOG-367 - Log forwarding

set -e
if [ -n "${DEBUG:-}" ]; then
    set -x
fi

source "$(dirname $0)/../common"

start_seconds=$(date +%s)
os::test::junit::declare_suite_start "${BASH_SOURCE[0]}"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

cleanup(){
  os::log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  local return_code="$?"
  set +e
  
  if [ "${SKIP_CLEANUP:-false}" == "false" ] ; then
    for item in "crd/elasticsearches.logging.openshift.io" "crd/clusterloggings.logging.openshift.io" "ns/openshift-logging" "ns/openshift-operators-redhat"; do
      oc delete $item --wait=true --ignore-not-found --force --grace-period=0
    done
    for item in "ns/openshift-logging" "ns/openshift-operators-redhat"; do
      os::cmd::try_until_failure "oc get project ${item}" "$((1 * $minute))"
    done
  fi

  os::cleanup::all "${return_code}"
  
  exit ${return_code}
}
trap cleanup exit

# deploy_elasticsearch-operator from marketplace under the assumption it is current and CLO does not depend on any changes
if [ -n "${DEPLOY_FROM_MARKETPLACE:-}" ] ; then
  os::log::info "Deploying elasticsearch-operator from the OLM marketplace"
  os::cmd::expect_success 'deploy_marketplace_operator "openshift-operators-redhat" "elasticsearch-operator"'
else
  os::log::info "Deploying elasticsearch-operator from the vendored manifest"
  deploy_elasticsearch_operator
fi

os::log::info "Deploying cluster-logging-operator"
deploy_clusterlogging_operator

os::log::info "Staring test of logforwarding"
ELASTICSEARCH_IMAGE=quay.io/openshift/origin-logging-elasticsearch5:latest go test -parallel=1 ./test/e2e/logforwarding  | tee -a $ARTIFACT_DIR/test.log