# Enable Go modules
export GO111MODULE := on

# Versioning variables
VERSION := $(shell git describe --abbrev=0 --tags 2>/dev/null || true)
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_TIME := $(shell date -u +%FT%TZ)
BUILD_AUTHOR := $(shell git config user.email)

# Package name
PKG_NAME := $(shell basename `pwd`)

# Setup ldflags for build interpolation
LDFLAGS := -ldflags "-w -s \
-X main.Version=$(VERSION) \
-X main.BuildCommit=$(BUILD_COMMIT) \
-X main.BuildBranch=$(BUILD_BRANCH) \
-X main.BuildTime=$(BUILD_TIME) \
-X main.BuildAuthor=$(BUILD_AUTHOR)"

# Make options
.SILENT: ;
.DEFAULT_GOAL := help
.PHONY: setup build lint test cover run ci clean help

setup: ## Get lint and build depedencies
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go mod download

build: ## Compile code and build binary file
	go build $(LDFLAGS) -o ./bin/$(PKG_NAME) ./...

lint: ## Run linters
	./bin/golangci-lint run --tests=false --enable-all ./...

test: ## Run test suite
	go test -failfast -race -cover -covermode=atomic -coverprofile=cover.out -v ./...

cover: ## Open test coverage report
	go tool cover -html=cover.out

run: ## Run binary
	./bin/$(PKG_NAME)

ci: setup build test lint ## Run all code checks and tests

clean: ## Remove object and cache files
	go clean

help:  ## Display this help
	awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
	{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

