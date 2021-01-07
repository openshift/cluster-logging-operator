#!/bin/bash

set -euo pipefail

repo_dir=${repo_dir:-$(dirname $0)/../..}
source "$repo_dir/hack/testing-olm/utils"

ARTIFACT_DIR=${ARTIFACT_DIR:-$repo_dir/_output}
LOGGING_NS=${LOGGING_NS:-openshift-logging}

test_name=test-999-fluentd-prometheus-metrics

cleanup() {
  local return_code="$?"
  set +e
  mkdir -p $ARTIFACT_DIR/$test_name
  oc -n $LOGGING_NS get configmap fluentd -o jsonpath={.data} > $ARTIFACT_DIR/$test_name/fluent-configmap.log
  get_all_logging_pod_logs $ARTIFACT_DIR/$test_name

  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
      ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
      oc delete ns openshift-operators-redhat
  fi

  set -e
  exit $return_code
}
trap "cleanup" EXIT

if [ "${DO_SETUP:-true}" == "true" ] ; then
  log::info "Deploying elasticsearch-operator"
  make deploy-elasticsearch-operator
	log::info "Deploying cluster-logging-operator"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh

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
    logs:
      type: "fluentd"
      fluentd: {}
EOL
fi
try_until_text "oc -n ${LOGGING_NS} get pod -l component=elasticsearch -o jsonpath={.items[0].status.phase}" "Running" "$((3 * $minute))"
try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.metadata.name} --ignore-not-found" "fluentd" "$((1 * $minute))"
expectedcollectors=$( oc get nodes | grep -c " Ready " )
try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.status.desiredNumberScheduled}" "${expectedcollectors}"  "$((1 * $minute))"
desired=$(oc -n ${LOGGING_NS} get ds fluentd  -o jsonpath={.status.desiredNumberScheduled})

log::info "Waiting for ${desired} fluent pods to be available...."
try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.status.numberReady}" "$desired" "$((5 * $minute))"

fpod=$(oc -n $LOGGING_NS get pod -l component=fluentd -o jsonpath={.items[0].metadata.name} --ignore-not-found)

log::info "Checking metrics using 'localhost'..."
try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://localhost:24231/metrics" "$((2 * $minute))"

log::info "Checking metrics using fluent service address..."
try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://fluentd.openshift-logging.svc:24231/metrics" "$((2 * $minute))"

fpod_ip="$(oc -n $LOGGING_NS get pod ${fpod} -o jsonpath='{.status.podIP}')"
log::info "Checking metrics using fluent podIP..."
try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://${fpod_ip}:24231/metrics" "$((2 * $minute))"

oc -n $LOGGING_NS exec $fpod -- curl -ks https://${fpod_ip}:24231/metrics >> $ARTIFACT_DIR/${fpod}-metrics-scrape 2>&1 || :
