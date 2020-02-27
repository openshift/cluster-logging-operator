#!/bin/bash
# This script inits an elasticsearch-operator
# to deploy an Elasticsearch cluster.  It assumes it is capable of login as a
# user who has the cluster-admin role

if [ "${DEBUG:-}" = "true" ]; then
  set -x
fi
set -euo pipefail

source "$(dirname $0)/common"

if [[ -z `oc get project ${NAMESPACE} 2> /dev/null` ]] ; then
cat <<EOF | oc create -f -
apiVersion: v1
kind: Namespace
metadata:
  name: ${NAMESPACE}
  annotations:
    openshift.io/node-selector: ""
  labels:
    openshift.io/cluster-logging: "true"
    openshift.io/cluster-monitoring: "true"
EOF
fi

load_manifest() {
  local repo=$1
  local namespace=${2:-}
  if [ -n "${namespace}" ] ; then
    namespace="-n ${namespace}"
  fi

  pushd ${repo}/manifests
    for m in $(ls); do
      if [ "$(echo ${EXCLUSIONS[@]} | grep -o ${m} | wc -w)" == "0" ] ; then
        oc create -f ${m} ${namespace:-} ||:
      fi
    done
  popd
}

load_manifest ${repo_dir} ${NAMESPACE}

oc create -f hack/prometheus-operator-crd-cluster-roles.yaml ||:

oc create clusterrolebinding elasticsearch-operator-prometheus-rolebinding \
    --serviceaccount=${NAMESPACE}:elasticsearch-operator \
    --clusterrole=prometheus-crd-edit ||:

if [ "${REMOTE_CLUSTER:-false}" = false ] ; then
  sudo sysctl -w vm.max_map_count=262144
fi

if [ "${CREATE_ES_SECRET:-true}" = true ] ; then
  # This is necessary for running the operator with go run
  if [ ! -d /tmp/_working_dir ] ; then
    mkdir /tmp/_working_dir
  fi

  oc create secret generic elasticsearch -n ${NAMESPACE} \
      --from-file=admin-key=test/files/system.admin.key \
      --from-file=admin-cert=test/files/system.admin.crt \
      --from-file=admin-ca=test/files/ca.crt \
      --from-file=test/files/elasticsearch.crt \
      --from-file=test/files/logging-es.key \
      --from-file=test/files/logging-es.crt \
      --from-file=test/files/elasticsearch.key \
      ||:
fi
