
# Define the target to run if make is called with no arguments.
.PHONY: default
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

ifdef OVERLAY
$(if $(wildcard $(OVERLAY)),,$(error "OVERLAY file '$(OVERLAY)' does not exist"))
# Set variables from overlay file instead of environment variables.
KUSTOMIZE_VALUE=$(shell grep '$(1): *' $(OVERLAY)/kustomization.yaml | sed 's/^.*$(1): *//')
IMAGE_NAME=$(or $(call KUSTOMIZE_VALUE,newName),$(call KUSTOMIZE_VALUE,name))
VERSION=$(call KUSTOMIZE_VALUE,newTag)
NAMESPACE=$(shell awk '/namespace:/{print $$2}' $(OVERLAY)/kustomization.yaml)

# Get operand image names from the deployment patch in the overlay.
DEPLOY_IMG=$(shell awk '/name:/ {NAME = $$NF} /value: / { if (NAME == "$(1)") { print $$NF; exit 0; }  }' $(OVERLAY_DIR)/deployment_patch.yaml)
IMAGE_LOGGING_VECTOR=$(call DEPLOY_ENV,RELATED_IMAGE_VECTOR)
IMAGE_LOGGING_FLUENTD=$(call DEPLOY_ENV,RELATED_IMAGE_FLUENTD)
IMAGE_LOGFILEMETRICEXPORTER=$(call DEPLOY_ENV,RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER)
IMAGE_LOGGING_CONSOLE_PLUGIN?=$(call DEPLOY_ENV,RELATED_IMAGE_LOGGING_CONSOLE_PLUGIN)

export IMAGE_TAG=$(IMAGE_NAME):$(VERSION)
BUNDLE_TAG=$(IMAGE_NAME)-bundle:$(VERSION)
LOGGING_VERSION=$(shell echo "$(VERSION)" | grep -o '^[0-9]\+\.[0-9]\+')

else
# Set variables from environment or hard-coded default

export OPERATOR_NAME=cluster-logging-operator
export CURRENT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD;)
export IMAGE_TAG?=127.0.0.1:5000/openshift/origin-$(OPERATOR_NAME):$(CURRENT_BRANCH)
BUNDLE_TAG=$(error set OVERLAY to deploy or run a bundle)

export LOGGING_VERSION?=5.7
export VERSION=$(LOGGING_VERSION).0
export NAMESPACE?=openshift-logging

IMAGE_LOGGING_FLUENTD?=quay.io/openshift-logging/fluentd:1.14.6
IMAGE_LOGGING_VECTOR?=quay.io/openshift-logging/vector:0.21-rh
IMAGE_LOGFILEMETRICEXPORTER?=quay.io/openshift-logging/log-file-metric-exporter:1.1
# Note: use logging-view-plugin:latest to pick up improvements in the console automatically.
# Unlike the other components, console changes do not risk breaking the collector,
# the console depends only on the format of the LokiStore records.
IMAGE_LOGGING_CONSOLE_PLUGIN?=quay.io/openshift-logging/logging-view-plugin:latest
endif # ifdef OVERLAY

REPLICAS?=0
export E2E_TEST_INCLUDES?=
export CLF_TEST_INCLUDES?=

.PHONY: force

.PHONY: tools
tools: $(BINGO) $(GOLANGCI_LINT) $(JUNITREPORT) $(OPERATOR_SDK) $(OPM) $(KUSTOMIZE) $(CONTROLLER_GEN) $(GEN_CRD_API_REFERENCE_DOCS)

.PHONY: pre-commit
# Should pass when run before commit.
pre-commit: clean bundle check docs

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

docs: docs/reference/operator/api.adoc docs/reference/datamodels/viaq/v1.adoc
.PHONY: docs

docs/reference/operator/api.adoc: $(GEN_CRD_API_REFERENCE_DOCS)
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir "github.com/openshift/cluster-logging-operator/apis/" -config "$(PWD)/config/docs/config.json" -template-dir "$(PWD)/config/docs/templates/apis/asciidoc" -out-file "$(PWD)/$@"
.PHONY: docs/reference/operator/api.adoc

docs/reference/datamodels/viaq/v1.adoc: $(GEN_CRD_API_REFERENCE_DOCS)
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir "github.com/openshift/cluster-logging-operator/internal/datamodels/viaq/v1" -config "$(PWD)/config/docs/config.json" -template-dir "$(PWD)/config/docs/templates/datamodels/asciidoc" -out-file "$(PWD)/$@"
.PHONY: docs/reference/datamodels/viaq/v1.adoc

# Run the CLO locally - see HACKING.md
RUN_CMD?=go run

.PHONY: run
run:
	@ls ./bundle/manifests/logging.openshift.io_*.yaml | xargs -n1 oc apply -f
	@mkdir -p $(CURDIR)/tmp
	LOG_LEVEL=$(LOG_LEVEL) \
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) \
	RELATED_IMAGE_VECTOR=$(IMAGE_LOGGING_VECTOR) \
	RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER=$(IMAGE_LOGFILEMETRICEXPORTER) \
	RELATED_IMAGE_LOGGING_CONSOLE_PLUGIN=$(IMAGE_LOGGING_CONSOLE_PLUGIN) \
	OPERATOR_NAME=$(OPERATOR_NAME) \
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

spotless: clean
	go clean -cache -testcache ./...

.PHONY: image
image: .target/image
.target/image: .target $(GEN_TIMESTAMP) $(shell find must-gather version files bundle .bingo apis controllers internal -type f) Dockerfile  go.mod go.sum
	podman build -t $(IMAGE_TAG) . -f Dockerfile
	touch $@

# Notes:
# - override the .cache dir for CI, where $HOME/.cache may not be writable.
# - don't run with --fix in CI, complain about everything. Do try to auto-fix outside of CI.
export GOLANGCI_LINT_CACHE=$(CURDIR)/.cache
lint:  $(GOLANGCI_LINT) lint-dockerfile
	$(GOLANGCI_LINT) run --color=never --timeout=3m $(if $(CI),,--fix)
.PHONY: lint

.PHONY: lint-dockerfile
lint-dockerfile:
	@hack/run-linter

.PHONY: fmt
fmt:
	@echo gofmt		# Show progress, real gofmt line is too long
	find version test internal controllers apis -name '*.go' | xargs gofmt -s -l -w

# Do all code/CRD generation at once, with timestamp file to check out-of-date.
GEN_TIMESTAMP=.target/codegen
generate: $(GEN_TIMESTAMP)
$(GEN_TIMESTAMP): $(shell find apis -name '*.go')  $(OPERATOR_SDK) $(CONTROLLER_GEN) $(KUSTOMIZE) .target
	@$(CONTROLLER_GEN) object paths="./apis/..."
	@$(CONTROLLER_GEN) crd:crdVersions=v1 rbac:roleName=clusterlogging-operator paths="./..." output:crd:artifacts:config=config/crd/bases
	echo -e "package version\n\nvar Version = \"$(or $(CI_CONTAINER_VERSION),$(LOGGING_VERSION))\"" > version/version.go
	@$(MAKE) fmt
	@touch $@

.PHONY: regenerate
regenerate: $(OPERATOR_SDK) $(CONTROLLER_GEN) $(KUSTOMIZE)
	@rm -f $(GEN_TIMESTAMP)
	@$(MAKE) generate

.PHONY: deploy-image
deploy-image: .target/image
	hack/deploy-image.sh

.PHONY: deploy
deploy:  deploy-image deploy-elasticsearch-operator deploy-catalog install

.PHONY: install
install:
	IMAGE_CLUSTER_LOGGING_OPERATOR=image-registry.openshift-image-registry.svc:5000/openshift/origin-cluster-logging-operator:$(CURRENT_BRANCH) \
	$(MAKE) cluster-logging-operator-install

.PHONY: deploy-catalog
deploy-catalog:
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
	RELATED_IMAGE_VECTOR=$(IMAGE_LOGGING_VECTOR) \
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) \
	RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER=$(IMAGE_LOGFILEMETRICEXPORTER) \
	RELATED_IMAGE_LOGGING_CONSOLE_PLUGIN=$(IMAGE_LOGGING_CONSOLE_PLUGIN) \
	go test -cover -race ./test/helpers/...

.PHONY: test-functional-fluentd
test-functional-fluentd:
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) \
	RELATED_IMAGE_VECTOR=$(IMAGE_LOGGING_VECTOR) \
	RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER=$(IMAGE_LOGFILEMETRICEXPORTER) \
	go test --tags=fluentd -race ./test/functional/... -ginkgo.noColor -timeout=40m -ginkgo.slowSpecThreshold=45.0

.PHONY: test-functional-vector
test-functional-vector:
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) \
	RELATED_IMAGE_VECTOR=$(IMAGE_LOGGING_VECTOR) \
	RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER=$(IMAGE_LOGFILEMETRICEXPORTER) \
	go test --tags=vector -race \
		./test/functional/outputs/elasticsearch/... \
		./test/functional/outputs/kafka/... \
		./test/functional/outputs/cloudwatch/... \
		./test/functional/outputs/loki/... \
		./test/functional/outputs/http/... \
		./test/functional/normalization \
		-ginkgo.noColor -timeout=40m -ginkgo.slowSpecThreshold=45.0

.PHONY: test-forwarder-generator
test-forwarder-generator: bin/forwarder-generator
	WATCH_NAMESPACE=openshift-logging bin/forwarder-generator --file hack/logforwarder.yaml --collector=fluentd > /dev/null
	WATCH_NAMESPACE=openshift-logging bin/forwarder-generator --file hack/logforwarder.yaml --collector=vector > /dev/null


.PHONY: test-functional-benchmarker
test-functional-benchmarker: bin/functional-benchmarker
	@out=$$(bin/functional-benchmarker --artifact-dir=/tmp/benchmark-test 2>&1); if [ "$$?" != "0" ] ; then echo "$$out"; exit 1; fi

.PHONY: test-unit
test-unit: test-forwarder-generator
	RELATED_IMAGE_VECTOR=$(IMAGE_LOGGING_VECTOR) \
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) \
	RELATED_IMAGE_LOG_FILE_METRIC_EXPORTER=$(IMAGE_LOGFILEMETRICEXPORTER) \
	RELATED_IMAGE_LOGGING_CONSOLE_PLUGIN=$(IMAGE_LOGGING_CONSOLE_PLUGIN) \
	go test -coverprofile=test.cov -race ./apis/... ./internal/... `go list ./test/... | grep -Ev 'test/(e2e|functional|client|helpers)'`

.PHONY: coverage
coverage: test-unit
	go tool cover -html=test.cov -o $${ARTIFACTS_DIR:-.}/coverage.html

.PHONY: test-cluster
test-cluster:
	go test  -cover -race ./test/... -- -root=$(CURDIR)

OPENSHIFT_VERSIONS?="v4.11-v4.14"
# Generate bundle manifests and metadata, then validate generated files.
BUNDLE_VERSION?=$(VERSION)
CHANNEL=$(or $(filename $(OVERLAY)),stable)
BUNDLE_CHANNELS := --channels=$(CHANNEL),$(CHANNEL)-${LOGGING_VERSION}
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(CHANNEL)
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
    BUNDLE_GEN_FLAGS += --use-image-digests
endif

.PHONY: bundle
bundle: $(GEN_TIMESTAMP) $(KUSTOMIZE) $(find config -name *.yaml) ## Generate operator bundle.
	$(OPERATOR_SDK) generate kustomize manifests -q
	$(KUSTOMIZE) build $(or $(OVERLAY),config/manifests) | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	MANIFEST_VERSION=${LOGGING_VERSION} OPENSHIFT_VERSIONS=${OPENSHIFT_VERSIONS} CHANNELS=${CHANNELS} DEFAULT_CHANNEL=${DEFAULT_CHANNEL} hack/generate-bundle.sh
	$(OPERATOR_SDK) bundle validate ./bundle
	@touch $@

.PHONY: deploy-bundle
deploy-bundle: bundle bundle.Dockerfile
	podman build -t $(BUNDLE_TAG) -f bundle.Dockerfile .
	podman push --tls-verify=false ${BUNDLE_TAG}
	@echo "To run the bundle without this Makefile:"
	@echo "    oc create ns $(NAMESPACE)"
	@echo "    run bundle -n $(NAMESPACE) --install-mode OwnNamespace $(BUNDLE_TAG)"
	@touch $@

.PHONY: clean-bundle
clean-bundle:
	$(OPERATOR_SDK) cleanup --delete-all cluster-logging

WATCH_EVENTS=oc get events -A --watch-only& trap "kill %%" EXIT;
WAIT_FOR_OPERATOR=oc wait -n $(NAMESPACE) --for=condition=available deployment/cluster-logging-operator

.PHONY: namespace
namespace:
	echo '{"apiVersion": "v1", "kind": "Namespace","metadata":{"name":"$(NAMESPACE)","labels":{"openshift.io/cluster-monitoring":"true"}}}' | oc apply -f -

.PHONY: run-bundle
run-bundle: namespace $(OPERATOR_SDK) ## Run the overlay bundle image, assumes it has been pushed
	$(OPERATOR_SDK) cleanup --delete-all cluster-logging || true
	$(WATCH_EVENTS)	$(OPERATOR_SDK) run bundle -n $(NAMESPACE) --install-mode OwnNamespace $(BUNDLE_TAG); $(WAIT_FOR_OPERATOR)

.PHONY: apply
apply: namespace $(OPERATOR_SDK) ## Install kustomized resources directly to the cluster.
	$(OPERATOR_SDK) generate kustomize manifests -q
	$(WATCH_EVENTS) $(KUSTOMIZE) build $(or $(OVERLAY),config/manifests) | oc apply -f -; $(WAIT_FOR_OPERATOR)

.PHONY: test-e2e-olm
# NOTE: This is the CI e2e entry point.
test-e2e-olm: $(JUNITREPORT)
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) INCLUDES="$(E2E_TEST_INCLUDES)" CLF_INCLUDES="$(CLF_TEST_INCLUDES)" LOG_LEVEL=3 hack/test-e2e-olm.sh

.PHONY: test-e2e-local
test-e2e-local: $(JUNITREPORT) deploy-image
	LOG_LEVEL=3 \
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD) \
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
cluster-logging-catalog-deploy: $(shell find olm_deploy -type f)
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
