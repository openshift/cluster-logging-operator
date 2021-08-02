#!/bin/bash
set -e
make deploy-image

#export FLUENTD_IMAGE=quay.io/openshift-logging/fluentd:latest
export LOGGING_SHARE_DIR=$PWD/files
export KUBECONFIG=$HOME/.kube/config
export SCRIPTS_DIR=$PWD/scripts
export IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest
export IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=image-registry.openshift-image-registry.svc:5000/openshift/cluster-logging-operator-registry:latest
export LOG_LEVEL=1

go test -count 1 -v -run TestLokiOutput ./test/functional/outputs
go test -count 1 -v ./test/e2e/logforwarding/loki/.
