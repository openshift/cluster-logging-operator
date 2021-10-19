### This is a generated file from Dockerfile.in ###
#@follow_tag(registry-proxy.engineering.redhat.com/rh-osbs/openshift-golang-builder:rhel_8_golang_1.16)
FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.8 AS builder

ENV BUILD_VERSION=${CI_CONTAINER_VERSION}
ENV OS_GIT_MAJOR=${CI_X_VERSION}
ENV OS_GIT_MINOR=${CI_Y_VERSION}
ENV OS_GIT_PATCH=${CI_Z_VERSION}
ENV SOURCE_GIT_COMMIT=${CI_CLUSTER_LOGGING_OPERATOR_UPSTREAM_COMMIT}
ENV SOURCE_GIT_URL=${CI_CLUSTER_LOGGING_OPERATOR_UPSTREAM_URL}
ENV REMOTE_SOURCE=${REMOTE_SOURCE:-.}


WORKDIR /go/src/github.com/openshift/cluster-logging-operator

COPY ${REMOTE_SOURCE}/main.go .
COPY ${REMOTE_SOURCE}/apis ./apis
COPY ${REMOTE_SOURCE}/controllers ./controllers
COPY ${REMOTE_SOURCE}/internal ./internal
COPY ${REMOTE_SOURCE}/must-gather ./must-gather
COPY ${REMOTE_SOURCE}/version ./version
COPY ${REMOTE_SOURCE}/scripts ./scripts
COPY ${REMOTE_SOURCE}/files ./files
COPY ${REMOTE_SOURCE}/go.mod .
COPY ${REMOTE_SOURCE}/go.sum .
COPY ${REMOTE_SOURCE}/vendor ./vendor
COPY ${REMOTE_SOURCE}/manifests ./manifests
COPY ${REMOTE_SOURCE}/.bingo .bingo
COPY ${REMOTE_SOURCE}/Makefile ./Makefile

RUN make build


#@follow_tag(registry-proxy.engineering.redhat.com/rh-osbs/openshift-ose-cli:v4.8)
FROM registry.ci.openshift.org/ocp/4.8:cli AS origincli

#@follow_tag(registry.redhat.io/ubi8:latest)
FROM registry.ci.openshift.org/ocp/4.8:base
RUN INSTALL_PKGS=" \
      openssl \
      rsync \
      file \
      xz \
      " && \
    yum install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all && \
    mkdir /tmp/ocp-clo && \
    chmod og+w /tmp/ocp-clo
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/bin/cluster-logging-operator /usr/bin/
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/scripts/* /usr/bin/scripts/
RUN mkdir -p /usr/share/logging/
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/files/ /usr/share/logging/

COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/manifests /manifests

COPY --from=origincli /usr/bin/oc /usr/bin
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/must-gather/collection-scripts/* /usr/bin/

# this is required because the operator invokes a script as `bash scripts/cert_generation.sh`
WORKDIR /usr/bin
#ENV ENABLE_VECTOR_COLLECTOR=true
CMD ["/usr/bin/cluster-logging-operator"]

LABEL \
        io.k8s.display-name="Cluster Logging Operator" \
        io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
        io.openshift.tags="openshift,logging" \
        com.redhat.delivery.appregistry="false" \
        maintainer="AOS Logging <aos-logging@redhat.com>" \
        License="Apache-2.0" \
        name="openshift-logging/cluster-logging-rhel8-operator" \
        com.redhat.component="cluster-logging-operator-container" \
        io.openshift.maintainer.product="OpenShift Container Platform" \
        io.openshift.build.commit.id=${CI_CLUSTER_LOGGING_OPERATOR_UPSTREAM_COMMIT} \
        io.openshift.build.source-location=${CI_CLUSTER_LOGGING_OPERATOR_UPSTREAM_URL} \
        io.openshift.build.commit.url=${CI_CLUSTER_LOGGING_OPERATOR_UPSTREAM_URL}/commit/${CI_CLUSTER_LOGGING_OPERATOR_UPSTREAM_COMMIT} \
        version=${CI_CONTAINER_VERSION}


