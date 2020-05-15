# Hacking on the Cluster Logging Operator

## Preparation

Clone this repository into `$GOPATH/github.com/openshift/cluster-logging-operator`

Install `podman`. For example on  Fedora:
```
sudo dnf install -y podman
```

Check out some other required repositories in your $GOPATH:
```
go get -d -u github.com/openshift/elasticsearch-operator/... github.com/openshift/origin-aggregated-logging/...
```

**Note:** `podman` copies our source tree very slowly. You can use `docker`
instead, this will not exactly reproduce how images are built for release or by
CI, but is close enough for many purposes. Put a symlink early in your path:
```
ln -s $(which docker) /somewhere/early/in/my/PATH/podman`
```

## Testing Overview

There are two types of test:

* *Unit tests:* run in seconds, no cluster needed, verify internal behavior - behind the API. New code *should* be reasonably well covered by unit tests but we don't have a formal requirement.
* *E2E tests:* run in minutes to hour(s), requires a cluster. There are several ways to run these:
   - *CI tests:* run automatically when you raise a PR in a controlled cluster environment. Can take hour(s) to get results.
   - *Deployed tests:* Build and deploy CLO image to your own cluster, allows you to run tests selectively.
   - *Local process:* requires a cluster with dependencies deployed (elasticsearch), but runs CLO as a *local process*. Faster and easier to debug.

All PRs should pass unit tests. CI runs all tests automatically and won't merge a PR till they pass, so it is OK to create a PR and let CI run the e2e tests if you've done enough testing to be confident. This allows for human review and feedback while CI tests are running.

To get a code review of unfinished work, create a PR with "[WIP]" at the start of the description. The CI system will not merge it until the "[WIP]" is removed and all tests are passing.

## Setting up a test cluster

You can use a real cluster (for example an AWS cluster) or create a virtual cluster on your development box with [Code Ready Containers](https://developers.redhat.com/products/codeready-containers/download)

If you use CRC, you may need to start with more memory than the default, e.g.:
```
crc start -m 12288
```

Log in as a user with the role `cluster-admin` - user `kubeadmin` is often predefined with this role.

You can `export KUBECONFIG=/path/to/cluster/config` or copy/add your config to `$HOME/.kube/config`. Depending on how you set it up, you may also need to `oc login` to your cluster and add pull secrets to your credentials.

## Building and running tests

Unit tests can be run without a cluster:
* `make check`: Generate and format code, run unit tests and lint.
   You should run *at least* this set of checks before a commit.

To build, deploy and test the CLO image in your own cluster:
* `make deploy`: Build CLO image and deploy to cluster with elasticsearch.
* `make undeploy`: Undo `make deploy`.

To run CLO as a local process:
* `make run`: Run CLO as a local process. Does not require `make deploy`.
* `make debug`: Run CLO in the `dlv` debugger.

After running `make deploy`, `make run` or `make debug`, you can run e2e tests:
* `make test-e2e-run` Run all e2e tests against an already-running CLO.
* `go run ./test/e2e/something`: Run tests individually, in a debugger or whatever.

To do a complete deploy, test, cleanup cycle as CI does (starting from clean cluster)
* `make test-e2e-olm`: Run e2e tests the same way CI does, requires some setup - see below.
* `make test-e2e-local`: Set up and run an e2e test using locally-built image.

Some other useful targets (read the Makefile for more)
* `make image`: Build the image as $IMAGE_TAG, do not deploy it.
* `make` or `make all`: Run `make check` and build the CLO image.

## More about `make run`

You can run the CLO as a local process, outside the cluster. This is *not* the
normal way to run an operator, and does not test all aspects of the CLO
(e.g. problems with the CSV or OLM interactions), but it has advantages:

* Fast edit/run/test cycle - runs from source code, skips some slow build/deploy steps.
* Directly access the CLO logs on stdout/stderr
* Control logging levels and other environment variables, e.g. `export LOG_LEVEL=debug`
* Run CLO in a debugger, profiler or other development tools.

*How it works*: An operator is actually a cluster *client*. It watches for
changes to its own custom resources, and creates/updates other resources
accordingly. This can all be done from outside the cluster.

Examples:
```
make run  # Run the CLO locally
make run-debug  # Run CLO under the dlv debugger
LOG_LEVEL=debug make run  # Run CLO with debug logging
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
``**

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

