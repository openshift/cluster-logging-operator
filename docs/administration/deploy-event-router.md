# Deploying Event Router
The EventRouter is a simple service that listens to kubernetes events and writes them to `STDOUT`.  It is not managed by the `cluster-logging-operator`.
Creating a deployment using the following template will enable event collection equivalent to the same capability from 3.11 releases or earlier.
Currently, all events are collected by the cluster logging collector and indexed to the `infra` index when
the `EventRouter` is deployed into an operations namespace (e.g. `openshift-*`). Elevated permissions are required to create the cluster level roles and bindings referenced in the following template:

```
kind: Template
apiVersion: v1
metadata:
  name: eventrouter-template
  annotations:
    description: "A pod forwarding kubernetes events to cluster logging stack."
    tags: "events,logging, cluster-logging"
objects:
  - kind: ServiceAccount
    apiVersion: v1
    metadata:
      name: eventrouter
      namespace: openshift-logging
  - kind: ClusterRole
    apiVersion: v1
    metadata:
      name: event-reader
    rules:
    - apiGroups: [""]
      resources: ["events"]
      verbs: ["get", "watch", "list"]
  - kind: ClusterRoleBinding
    apiVersion: v1
    metadata:
      name: event-reader-binding
    subjects:
    - kind: ServiceAccount
      name: eventrouter
      namespace: openshift-logging
    roleRef:
      kind: ClusterRole
      name: event-reader
  - kind: ConfigMap
    apiVersion: v1
    metadata:
      name: eventrouter
      namespace: openshift-logging
    data:
      config.json: |-
        {
          "sink": "stdout"
        }
  - kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: eventrouter
      namespace: eventrouter
      labels:
        component: eventrouter
        logging-infra: eventrouter
        provider: openshift
    spec:
      selector:
        matchLabels:
          component: eventrouter
          logging-infra: eventrouter
          provider: openshift
      replicas: 1
      template:
        metadata:
          labels:
            component: eventrouter
            logging-infra: eventrouter
            provider: openshift
          name: eventrouter
        spec:
          serviceAccount: eventrouter
          containers:
            - name: kube-eventrouter
              image: ${IMAGE}
              imagePullPolicy: IfNotPresent
              resources:
                requests:
                  cpu: ${CPU}
                  memory: ${MEMORY}
              volumeMounts:
              - name: config-volume
                mountPath: /etc/eventrouter
          volumes:
            - name: config-volume
              configMap:
                name: eventrouter
parameters:
  - name: IMAGE
    displayName: Image
    value: "quay.io/openshift-logging/eventrouter:0.3"
  - name: MEMORY
    displayName: Memory
    value: "20Mi"
  - name: CPU
    displayName: CPU
    value: "100m"

```
Process and apply this above template like:

```
oc process -f <templatefile> | oc apply -f -
```
