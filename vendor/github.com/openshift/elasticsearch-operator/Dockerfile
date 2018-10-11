# first stage
FROM openshift/origin-release:golang-1.10 AS builder

WORKDIR /tmp/workdir
COPY vendor/ ./vendor
COPY cmd/ ./cmd
COPY manifests/ manifests/
COPY pkg/ ./pkg
COPY Makefile .

RUN make

# second stage
FROM openshift/origin-base
ENV PATH="/usr/local/bin:${PATH}"

COPY --from=builder /tmp/workdir/_output/bin/elasticsearch-operator /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/elasticsearch-operator"]

LABEL io.k8s.display-name="elasticsearch-operator" \
      io.k8s.description="This is the component that manages an Elasticsearch cluster on a kubernetes based platform" \
      io.openshift.tags="openshift,logging,elasticsearch" \
      io.openshift.release.operator=true \
      maintainer="AOS Logging<aos-logging@redhat.com>"
