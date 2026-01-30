#!/bin/sh
set -eou pipefail
current_dir=$(dirname "${BASH_SOURCE[0]}" )
source "${current_dir}/env.sh"

function deploy_catalogsource(){
  echo "Create catalogsource from index image ${LOKI_OPERATOR_CATALOG_IMAGE}"
  if [ "LOKI_OPERATOR_CATALOG" == "redhat-operators" ] ; then
    echo "Warning!!!! update perserved catalogsoruce redhat-operators is forbiden"
    exit 1
  fi
  oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name:  ${LOKI_OPERATOR_CATALOG}
  namespace: ${LOKI_OPERATOR_CATALOG_NAMESPACE}
spec:
  displayName: loki-operator catalog
  grpcPodConfig:
    nodeSelector:
      kubernetes.io/os: linux
      node-role.kubernetes.io/master: ""
    priorityClassName: system-cluster-critical
    securityContextConfig: restricted
    tolerations:
    - effect: NoSchedule
      key: node-role.kubernetes.io/master
      operator: Exists
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 120
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 120
  icon:
    base64data: ""
    mediatype: ""
  image: ${LOKI_OPERATOR_CATALOG_IMAGE}
  priority: -100
  publisher: Openshift QE
  sourceType: grpc
  updateStrategy:
    registryPoll:
      interval: 10m
EOF
}

echo "prepare catalogsource for loki-operator"
if [[ -n "$LOKI_OPERATOR_CATALOG_IMAGE" ]];then
  deploy_catalogsource
else
  echo "Using existing catalogsource ${LOKI_OPERATOR_CATALOG}"
fi
