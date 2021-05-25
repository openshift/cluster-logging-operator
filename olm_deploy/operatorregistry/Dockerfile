FROM quay.io/operator-framework/upstream-registry-builder AS registry-builder

FROM registry.ci.openshift.org/ocp/4.7:base

WORKDIR /
COPY manifests/ /manifests

RUN chmod -R g+w /manifests && \
  rm /manifests/*/image-references /manifests/*/kustomization.yaml /manifests/*/crd-v1-singleton-patch.yaml /manifests/*/crd-v1-clusterloggings-patches.yaml

COPY olm_deploy/scripts/registry-init.sh /scripts/

COPY --from=registry-builder /bin/initializer /usr/bin/initializer
COPY --from=registry-builder /bin/registry-server /usr/bin/registry-server
COPY --from=registry-builder /bin/grpc_health_probe /usr/bin/grpc_health_probe

# Change working directory to enable registry migrations
# See https://bugzilla.redhat.com/show_bug.cgi?id=1843702
# See https://bugzilla.redhat.com/show_bug.cgi?id=1827612
WORKDIR /bundle
