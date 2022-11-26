.PHONY: help build test docker deps clean
NAME = calibre-api
VERSION = latest
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build executable files
	@goreleaser release --rm-dist --snapshot

test: ## Run tests
	go install "github.com/rakyll/gotest@latest"
	GIN_MODE=release
	LOG_LEVEL=fatal ## disable log for test
	gotest -v -coverprofile=coverage.out -covermode=atomic ./...

docker: ## Build docker images
	docker build -t $(NAME):$(VERSION) .

deps: ## Update vendor.
	go mod verify
	go mod tidy -v
	go get -u ./...

clean: ## Clean up build files.
	rm -rf dist/
