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
  local return_code="$?"
  set +e
  os::log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  
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

failed=0
for dir in $(ls -d ./test/e2e/logforwarding/*); do
  os::log::info "==========================="
  os::log::info "Staring test of logforwarding $dir"
  os::log::info "==========================="
  os::log::info "Deploying cluster-logging-operator"
  deploy_clusterlogging_operator
  artifact_dir=$ARTIFACT_DIR/$(basename $dir)
  mkdir -p $artifact_dir
  if ELASTICSEARCH_IMAGE=quay.io/openshift/origin-logging-elasticsearch5:latest go test -parallel=1 $dir  | tee -a $artifact_dir/test.log ; then
    os::log::info "==========================="
    os::log::info "Logforwarding $dir passed"
    os::log::info "==========================="
  else
    failed=1
    os::log::info "==========================="
    os::log::info "Logforwarding $dir failed"
    os::log::info "==========================="
    for p in $(oc -n openshift-logging get pods -o jsonpath={.items[*].metadata.name}); do
      oc -n openshift-logging logs $p > $artifact_dir/$p.logs||:
    done
    for p in $(oc -n openshift-logging get pods -lcomponent=fluentd -o jsonpath={.items[*].metadata.name}); do
      oc -n openshift-logging exec $p -- logs > $artifact_dir/$p.logs||:
    done
    for p in $(oc -n openshift-logging get pods -lcomponent=elasticsearch -o jsonpath={.items[*].metadata.name}); do
      oc -n openshift-logging -c elasticsearch exec $p -- logs > $artifact_dir/$p.logs||:
    done
    oc -n openshift-logging get pods > $artifact_dir/pods||:
    oc -n openshift-logging describe ds fluentd > $artifact_dir/fluent.describe||:
    oc -n openshift-logging configmap fluentd -o yaml > $artifact_dir/fluent-configmap.yaml||:
  fi
  oc delete ns openshift-logging --wait=true --ignore-not-found --force --grace-period=0
  os::cmd::try_until_failure "oc get project openshift-logging" "$((1 * $minute))"
done
exit $failed
