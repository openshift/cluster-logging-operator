#
# Temporarily set this to internal golang-builder until UBI 9.5 is released with
# a go-toolset base image providing go1.22
#
# FROM registry.redhat.io/ubi9/go-toolset:latest AS builder
FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.22-openshift-4.17 as builder

ENV REMOTE_SOURCES=${REMOTE_SOURCES:-.}
ENV REMOTE_SOURCES_DIR=${REMOTE_SOURCES_DIR:-.}
ENV APP_DIR=.
ENV CACHE_DEPS="true"
WORKDIR /opt/apt-root/src


COPY ${APP_DIR}/api ./api
COPY ${APP_DIR}/go.mod ${APP_DIR}/go.sum ./
RUN if [ -n $CACHE_DEPS ]; then go mod download ; fi
COPY ${APP_DIR}/.bingo .bingo
COPY ${APP_DIR}/Makefile ./Makefile
COPY ${APP_DIR}/version ./version
COPY ${APP_DIR}/cmd/main.go ./cmd/main.go
COPY ${APP_DIR}/internal ./internal

USER 0
RUN make build

FROM quay.io/openshift/origin-cli-artifacts:4.16 AS origincli

RUN case $(uname -m) in \
    x86_64) cp /usr/share/openshift/linux_amd64/oc.rhel9 /tmp/oc ;; \
    aarch64) cp /usr/share/openshift/linux_arm64/oc.rhel9 /tmp/oc ;; \
    ppc64le) cp /usr/share/openshift/linux_ppc64le/oc.rhel9 /tmp/oc ;; \
    s390x) cp /usr/share/openshift/linux_s390x/oc /tmp/oc ;; \
    *) echo "Unsupported architecture"; exit 1 ;; \
esac

FROM registry.access.redhat.com/ubi9/ubi-minimal

ENV APP_DIR=/opt/apt-root/src
ENV SRC_DIR=./

RUN INSTALL_PKGS=" \
      openssl \
      rsync \
      file \
      xz \
      " && \
    microdnf install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    microdnf clean all && \
    mkdir /tmp/ocp-clo && \
    chmod og+w /tmp/ocp-clo

COPY --from=builder $APP_DIR/bin/cluster-logging-operator /usr/bin/

COPY --from=origincli /tmp/oc /usr/bin/oc

COPY $SRC_DIR/must-gather/collection-scripts/* /usr/bin/

USER 1000
WORKDIR /usr/bin
CMD ["/usr/bin/cluster-logging-operator"]

LABEL \
        io.k8s.display-name="Cluster Logging Operator" \
        io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
        io.openshift.tags="openshift,logging" \
        com.redhat.delivery.appregistry="false" \
        maintainer="AOS Logging <team-logging@redhat.com>" \
        License="Apache-2.0" \
        name="openshift-logging/cluster-logging-rhel9-operator" \
        com.redhat.component="cluster-logging-operator-container" \
        io.openshift.maintainer.product="OpenShift Container Platform" \


