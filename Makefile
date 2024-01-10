# Go version to require, run go version to see what version you are using;
GO_VERSION := "go1.21.6"
GO_VERSION ?= $(GO_VERSION)

TAG_VERSION := $(shell git describe --tags --always --dirty)

# Project Name & Version
INDEXER_MODULE_NAME := "indexer"
INDEXER_MODULE_VERSION := "alpha-0.0.1"

API_MODULE_NAME := "apiserver"
API_MODULE_VERSION := "alpha-0.0.1"

TIMESTAMP := $(shell date +%Y%m%d%H%M)

#INDEXER_OUTFILE_NAME=$(INDEXER_MODULE_NAME)-$(TAG_VERSION)
INDEXER_OUTFILE_NAME=$(INDEXER_MODULE_NAME)-$(INDEXER_MODULE_VERSION)
#API_OUTFILE_NAME=$(API_MODULE_NAME)-$(TAG_VERSION)
API_OUTFILE_NAME=$(API_MODULE_NAME)-$(API_MODULE_VERSION)

GO_BIN_DIR?=$(shell dirname `which go`)

.PHONY: all
all: fmt build

.PHONY: build
build: export GOBIN=$(CURDIR)/bin
build:
	go install -v $(CURDIR)/cmd/...

.PHONY: clean
clean:
	go clean
	rm -rf bin/*
	rm -rf tmp


.PHONY: lint
lint: fmt go-lint

.PHONY: go-lint
go-lint: export PATH:=$(CURDIR)/tools:$(PATH)
go-lint:
	golangci-lint --version
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...
	go vet ./...

.PHONY: check-go-version
check-go-version:
	@if ! go version | grep "$(GO_VERSION)" >/dev/null; then \
        printf "Wrong go version: "; \
        go version; \
        echo "Requires go version: $(GO_VERSION)"; \
        exit 2; \
    fi


TIMESTAMP := $(shell date +%Y%m%d%H%M)
OUTFILE_NAME=$(MODULE_NAME)-$(TIMESTAMP)


dev-indexer-build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/$(INDEXER_OUTFILE_NAME) -v $(CURDIR)/cmd/indexer/main.go

dev-indexer-build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/$(INDEXER_OUTFILE_NAME) -v $(CURDIR)/cmd/indexer/main.go

dev-apiserver-build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/$(API_OUTFILE_NAME) -v $(CURDIR)/cmd/jsonrpc/main.go

dev-apiserver-build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/$(API_OUTFILE_NAME) -v $(CURDIR)/cmd/jsonrpc/main.go
