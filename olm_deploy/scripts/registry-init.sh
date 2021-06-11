#!/bin/bash

set -eou pipefail

echo -e "Dumping IMAGE env vars\n"
env | grep IMAGE
echo -e "\n\n"


IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-quay.io/openshift-logging/cluster-logging-operator:5.1}
IMAGE_OAUTH_PROXY=${IMAGE_OAUTH_PROXY:-quay.io/openshift/origin-oauth-proxy:latest}
IMAGE_LOGGING_FLUENTD=${IMAGE_LOGGING_FLUENTD:-quay.io/openshift-logging/fluentd:5.1}
IMAGE_ELASTICSEARCH6=${IMAGE_ELASTICSEARCH6:-quay.io/openshift-logging/elasticsearch6:5.1}
IMAGE_LOGGING_KIBANA6=${IMAGE_LOGGING_KIBANA6:-quay.io/openshift-logging/kibana6:5.1}

# update the manifest with the image built by ci
sed -i "s,quay.io/openshift-logging/cluster-logging-operator:latest,${IMAGE_CLUSTER_LOGGING_OPERATOR}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-oauth-proxy:latest,${IMAGE_OAUTH_PROXY}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/fluentd:latest,${IMAGE_LOGGING_FLUENTD}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/elasticsearch6:latest,${IMAGE_ELASTICSEARCH6}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/kibana6:latest,${IMAGE_LOGGING_KIBANA6}," /manifests/*/*clusterserviceversion.yaml

# update the manifest to pull always the operator image for non-CI environments
if [ "${OPENSHIFT_CI:-false}" == "false" ] ; then
    echo -e "Set operator deployment's imagePullPolicy to 'Always'\n\n"
    sed -i 's,imagePullPolicy:\ IfNotPresent,imagePullPolicy:\ Always,' /manifests/*/*clusterserviceversion.yaml
fi

echo -e "substitution complete, dumping new csv\n\n"
cat /manifests/*/*clusterserviceversion.yaml

echo "generating sqlite database"

/usr/bin/initializer --manifests=/manifests --output=/bundle/bundles.db --permissive=true
