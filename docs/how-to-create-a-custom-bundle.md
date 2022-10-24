# Creating and Installing a Custom Bundle

You can build a cluster-logging-operator directly from source code,
and create a *bundle* image that will allow you to deploy it to your cluster.
There are two images:

* operator image: Contains your build of the operator.
* bundle image: Contains instructions for OLM to install you operator.

## Log-in to Image Repositories

You need write permission to a *public* image registry (for example `quay.io/somebody`) for your operator and bundle images.

**NOTE**: The images you create must be *public* so cluster are able to pull them.
If you use `quay.io` you may need to mark newly-created images as public in "organization settings".

You also need to log in to the red Hat registry for related images.
Use your Red Hat customer portal user name and password.
See https://access.redhat.com/RegistryAuthentication for more.

``` bash
podman login quay.io/somebody
podman login registry.redhat.io # Use Red Hat customer portal username/password
```

**NOTE**: this document uses `podman` but you can substitute `docker` instead, they are equivalent.

## Build and Push Operator and Bundle Images

### Clone the cluster-logging-operator repository

``` bash
git clone http://github.com/openshift/cluster-logging-operator
cd cluster-logging-operator
```

### Create a custom overlay

An *overlay* provides parameters to generate customized operator and bundle images.
You can modify:

* The image name/tag for the operator image (the bundle will have the same name with "-bundle" added)
* The image names/tags for related images used by the operator

See [../config/overlays/README.md](../config/overlays/README.md) for instructions to create your overlay.

### Build and push operator and bundle images

Assuming your overlay is in directory `config/overlays/custom`, in the repository root directory:

```bash
make OVERLAY=config/overlays/custom deploy-image
make OVERLAY=config/overlays/custom deploy-bundle
```

This will push your operator and bundle images to the repository specified by your overlay.


## Installing the Bundle Image

You need `operator-sdk` which you can get here: https://sdk.operatorframework.io/docs/installation

**NOTE**: use `operator-sdk.v1.22.2` or later. Some earlier versions cause problems.

To install the operator in your cluster:

```
oc create namespace openshift-logging
operator-sdk run bundle -n openshift-logging <YOUR_BUNDLE_IMAGE_NAME>
```

To uninstall:

``` bash
operator-sdk cleanup --delete-all cluster-logging
oc delete namespace openshift-logging
```
