#!/bin/sh
set -eou pipefail

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}

oc delete --wait --ignore-not-found project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
