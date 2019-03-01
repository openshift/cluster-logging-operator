# cluster-logging-operator
An operator to support OKD aggregated cluster logging.  Cluster logging configuration information
is found in the [configuration](./docs/configuration.md) documentation.

## Quick Start
To get started with cluster logging and the `cluster-logging-operator`:
* Ensure Docker is installed on your local system
```
$ oc login $CLUSTER -u $ADMIN_USER -p $ADMIN_PASSWD
$ REMOTE_CLUSTER=true make deploy-example
```
This will stand up a cluster logging stack named 'example'.


## Hacking

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

```oc login --username=kubeadmin --password=$( cat ../installer/auth/kubeadmin_password )```

The user should already have `cluster-admin` rights.

**Note:** If you are using `REMOTE_REGISTRY=true`, ensure you have `docker` package installed and `docker` daemon up and running on the workstation you are running these commands.


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
|`deploy-setup`|Setup the OKD cluster to support cluster-logging without deploying an instance of the `cluster-logging-operator`|
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
$ make deploy-image && IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc/openshift/origin-cluster-logging-operator:latest make test-e2e
```
