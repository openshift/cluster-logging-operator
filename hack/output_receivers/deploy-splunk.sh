#!/bin/bash

DOMAIN=$(oc get ingresses.config/cluster -o jsonpath={.spec.domain})
echo ${DOMAIN}

oc apply --server-side --force-conflicts -f https://github.com/splunk/splunk-operator/releases/download/2.8.1/splunk-operator-cluster.yaml
oc adm policy add-scc-to-user privileged -z splunk-operator-controller-manager -n splunk-operator
oc adm policy add-scc-to-user privileged -z default -n splunk-operator
oc set image -n splunk-operator deployment/splunk-operator-controller-manager kube-rbac-proxy=registry.k8s.io/kubebuilder/kube-rbac-proxy:v0.13.1
oc wait -n splunk-operator --timeout=180s --for=condition=available deployment/splunk-operator-controller-manager

cat <<EOF | oc apply -n splunk-operator -f -
apiVersion: enterprise.splunk.com/v4
kind: Standalone
metadata:
  name: s1
  finalizers:
  - enterprise.splunk.com/delete-pvc
EOF

cat <<EOF | oc create -n splunk-operator -f -
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: splunk-web
spec:
  host: splunk-web-splunk-operator.${DOMAIN}
  port:
    targetPort: http-splunkweb
  to:
    kind: Service
    name: splunk-s1-standalone-service
EOF

cat <<EOF | oc create -n splunk-operator -f -
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: splunk-hec
spec:
  host: splunk-hec-splunk-operator.${DOMAIN}
  port:
    targetPort: http-hec
  to:
    kind: Service
    name: splunk-s1-standalone-service
EOF

