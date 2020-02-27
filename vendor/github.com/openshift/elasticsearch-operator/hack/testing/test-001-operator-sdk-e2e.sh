#!/bin/bash
set -euo pipefail

if [ -n "${DEBUG:-}" ]; then
    set -x
fi

IMAGE_ELASTICSEARCH_OPERATOR=${IMAGE_ELASTICSEARCH_OPERATOR:-quay.io/openshift/origin-elasticsearch-operator:latest}

if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_ELASTICSEARCH_OPERATOR=$(sed -e "s,\${component},elasticsearch-operator," <(echo $IMAGE_FORMAT))
fi
LOGGING_IMAGE_STREAM=${LOGGING_IMAGE_STREAM:-stable}
ELASTICSEARCH_IMAGE=${ELASTICSEARCH_IMAGE:-registry.svc.ci.openshift.org/ocp/$LOGGING_IMAGE_STREAM:logging-elasticsearch6}

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

repo_dir="$(dirname $0)/../.."
source "${repo_dir}/hack/lib/log/output.sh"
source "${repo_dir}/hack/testing/utils"
ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output/$(basename ${BASH_SOURCE[0]})"}
test_artifact_dir=$ARTIFACT_DIR/test-001-operator-sdk
if [ ! -d $test_artifact_dir ] ; then
  mkdir -p $test_artifact_dir
fi

manifest=$(mktemp)
files="01-service-account.yaml 02-role.yaml 03-role-bindings.yaml 05-deployment.yaml"
pushd manifests;
  for f in ${files}; do
     cat ${f} >> ${manifest};
  done;
popd
# update the manifest with the image built by ci
sed -i "s,quay.io/openshift/origin-elasticsearch-operator:latest,${IMAGE_ELASTICSEARCH_OPERATOR}," ${manifest}
sed -i "s,quay.io/openshift/origin-logging-elasticsearch6:latest,${ELASTICSEARCH_IMAGE}," ${manifest}

if [ "${REMOTE_CLUSTER:-false}" = false ] ; then
  sudo sysctl -w vm.max_map_count=262144 ||:
fi

TEST_NAMESPACE="${TEST_NAMESPACE:-e2e-test-${RANDOM}}"

start_seconds=$(date +%s)
cleanup(){
  local return_code="$?"
  set +e
  os::log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  
  if [ "${SKIP_CLEANUP:-false}" == "false" ] ; then
    get_all_logging_pod_logs ${TEST_NAMESPACE} $test_artifact_dir
    for item in "ns/${TEST_NAMESPACE}" "clusterrole/elasticsearch-operator" "clusterrolebinding/elasticsearch-operator-rolebinding"; do
      oc delete $item --wait=true --ignore-not-found --force --grace-period=0
    done
  fi
  
  exit ${return_code}
}
trap cleanup exit

if oc get project ${TEST_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${TEST_NAMESPACE}
else
  oc create namespace ${TEST_NAMESPACE}
fi

sed -i "s/namespace: openshift-logging/namespace: ${TEST_NAMESPACE}/g" ${manifest}

oc create -n ${TEST_NAMESPACE} -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml ||:
oc create -n ${TEST_NAMESPACE} -f \
https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml ||:

TEST_NAMESPACE=${TEST_NAMESPACE} go test ./test/e2e/... \
  -root=$(pwd) \
  -kubeconfig=${KUBECONFIG} \
  -globalMan manifests/04-crd.yaml \
  -namespacedMan ${manifest} \
  -v \
  -parallel=1 \
  -singleNamespace \
  -timeout 1200s
