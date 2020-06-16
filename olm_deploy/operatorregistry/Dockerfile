FROM quay.io/operator-framework/upstream-registry-builder AS registry-builder

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.12

WORKDIR /
COPY manifests/ /manifests

RUN chmod -R g+w /manifests && rm /manifests/art.yaml /manifests/*/image-references

COPY olm_deploy/scripts/registry-init.sh /scripts/

COPY --from=registry-builder /bin/initializer /usr/bin/initializer
COPY --from=registry-builder /bin/registry-server /usr/bin/registry-server
COPY --from=registry-builder /bin/grpc_health_probe /usr/bin/grpc_health_probe

# Change working directory to enable registry migrations
# See https://bugzilla.redhat.com/show_bug.cgi?id=1843702
# See https://bugzilla.redhat.com/show_bug.cgi?id=1827612
WORKDIR /bundle
