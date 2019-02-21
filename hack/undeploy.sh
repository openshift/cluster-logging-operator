#!/bin/bash

source "$(dirname $0)/common"

for repo in ${repo_dir} ${ELASTICSEARCH_OP_REPO}; do
  oc delete -f ${repo}/manifests --ignore-not-found
done

oc delete -n openshift is origin-cluster-logging-operator || :
oc delete -n openshift bc cluster-logging-operator || :

NAMESPACE=openshift-logging make -C ${ELASTICSEARCH_OP_REPO} undeploy
