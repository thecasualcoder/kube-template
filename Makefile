.PHONY: help
help: ## Prints help (only for targets with comments)
	@grep -E '^[a-zA-Z._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

APP=kube-template
SRC_PACKAGES=$(shell go list ./... | grep -v "vendor")
VERSION?=dev
GOLINT:=$(shell command -v golint 2> /dev/null)
APP_EXECUTABLE="./out/$(APP)"
RICHGO=$(shell command -v richgo 2> /dev/null)
GOMETA_LINT=$(shell command -v golangci-lint 2> /dev/null)
GO111MODULE=on
SHELL=/bin/bash -o pipefail

ifeq ($(GOMETA_LINT),)
	GOMETA_LINT=$(shell command -v $(PWD)/bin/golangci-lint 2> /dev/null)
endif

ifeq ($(RICHGO),)
	GO_BINARY=go
else
	GO_BINARY=richgo
endif

ifeq ($(BUILD),)
	BUILD=dev
endif

ifdef CI_COMMIT_SHORT_SHA
	BUILD=$(CI_COMMIT_SHORT_SHA)
endif

all: setup build

ensure-build-dir:
	mkdir -p out

build-deps: ## Install dependencies
	go mod tidy

update-deps: ## Update dependencies
	go get -u

compile: ensure-build-dir ## Compile
	$(GO_BINARY) build -ldflags "-X main.version=$(VERSION)" -o $(APP_EXECUTABLE) ./main.go

compile-linux: ensure-build-dir ## Compile for linux explicitly
	GOOS=linux GOARCH=amd64 $(GO_BINARY) build -ldflags "-X main.version=$(VERSION)" -o $(APP_EXECUTABLE) ./main.go

build: build-deps fmt vet lint test compile ## Build the application

compress: compile ## Compress the binary
	upx $(APP_EXECUTABLE)

fmt:
	$(GO_BINARY) fmt $(SRC_PACKAGES)

vet:
	$(GO_BINARY) vet $(SRC_PACKAGES)

setup-golangci-lint:
ifeq ($(GOMETA_LINT),)
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
endif

setup: setup-golangci-lint ensure-build-dir ## Setup environment
ifeq ($(GOLINT),)
	GO111MODULE=off $(GO_BINARY) get -u golang.org/x/lint/golint
endif
ifeq ($(RICHGO),)
	GO111MODULE=off $(GO_BINARY) get -u github.com/kyoh86/richgo
endif

lint: setup-golangci-lint
	$(GOMETA_LINT) run

test: ensure-build-dir ## Run tests
	ENVIRONMENT=test $(GO_BINARY) test -v ./...

test-cover-html: ## Run tests with coverage
	mkdir -p ./out
	@echo "mode: count" > coverage-all.out
	$(foreach pkg, $(SRC_PACKAGES),\
	ENVIRONMENT=test $(GO_BINARY) test -coverprofile=coverage.out -covermode=count $(pkg);\
	tail -n +2 coverage.out >> coverage-all.out;)
	$(GO_BINARY) tool cover -html=coverage-all.out -o out/coverage.html