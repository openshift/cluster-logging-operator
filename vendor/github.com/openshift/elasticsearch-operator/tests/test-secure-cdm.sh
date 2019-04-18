#!/bin/sh
# Integration test for the operator
# Input parameters:
# cmd: oc or kubectl
# insecure: True/False
# roles: [master, data, client]
# replicas: int
# nodeSelector: ???
# storage: EmptyDir, HostPath, volumeClaimTemplate, persistentVolumeClaim
# configMapName: string
# serviceAccountName: string

# Validation
# service created
# pod deployed
# configmap created
# serviceAccount created
# storage mouted
# secret created (if needed)

# Variant2:
# kubectl
# insecure: False
# roles: [master, data, client]
# replicas: 1
# nodeSelector: n/a
# storage: EmptyDir
# configMapName: logging-elasticsearch
# serviceAccountName: logging-elasticsearch

set -x
set -o errexit
set -o nounset

NAMESPACE="${NAMESPACE:-default}"
CONFIGMAP_NAME="${CONFIGMAP_NAME:-${CLUSTER_NAME}}"
SECRET_NAME="${SECRET_NAME:-${CLUSTER_NAME}}"
SERVICEACCOUNT_NAME="${SERVICEACCOUNT_NAME:-aggregated-logging-elasticsearch}"

. "./tests/utils.sh"

# Modify the CR to use secure image and enable secure Elasticsearch config
sed -i -e 's#image: docker.io/t0ffel/logging-insecure-elasticsearch5#image: registry.access.redhat.com/openshift3/logging-elasticsearch:v3.9#g' deploy/cr.yaml

# Create secret
kubectl create -n ${NAMESPACE} -f tests/test-secret.yaml

kubectl create -n ${NAMESPACE} -f deploy/cr.yaml
timeout 20m "./tests/wait-for-container.sh" elastic1-clientdatamaster
kubectl get deployment
kubectl get po

actual_configmap=$( get_configmap $CONFIGMAP_NAME )

if [ -z "$actual_configmap" ]; then
    echo "Desired configmap $CONFIGMAP_NAME is missing"
    exit 1
fi
echo "ConfigMap: OK"

actual_sa=$( get_serviceaccount $SERVICEACCOUNT_NAME )

if [ -z "$actual_sa" ]; then
    echo "Desired serviceaccount $SERVICEACCOUNT_NAME is missing"
    exit 1
fi
echo "ServiceAccount: OK"


kubectl delete -f deploy/cr.yaml
