export KUBECONFIG?=$(HOME)/.kube/config

export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=-mod=vendor
export GO111MODULE=on

export APP_NAME=cluster-logging-operator
export IMAGE_TAG?=127.0.0.1:5000/openshift/origin-$(APP_NAME):latest

export OCP_VERSION?=$(shell basename $(shell find manifests/  -maxdepth 1  -not -name manifests -type d))
export NAMESPACE?=openshift-logging

FLUENTD_IMAGE?=quay.io/openshift/origin-logging-fluentd:latest

.PHONY: all build clean fmt generate regenerate deploy-setup deploy-image image deploy deploy-example test-unit test-e2e test-sec undeploy run operator-sdk golangci-lint

# Update code (generate, format), run unit tests and lint. Make sure e2e tests build.
check: generate fmt test-unit
	go test ./test/... -exec true > /dev/null # Build but don't run e2e tests.
	$(MAKE) lint

# Download tools if not available on local PATH.
operator-sdk:
	@type -p operator-sdk > /dev/null || bash hack/get-operator-sdk.sh

golangci-lint:
	@type -p golangci-lint > /dev/null || GOFLAGS="" GO111MODULE=off go get github.com/golangci/golangci-lint/cmd/golangci-lint

build:
	go build $(BUILD_OPTS) -o bin/cluster-logging-operator ./cmd/manager

build-debug:
	$(MAKE) build BUILD_OPTS='-gcflags=all="-N -l"'

# Run the CLO locally - see HACKING.md
RUN_CMD?=go run
run: deploy-elasticsearch-operator test-cleanup
	@ls $(MANIFESTS)/*crd.yaml | xargs -n1 oc apply -f
	@mkdir -p $(CURDIR)/tmp
	@ELASTICSEARCH_IMAGE=quay.io/openshift/origin-logging-elasticsearch6:latest \
	FLUENTD_IMAGE=$(FLUENTD_IMAGE) \
	KIBANA_IMAGE=quay.io/openshift/origin-logging-kibana6:latest \
	CURATOR_IMAGE=quay.io/openshift/origin-logging-curator6:latest \
	OAUTH_PROXY_IMAGE=quay.io/openshift/origin-oauth-proxy:latest \
	PROMTAIL_IMAGE=quay.io/openshift/origin-promtail:latest \
	OPERATOR_NAME=cluster-logging-operator \
	WATCH_NAMESPACE=$(NAMESPACE) \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(CURDIR)/tmp \
	LOGGING_SHARE_DIR=$(CURDIR)/files \
	$(RUN_CMD) cmd/manager/main.go

run-debug:
	$(MAKE) run RUN_CMD='dlv debug'

clean:
	@rm -f bin/operator-sdk bin/imagebuilder bin/golangci-lint bin/cluster-logging-operator
	rm -rf tmp _output
	go clean -cache -testcache ./...

image:
	@if [ $${SKIP_BUILD:-false} = false ] ; then \
		podman build -t $(IMAGE_TAG) . ; \
	fi

lint: fmt
	@$(MAKE) golangci-lint
	golangci-lint run -c golangci.yaml

fmt:
	@echo gofmt		# Show progress, real gofmt line is too long
	@gofmt -s -l -w $(shell find pkg cmd -name '*.go')

# Do all code/CRD generation at once, with timestamp file to check out-of-date.
GEN_TIMESTAMP=.zz_generate_timestamp
MANIFESTS=manifests/$(OCP_VERSION)
generate: $(GEN_TIMESTAMP)
$(GEN_TIMESTAMP): $(shell find pkg/apis -name '*.go')
	@echo generating code
	@$(MAKE) operator-sdk
	@operator-sdk generate k8s
	@operator-sdk generate crds
	@mv deploy/crds/logging.openshift.io_clusterlogforwarders_crd.yaml $(MANIFESTS)
	@rm -rf deploy
	@$(MAKE) fmt
	@touch $@

regenerate:
	@rm -f $(GEN_TIMESTAMP) $(shell find pkg -name zz_generated_*.go)
	@$(MAKE) generate

deploy-image: image
	hack/deploy-image.sh

deploy-image-no-build:
	hack/deploy-image.sh

deploy:  deploy-image deploy-elasticsearch-operator deploy-catalog install

install:
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
	$(MAKE) cluster-logging-operator-install

deploy-catalog:
	LOCAL_IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=127.0.0.1:5000/openshift/cluster-logging-operator-registry \
	$(MAKE) cluster-logging-catalog-build
	IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=image-registry.openshift-image-registry.svc:5000/openshift/cluster-logging-operator-registry \
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
	$(MAKE) cluster-logging-catalog-deploy

deploy-elasticsearch-operator:
	hack/deploy-eo.sh
undeploy-elasticsearch-operator:
	make -C ../elasticsearch-operator elasticsearch-cleanup

deploy-example: deploy
	oc create -n $(NAMESPACE) -f hack/cr.yaml

test-unit:
	LOGGING_SHARE_DIR=$(CURDIR)/files \
	LOG_LEVEL=fatal \
	go test -cover -race ./pkg/...

test-cluster:
	go test  -cover -race ./test/... -- -root=$(CURDIR)

test-cleanup:			# Only needed if e2e tests are interrupted.
	oc delete --ignore-not-found -n $(NAMESPACE) all --all
	oc delete --ignore-not-found ns $(NAMESPACE)
	oc create ns $(NAMESPACE)
	@for TYPE in sa role rolebinding configmap; do oc get $$TYPE -o name; done | grep -e -receiver | xargs --verbose --no-run-if-empty oc delete
	@oc get ns -o name | grep clo- | xargs --verbose --no-run-if-empty oc delete

# NOTE: This is the CI e2e entry point.
test-e2e-olm:
	hack/test-e2e-olm.sh

test-e2e-local: deploy-image
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
	hack/test-e2e.sh

test-sec:
	GOFLAGS="" GO111MODULE=off go get -u github.com/securego/gosec/cmd/gosec
	gosec -severity medium --confidence medium -quiet ./...

undeploy:
	hack/undeploy.sh
redeploy:
	$(MAKE) undeploy
	$(MAKE) deploy

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
