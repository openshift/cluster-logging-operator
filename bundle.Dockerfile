FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=cluster-logging-operator
LABEL operators.operatorframework.io.bundle.channels.v1=4.7
LABEL operators.operatorframework.io.bundle.channel.default.v1=4.7

COPY bundle/manifests /manifests/
COPY bundle/metadata /metadata/

LABEL com.redhat.delivery.operator.bundle=true
LABEL com.redhat.openshift.versions="v4.6-v4.7"

LABEL \
    com.redhat.component="cluster-logging-operator" \
    version="v1.1" \
    name="cluster-logging-operator" \
    License="ASL 2.0" \
    io.k8s.display-name="cluster-logging-operator bundle" \
    io.k8s.description="bundle for the cluster-logging-operator" \
    summary="This is the bundle for the cluster-logging-operator" \
    maintainer="AOS Logging <aos-logging@redhat.com>"
