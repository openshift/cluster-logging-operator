TARGET_DIR=$(CURDIR)/_output
KUBECONFIG?=$(HOME)/.kube/config

GOBUILD=go build
BUILD_GOPATH=$(TARGET_DIR):$(TARGET_DIR)/vendor:$(CURDIR)/cmd

IMAGE_BUILDER_OPTS=
IMAGE_BUILDER?=imagebuilder
IMAGE_BUILD=$(IMAGE_BUILDER)
export IMAGE_TAG_CMD?=docker tag

export APP_NAME=cluster-logging-operator
APP_REPO=github.com/openshift/$(APP_NAME)
TARGET=$(TARGET_DIR)/bin/$(APP_NAME)
IMAGE_TAG?=quay.io/openshift/origin-$(APP_NAME):latest
export IMAGE_TAG
MAIN_PKG=cmd/manager/main.go
export OCP_VERSION?=$(shell basename $(shell find manifests/  -maxdepth 1  -not -name manifests -type d))
export CSV_FILE=$(CURDIR)/manifests/$(OCP_VERSION)/cluster-logging.v$(OCP_VERSION).0.clusterserviceversion.yaml
export NAMESPACE?=openshift-logging
export EO_CSV_FILE=$(CURDIR)/vendor/github.com/openshift/elasticsearch-operator/manifests/$(OCP_VERSION)/elasticsearch-operator.v$(OCP_VERSION).0.clusterserviceversion.yaml

FLUENTD_IMAGE?=quay.io/openshift/origin-logging-fluentd:latest

PKGS=$(shell go list ./... | grep -v -E '/vendor/|/test|/examples')
TEST_PKGS=$(shell go list ./test)

TEST_OPTIONS?=

# go source files, excluding generated code.
SRC = $(shell find cmd pkg version -type f -name '*.go' -not -name zz_generated*)

.PHONY: all imagebuilder build clean fmt simplify generate deploy-setup deploy-image deploy deploy-example test-unit test-e2e test-sec undeploy run

all: build #check install

# Download a known released version of operator-sdk.
OPERATOR_SDK_RELEASE?=v0.15.2
OPERATOR_SDK=./operator-sdk-$(OPERATOR_SDK_RELEASE)
$(OPERATOR_SDK):
	curl -f -L -o $@ https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_RELEASE}/operator-sdk-${OPERATOR_SDK_RELEASE}-$(shell uname -i)-linux-gnu
	chmod +x $(OPERATOR_SDK)

imagebuilder:
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && ! type -p imagebuilder ; \
	then go get -u github.com/openshift/imagebuilder/cmd/imagebuilder ; \
	fi

build: generate fmt
	@mkdir -p $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURDIR)/pkg $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURDIR)/vendor/* $(TARGET_DIR)/src
	GOPATH=$(BUILD_GOPATH) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_PKG)

run:
	ELASTICSEARCH_IMAGE=quay.io/openshift/origin-logging-elasticsearch6:latest \
	FLUENTD_IMAGE=$(FLUENTD_IMAGE) \
	KIBANA_IMAGE=quay.io/openshift/origin-logging-kibana6:latest \
	CURATOR_IMAGE=quay.io/openshift/origin-logging-curator6:latest \
	OAUTH_PROXY_IMAGE=quay.io/openshift/origin-oauth-proxy:latest \
	PROMTAIL_IMAGE=quay.io/openshift/origin-promtail:latest \
	OPERATOR_NAME=cluster-logging-operator \
	WATCH_NAMESPACE=openshift-logging \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(TARGET_DIR)/ocp-clo \
	LOGGING_SHARE_DIR=$(CURDIR)/files \
	go run ${MAIN_PKG}

clean:
	@rm -rf $(TARGET_DIR) && \
	go clean -cache -testcache  $(TEST_PKGS) $(PKGS)

image: imagebuilder
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && [ $${SKIP_BUILD:-false} = false ] ; \
	then hack/build-image.sh $(IMAGE_TAG) $(IMAGE_BUILDER) $(IMAGE_BUILDER_OPTS) ; \
	fi

lint:
	golangci-lint run -c golangci.yaml

fmt:
	gofmt -l -w cmd/ pkg/ version/

simplify:
	gofmt -s -l -w $(SRC)

GEN_TIMESTAMP=.zz_generate_timestamp
generate: $(GEN_TIMESTAMP)
$(GEN_TIMESTAMP): $(SRC) $(OPERATOR_SDK)
	$(OPERATOR_SDK) generate k8s
	$(OPERATOR_SDK) generate crds
	@touch $@

# spotless does make clean and removes generated code. Don't commit without re-generating.
spotless: clean
	@find pkg -name 'zz_generated*' -delete -print
	@rm -vrf deploy/crds/*.yaml
	@rm -vf $(GEN_TIMESTAMP)

deploy-image: image
	hack/deploy-image.sh

deploy:  deploy-image deploy-elasticsearch-operator
	hack/deploy.sh

deploy-no-build:  deploy-elasticsearch-operator
	hack/deploy.sh

deploy-elasticsearch-operator:
	hack/deploy-eo.sh

deploy-example: deploy
	oc create -n $(NAMESPACE) -f hack/cr.yaml

test-unit: fmt
	@LOGGING_SHARE_DIR=$(CURDIR)/files go test $(TEST_OPTIONS) $(PKGS)

test-e2e:
	hack/test-e2e.sh

test-e2e-local: deploy-image
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
	 hack/test-e2e.sh

test-sec:
	go get -u github.com/securego/gosec/cmd/gosec
	gosec -severity medium --confidence medium -quiet ./...

undeploy:
	hack/undeploy.sh
