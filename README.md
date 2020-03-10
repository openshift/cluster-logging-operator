# cluster-logging-operator - TESTDONOTMERGE
An operator to support OKD aggregated cluster logging.  Cluster logging configuration information
is found in the [configuration](./docs/configuration.md) documentation.

## Quick Start
To get started with cluster logging and the `cluster-logging-operator`:
* Ensure Docker is installed on your local system [(**Note** for Fedora 31)](./docs/fedora31.md)
* Ensure skopeo is installed, e.g. `sudo dnf install -y skopeo`
```
$ oc login $CLUSTER -u $ADMIN_USER -p $ADMIN_PASSWD
$ REMOTE_CLUSTER=true make deploy-example
```
This will stand up a cluster logging stack named 'example'.


## Hacking

### Building a Universal Base Image (UBI) based image

You must first `oc login api.ci.openshift.org`.  You'll need these credentials in order
to pull images from the UBI registry.

The image build process for UBI based images uses a private yum repo.
In order to use the private yum repo, you will need access to
https://github.com/openshift/release/blob/master/ci-operator/infra/openshift/release-controller/repos/ocp-4.1-default.repo
and
https://github.com/openshift/shared-secrets/blob/master/mirror/ops-mirror.pem
Note that the latter is private and requires special permission to access.

The best approach is to clone these repos under `$GOPATH/src/github.com/openshift`
which the build scripts will pick up automatically.  If you do not, the build script
will attempt to clone them to a temporary directory.

### Running a full integration test with a 4.x cluster

If you have a local clone of [origin-aggregated-logging](https://github.com/openshift/origin-aggregated-logging)
under `$GOPATH/github.com/openshift/origin-aggregated-logging` you can use its `hack/get-cluster-run-tests.sh`
to build the logging, elasticsearch-operator, and cluster-logging-operator images; deploy a 4.x cluster;
push the images to the cluster; deploy logging; and launch the logging CI tests.


### Running the operator

Running locally outside an OKD cluster:
```
 $ make run
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
* various other commands such as `imagebuilder` and `operator-sdk`. The
   [sdk_setup.sh](https://raw.githubusercontent.com/openshift/origin-aggregated-logging/master/hack/sdk_setup.sh) will setup your environment.

The deployment can be optionally modified using any of the following:

| Env Var | Default | Description|
|---------|---------|------------|
|`IMAGE_BUILDER`|`imagebuilder`| The command to build the container image|
|`EXCLUSIONS`|none|The list of manifest files that should will be ignored|
|`OC`|`oc` in `PATH`| The openshift binary to use to deploy resources|
|`REMOTE_REGISTRY`|false|`true` if you are running the cluster on a different machine than the one you are developing on|
|`PUSH_USER`|none|The name of the user e.g. `kubeadmin` used to push images to the remote registry|
|`PUSH_PASSWORD`|none|The password for `PUSH_USER`|
|`SKIP_BUILD`|false|`true` if you are pushing an image to a remote registry and want to skip building it again|

**Note:** Use `REMOTE_REGISTRY=true`, for example, if you are running a cluster in a
    local libvirt or minishift environment; you may want to build the image on the host
    and push them to the cluster running in the VM. This requires a username with a password (i.e. not the default `system:admin` user).
    If your cluster was deployed with the `allow_all` identity provider, you can:
create a user and assign it rights:
```
    oc login --username=admin --password=admin
    oc adm policy add-cluster-role-to-user cluster-admin admin
```

If you used the new `openshift-installer`, it creates a user named `kubeadmin`
    with the password in the file `installer/auth/kubeadmin_password`.

To push the image to the remote registry, set `REMOTE_REGISTRY=true`, `PUSH_USER=kubeadmin`,
`KUBECONFIG=/path/to/installer/auth/kubeconfig`, and
`PUSH_PASSWORD=$( cat /path/to/installer/auth/kubeadmin_password )`.  You do not have to
`oc login` to the cluster.  If you already built the image, use `SKIP_BUILD=true` to do
a push only.

```bash
REMOTE_REGISTRY=true PUSH_USER=kubeadmin SKIP_BUILD=true \
KUBECONFIG=/path/to/installer/auth/kubeconfig \
PUSH_PASSWORD=$( cat /path/to/installer/auth/kubeadmin_password ) \
make deploy-image
```

**Note:** If you are using `REMOTE_REGISTRY=true`, ensure you have `docker` package installed and `docker` daemon up and running on the workstation you are running these commands.  Also ensure you have `podman` (`docker` may be deprecated in the future).


**Note:**  If while hacking you find your changes are not being applied, use
`docker images` to see if there is a local version of the `cluster-logging-operator`
on your machine which may being used by the cluster instead of the one pushed to
the docker registry.  You may need to delete it (e.g. `docker rmi $IMAGE`)

Following is a list of some of the existing `make` targets to facilitate deployment:

|Target|Description|
|------|-----------|
|`deploy-example`|Deploys all resources and creates an instance of the `cluster-logging-operator`, creates a ClusterLogging custom resource named `example`, starts the Elasticsearch, Kibana, and Fluentd pods running.|
|`deploy`|Deploys all resources and creates an instance of the `cluster-logging-operator`, but does not create the CR nor start the components running.|
|`undeploy`|Removes all `cluster-logging` resources and `openshift-logging` project|
|`deploy-image`|If you already have a remote cluster, and you just want to rebuild and redeploy the image to the remote cluster.  Use `make image` instead if you only want to build the image but not push it.|


## Testing

### E2E Testing
This test assumes:
* the cluster-logging-operator image is available
* the cluster-logging component images are available (i.e. $docker_registry_ip/openshift/$component)

**Note:** This test will fail if the images are not pushed to the cluster
on which the operator runs or can be pulled from a visible registry.

**Note:** It is necessary to set the `IMAGE_CLUSTER_LOGGING_OPERATOR` environment variable to a valid pull spec
in order to run this test against local changes to the `cluster-logging-operator`. For example:
```
$ make deploy-image && IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest make test-e2e
```
