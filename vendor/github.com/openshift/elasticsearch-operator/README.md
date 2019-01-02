# elasticsearch-operator

[![Build Status](https://travis-ci.org/openshift/elasticsearch-operator.svg?branch=master)](https://travis-ci.org/openshift/elasticsearch-operator)

*WORK IN PROGRESS*

Elasticsearch operator to run Elasticsearch cluster on top of Openshift and Kubernetes.
Operator uses [Operator Framework SDK](https://github.com/operator-framework/operator-sdk).

## Why Use An Operator?

Operator is designed to provide self-service for the Elasticsearch cluster operations. See the [diagram](https://github.com/operator-framework/operator-sdk/blob/master/doc/images/Operator-Maturity-Model.png) on operator maturity model.

- Elasticsearch operator ensures proper layout of the pods
- Elasticsearch operator enables proper rolling cluster restarts
- Elasticsearch operator provides kubectl interface to manage your Elasticsearch cluster
- Elasticsearch operator provides kubectl interface to monitor your Elasticsearch cluster

## Getting started

### Prerequisites

- Cluster administrator must set `vm.max_map_count` sysctl to 262144 on the host level of each node in your cluster prior to running the operator.
- In case hostmounted volume is used, the directory on the host must have 777 permissions and the following selinux labels (TODO).
- In case secure cluster is used, the certificates must be pre-generated and uploaded to the secret `<elasticsearch_cluster_name>-certs`

### Kubernetes

Make sure certificates are pre-generated and deployed as secret.
Upload the Custom Resource Definition to your Kubernetes cluster:

    $ kubectl create -f deploy/crd.yaml

Deploy the required roles to the cluster:

    $ kubectl create -f deploy/rbac.yaml

Deploy custom resource and the Deployment resource of the operator:

    $ kubectl create -f deploy/cr.yaml
    $ kubectl create -f deploy/operator.yaml

### OpenShift

As a cluster admin apply the template with the roles and permissions:

    $ oc process -f deploy/openshift/admin-elasticsearch-template.yaml | oc apply -f -

The template deploys CRD, roles and rolebindings. You can pass variables:

- `NAMESPACE` to specify which namespace's default ServiceAccount will be allowed to manage the Custom Resource.
- `ELASTICSEARCH_ADMIN_USER` to specify which user of OpenShift will be allowed to manage the Custom Resource.

For example:

    $ oc process NAMESPACE=myproject ELASTICSEARCH_ADMIN_USER=developer -f deploy/openshift/admin-elasticsearch-template.yaml | oc apply -f -

Grant permissions to extra users by giving them the role `elasticsearch-operator`.

As the user which was specified as `ELASTICSEARCH_ADMIN_USER` on previous step:

Make sure the secret with Elasticsearch certificates exists and is named `<elasticsearch_cluster_name>-certs`

Then process the following template:

    $ oc process -f deploy/openshift/elasticsearch-template.yaml | oc apply -f -

The template deploys the Custom Resource and the operator deployment. You can pass the following variables to the template:

- `NAMESPACE` - namespace where the Elasticsearch cluster will be deployed. Must be the same as the one specified by admin
- `ELASTICSEARCH_CLUSTER_NAME` - name of the Elasticsearch cluster to be deployed

For example:

    $ oc process NAMESPACE=myproject ELASTICSEARCH_CLUSTER_NAME=elastic1 -f deploy/openshift/elasticsearch-template.yaml | oc apply -f -

### Openshift Alternatives - Make targets
If you are using an image built in a different node, you can specify to use a remote registry by setting
the environment variable `REMOTE_REGISTRY=true` before running any of the targets below.  See `hack/deploy-image.sh`
and `hack/deploy.sh` for more details.

*  `REMOTE_REGISTRY` Set to `true` if you are running the cluster on a different machine
    than the one you are developing on. For example, if you are running a cluster in a
    local libvirt or minishift environment, you may want to build the image on the host
    and push them to the cluster running in the VM.
    You will need a username with a password (i.e. not the default `system:admin` user).
    If your cluster was deployed with the `allow_all` identity provider, you can create
    a user like this: `oc login --username=admin --password=admin`, then assign it rights:
    `oc login --username=system:admin`
    `oc adm policy add-cluster-role-to-user cluster-admin admin`
    If you used the new `openshift-installer`, it created a user named `kubeadmin`
    with the password in the file `installer/auth/kubeadmin_password`.
    `oc login --username=kubeadmin --password=$( cat ../installer/auth/kubeadmin_password )`
    The user should already have `cluster-admin` rights.

It is additionally possible to deploy the operator to an Openshift cluster using the provided make targets.  These
targets assume you have cluster admin access. Following are a few of these targets:

#### deploy
Deploy the resources for the operator, build the operator image, push the image to the Openshift registry

#### deploy-setup
Deploy the pre-requirements for the operator to function (i.e. CRD, RBAC, sample secret)

#### deploy-example
Deploy an example custom resource for a single node Elasticsearch cluster

#### deploy-undeploy
Remove all deployed resources

#### go-run
Deploy the example cluster and start running the operator.  The end result is that there will be an
`elasticsearch` custom resource, and an elasticsearch pod running.  You can view the operator log by
looking at the log file specified by `$(RUN_LOG)` (default `elasticsearch-operator.log`).  The command
is run in the background - when finished, kill the process by killing the pid, which is written to the
file `$(RUN_PID)` (default `elasticsearch-operator.pid`) e.g. `kill $(cat elasticsearch-operator.pid)`

## Customize your cluster

### Image customization

The operator is designed to work with `openshift/origin-logging-elasticsearch5` image.  To use
a different image, edit `manifests/image-references` before deployment, or edit the elasticsearch
cr after deployment e.g. `oc edit elasticsearch elasticsearch`.

### Storage configuration

Storage is configurable per individual node type. Possible configuration
options:

- Hostmounted directory
- Empty directory
- Existing PersistentVolume
- New PersistentVolume generated by StorageClass

### Elasticsearch cluster topology customization

Decide how many nodes you want to run.

### Elasticsearch node configuration customization

TODO

## Supported features

Kubernetes TBD+ and OpenShift TBD+ are supported.

- [x] SSL-secured deployment (using Searchguard)
- [x] Insecure deployment (requires different image)
- [x] Index per tenant
- [x] Logging to a file or to console
- [ ] Elasticsearch 6.x support
- [x] Elasticsearch 5.6.x support
- [x] Master role
- [x] Client role
- [x] Data role
- [x] Clientdata role
- [x] Clientdatamaster role
- [ ] Elasticsearch snapshots
- [x] Prometheus monitoring
- [ ] Status monitoring
- [ ] Rolling restarts

## Testing

In a real deployment OpenShift monitoring will be installed.  However
for testing purposes, you should install the monitoring CRDs:
```
[REMOTE_REGISTRY=true] make deploy-setup
```

Use `REMOTE_REGISTRY=true make deploy-image` to build the image and copy it
to the remote registry.

### E2E Testing
To run the e2e tests, install the above CRDs and from the repo directory, run:
```
make test-e2e
```

### Dev Testing
You should first ensure that you have commands such as `imagebuilder` and `operator-sdk`
available by using something like `https://github.com/openshift/origin-aggregated-logging/blob/master/hack/sdk_setup.sh`.

To set up your local environment based on what will be provided by OLM, run:
```
sudo sysctl -w vm.max_map_count=262144
ELASTICSEARCH_OPERATOR=$GOPATH/src/github.com/openshift/elasticsearch-operator
[REMOTE_REGISTRY=true] make deploy-setup
[REMOTE_REGISTRY=true] make deploy-example
```

To test on an OCP cluster, you can run:

    make go-run

To remove created API objects:
```
make undeploy
```
