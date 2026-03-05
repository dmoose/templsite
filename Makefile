.PHONY: help build test clean install dev setup deps setup-tailwind generate

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

BINARY_NAME=templsite
BINARY_PATH=./bin/$(BINARY_NAME)

# Tailwind CSS standalone CLI
# Note: Update this to the latest version from https://github.com/tailwindlabs/tailwindcss/releases
TAILWIND_VERSION=latest
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Detect OS
ifeq ($(UNAME_S),Linux)
    TAILWIND_OS=linux
endif
ifeq ($(UNAME_S),Darwin)
    TAILWIND_OS=macos
endif

# Detect Architecture
ifeq ($(UNAME_M),x86_64)
    TAILWIND_ARCH=x64
endif
ifeq ($(UNAME_M),arm64)
    TAILWIND_ARCH=arm64
endif
ifeq ($(UNAME_M),aarch64)
    TAILWIND_ARCH=arm64
endif

TAILWIND_URL=https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-$(TAILWIND_OS)-$(TAILWIND_ARCH)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	@echo "Downloading Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded successfully"

setup-tailwind: ## Download Tailwind CSS standalone CLI (if not already available)
	@if command -v tailwindcss >/dev/null 2>&1; then \
		echo "✓ Tailwind CSS CLI found in system PATH"; \
		tailwindcss --version | head -n1; \
	elif [ -f bin/tailwindcss ]; then \
		echo "✓ Tailwind CSS CLI found at bin/tailwindcss"; \
		./bin/tailwindcss --version | head -n1; \
	else \
		echo "Downloading Tailwind CSS v4 standalone CLI..."; \
		echo "OS: $(TAILWIND_OS), Architecture: $(TAILWIND_ARCH)"; \
		mkdir -p bin; \
		curl -sL $(TAILWIND_URL) -o bin/tailwindcss; \
		chmod +x bin/tailwindcss; \
		echo "✓ Tailwind CSS CLI installed to bin/tailwindcss"; \
		./bin/tailwindcss --version | head -n1; \
	fi

setup: deps setup-tailwind ## Complete project setup
	@echo "✓ Setup complete!"

generate: ## Generate templ components
	@echo "Generating templ components..."
	@command -v templ >/dev/null 2>&1 || { echo "Installing templ CLI..."; $(GOINSTALL) github.com/a-h/templ/cmd/templ@latest; }
	@templ generate
	@echo "✓ Components generated"

build: deps ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) ./cmd/templsite
	@echo "✓ Binary built: $(BINARY_PATH)"

build-release: deps ## Build release binary with version info
	@echo "Building release binary..."
	@mkdir -p bin
	$(GOBUILD) -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_PATH) ./cmd/templsite
	@echo "✓ Release binary built: $(BINARY_PATH)"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@find . -name '*_templ.go' -delete
	@echo "✓ Cleaned"

install: build ## Install binary to GOPATH
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) ./cmd/templsite
	@echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

dev: build ## Start development mode
	@echo "Starting development server..."
	./$(BINARY_PATH) serve --watch

example: build setup-tailwind ## Create and serve example site
	@echo "Creating example site..."
	@mkdir -p example
	@cd example && ../$(BINARY_PATH) new . --template business
	@echo "Starting example site..."
	@cd example && ../$(BINARY_PATH) serve

lint: ## Run linters
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install from: https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "✓ Vet complete"

mod-update: ## Update Go dependencies
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "✓ Dependencies updated"

all: clean deps build test ## Clean, build and test
	@echo "✓ All tasks complete"
