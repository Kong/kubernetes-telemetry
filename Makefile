# ------------------------------------------------------------------------------
# Configuration - Tooling
# ------------------------------------------------------------------------------

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: _download_tool
_download_tool:
	(cd third_party && go mod tidy ) && \
		GOBIN=$(PROJECT_DIR)/bin go install -modfile third_party/go.mod $(TOOL)

GOLANGCI_LINT = $(PROJECT_DIR)/bin/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(MAKE) _download_tool TOOL=github.com/golangci/golangci-lint/cmd/golangci-lint

# ------------------------------------------------------------------------------
# Build & Tests
# ------------------------------------------------------------------------------

.PHONY: lint
lint: golangci-lint
	$(GOLANGCI_LINT) run -v

.PHONY: test.unit
test.unit:
	go test -count 1 -v ./...
