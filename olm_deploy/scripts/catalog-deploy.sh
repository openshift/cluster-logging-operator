#!/bin/sh
set -eou pipefail
source $(dirname "${BASH_SOURCE[0]}")/env.sh

echo "Deploying operator catalog with bundle using images: "
echo "cluster logging operator registry: ${IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY}"
echo "cluster logging operator: ${IMAGE_CLUSTER_LOGGING_OPERATOR}"
echo "fluentd: ${IMAGE_LOGGING_FLUENTD}"
echo "vector: ${IMAGE_LOGGING_VECTOR}"
echo "log-file-metric-exporter: ${IMAGE_LOG_FILE_METRIC_EXPORTER}"
echo "console-plugin: ${IMAGE_LOGGING_CONSOLE_PLUGIN}"

echo "In namespace: ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}"

if oc get project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} > /dev/null 2>&1 ; then
  echo using existing project ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} for operator catalog deployment
else
  oc create namespace ${CLUSTER_LOGGING_OPERATOR_NAMESPACE}
fi
# sleep helps to solve 'unauthorized: authentication required' for imagestream images
# if deployment is created immediately after the namespace is created, we get auth errors in cluster image registry
sleep 2

# substitute image names into the catalog deployment yaml and deploy it
envsubst < olm_deploy/operatorregistry/registry-deployment.yaml | oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -
oc wait -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} --timeout=120s --for=condition=available deployment/cluster-logging-operator-registry

# create the catalog service
oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f olm_deploy/operatorregistry/service.yaml

# find the catalog service ip, substitute it into the catalogsource and create the catalog source
export CLUSTER_IP=$(oc get -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} service cluster-logging-operator-registry -o jsonpath='{.spec.clusterIP}')
envsubst < olm_deploy/operatorregistry/catalog-source.yaml | oc create -n ${CLUSTER_LOGGING_OPERATOR_NAMESPACE} -f -
