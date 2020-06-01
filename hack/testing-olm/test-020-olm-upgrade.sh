#!/bin/bash
# Given an OLM manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.

# TODO: Renable this test when the OLM schema validation regression(https://bugzilla.redhat.com/show_bug.cgi?id=1825330)
# is resolved.
exit 0

set -euo pipefail

repo_dir="$( cd "$(dirname "$0")/../.." ; pwd -P )"
source "$repo_dir/hack/testing-olm/utils"
source "$repo_dir/hack/testing-olm/assertions"

start_seconds=$(date +%s)

ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output/"}
ARTIFACT_DIR="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
TIMEOUT_MIN=$((2 * $minute))
export CLUSTER_LOGGING_OPERATOR_NAMESPACE="openshift-logging"
manifest_dir=${repo_dir}/manifests
version=$(basename $(find $manifest_dir -type d | sort -r | head -n 1))
#use N-2 since OLM content is not matched to OCP version until release
previous_version=$(echo $version | awk '{print $1 - 0.2}')

cleanup(){
  local return_code="$?"
  set +e
  log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  oc -n openshift-operators-redhat -o yaml get subscription elasticsearch-operator > $ARTIFACT_DIR/subscription-eo.yml 2>&1 ||:
  oc -n openshift-operators-redhat -o yaml get deployment/elasticsearch-operator > $ARTIFACT_DIR/elasticsearch-operator-deployment.yml 2>&1 ||:
  oc -n openshift-operators-redhat -o yaml get pods > $ARTIFACT_DIR/openshift-operators-redhat-pods.yml 2>&1 ||:
  oc -n openshift-operators-redhat -o yaml get configmaps > $ARTIFACT_DIR/openshift-operators-redhat-configmaps.yml 2>&1 ||:
  oc -n openshift-operators-redhat describe deployment/elasticsearch-operator > $ARTIFACT_DIR/elasticsearch-operator.describe 2>&1 ||:

  oc logs -n ${NAMESPACE} deployment/cluster-logging-operator > $ARTIFACT_DIR/cluster-logging-operator.log 2>&1 ||:
  oc -n ${NAMESPACE} -o yaml get subscription cluster-logging-operator > $ARTIFACT_DIR/subscription-clo.yml 2>&1 ||:
  oc -n ${NAMESPACE} -o yaml get configmap cluster-logging > $ARTIFACT_DIR/configmap-for-catalogsource-clo.yml 2>&1 ||:
  oc -n ${NAMESPACE} -o yaml get catalogsource cluster-logging > $ARTIFACT_DIR/catalogsource-clo.yml 2>&1 ||:
  oc describe -n ${NAMESPACE} deployment/cluster-logging-operator > $ARTIFACT_DIR/cluster-logging-operator.describe.after_update  2>&1 ||:

  get_all_olm_logs $outdir $runtime

  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
      ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
  fi
  set -e
  exit ${return_code}
}
trap cleanup exit


# verify operator is ready
log::info "Verifing elasticsearch-operator is ready..."
try_until_text "oc -n openshift-operators-redhat get deployment elasticsearch-operator -o jsonpath={.status.availableReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# deploy CLO
log::info "Deploying cluster-logging-operator from marketplace..."
deploy_marketplace_operator "openshift-logging" "cluster-logging-operator" "$previous_version" "cluster-logging"

# verify operator is ready
try_until_text "oc -n openshift-logging get deployment cluster-logging-operator -o jsonpath={.status.updatedReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# deploy cluster logging
oc -n $NAMESPACE create -f ${repo_dir}/hack/cr.yaml

assert_resources_exist

oc describe -n ${NAMESPACE} deployment/cluster-logging-operator > $ARTIFACT_DIR/cluster-logging-operator.describe.before_update 2>&1

# deploy cluster logging catalog from local code
${repo_dir}/olm_deploy/scripts/catalog-deploy.sh

# patch subscription
payload="{\"op\":\"replace\",\"path\":\"/spec/source\",\"value\":\"cluster-logging-catalog\"}"
payload="$payload,{\"op\":\"replace\",\"path\":\"/spec/sourceNamespace\",\"value\":\"openshift-logging\"}"
payload="$payload,{\"op\":\"replace\",\"path\":\"/spec/channel\",\"value\":\"$version\"}"
oc -n $NAMESPACE patch subscription cluster-logging-operator --type json -p "[$payload]"

#verify deployment is rolled out
OPENSHIFT_VERSION=${OPENSHIFT_VERSION:-4.6}
IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-registry.svc.ci.openshift.org/ocp/${OPENSHIFT_VERSION}:cluster-logging-operator}
if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_CLUSTER_LOGGING_OPERATOR=$(echo $IMAGE_FORMAT | sed -e "s,\${component},cluster-logging-operator,")
fi
try_until_text "oc -n openshift-logging get deployment cluster-logging-operator -o jsonpath={.spec.template.spec.containers[0].image}" "${IMAGE_CLUSTER_LOGGING_OPERATOR}" ${TIMEOUT_MIN}

# verify operator is ready
try_until_text "oc -n openshift-logging get deployment cluster-logging-operator -o jsonpath={.status.updatedReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}


assert_resources_exist
