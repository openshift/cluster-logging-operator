FROM golang:1.26 AS builder

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
COPY ./must-gather ./must-gather

USER 0
RUN make build

FROM registry.access.redhat.com/ubi9/ubi:latest AS packages

# Install ubi-micro runtime packages into a staging root directory using DNF
RUN mkdir -p /mnt/rootfs && \
    dnf install -y --installroot /mnt/rootfs \
      --releasever 9 \
      --setopt=install_weak_deps=false \
      --nodocs \
      openssl-libs \
      ca-certificates \
      rsync \
      file \
      xz && \
    dnf --installroot /mnt/rootfs clean all && \
    mkdir -p /mnt/rootfs/tmp/ocp-clo && \
    chmod og+w /mnt/rootfs/tmp/ocp-clo

FROM registry.access.redhat.com/ubi9/ubi-micro

COPY --from=packages /mnt/rootfs/ /
COPY --from=builder /opt/app-root/src/bin/cluster-logging-operator /usr/bin/
COPY --from=builder /opt/app-root/src/bin/must-gather /usr/bin/

RUN ln -s /usr/bin/must-gather /usr/bin/gather

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
