.PHONY: migrate seed status migration test coverage lint lint-fix fmt vet build ci tidy-check fmt-check

GOLANGCI_CONFIG ?= ../../configs/golangci.yml
GOLANGCI_LINT ?= golangci-lint
LOCALBIN ?= $(shell pwd)/bin
GOLANGCI_LINT_BIN = $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.1.0

# Apply pending migrations
migrate:
	atlas migrate apply --env local

# Seed database (non-prod DBs)
seed:
	go run -mod=mod cmd/seed/main.go

# Get status of DB migrations (applied, pending)
status:
	atlas migrate status --env local

# Generate a new migration (requires NAME variable)
# Usage: make migration NAME="add_some_column"
migration:
	atlas migrate diff "$(NAME)" --env local

##@ Development

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

test: ## Run tests with coverage profile.
	go test ./... -coverprofile=cover.out -covermode=atomic

coverage: test ## Print test coverage summary.
	go tool cover -func=cover.out

lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT_BIN) run --config $(GOLANGCI_CONFIG)

lint-fix: golangci-lint ## Run golangci-lint and apply fixes.
	$(GOLANGCI_LINT_BIN) run --config $(GOLANGCI_CONFIG) --fix

build: ## Build all packages.
	go build ./...

ci: tidy-check vet fmt-check lint build test ## Run all CI checks locally.

tidy-check: ## Verify go.mod and go.sum are tidy.
	go mod tidy
	@git diff --exit-code go.mod go.sum || { \
		echo "Please run go mod tidy and commit changes"; \
		git diff go.mod go.sum; \
		exit 1; \
	}

fmt-check: ## Verify code is gofmt-formatted.
	@FILES=$$(gofmt -l .); \
	if [ -n "$$FILES" ]; then \
		echo "The following files need formatting:"; \
		echo "$$FILES"; \
		exit 1; \
	fi

##@ Dependencies

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT_BIN) ## Download golangci-lint locally if necessary.

$(GOLANGCI_LINT_BIN): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT_BIN),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3); \
echo "Downloading $${package}"; \
rm -f $(1) || true; \
GOBIN=$(LOCALBIN) go install $${package}; \
mv $(1) $(1)-$(3); \
}; \
ln -sf $(1)-$(3) $(1)
endef
