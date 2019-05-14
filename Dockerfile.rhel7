FROM registry.svc.ci.openshift.org/ocp/builder:golang-1.10 AS builder
WORKDIR /go/src/github.com/openshift/cluster-logging-operator
COPY . .
RUN make

FROM registry.svc.ci.openshift.org/ocp/4.0:base
ARG CSV=4.1
RUN INSTALL_PKGS=" \
      openssl \
      " && \
    yum install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all && \
    mkdir /tmp/_working_dir && \
    chmod og+w /tmp/_working_dir
COPY --from=builder _output/bin/cluster-logging-operator /usr/bin/
COPY scripts/* /usr/bin/scripts/
RUN mkdir -p /usr/share/logging/
COPY files/ /usr/share/logging/
COPY manifests/$CSV /manifests/$CSV
COPY manifests/cluster-logging.package.yaml /manifests/
# this is required because the operator invokes a script as `bash scripts/cert_generation.sh`
WORKDIR /usr/bin
ENTRYPOINT ["/usr/bin/cluster-logging-operator"]
LABEL io.k8s.display-name="OpenShift cluster-logging-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
      io.openshift.tags="openshift,logging,cluster-logging" \
      com.redhat.delivery.appregistry=true \
      maintainer="AOS Logging <aos-logging@redhat.com>"
