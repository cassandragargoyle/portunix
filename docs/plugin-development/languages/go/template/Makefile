# Go Plugin Template Makefile

.PHONY: help build test clean proto docker install dev

# Build configuration
# The binary lands next to plugin.json, where its "binary" field points
BINARY_NAME=plugin
PROTO_DIR=proto
DOCKER_IMAGE=go-plugin-template
VERSION ?= $(shell git describe --tags --always --dirty)

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: proto ## Build the plugin binary
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .
	@echo "Build completed: $(BINARY_NAME)"

proto: ## Generate protocol buffer files
	@echo "Generating protocol buffers..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Protocol buffers generated"

test: proto ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Tests completed. Coverage report: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME) $(PROTO_DIR)/*.pb.go
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "Clean completed"

dev: proto ## Run in development mode
	@echo "Starting development server..."
	@go run . --config=config.yaml

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
	@tar -czf dist/$(BINARY_NAME)-$(VERSION).tar.gz $(BINARY_NAME) plugin.json config.yaml
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
		inotifywait -r -e modify,create,delete --exclude='\.git|dist|\.pb\.go' .; \
		make build; \
	done

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Default target
all: clean proto build test ## Clean, generate proto, build and test