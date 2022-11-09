#!/bin/bash
# Given an OLM manifest, verify a green field deployment
# of cluster logging by asserting CLO creates the resources
# that begets the operands that make up logging.

set -eou pipefail

source "$(dirname "${BASH_SOURCE[0]}" )/../lib/init.sh"
source "$(dirname $0)/assertions"

LOGGING_NS=${LOGGING_NS:-"openshift-logging"}

mkdir -p /tmp/artifacts/junit
os::test::junit::declare_suite_start "[ClusterLogging] Deploy via OLM minimal"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$(pwd)/_output"}
test_artifactdir="${ARTIFACT_DIR}/$(basename ${BASH_SOURCE[0]})"
if [ ! -d $test_artifactdir ] ; then
  mkdir -p $test_artifactdir
fi

export CLUSTER_LOGGING_OPERATOR_NAMESPACE=$LOGGING_NS
repo_dir="$(dirname $0)/../.."

cleanup(){
  local return_code="$?"

  set +e
  if [ "$return_code" != "0" ] ; then 
    gather_logging_resources ${LOGGING_NS} $test_artifactdir
  fi

  if [ "${DO_TEST_CLEANUP:-true}" == "true" ] ; then
    for r in "clusterlogging/instance" "clusterlogforwarder/instance"; do
      oc -n $LOGGING_NS delete $r --ignore-not-found --force --grace-period=0||:
      os::cmd::try_until_failure "oc -n $LOGGING_NS get $r" "$((1 * $minute))"
    done
  fi

  os::test::junit::declare_suite_end

  set -e
  exit ${return_code}
}
trap cleanup exit

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

if [ "${DO_SETUP:-false}" == "true" ] ; then
  if [ "${DO_EO_SETUP:-true}" == "true" ] ; then
      pushd ../../elasticsearch-operator
      # install the catalog containing the elasticsearch operator csv
      ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/catalog-deploy.sh
      # install the elasticsearch operator from that catalog
      ELASTICSEARCH_OPERATOR_NAMESPACE=openshift-operators-redhat olm_deploy/scripts/operator-install.sh
      popd
  fi

  os::log::info "Deploying cluster-logging-operator"
  ${repo_dir}/olm_deploy/scripts/catalog-deploy.sh
  ${repo_dir}/olm_deploy/scripts/operator-install.sh
fi

TIMEOUT_MIN=$((2 * $minute))

##verify metrics rbac
# extra resources not support for ConfigMap based catelogs for now.
#os::cmd::expect_success "oc get clusterrole clusterlogging-collector-metrics"
#os::cmd::expect_success "oc get clusterrolebinding clusterlogging-collector-metrics"

# wait for operator to be ready
os::cmd::try_until_text "oc -n $LOGGING_NS get deployment cluster-logging-operator -o jsonpath={.status.availableReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# test the validation of an invalid cr
os::cmd::expect_failure_and_text "oc -n $LOGGING_NS create -f ${repo_dir}/hack/cr_invalid.yaml" "invalid: metadata.name: Unsupported value"

# deploy cluster logging with unmanaged state
os::cmd::expect_success "oc -n $LOGGING_NS create -f ${repo_dir}/hack/cr-unmanaged.yaml"

# wait few seconds
sleep 10
# assert does not exist
assert_resources_does_not_exist
# assert kibana instance does not exists
assert_kibana_instance_does_not_exists

# wait few seconds
sleep 10
# delete cluster logging
os::cmd::expect_success "oc -n $LOGGING_NS delete -f ${repo_dir}/hack/cr-unmanaged.yaml"

# deploy cluster logging
os::cmd::expect_success "oc -n $LOGGING_NS create -f ${repo_dir}/hack/cr.yaml"

# assert deployment
assert_resources_exist
# assert kibana instance exists
assert_kibana_instance_exists

# modify the collector
os::cmd::expect_success "oc  -n $LOGGING_NS patch clusterlogging instance --type='json' -p '[{\"op\":\"replace\",\"path\":\"/spec/collection/type\",\"value\":\"vector\"}]'"

# wait few seconds since CLO reconciles every 30
sleep 40

# assert deployment
assert_resources_exist