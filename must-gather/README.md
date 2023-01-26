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
- The openshift-operators-redhat namespace and its children objects
- The cluster-logging install objects
- All cluster-logging CRD's definitions
- All nodes objects
- All persistent volumes objects
- Custom logs, configurations and health status per component, i.e. collection, logStore, curation, visualization

In order to get data about other parts of the cluster (not specific to cluster-logging) you should
run `oc adm must-gather` (without passing a custom image). Run `oc adm must-gather -h` to see more options.

Example must-gather for cluster-logging output:
```
├── cluster-logging
│  ├── clo
│  │  ├── cluster-logging-operator-74dd5994f-6ttgt
│  │  ├── clusterlogforwarder_cr
│  │  ├── cr
│  │  ├── csv
│  │  ├── deployment
│  │  └── logforwarding_cr
│  ├── collector
│  │  ├── fluentd-2tr64
│  ├── eo
│  │  ├── csv
│  │  ├── deployment
│  │  └── elasticsearch-operator-7dc7d97b9d-jb4r4
│  ├── es
│  │  ├── cluster-elasticsearch
│  │  │  ├── aliases.cat
│  │  │  ├── health.cat
│  │  │  ├── hot_threads.txt
│  │  │  ├── indices.cat
│  │  │  ├── indices_size.cat
│  │  │  ├── latest_documents.json
│  │  │  ├── nodes.cat
│  │  │  ├── nodes_state.json
│  │  │  ├── nodes_stats.json
│  │  │  ├── pending_tasks.cat      (iff cluster health not green)
│  │  │  ├── recovery.cat           (iff cluster health not green)
│  │  │  ├── shards.cat             (iff cluster health not green)
│  │  │  ├── thread_pool.cat
│  │  │  └── unassigned_shards.cat  (iff cluster health not green)
│  │  ├── cr
│  │  ├── elasticsearch-cdm-lp8l38m0-1-794d6dd989-4jxms
│  │  └── logs
│  │     ├── elasticsearch-cdm-lp8l38m0-1-794d6dd989-4jxms
│  ├── install
│  │  ├── co_logs
│  │  ├── install_plan
│  │  ├── olmo_logs
│  │  └── subscription
│  └── kibana
│     ├── cr
│     ├── kibana-9d69668d4-2rkvz
├── cluster-scoped-resources
│  └── core
│     ├── nodes
│     │  ├── ip-10-0-146-180.eu-west-1.compute.internal.yaml
│     └── persistentvolumes
│        ├── pvc-0a8d65d9-54aa-4c44-9ecc-33d9381e41c1.yaml
├── event-filter.html
├── gather-debug.log
└── namespaces
   ├── openshift-logging
   │  ├── apps
   │  │  ├── daemonsets.yaml
   │  │  ├── deployments.yaml
   │  │  ├── replicasets.yaml
   │  │  └── statefulsets.yaml
   │  ├── batch
   │  │  ├── cronjobs.yaml
   │  │  └── jobs.yaml
   │  ├── core
   │  │  ├── configmaps.yaml
   │  │  ├── endpoints.yaml
   │  │  ├── events
   │  │  │  ├── elasticsearch-delete-app-1596020400-gm6nl.1626341a296c16a1.yaml
   │  │  │  ├── elasticsearch-delete-audit-1596020400-9l9n4.1626341a2af81bbd.yaml
   │  │  │  ├── elasticsearch-delete-infra-1596020400-v98tk.1626341a2d821069.yaml
   │  │  │  ├── elasticsearch-rollover-app-1596020400-cc5vc.1626341a3019b238.yaml
   │  │  │  ├── elasticsearch-rollover-audit-1596020400-s8d5s.1626341a31f7b315.yaml
   │  │  │  ├── elasticsearch-rollover-infra-1596020400-7mgv8.1626341a35ea59ed.yaml
   │  │  ├── events.yaml
   │  │  ├── persistentvolumeclaims.yaml
   │  │  ├── pods.yaml
   │  │  ├── replicationcontrollers.yaml
   │  │  ├── secrets.yaml
   │  │  └── services.yaml
   │  ├── openshift-logging.yaml
   │  ├── pods
   │  │  ├── cluster-logging-operator-74dd5994f-6ttgt
   │  │  │  ├── cluster-logging-operator
   │  │  │  │  └── cluster-logging-operator
   │  │  │  │     └── logs
   │  │  │  │        ├── current.log
   │  │  │  │        ├── previous.insecure.log
   │  │  │  │        └── previous.log
   │  │  │  └── cluster-logging-operator-74dd5994f-6ttgt.yaml
   │  │  ├── cluster-logging-operator-registry-6df49d7d4-mxxff
   │  │  │  ├── cluster-logging-operator-registry
   │  │  │  │  └── cluster-logging-operator-registry
   │  │  │  │     └── logs
   │  │  │  │        ├── current.log
   │  │  │  │        ├── previous.insecure.log
   │  │  │  │        └── previous.log
   │  │  │  ├── cluster-logging-operator-registry-6df49d7d4-mxxff.yaml
   │  │  │  └── mutate-csv-and-generate-sqlite-db
   │  │  │     └── mutate-csv-and-generate-sqlite-db
   │  │  │        └── logs
   │  │  │           ├── current.log
   │  │  │           ├── previous.insecure.log
   │  │  │           └── previous.log
   │  │  ├── elasticsearch-cdm-lp8l38m0-1-794d6dd989-4jxms
   │  │  ├── elasticsearch-delete-app-1596030300-bpgcx
   │  │  │  ├── elasticsearch-delete-app-1596030300-bpgcx.yaml
   │  │  │  └── indexmanagement
   │  │  │     └── indexmanagement
   │  │  │        └── logs
   │  │  │           ├── current.log
   │  │  │           ├── previous.insecure.log
   │  │  │           └── previous.log
   │  │  ├── fluentd-2tr64
   │  │  │  ├── fluentd
   │  │  │  │  └── fluentd
   │  │  │  │     └── logs
   │  │  │  │        ├── current.log
   │  │  │  │        ├── previous.insecure.log
   │  │  │  │        └── previous.log
   │  │  │  ├── fluentd-2tr64.yaml
   │  │  │  └── fluentd-init
   │  │  │     └── fluentd-init
   │  │  │        └── logs
   │  │  │           ├── current.log
   │  │  │           ├── previous.insecure.log
   │  │  │           └── previous.log
   │  │  ├── kibana-9d69668d4-2rkvz
   │  │  │  ├── kibana
   │  │  │  │  └── kibana
   │  │  │  │     └── logs
   │  │  │  │        ├── current.log
   │  │  │  │        ├── previous.insecure.log
   │  │  │  │        └── previous.log
   │  │  │  ├── kibana-9d69668d4-2rkvz.yaml
   │  │  │  └── kibana-proxy
   │  │  │     └── kibana-proxy
   │  │  │        └── logs
   │  │  │           ├── current.log
   │  │  │           ├── previous.insecure.log
   │  │  │           └── previous.log
   │  └── route.openshift.io
   │     └── routes.yaml
   └── openshift-operators-redhat
      ├── ...
```
