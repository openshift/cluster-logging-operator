FROM quay.io/operator-framework/upstream-registry-builder AS registry-builder

FROM registry.redhat.io/ubi8:8.5

WORKDIR /
COPY bundle/manifests/ /manifests
COPY bundle/metadata /metadata/

#
# TODO: Remove this once we migrate olm-deploy to use `opm index add`
#       to deploy using the operator bundle image instead. This is
#       currently a temporary solution as per /bin/initializer
#       does not support loading package and channel information from
#       the metadata/annotations.yaml
#
COPY olm_deploy/operatorregistry/cluster-logging.package.yaml /manifests/

RUN chmod -R g+w /manifests /metadata

COPY olm_deploy/scripts/registry-init.sh olm_deploy/scripts/env.sh /scripts/

COPY --from=registry-builder /bin/initializer /usr/bin/initializer
COPY --from=registry-builder /bin/opm /usr/bin/opm
#
# TODO: Remove this after merging switch to opm to get registry builds into CI first.
#
COPY --from=registry-builder /bin/registry-server /usr/bin/registry-server
COPY --from=registry-builder /bin/grpc_health_probe /usr/bin/grpc_health_probe

# Change working directory to enable registry migrations
# See https://bugzilla.redhat.com/show_bug.cgi?id=1843702
# See https://bugzilla.redhat.com/show_bug.cgi?id=1827612
WORKDIR /bundle
