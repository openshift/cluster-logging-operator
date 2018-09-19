FROM openshift/origin-base

ENV APP_HOME=/opt/app/src \
    GOPATH=/go \
    PATH="${PATH}:${GOPATH}/bin:/usr/local/bin" \
    INSTALL_PKGS="golang make git"

RUN yum install -y ${INSTALL_PKGS} --setopt=tsflags=nodocs

WORKDIR /tmp/workdir
COPY vendor/ ./vendor
COPY cmd/ ./cmd
COPY pkg/ ./pkg
COPY Makefile .

RUN make && \
    cp _output/bin/elasticsearch-operator /usr/local/bin/ && \
    rm -rf /tmp/workdir && \
    yum erase -y ${INSTALL_PKGS} && \
    yum clean all

LABEL io.k8s.display-name="elasticsearch-operator" \
      io.k8s.description="This is the component that manages an Elasticsearch cluster on a kubernetes based platform" \
      io.openshift.tags="openshift,logging,elasticsearch" \
      maintainer="AOS Logging<aos-logging@redhat.com>"

ENTRYPOINT ["/usr/local/bin/elasticsearch-operator"]
