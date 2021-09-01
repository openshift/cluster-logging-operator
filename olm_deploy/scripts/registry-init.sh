#!/bin/bash

set -eou pipefail

echo -e "Dumping IMAGE env vars\n"
env | grep IMAGE
echo -e "\n\n"

IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-quay.io/openshift/origin-cluster-logging-operator:latest}
IMAGE_OAUTH_PROXY=${IMAGE_OAUTH_PROXY:-quay.io/openshift/origin-oauth-proxy:latest}
IMAGE_LOGGING_CURATOR5=${IMAGE_LOGGING_CURATOR5:-quay.io/openshift-logging/curator5:5.8.1}
IMAGE_LOGGING_FLUENTD=${IMAGE_LOGGING_FLUENTD:-quay.io/openshift-logging/fluentd:1.7.4}

# update the manifest with the image built by ci
sed -i "s,quay.io/openshift/origin-cluster-logging-operator:latest,${IMAGE_CLUSTER_LOGGING_OPERATOR}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-oauth-proxy:latest,${IMAGE_OAUTH_PROXY}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/curator5:5.8.1,${IMAGE_LOGGING_CURATOR5}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/fluentd:1.7.4,${IMAGE_LOGGING_FLUENTD}," /manifests/*/*clusterserviceversion.yaml

# update the manifest to pull always the operator image for non-CI environments
if [ "${OPENSHIFT_CI:-false}" == "false" ] ; then
    echo -e "Set operator deployment's imagePullPolicy to 'Always'\n\n"
    sed -i 's,imagePullPolicy:\ IfNotPresent,imagePullPolicy:\ Always,' /manifests/*/*clusterserviceversion.yaml
fi

echo -e "substitution complete, dumping new csv\n\n"
cat /manifests/*/*clusterserviceversion.yaml

echo "generating sqlite database"

/usr/bin/initializer --manifests=/manifests --output=/bundle/bundles.db --permissive=true
