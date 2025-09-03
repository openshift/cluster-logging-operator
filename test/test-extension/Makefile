all: update
.PHONY: all

OUT_DIR=bin
BINARY=$(OUT_DIR)/logging-test

GO_LD_FLAGS := -s -w
ifeq ($(shell go env GOOS),darwin)
    GO_LD_FLAGS += -extldflags='-ld_classic'
endif

GOBIN := $(shell go env GOPATH)/bin
GO_BINDATA := $(GOBIN)/go-bindata

.PHONY: update
update: bindata build

.PHONY: build
build:
	mkdir -p "$(OUT_DIR)"
	export CGO_ENABLED=0 && \
	export GO111MODULE=on && \
	export GOWORK=off && \
	go build -ldflags="$(GO_LD_FLAGS)" -mod=mod -o "$(BINARY)" ./cmd/...


.PHONY: bindata
bindata: $(GO_BINDATA) generated/bindata.go

$(GO_BINDATA):
	go install github.com/go-bindata/go-bindata/go-bindata@latest

generated/bindata.go: $(shell find testdata -type f)
	mkdir -p $(@D)
	$(GO_BINDATA) -nocompress -nometadata \
		-pkg testdata -o $@ \
		testdata/...
	gofmt -s -w $@

.PHONY: clean
clean:
	@echo "Cleaning up artifacts..."
	rm -rf $(OUT_DIR)
	# Specifically remove the generated bindata to force a sync next build
	rm -f generated/bindata.go
