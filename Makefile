CURPATH=$(PWD)
TARGET_DIR=$(CURPATH)/_output

GOBUILD=go build
BUILD_GOPATH=$(TARGET_DIR):$(TARGET_DIR)/vendor:$(CURPATH)/cmd

IMAGE_BUILDER_OPTS=
IMAGE_BUILDER?=imagebuilder
IMAGE_BUILD=$(IMAGE_BUILDER)
export IMAGE_TAG_CMD?=docker tag

export APP_NAME=cluster-logging-operator
APP_REPO=github.com/openshift/$(APP_NAME)
TARGET=$(TARGET_DIR)/bin/$(APP_NAME)
export IMAGE_TAG=openshift/$(APP_NAME):latest
MAIN_PKG=cmd/$(APP_NAME)/main.go
export NAMESPACE?=openshift-logging

OC?=oc

# These will be provided to the target
#VERSION := 1.0.0
#BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
#LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

#.PHONY: all build clean install uninstall fmt simplify check run
.PHONY: all operator-sdk imagebuilder build clean fmt simplify gendeepcopy deploy-setup deploy-image deploy deploy-example test-e2e undeploy

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
	@if ! type -p imagebuilder ; \
	then go get -u github.com/openshift/imagebuilder/cmd/imagebuilder ; \
	fi

build:
	@mkdir -p $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/pkg $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/vendor/* $(TARGET_DIR)/src
	@GOPATH=$(BUILD_GOPATH) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_PKG)

clean:
	@rm -rf $(TARGET_DIR)

image: imagebuilder
	$(IMAGE_BUILDER) -t $(IMAGE_TAG) . $(IMAGE_BUILDER_OPTS)

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

gendeepcopy: operator-sdk
	@operator-sdk generate k8s

deploy-setup:
	EXCLUSIONS="01-namespace.yaml 10-service-monitor-fluentd.yaml 05-deployment.yaml image-references" hack/deploy-setup.sh

deploy-image: image
	hack/deploy-image.sh

deploy: deploy-setup deploy-image
	hack/deploy.sh

deploy-example: deploy
	oc create -n $(NAMESPACE) -f hack/cr.yaml

test-e2e: deploy-image
	hack/test-e2e.sh

undeploy:
	hack/undeploy.sh
