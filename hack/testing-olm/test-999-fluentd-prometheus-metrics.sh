#!/bin/bash

set -euo pipefail

repo_dir=${repo_dir:-$(dirname $0)/../..}
source "$(dirname "${BASH_SOURCE[0]}" )/../lib/init.sh"
source "$repo_dir/hack/testing-olm/utils"

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Prometheus metrics"

ARTIFACT_DIR=${ARTIFACT_DIR:-$repo_dir/_output}
LOGGING_NS=${LOGGING_NS:-openshift-logging}

test_name=test-999-fluentd-prometheus-metrics

cleanup() {
  local return_code="$?"

  os::test::junit::declare_suite_end

  set +e
  mkdir -p $ARTIFACT_DIR/$test_name
  gather_logging_resources ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} $ARTIFACT_DIR/$test_name

  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
      ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
  fi

  set -e
  exit $return_code
}
trap "cleanup" EXIT

if [ "${DO_SETUP:-true}" == "true" ] ; then
	os::log::info "Deploying cluster-logging-operator"
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
os::cmd::try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.metadata.name} --ignore-not-found" "fluentd" "$((3 * $minute))"
expectedcollectors=$( oc get nodes | grep -c " Ready " )
os::cmd::try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.status.desiredNumberScheduled}" "${expectedcollectors}"  "$((5 * $minute))"
desired=$(oc -n ${LOGGING_NS} get ds fluentd  -o jsonpath={.status.desiredNumberScheduled})

os::log::info "Waiting for ${desired} fluent pods to be available...."
os::cmd::try_until_text "oc -n ${LOGGING_NS} get ds fluentd -o jsonpath={.status.numberReady}" "$desired" "$((5 * $minute))"

fpod=$(oc -n $LOGGING_NS get pod -l component=fluentd -o jsonpath={.items[0].metadata.name} --ignore-not-found)

os::log::info "Checking metrics using fluent service address..."
os::cmd::try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://fluentd.openshift-logging.svc:24231/metrics" "$((2 * $minute))"

fpod_ip="$(oc -n $LOGGING_NS get pod ${fpod} -o jsonpath='{.status.podIP}')"
os::log::info "Checking metrics using fluent podIP..."
os::cmd::try_until_success "oc -n $LOGGING_NS exec $fpod -- curl -ks https://${fpod_ip}:24231/metrics" "$((2 * $minute))"

oc -n $LOGGING_NS exec $fpod -- curl -ks https://${fpod_ip}:24231/metrics >> $ARTIFACT_DIR/${fpod}-metrics-scrape 2>&1 || :
