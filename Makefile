# Define the target to run if make is called with no arguments.
default: pre-commit

export LOG_LEVEL?=9
export KUBECONFIG?=$(HOME)/.kube/config

export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

include .bingo/Variables.mk

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=
export GO111MODULE=on
export GODEBUG=x509ignoreCN=0

export APP_NAME=cluster-logging-operator
export CURRENT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD;)
export IMAGE_TAG?=127.0.0.1:5000/openshift/origin-$(APP_NAME):$(CURRENT_BRANCH)

export LOGGING_VERSION?=$(shell basename $(shell ls -d manifests/[0-9]*))
export NAMESPACE?=openshift-logging

IMAGE_LOGGING_FLUENTD?=quay.io/openshift-logging/fluentd:1.14.5
IMAGE_LOGGING_VECTOR?=quay.io/openshift-logging/vector:0.21-rh
IMAGE_LOGFILEMETRICEXPORTER?=quay.io/openshift-logging/log-file-metric-exporter:1.1
REPLICAS?=0
export E2E_TEST_INCLUDES?=
export CLF_TEST_INCLUDES?=

.PHONY: force all build clean generate regenerate deploy-setup deploy-image image deploy deploy-example test-functional test-unit test-e2e test-sec undeploy run

.PHONY: tools
tools: $(BINGO) $(GOLANGCI_LINT) $(JUNITREPORT) $(OPERATOR_SDK) $(OPM) $(KUSTOMIZE) $(CONTROLLER_GEN)

.PHONY: pre-commit
# Should pass when run before commit.
pre-commit: clean generate-bundle check

.PHONY: check
# check health of the code:
# - Update generated code
# - Build all code and test
# - Run lint, automatically fix trivial issues
# - Run unit tests
#
check: compile-tests bin/forwarder-generator bin/cluster-logging-operator bin/functional-benchmarker lint test-unit

# Compile all tests and code but don't run the tests.
compile-tests: generate
	go test ./test/... -exec true > /dev/null # Build all tests but don't run

.PHONY: ci-check
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

.PHONY: openshift-client
openshift-client:
	@type -p oc > /dev/null || bash hack/get-openshift-client.sh

.PHONY: build
build: bin/cluster-logging-operator

.PHONY: build-debug
build-debug:
	$(MAKE) build BUILD_OPTS='-gcflags=all="-N -l"'

# Run the CLO locally - see HACKING.md
RUN_CMD?=go run

.PHONY: run
run:
	@ls $(MANIFESTS)/*crd.yaml | xargs -n1 oc apply -f
	@mkdir -p $(CURDIR)/tmp
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	VECTOR_IMAGE=$(IMAGE_LOGGING_VECTOR) \
	LOGFILEMETRICEXPORTER_IMAGE=$(IMAGE_LOGFILEMETRICEXPORTER) \
	OPERATOR_NAME=cluster-logging-operator \
	WATCH_NAMESPACE=$(NAMESPACE) \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(CURDIR)/tmp \
	$(RUN_CMD) main.go

.PHONY: run-debug
run-debug:
	$(MAKE) run RUN_CMD='dlv debug'

.PHONY: scale-cvo
scale-cvo:
	@oc -n openshift-cluster-version scale deployment/cluster-version-operator --replicas=$(REPLICAS)

.PHONY: scale-olm
scale-olm:
	@oc -n openshift-operator-lifecycle-manager scale deployment/olm-operator --replicas=$(REPLICAS)

.PHONY: clean
clean:
	rm -rf bin/cluster-logging-operator bin/forwarder-generator bin/functional-benchmarker tmp _output .target .cache
	find -name .kube | xargs rm -rf
	go clean -cache -testcache ./...

.PHONY: image
image: .target/image
.target/image: .target $(shell find must-gather version scripts files manifests .bingo apis controllers internal -type f) Makefile Dockerfile  go.mod go.sum
	podman build -t $(IMAGE_TAG) . -f Dockerfile
	touch $@

# Notes:
# - override the .cache dir for CI, where $HOME/.cache may not be writable.
# - don't run with --fix in CI, complain about everything. Do try to auto-fix outside of CI.
export GOLANGCI_LINT_CACHE=$(CURDIR)/.cache
lint:  $(GOLANGCI_LINT) lint-dockerfile
	$(GOLANGCI_LINT) run --color=never $(if $(CI),,--fix)
.PHONY: lint

.PHONY: lint-dockerfile
lint-dockerfile:
	@hack/run-linter


.PHONY: fmt
fmt:
	@echo gofmt		# Show progress, real gofmt line is too long
	find version test internal controllers apis -name '*.go' | xargs gofmt -s -l -w

MANIFESTS=manifests/$(LOGGING_VERSION)

# Do all code/CRD generation at once, with timestamp file to check out-of-date.
GEN_TIMESTAMP=.target/codegen
DEFAULT_VERSION=5.5
generate: $(GEN_TIMESTAMP)
$(GEN_TIMESTAMP): $(shell find apis -name '*.go')  $(OPERATOR_SDK) $(CONTROLLER_GEN) $(KUSTOMIZE) .target
	@$(CONTROLLER_GEN) object paths="./apis/..."
	@$(CONTROLLER_GEN) crd:crdVersions=v1 rbac:roleName=clusterlogging-operator paths="./..." output:crd:artifacts:config=config/crd/bases
	@bash ./hack/generate-crd.sh
	echo -e "package version\n\nvar Version = \"$(or $(CI_CONTAINER_VERSION),$(LOGGING_VERSION), DEFAULT_VERSION)\"" > version/version.go
	@touch $@

.PHONY: regenerate
regenerate: $(OPERATOR_SDK) $(CONTROLLER_GEN) $(KUSTOMIZE)
	@rm -f $(GEN_TIMESTAMP)
	@$(MAKE) generate

.PHONY: deploy-image
deploy-image: image
	hack/deploy-image.sh

.PHONY: deploy
deploy:  deploy-image deploy-elasticsearch-operator deploy-catalog install

.PHONY: install
install:
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:$(CURRENT_BRANCH) \
	$(MAKE) cluster-logging-operator-install

.PHONY: deploy-catalog
deploy-catalog: .target
	LOCAL_IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=127.0.0.1:5000/openshift/cluster-logging-operator-registry:$(CURRENT_BRANCH) \
	$(MAKE) cluster-logging-catalog-build
	IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=image-registry.openshift-image-registry.svc:5000/openshift/cluster-logging-operator-registry:$(CURRENT_BRANCH) \
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:$(CURRENT_BRANCH) \
	$(MAKE) cluster-logging-catalog-deploy

.PHONY: deploy-elasticsearch-operator
deploy-elasticsearch-operator:
	hack/deploy-eo.sh

.PHONY: undeploy-elasticsearch-operator
undeploy-elasticsearch-operator:
	make -C ../elasticsearch-operator elasticsearch-cleanup

.PHONY: deploy-example
deploy-example: deploy
	oc create -n $(NAMESPACE) -f hack/cr.yaml

.PHONY: test-functional
test-functional: test-functional-benchmarker test-functional-fluentd test-functional-vector
	VECTOR_IMAGE=$(IMAGE_LOGGING_VECTOR) \
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	LOGFILEMETRICEXPORTER_IMAGE=$(IMAGE_LOGFILEMETRICEXPORTER) \
	go test -cover -race ./test/helpers/... $(GOTESTFLAGS)

.PHONY: test-functional-fluentd
test-functional-fluentd:
	@echo == $@
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	LOGFILEMETRICEXPORTER_IMAGE=$(IMAGE_LOGFILEMETRICEXPORTER) \
	go test --tags=fluentd -race ./test/functional/... -ginkgo.noColor -timeout=40m -ginkgo.slowSpecThreshold=45.0 $(GOTESTFLAGS)

.PHONY: test-functional-vector
test-functional-vector:
	@echo == $@
	VECTOR_IMAGE=$(IMAGE_LOGGING_VECTOR) \
	LOGFILEMETRICEXPORTER_IMAGE=$(IMAGE_LOGFILEMETRICEXPORTER) \
	go test --tags=vector -race ./test/functional/outputs/cloudwatch/... ./test/functional/outputs/elasticsearch/... ./test/functional/normalization -ginkgo.noColor -timeout=40m -ginkgo.slowSpecThreshold=45.0  $(GOTESTFLAGS)

.PHONY: test-forwarder-generator
test-forwarder-generator: bin/forwarder-generator
	@bin/forwarder-generator --file hack/logforwarder.yaml --collector=fluentd > /dev/null
	@bin/forwarder-generator --file hack/logforwarder.yaml --collector=vector > /dev/null


.PHONY: test-functional-benchmarker
test-functional-benchmarker: bin/functional-benchmarker
	@out=$$(bin/functional-benchmarker --artifact-dir=/tmp/benchmark-test 2>&1); if [ "$$?" != "0" ] ; then echo "$$out"; exit 1; fi

.PHONY: test-unit
test-unit: test-forwarder-generator
	VECTOR_IMAGE=$(IMAGE_LOGGING_VECTOR) \
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	LOGFILEMETRICEXPORTER_IMAGE=$(IMAGE_LOGFILEMETRICEXPORTER) \
	go test -cover -race ./internal/... `go list ./test/... | grep -Ev 'test/(e2e|functional|client|helpers)'` $(GOTESTFLAGS)

.PHONY: test-cluster
test-cluster:
	go test  -cover -race ./test/... $(GOTESTFLAGS) -- -root=$(CURDIR)

OPENSHIFT_VERSIONS?="v4.7"
CHANNELS="stable,stable-${LOGGING_VERSION}"
DEFAULT_CHANNEL="stable"

.PHONY: generate-bundle
generate-bundle: regenerate $(OPM)
	MANIFEST_VERSION=${LOGGING_VERSION} OPENSHIFT_VERSIONS=${OPENSHIFT_VERSIONS} CHANNELS=${CHANNELS} DEFAULT_CHANNEL=${DEFAULT_CHANNEL} hack/generate-bundle.sh

.PHONY: bundle
bundle: generate-bundle

.PHONY: test-e2e-olm
# NOTE: This is the CI e2e entry point.
test-e2e-olm: $(JUNITREPORT)
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) INCLUDES="$(E2E_TEST_INCLUDES)" CLF_INCLUDES="$(CLF_TEST_INCLUDES)" LOG_LEVEL=3 hack/test-e2e-olm.sh

.PHONY: test-e2e-local
test-e2e-local: $(JUNITREPORT) deploy-image
	FLUENTD_IMAGE=$(IMAGE_LOGGING_FLUENTD) \
	CLF_INCLUDES=$(CLF_TEST_INCLUDES) \
	INCLUDES=$(E2E_TEST_INCLUDES) \
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:$(CURRENT_BRANCH) \
	IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY=image-registry.openshift-image-registry.svc:5000/openshift/cluster-logging-operator-registry:$(CURRENT_BRANCH) \
	hack/test-e2e-olm.sh

.PHONY: test-e2e-clo-metric
test-e2e-clo-metric: 
	test/e2e/telemetry/clometrics_test.sh

.PHONY: test-svt
test-svt:
	hack/svt/test-svt.sh

.PHONY: undeploy
undeploy:
	hack/undeploy.sh

.PHONY: redeploy
redeploy:
	$(MAKE) undeploy
	$(MAKE) deploy

.PHONY: undeploy-all
undeploy-all: undeploy undeploy-elasticsearch-operator

.PHONY: cluster-logging-catalog
cluster-logging-catalog: cluster-logging-catalog-build cluster-logging-catalog-deploy

.PHONY: cluster-logging-cleanup
cluster-logging-cleanup: cluster-logging-operator-uninstall cluster-logging-catalog-uninstall

.PHONY: cluster-logging-catalog-build
# builds an operator-registry image containing the cluster-logging operator
cluster-logging-catalog-build: .target/cluster-logging-catalog-build
.target/cluster-logging-catalog-build: $(shell find olm_deploy -type f)
	olm_deploy/scripts/catalog-build.sh
	touch $@

.PHONY: cluster-logging-catalog-deploy
# deploys the operator registry image and creates a catalogsource referencing it
cluster-logging-catalog-deploy: .target/cluster-logging-catalog-deploy
.target/cluster-logging-catalog-deploy: $(shell find olm_deploy -type f)
	olm_deploy/scripts/catalog-deploy.sh

.PHONY: cluster-logging-catalog-uninstall
# deletes the catalogsource and catalog namespace
cluster-logging-catalog-uninstall:
	olm_deploy/scripts/catalog-uninstall.sh

.PHONY: cluster-logging-operator-install
# installs the cluster-logging operator from the deployed operator-registry/catalogsource.
cluster-logging-operator-install:
	olm_deploy/scripts/operator-install.sh

.PHONY: cluster-logging-operator-uninstall
# uninstalls the cluster-logging operator
cluster-logging-operator-uninstall:
	olm_deploy/scripts/operator-uninstall.sh

.PHONY: gen-dockerfiles
gen-dockerfiles:
	./hack/generate-dockerfile-from-midstream > Dockerfile
