# cluster-logging-operator
Operator to support OKD cluster logging

## Testing

### E2E Testing
To run the e2e tests, from the repo directory, run:
```
oc adm policy add-scc-to-user privileged -z fluentd -n openshift-logging
imagebuilder -t quay.io/openshift/cluster-logging-operator .
operator-sdk test local --namespace openshift-logging ./test/e2e/
```

### Dev Testing
To set up your local environment based on what will be provided by OLM, run:
```
mkdir /tmp/_working_dir

CLUSTER_LOGGING_OPERATOR=$GOPATH/src/github.com/openshift/cluster-logging-operator
ELASTICSEARCH_OPERATOR=$GOPATH/src/github.com/openshift/elasticsearch-operator

oc adm ca create-signer-cert --cert='/tmp/ca.crt' --key='/tmp/ca.key' --serial='/tmp/ca.serial.txt'
oc adm ca create-server-cert --cert='/tmp/kibana-internal.crt' --key='/tmp/kibana-internal.key' --hostnames='kibana,kibana-ops' --signer-cert='/tmp/ca.crt' --signer-key='/tmp/ca.key' --signer-serial='/tmp/ca.serial.txt'

oc create -n openshift-logging secret generic logging-master-ca --from-file=masterca=/tmp/ca.crt --from-file=masterkey=/tmp/ca.key --from-file=kibanacert=/tmp/kibana-internal.crt --from-file=kibanakey=/tmp/kibana-internal.key

oc label node --all logging-infra-fluentd=true
oc adm policy add-scc-to-user privileged -z fluentd -n openshift-logging

oc create -n openshift-logging -f $CLUSTER_LOGGING_OPERATOR/deploy/service_account.yaml
oc create -n openshift-logging -f $CLUSTER_LOGGING_OPERATOR/deploy/role.yaml
oc create -n openshift-logging -f $CLUSTER_LOGGING_OPERATOR/deploy/role_binding.yaml
oc create -n openshift-logging -f $CLUSTER_LOGGING_OPERATOR/deploy/crds/crd.yaml
oc create -n openshift-logging -f $ELASTICSEARCH_OPERATOR/deploy/rbac.yaml
oc create -n openshift-logging -f $ELASTICSEARCH_OPERATOR/deploy/crd.yaml

oc create -n openshift-logging -f $CLUSTER_LOGGING_OPERATOR/deploy/cr.yaml
```

To test on an OCP cluster, you can run:
`REPO_PREFIX=openshift/ IMAGE_PREFIX=origin- OPERATOR_NAME=cluster-logging-operator WATCH_NAMESPACE=openshift-logging KUBERNETES_CONFIG=/etc/origin/master/admin.kubeconfig go run cmd/cluster-logging-operator/main.go`


To remove created API objects:
`oc delete ClusterLogging example -n openshift-logging`
