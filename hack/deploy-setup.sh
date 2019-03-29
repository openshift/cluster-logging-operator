#!/bin/bash
# This script inits a cluster to allow cluster-logging-operator
# to deploy logging.  It assumes it is capable of login as a
# user who has the cluster-admin role

set -euxo pipefail

source "$(dirname $0)/common"

# This is required for when running the operator locally using go run
rm -rf /tmp/_working_dir
mkdir /tmp/_working_dir

manifest=$(mktemp)
if ! oc get project openshift-logging > /dev/null 2>&1 ; then
  $repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'ns' >> $manifest
fi
$repo_dir/hack/gen-olm-artifacts.sh ${CSV_FILE} ${NAMESPACE} 'sa,role,clusterrole,crd'  >> $manifest
oc create -f $manifest

CREATE_ES_SECRET=false NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} deploy-setup
