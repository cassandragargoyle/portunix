# Go Plugin Template Makefile

.PHONY: help build test clean proto docker install dev

# Build configuration
BINARY_NAME=plugin
BUILD_DIR=bin
PROTO_DIR=proto
DOCKER_IMAGE=go-plugin-template
VERSION ?= $(shell git describe --tags --always --dirty)

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: proto ## Build the plugin binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

proto: ## Generate protocol buffer files
	@echo "Generating protocol buffers..."
	@protoc --go_out=. --go-grpc_out=. $(PROTO_DIR)/*.proto
	@echo "Protocol buffers generated"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Tests completed. Coverage report: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "Clean completed"

dev: proto ## Run in development mode
	@echo "Starting development server..."
	@go run main.go -config=config.dev.yaml

deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run

docker: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "Docker image built: $(DOCKER_IMAGE):$(VERSION)"

package: build ## Create distribution package
	@echo "Creating package..."
	@mkdir -p dist
	@tar -czf dist/$(BINARY_NAME)-$(VERSION).tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	@cp plugin.yaml dist/
	@echo "Package created: dist/$(BINARY_NAME)-$(VERSION).tar.gz"

install: build ## Install plugin to Portunix
	@echo "Installing plugin..."
	@portunix plugin install .
	@echo "Plugin installed"

validate: ## Validate plugin
	@echo "Validating plugin..."
	@portunix plugin validate .
	@echo "Plugin validation completed"

# Development helpers
watch: ## Watch for changes and rebuild
	@echo "Watching for changes..."
	@while true; do \
		inotifywait -r -e modify,create,delete --exclude='\.git|bin|dist' .; \
		make build; \
	done

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

profile: ## Run with profiling
	@echo "Starting with profiling..."
	@go run main.go -config=config.dev.yaml -cpuprofile=cpu.prof -memprofile=mem.prof

# Default target
all: clean proto build test ## Clean, generate proto, build and test