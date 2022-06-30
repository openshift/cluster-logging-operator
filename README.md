# Cluster Logging Operator
An operator to support aggregated cluster logging in an OKD or kubernetes cluster.

## Build and Deploy

See the [HACKING guide](docs/HACKING.md) for instructions to build and deploy the operator.

## Overview

The CLO (Cluster Logging Operator) provides a set of APIs to control
collection and forwarding of logs from all pods and nodes in a
cluster.  This includes *application* logs (from regular pods),
*infrastructure* logs (from system pods and node logs), and *audit*
logs (logs that track activity on the node or cluster)

The CLO is an _operator_. It does not do the work of collecting or
forwarding logs at runtime; it starts, configures, monitors and
manages the components that do the work according to the API
configuration.

CLO currently uses:
* Fluentd as collector/forwarder.
* Elasticsearch as default store.
* Kibana for visualization.

The goal is to *encapsulate* those technologies behind APIs so that:

1. The user has less to learn, and has a simpler experience to control logging.
2. These technologies can be replaced in the future without affecting the user experience.

The CLO can also *forward* logs over multiple protocols, to multiple
types of log stores, on- or off-cluster

The CLO *owns* the following APIs:

* ClusterLogging: Top level control of cluster-wide logging resources
* ClusterLogForwarder: Configure forwarding of logs to external sources

## Relevant documentation

For configuration information see [configuration](./docs/configuration.md) documentation.

To install a GA released version of cluster logging see the [Openshift Documentation](https://docs.openshift.com/container-platform/latest/logging/cluster-logging.html)

To build and deploy the operator for yourself, or to contribute to the project, see [HACKING.md](./docs/HACKING.md). To submit a pull request see [REVIEW.md](./docs/REVIEW.md)

To debug problems with the cluster logging stack, see the [must-gather tool](./must-gather/README.md)

To find currently known Cluster Logging Operator issues with work-arounds, see the [troubleshooting](./docs/troubleshooting.md) guide.


## Introduction to concepts
This is a _very_ brief overview of some relevant concepts for the unfamiliar.
There are lots of good resources available for more details.
Starting points include https://docs.openshift.com/ and  https://kubernetes.io/

### Deployment concepts

There are two ways to deploy the operator:

_Plain kubernetes_:
The operator is a collection of resources (deployment, service accounts, roles, etc.)
These are just YAML files, and can be deployed directly with `kubectl apply`

_Operator Lifecycle Manager_:
OLM allows automatic updates to operators by subcribing to _channels_.
It provides version management, control of upgrades, and other features.
OLM requires a specially formatted _bundle image_ containing the operator's resources, and some extra channel, version and upgrade information.

### Operator concepts

This is a _very_ brief overview of how operators actually work.
It assumes you are familiar with basic kubernetes concepts including custom resources.

An _operator_ consists of
* an executable manager program (in our case written in Go) packaged as an image
* a set of resources (YAML files) describing how the manager should run in a cluster.
* a set of _custom resource descriptions_ (CRDs) for the resources managed by the operator.

The operator watches the kubernetes API server for changes to its _custom resources_ (CRs).
In our case, that is the `ClusterLogging` and `ClusterLogForwarder` CRs.

When a CR is created or modified, the operator looks at the `spec` section.
It creates _resources_ (pods, deployments, daemonsets, etc.) to run its _operands_ (in our case `fluentd`) to bring about the state requested in the `spec`.

The operands do the real work, the operator's job is to make sure they are configured and running as requested by the CR `spec`, and to update the CR `status` to reflect the actual state of its operands.

For example: consider this ClusterLogForwarder configuration

``` yaml
---
spec:
  outputs:
  - name: lokiOut
    type: loki
    url: http://grover.example.com/something
  - name: kafkaOut
    type: kafka
    url: tcp://bert.example.com:12345
  pipelines:
  - inputRefs: [application]
    outputRefs: [fluentOut]
  - inputRefs: [infrastructure]
    outputRefs: [kafkaOut]
```

The operator will

1. Generate a fluent.conf configuration to send application logs to the loki URL and infrastructure logs to the kafka URL.
2. Create or restart a `daemonset` running fluentd with that configuration.
3. Update the CR `status` section to indicate it was successfully configured.

If the CR is modified later (e.g. using `kubectl apply`) the operator will react to
update its operands accordingly.

Normally an operator runs as a kubernetes 'deployment' on the cluster.
However operators do all their work via the kubernetes API,
so it is possible to run the operator executable _outside the cluster_ for debugging.
