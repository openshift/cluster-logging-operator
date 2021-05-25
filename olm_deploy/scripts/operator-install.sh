#!/bin/sh
set -eou pipefail

export CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}


if oc get project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} for operator installation
else
  oc create namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
fi

set +e
oc label ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} openshift.io/cluster-monitoring=true --overwrite
oc annotate ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} openshift.io/node-selector="" --overwrite
set -e

echo "##################"
echo "oc version"
oc version
echo "##################"

# create the operatorgroup
envsubst < olm_deploy/subscription/operator-group.yaml | oc apply -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -

# create the subscription
export OPERATOR_PACKAGE_CHANNEL=\"$(grep name manifests/cluster-logging.package.yaml | grep  -oh "[0-9]\+\.[0-9]\+")\"
echo "Deploying CLO from channel ${OPERATOR_PACKAGE_CHANNEL}"
subscription=$(envsubst < olm_deploy/subscription/subscription.yaml)
echo "Creating:"
echo "$subscription"
echo "$subscription" | oc apply -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -

olm_deploy/scripts/wait_for_deployment.sh ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} cluster-logging-operator
oc wait -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} --timeout=180s --for=condition=available deployment/cluster-logging-operator
