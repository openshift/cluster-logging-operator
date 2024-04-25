# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.8. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Below generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for bingo variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(BINGO)
#	@echo "Running bingo"
#	@$(BINGO) <flags/args..>
#
BINGO := $(GOBIN)/bingo-v0.8.0
$(BINGO): $(BINGO_DIR)/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.8.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.8.0 "github.com/bwplotka/bingo"

CONTROLLER_GEN := $(GOBIN)/controller-gen-v0.9.2
$(CONTROLLER_GEN): $(BINGO_DIR)/controller-gen.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/controller-gen-v0.9.2"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=controller-gen.mod -o=$(GOBIN)/controller-gen-v0.9.2 "sigs.k8s.io/controller-tools/cmd/controller-gen"

GEN_CRD_API_REFERENCE_DOCS := $(GOBIN)/gen-crd-api-reference-docs-v0.0.5
$(GEN_CRD_API_REFERENCE_DOCS): $(BINGO_DIR)/gen-crd-api-reference-docs.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/gen-crd-api-reference-docs-v0.0.5"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=gen-crd-api-reference-docs.mod -o=$(GOBIN)/gen-crd-api-reference-docs-v0.0.5 "github.com/ViaQ/gen-crd-api-reference-docs"

GOLANGCI_LINT := $(GOBIN)/golangci-lint-v1.54.1
$(GOLANGCI_LINT): $(BINGO_DIR)/golangci-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/golangci-lint-v1.54.1"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=golangci-lint.mod -o=$(GOBIN)/golangci-lint-v1.54.1 "github.com/golangci/golangci-lint/cmd/golangci-lint"

JUNITREPORT := $(GOBIN)/junitreport-v0.0.0-20220531132329-92dee95b5da2
$(JUNITREPORT): $(BINGO_DIR)/junitreport.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/junitreport-v0.0.0-20220531132329-92dee95b5da2"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=junitreport.mod -o=$(GOBIN)/junitreport-v0.0.0-20220531132329-92dee95b5da2 "github.com/openshift/release/tools/junitreport"

KUSTOMIZE := $(GOBIN)/kustomize-v4.5.7
$(KUSTOMIZE): $(BINGO_DIR)/kustomize.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/kustomize-v4.5.7"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=kustomize.mod -o=$(GOBIN)/kustomize-v4.5.7 "sigs.k8s.io/kustomize/kustomize/v4"

OPERATOR_SDK := $(GOBIN)/operator-sdk-v1.34.1
$(OPERATOR_SDK): $(BINGO_DIR)/operator-sdk.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/operator-sdk-v1.34.1"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=operator-sdk.mod -o=$(GOBIN)/operator-sdk-v1.34.1 "github.com/operator-framework/operator-sdk/cmd/operator-sdk"

OPM := $(GOBIN)/opm-v1.23.0
$(OPM): $(BINGO_DIR)/opm.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/opm-v1.23.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=opm.mod -o=$(GOBIN)/opm-v1.23.0 "github.com/operator-framework/operator-registry/cmd/opm"

