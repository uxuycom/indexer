PKG := github.com/uxuycom/indexer

GO_BIN := ${GOPATH}/bin
GOACC_BIN := $(GO_BIN)/go-acc
GOIMPORTS_BIN := $(GO_BIN)/goimports
MIGRATE_BIN := $(GO_BIN)/migrate

COMMIT := $(shell git describe --tags --dirty)

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v
GOTEST := GO111MODULE=on go test -v
GOMOD := GO111MODULE=on go mod

GOLIST := go list -deps $(PKG)/... | grep '$(PKG)'
GOLIST_COVER := $$(go list -deps $(PKG)/... | grep '$(PKG)')

RM := rm -f
CP := cp
MAKE := make
XARGS := xargs -L 1
UNAME_S := $(shell uname -s)

DEV_TAGS := $(if ${tags},$(DEV_TAGS) ${tags},$(DEV_TAGS))

make_ldflags = $(1) -X $(PKG).Commit=$(COMMIT)

DEV_GCFLAGS := -gcflags "all=-N -l"
DEV_LDFLAGS := -ldflags "$(call make_ldflags)"

# For the release, we want to remove the symbol table and debug information (-s)
# and omit the DWARF symbol table (-w). Also we clear the build ID.
RELEASE_LDFLAGS := $(call make_ldflags, -s -w -buildid=)

# Go version to require, run go version to see what version you are using;
GO_VERSION := "go1.21.6"
GO_VERSION ?= $(GO_VERSION)

GO_BIN_DIR?=$(shell dirname `which go`)

.PHONY: all
all: fmt build

.PHONY: build
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

build:
	@$(call print, "Building debug uxuy indexer and apiserver.")
	$(GOBUILD) -tags="$(DEV_TAGS)" -o bin/$(INDEXER_OUTFILE_NAME) $(DEV_GCFLAGS) $(DEV_LDFLAGS) $(PKG)/cmd/indexer
	$(GOBUILD) -tags="$(DEV_TAGS)" -o bin/$(API_OUTFILE_NAME) $(DEV_GCFLAGS) $(DEV_LDFLAGS) $(PKG)/cmd/jsonrpc

install: install-indexer install-jsonrpc

install-indexer:
	@$(call print, "Installing uxuy indexer")
	install -C ./bin/indexer /usr/local/bin/indexer

install-jsonrpc:
	@$(call print, "Installing uxuy jsonrpc api server")
	install -C ./bin/jsonrpc /usr/local/bin/apiserver

release-install:
	@$(call print, "Installing release uxuy indexer and apiserver.")
	env CGO_ENABLED=0 $(GOINSTALL) -v -trimpath -ldflags="$(RELEASE_LDFLAGS)" -tags="$(RELEASE_TAGS)" $(PKG)/cmd/indexer
	env CGO_ENABLED=0 $(GOINSTALL) -v -trimpath -ldflags="$(RELEASE_LDFLAGS)" -tags="$(RELEASE_TAGS)" $(PKG)/cmd/jsonrpc
