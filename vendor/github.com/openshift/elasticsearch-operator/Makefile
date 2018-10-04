CURPATH=$(PWD)
TARGET_DIR=$(CURPATH)/_output

GOBUILD=go build
GOPATH=$(TARGET_DIR):$(TARGET_DIR)/vendor:$(CURPATH)/cmd

DOCKER_OPTS=

APP_NAME=elasticsearch-operator
APP_REPO=github.com/ViaQ/$(APP_NAME)
DOCKER_TAG=github.com/openshift/origin-$(APP_NAME)
TARGET=$(TARGET_DIR)/bin/$(APP_NAME)
MAIN_PKG=cmd/$(APP_NAME)/main.go

# These will be provided to the target
#VERSION := 1.0.0
#BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
#LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
#LDFLAGS=-ldflags "-X=main.Build=$(BUILD)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

#.PHONY: all build clean install uninstall fmt simplify check run
.PHONY: all build clean fmt simplify run

all: build #check install

build: $(SRC)
	@mkdir -p $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/pkg $(TARGET_DIR)/src/$(APP_REPO)
	@cp -ru $(CURPATH)/vendor/* $(TARGET_DIR)/src
	@GOPATH=$(GOPATH) $(GOBUILD) $(LDFLAGS) -o $(TARGET) $(MAIN_PKG)

clean:
	@rm -rf $(TARGET_DIR)

image:
	@docker build -t $(DOCKER_TAG) . $(DOCKER_OPTS)

#install:
#	@go install $(LDFLAGS)

#uninstall: clean
#	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)
