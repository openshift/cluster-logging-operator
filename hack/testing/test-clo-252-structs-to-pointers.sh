#!/bin/bash
# CLO issue #252. Verify Logging still functions 
# after API converts component specs to pointers 
# from empty structs

set -e
if [ "${DEBUG:-}" = "true" ]; then
	set -x
fi

repo_dir="$( cd "$(dirname "$0")/../.." ; pwd -P )"
source "$repo_dir/hack/lib/init.sh"
source "$repo_dir/hack/testing/assertions"
source "$repo_dir/hack/testing/utils"

start_seconds=$(date +%s)
os::test::junit::declare_suite_start "${BASH_SOURCE[0]}"

ARTIFACT_DIR=${ARTIFACT_DIR:-"$repo_dir/_output"}
if [ ! -d $ARTIFACT_DIR ] ; then
  mkdir -p $ARTIFACT_DIR
fi

KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
TIMEOUT_MIN=$((2 * $minute))
NAMESPACE="openshift-logging"
manifest_dir=${repo_dir}/manifests
version=$(basename $(find $manifest_dir -type d | sort -r | head -n 1))

cleanup(){
  local return_code="$?"
  set +e
  os::log::info "Running cleanup"
  end_seconds=$(date +%s)
  runtime="$(($end_seconds - $start_seconds))s"
  oc -n openshift-operators-redhat -o yaml get subscription elasticsearch-operator > $ARTIFACT_DIR/subscription-eo.yml 2>&1 ||:
  oc logs -n ${NAMESPACE} deployment/cluster-logging-operator > $ARTIFACT_DIR/cluster-logging-operator.log 2>&1 ||:
  oc -n ${NAMESPACE} -o yaml get subscription cluster-logging-operator > $ARTIFACT_DIR/subscription-clo.yml 2>&1 ||:
  oc -n ${NAMESPACE} -o yaml get configmap cluster-logging > $ARTIFACT_DIR/configmap-for-catalogsource-clo.yml 2>&1 ||:
  oc -n ${NAMESPACE} -o yaml get catalogsource cluster-logging > $ARTIFACT_DIR/catalogsource-clo.yml 2>&1 ||:
  oc  -n openshift-operator-lifecycle-manager logs --since=$runtime deployment/catalog-operator > $ARTIFACT_DIR/catalog-operator.logs 2>&1 ||:
  oc  -n openshift-operator-lifecycle-manager logs --since=$runtime deployment/olm-operator > $ARTIFACT_DIR/olm-operator.logs 2>&1 ||:
  
  for item in "crd/elasticsearches.logging.openshift.io" "crd/clusterloggings.logging.openshift.io" "ns/openshift-logging" "ns/openshift-operators-redhat"; do
    oc delete $item --wait=true --ignore-not-found --force --grace-period=0
  done
  for item in "ns/openshift-logging" "ns/openshift-operators-redhat"; do
    os::cmd::try_until_failure "oc get project ${item}" "$((1 * $minute))"
  done

  os::cleanup::all "${return_code}"
  
  exit ${return_code}
}
trap cleanup exit

if [ -n "${IMAGE_CLUSTER_LOGGING_OPERATOR:-}" ] ; then
  source "$repo_dir/hack/common"
fi
if [ -n "${IMAGE_FORMAT:-}" ] ; then
  IMAGE_CLUSTER_LOGGING_OPERATOR=$(sed -e "s,\${component},cluster-logging-operator," <(echo $IMAGE_FORMAT))
else
  IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-registry.svc.ci.openshift.org/origin/${version}:cluster-logging-operator}
fi


# deploy EO
os::cmd::expect_success 'deploy_marketplace_operator "openshift-operators-redhat" "elasticsearch-operator" "elasticsearch-operator" "true" '

# verify operator is ready
os::cmd::try_until_text "oc -n openshift-operators-redhat get deployment elasticsearch-operator -o jsonpath={.status.availableReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# deploy CLO
os::cmd::expect_success 'deploy_marketplace_operator "openshift-logging" "cluster-logging-operator" "cluster-logging"'

# verify operator is ready
os::cmd::try_until_text "oc -n openshift-logging get deployment cluster-logging-operator -o jsonpath={.status.updatedReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

# deploy cluster logging
cl_yaml=$(cat <<EOM
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
      storage: {}
      resources:
        requests:
          memory: 3Gi
      redundancyPolicy: "ZeroRedundancy"
EOM
)

os::cmd::expect_success "echo \"$cl_yaml\" | oc -n $NAMESPACE create -f -"

os::cmd::try_until_success "oc -n $NAMESPACE get elasticsearch elasticsearch" ${TIMEOUT_MIN}
os::cmd::expect_failure "oc -n $NAMESPACE get deployment kibana"
os::cmd::expect_failure "oc -n $NAMESPACE get cronjob curator"
os::cmd::expect_failure "oc -n $NAMESPACE get ds fluentd"

os::cmd::expect_success "deploy_config_map_catalog_source $NAMESPACE ${repo_dir}/manifests '${IMAGE_CLUSTER_LOGGING_OPERATOR}'"

# patch subscription
payload="{\"op\":\"replace\",\"path\":\"/spec/source\",\"value\":\"cluster-logging\"}"
payload="$payload,{\"op\":\"replace\",\"path\":\"/spec/sourceNamespace\",\"value\":\"openshift-logging\"}"
payload="$payload,{\"op\":\"replace\",\"path\":\"/spec/channel\",\"value\":\"$version\"}"
os::cmd::expect_success "oc -n $NAMESPACE patch subscription cluster-logging-operator --type json -p '[$payload]'"

#verify deployment is rolled out
os::cmd::try_until_text "oc -n openshift-logging get deployment cluster-logging-operator -o jsonpath={.spec.template.spec.containers[0].image}" "${IMAGE_CLUSTER_LOGGING_OPERATOR}" ${TIMEOUT_MIN}

# verify operator is ready
os::cmd::try_until_text "oc -n openshift-logging get deployment cluster-logging-operator -o jsonpath={.status.updatedReplicas} --ignore-not-found" "1" ${TIMEOUT_MIN}

os::cmd::try_until_success "oc -n $NAMESPACE get elasticsearch elasticsearch" ${TIMEOUT_MIN}
os::cmd::expect_failure "oc -n $NAMESPACE get deployment kibana"
os::cmd::expect_failure "oc -n $NAMESPACE get cronjob curator"
os::cmd::expect_failure "oc -n $NAMESPACE get ds fluentd"
