#!/bin/bash

set -e

repo_dir=${repo_dir:-$(dirname $0)/../..}
source "$repo_dir/hack/testing/utils"

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
		for item in "crd/elasticsearches.logging.openshift.io" "crd/clusterloggings.logging.openshift.io" "ns/openshift-logging" "ns/openshift-operators-redhat"; do
		oc delete $item --wait=true --ignore-not-found --force --grace-period=0
		done
		for item in "ns/openshift-logging" "ns/openshift-operators-redhat"; do
		try_until_failure "oc get ${item}" "$((1 * $minute))"
		done
	fi

    cleanup_olm_catalog_unsupported_resources
    set -e
    exit $return_code
}
trap "cleanup" EXIT

if [ "${DO_SETUP:-true}" == "true" ] ; then
	log::info "Deploying elasticsearch-operator from the vendored manifest"
	deploy_elasticsearch_operator
	log::info "Deploying cluster-logging-operator"
	deploy_clusterlogging_operator
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
try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.metadata.name} --ignore-not-found" "fluentd" "$((1 * $minute))"
desired=$(oc -n ${LOGGING_NS} get ds fluentd  -o jsonpath={.status.desiredNumberScheduled})
log::info "Waiting for ${desired} fluent pods to be available...."
try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.status.numberReady}" "$desired" "$((2 * $minute))"

fpod=$(oc -n $LOGGING_NS get pod -l component=fluentd -o jsonpath={.items[0].metadata.name} --ignore-not-found)

log::info "Checking metrics using 'localhost'..."
try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://localhost:24231/metrics" "$((2 * $minute))"

log::info "Checking metrics using fluent service address..."
try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://fluentd.openshift-logging.svc:24231/metrics" "$((2 * $minute))"

fpod_ip="$(oc -n $LOGGING_NS get pod ${fpod} -o jsonpath='{.status.podIP}')"
log::info "Checking metrics using fluent podIP..."
try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://${fpod_ip}:24231/metrics" "$((2 * $minute))"

oc -n $LOGGING_NS exec $fpod -- curl -ks https://${fpod_ip}:24231/metrics >> $ARTIFACT_DIR/${fpod}-metrics-scrape 2>&1 || :
