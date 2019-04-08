FROM registry.svc.ci.openshift.org/ocp/builder:golang-1.10 AS builder
WORKDIR /go/src/github.com/openshift/elasticsearch-operator
COPY . .
RUN make

FROM registry.svc.ci.openshift.org/ocp/4.0:base
ARG CSV=4.1

ENV ALERTS_FILE_PATH="/etc/elasticsearch-operator/files/prometheus_alerts.yml"
ENV RULES_FILE_PATH="/etc/elasticsearch-operator/files/prometheus_rules.yml"

COPY --from=builder /go/src/github.com/openshift/elasticsearch-operator/_output/bin/elasticsearch-operator /usr/bin/
COPY --from=builder /go/src/github.com/openshift/elasticsearch-operator/files/ /etc/elasticsearch-operator/files/
COPY --from=builder /go/src/github.com/openshift/elasticsearch-operator/manifests/$CSV /manifests/$CSV
COPY --from=builder /go/src/github.com/openshift/elasticsearch-operator/manifests/elasticsearch-operator.package.yaml /manifests/

WORKDIR /usr/bin
ENTRYPOINT ["elasticsearch-operator"]

LABEL io.k8s.display-name="OpenShift elasticsearch-operator" \
      io.k8s.description="This is the component that manages an Elasticsearch cluster on a kubernetes based platform" \
      io.openshift.tags="openshift,logging,elasticsearch" \
      com.redhat.delivery.appregistry=true \
      maintainer="AOS Logging <aos-logging@redhat.com>"
