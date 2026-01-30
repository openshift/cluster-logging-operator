#!/bin/sh
set -eou pipefail
current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/env.sh"

if oc get project ${LOKI_OPERATOR_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${LOKI_OPERATOR_NAMESPACE} for operator installation
else
  oc create namespace ${LOKI_OPERATOR_NAMESPACE}
fi

oc label ns/${LOKI_OPERATOR_NAMESPACE} openshift.io/cluster-monitoring=true --overwrite
oc annotate ns/${LOKI_OPERATOR_NAMESPACE} openshift.io/node-selector="" --overwrite

echo "##################"
echo "oc version"
oc version
echo "##################"

# create the operatorgroup
operator_group=$(oc get pods -n "${LOKI_OPERATOR_NAMESPACE}" 2>/dev/null)
if [[ -z "${operator_group//[[:space:]]/}" ]]; then
  echo "create operator group loki-operator"
  oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: loki-operator
  namespace: ${LOKI_OPERATOR_NAMESPACE}
spec:
  targetNamespaces: []
EOF
else
  echo "using existing $operator_group"
fi

# create the subscription which is same channel with CLO here
echo "Deploying LO from channel ${OPERATOR_PACKAGE_CHANNEL} in ${LOKI_OPERATOR_NAMESPACE}"
echo "Creating:"
echo "loki-operator"
oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: loki-operator
  namespace: ${LOKI_OPERATOR_NAMESPACE}
spec:
  channel: ${OPERATOR_PACKAGE_CHANNEL}
  name: loki-operator
  source: ${LOKI_OPERATOR_CATALOG}
  sourceNamespace: ${LOKI_OPERATOR_CATALOG_NAMESPACE}
EOF
set -x 
${current_dir}/../../olm_deploy/scripts/wait_for_deployment.sh ${LOKI_OPERATOR_NAMESPACE} loki-operator-controller-manager
oc wait -n ${LOKI_OPERATOR_NAMESPACE} --timeout=180s --for=condition=available deployment/loki-operator-controller-manager

echo " done"
