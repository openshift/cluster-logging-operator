#!/bin/bash

DOMAIN=$(oc get ingresses.config/cluster -o jsonpath={.spec.domain})
echo ${DOMAIN}

oc apply -f https://github.com/splunk/splunk-operator/releases/download/2.0.0/splunk-operator-cluster.yaml
oc adm policy add-scc-to-user nonroot -z splunk-operator-controller-manager -n splunk-operator
oc adm policy add-scc-to-user nonroot -z default -n splunk-operator
oc wait -n splunk-operator --timeout=180s --for=condition=available deployment/splunk-operator-controller-manager

cat <<EOF | oc apply -n splunk-operator -f -
apiVersion: enterprise.splunk.com/v3
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

