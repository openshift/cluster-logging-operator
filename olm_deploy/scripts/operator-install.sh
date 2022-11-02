#!/bin/sh
set -eou pipefail
source $(dirname "${BASH_SOURCE[0]}")/env.sh

if oc get project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} for operator installation
else
  oc create namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
fi

set +e
oc label ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} openshift.io/cluster-monitoring=true --overwrite
oc annotate ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} openshift.io/node-selector="" --overwrite
# LOG-2620: containers violate PodSecurity
oc label ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} pod-security.kubernetes.io/enforce=privileged --overwrite
oc label ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} pod-security.kubernetes.io/audit=privileged --overwrite
oc label ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} pod-security.kubernetes.io/warn=privileged --overwrite
oc label ns/${CLUSTER_LOGGING_OPERATOR_NAMESPACE} security.openshift.io/scc.podSecurityLabelSync=false --overwrite

set -e

echo "##################"
echo "oc version"
oc version
echo "##################"

# create the operatorgroup
envsubst < olm_deploy/subscription/operator-group.yaml | oc apply -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -

# create the subscription
export OPERATOR_PACKAGE_CHANNEL=\"$(grep -Eo 'bundle\.channel\.default\.v1:(.*)$' bundle/metadata/annotations.yaml | cut -d ':' -f2 | tr -d '[:space:]')\"
echo "Deploying CLO from channel ${OPERATOR_PACKAGE_CHANNEL}"
subscription=$(envsubst < olm_deploy/subscription/subscription.yaml)
echo "Creating:"
echo "$subscription"
echo "$subscription" | oc apply -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -

olm_deploy/scripts/wait_for_deployment.sh ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} cluster-logging-operator
oc wait -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} --timeout=180s --for=condition=available deployment/cluster-logging-operator
