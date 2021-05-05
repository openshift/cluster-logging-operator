#!/bin/sh
set -eou pipefail
LOGGING_VERSION=${LOGGING_VERSION:-5.2}
LOGGING_IS=${LOGGING_IS:-logging}
export IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY:-registry.ci.openshift.org/${LOGGING_IS}/${LOGGING_VERSION}:cluster-logging-operator-registry}
export IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-registry.ci.openshift.org/${LOGGING_IS}/${LOGGING_VERSION}:cluster-logging-operator}
export IMAGE_LOGGING_CURATOR5=${IMAGE_LOGGING_CURATOR5:-registry.ci.openshift.org/${LOGGING_IS}/${LOGGING_VERSION}:logging-curator5}
export IMAGE_LOGGING_FLUENTD=${IMAGE_LOGGING_FLUENTD:-registry.ci.openshift.org/${LOGGING_IS}/${LOGGING_VERSION}:logging-fluentd}

CLUSTER_LOGGING_OPERATOR_NAMESPACE=${CLUSTER_LOGGING_OPERATOR_NAMESPACE:-openshift-logging}

echo "Deploying operator catalog with bundle using images: "
echo "cluster logging operator registry: ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
echo "cluster logging operator: ${IMAGE_CLUSTER_LOGGING_OPERATOR}"
echo "curator5: ${IMAGE_LOGGING_CURATOR5}"
echo "fluentd: ${IMAGE_LOGGING_FLUENTD}"

echo "In namespace: ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}"

if oc get project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} for operator catalog deployment
else
  oc create namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
fi

# substitute image names into the catalog deployment yaml and deploy it
envsubst < olm_deploy/operatorregistry/registry-deployment.yaml | oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -
oc wait -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} --timeout=120s --for=condition=available deployment/cluster-logging-operator-registry

# create the catalog service
oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f olm_deploy/operatorregistry/service.yaml

# find the catalog service ip, substitute it into the catalogsource and create the catalog source
export CLUSTER_IP=$(oc get -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} service cluster-logging-operator-registry -o jsonpath='{.spec.clusterIP}')
envsubst < olm_deploy/operatorregistry/catalog-source.yaml | oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -
