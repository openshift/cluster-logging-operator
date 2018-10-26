# cluster-logging-operator
Operator to support OKD cluster logging

## Hacking

### Running the operator

Running locally outside an OKD cluster:
```
 $ REPO_PREFIX=openshift/ \
   IMAGE_PREFIX=origin- \
   OPERATOR_NAME=cluster-logging-operator \
   WATCH_NAMESPACE=openshift-logging \
   KUBERNETES_CONFIG=/etc/origin/master/admin.kubeconfig \
   go run cmd/cluster-logging-operator/main.go`
```
### `make` targets
Various `make` targets are included to simplify building and deploying the operator
from the repository root directory.  Hacking and deploying the operator assumes:
* a running OKD cluster
* `oc` binary in your `$PATH`
* Logged into a cluster with an admin user who has the cluster role of `cluster-admin`
* Access to a clone of the `openshift/elasticsearch-operator` repository
* `make` targets are executed from the `openshift/cluster-logging-operator` root directory

The deployment can be optionally modified using any of the following:

*  `IMAGE_BUILDER` is the command to build the container image (default: `docker build`)
*  `OC` is the openshift binary to use to deploy resources (default: `oc` in path)

#### Full Deploy
Deploys all resources and creates an instance of the `cluster-logging-operator`
```
$ make deploy IMAGE_BUILDER=imagebuilder
```

#### Undeploy
Removes all `cluster-logging` resources and `openshift-logging` project
```
$ make undeploy
```
#### Deploy Prerequisites
Setup the OKD cluster to support cluster-logging without deploying an instance of
the `cluster-logging-operator`
```
$ make deploy-setup
```

## Testing

### E2E Testing
This test assumes the cluster-logging component images were already pushed 
to the OKD cluster and can be found at $docker_registry_ip/openshift/$component

**Note:** This test will fail if the component images are not pushed to the cluster
on which the operator runs.

Running the tests directly from the repo directory:
```
$ make test-e2e
```
