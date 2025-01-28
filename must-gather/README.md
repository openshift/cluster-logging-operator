cluster-logging must-gather
=================

`cluster-logging-must-gather` is a tool built on top of [OpenShift must-gather](https://github.com/openshift/must-gather)
that expands its capabilities to gather Openshift Cluster Logging information.

**Note:** This image is only built for x86_64 architecture

### Usage
To gather only Openshift Cluster Logging information: 
```sh
oc adm must-gather --image=quay.io/openshift-logging/cluster-logging-operator:latest -- /usr/bin/gather
```

To gather default [OpenShift must-gather](https://github.com/openshift/must-gather) in addition to Openshift Cluster Logging information: 
```sh
oc adm must-gather --image-stream=openshift/must-gather --image=quay.io/openshift/origin-cluster-logging-operator -- /usr/bin/gather
```

The command above will create a local directory with a dump of the cluster-logging state.
Note that this command will only get data related to the cluster-logging part of the OpenShift cluster.

You will get a dump of:
- The openshift-logging namespace and its children objects
- The openshift-operators and openshift-operators-redhat namespaces and its children objects
- Other namespaces with ClusterLogForwarder object
- The UIplugin and console namespace
- The cluster-logging install objects
- All cluster-logging CRD's definitions
- All nodes objects
- All persistent volumes objects
- Custom logs, configurations and health status per component, i.e. collection, logStore

In order to get data about other parts of the cluster (not specific to cluster-logging) you should
run `oc adm must-gather` (without passing a custom image). Run `oc adm must-gather -h` to see more options.

### Resources in the generated must-gather directory
CLO information is located in `cluster-logging/clo/`, and some resources information (like collectors) are in individual namespaces under `cluster-logging/[namespace_name]`.

The `clusterlogging` and `clusterlogforwarder` resources, and also `installplans`, `subscriptions`, `clusterserviceversions`, `logfilemetricexporter`, `uiplugion`, etc. are now collected by `oc adm inspect` command in the different `namespaces/[namespace_name]/` directories and no longer in `cluster-logging/clo`. This directory structure allows tools like [`omc`](https://github.com/gmeghnag/omc/) to work with those resources in a similar way to `oc` commands on a cluster.

The `deployments`, `daemonsets` and `secrets` are also found under `namespaces/[namespace_name]/` and can also be seen using the [`omc`](https://github.com/gmeghnag/omc/) tool.

Example must-gather for cluster-logging output (use `tree` for up-to-date structure):
```
├── cluster-logging
│  ├── clo
│  │  ├── cluster-logging-operator-xxxxxxxxxx-xxxxx
│  │  └── version
│  └── namespaces
│  │  └── [nampespace_name]               ## including openshift-logging
│  │     ├── collector-xxxxx.describe
│  │     ├── collector-yyyyy.describe
│  │     └── configmap_collector-config_vector.toml
├── cluster-scoped-resources
│  ├── apiextensions.k8s.io
│  │  └── customresourcedefinitions
│  │     ├── [all_CRDs]
│  ├── config.openshift.io
│  │  ├── [...]
│  ├── console.openshift.io
│  │  └── consoleplugins
│  │     ├── logging-view-plugin.yaml
│  │     └── [...]
│  └── core
│     ├── nodes
│     │  └── ip-10-0-146-180.eu-west-1.compute.internal.yaml
│     └── persistentvolumes
│        └── pvc-0a8d65d9-54aa-4c44-9ecc-33d9381e41c1.yaml
│  ├── machineconfiguration.openshift.io
│  │  └── machineconfigpools
│  │     └── [...]
│  ├── [...]
│  ├── observability.openshift.io
│  │  └── uiplugins
│  │     └── logging.yaml
│  └── [...]
├── [...]
├── event-filter.html
├── gather-debug.log
└── namespaces
│  ├── [namespace_name]       ## including openshift-logging
│  │  ├── apps
│  │  │  ├── daemonsets.yaml
│  │  │  ├── deployments.yaml
│  │  │  ├── replicasets.yaml
│  │  │  └── statefulsets.yaml
│  │  ├── batch
│  │  │  ├── cronjobs.yaml
│  │  │  └── jobs.yaml
│  │  ├── core
│  │  │  ├── configmaps
│  │  │  │  ├── collector-config.yaml
│  │  │  │  ├── collector-trustbundle.yaml
│  │  │  │  ├── kube-root-ca.crt.yaml
│  │  │  │  ├── logging-loki-ca-bundle.yaml
│  │  │  │  ├── logging-loki-config.yaml
│  │  │  │  ├── logging-loki-gateway-ca-bundle.yaml
│  │  │  │  ├── logging-loki-gateway.yaml
│  │  │  │  └── openshift-service-ca.crt.yaml
│  │  │  ├── configmaps.yaml
│  │  ├── [...]
│  │  │  ├── persistentvolumeclaims.yaml
│  │  │  ├── pods
│  │  │  │  ├── cluster-logging-operator-xxxxxxxxxx-xxxxx.yaml
│  │  │  │  ├── collector-xxxxx.yaml
│  │  │  │  ├── collector-xxxxx.yaml
│  │  │  │  ├── collector-xxxxx.yaml
│  │  │  │  ├── collector-xxxxx.yaml
│  │  │  │  ├── collector-xxxxx.yaml
│  │  │  │  ├── logfilesmetricexporter-xxxxx.yaml
│  │  │  │  ├── logfilesmetricexporter-xxxxx.yaml
│  │  │  │  ├── logfilesmetricexporter-xxxxx.yaml
│  │  │  │  ├── logfilesmetricexporter-xxxxx.yaml
│  │  │  │  ├── logfilesmetricexporter-xxxxx.yaml
│  │  │  │  ├── logging-loki-compactor-0.yaml
│  │  │  │  ├── logging-loki-distributor-xxxxxxxxxx-xxxxx.yaml
│  │  │  │  ├── logging-loki-gateway-xxxxxxxxxx-xxxxx.yaml
│  │  │  │  ├── logging-loki-gateway-xxxxxxxxxx-xxxxx.yaml
│  │  │  │  ├── logging-loki-index-gateway-0.yaml
│  │  │  │  ├── logging-loki-ingester-0.yaml
│  │  │  │  ├── logging-loki-querier-xxxxxxxxxx-xxxxx.yaml
│  │  │  │  └── logging-loki-query-frontend-xxxxxxxxxx-xxxxx.yaml
│  │  │  ├── pods.yaml
│  │  ├── [...]
│  │  ├── logging.openshift.io
│  │  │  └── logfilemetricexporters
│  │  │     └── instance.yaml
│  │  ├── loki.grafana.com
│  │  │  └── lokistacks
│  │  │     └── logging-loki.yaml
│  │  ├── [...]
│  │  ├── observability.openshift.io
│  │  │   └── clusterlogforwarders
│  │  │       └── collector.yaml
│  │  ├── openshift-logging.yaml
│  │  ├── operators.coreos.com
│  │  │  ├── clusterserviceversions
│  │  │  │  ├── cluster-logging.v6.1.1.yaml
│  │  │  │  ├── cluster-observability-operator.0.4.1.yaml
│  │  │  │  ├── elasticsearch-operator.v5.8.16.yaml
│  │  │  │  ├── jaeger-operator.v1.62.0-1.yaml
│  │  │  │  ├── kiali-operator.v1.89.8.yaml
│  │  │  │  ├── loki-operator.v6.1.1.yaml
│  │  │  │  └── servicemeshoperator.v2.6.4.yaml
│  │  │  ├── installplans
│  │  │  │  └── install-bb9t6.yaml
│  │  │  └── subscriptions
│  │  │     └── cluster-logging.yaml
│  │  ├── pods
│  │  │  ├── cluster-logging-operator-xxxxxxxxxx-xxxxx
│  │  │  │  ├── cluster-logging-operator
│  │  │  │  │  └── cluster-logging-operator
│  │  │  │  │     └── logs
│  │  │  │  │        ├── current.log
│  │  │  │  │        ├── previous.insecure.log
│  │  │  │  │        └── previous.log
│  │  │  │  └── cluster-logging-operator-xxxxxxxxxx-xxxxx.yaml
│  │  │  ├── collector-xxxxx
│  │  │  │  ├── [...]
│  │  │  └── [...]
│  │  └── [...]
│  └── [...]
│  ├── openshift-operators
│     ├── [...]
│  └── openshift-operators-redhat
│     ├── [...]
└── timestamp
