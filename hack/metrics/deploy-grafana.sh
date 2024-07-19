 #!/bin/bash

 # This script deploys an http receiver that can be used for manual testing.
 # It splits incoming logs based upon the log_type and writes them to files.

 set -euo pipefail

#create NS
oc new-project grafana-operator||:

#Deploy Operator
cat << EOF | oc create -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  generateName: grafana-operator-
  namespace: grafana-operator
spec:
  targetNamespaces:
  - grafana-operator
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  generateName: grafana-operator-
  namespace: grafana-operator
spec:
  channel: v4
  name: grafana-operator
  installPlanApproval: Automatic
  source: community-operators
  sourceNamespace: openshift-marketplace
EOF
sleep 20s
oc -n grafana-operator wait --for=condition=available=true deployment/grafana-operator-controller-manager

# Deploy Grafana
helm repo add mobb https://rh-mobb.github.io/helm-charts/
helm upgrade --install -n grafana-operator \
  grafana mobb/grafana-cr --set "basicAuthPassword=myPassword"

oc adm policy add-cluster-role-to-user \
  cluster-monitoring-view -z grafana-serviceaccount ||:

BEARER_TOKEN=$(oc create token grafana-serviceaccount)

cat << EOF | oc apply -f -
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDataSource
metadata:
  name: prometheus-grafanadatasource
spec:
  datasources:
    - access: proxy
      editable: true
      isDefault: true
      jsonData:
        httpHeaderName1: 'Authorization'
        timeInterval: 5s
        tlsSkipVerify: true
      name: Prometheus
      secureJsonData:
        httpHeaderValue1: 'Bearer ${BEARER_TOKEN}'
      type: prometheus
      url: 'https://thanos-querier.openshift-monitoring.svc.cluster.local:9091'
  name: prometheus-grafanadatasource.yaml
EOF

echo "Route: $(oc get route grafana-route -o jsonpath='{"https://"}{.spec.host}{"\n"}')"