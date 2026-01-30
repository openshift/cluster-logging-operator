FROM golang:1.24 AS builder

ENV APP_DIR=.
WORKDIR /opt/apt-root/src
COPY ${APP_DIR}/cmd ./cmd
COPY ${APP_DIR}/specs ./specs
COPY ${APP_DIR}/testdata ./testdata
COPY ${APP_DIR}/util ./util
COPY ${APP_DIR}/Makefile ./Makefile
COPY ${APP_DIR}/README ./README
COPY ${APP_DIR}/go.mod ./go.mod

USER 0
RUN go env -w GOPRIVATE=github.com/openshift/*
RUN go env -w GONOSUMDB=github.com/openshift/*
RUN make build

FROM quay.io/openshift/origin-cli-artifacts:4.16 AS origincli

RUN case $(uname -m) in \
    x86_64) cp /usr/share/openshift/linux_amd64/oc.rhel9 /tmp/oc ;; \
    aarch64) cp /usr/share/openshift/linux_arm64/oc.rhel9 /tmp/oc ;; \
    ppc64le) cp /usr/share/openshift/linux_ppc64le/oc.rhel9 /tmp/oc ;; \
    s390x) cp /usr/share/openshift/linux_s390x/oc /tmp/oc ;; \
    *) echo "Unsupported architecture"; exit 1 ;; \
esac

FROM registry.access.redhat.com/ubi9/ubi-minimal

ENV APP_DIR=/opt/apt-root/src
ENV SRC_DIR=./

RUN INSTALL_PKGS=" \
      openssl \
      rsync \
      file \
      xz \
      " && \
    microdnf install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    microdnf clean all

COPY --from=builder $APP_DIR/bin/test-e2e-extension /usr/bin/

COPY --from=origincli /tmp/oc /usr/bin/oc

USER 1000
WORKDIR /usr/bin
CMD ["/usr/bin/test-e2e-extension]

LABEL \
        io.k8s.display-name="test-e2e-extension" \
        io.k8s.description="This is an Openshift Logging system test image which container the binary used for Logging Test." \
        io.openshift.tags="openshift,logging" \
        com.redhat.delivery.appregistry="false" \
        maintainer="AOS Logging <team-logging@redhat.com>" \
        License="Apache-2.0" \
        name="test-e2e-extension" \
        com.redhat.component="openshift-logging-e2e-container" \
        io.openshift.maintainer.product="OpenShift Container Platform" \


