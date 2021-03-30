# Hacking on the Cluster Logging Operator

## Preparation

* Clone this repository into `$GOPATH/github.com/openshift/cluster-logging-operator`
* Install `podman`. For example on Fedora:
  ```
  sudo dnf install -y podman
  ```

## Makefile tagets

Quick summary of main targets only, see below and the Makefile itself for more details.

* `make check`: Generate and format code, run unit tests and linter.

To build, deploy and test the CLO image in your own cluster:
* `make deploy`: Build CLO image and deploy to cluster with elasticsearch.
* `make undeploy`: Undo `make deploy`.
* `make undeploy-all`: Undeploys CLO and Elasticsearch operator.

To run CLO as a local process:
* `make run`: Run CLO as a local process. Does not require `make deploy`.
* `make debug`: Run CLO in the `dlv` debugger.

To run the e2e tests:
* `make test-e2e-local`: Run e2e tests using locally built images.
* `make test-e2e-olm`: Run e2e tests using CI latest images.

**Note**: e2e tests require a clean cluster starting point, `make undeploy-all` will provide it.

**Note**: There is a temporary issue with openshift base images using RHEL8
repositories that are not accessible to all. This will be fixed, but until it is
you can build using CentOS as the base instead, as follows:

```bash
    PATCH=Dockerfile.centos.patch make ...
```

## Testing Overview

There are two types of test:

* *Unit tests:* run in seconds, no cluster needed, verify internal behavior - behind the API. New code *should* be reasonably well covered by unit tests but we don't have a formal requirement.
* *E2E tests:* run in minutes to hour(s), require a cluster. There are several ways to run these:
   - *CI tests:* run automatically when you raise a PR in a controlled cluster environment. Can take hour(s) to get results.
   - *Deployed tests:* Build and deploy CLO image to your own cluster, allows you to run tests selectively.
   - *Local process:* requires a cluster with dependencies deployed (elasticsearch), but runs CLO as a *local process*. Faster and easier to debug.

All PRs should pass unit tests. CI runs all tests automatically and won't merge a PR till they pass, so it is OK to create a PR and let CI run the e2e tests if you've done enough testing to be confident. This allows for human review and feedback while CI tests are running.

To get a code review of unfinished work, create a PR with "[WIP]" at the start of the description. The CI system will not merge it until the "[WIP]" is removed and all tests are passing.

## Setting up a test cluster

Developers can use a public cluster provider or a local cluster using [Code Ready Containers](https://developers.redhat.com/products/codeready-containers/download).

### Pull secret

Pull secrets are tokens used to pull OCI images from multiple sources, such as quay.io or cloud.openshift.com.

**Note**: From now on we assume `$PULL_SECRET` is the path to your pull secret file, e.g.
  ```
  PULL_SECRET=~/my-pull-secret
  ```

1. In a browser, visit https://cloud.redhat.com/openshift/install#pull-secret (log in if necessary)
1. Select your preferred cluster deployment method.
1. Use the `download pull secret` button, save the file to `$PULL_SECRET`

### CI token

For development you also need an openshift CI repository token.

1. In a browser, visit https://api.ci.openshift.org/oauth/token/request (log in if needed)
1. Execute the`oc login ... --server=https://api.ci.openshift.org` command shown.\
   This will log you into the CI system and update your `KUBECONFIG` (default `~/.kube/config`) with CI credentials. You should see\
   `Logged into "https://api.ci.openshift.org:443" as "<user>" using the token provided.`
1. Add the CI registry token to your pull-secret file, and log out of the CI system
   ```
   oc registry login --to $PULL_SECRET
   oc logout
   ```

### Create a cluster

Follow instructions from  https://cloud.redhat.com/openshift/install to create your cluster.
Provide `$PULL_SECRET` as the pull-secret file when requested.

Log in to your cluster as a user with the `cluster-admin` role.
User `kubeadmin` is usually predefined with this role.

**IMPORTANT**: Pull secrets and CI tokens expire after a period of several weeks.
The symptoms are:

* `make` prints "unauthorized: authentication required" errors when building images.
* `make deploy` times out and `oc get events -A | grep Pull` shows `ImagePullBackoff` events.

When a CI token expires you can simply create a new cluster.
You can also update the pull secret of an existing cluster.

### Modify the pull-secret of an existing cluster

1. Repeat the steps in "Pull secret" and "CI token" above to create an up-to-date pull secret.
1. Use the script provided to encode and update the secret
	```
	../hack/update_pull_secret.sh $PULL_SECRET
	```

The secret is base64 encoded, read the script for details.

### Configuration for CRC

Download the crc command from [Code Ready Containers](https://developers.redhat.com/products/codeready-containers/download) and read the install instructions.

The following steps are recommended before calling `crc start`
```
# Increase memory - default 8k may not be enough
crc config set memory 12288
# my-pull-secret is the CRC pull secret with the openshift CI secret added, explained above.
crc config set pull-secret-file my-pull-secret
# This test sometimes fails incorrectly on systems where libvirtd is triggered by a socket.
crc config set skip-check-libvirt-running true
```

## More about `make run`

You can run the CLO as a local process, outside the cluster. This is *not* the
normal way to run an operator, and does not test all aspects of the CLO
(e.g. problems with the CSV or OLM interactions), but it has advantages:

* Fast edit/run/test cycle - runs from source code, skips some slow build/deploy steps.
* Directly access the CLO logs on stdout/stderr
* Control logging levels and other environment variables, e.g. `export LOG_LEVEL=5`
* Run CLO in a debugger, profiler or other development tools.

*How it works*: An operator is actually a cluster *client*. It watches for
changes to its own custom resources, and creates/updates other resources
accordingly. This can all be done from outside the cluster.

**Note:** You must run the `make deploy-elasticsearch-operator` as a separate target which will:
* Deploy the Elasticsearch CRD upon which the CLO depends
* Define ClusterLogForwarding to the default log store

Examples:
```
make run  # Run the CLO locally
make run-debug  # Run CLO under the dlv debugger
LOG_LEVEL=4 make run  # Run CLO with greater log verbosity
RUN_CMD=foo make run # Run CLO under imaginary "foo" debugger/profiler.
```

Note `make run` will not return until you terminate the CLO.

### More about `make test-e2e-olm`

This test assumes:
* the cluster-logging-catalog image is available
* the cluster-logging-operator image is available
* the cluster-logging component images are available (i.e. $docker_registry_ip/openshift/$component)

**Note:** This test will fail if the images are not pushed to the cluster
on which the operator runs or can be pulled from a visible registry.

**Note:** It is necessary to set the `IMAGE_CLUSTER_LOGGING_OPERATOR` environment variable to a valid pull spec
in order to run this test against local changes to the `cluster-logging-operator`. For example:
```
$ make deploy-image && IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest make test-e2e
```

**Note:** To skip cleanup of resources while hacking/debugging an E2E test apply `DO_CLEANUP=false`.

## Building a Universal Base Image (UBI) based image

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

## Updating Dependencies

Bumping the release and manifest versions typically require updating the `elasticsearch-operator` as well.
* `dep ensure -update github.com/openshift/elasticsearch-operator`

## Deploying without OLM
Production relies upon OLM to manage and control the operator deployment, permissions, etc. The manifest defines all the resources needed by OLM.  We can use this same manifest to generate a list of resources to deploy without using OLM.
```
make deploy-image
```
will produce output that should give you the pullspec on the cluster like:
```
image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest
```

which will allow you to use the script like:
```                                                                                                                                                                                                                                                                               
CLO_IMAGE=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
./hack/gen-olm-artifacts.py manifests/5.1/cluster-logging.v5.1.0.clusterserviceversion.yaml  $CLO_IMAGE | oc create -f -
```