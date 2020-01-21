# Unmanaged Configurations for Cluster Logging
This document describes the various settings that can be tuned in cluster logging but are not
managed by the `cluster-logging-operator`.  Administrators assume responsibility of managing and
maintaining the deployment by modifying the cluster logging CRD to be unmanaged.

```
  spec:
    managementState: "Unmanaged"
```
**NOTE:** Many of the configurations described in the following sections could previously be modified by
updating the ansible inventory file.  These changes are no longer supported in a managed capacity and are
subject to change in the future.

## Fluentd
### MERGE_JSON_LOG
This setting configures Fluentd to inspect each log message to determine if it's format is 'JSON' and to merge
it into the JSON payload document posted to Elasticsearch.  This setting is `false` by default and is a change
from 3.11 and earlier deployments where it was `true`.  

Enable the feature by:
```
oc set env dc/fluentd MERGE_JSON_LOG=true
```
**NOTE:** Enabling this feature comes with [risks](https://github.com/openshift/origin-aggregated-logging/issues/1492) summarized here:
* Possible log loss due to Elasticsearch rejecting documents due to inconsistent type mappings
* Potential buffer storage leak caused by rejected message cycling
* Overwrite of data for field with same names


## ElasticSearch
### Increase vm.max_map_count
In order to deploy an instance of ElasticSearch on a selected node of the cluster you will need to inscrease this Kernel parameter `vm.max_map_count` which is set by default to `65530` and on the ElasticSearch case you need to set it up to `262144`. This could be done in multiple ways but we recommend to use [Cluster Node Tuning Operator](https://github.com/openshift/cluster-node-tuning-operator). 

- This is the Tuned Profile [you need to apply](https://github.com/openshift/cluster-node-tuning-operator/blob/master/examples/elasticsearch.yaml) with Cluster Node Tuning Operator.
```
apiVersion: tuned.openshift.io/v1
kind: Tuned
metadata:
  name: elasticsearch
  namespace: openshift-cluster-node-tuning-operator
spec:
  profile:
  - data: |
      [main]
      summary=Optimize systems running ES on OpenShift nodes
      include=openshift-node
      [sysctl]
      vm.max_map_count=262144
    name: openshift-node-elasticsearch
  recommend:
  - match:
    - label: tuned.openshift.io/elasticsearch
      type: pod
    priority: 30
    profile: openshift-node-elasticsearch
```