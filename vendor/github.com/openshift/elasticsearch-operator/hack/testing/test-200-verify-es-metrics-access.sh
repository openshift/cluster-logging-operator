#!/bin/bash
# This test verifies only serviceaccounts with the desired rolebindings are 
# allowed to retrieve metrices from elasticsearch
set -euo pipefail

if [ -n "${DEBUG:-}" ]; then
    set -x
fi

repo_dir="$(dirname $0)/../.."
source "${repo_dir}/hack/testing/utils"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}/$(basename ${BASH_SOURCE[0]})

if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

LOGGING_NS=${LOGGING_NS:-openshift-logging}
suffix=$RANDOM
UNAUTHORIZED_SA="unauthorized-sa-${suffix}"
AUTHORIZED_SA="authorized-sa-${suffix}"
CLUSTERROLE="prometheus-k8s-${suffix}"

function cleanup() {
    local result_code="$?"
    set +e

    if [ "${DO_CLEANUP:-true}" == "true" ] ; then  
      oc -n ${LOGGING_NS} exec $(oc ${LOGGING_NS} get pods -l component=elasticsearch -o jsonpath={.items[0].metadata.name}) -c elasticsearch -- cat /tmp/metrics.txt ||: >$ARTIFACT_DIR/metrics.log
      oc -n ${LOGGING_NS} get configmap/elasticsearch  -o jsonpath={.data}> $ARTIFACT_DIR/configmap-elasticsearch.log 2>&1
      get_all_logging_pod_logs ${LOGGING_NS} $ARTIFACT_DIR

      for name in "ns/openshift-operators-redhat" "ns/${LOGGING_NS}" ; do
          oc delete ${name}  > $ARTIFACT_DIR/cleanup.log 2>&1
          try_until_failure "oc get ${name}" "$((1 * $minute))"
      done
      oc delete clusterrole ${CLUSTERROLE} >> $ARTIFACT_DIR/cleanup.log 2>&1 ||:
      oc delete clusterrolebinding ${CLUSTERROLE} >> $ARTIFACT_DIR/cleanup.log 2>&1 ||:
      oc delete clusterrolebinding view-${CLUSTERROLE} >> $ARTIFACT_DIR/cleanup.log 2>&1 ||:
    fi
    set -e
    exit ${result_code}
}
trap cleanup EXIT

if [ "${DO_SETUP:-true}" == "true" ] ; then
  #deploy elasticsearch-operator
  deploy_elasticsearch_operator

  #deploy elasticsearch cluster
  expect_success "oc -n ${LOGGING_NS} create ns ${LOGGING_NS}"
  expect_success "${repo_dir}/hack/deploy-example-secrets.sh  ${LOGGING_NS}"
  expect_success "oc -n ${LOGGING_NS} create -f ${repo_dir}/hack/cr.yaml"

  #wait for pod
  wait_for_deployment_to_be_ready ${LOGGING_NS} $(oc -n ${LOGGING_NS} get deployment -l component=elasticsearch -o jsonpath={.items[0].metadata.name}) $((2 * $minute))
fi

log::info Creating serviceaccounts to verify metrics
oc -n ${LOGGING_NS} create serviceaccount ${UNAUTHORIZED_SA}
oc -n ${LOGGING_NS} create serviceaccount ${AUTHORIZED_SA}

result=$(oc get clusterrole ${CLUSTERROLE} --ignore-not-found ||:)
if [ "$result" == "" ] ; then
  echo "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\", \"kind\":\"ClusterRole\",\"metadata\":{\"name\":\"${CLUSTERROLE}\"},\"rules\":[{\"nonResourceURLs\":[\"/metrics\"],\"verbs\":[\"get\"]}]}" | oc create -f -
fi
result=$(oc get clusterrolebinding ${CLUSTERROLE} --ignore-not-found ||:)
if [ "$result" == "" ] ; then
  log::info Binding ${AUTHORIZED_SA} to be cable of reading metrics
  oc create clusterrolebinding --clusterrole=${CLUSTERROLE} ${CLUSTERROLE} --serviceaccount=${LOGGING_NS}:${AUTHORIZED_SA}
fi
result=$(oc get clusterrolebinding view-${CLUSTERROLE} --ignore-not-found ||:)
if [ "$result" == "" ] ; then
  log::info Binding ${AUTHORIZED_SA} to be cable of getting namespaces
  oc create clusterrolebinding --clusterrole=basic-user view-${CLUSTERROLE} --serviceaccount=${LOGGING_NS}:${AUTHORIZED_SA}
fi
result=$(oc get clusterrolebinding view-${CLUSTERROLE}-unauth --ignore-not-found ||:)
if [ "$result" == "" ] ; then
  log::info Binding ${UNAUTHORIZED_SA} to be cable of getting namespaces
  oc create clusterrolebinding --clusterrole=basic-user view-${CLUSTERROLE}-unauth --serviceaccount=${LOGGING_NS}:${UNAUTHORIZED_SA}
fi

es_pod=$(oc -n ${LOGGING_NS} get pod -l component=elasticsearch -o jsonpath={.items[0].metadata.name})

push_test_script_to_es(){
  es_pod=$1
  token=$2
  service_ip=elasticsearch-metrics.${LOGGING_NS}.svc
  echo "curl -ks -o /tmp/metrics.txt https://${service_ip}:60000/_prometheus/metrics -H Authorization:'Bearer ${token}' -w '%{response_code}\n'" > /tmp/test
  expect_success "oc -n ${LOGGING_NS} cp /tmp/test ${es_pod}:/tmp/test -c elasticsearch"
  expect_success "oc -n ${LOGGING_NS} exec ${es_pod} -c elasticsearch -- chmod 777 /tmp/test"
}

log::info Checking ${UNAUTHORIZED_SA} ability to read metrics through metrics service
token=$(oc -n ${LOGGING_NS} serviceaccounts get-token $UNAUTHORIZED_SA)
push_test_script_to_es $es_pod $token
expect_success_and_text "oc -n ${LOGGING_NS} exec ${es_pod} -c elasticsearch -- bash -c /tmp/test" '403'

log::info Checking ${AUTHORIZED_SA} ability to read metrics
token=$(oc -n ${LOGGING_NS} serviceaccounts get-token $AUTHORIZED_SA)
push_test_script_to_es $es_pod $token
expect_success_and_text "oc -n ${LOGGING_NS} exec ${es_pod} -c elasticsearch -- bash -c /tmp/test" '200'
