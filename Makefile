KUBECONFIG?=$(HOME)/.kube/config

export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

export IMAGE_TAG_CMD?=docker tag

export APP_NAME=cluster-logging-operator
export IMAGE_TAG?=quay.io/openshift/origin-$(APP_NAME):latest

export OCP_VERSION?=4.5
export NAMESPACE?=openshift-logging

FLUENTD_IMAGE?=quay.io/openshift/origin-logging-fluentd:latest

.PHONY: all build clean fmt generate regenerate deploy-setup deploy-image image deploy deploy-example test-unit test-e2e test-sec undeploy run operator-sdk imagebuilder golangci-lint

# Run all local pre-merge tests. CI runs additional system tests on the image.
all: generate fmt test-unit lint

# Download tools if not available on local PATH.
operator-sdk:
	@type -p operator-sdk > /dev/null || bash hack/get-operator-sdk.sh
imagebuilder:
	@type -p imagebuilder > /dev/null || go get github.com/openshift/imagebuilder/cmd/imagebuilder
golangci-lint:
	@type -p golangci-lint > /dev/null || go get github.com/golangci/golangci-lint/cmd/golangci-lint

build:
	go build $(BUILD_OPTS) -o bin/cluster-logging-operator ./cmd/manager

build-debug:
	$(MAKE) build BUILD_OPTS='-gcflags=all="-N -l"'

run:
	ELASTICSEARCH_IMAGE=quay.io/openshift/origin-logging-elasticsearch6:latest \
	FLUENTD_IMAGE=$(FLUENTD_IMAGE) \
	KIBANA_IMAGE=quay.io/openshift/origin-logging-kibana6:latest \
	CURATOR_IMAGE=quay.io/openshift/origin-logging-curator6:latest \
	OAUTH_PROXY_IMAGE=quay.io/openshift/origin-oauth-proxy:latest \
	PROMTAIL_IMAGE=quay.io/openshift/origin-promtail:latest \
	OPERATOR_NAME=cluster-logging-operator \
	WATCH_NAMESPACE=$(NAMESPACE) \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(TARGET_DIR)/ocp-clo \
	LOGGING_SHARE_DIR=$(CURDIR)/files \
	$(RUN_OPTS) bin/cluster-logging-operator

run-debug:
	$(MAKE) run RUN_OPTS='dlv exec --'

clean:
	@rm -f bin/operator-sdk bin/imagebuilder bin/golangci-lint bin/cluster-logging-operator
	go clean -cache -testcache ./...

image: build
	$(MAKE) imagebuilder
	@if [ $${USE_IMAGE_STREAM:-false} = false ] && [ $${SKIP_BUILD:-false} = false ] ; \
	then hack/build-image.sh $(IMAGE_TAG) imagebuilder $(IMAGE_BUILDER_OPTS) ; \
	fi

lint: fmt
	@$(MAKE) golangci-lint
	golangci-lint run -c golangci.yaml

fmt:
	@echo gofmt		# Show progress, real gofmt line is too long
	@gofmt -s -l -w $(shell find pkg cmd -name '*.go')

# Do all code/CRD generation at once, with timestamp file to check out-of-date.
GEN_TIMESTAMP=.zz_generate_timestamp
generate: $(GEN_TIMESTAMP)
$(GEN_TIMESTAMP): $(shell find pkg/apis -name '*.go')
	@echo generating code
	@$(MAKE) operator-sdk
	@operator-sdk generate k8s
	@operator-sdk generate crds
	@$(MAKE) fmt
	@touch $@

regenerate:
	@rm -f $(GEN_TIMESTAMP)
	@$(MAKE) generate

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

test-unit:
	go test ./pkg/...

test-e2e:
	hack/test-e2e.sh

test-e2e-olm:
	hack/test-e2e-olm.sh

test-e2e-local: deploy-image
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
	hack/test-e2e.sh

test-sec:
	go get -u github.com/securego/gosec/cmd/gosec
	gosec -severity medium --confidence medium -quiet ./...

undeploy:
	hack/undeploy.sh



cluster-logging-catalog: cluster-logging-catalog-build cluster-logging-catalog-deploy

cluster-logging-cleanup: cluster-logging-operator-uninstall cluster-logging-catalog-uninstall

# builds an operator-registry image containing the cluster-logging operator
cluster-logging-catalog-build:
	olm_deploy/scripts/catalog-build.sh

# deploys the operator registry image and creates a catalogsource referencing it
cluster-logging-catalog-deploy:
	olm_deploy/scripts/catalog-deploy.sh

# deletes the catalogsource and catalog namespace
cluster-logging-catalog-uninstall:
	olm_deploy/scripts/catalog-uninstall.sh

# installs the cluster-logging operator from the deployed operator-registry/catalogsource.
cluster-logging-operator-install:
	olm_deploy/scripts/operator-install.sh

# uninstalls the cluster-logging operator
cluster-logging-operator-uninstall:
	olm_deploy/scripts/operator-uninstall.sh
