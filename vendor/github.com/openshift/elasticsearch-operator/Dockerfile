# first stage
FROM openshift/origin-release:golang-1.10 AS builder

WORKDIR /tmp/elasticsearch-operator
COPY vendor/ /tmp/elasticsearch-operator/vendor/
COPY manifests/ manifests/
COPY cmd/ /tmp/elasticsearch-operator/cmd/
COPY pkg/ /tmp/elasticsearch-operator/pkg/
COPY Makefile /tmp/elasticsearch-operator/

RUN make

# second stage

FROM openshift/origin-base

ENV PATH="/usr/local/bin:${PATH}"
ENV ALERTS_FILE_PATH="/etc/elasticsearch-operator/files/prometheus_alerts.yml"
ENV RULES_FILE_PATH="/etc/elasticsearch-operator/files/Prometheus_rules.yml"

COPY --from=builder /tmp/elasticsearch-operator/_output/bin/elasticsearch-operator /usr/local/bin/
COPY files/ /etc/elasticsearch-operator/files/

WORKDIR /usr/local/bin
ENTRYPOINT ["elasticsearch-operator"]

LABEL io.k8s.display-name="OpenShift elasticsearch-operator" \
      io.k8s.description="This is the component that manages an Elasticsearch cluster on a kubernetes based platform" \
      io.openshift.tags="openshift,logging,elasticsearch" \
      io.openshift.release.operator=true \
      maintainer="AOS Logging <aos-logging@redhat.com>"
