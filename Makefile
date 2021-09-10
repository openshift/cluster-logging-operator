# Define the target to run if make is called with no arguments.
default: check

export LOG_LEVEL?=9
export KUBECONFIG?=$(HOME)/.kube/config

export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

include .bingo/Variables.mk

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=-mod=vendor
export GO111MODULE=on
export GODEBUG=x509ignoreCN=0

export APP_NAME=cluster-logging-operator
export IMAGE_TAG?=127.0.0.1:5000/openshift/origin-$(APP_NAME):latest

export LOGGING_VERSION?=$(shell basename $(shell ls -d manifests/[0-9]*))
export NAMESPACE?=openshift-logging

IMAGE_LOGGING_FLUENTD?=quay.io/openshift-logging/fluentd:1.7.4
REPLICAS?=0
export E2E_TEST_INCLUDES?=
export CLF_TEST_INCLUDES?=

.PHONY: force all build clean fmt generate regenerate deploy-setup deploy-image image deploy deploy-example test-functional test-unit test-e2e test-sec undeploy run

tools: $(BINGO) $(GOLANGCI_LINT) $(JUNITREPORT) $(OPERATOR_SDK) $(OPM) $(KUSTOMIZE) $(CONTROLLER_GEN)

# check health of the code:
# - Update generated code
# - Apply standard source format
# - Run unit tests
# - Build all code (including e2e tests)
# - Run lint check
#
check: generate fmt test-unit bin/forwarder-generator bin/cluster-logging-operator bin/functional-benchmarker
	go test ./test/... -exec true > /dev/null # Build but don't run e2e tests.
	go test ./test/functional/... -exec true > /dev/null # Build but don't run test functional tests.
	go test ./test/helpers/... -exec true > /dev/null # Build but don't run test helpers tests.
	$(MAKE) lint				  # Only lint if all code builds.

# CI calls ci-check first.
ci-check: check
	@echo
	@git diff-index --name-status --exit-code HEAD || { \
		echo -e '\nerror: files changed during "make check", not up-to-date\n' ; \
		exit 1 ; \
	}

# .target is used to hold timestamp files to avoid un-necessary rebuilds. Do NOT check in.
.target:
	mkdir -p .target

# Note: Go has built-in build caching, so always run `go build`.
# It will do a better job than using source dependencies to decide if we need to build.
bin/functional-benchmarker: force
	go build $(BUILD_OPTS) -o $@ ./internal/cmd/functional-benchmarker

bin/forwarder-generator: force
	go build $(BUILD_OPTS) -o $@ ./internal/cmd/forwarder-generator

bin/cluster-logging-operator: force
	go build $(BUILD_OPTS) -o $@

openshift-client:
	@type -p oc > /dev/null || bash hack/get-openshift-client.sh

build: bin/cluster-logging-operator

build-debug:
	$(MAKE) build BUILD_OPTS='-gcflags=all="-N -l"'

# Run the CLO locally - see HACKING.md
RUN_CMD?=go run
run:
	@ls $(MANIFESTS)/*crd.yaml | xargs -n1 oc apply -f
	@mkdir -p $(CURDIR)/tmp
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	OPERATOR_NAME=cluster-logging-operator \
	WATCH_NAMESPACE=$(NAMESPACE) \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(CURDIR)/tmp \
	$(RUN_CMD) main.go

run-debug:
	$(MAKE) run RUN_CMD='dlv debug'

scale-cvo:
	@oc -n openshift-cluster-version scale deployment/cluster-version-operator --replicas=$(REPLICAS)
.PHONY: scale-cvo

scale-olm:
	@oc -n openshift-operator-lifecycle-manager scale deployment/olm-operator --replicas=$(REPLICAS)
.PHONY: scale-olm

clean:
	rm -rf bin tmp _output .target
	go clean -cache -testcache ./...

PATCH?=Dockerfile.patch
image: .target/image
.target/image: .target $(shell find must-gather version scripts files vendor manifests .bingo apis controllers internal -type f) Makefile Dockerfile  go.mod go.sum
	patch -o Dockerfile.local Dockerfile $(PATCH)
	podman build -t $(IMAGE_TAG) . -f Dockerfile.local
	touch $@

lint: $(GOLANGCI_LINT) lint-dockerfile
	@GOLANGCI_LINT_CACHE="$(CURDIR)/.cache" $(GOLANGCI_LINT) run -c golangci.yaml
.PHONY: lint

lint-dockerfile:
	@hack/run-linter
.PHONY: lint-dockerfile

fmt:
	@echo gofmt		# Show progress, real gofmt line is too long
	find test internal controllers apis -name '*.go' | xargs gofmt -s -l -w

MANIFESTS=manifests/$(LOGGING_VERSION)

# Do all code/CRD generation at once, with timestamp file to check out-of-date.
GEN_TIMESTAMP=.zz_generate_timestamp
generate: $(GEN_TIMESTAMP) $(OPERATOR_SDK) $(CONTROLLER_GEN)
$(GEN_TIMESTAMP): $(shell find apis -name '*.go')
	@$(CONTROLLER_GEN) object paths="./apis/..."
	@$(CONTROLLER_GEN) crd:crdVersions=v1 rbac:roleName=clusterlogging-operator paths="./..." output:crd:artifacts:config=config/crd/bases
	@bash ./hack/generate-crd.sh
	@$(MAKE) fmt
	@touch $@

regenerate: $(OPERATOR_SDK) $(CONTROLLER_GEN)
	@rm -f $(GEN_TIMESTAMP)
	@$(MAKE) generate

deploy-image: image
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

test-functional: test-functional-benchmarker
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	LOGGING_SHARE_DIR=$(CURDIR)/files \
	SCRIPTS_DIR=$(CURDIR)/scripts \
	go test -race ./test/functional/... -ginkgo.noColor -timeout=40m
	go test -cover -race ./test/helpers/...
.PHONY: test-functional

test-forwarder-generator: bin/forwarder-generator
	@bin/forwarder-generator --file hack/logforwarder.yaml > /dev/null
.PHONY: test-forwarder-generator

test-functional-benchmarker: bin/functional-benchmarker
	@bin/functional-benchmarker > /dev/null
.PHONY: test-functional-benchmarker

test-unit: test-forwarder-generator
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	go test -cover -race ./internal/... ./test ./test/helpers ./test/matchers ./test/runtime

test-cluster:
	go test  -cover -race ./test/... -- -root=$(CURDIR)

OPENSHIFT_VERSIONS?="v4.7"
CHANNELS="stable,stable-${LOGGING_VERSION}"
DEFAULT_CHANNEL="stable"
generate-bundle: regenerate $(OPM)
	MANIFEST_VERSION=${LOGGING_VERSION} OPENSHIFT_VERSIONS=${OPENSHIFT_VERSIONS} CHANNELS=${CHANNELS} DEFAULT_CHANNEL=${DEFAULT_CHANNEL} hack/generate-bundle.sh
.PHONY: generate-bundle

bundle: generate-bundle
.PHONY: bundle

# NOTE: This is the CI e2e entry point.
test-e2e-olm: $(JUNITREPORT)
	INCLUDES="$(E2E_TEST_INCLUDES)" CLF_INCLUDES="$(CLF_TEST_INCLUDES)" hack/test-e2e-olm.sh

test-e2e-local: $(JUNITREPORT) deploy-image
	CLF_INCLUDES=$(CLF_TEST_INCLUDES) \
	INCLUDES=$(E2E_TEST_INCLUDES) \
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:latest \
	IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=image-registry.openshift-image-registry.svc:5000/openshift/cluster-logging-operator-registry:latest \
	hack/test-e2e-olm.sh

test-svt:
	hack/svt/test-svt.sh

undeploy:
	hack/undeploy.sh

redeploy:
	$(MAKE) undeploy
	$(MAKE) deploy

undeploy-all: undeploy undeploy-elasticsearch-operator

cluster-logging-catalog: cluster-logging-catalog-build cluster-logging-catalog-deploy

cluster-logging-cleanup: cluster-logging-operator-uninstall cluster-logging-catalog-uninstall

# builds an operator-registry image containing the cluster-logging operator
cluster-logging-catalog-build: .target/cluster-logging-catalog-build
.target/cluster-logging-catalog-build: $(shell find olm_deploy -type f)
	olm_deploy/scripts/catalog-build.sh
	touch $@

# deploys the operator registry image and creates a catalogsource referencing it
cluster-logging-catalog-deploy: .target/cluster-logging-catalog-deploy
.target/cluster-logging-catalog-deploy: $(shell find olm_deploy -type f)
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

gen-dockerfiles:
	./hack/generate-dockerfile-from-midstream > Dockerfile
