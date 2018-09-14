FROM openshift/origin-base

ENV APP_HOME=/opt/app/src \
    GOPATH=/go \
    PATH="${PATH}:${GOPATH}/bin:/usr/local/bin" \
    INSTALL_PKGS="golang make git"

RUN yum install -y ${INSTALL_PKGS} --setopt=tsflags=nodocs

WORKDIR /tmp/cluster-logging-operator
COPY vendor/ /tmp/cluster-logging-operator/vendor
COPY scripts/* /usr/local/bin/scripts/
COPY cmd/ /tmp/cluster-logging-operator/cmd
COPY pkg/ /tmp/cluster-logging-operator/pkg
COPY Makefile /tmp/cluster-logging-operator

RUN make && \
    cp _output/bin/cluster-logging-operator /usr/local/bin/ && \
    rm -rf /tmp/cluster-logging-operator && \
    yum erase -y ${INSTALL_PKGS} && \
    yum clean all

LABEL io.k8s.display-name="OpenShift cluster-logging-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
      io.openshift.tags="openshift,logging" \
      maintainer="AOS Logging <aos-logging@redhat.com>"

ENTRYPOINT ["/usr/local/bin/cluster-logging-operator"]
