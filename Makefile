all: check

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: addlicense
addlicense: ## Add license headers to all go files.
	go run github.com/google/addlicense -c 'OnMetal authors' **/*.go

.PHONY: generate
generate: ## Generate code (mocks etc.).
	go generate ./...

.PHONY: test
test: ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: ## Run the linter.
	golangci-lint run ./...

.PHONY: checklicense
checklicense: ## Check that every file has a license header present.
	go run github.com/google/addlicense -check -c 'OnMetal authors' **/*.go

.PHONY: check
check: generate test lint checklicense ## Execute multiple checks.

