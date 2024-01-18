#!/bin/bash

# This script deploys an http receiver that can be used for manual testing.
# It splits incoming logs based upon the log_type and writes them to files.

set -euo pipefail

NAMESPACE=$1

cat <<EOF | oc create -n $NAMESPACE -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: http-receiver
data:
  vector.toml: |
    [sources.my_source]
    type = "http_server"
    address = "0.0.0.0:8090"
    decoding.codec = "json"
    framing.method = "newline_delimited"
    [transforms.route_logs]
    type = "route"
    inputs = ["my_source"]
    route.audit = '.log_type == "audit"'
    route.container = 'exists(.kubernetes)'
    route.journal = '!exists(.kubernetes)'
    [sinks.container]
    inputs = ["route_logs.container"]
    type = "file"
    path = "/tmp/container/{{kubernetes.namespace_name}}_{{kubernetes.pod_name}}_{{kubernetes.container_name}}.json"

    [sinks.container.encoding]
    codec = "json"
    [sinks.out_journal]
    inputs = ["route_logs.journal"]
    type = "file"
    path = "/tmp/journal/journal.json"

    [sinks.out_journal.encoding]
    codec = "json"
    [sinks.out_audit]
    inputs = ["route_logs.audit"]
    type = "file"
    path = "/tmp/audit/audit.json"

    [sinks.out_audit.encoding]
    codec = "json"
EOF

cat <<EOF | oc create -n $NAMESPACE -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: http-receiver
  labels:
    app: http-receiver
    component: http-receiver
spec:
  replicas: 1
  selector:
    matchLabels:
      app: http-receiver
  template:
    metadata:
      labels:
        app: http-receiver
        component: http-receiver
    spec:
      volumes:
      - configMap:
          defaultMode: 420
          name: http-receiver
        name: config
      containers:
      - name: receiver
        image: quay.io/openshift-logging/vector:5.8.0
        command:
        - /usr/bin/vector
        - --config-toml=/etc/vector/vector.toml
        ports:
        - containerPort: 8090
        volumeMounts:
        - mountPath: /etc/vector
          name: config
EOF

cat <<EOF | oc create -n $NAMESPACE -f -
apiVersion: v1
kind: Service
metadata:
  labels:
    app: http-receiver
    component: http-receiver
  name: http-receiver
spec:
  ports:
  - port: 8090
    protocol: TCP
  selector:
    app: http-receiver
    component: http-receiver
EOF

