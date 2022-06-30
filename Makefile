# Makefile for cluster logging operator.

# NOTES on Makefile design
#
# Developers must start with `make init REGISTRY=quay.io/myname`.
# This generates a config/overlays/develop overlay with the REGISTRY
# and some default settings that can be edited (e.g. to use alternate
# operand images)
#
# All non-test targets have complete dependency rules.
# Any target can be built directly from a clean environment,
# make builds pre-requisites as needed.
#
# Executable targets depend on the apis: target to generate code, but
# are independent of OVERLAY, config, image settings etc. In other
# words, you can build executables without knowing how they will be
# deployed in the cluster. The operator executable takes runtime
# values (image names, namespace) from env. vars.
#
# The config/overlays/$(OVERLAY)/kustomization.yaml defines how to
# modify config resources to deploy in release or develop mode. The
# Makefile sets key variables (NAMESPACE, IMG, etc.)  that must match
# the deployment values from the $(OVERLAY) files.
# Editing overlay files will rebuild with new values as needed.
#
# The ./config directory contains YAML to deploy the operator, the
# $(OVERLAY)/kustomization.yaml modifies the YAML for release or
# developer deployments. Some files under config (CRD, CSV, some RBAC
# files) are generated/updated from *_type.go code. The config: target
# updates everything under ./config.
#
# The operator can be deployed directly from resources in ./config
# using kustomize by `make deploy`.
#
# Release builds and CI get special treatment.
#
# `make OVERLAY=release` updates the ./bundle and bundle.Dockerfile
# used by CI to build images.  Beware it will push images to the
# openshift release registry if you have permissions!
#
# You shouldn't need to explicitly use `make OVERLAY=release`, updates
# to the develop/bundle will automatically update the release
# ./bundle as well.
#
# CI uses the checked-in Dockerfile* and bundle.Dockerfile directly
# (not via the Makefile) to build images. It uses the Makefile to run
# lint, test-functional and test-e2e-olm targets.
#
# The first version will download $(SOME_TOOL) if not present and then re build my-thing,
# even if real-dependencies haven't changed. In particular this will always happen in CI.
#
# The second version will only download $(SOME_TOOL) IF we really need to update my-thing.
# Since everything should be up to date in CI, nothing happens.

# Display in-line help for this Makefile.
#
# Comments starting with '##' on the same line as the target are target descriptions.
# Stand-alone starting with '##@' are section headers.
.PHONY: help
.PHONY: help
help: ## Display this help
	@echo 'MAKE TARGETS'
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 } /^##@/ { printf "%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo 'QUICK START'
	@echo '  make init REGISTRY=quay.io/YOUR_QUAY_ACCOUNT_NAME'
	@echo '  make deploy'
	@echo 'RUNNING INDIVIDUAL TESTS'
	@echo '  env $$(make -s env) go test -v ./test/functional/outputs # Use go test directly'
	@echo '  make go-test VERBOSE=1 PKG=./test/functional/outputs # Run via make'

.PHONY: all
all: check clean-cluster test-functional olm-deploy deploy-elastic test-e2e ## Build and test everything. `make -k all` to keep going if a test fails.

.PHONY: check
check: compile-tests bundle bin/forwarder-generator bin/cluster-logging-operator bin/functional-benchmarker lint test-unit  ## Checks that don't need a cluster: generate, compile, lint, unit test.
.PHONY: pre-commit
pre-commit: check

.PHONY: compile-tests
compile-tests: apis
	go test ./test/... -exec true > /dev/null # Build all tests but don't run

# Define targets and variables for executable tools.
include .bingo/Variables.mk

# Disable vendor mode in case we are in a directory tree with vendor/ directories higher up.
export GOFLAGS=-mod=mod
GOTEST_FLAGS=-ginkgo.noColor

# VERBOSE gives verbose output from tools and tests, and watches test events for cluster tests.
# Always be verbose for CI builds.
ifneq (,$(or $(VERBOSE),$(CI)))
export VERBOSE=1 # For sub-make
GOTEST_FLAGS:=$(GOTEST_FLAGS) -v -ginkgo.v
RUN_TEST=$(TEST_ENV) $(WATCH_EVENTS)
VERBOSE_FLAG=--verbose
else
PUSH_FLAGS=--quiet
RUN_TEST=$(TEST_ENV)
endif

# FAST skips coverage and race detection in tests.
ifeq (,$(FAST))
GOTEST_FLAGS:=-cover -race $(GOTEST_FLAGS)
export FAST			# For sub-make
endif

# List of supported vX.Y versions of Openshift, comma separated, no spaces.
OPENSHIFT_VERSIONS?=v4.9

# OVERLAY kustomizations in config/overlays/$(OVERLAY)/kustomization.yaml
# "develop" builds develop/bundle using config/overlays/develop/ kustomization.
# "release" builds root ./bundle using config/overlays/release kustomization.
OVERLAY?=$(if $(CI),release,develop)
OVERLAY_DIR=config/overlays/$(OVERLAY)
OVERLAY_FILES=$(OVERLAY_DIR) $(shell find $(OVERLAY_DIR) 2>/dev/null)
ifeq ($(OVERLAY),release)
# For the release overlay, ./bundle and other output goes in the root dir.
OUT_DIR=.
else
# For a developer overlay, use a separate OUT_DIR for generated bundle etc.
OUT_DIR=$(OVERLAY)
CLEAN:=$(OUT_DIR)
endif

VERSION_GO=$(OUT_DIR)/version/version.go

# PODMAN command. `PODMAN=docker` may work, but is not well tested.
PODMAN?=podman

# Don't ignore failures in shell pipelines.
SHELL=/bin/bash -o pipefail

# Get values from overlay files
IMG_NAME=$(shell awk '/newName:/{print $$2}' $(OVERLAY_DIR)/kustomization.yaml)
IMG_TAG=$(shell awk '/newTag:/{print $$2}' $(OVERLAY_DIR)/kustomization.yaml)

# If CI env. vars are set, use them in preference to kustomization values.
NAMESPACE?=$(shell awk '/namespace:/{print $$2}' $(OVERLAY_DIR)/kustomization.yaml)
IMG=$(or $(IMAGE_CLUSTER_LOGGING_OPERATOR),$(IMG_NAME):$(IMG_TAG))
BUNDLE_IMG=$(or $(IMAGE_CLUSTER_LOGGING_OPERATOR_REGISTRY),$(IMG_NAME)-bundle:$(IMG_TAG))

# If image tag is a semver, use it as VERSION, otherwise use 0.0.0-git-branch for development versions.
VERSION=$(or $(CI_CONTAINER_VERSION),$(shell echo $(IMG_TAG) | grep -o '^[0-9]\+\.[0-9]\+\.[0-9]\+'),0.0.0-$(IMG_TAG))
VERSION_XY=$(shell echo "$(VERSION)" | grep -o '^[0-9]\+\.[0-9]\+')

.PHONY: force

################ Rules

##@ Build the operator

# init copies the release workspace and makes some default adjustments.
# Developer can edit the overlay for further kustomization.
.PHONY: init
init: ## (re-)initialize workspace for development. REGISTRY must be set, e.g. 'REGISTRY=quay.io/myquayname'
	rm -rf $(OVERLAY_DIR)
	@$(MAKE) --no-print-directory $(OVERLAY_DIR)

$(OVERLAY_DIR):
ifneq ($(OVERLAY),release)     # Never initialize the release overlay, it is checked in.
ifeq (,$(REGISTRY))
	@echo "$(OVERLAY_DIR) not found and REGISTRY not set."
	@echo "Initialize you workspace with a REGISTRY, for example:"
	@echo "  make init REGISTRY=quay.io/myquayname"
	@exit 1
else
	cp -r config/overlays/release $(OVERLAY_DIR)
	sed -e 's|newName:.*|newName: $(REGISTRY)/origin-cluster-logging-operator|'	\
	     -e	's|newTag:.*|newTag: $(or $(shell git branch --show-current),dev)|'	\
	     -e 's|../../manifests|../../default|'					\
	     -i $(OVERLAY_DIR)/kustomization.yaml
	sed -e 's|imagePullPolicy:.*|imagePullPolicy: Always|'				\
	     -i $(OVERLAY_DIR)/deployment_patch.yaml
	@echo == Initialized $(OVERLAY_DIR)
endif
endif

.PHONY: tools
tools: $(BINGO) $(GOLANGCI_LINT) $(JUNITREPORT) $(OPERATOR_SDK) $(KUSTOMIZE) $(CONTROLLER_GEN)

# Generate Go code for API types.
apis: $(CONTROLLER_GEN) $(shell find apis -name '*.go' | grep -v zz_generated) $(VERSION_GO)
	$(CONTROLLER_GEN) object paths="./apis/..."
	@touch $@

$(VERSION_GO): $(OVERLAY_FILES)
	@mkdir -p $(dir $@)
	@echo 'package version; var Version = "$(VERSION)"' > $(VERSION_GO)
ifneq ($(OVERLAY),release)	#Update the release bundle also.
	$(MAKE) -s OVERLAY=release ./version/version.go
endif

# Generate config manifests from API types.
config: $(CONTROLLER_GEN) apis $(shell find config -type f 2> /dev/null || true)
	$(CONTROLLER_GEN) object paths="./apis/..."
	$(CONTROLLER_GEN) crd:crdVersions=v1 output:crd:artifacts:config=config/crd/bases paths="./apis/..."
	$(CONTROLLER_GEN) rbac:roleName=clusterlogging-operator paths="./apis/..."
	@touch $@

.PHONY: build
build: bin/cluster-logging-operator

# Note: Go has built-in build caching, so always run `go build`.
# It will do a better job than using source dependencies to decide if we need to build.
bin/functional-benchmarker: apis force
	go build -o $@ ./internal/cmd/functional-benchmarker

bin/forwarder-generator: apis force
	go build -o $@ ./internal/cmd/forwarder-generator

bin/cluster-logging-operator: apis $(VERSION_GO) force
	go build -o $@

.PHONY: openshift-client
openshift-client:
	@type -p oc > /dev/null || bash hack/get-openshift-client.sh

# lint is a CI target, called before any others to verify valid sources.
.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --color=never $(if $(CI),,--fix)

.PHONY: clean
clean: ## Delete build artifacts and temporary files, keep downloaded tools and overlay.
	rm -f bin/cluster-logging-operator bin/forwarder-generator bin/functional-benchmarker
	rm -rf $(CLEAN) tmp _output .cache bin Dockerfile.local .make*  benchmark-[0-9]* bundle-[0-9][0-9]*
	find -name .kube -type d | xargs rm -rf
	go clean -cache -testcache ./...
	rm -rf /tmp/ocp_clo

.PHONY: spotless
spotless: clean ## Delete everything including tools, overlay and bundle files.
	rm -rf bin/* $(OVERLAY_DIR) $(OUT_DIR)

Dockerfile: hack/generate-dockerfile-from-midstream Dockerfile.in origin-meta.yaml
	$< > Dockerfile

.PHONY: image
image: .make-$(OVERLAY)-image  ## Build the operator image
.make-$(OVERLAY)-image: config Dockerfile main.go go.mod go.sum $(shell find must-gather version scripts files .bingo apis controllers internal -type f) ## Build the operator image.
	mkdir -p $(dir $@)
	$(PODMAN) build -t $(IMG) -f Dockerfile
	@touch $@

.PHONY: image
image-push: .make-$(OVERLAY)-image-push ## Push the operator image
.make-$(OVERLAY)-image-push: .make-$(OVERLAY)-image
	$(PODMAN) push $(PUSH_FLAGS) $(IMG)
	@touch $@

##@ Deploy directly (without OLM)

.PHONY: namespace
namespace: config
	oc create namespace $(NAMESPACE) --dry-run=client -o yaml | oc apply --wait -f -
	oc label --overwrite=true namespace/$(NAMESPACE) openshift.io/cluster-monitoring=true

.PHONY: deploy
deploy: $(KUSTOMIZE) config image-push namespace  ## Deploy kustomized resources.
	@echo == $@
	$(KUSTOMIZE) build $(OVERLAY_DIR) | oc apply --wait -f -
	@hack/oc-wait.sh --for=condition=available $(NAMESPACE) deployment/cluster-logging-operator

.PHONY: deploy-image
deploy-image: deploy		# Alias for compatibility with old Makefile

.PHONY: deploy-yaml
deploy-yaml: $(KUSTOMIZE) config
	$(KUSTOMIZE) build $(OVERLAY_DIR)

.PHONY: undeploy
undeploy: $(KUSTOMIZE) config ## Delete deployed resources from cluster.
	@echo == $@
	@oc delete --ignore-not-found --wait ns/$(NAMESPACE)
	@$(KUSTOMIZE) build $(OVERLAY_DIR) | oc delete --ignore-not-found --wait -f - 2> /dev/null || true
	@oc get -n openshift imagestreams -o name | grep cluster-logging | xargs -r oc delete --ignore-not-found || true

##@ Deploy via OLM (Operater Lifecycle Manager)

# Release bundles go in ./bundle, developer builds in $(OUT_DIR)/bundle.
ifneq ($(OVERLAY),release)
.PHONY: bundle
bundle: $(OUT_DIR)/bundle
endif

CHANNELS?="stable,stable-$(VERSION_XY)"
DEFAULT_CHANNEL?=stable
# clusterlogging or cluster-logging?
PACKAGE?=clusterlogging

BUNDLE_ARGS?=-q --overwrite --package=$(PACKAGE) --version=$(VERSION) --channels=$(CHANNELS) --default-channel=$(DEFAULT_CHANNEL) $(and $(USE_IMAGE_DIGESTS),--use-image-digests)

$(OUT_DIR)/bundle: $(KUSTOMIZE) $(OPERATOR_SDK) config
	@echo == $@
	@rm -rf $(OUT_DIR)/bundle
	@mkdir -p $(OUT_DIR)/bundle
	$(OPERATOR_SDK) $(VERBOSE_FLAG) generate kustomize manifests
	$(KUSTOMIZE) build $(OVERLAY_DIR) | { cd $(OUT_DIR) && $(OPERATOR_SDK) generate bundle $(BUNDLE_ARGS) ; }
	$(OPERATOR_SDK) bundle validate $(VERBOSE_FLAG) $@
	OPENSHIFT_VERSIONS=$(OPENSHIFT_VERSIONS) hack/edit-bundle.sh $(OUT_DIR)
ifneq ($(OVERLAY),release)	# Update the release bundle automatically when dev bundle is updated.
	@echo "== Update release bundle"
	$(MAKE) -s OVERLAY=release bundle
endif
	@touch config # op-sdk touches files in the config dir, mark it up to date.
	@touch $@

.PHONY: bundle-image
bundle-image: .make-$(OVERLAY)-$@ ## Build the OLM bundle image.
.make-$(OVERLAY)-bundle-image: $(OUT_DIR)/bundle
	cd $(OUT_DIR); $(PODMAN) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .
	@touch $@

.PHONY: bundle-push
bundle-push: .make-$(OVERLAY)-bundle-push ## Push the OLM bundle image.
.make-$(OVERLAY)-bundle-push: .make-$(OVERLAY)-bundle-image
	$(PODMAN) push $(PUSH_FLAGS) $(BUNDLE_IMG)
	@touch $@

.PHONY: olm-deploy
olm-deploy: $(OPERATOR_SDK) image-push bundle-push undeploy namespace ## Deploy the OLM bundle.
	@echo == $@
	@$(WATCH_EVENTS) $(OPERATOR_SDK) $(VERBOSE_FLAG) run bundle -n $(NAMESPACE) --install-mode OwnNamespace $(BUNDLE_IMG)

##@ Testing

DEPLOY_ENV=$(shell awk '/name:/ {NAME = $$NF} /value: / { if (NAME == "$(1)") { print $$NF; exit 0; }  }' $(OVERLAY_DIR)/deployment_patch.yaml)

# Set variables from kubernetes deployment, unless already set.
# Allows CI to change the images.
IMAGE_LOGGING_VECTOR?=$(call DEPLOY_ENV,RELATED_IMAGE_VECTOR)
IMAGE_LOGGING_FLUENTD?=$(call DEPLOY_ENV,RELATED_IMAGE_FLUENTD)
IMAGE_LOGGING_LOGFILEMETRICEXPORTER?=$(call DEPLOY_ENV,RELATED_IMAGE_LOGFILEMETRICEXPORTER)

# Test environment.
TEST_ENV=										\
	RELATED_IMAGE_VECTOR=$(IMAGE_LOGGING_VECTOR)					\
	RELATED_IMAGE_FLUENTD=$(IMAGE_LOGGING_FLUENTD)					\
	RELATED_IMAGE_LOGFILEMETRICEXPORTER=$(IMAGE_LOGGING_LOGFILEMETRICEXPORTER)	\
	OPERATOR_NAME=$(call DEPLOY_ENV,OPERATOR_NAME)

# Print k8s events while running a command.
WATCH_EVENTS=hack/test-watch.sh $(NAMESPACE) --

.PHONY: env
env: config ## Print environment for running tests. Use 'make -s' to suppress make output.
	@for E in  $(TEST_ENV); do echo $$E; done

.PHONY: run
run: $(KUSTOMIZE) config namespace ## Run the operator outside the cluster for debugging
	oc delete --ignore-not-found -n $(NAMESPACE) deployment cluster-logging-operator # Stop the in-cluster operator
	$(KUSTOMIZE) build config/crd | oc apply -f -
	@mkdir -p $(CURDIR)/tmp
	$(RUN_TEST) \
	KUBERNETES_CONFIG=$(KUBECONFIG) \
	WORKING_DIR=$(CURDIR)/tmp \
	go run main.go

.PHONY: test-unit
test-unit: ## Run unit tests, no cluster required.
	@echo == $@
	@$(MAKE) -k test-unit-go test-unit-vector test-unit-fluentd # keep going on failure.
.PHONY: test-unit-go
test-unit-go:
	 go test ./internal/... $(shell go list ./test/... | grep -Ev 'test/(e2e|functional|client|helpers)') $(GOTEST_FLAGS)
.PHONY: test-unit-fluentd
test-unit-fluentd: bin/forwarder-generator
	bin/forwarder-generator --file hack/logforwarder.yaml --collector=fluentd > /dev/null
.PHONY: test-unit-vector
test-unit-vector: bin/forwarder-generator
	bin/forwarder-generator --file hack/logforwarder.yaml --collector=vector > /dev/null

.PHONY: test-functional
test-functional: ## Run functional tests, requires a cluster but not a deployment.
	$(MAKE) -k test-functional-fluentd test-functional-vector test-functional-other test-functional-benchmarker
.PHONY: test-functional-fluentd
test-functional-fluentd:	# FIXME Skipping most vector functional tests!
	@echo == $@
	@$(RUN_TEST) COLLECTOR=fluentd go test ./test/functional/outputs/elasticsearch/... $(GOTEST_FLAGS) -tags=fluentd -ginkgo.slowSpecThreshold=45
.PHONY: test-functional-vector
test-functional-vector:
	@echo == $@
	@$(RUN_TEST) COLLECTOR=vector go test ./test/functional/... $(GOTEST_FLAGS) -tags=vector -ginkgo.slowSpecThreshold=45
.PHONY: test-functional-other
test-functional-other:
	@echo == $@
	@$(RUN_TEST) go test ./test/client ./test/framework/... ./test/helpers/... $(GOTEST_FLAGS)
.PHONY: test-functional-benchmarker
test-functional-benchmarker: bin/functional-benchmarker
	@echo == $@
	bin/functional-benchmarker --artifact-dir=$$(mktemp -d -u) > /dev/null

.PHONY: deploy-elastic
deploy-elastic: $(KUSTOMIZE) config ## Deploy elasticsearch for e2e tests.
	@echo == $@
	$(KUSTOMIZE) build config/elasticsearch | oc apply -f -
	@hack/oc-wait.sh --for=condition=available openshift-operators-redhat deployment/elasticsearch-operator

.PHONY: undeploy-elastic
undeploy-elastic: $(KUSTOMIZE) config ## Undeploy elasticsearch.
	@echo == $@
	@$(KUSTOMIZE) build config/elasticsearch  | oc delete --ignore-not-found -f - 2> /dev/null || true
	@oc delete --ignore-not-found -n openshift-operators-redhat  catalogsource/elasticsearch-catalog

test-e2e: ## Run e2e tests, requires a cluster with operators deployed (make deploy deploy-elastic)
	@$(RUN_TEST) go test ./test/e2e/... $(GOTEST_FLAGS) -ginkgo.slowSpecThreshold=60

# NOTE: This is the CI e2e entry point.
# Use the old test and deploy scripts until CI is updated to use bundles.
test-e2e-olm: clean-cluster ## Deploy and run e2e tests from scratch for CI.
	@echo == $@
	$(RUN_TEST) LOG_LEVEL=${LOG_LEVEL:-3} hack/test-e2e-olm.sh

.PHONY: go-test
go-test: ## Test selected packages, specify packages like this TEST_PKG="./test/x ./test/y ..."
	$(RUN_TEST) go test $(PKG) $(GOTEST_FLAGS)

.PHONY: delete-tests
delete-tests: ## Delete left-over test resources in the cluster
	@echo == $@
	for KIND in Namespace ClusterRole ClusterRoleBinding; do \
	  oc delete -l test-client --ignore-not-found --wait=false $$KIND ; \
	done
	oc get ns -o name | grep 'clo-test-[0-9]' | xargs -r oc delete --ignore-not-found --wait=false || :

.PHONY: clean-cluster
clean-cluster: undeploy undeploy-elastic delete-tests ## Clean cluster: undeploy operators, delete tests.

test-e2e-clo-metric:
	test/e2e/telemetry/clometrics_test.sh

.PHONY: test-svt
test-svt: deploy deploy-elastic
	hack/svt/test-svt.sh

REPLICAS?=0
.PHONY: scale-cvo
scale-cvo:
	oc -n openshift-cluster-version scale deployment/cluster-version-operator --replicas=$(REPLICAS)

.PHONY: scale-olm
scale-olm:
	oc -n openshift-operator-lifecycle-manager scale deployment/olm-operator --replicas=$(REPLICAS)

# CI calls ci-check first.
.PHONY: ci-check
ci-check: check
	@echo
	@git diff-index --name-status --exit-code HEAD || { \
		echo -e '\nerror: files changed during "make check", not up-to-date\n' ; \
		exit 1 ; \
	}
