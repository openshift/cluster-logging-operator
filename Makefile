CURPATH=$(PWD)
TARGET_DIR=$(CURPATH)/_output
KUBECONFIG?=$(HOME)/.kube/config

GOBUILD=go build
BUILD_GOPATH=$(TARGET_DIR):$(TARGET_DIR)/vendor:$(CURPATH)/cmd

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
export OCP_VERSION?=4.3
IMAGE_CLUSTER_LOGGING_OPERATOR?=registry.svc.ci.openshift.org/origin/$VERSION:cluster-logging-operator
export CSV_FILE=$(CURPATH)/manifests/$(OCP_VERSION)/cluster-logging.v$(OCP_VERSION).0.clusterserviceversion.yaml
export NAMESPACE?=openshift-logging
export EO_CSV_FILE=$(CURPATH)/vendor/github.com/openshift/elasticsearch-operator/manifests/$(OCP_VERSION)/elasticsearch-operator.v$(OCP_VERSION).0.clusterserviceversion.yaml

FLUENTD_IMAGE?=quay.io/openshift/origin-logging-fluentd:latest

PKGS=$(shell go list ./... | grep -v -E '/vendor/|/test|/examples')
TEST_PKGS=$(shell go list ./test)

TEST_OPTIONS?=

OC?=oc

# These will be provided to the target
#VERSION := 1.0.0
#BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
#LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

#.PHONY: all build clean install uninstall fmt simplify check run
.PHONY: all operator-sdk imagebuilder build clean fmt simplify gendeepcopy deploy-setup deploy-image deploy deploy-example test-unit test-e2e test-sec undeploy run

all: build #check install

operator-sdk:
	@if ! type -p operator-sdk ; \
	then if [ ! -d $(GOPATH)/src/github.com/operator-framework/operator-sdk ] ; \
	  then git clone https://github.com/operator-framework/operator-sdk --branch master $(GOPATH)/src/github.com/operator-framework/operator-sdk ; \
	  fi ; \
	  cd $(GOPATH)/src/github.com/operator-framework/operator-sdk ; \
	  make dep ; \
	  make install || sudo make install || cd commands/operator-sdk && sudo go install ; \
	fi

imagebuilder:
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && ! type -p imagebuilder ; \
	then go get -u github.com/openshift/imagebuilder/cmd/imagebuilder ; \
	fi

build: fmt
	@mkdir -p $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/pkg $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/vendor/* $(TARGET_DIR)/src
	@GOPATH=$(BUILD_GOPATH) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_PKG)

run:
	ELASTICSEARCH_IMAGE=quay.io/openshift/origin-logging-elasticsearch5:latest \
	FLUENTD_IMAGE=$(FLUENTD_IMAGE) \
	KIBANA_IMAGE=quay.io/openshift/origin-logging-kibana5:latest \
	CURATOR_IMAGE=quay.io/openshift/origin-logging-curator5:latest \
	OAUTH_PROXY_IMAGE=quay.io/openshift/origin-oauth-proxy:latest \
	OPERATOR_NAME=cluster-logging-operator \
	WATCH_NAMESPACE=openshift-logging \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(TARGET_DIR)/_working_dir \
	LOGGING_SHARE_DIR=$(CURPATH)/files \
	go run ${MAIN_PKG}

clean:
	@rm -rf $(TARGET_DIR) && \
	go clean -testcache $(TEST_PKGS) $(PKGS)

image: imagebuilder
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && [ $${SKIP_BUILD:-false} = false ] ; \
	then hack/build-image.sh $(IMAGE_TAG) $(IMAGE_BUILDER) $(IMAGE_BUILDER_OPTS) ; \
	fi

fmt:
	@gofmt -l -w cmd && \
	gofmt -l -w pkg && \
	gofmt -l -w test

simplify:
	@gofmt -s -l -w $(SRC)

gendeepcopy: operator-sdk
	@operator-sdk generate k8s

deploy-setup:
	hack/deploy-setup.sh

deploy-image: image
	hack/deploy-image.sh

deploy:  deploy-image
	IMAGE_CLUSTER_LOGGING_OPERATOR=$(IMAGE_CLUSTER_LOGGING_OPERATOR) hack/deploy.sh

deploy-no-build: deploy-setup
	NO_BUILD=true hack/deploy.sh

deploy-example: deploy
	oc create -n $(NAMESPACE) -f hack/cr.yaml

test-unit: fmt
	@LOGGING_SHARE_DIR=$(CURPATH)/files go test $(TEST_OPTIONS) $(PKGS)

test-e2e:
	hack/test-e2e.sh
	
test-sec:
	go get -u github.com/securego/gosec/cmd/gosec
	gosec -severity medium --confidence medium -quiet ./...

deploy-example-no-build: deploy-no-build
	oc create -n $(NAMESPACE) -f hack/cr.yaml

undeploy:
	hack/undeploy.sh
