# cluster-logging-operator
Operator to support OKD cluster logging

## Hacking

### Running the operator

Running locally outside an OKD cluster:
```
 $ ELASTICSEARCH_IMAGE=docker.io/openshift/origin-logging-elasticsearch5:latest \
   FLUENTD_IMAGE=docker.io/openshift/origin-logging-fluentd:latest \
   KIBANA_IMAGE=docker.io/openshift/origin-logging-kibana5:latest \
   CURATOR_IMAGE=docker.io/openshift/origin-logging-curator5:latest \
   OAUTH_PROXY_IMAGE=docker.io/openshift/oauth-proxy:latest \
   RSYSLOG_IMAGE=docker.io/viaq/rsyslog:latest \
   OPERATOR_NAME=cluster-logging-operator \
   WATCH_NAMESPACE=openshift-logging \
   KUBERNETES_CONFIG=/etc/origin/master/admin.kubeconfig \
   go run cmd/cluster-logging-operator/main.go
```
### `make` targets
Various `make` targets are included to simplify building and deploying the operator
from the repository root directory, all of which are not listed here.  Hacking and
deploying the operator assumes:
* a running OKD cluster
* `oc` binary in your `$PATH`
* Logged into a cluster with an admin user who has the cluster role of `cluster-admin`
* Access to a clone of the `openshift/elasticsearch-operator` repository
* `make` targets are executed from the `openshift/cluster-logging-operator` root directory

The deployment can be optionally modified using any of the following:

*  `IMAGE_BUILDER` is the command to build the container image (default: `docker build`)
*  `EXCLUSIONS` is list of manifest files that should be ignored (default: '')
*  `OC` is the openshift binary to use to deploy resources (default: `oc` in path)

**Note:**  If while hacking you find your changes are not being applied, use
`docker images` to see if there is a local version of the `cluster-logging-operator`
on your machine which may being used by the cluster instead of the one pushed to
the docker registry.  You may need to delete it (e.g. `docker rmi $IMAGE`)

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
