#!/bin/bash

set -eou pipefail

echo -e "Dumping IMAGE env vars\n"
env | grep IMAGE
echo -e "\n\n"


IMAGE_CLUSTER_LOGGING_OPERATOR=${IMAGE_CLUSTER_LOGGING_OPERATOR:-quay.io/openshift/origin-cluster-logging-operator:latest}
IMAGE_OAUTH_PROXY=${IMAGE_OAUTH_PROXY:-quay.io/openshift/origin-oauth-proxy:latest}
IMAGE_LOGGING_CURATOR5=${IMAGE_LOGGING_CURATOR5:-quay.io/openshift/origin-logging-curator5:latest}
IMAGE_LOGGING_FLUENTD=${IMAGE_LOGGING_FLUENTD:-quay.io/openshift/origin-logging-fluentd:latest}
IMAGE_PROMTAIL=${IMAGE_PROMTAIL:-name: quay.io/openshift/origin-promtail:latest}
IMAGE_ELASTICSEARCH6=${IMAGE_ELASTICSEARCH6:-quay.io/openshift/origin-logging-elasticsearch6:latest}
IMAGE_LOGGING_KIBANA6=${IMAGE_OAUTH_KIBANA6:-quay.io/openshift/origin-logging-kibana6:latest}

# update the manifest with the image built by ci
sed -i "s,quay.io/openshift/origin-cluster-logging-operator:latest,${IMAGE_CLUSTER_LOGGING_OPERATOR}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-oauth-proxy:latest,${IMAGE_OAUTH_PROXY}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-logging-curator5:latest,${IMAGE_LOGGING_CURATOR5}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-logging-fluentd:latest,${IMAGE_LOGGING_FLUENTD}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-promtail:latest,${IMAGE_PROMTAIL}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-logging-elasticsearch6:latest,${IMAGE_ELASTICSEARCH6}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-logging-kibana6:latest,${IMAGE_LOGGING_KIBANA6}," /manifests/*/*clusterserviceversion.yaml


echo -e "substitution complete, dumping new csv\n\n"
cat /manifests/*/*clusterserviceversion.yaml

echo "generating sqlite database"

/usr/bin/initializer --manifests=/manifests --output=/bundle/bundles.db --permissive=true
