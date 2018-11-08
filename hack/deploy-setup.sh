#!/bin/bash
# This script inits a cluster to allow cluster-logging-operator
# to deploy logging.  It assumes it is capable of login as a
# user who has the cluster-admin role

set -e

source "$(dirname $0)/common"

load_manifest() {
  local repo=$1
  local namespace=${2:-}
  if [ -n "${namespace}" ] ; then
    namespace="-n ${namespace}"
  fi

  pushd ${repo}/manifests
    for m in $(ls); do
      if [ "$(echo ${EXCLUSIONS[@]} | grep -o ${m} | wc -w)" == "0" ] ; then
        oc create -f ${m} ${namespace:-}
      fi
    done
  popd
}

# This is required for when running the operator locally using go run
mkdir /tmp/_working_dir

load_manifest ${repo_dir}
load_manifest ${ELASTICSEARCH_OP_REPO} openshift-logging

# This is necessary for running the operator with go run
mkdir /tmp/_working_dir

tmpdir=$(mktemp -d)
pushd ${tmpdir}
  oc adm ca create-signer-cert --cert='ca.crt' --key='ca.key' --serial='ca.serial.txt'
  oc adm ca create-server-cert --cert='kibana-internal.crt' --key='kibana-internal.key' \
      --hostnames='kibana,kibana-infra' --signer-cert='ca.crt' --signer-key='ca.key' --signer-serial='ca.serial.txt'
  oc create -n openshift-logging secret generic logging-master-ca --from-file=masterca=ca.crt --from-file=masterkey=ca.key --from-file=kibanacert=kibana-internal.crt \
      --from-file=kibanakey=kibana-internal.key
popd

oc label node --all logging-infra-fluentd=true
oc adm policy add-scc-to-user privileged -z fluentd -n openshift-logging
