# Portunix Testing and Build Automation
# Modern testing infrastructure with Go best practices

.PHONY: help benchmark benchmark-docker build build-helpers build-main build-race build-release build-version ci-integration ci-setup ci-test clean clean-all deps dev-setup dist docker-cleanup docker-test-env docs-serve fmt lint mocks security setup-test status test test-cli test-coverage test-coverage-ci test-docker test-docker-unit test-e2e test-fixtures test-integration test-performance test-release test-report test-unit vet watch-test

# Detect OS and set executable extension and commands
ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
    RM := del /f /q
    RMDIR := rmdir /s /q
    # UTF-8 support for Windows - run chcp 65001 before make commands
    CHCP := chcp 65001 
else
    EXE_EXT :=
    RM := rm -f
    RMDIR := rm -rf
    CHCP :=
endif

# Default target
help: ## Show this help message
ifeq ($(OS),Windows_NT)
	@$(CHCP)
endif
	@echo "🔧 Portunix Testing and Build Commands"
	@echo "========================================"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "📖 Usage: make [target]"
	@echo "💡 Tip: Use 'make build' to compile, 'make test' to run tests"

benchmark: ## Run benchmark tests
	@echo "🚀 Running benchmarks..."
	go test -bench=. -benchmem ./...

benchmark-docker: ## Run Docker-specific benchmarks
	@echo "🐳 Running Docker benchmarks..."
	go test -tags=integration -bench=BenchmarkDocker -benchmem ./pkg/docker/

build: build-main build-helpers ## Build main binary and all helpers (default)
	@echo "🎉 All binaries built successfully"

build-helpers: ## Build all helper binaries
	@echo "🔧 Building helper binaries..."
	@cd src/helpers/ptx-container && go build -o ../../../ptx-container$(EXE_EXT) .
	@cd src/helpers/ptx-mcp && go build -o ../../../ptx-mcp$(EXE_EXT) .
	@cd src/helpers/ptx-virt && go build -o ../../../ptx-virt$(EXE_EXT) .
	@cd src/helpers/ptx-ansible && go build -o ../../../ptx-ansible$(EXE_EXT) .
	@cd src/helpers/ptx-prompting && go build -o ../../../ptx-prompting$(EXE_EXT) .
	@cd src/helpers/ptx-python && go build -o ../../../ptx-python$(EXE_EXT) .
	@cd src/helpers/ptx-installer && go build -o ../../../ptx-installer$(EXE_EXT) .
	@cd src/helpers/ptx-aiops && go build -o ../../../ptx-aiops$(EXE_EXT) .
	@cd src/helpers/ptx-make && go build -o ../../../ptx-make$(EXE_EXT) .
	@cd src/helpers/ptx-pft && go build -o ../../../ptx-pft$(EXE_EXT) .
	@echo "✅ Helper binaries built: ptx-container, ptx-mcp, ptx-virt, ptx-ansible, ptx-prompting, ptx-python, ptx-installer, ptx-aiops, ptx-make, ptx-pft"

build-main: ## Build only the main Portunix binary
	@echo "📦 Building Portunix..."
	go build -o portunix$(EXE_EXT) .

build-race: ## Build with race detection
	@echo "🏃 Building with race detection..."
	go build -race -o portunix$(EXE_EXT) .

build-release: ## Build release version with proper version embedding
	@echo "🎁 Building Portunix release..."
	./build-with-version.sh

build-version: ## Build with custom version (use VERSION=v1.6.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make build-version VERSION=v1.6.0"; \
		exit 1; \
	fi
	@echo "🏷️  Building Portunix $(VERSION)..."
	./build-with-version.sh $(VERSION)

ci-integration: test-integration ## Run CI integration tests
	@echo "✅ CI integration tests completed"

ci-setup: deps setup-test test-fixtures ## Setup CI environment
	@echo "🚀 CI environment ready"

ci-test: lint vet test-coverage-ci ## Run CI test suite
	@echo "✅ CI tests completed"

clean: ## Clean build artifacts and test files
	@echo "🧹 Cleaning up..."
	-$(RM) portunix$(EXE_EXT) ptx-container$(EXE_EXT) ptx-mcp$(EXE_EXT) ptx-virt$(EXE_EXT) ptx-ansible$(EXE_EXT) ptx-prompting$(EXE_EXT) ptx-python$(EXE_EXT) ptx-installer$(EXE_EXT) ptx-aiops$(EXE_EXT) ptx-make$(EXE_EXT) ptx-pft$(EXE_EXT) ptx-vocalio$(EXE_EXT)
	-$(RM) coverage.out coverage.html
	-$(RMDIR) test/tmp/
	go clean -testcache
	@echo "✅ Cleanup complete"

clean-all: clean ## Clean everything including dependencies
	go clean -modcache
	-$(RMDIR) test/mocks/generated_*

deps: ## Install testing dependencies
	@echo "📥 Installing testing dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/sonatype-nexus-community/nancy@latest
	go install github.com/axw/gocov/gocov@latest
	@echo "✅ Testing dependencies installed"

dev-setup: deps setup-test test-fixtures ## Setup development environment
	@echo "🔧 Development environment ready"
	@echo "Run 'make test' to verify setup"

dist: ## Create distribution release (use VERSION=1.7.9 to override portunix.rc)
	@echo "📦 Creating distribution release..."
ifeq ($(OS),Windows_NT)
	@powershell -Command "$$paramVersion = '$(VERSION)'; if ($$paramVersion) { $$ver = if ($$paramVersion.StartsWith('v')) { $$paramVersion } else { \"v$$paramVersion\" }; Write-Host \"🏷️  Using provided version: $$ver\"; bash scripts/make-release.sh \"$$ver\" } else { $$version = (Select-String -Path 'portunix.rc' -Pattern '\"FileVersion\", \"([0-9]+\.[0-9]+\.[0-9]+)\"').Matches.Groups[1].Value; if ($$version) { Write-Host \"🏷️  Detected version from portunix.rc: v$$version\"; bash scripts/make-release.sh \"v$$version\" } else { Write-Host 'Error: Could not read version from portunix.rc'; exit 1 } }"
else
	@if [ -n "$(VERSION)" ]; then \
		VER="$(VERSION)"; \
		case "$$VER" in v*) ;; *) VER="v$$VER" ;; esac; \
		echo "🏷️  Using provided version: $$VER"; \
		bash scripts/make-release.sh "$$VER"; \
	else \
		VERSION=$$(grep -oP 'VALUE "FileVersion", "\K[0-9]+\.[0-9]+\.[0-9]+' portunix.rc | head -n1); \
		if [ -z "$$VERSION" ]; then \
			echo "❌ Error: Could not read version from portunix.rc"; \
			exit 1; \
		fi; \
		echo "🏷️  Detected version from portunix.rc: v$$VERSION"; \
		bash scripts/make-release.sh "v$$VERSION"; \
	fi
endif

docker-cleanup: ## Cleanup Docker test environment
	@echo "🧹 Cleaning up Docker test environment..."
	-docker stop portunix-test-registry
	-docker rm portunix-test-registry
	-docker system prune -f

docker-test-env: ## Start Docker test environment
	@echo "🐳 Starting Docker test environment..."
	docker run -d --name portunix-test-registry -p 5000:5000 registry:2
	@echo "✅ Test registry started on localhost:5000"

docs-serve: ## Serve documentation locally (Hugo dev server)
ifeq ($(OS),Windows_NT)
	@scripts\docs-serve.cmd
else
	@./scripts/docs-serve.sh
endif

fmt: ## Format code
	@echo "✨ Formatting code..."
	go fmt ./...
	goimports -w .

lint: ## Run linters
	@echo "🔍 Running linters..."
	golangci-lint run ./...

mocks: ## Generate mocks
	@echo "🎭 Generating mocks..."
	go generate ./...

security: ## Run security scans
	@echo "🔒 Running security scans..."
	gosec ./...
	go list -json -deps ./... | nancy sleuth

setup-test: ## Setup test environment
	@echo "🛠️  Setting up test environment..."
	mkdir -p test/{fixtures,mocks,testdata,integration}
	mkdir -p test/fixtures/{docker,install,system}
	mkdir -p internal/{testutils,testcontainers}
	@echo "✅ Test directories created"

status: ## Show current testing status
	@echo "📊 Portunix Testing Status"
	@echo "========================="
	@echo "Go version: $$(go version)"
	@echo "Git branch: $$(git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo ""
	@echo "Test files found:"
	@find . -name "*_test.go" -type f | wc -l | sed 's/^/  /'
	@echo ""
	@echo "Dependencies status:"
	@command -v golangci-lint >/dev/null 2>&1 && echo "  ✅ golangci-lint" || echo "  ❌ golangci-lint"
	@command -v mockgen >/dev/null 2>&1 && echo "  ✅ mockgen" || echo "  ❌ mockgen"
	@command -v gosec >/dev/null 2>&1 && echo "  ✅ gosec" || echo "  ❌ gosec"
	@command -v docker >/dev/null 2>&1 && echo "  ✅ docker" || echo "  ❌ docker"

test: ## Run all tests (unit + integration)
	@echo "🧪 Running all tests..."
	go test -v ./...

test-cli: ## Test CLI commands
	@echo "💻 Testing CLI commands..."
	go test -v ./cmd/...

test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-ci: ## Run coverage for CI (with codecov format)
	@echo "📊 Running coverage for CI..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

test-docker: ## Test Docker functionality specifically
	@echo "🐳 Testing Docker functionality..."
	go test -tags=integration -v ./pkg/docker/... ./cmd/docker*

test-docker-unit: ## Test Docker unit tests only
	@echo "⚡ Testing Docker unit tests..."
	go test -tags=unit -v ./pkg/docker/...

test-e2e: ## Run end-to-end tests
	@echo "🎯 Running E2E tests..."
	go test -tags=e2e -v -timeout=30m ./test/e2e/...

test-fixtures: ## Create test fixtures
	@echo "📋 Creating test fixtures..."
	@mkdir -p test/fixtures/docker
	@echo "FROM alpine:latest" > test/fixtures/docker/valid_dockerfile
	@echo "INVALID DOCKERFILE CONTENT" > test/fixtures/docker/invalid_dockerfile
	@mkdir -p test/fixtures/install
	@echo '{"packages": ["docker", "python"]}' > test/fixtures/install/package.json
	@echo "invalid json content" > test/fixtures/install/invalid_config.json
	@echo "✅ Test fixtures created"

test-integration: ## Run integration tests (requires Docker)
	@echo "🔗 Running integration tests..."
	go test -tags=integration -v -timeout=10m ./...

test-performance: ## Run performance tests
	@echo "🚀 Running performance tests..."
	go test -tags=performance -v -timeout=15m ./test/performance/...

test-release: ## Test release build
	@echo "🎁 Testing release build..."
	GOOS=linux GOARCH=amd64 go build -o portunix-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o portunix-windows-amd64.exe .
	@echo "✅ Cross-platform builds successful"

test-report: ## Generate comprehensive test report
	@echo "📊 Generating test report..."
	@echo "==================== PORTUNIX TEST REPORT ====================" > test_report.txt
	@echo "Date: $$(date)" >> test_report.txt
	@echo "Git Commit: $$(git rev-parse HEAD)" >> test_report.txt
	@echo "" >> test_report.txt
	@echo "Unit Tests:" >> test_report.txt
	go test -tags=unit ./... >> test_report.txt 2>&1
	@echo "" >> test_report.txt
	@echo "Coverage:" >> test_report.txt
	go test -coverprofile=coverage.out ./... >> test_report.txt 2>&1
	go tool cover -func=coverage.out >> test_report.txt
	@echo "✅ Test report generated: test_report.txt"

test-unit: ## Run unit tests only
	@echo "⚡ Running unit tests..."
	go test -tags=unit -v ./...

vet: ## Run go vet (examines Go source code and reports suspicious constructs)
	@echo "🔎 Running go vet..."
	go vet ./...

watch-test: ## Watch for changes and run tests
	@echo "👀 Watching for changes..."
	find . -name "*.go" | entr -c make test-unit
