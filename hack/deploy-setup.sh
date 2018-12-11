#!/bin/bash
# This script inits a cluster to allow cluster-logging-operator
# to deploy logging.  It assumes it is capable of login as a
# user who has the cluster-admin role

set -euxo pipefail

source "$(dirname $0)/common"

load_manifest() {
  local repo=$1
  local namespace=${2:-}
  if [ -n "${namespace}" ] ; then
    namespace="-n ${namespace}"
  fi

  pushd ${repo}/manifests
    if ! oc get project openshift-logging > /dev/null 2>&1 && test -f 01-namespace.yaml ; then
      oc create -f 01-namespace.yaml
    fi
    for m in $(ls); do
      if [ "$(echo ${EXCLUSIONS[@]} | grep -o ${m} | wc -w)" == "0" ] ; then
        oc create -f ${m} ${namespace:-}
      fi
    done
  popd
}

# This is required for when running the operator locally using go run
rm -rf /tmp/_working_dir
mkdir /tmp/_working_dir

load_manifest ${repo_dir}
CREATE_ES_SECRET=false NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} deploy-setup

oc adm policy add-scc-to-user privileged -z rsyslog -n openshift-logging
oc adm policy add-cluster-role-to-user cluster-reader -z rsyslog -n openshift-logging
oc adm policy add-scc-to-user privileged -z fluentd -n openshift-logging
oc adm policy add-cluster-role-to-user cluster-reader -z fluentd -n openshift-logging
if [ "${USE_RSYSLOG:-false}" = true ] ; then
  oc label node --all logging-infra-rsyslog=true
else
  oc label node --all logging-infra-fluentd=true
fi
