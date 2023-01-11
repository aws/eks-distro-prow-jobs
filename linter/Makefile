ifndef PULL_BASE_SHA
export PULL_BASE_SHA=$(shell git show -s --format=%H upstream/main)
endif
ifndef PULL_PULL_SHA
export PULL_PULL_SHA=$(shell git show -s --format=%H)
endif

BIN_DIR := bin

.PHONY: all
all: build lint

.PHONY: build
build:  ## Build linter
build: fmt vet
	go build -o ./$(BIN_DIR)/prow-linter github.com/rcrozean/eks-distro-prow-jobs/linter

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: ## Run linter
lint: build
	./$(BIN_DIR)/prow-linter

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[%\/0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
