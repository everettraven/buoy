include .bingo/Variables.mk

export GIT_VERSION       ?= $(shell git describe --tags --always --dirty)
export VERSION_PKG       ?= $(shell go list -m)/internal/cli
export GO_BUILD_ASMFLAGS ?= all=-trimpath=${PWD}
export GO_BUILD_GCFLAGS  ?= all=-trimpath=${PWD}

export IMAGE_REPO ?= docker.io/bpalmer/buoy
export IMAGE_TAG  ?= devel

build:
	go build \
	-asmflags '$(GO_BUILD_ASMFLAGS)' \
	-gcflags '$(GO_BUILD_GCFLAGS)' \
	-o buoy .

export ENABLE_RELEASE_PIPELINE ?= false
export GORELEASER_ARGS         ?= --snapshot --clean
release: $(GORELEASER)
	$(GORELEASER) $(GORELEASER_ARGS)

GOLANGCI_LINT_ARGS ?=
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_ARGS)
unit:
	go test ./... -coverprofile=cover.out -covermode=atomic

