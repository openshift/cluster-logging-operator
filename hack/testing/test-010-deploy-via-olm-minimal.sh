#!/bin/bash
# Given an OLM manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.

set -e

source "$(dirname "${BASH_SOURCE[0]}" )/../lib/init.sh"
source "$(dirname $0)/assertions"

os::test::junit::declare_suite_start "${BASH_SOURCE[0]}"

test_artifact_dir="${ARTIFACT_DIR:-"$repo_dir/_output"}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifact_dir ] ; then
  mkdir -p $test_artifact_dir
fi
export NAMESPACE="openshift-logging"
repo_dir="$(dirname $0)/../.."
manifest=${repo_dir}/manifests
version=$(basename $(find $manifest -type d | sort -r | head -n 1))

cleanup(){
  local return_code="$?"
  set +e
  gather_logging_resources ${NAMESPACE} $test_artifact_dir
  oc delete ns ${NAMESPACE} --wait=true --ignore-not-found
  oc delete crd elasticsearches.logging.openshift.io --wait=false --ignore-not-found
  oc delete crd kibanas.logging.openshift.io --wait=false --ignore-not-found
  os::cmd::try_until_failure "oc get project ${NAMESPACE}" "$((1 * $minute))"

  os::cleanup::all "${return_code}"
  
  set -e
  exit ${return_code}
}
trap cleanup exit

if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_CLUSTER_LOGGING_OPERATOR=$(sed -e "s,\${component},cluster-logging-operator," <(echo $IMAGE_FORMAT))
else
  IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-registry.svc.ci.openshift.org/origin/${OCP_VERSION}:cluster-logging-operator}
fi

os::log::info "Using image: $IMAGE_CLUSTER_LOGGING_OPERATOR"

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

oc create ns ${NAMESPACE} || :

tempdir=$(mktemp -d /tmp/elasticsearch-operator-XXXXXXXX)
get_operator_files $tempdir elasticsearch-operator ${EO_REPO:-openshift} ${EO_BRANCH:-master}
eo_manifest=${tempdir}/manifests
eo_version=$(basename $(find $eo_manifest -type d | sort -r | head -n 1))
os::log::info "Using elasticsearch operator version: ${eo_version}"

if ! is_crd_installed "elasticsearches.logging.openshift.io"; then
  os::cmd::expect_success "oc create -f ${eo_manifest}/${eo_version}/elasticsearches.crd.yaml"
fi

if ! is_crd_installed "kibanas.logging.openshift.io"; then
  os::cmd::expect_success "oc create -f ${eo_manifest}/${eo_version}/kibanas.crd.yaml"
fi
# Create static cluster roles and rolebindings
deploy_olm_catalog_unsupported_resources

os::log::info "Deploying operator from ${manifest}"
NAMESPACE=${NAMESPACE} \
VERSION=${version} \
OPERATOR_IMAGE=${IMAGE_CLUSTER_LOGGING_OPERATOR} \
MANIFEST_DIR=${manifest} \
TEST_NAMESPACE=${NAMESPACE} \
TARGET_NAMESPACE=${NAMESPACE} \
${repo_dir}/hack/vendor/olm-test-script/e2e-olm.sh

if [ "$?" != "0" ] ; then
	os::log::error "Error deploying operator via OLM using manifest: $manifest"
	exit 1
fi

TIMEOUT_MIN=$((2 * $minute))

##verify metrics rbac
# extra resources not support for ConfigMap based catelogs for now.
#os::cmd::expect_success "oc get clusterrole clusterlogging-collector-metrics"
#os::cmd::expect_success "oc get clusterrolebinding clusterlogging-collector-metrics"

# wait for operator to be ready
os::cmd::try_until_text "oc -n $NAMESPACE get deployment cluster-logging-operator -o jsonpath={.status.availableReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# deploy cluster logging
os::cmd::expect_success "oc -n $NAMESPACE create -f ${repo_dir}/hack/cr.yaml"

# assert deployment
assert_resources_exist
# assert kibana instance exists
assert_kibana_instance_exists
