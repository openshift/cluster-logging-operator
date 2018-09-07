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

# Variant1(done):
# kubectl
# insecure: True
# roles: [master, data, client]
# replicas: 1
# nodeSelector: n/a
# storage: EmptyDir
# configMapName: n/a
# serviceAccountName: n/a

set -x
set -o errexit
set -o nounset

NAMESPACE="${NAMESPACE:-default}"
CONFIGMAP_NAME="${CONFIGMAP_NAME:-${CLUSTER_NAME}}"
SECRET_NAME="${SECRET_NAME:-${CLUSTER_NAME}}"
SERVICEACCOUNT_NAME="${SERVICEACCOUNT_NAME:-aggregated-logging-elasticsearch}"

. "./tests/utils.sh"

kubectl create -f deploy/cr.yaml
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


echo "*** Test changing deployment by changing a label"
# Save name of the current pod, to know what to delete later
old_pod=$(kubectl get po | awk '/elastic1-.*/{print $1}')

# Add label and see that deployment is respawned
kubectl patch elasticsearch/elastic1 --type=merge --patch '{"metadata": {"labels": {"testlabel": "addedvalue" }}}'

# Old pod must be disposed
wait_pod_completion $old_pod

# new pod must be created
timeout 20m "./tests/wait-for-container.sh" elastic1-clientdatamaster

pod=$(kubectl get po -n $NAMESPACE -l testlabel=addedvalue -o name)

if [ -z "$pod" ]; then
  echo "No pod found via label.."
  exit 1
fi
echo "Pod successfully found via label"

kubectl delete -f deploy/cr.yaml