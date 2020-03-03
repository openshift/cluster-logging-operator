CURPATH=$(PWD)
TARGET_DIR=$(CURPATH)/_output

GOBUILD=go build
BUILD_GOPATH=$(TARGET_DIR):$(TARGET_DIR)/vendor:$(CURPATH)/cmd

IMAGE_BUILDER_OPTS=
IMAGE_BUILDER?=imagebuilder
IMAGE_BUILD=$(IMAGE_BUILDER)
export IMAGE_TAGGER?=docker tag

export APP_NAME=elasticsearch-operator
IMAGE_TAG?=quay.io/openshift/origin-$(APP_NAME):latest
export IMAGE_TAG
APP_REPO=github.com/openshift/$(APP_NAME)
TARGET=$(TARGET_DIR)/bin/$(APP_NAME)
KUBECONFIG?=$(HOME)/.kube/config
MAIN_PKG=cmd/manager/main.go
RUN_LOG?=elasticsearch-operator.log
RUN_PID?=elasticsearch-operator.pid
LOGGING_IMAGE_STREAM?=feature-es6x
OPERATOR_NAMESPACE=openshift-operators-redhat
DEPLOYMENT_NAMESPACE=openshift-logging
REPLICAS?=0

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

#.PHONY: all build clean install uninstall fmt simplify check run
.PHONY: all build clean fmt simplify run sec

all: build #check install

operator-sdk: get-dep
	@if ! type -p operator-sdk ; \
	then if [ ! -d $(GOPATH)/src/github.com/operator-framework/operator-sdk ] ; \
	  then git clone https://github.com/operator-framework/operator-sdk --branch master $(GOPATH)/src/github.com/operator-framework/operator-sdk ; \
	  fi ; \
	  cd $(GOPATH)/src/github.com/operator-framework/operator-sdk ; \
	  make dep ; \
	  make install || sudo make install || cd commands/operator-sdk && sudo go install ; \
	fi

gendeepcopy: operator-sdk
	@operator-sdk generate k8s

imagebuilder:
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && ! type -p imagebuilder ; \
	then go get -u github.com/openshift/imagebuilder/cmd/imagebuilder ; \
	fi

get-dep:
	@if ! type -p dep ; then \
		cd $(GOPATH) ; \
		if [ ! -d $(GOPATH)/bin ]; then \
			mkdir bin; \
			fi; \
			curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh ; \
	fi


build: fmt
	@mkdir -p $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/pkg $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/vendor/* $(TARGET_DIR)/src
	@GOPATH=$(BUILD_GOPATH) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_PKG)

clean:
	@rm -rf $(TARGET_DIR)

image: imagebuilder
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && [ $${SKIP_BUILD:-false} = false ] ; \
	then hack/build-image.sh $(IMAGE_TAG) $(IMAGE_BUILDER) $(IMAGE_BUILDER_OPTS) ; \
	fi

test-e2e: gen-example-certs
	LOGGING_IMAGE_STREAM=$(LOGGING_IMAGE_STREAM) REMOTE_CLUSTER=true hack/test-e2e.sh

test-unit:
	@go test -v ./pkg/... ./cmd/...

sec:
	go get -u github.com/securego/gosec/cmd/gosec
	gosec -severity medium -confidence medium -exclude G304 -quiet ./...

fmt:
	@gofmt -l -w cmd && \
	gofmt -l -w pkg && \
	gofmt -l -w test

simplify:
	@gofmt -s -l -w $(SRC)

deploy: deploy-image
	hack/deploy.sh
.PHONY: deploy

deploy-no-build:
	hack/deploy.sh
.PHONY: deploy-no-build

deploy-image: image
	hack/deploy-image.sh
.PHONY: deploy-image

deploy-example: deploy deploy-example-secret
	@oc create -n $(DEPLOYMENT_NAMESPACE) -f hack/cr.yaml
.PHONY: deploy-example

deploy-example-secret: gen-example-certs
	@oc -n $(DEPLOYMENT_NAMESPACE) delete secret elasticsearch ||: && \
	oc -n $(DEPLOYMENT_NAMESPACE) create secret generic elasticsearch  \
		--from-file=admin-key=/tmp/example-secrets/system.admin.key \
		--from-file=admin-cert=/tmp/example-secrets/system.admin.crt \
		--from-file=admin-ca=/tmp/example-secrets/ca.crt \
		--from-file=/tmp/example-secrets/elasticsearch.crt \
		--from-file=/tmp/example-secrets/logging-es.key \
		--from-file=/tmp/example-secrets/logging-es.crt \
		--from-file=/tmp/example-secrets/elasticsearch.key
.PHONY: deploy-example-secret

gen-example-certs:
	@rm -rf /tmp/example-secrets ||: \
	mkdir /tmp/example-secrets && \
	hack/cert_generation.sh /tmp/example-secrets $(DEPLOYMENT_NAMESPACE) elasticsearch
.PHONY: gen-example-certs

run: deploy deploy-example
	@ALERTS_FILE_PATH=files/prometheus_alerts.yml \
	RULES_FILE_PATH=files/prometheus_rules.yml \
	OPERATOR_NAME=elasticsearch-operator WATCH_NAMESPACE=$(DEPLOYMENT_NAMESPACE) \
	KUBERNETES_CONFIG=/etc/origin/master/admin.kubeconfig \
	go run ${MAIN_PKG} > $(RUN_LOG) 2>&1 & echo $$! > $(RUN_PID)

run-local:
	@ALERTS_FILE_PATH=files/prometheus_alerts.yml \
	RULES_FILE_PATH=files/prometheus_rules.yml \
	OPERATOR_NAME=elasticsearch-operator WATCH_NAMESPACE=$(DEPLOYMENT_NAMESPACE) \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	go run ${MAIN_PKG} LOG_LEVEL=debug
.PHONY: run-local

scale-cvo:
	@oc -n openshift-cluster-version scale deployment/cluster-version-operator --replicas=$(REPLICAS)
.PHONEY: scale-cvo

scale-olm:
	@oc -n openshift-operator-lifecycle-manager scale deployment/olm-operator --replicas=$(REPLICAS)
.PHONEY: scale-olm

undeploy:
	hack/undeploy.sh
.PHONY: undeploy
