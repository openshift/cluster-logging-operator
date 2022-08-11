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
EVENT_ROUTER_VERSION="0.3"
IMAGE_LOGGING_EVENTROUTER=${IMAGE_LOGGING_EVENTROUTER:-"quay.io/openshift-logging/eventrouter:${EVENT_ROUTER_VERSION}"}
EVENT_ROUTER_TEMPLATE=${repo_dir}/hack/eventrouter-template.yaml
MAX_DEPLOY_WAIT_SECONDS=${MAX_DEPLOY_WAIT_SECONDS:-120}
GENERATOR_NS="clo-eventrouter-test-$RANDOM"
mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] event router e2e test"

if [ "${DO_SETUP:-false}" == "true" ] ; then
  if [ "${DO_EO_SETUP:-true}" == "true" ] ; then
      pushd ../../elasticsearch-operator
      # install the catalog containing the elasticsearch operator csv
      ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/catalog-deploy.sh
      # install the elasticsearch operator from that catalog
      ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/operator-install.sh
      popdgz
  fi

  os::log::info "Deploying cluster-logging-operator"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh
fi

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

wait_elasticsearch() {
  local tries=90
  local i
  for (( i=0; i<$tries; i++ ))
    do
       os::log::info "Checking deployment of elasticsearch... Attempt " $i
       if [[ $(oc get pods -l component=elasticsearch -n $LOGGING_NS -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; then
          sleep 2
       else
          break
       fi
    done
  if [ $i -eq $tries ] ; then
    os::log::error could not start elasticsearch pod after $tries*2 seconds
    oc get pods -l component=elasticsearch -n $LOGGING_NS -oyaml
    exit 1
  fi
}

reset_logging(){
    for r in "ns/$GENERATOR_NS" "clusterlogging/instance" "clusterlogforwarder/instance"; do
      oc delete $r --ignore-not-found --force --grace-period=0||:
      os::cmd::try_until_failure "oc get $r" "$((1 * $minute))"
    done
}

cleanup() {
    local return_code="$?"
    local test_name=$(basename $0)


    if [ "true" == "${DO_CLEANUP:-"true"}" ] ; then
      os::test::junit::declare_suite_end
      set +e
      if [ "$return_code" != "0" ] ; then
	      gather_logging_resources ${LOGGING_NS} $test_artifactdir

        mkdir -p $ARTIFACT_DIR/$test_name
        oc -n $LOGGING_NS get configmap fluentd -o jsonpath={.data} --ignore-not-found > $ARTIFACT_DIR/$test_name/fluent-configmap.log ||:
      fi

      oc process -p SA_NAMESPACE=${LOGGING_NS} -p IMAGE=${IMAGE_LOGGING_EVENTROUTER} \
         -f $EVENT_ROUTER_TEMPLATE | oc -n ${LOGGING_NS} delete -f -

      reset_logging

      os::cleanup::all "${return_code}"
    fi

    exit $return_code
}
trap "cleanup" EXIT

function warn_nonformatted() {
    local index=$1
    # check if eventrouter and fluentd with correct ViaQ plugin are deployed
    local non_formatted_event_count=$(oc -n $LOGGING_NS exec -c elasticsearch $espod -- es_util --query="$index/_count?q=verb:*" | jq .count )
    if [ "$non_formatted_event_count" != 0 ]; then
        os::log::warning "$non_formatted_event_count events from eventrouter in index $index were not processed by ViaQ fluentd plugin"
    else
        os::log::info "good - looks like all eventrouter events were processed by fluentd"
    fi
}

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

reset_logging
cat <<EOL | oc -n ${LOGGING_NS} create -f -
apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogging"
metadata:
  name: "instance"
spec:
  managementState: "Managed"
  logStore:
    type: "elasticsearch"
    elasticsearch:
      nodeCount: 1
      redundancyPolicy: "ZeroRedundancy"
      resources:
        request:
          memory: 1Gi
          cpu: 100m
  collection:
    type: "fluentd"
EOL

deploy_eventrouter
wait_elasticsearch
espod=$(oc -n $LOGGING_NS get pods -l component=elasticsearch -o jsonpath={.items[0].metadata.name})
os::log::info "Testing Elasticsearch pod ${espod}..."
os::cmd::try_until_text "oc -n $LOGGING_NS exec -c elasticsearch ${espod} -- es_util --query=/ --request HEAD --head --output /dev/null --write-out %{response_code}" "200" "$(( 1*$minute ))"

warn_nonformatted 'infra-*'

evpod=$(oc -n $LOGGING_NS get pods -l component=eventrouter -o jsonpath={.items[0].metadata.name})

os::log::info "Checking if 1) the doc _id is the same as the kube id 2) there's no duplicates"
hit_count=10
os::cmd::try_until_text "oc -n $LOGGING_NS exec -c elasticsearch $espod -- es_util --query='_search?pretty&q=kubernetes.event:*&size=$hit_count' | jq -r '.hits.hits | length'" $hit_count "$((10 * $minute))" 30
oc -n $LOGGING_NS exec -c elasticsearch $espod -- es_util --query='_search?pretty&q=kubernetes.event:*&size=9999' > $ARTIFACT_DIR/id-dup-search-raw.json
cat $ARTIFACT_DIR/id-dup-search-raw.json | jq -r '.hits.hits[] | ._id + " " + ._source.kubernetes.event.metadata.uid' | sort > $ARTIFACT_DIR/id-and-uid
if ! $(test -s "$ARTIFACT_DIR/id-and-uid") ; then
    os::log::error "Fail: no events found"
    exit 1
fi
cat $ARTIFACT_DIR/id-and-uid | awk '{
    if ($1 != $2) {print "Error: es _id " $1 " not equal to kube uid " $2; exit 1}
    if ($1 == last1) {print "Error: found duplicate es _id " $1; exit 1}
    if ($2 == last2) {print "Error: found duplicate kube uid " $2; exit 1}
    last1 = $1; last2 = $2
}'
