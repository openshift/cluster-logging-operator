# Cluster Logging Operator

## Overview

The CLO (Cluster Logging Operator) provides a set of APIs to control collection and forwarding of logs from
all pods and nodes in a cluster.  This includes *application* logs (from regular
pods), *infrastructure* logs (from system pods and node logs), and *audit* logs
(special node logs with legal/security implications)

The CLO does not collect or forward logs itself: it starts, configures, monitors
and manages the components that do the work.

CLO currently uses Fluentd as collector/forwarder, Elasticsearch as default store and Kibana for visualization.
The goal is to *encapsulate* those technologies behind APIs so that:

1. The user has less to learn, and has a simpler experience to control logging.
2. These technologies can be replaced in the future without affecting the user experience.

The CLO can also *forward* logs over multiple protocols, to multiple types of log stores, on- or off-cluster

The CLO *owns* the following APIs:

* ClusterLogging: Top level control of cluster-wide logging resources
* ClusterLogForwarder: Configure forwarding of logs to external sources
* Collector: configure global collector resource use.

To install a released version of cluster logging see the [Openshift Documentation](https://docs.openshift.com/)

To experiment or contribute to the development of cluster logging, see [HACKING.md](./docs/HACKING.md) and [REVIEW.md](./docs/REVIEW.md)

To debug the cluster logging stack, see [README.md](./must-gather/README.md)

To find currently known Cluster Logging Operator issues with work-arounds, see the [Troubleshooting](./docs/troubleshooting.md) guide.
