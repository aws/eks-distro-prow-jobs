BIN_DIR := bin
SCRIPTS_DIR := scripts

.PHONY: all
all: build prowjobs

.PHONY: build
build:  ## Build templater binary
build: fmt generate vet
	go build -o ./$(BIN_DIR)/generate-prowjobs github.com/kevinsalerno/eks-distro-prow-jobs/templater

fmt: ## Run go fmt against code.
	go fmt ./...

generate: ## Run go generate against code.
	go generate ./...

vet: ## Run go vet against code.
	go vet ./...

.PHONY: clean
clean:
	rm -rf bin

.PHONY: prowjobs
prowjobs: ## Generate Prowjob YAML files using templater
prowjobs: build
	./$(BIN_DIR)/generate-prowjobs

.PHONY: verify-prowjobs
verify-prowjobs: prowjobs
	./$(SCRIPTS_DIR)/verify-prowjobs.sh

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[%\/0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
