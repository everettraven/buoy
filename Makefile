include .bingo/Variables.mk

export GIT_VERSION       ?= $(shell git describe --tags --always --dirty)
export VERSION_PKG       ?= $(shell go list -m)/internal/cli
export GO_BUILD_ASMFLAGS ?= all=-trimpath=${PWD}
export GO_BUILD_LDFLAGS  ?= -s -w -X "$(VERSION_PKG).version=$(GIT_VERSION)"
export GO_BUILD_GCFLAGS  ?= all=-trimpath=${PWD}

export IMAGE_REPO ?= docker.io/bpalmer/buoy
export IMAGE_TAG  ?= devel

build:
	go build \
	-asmflags '$(GO_BUILD_ASMFLAGS)' \
	-ldflags '$(GO_BUILD_LDFLAGS)' \
	-gcflags '$(GO_BUILD_GCFLAGS)' \
	-o buoy main.go

export ENABLE_RELEASE_PIPELINE ?= false
export GORELEASER_ARGS         ?= --snapshot --clean
release: $(GORELEASER)
	$(GORELEASER) $(GORELEASER_ARGS)

GOLANGCI_LINT_ARGS ?=
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_ARGS)
unit:
	go test ./...

