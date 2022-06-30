# Hacking on the Cluster Logging Operator

## Before you start

You need:
1. *administrator* access to an openshift cluster, normally user kube:admin.
2. commands `git`, `make` and`podman` installed in your working environment.
3. *public* image repositories that you can push to for the images you will build.\
   For example you could use a free account on `quay.io` or `docker.io`.

You need two *public* image repositories called:
```
origin-cluster-logging-operator
origin-cluster-logging-operator-bundle
```

**WARNING**:
- Auto-created image repositories *may be private* depending on your account.
If in doubt, log-in to the repository website and pre-create the repos or set them to be public.
- Credentials and log-ins for your image repositories will time-out periodically and need to be refreshed.

## Quick start

Initialize your workspace:
```bash
make init REGISTRY=quay.io/myname
```

Build and deploy the operator:
```bash
make deploy
```

The logging operator is now running in your cluster!

* `make deploy` deploys resources directly to your cluster. It does not use the Operator Lifecycle Manager (OLM)
* `make deploy` is idempotent, you can run it to update the deployment without an undeploy.

## Deploy with OLM

To deploy via OLM, you need to create a "bundle" image.
Build and deploy an OLM bundle like this:

``` bash
make olm-deploy
```

**NOTE**: Once you have a bundle image, you can deploy it directly using
[`operator-sdk`](https://sdk.operatorframework.io/docs/installation/).
You don't need access to this repository to deploy it.

For example:
``` bash
oc create namespace openshift-logging # Make sure this namespace exists
operator-sdk run bundle -n openshift-logging --install-mode OwnNamespace quay.io/MY_QUAY_NAME/cluster-logging-operator-bundle:0.0.0-myname-mybranch
```

## Additional build tools

Kubernetes and Openshift tools used to build the images are automatically installed with `go install` by `make`.
The Makefile uses specific versions of the tools to avoid unexpected changes in behavior.
To update the versions used, see [../.bingo/README.md](../.bingo/README.md)

## Image Registries and Pull Secrets

To push images you need to be logged in to your registry with push access.
The cluster does _not_ need push access to deploy the images, only the user calling `make` does.

The cluster must have pull-access to the images you build.
The simplest way to arrange that is to use public repositories like quay.io or docker.io.
If you use private repositories, you must modify your cluster's pull-secret to give it pull access.

**NOTE**: podman can store and read secrets from multiple places, if you get into a confused state check all these files:

```bash
$XDG_RUNTIME_DIR/containers/auth.json
$HOME/.config/containers/auth.json
$HOME/.docker/config.json
$HOME/.dockercfg
```
See the podman man page for details.

## Running Tests

Major test groups can be run via `make` targets, selected tests can be run directly with `go test`.
For example:

``` bash
go test ./internal/k8shandler -v
```

For some functional and e2e tests you must set some environment variables provided by `make env1`.
For example:

``` bash
  export $(make env)
  LOG_LEVEL=2 go test -v test/functional/outputs
  make deploy deploy-elastic
  go test ./test/e2e/logforwarding/loki/forward_to_loki_test.go
```

# Overview of the project

The operator's "manager" is a Go executable `bin/cluster-logging-operator`
- Built from `main.go` and the `internal/...` packages.
- Unit tests are included in packages using the normal Go `_test.go` conventions.
- Tests are written using the `github.com/onsi/ginkgo` framework.

Kubernetes resources and kustomizations are stored in `config`
- `config/`: resources (Deployment, Roles, Service Account etc.) needed by the operator.
- `config/overlays`: modify the resources for different deployments, in particular release vs development.
  - `config/overlays/*: the YAML files in each overlay directory specify what is different about the overlay.

Additional tests that require a cluster in the `tests` directory.
- `tests/functional`: test pieces of the operator in isolation, requires a cluster but not a deployment of the operator.
- `tests/e2e`: test a deployed operator.

For OLM deployments you need a 'bundle' which describes the operator and the OLM channels to install it.
- `bundle/`: This is the release bundle that will be pushed to operatorhub when the operator is released.
- `develop/bundle`: This is a local developer bundle using your own image repository.
