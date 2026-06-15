.PHONY: build install clean test coverage lint lint-fix fmt vet ci fmt-check kubectl-plugin cross-compile release

# Binary name
BINARY_NAME=konflux-issues
KUBECTL_PLUGIN_NAME=kubectl-issues

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt

GOLANGCI_CONFIG ?= ../../configs/golangci.yml
LOCALBIN ?= $(shell pwd)/bin
GOLANGCI_LINT_BIN = $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.1.0

# Default target
all: fmt test build

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v main.go

# Format the code
fmt:
	$(GOFMT) ./...

vet: ## Run go vet against code.
	go vet ./...

# Run tests with coverage profile
test:
	$(GOTEST) ./... -coverprofile=cover.out -covermode=atomic

coverage: test ## Print test coverage summary.
	go tool cover -func=cover.out

lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT_BIN) run --config $(GOLANGCI_CONFIG)

lint-fix: golangci-lint ## Run golangci-lint and apply fixes.
	$(GOLANGCI_LINT_BIN) run --config $(GOLANGCI_CONFIG) --fix

ci: fmt-check vet lint build test ## Run all CI checks locally.

fmt-check: ## Verify code is gofmt-formatted.
	@FILES=$$(gofmt -l .); \
	if [ -n "$$FILES" ]; then \
		echo "The following files need formatting:"; \
		echo "$$FILES"; \
		exit 1; \
	fi

# Install the binary
install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY_NAME) $(HOME)/.local/bin/

# Install as kubectl plugin
kubectl-plugin: install
	ln -sf $(HOME)/.local/bin/$(BINARY_NAME) $(HOME)/.local/bin/$(KUBECTL_PLUGIN_NAME)

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) cover.out

# Cross-compile for multiple platforms
cross-compile:
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64 -v main.go
	# MacOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-darwin-amd64 -v main.go
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows-amd64.exe -v main.go

# Create GitHub release archives
release: cross-compile
	mkdir -p release
	tar -czvf release/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	tar -czvf release/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	zip -r release/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

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
