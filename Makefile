# Enable Go modules
export GO111MODULE := on

# Versioning variables
VERSION := $(shell git describe --abbrev=0 --tags 2>/dev/null || true)
BUILD_COMMIT := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u +%FT%TZ)

# Package name
PKG_NAME := $(shell basename `pwd`)

# Setup ldflags for build interpolation
LDFLAGS := -ldflags "-w -s \
-X main.version=$(VERSION) \
-X main.commit=$(BUILD_COMMIT) \
-X main.date=$(BUILD_TIME)"

# Make options
.SILENT: ;
.DEFAULT_GOAL := help
.PHONY: setup build lint test cover bench benchcmp profile pprof-cpu pprof-mem run ci clean help

setup: ## Get lint, test and build depedencies
	[ -d ./bin ] || mkdir ./bin
	[ -d ./_tmp ] || mkdir ./_tmp
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	cd /tmp && go get golang.org/x/tools/cmd/benchcmp
	go mod download

build: ## Compile code and build binary file
	go build $(LDFLAGS) -o ./bin/$(PKG_NAME) ./...

lint: ## Run linters
	./bin/golangci-lint run --tests=false --enable-all --disable=gochecknoglobals,gochecknoinits ./...

test: ## Run test suite
	go test -failfast -race -cover -covermode=atomic -coverprofile=./_tmp/cover.out -v ./...

cover: ## Open test coverage report
	go tool cover -html=./_tmp/cover.out

bench: ## Run benchmarks
	if [ -f ./_tmp/bench-new.txt ]; then mv ./_tmp/bench-new.txt ./_tmp/bench-old.txt; fi;
	go test -bench=. -benchmem -benchtime=5s ./... | tee ./_tmp/bench-new.txt

benchcmp: bench ## Run benchmarks and compare with previous results
	benchcmp ./_tmp/bench-old.txt ./_tmp/bench-new.txt

profile: ## Profile CPU and memory usage
	go test -bench=. -o=./_tmp/profile.test -cpuprofile=./_tmp/cpuprofile.out -memprofile=./_tmp/memprofile.out .

pprof-cpu: profile ## Analyze CPU profile
	go tool pprof -http=: ./_tmp/cpuprofile.out

pprof-mem: profile ## Analyze memory profile
	go tool pprof -http=: ./_tmp/memprofile.out

run: ## Run binary
	go run .

ci: build test lint bench ## Run all code checks and tests

clean: ## Remove object and cache files
	go clean

help:  ## Display this help
	awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
	{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

