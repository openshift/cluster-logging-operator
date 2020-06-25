#!/bin/sh
set -eou pipefail

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}


oc delete --wait --ignore-not-found ns ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}

oc delete --wait --ignore-not-found crd clusterloggings.logging.openshift.io
oc delete --wait --ignore-not-found crd collectors.logging.openshift.io
oc delete --wait --ignore-not-found crd clusterlogforwarders.logging.openshift.io

oc delete --wait --ignore-not-found clusterrolebinding clusterlogging-collector-metrics
oc delete --wait --ignore-not-found clusterrole clusterlogging-collector-metrics
