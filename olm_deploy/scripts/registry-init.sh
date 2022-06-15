#!/bin/bash

set -eou pipefail
source $(dirname "${BASH_SOURCE[0]}")/env.sh

echo -e "Dumping IMAGE env vars\n"
env | grep IMAGE
echo -e "\n\n"

# update the manifest with the image built by ci
sed -i "s,quay.io/openshift-logging/cluster-logging-operator:latest,${IMAGE_CLUSTER_LOGGING_OPERATOR}," /manifests/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/fluentd:.*,${IMAGE_LOGGING_FLUENTD}," /manifests/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/vector:.*,${IMAGE_LOGGING_VECTOR}," /manifests/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift-logging/log-file-metric-exporter:.*,${IMAGE_LOG_FILE_METRIC_EXPORTER}," /manifests/*clusterserviceversion.yaml

# update the manifest to pull always the operator image for non-CI environments
if [ "${OPENSHIFT_CI:-false}" == "false" ] ; then
    echo -e "Set operator deployment's imagePullPolicy to 'Always'\n\n"
    sed -i 's,imagePullPolicy:\ IfNotPresent,imagePullPolicy:\ Always,' /manifests/*clusterserviceversion.yaml
fi

echo -e "substitution complete, dumping new csv\n\n"
cat /manifests/*clusterserviceversion.yaml

echo "generating sqlite database"

/usr/bin/initializer --manifests=/manifests --output=/bundle/bundles.db --permissive=true
