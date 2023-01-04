# Configuring and Tuning Cluster Logging

OKD cluster logging is configurable by modifying the CustomResource deployed
in the **openshift-logging** namespace.  The [sample](../../hack/cr.yaml) resource
defines a complete cluster logging deployment that includes all the subcomponents
of the logging stack to collect, store and visualize logs.  The `cluster-logging-operator`
watches the a `ClusterLogging` CustomResources and adjusts the logging deployment accordingly.

**Note:** The cluster-logging-operator **will only** handle a `ClusterLogging` instance named `instance`.  This
is to enforce a single ClusterLogging deployment for the entire OKD cluster.

## Management State
Cluster logging is managed by the `cluster-logging-operator`.  It can be placed
into an unmanaged state allowing an administrator to assume full control of individual
component configurations and upgrades by changing the `managementState` from `Managed` to `Unmanaged`.
**Note:** An unmanaged deployment will not receive updates until the `ClusterLogging` custom resource is placed back into a managed state.
```
  spec:
    managementState: "Managed"
```

## Component Images
The image for each subcomponent are overwriteable by modifying the appropriate
environment variable in the `cluster-logging-operator` Deployment.  Each variable
takes a full pull spec of the image:

* `RELATED_IMAGE_FLUENTD`
* `LOGFILEMETRICEXPORTER_IMAGE`
* `VECTOR_IMAGE`

## Common Configurations
The following configuration options apply generally to all components defined in for a ClusterLogging object (e.g. logStore, visualization, etc).

### Memory and CPU
Each component specification allows for adjustments to both the CPU and
memory limits.  This is defined by modifying the `resources`
block with valid memory (e.g. 16Gi) and CPU values (e.g 1):
```
  spec:
    logStore:
      type: "elasticsearch"
      elasticsearch:
        resources:
          limits:
            cpu:
            memory:
          requests:
            cpu:
            memory:
```
### Node Selectors
Each component specification allows the component to target a specific node.  This is defined
by modifying the `nodeSelector` block to include key/value pairs that correspond
to specifically labeled nodes:
```
  spec:
    logStore:
      type: "elasticsearch"
      elasticsearch:
        nodeSelector:
          node-type: infra
          cluster-logging-component: es
```

## Data Aggregation and Storage
An Elasticsearch cluster is responsible for log aggregation.  Following is a sample
of the spec:
```
  spec:
    logStore:
      type: "elasticsearch"
      elasticsearch:
        nodeCount: 3
        storage:
          storageClassName: "gp2"
          size: "200G"
        redundancyPolicy: "SingleRedundancy"
```
This example specifies each data node in the cluster will be bound to a `PersistentVolumeClaim` that
requests "200G" of "gp2" storage.  Additionally, each primary shard will be backed by a single replica.

### Backing storage
The `cluster-logging-operator` will create a `PersistentVolumeClaim` for each data node in the Elasticsearch cluster
using the size and storage class name from the spec.  

**Note**: Omission of the storage block will result in a deployment that includes ephemeral storage only. E.g.:
```
  spec:
    logStore:
      type: "elasticsearch"
      elasticsearch:
        resources
        nodeCount: 3
        redundancyPolicy: "SingleRedundancy"
```

### Data Redundancy
The policy that defines how shards are replicated across data nodes in the cluster

|Policy | Description |
|----- | ----- |
|`FullRedundancy` | The shards for each index are fully replicated to every data node|
|`MultipleRedundancy`|The shards for each index are spread over half of the data nodes|
|`SingleRedundancy`| A single copy of each shard. Logs are always available and recoverable as long as at least two data nodes exist |
|`ZeroRedundancy`| No copies of any shards.  Logs may be unavailable (or lost) in the event a node is down or fails|

## Log Collectors
Log collectors are deployed as a Daemonset to each node in the OKD cluster.  Following are the
supported log collectors for Cluster Logging:
* Fluentd - The default log collector based on Fluentd.
* Vector - An alterate log collector based on Vector.

```
  spec:
    collection:
      type: "fluentd"
      resources:
        limits:
          cpu:
          memory:
        requests:
          cpu:
          memory:
```

## Kibana and Visualization
Kibana is fronted by an oauth-proxy container, which additionally allows memory and CPU
configuration.
```
  spec:
    visualization:
      type: "kibana"
      kibana:
        replicas: 1
        resources:
        proxy:
          resources:
            limits:
              cpu:
              memory:
            requests:
              cpu:
              memory:

```
## Curator and Data Curation
**Note:** Curator is deprecated and no longer supported
