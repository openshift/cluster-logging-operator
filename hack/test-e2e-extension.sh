#!/bin/bash

current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/lib/init.sh"
source "${current_dir}/lib/util/logs.sh"
source "${current_dir}/testing-olm/utils"

get_setup_artifacts=true
CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}
ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
repo_dir="$(dirname $0)/.."
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi

cleanup(){
  local return_code="$?"
  
  set +e
  if [ "$return_code" != "0" -a $get_setup_artifacts ] ; then 
    oc get all -n openshift-operators-redhat > $test_artifactdir/openshift-operators-redhat_all.txt
    gather_logging_resources ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} $test_artifactdir
  fi

  if [ "${DO_CLEANUP:-true}" == "true" ] ; then
      ${repo_dir}/olm_deploy/scripts/operator-uninstall.sh
      ${repo_dir}/olm_deploy/scripts/catalog-uninstall.sh
      ${current_dir}/loki-operator-deploy/operator-uninstall.sh
      ${current_dir}/loki-operator-deploy/catalog-uninstall.sh
  fi
  os::cleanup::all "${return_code}"
  set -e
  exit ${return_code}
}
#trap cleanup exit

if [[ -z "${RELATED_CLO_CATALOG}" ]] ;then
  os::log::info "Deploying Operator Catalog from code"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
else
  os::log::info "use Catalog $RELATED_CLO_CATALOG"
  export CLUSTER_LOGGING_CATALOG_NAME="${RELATED_CLO_CATALOG}"
  export CLUSTER_LOGGING_CATALOG_NAMESPACE="openshift-marketplace"
fi

os::log::info "Deploying cluster-logging-operator"
${repo_dir}/olm_deploy/scripts/operator-install.sh

os::log::info "Deploying loki-operator"
# deploy loki-operator
${repo_dir}/hack/loki-operator-deploy/catalog-install.sh
${repo_dir}/hack/loki-operator-deploy/operator-install.sh

get_setup_artifacts=false

export JUNIT_REPORT_OUTPUT="/tmp/artifacts/junit/test-extension"
test_script="${current_dir}/testing-extension/test-by-labels.sh"
os::log::info "==============================================================="
os::log::info "running e2e extension test"
if "${test_script}" ; then
  os::log::info "==============================================================="
  os::log::info "e2e $test succeeded at $( date )"
  os::log::info "==========================================================="
else
  os::log::error "============= FAILED FAILED ============= "
  os::log::error "e2e $test failed at $( date )"
  os::log::error "============= FAILED FAILED ============= "
fi

get_logging_pod_logs

ARTIFACT_DIR="/tmp/artifacts/junit/" os::test::junit::generate_report ||:

if [[ -n "${failed:-}" ]]; then
    exit 1
fi
