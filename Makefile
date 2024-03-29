PROJECT_NAME := "gino-keva"
VERSION := "$(shell git describe --tags --match "v*.*.*")"
LAST_VERSION := "$(shell git describe --tags --match "v*.*.*" --abbrev=0 HEAD^)"
PKG := "github.com/philips-software/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep lint vet test test-coverage build clean
 
all: build lint test

dep: ## Get the dependencies
	@go mod tidy

lint: ## Lint Golang files
	@golint -set_exit_status ${PKG_LIST}
	@staticcheck ./...

vet: ## Run go vet
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST} 
	@cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@CGO_ENABLED=0 govvv build -pkg $(PKG)/internal/versioninfo -version $(VERSION) -o build/$(PROJECT_NAME) $(PKG)
 
clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)/build

release-notes: ## Generate release notes
	@git log $(LAST_VERSION)..HEAD --pretty=format:%s
	 
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
