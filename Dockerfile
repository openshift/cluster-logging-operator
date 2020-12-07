FROM registry.svc.ci.openshift.org/ocp/builder:rhel-8-golang-1.15-openshift-4.7 AS builder
WORKDIR /go/src/github.com/openshift/cluster-logging-operator

# COPY steps are in the reverse order of frequency of change
COPY cmd ./cmd
COPY must-gather ./must-gather
COPY version ./version
COPY files ./files
COPY go.mod .
COPY go.sum .
COPY vendor ./vendor
COPY manifests ./manifests
COPY .bingo .bingo
COPY Makefile ./Makefile
COPY pkg ./pkg

RUN make build

FROM registry.svc.ci.openshift.org/ocp/4.7:cli as origincli

FROM registry.svc.ci.openshift.org/ocp/4.7:base
RUN INSTALL_PKGS=" \
      rsync \
      xz \
      " && \
    yum install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/bin/cluster-logging-operator /usr/bin/

RUN mkdir -p /usr/share/logging/
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/files/ /usr/share/logging/

COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/manifests /manifests
RUN rm /manifests/art.yaml
RUN rm -rf /manifests/5.0

COPY --from=origincli /usr/bin/oc /usr/bin/
COPY --from=builder /go/src/github.com/openshift/cluster-logging-operator/must-gather/collection-scripts/* /usr/bin/

WORKDIR /usr/bin
CMD ["/usr/bin/cluster-logging-operator"]
LABEL io.k8s.display-name="OpenShift cluster-logging-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
      io.openshift.tags="openshift,logging,cluster-logging" \
      com.redhat.delivery.appregistry=true \
      maintainer="AOS Logging <aos-logging@redhat.com>"
