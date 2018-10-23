# first stage
FROM openshift/origin-release:golang-1.10 AS builder

WORKDIR /tmp/cluster-logging-operator
COPY vendor/ /tmp/cluster-logging-operator/vendor/
COPY manifests/ /manifests/
COPY scripts/* /usr/local/bin/scripts/
COPY files/* /usr/local/bin/files/
COPY cmd/ /tmp/cluster-logging-operator/cmd/
COPY pkg/ /tmp/cluster-logging-operator/pkg/
COPY Makefile /tmp/cluster-logging-operator/

RUN make

# second stage

FROM openshift/origin-base

ENV PATH="/usr/local/bin:${PATH}"

COPY scripts/* /usr/local/bin/scripts/
COPY files/* /usr/local/bin/files/
COPY --from=builder /tmp/cluster-logging-operator/_output/bin/cluster-logging-operator /usr/local/bin/

# TODO: update the operator code to not rely on this tmp dir to be created and remove from here
RUN mkdir /tmp/_working_dir && \
    chmod og+w /tmp/_working_dir && \
    yum install -y openssl

# this is required because the operator invokes a script as `bash scripts/cert_generation.sh`
WORKDIR /usr/local/bin
ENTRYPOINT ["cluster-logging-operator"]

LABEL io.k8s.display-name="OpenShift cluster-logging-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
      io.openshift.tags="openshift,logging,cluster-logging" \
      io.openshift.release.operator=true \
      maintainer="AOS Logging <aos-logging@redhat.com>"
