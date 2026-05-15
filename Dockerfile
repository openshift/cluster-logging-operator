FROM golang:1.25 AS builder

ARG CACHE_DEPS="true"
WORKDIR /opt/app-root/src


COPY ./api ./api
COPY ./go.mod ./go.sum ./
RUN if [ "${CACHE_DEPS}" = "true" ] ; then go mod download ; fi
COPY ./.bingo .bingo
COPY ./Makefile ./Makefile
COPY ./version ./version
COPY ./cmd/main.go ./cmd/main.go
COPY ./internal ./internal

USER 0
RUN make build

FROM quay.io/openshift/origin-cli-artifacts:latest AS origincli

RUN case $(uname -m) in \
    x86_64) cp /usr/share/openshift/linux_amd64/oc.rhel9 /tmp/oc ;; \
    aarch64) cp /usr/share/openshift/linux_arm64/oc.rhel9 /tmp/oc ;; \
    ppc64le) cp /usr/share/openshift/linux_ppc64le/oc.rhel9 /tmp/oc ;; \
    s390x) cp /usr/share/openshift/linux_s390x/oc /tmp/oc ;; \
    *) echo "Unsupported architecture"; exit 1 ;; \
esac

FROM registry.access.redhat.com/ubi9/ubi-minimal

RUN INSTALL_PKGS=" \
      openssl \
      rsync \
      file \
      xz \
      " && \
    microdnf install -y ${INSTALL_PKGS} && \
    rpm -V ${INSTALL_PKGS} && \
    microdnf clean all && \
    mkdir /tmp/ocp-clo && \
    chmod og+w /tmp/ocp-clo

COPY --from=builder /opt/app-root/src/bin/cluster-logging-operator /usr/bin/

COPY --from=origincli /tmp/oc /usr/bin/oc

COPY ./must-gather/collection-scripts/* /usr/bin/

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


