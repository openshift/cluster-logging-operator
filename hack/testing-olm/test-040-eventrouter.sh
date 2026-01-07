#!/bin/bash

# This is a test suite for the eventrouter

set -euo pipefail

current_dir=$(dirname "${BASH_SOURCE[0]}")
repo_dir=${repodir:-"$current_dir/../.."}
source ${current_dir}/../common

ARTIFACT_DIR=${ARTIFACT_DIR:-$repo_dir/_output}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi
EVENT_ROUTER_VERSION="v0.5.0"
IMAGE_LOGGING_EVENTROUTER=${IMAGE_LOGGING_EVENTROUTER:-"quay.io/openshift-logging/eventrouter:${EVENT_ROUTER_VERSION}"}
EVENT_ROUTER_TEMPLATE=${repo_dir}/hack/eventrouter-template.yaml
MAX_DEPLOY_WAIT_SECONDS=${MAX_DEPLOY_WAIT_SECONDS:-120}
mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] event router e2e test"

deploy_eventrouter() {
    oc process --local -p=SA_NAMESPACE=${LOGGING_NS} -p=IMAGE=${IMAGE_LOGGING_EVENTROUTER} \
        -f $EVENT_ROUTER_TEMPLATE | oc create -n ${LOGGING_NS} -f - 2>&1

    local looptries=${MAX_DEPLOY_WAIT_SECONDS}
    local ii
    for (( ii=0; ii<$looptries; ii++ ))
    do
      if [[ $(oc get pods -l component=eventrouter -n $LOGGING_NS -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; then
          sleep 1
      else
          break
      fi
    done
    if [ $ii -eq $looptries ] ; then
      os::log::error could not start eventrouter pod after $looptries seconds
      oc get pods -l component=eventrouter -n $LOGGING_NS -oyaml
      exit 1
    fi
}

reset_logging(){
  local resources=$(cat <<EOL
    serviceaccount/eventrouter
    clusterrole/event-reader
    clusterrolebinding/event-reader-binding
    configmap/eventrouter
    deployment/eventrouter
EOL
)
    oc -n $LOGGING_NS delete $resources --ignore-not-found --force --grace-period=0||:
}

cleanup() {
    local return_code="$?"
    local test_name=$(basename $0)

    os::test::junit::declare_suite_end
    set +e
    if [ "true" == "${DO_CLEANUP:-"true"}" ] ; then
      os::log::info "Running cleanup"
      if [ "$return_code" != "0" ] ; then
	      gather_logging_resources ${LOGGING_NS} $test_artifactdir
      fi

      oc process -p SA_NAMESPACE=${LOGGING_NS} -p IMAGE=${IMAGE_LOGGING_EVENTROUTER} \
         -f $EVENT_ROUTER_TEMPLATE | oc -n ${LOGGING_NS} delete -f -

      reset_logging

      os::cleanup::all "${return_code}"
    fi

    set -e
    exit $return_code
}
trap "cleanup" EXIT

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

reset_logging
deploy_eventrouter
