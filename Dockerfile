# first stage
FROM openshift/origin-release:golang-1.10 AS builder

WORKDIR /tmp/cluster-logging-operator
COPY vendor/ /tmp/cluster-logging-operator/vendor
COPY cmd/ /tmp/cluster-logging-operator/cmd
COPY pkg/ /tmp/cluster-logging-operator/pkg
COPY Makefile /tmp/cluster-logging-operator

RUN make

# second stage

FROM openshift/origin-base

ENV PATH="/usr/local/bin:${PATH}"

COPY scripts/* /usr/local/bin/scripts/
COPY --from=builder /tmp/cluster-logging-operator/_output/bin/cluster-logging-operator /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cluster-logging-operator"]

LABEL io.k8s.display-name="OpenShift cluster-logging-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform that manages the lifecycle of the Aggregated logging stack." \
      io.openshift.tags="openshift,logging" \
      maintainer="AOS Logging <aos-logging@redhat.com>"
