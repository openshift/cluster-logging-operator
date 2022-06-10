#!/usr/bin/bash

source .bingo/variables.env

set -euo pipefail

sed -i 's/bundle\.package\.v1=.*logging/bundle\.package\.v1=cluster-logging/g' bundle.Dockerfile
sed -i 's/.*scorecard.*//g' bundle.Dockerfile

sed -i 's/bundle\.package\.v1\: .*logging/bundle\.package\.v1\: cluster-logging/g' ./bundle/metadata/annotations.yaml
sed -i 's/.*scorecard.*//g' ./bundle/metadata/annotations.yaml

cat >> bundle.Dockerfile <<EOF

LABEL com.redhat.delivery.operator.bundle=true
LABEL com.redhat.openshift.versions="${OPENSHIFT_VERSIONS}"

LABEL \\
    com.redhat.component="cluster-logging-operator" \\
    version="v1.1" \\
    name="cluster-logging-operator" \\
    License="Apache-2.0" \\
    io.k8s.display-name="cluster-logging-operator bundle" \\
    io.k8s.description="bundle for the cluster-logging-operator" \\
    summary="This is the bundle for the cluster-logging-operator" \\
    maintainer="AOS Logging <team-logging@redhat.com>"
EOF

echo "validating bundle..."
$OPERATOR_SDK bundle validate --verbose ./bundle
