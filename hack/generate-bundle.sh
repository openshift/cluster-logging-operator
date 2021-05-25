#!/usr/bin/bash

source .bingo/variables.env

set -euo pipefail

mkdir -p bundle

$OPM alpha bundle generate --directory manifests/${MANIFEST_VERSION} --package cluster-logging-operator --channels "tech-preview" --default "tech-preview" --output-dir bundle/

cat >> bundle.Dockerfile <<EOF

LABEL com.redhat.delivery.operator.bundle=true
LABEL com.redhat.openshift.versions="${OPENSHIFT_VERSIONS}"

LABEL \\
    com.redhat.component="cluster-logging-operator" \\
    version="v1.1" \\
    name="cluster-logging-operator" \\
    License="ASL 2.0" \\
    io.k8s.display-name="cluster-logging-operator bundle" \\
    io.k8s.description="bundle for the cluster-logging-operator" \\
    summary="This is the bundle for the cluster-logging-operator" \\
    maintainer="AOS Logging <aos-logging@redhat.com>"
EOF

find bundle/manifests/ -type f ! \( -name "cluster-logging*" -o -name "*crd.yaml" \) -delete

echo "validating bundle..."
$OPERATOR_SDK bundle validate --verbose bundle
