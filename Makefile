CURPATH=$(PWD)
TARGET_DIR=$(CURPATH)/_output

GOBUILD=go build
GOPATH=$(TARGET_DIR):$(TARGET_DIR)/vendor:$(CURPATH)/cmd

DOCKER_OPTS=
IMAGE_BUILDER?=docker build
IMAGE_BUILD=$(IMAGE_BUILDER)
IMAGE_TAG?=docker tag

APP_NAME=cluster-logging-operator
APP_REPO=github.com/openshift/$(APP_NAME)
TARGET=$(TARGET_DIR)/bin/$(APP_NAME)
DOCKER_TAG=quay.io/openshift/$(APP_NAME)
MAIN_PKG=cmd/$(APP_NAME)/main.go

OC?=oc

# These will be provided to the target
#VERSION := 1.0.0
#BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
#LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

#.PHONY: all build clean install uninstall fmt simplify check run
.PHONY: all build clean fmt simplify run deploy deploy-setup deploy-image undeploy test-e2e

all: build #check install

build: $(SRC)
	@mkdir -p $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/pkg $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/vendor/* $(TARGET_DIR)/src
	@GOPATH=$(GOPATH) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_PKG)

clean:
	@rm -rf $(TARGET_DIR)

image:
	$(IMAGE_BUILDER) -t $(DOCKER_TAG) . $(DOCKER_OPTS)

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

gendeepcopy:
	@operator-sdk generate k8s

deploy-setup:
	EXCLUSIONS="01-namespace.yaml 10-service-monitor-fluentd.yaml 05-deployment.yaml image-references" hack/deploy-setup.sh

deploy-image:
	hack/deploy-image.sh

deploy: deploy-setup image deploy-image
	hack/deploy.sh

test-e2e: image deploy-image
	hack/test-e2e.sh

undeploy:
	hack/undeploy.sh
