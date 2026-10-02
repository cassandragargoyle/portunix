# Portunix Testing and Build Automation
# Modern testing infrastructure with Go best practices

.PHONY: help benchmark benchmark-docker build build-helpers build-main build-race build-release build-version ci-integration ci-setup ci-test clean clean-all deploy-local deps dev-setup dist docker-cleanup docker-test-env docs-serve fmt gen-resources install-build-tools lint lint-md mocks security setup-test status test test-cli test-coverage test-coverage-ci test-docker test-docker-unit test-e2e test-fixtures test-integration test-performance test-release test-report test-unit undeploy-local vet watch-test

# Detect OS and set executable extension and commands
ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
    RM := del /f /q
    RMDIR := rmdir /s /q
    # UTF-8 support for Windows - run chcp 65001 before make commands
    CHCP := chcp 65001
    # Resolve Git's bundled bash for POSIX build scripts. `bash` on the cmd.exe
    # PATH resolves to the WSL launcher / WindowsApps stub, not a usable shell,
    # so locate git.exe via PATH and derive its sibling bash.exe (Git for
    # Windows layout: <git>\cmd\git.exe -> <git>\bin\bash.exe).
    GIT_BASH := $(shell for %%i in (git.exe) do @echo %%~dp$$PATH:i)..\bin\bash.exe
else
    EXE_EXT :=
    RM := rm -f
    RMDIR := rm -rf
    CHCP :=
endif

# Force uv to use only its managed Python; never probe the system PATH.
# `uv run --script` ignores [tool.uv] in pyproject.toml, so the preference must
# be an environment variable. This avoids the Windows Store python3.exe stub
# (present in the cmd.exe PATH) that hangs uv's interpreter probe.
export UV_PYTHON_PREFERENCE := only-managed

# All Go modules in the repository, discovered dynamically (Issue #194)
# so new helper modules are tested without editing the Makefile.
# Excludes vendored docs-site theme and plugin-development template.
# Satellite modules have stub go.mod files resolvable only through the root
# module's replace directives — their packages are tested from the root
# module by import path instead of a standalone `cd && go test`.
SATELLITE_MODULES := ./src/cmd ./src/app/sandbox ./src/app/virt
SATELLITE_PKGS := portunix.ai/cmd/... portunix.ai/app/sandbox/... portunix.ai/app/virt/...
GO_MODULES := $(filter-out $(SATELLITE_MODULES),$(sort $(patsubst %/go.mod,%,\
    $(shell find . -name go.mod \
    -not -path './docs/*' -not -path './docs-site/*' -not -path './.git/*'))))

# Default target
help: ## Show this help message
ifeq ($(OS),Windows_NT)
	@$(CHCP)
	@echo Portunix Testing and Build Commands
	@echo ========================================
	@echo.
	@findstr /R "^[a-zA-Z_-][a-zA-Z_-]*:.*##" $(MAKEFILE_LIST)
	@echo.
	@echo Usage: make [target]
	@echo Tip: Use 'make build' to compile, 'make test' to run tests
else
	@echo "Portunix Testing and Build Commands"
	@echo "========================================"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort
	@echo ""
	@echo "Usage: make [target]"
	@echo "Tip: Use 'make build' to compile, 'make test' to run tests"
endif

benchmark: ## Run benchmark tests
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

benchmark-docker: ## Run Docker-specific benchmarks
	@echo "Running Docker benchmarks..."
	go test -tags=integration -bench=BenchmarkDocker -benchmem ./pkg/docker/

build: build-main build-helpers ## Build main binary and all helpers (default)
	@echo "All binaries built successfully"

build-helpers: ## Build all helper binaries
	@echo "Building helper binaries..."
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
	@cd src/helpers/ptx-credential && go build -o ../../../ptx-credential$(EXE_EXT) .
	@cd src/helpers/ptx-trace && go build -o ../../../ptx-trace$(EXE_EXT) .
	@cd src/helpers/ptx-ssh && go build -o ../../../ptx-ssh$(EXE_EXT) .
	@cd src/helpers/ptx-plugin-registry && go build -o ../../../ptx-plugin-registry$(EXE_EXT) .
	@cd src/helpers/ptx-proxmox && go build -o ../../../ptx-proxmox$(EXE_EXT) .
	@cd src/helpers/ptx-specpm && go build -o ../../../ptx-specpm$(EXE_EXT) .
	@cd src/helpers/ptx-database && go build -o ../../../ptx-database$(EXE_EXT) .
	@cd src/helpers/ptx-github && go build -o ../../../ptx-github$(EXE_EXT) .
	@cd src/helpers/ptx-wizard && go build -o ../../../ptx-wizard$(EXE_EXT) .
	@echo "Helper binaries built: ptx-container, ptx-mcp, ptx-virt, ptx-ansible, ptx-prompting, ptx-python, ptx-installer, ptx-aiops, ptx-make, ptx-pft, ptx-credential, ptx-trace, ptx-ssh, ptx-plugin-registry, ptx-proxmox, ptx-specpm, ptx-database, ptx-github, ptx-wizard"

build-main: gen-resources ## Build only the main Portunix binary
	@echo "Building Portunix..."
	go build -o portunix$(EXE_EXT) .

# Regenerate portunix.syso from versioninfo.json (embeds version + manifest + icon).
# .syso is Windows-only — it's only embedded by Go when GOOS=windows. On
# non-Windows hosts native builds simply ignore it; cross-compiling to Windows
# from those hosts still picks up the latest .syso via this target.
#
# We resolve goversioninfo at recipe-time (inside sh) rather than via $(shell)
# to avoid Make's host-charset issues with non-ASCII GOPATH on Windows.
# Override by exporting GOVERSIONINFO=/path/to/goversioninfo.
gen-resources: ## Regenerate Windows .syso (version, manifest, icon) from versioninfo.json
ifeq ($(OS),Windows_NT)
	@powershell -Command "[Console]::OutputEncoding = [Text.Encoding]::UTF8; $$bin = if ($$env:GOVERSIONINFO) { $$env:GOVERSIONINFO } else { Join-Path (go env GOPATH) 'bin\goversioninfo.exe' }; if (-not (Test-Path $$bin)) { Write-Host \"goversioninfo not found at $$bin\"; Write-Host 'Run: make install-build-tools  (or:  go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest )'; exit 1 }; Write-Host 'Generating portunix.syso (version + manifest + icon)...'; & $$bin -o portunix.syso versioninfo.json; exit $$LASTEXITCODE"
else
	@bin="$${GOVERSIONINFO:-$$(go env GOPATH)/bin/goversioninfo$(EXE_EXT)}"; \
	if [ ! -f "$$bin" ]; then \
		echo "WARNING: goversioninfo not found at $$bin - keeping existing portunix.syso"; \
		echo "Run: make install-build-tools  (or:  go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest )"; \
		exit 0; \
	fi; \
	echo "Generating portunix.syso (version + manifest + icon)..."; \
	"$$bin" -o portunix.syso versioninfo.json
endif

install-build-tools: ## Install Go build tooling required for resource generation
	@echo "Installing goversioninfo..."
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	@echo "Build tools installed"

build-race: ## Build with race detection
	@echo "Building with race detection..."
	go build -race -o portunix$(EXE_EXT) .

build-release: ## Build release version with proper version embedding
ifeq ($(OS),Windows_NT)
	@echo Building Portunix release...
	@"$(GIT_BASH)" build-with-version.sh
else
	@echo "Building Portunix release..."
	./build-with-version.sh
endif

build-version: ## Build with custom version (use VERSION=v1.6.0)
ifeq ($(OS),Windows_NT)
	@if "$(VERSION)"=="" (echo Usage: make build-version VERSION=v1.6.0 & exit 1)
	@echo Building Portunix $(VERSION)...
	@"$(GIT_BASH)" build-with-version.sh $(VERSION)
else
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make build-version VERSION=v1.6.0"; \
		exit 1; \
	fi
	@echo "Building Portunix $(VERSION)..."
	./build-with-version.sh $(VERSION)
endif

ci-integration: test-integration ## Run CI integration tests
	@echo "CI integration tests completed"

ci-setup: deps setup-test test-fixtures ## Setup CI environment
	@echo "CI environment ready"

ci-test: lint vet test-coverage-ci ## Run CI test suite
	@echo "CI tests completed"

clean: ## Clean build artifacts and test files
	@echo "Cleaning up..."
	-$(RM) portunix$(EXE_EXT) ptx-container$(EXE_EXT) ptx-mcp$(EXE_EXT) ptx-virt$(EXE_EXT) ptx-ansible$(EXE_EXT) ptx-prompting$(EXE_EXT) ptx-python$(EXE_EXT) ptx-installer$(EXE_EXT) ptx-aiops$(EXE_EXT) ptx-make$(EXE_EXT) ptx-pft$(EXE_EXT) ptx-credential$(EXE_EXT) ptx-trace$(EXE_EXT) ptx-ssh$(EXE_EXT) ptx-plugin-registry$(EXE_EXT) ptx-proxmox$(EXE_EXT) ptx-specpm$(EXE_EXT) ptx-database$(EXE_EXT) ptx-github$(EXE_EXT) ptx-wizard$(EXE_EXT) ptx-vocalio$(EXE_EXT)
	-$(RM) coverage.out coverage.html
	-$(RMDIR) coverage/
	-$(RMDIR) test/tmp/
	go clean -testcache
	@echo "Cleanup complete"

clean-all: clean ## Clean everything including dependencies
	go clean -modcache
	-$(RMDIR) test/mocks/generated_*

deploy-local: ## Deploy pre-built binaries to local system (VERBOSE=1 for diagnostics)
	@echo "Deploying local binaries via uv..."
	@uv run --script scripts/deploy-local.py $(if $(VERBOSE),--verbose)

undeploy-local: ## Remove Portunix and helpers from local system (auto-detects install path)
	@echo "Removing local binaries via uv..."
	@uv run --script scripts/undeploy-local.py

deps: ## Install testing dependencies
	@echo "Installing testing dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/sonatype-nexus-community/nancy@latest
	go install github.com/axw/gocov/gocov@latest
	@echo "Testing dependencies installed"

dev-setup: deps setup-test test-fixtures ## Setup development environment
	@echo "Development environment ready"
	@echo "Run 'make test' to verify setup"

dist: ## Create distribution release (use VERSION=1.7.9 to override portunix.rc)
	@echo "Creating distribution release..."
ifeq ($(OS),Windows_NT)
	@powershell -Command "$$paramVersion = '$(VERSION)'; if ($$paramVersion) { $$ver = if ($$paramVersion.StartsWith('v')) { $$paramVersion } else { \"v$$paramVersion\" }; Write-Host \"Using provided version: $$ver\"; python scripts/make-release.py \"$$ver\" } else { $$version = (Select-String -Path 'portunix.rc' -Pattern '\"FileVersion\", \"([0-9]+\.[0-9]+\.[0-9]+)\"').Matches.Groups[1].Value; if ($$version) { Write-Host \"Detected version from portunix.rc: v$$version\"; python scripts/make-release.py \"v$$version\" } else { Write-Host 'Error: Could not read version from portunix.rc'; exit 1 } }"
else
	@if [ -n "$(VERSION)" ]; then \
		VER="$(VERSION)"; \
		case "$$VER" in v*) ;; *) VER="v$$VER" ;; esac; \
		echo "Using provided version: $$VER"; \
		python3 scripts/make-release.py "$$VER"; \
	else \
		VERSION=$$(grep -oP 'VALUE "FileVersion", "\K[0-9]+\.[0-9]+\.[0-9]+' portunix.rc | head -n1); \
		if [ -z "$$VERSION" ]; then \
			echo "Error: Could not read version from portunix.rc"; \
			exit 1; \
		fi; \
		echo "Detected version from portunix.rc: v$$VERSION"; \
		python3 scripts/make-release.py "v$$VERSION"; \
	fi
endif

docker-cleanup: ## Cleanup Docker test environment
	@echo "Cleaning up Docker test environment..."
	-docker stop portunix-test-registry
	-docker rm portunix-test-registry
	-docker system prune -f

docker-test-env: ## Start Docker test environment
	@echo "Starting Docker test environment..."
	docker run -d --name portunix-test-registry -p 5000:5000 registry:2
	@echo "Test registry started on localhost:5000"

docs-serve: ## Serve documentation locally (Hugo dev server)
ifeq ($(OS),Windows_NT)
	@scripts\docs-serve.cmd
else
	@./scripts/docs-serve.sh
endif

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

lint-md: ## Run Markdown linter
	@echo "Running Markdown linter..."
	markdownlint-cli2

mocks: ## Generate mocks
	@echo "Generating mocks..."
	go generate ./...

security: ## Run security scans
	@echo "Running security scans..."
	gosec ./...
	go list -json -deps ./... | nancy sleuth

setup-test: ## Setup test environment
	@echo "Setting up test environment..."
	mkdir -p test/{fixtures,mocks,testdata,integration}
	mkdir -p test/fixtures/{docker,install,system}
	mkdir -p internal/{testutils,testcontainers}
	@echo "Test directories created"

status: ## Show current testing status
	@echo "Portunix Testing Status"
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

test: ## Run all tests (unit + integration) across all Go modules
	@echo "Running all tests..."
	@set -e; for mod in $(GO_MODULES); do \
		pkgs="./..."; \
		if [ "$$mod" = "." ]; then pkgs="./... $(SATELLITE_PKGS)"; fi; \
		echo "==> go test $$mod"; \
		(cd $$mod && go test -v $$pkgs); \
	done

test-cli: ## Test CLI commands (src/cmd module, via root replace graph)
	@echo "Testing CLI commands..."
	go test -v portunix.ai/cmd/...

test-coverage: ## Run tests with coverage report (per-module HTML in coverage/)
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	@set -e; for mod in $(GO_MODULES); do \
		name=$$(echo $$mod | sed 's|^\./||; s|^\.$$|root|; s|/|-|g'); \
		pkgs="./..."; \
		if [ "$$mod" = "." ]; then pkgs="./... $(SATELLITE_PKGS)"; fi; \
		echo "==> coverage $$mod"; \
		(cd $$mod && go test -race -coverprofile=coverage.out -covermode=atomic $$pkgs \
			&& go tool cover -html=coverage.out -o coverage.html); \
		mv $$mod/coverage.out coverage/$$name.out; \
		mv $$mod/coverage.html coverage/$$name.html; \
	done
	@echo "Coverage reports generated in coverage/"

test-coverage-ci: ## Run coverage for CI (merged codecov profile in coverage.out)
	@echo "Running coverage for CI..."
	@set -e; echo "mode: atomic" > coverage.out; \
	for mod in $(GO_MODULES); do \
		pkgs="./..."; \
		if [ "$$mod" = "." ]; then pkgs="./... $(SATELLITE_PKGS)"; fi; \
		echo "==> coverage $$mod"; \
		(cd $$mod && go test -race -coverprofile=coverage.mod.out -covermode=atomic $$pkgs \
			&& go tool cover -func=coverage.mod.out); \
		tail -n +2 $$mod/coverage.mod.out >> coverage.out; \
		rm -f $$mod/coverage.mod.out; \
	done
	@echo "Merged coverage profile: coverage.out"

test-docker: ## Test Docker functionality specifically
	@echo "Testing Docker functionality..."
	go test -tags=integration -v ./pkg/docker/... ./cmd/docker*

test-docker-unit: ## Test Docker unit tests only
	@echo "Testing Docker unit tests..."
	go test -tags=unit -v ./pkg/docker/...

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	go test -tags=e2e -v -timeout=30m ./test/e2e/...

test-fixtures: ## Create test fixtures
	@echo "Creating test fixtures..."
	@mkdir -p test/fixtures/docker
	@echo "FROM alpine:latest" > test/fixtures/docker/valid_dockerfile
	@echo "INVALID DOCKERFILE CONTENT" > test/fixtures/docker/invalid_dockerfile
	@mkdir -p test/fixtures/install
	@echo '{"packages": ["docker", "python"]}' > test/fixtures/install/package.json
	@echo "invalid json content" > test/fixtures/install/invalid_config.json
	@echo "Test fixtures created"

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	go test -tags=integration -v -timeout=10m ./...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	go test -tags=performance -v -timeout=15m ./test/performance/...

test-release: ## Test release build
	@echo "Testing release build..."
	GOOS=linux GOARCH=amd64 go build -o portunix-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o portunix-windows-amd64.exe .
	@echo "Cross-platform builds successful"

# Cross-platform binary distribution targets (ADR-031, Issue #125)
PLATFORMS := linux-amd64 linux-arm64 windows-amd64 darwin-amd64

build-all-platforms: ## Build all binaries for all platforms (cross-platform distribution)
	@echo "Building binaries for all platforms..."
	@mkdir -p dist/platforms
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'-' -f1); \
		arch=$$(echo $$platform | cut -d'-' -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building for $$os/$$arch..."; \
		mkdir -p dist/platforms/$$platform; \
		abs_dist=$$(pwd)/dist/platforms/$$platform; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/portunix$$ext . || true; \
		cd src/helpers/ptx-container && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-container$$ext . && cd ../../..; \
		cd src/helpers/ptx-mcp && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-mcp$$ext . && cd ../../..; \
		cd src/helpers/ptx-virt && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-virt$$ext . && cd ../../..; \
		cd src/helpers/ptx-ansible && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-ansible$$ext . && cd ../../..; \
		cd src/helpers/ptx-prompting && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-prompting$$ext . && cd ../../..; \
		cd src/helpers/ptx-python && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-python$$ext . && cd ../../..; \
		cd src/helpers/ptx-installer && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-installer$$ext . && cd ../../..; \
		cd src/helpers/ptx-aiops && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-aiops$$ext . && cd ../../..; \
		cd src/helpers/ptx-make && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-make$$ext . && cd ../../..; \
		cd src/helpers/ptx-pft && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-pft$$ext . && cd ../../..; \
		cd src/helpers/ptx-credential && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-credential$$ext . && cd ../../..; \
		cd src/helpers/ptx-trace && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-trace$$ext . && cd ../../..; \
		cd src/helpers/ptx-ssh && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-ssh$$ext . && cd ../../..; \
		cd src/helpers/ptx-plugin-registry && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-plugin-registry$$ext . && cd ../../..; \
		cd src/helpers/ptx-proxmox && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-proxmox$$ext . && cd ../../..; \
		cd src/helpers/ptx-specpm && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-specpm$$ext . && cd ../../..; \
		cd src/helpers/ptx-database && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-database$$ext . && cd ../../..; \
		cd src/helpers/ptx-github && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-github$$ext . && cd ../../..; \
		cd src/helpers/ptx-wizard && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-wizard$$ext . && cd ../../..; \
	done
	@echo "All platform binaries built in dist/platforms/"

create-platform-archives: build-all-platforms ## Create platform archives for cross-platform distribution
	@echo "Creating platform archives..."
ifeq ($(OS),Windows_NT)
	@python scripts/create-platform-archives.py
else
	@python3 scripts/create-platform-archives.py
endif
	@echo "Platform archives created in dist/platforms/"

test-report: ## Generate comprehensive test report
	@echo "Generating test report..."
	@echo "==================== PORTUNIX TEST REPORT ====================" > test_report.txt
	@echo "Date: $$(date)" >> test_report.txt
	@echo "Git Commit: $$(git rev-parse HEAD)" >> test_report.txt
	@set -e; for mod in $(GO_MODULES); do \
		pkgs="./..."; \
		if [ "$$mod" = "." ]; then pkgs="./... $(SATELLITE_PKGS)"; fi; \
		echo "" >> test_report.txt; \
		echo "==> Module: $$mod" >> test_report.txt; \
		echo "Unit Tests:" >> test_report.txt; \
		(cd $$mod && go test -tags=unit $$pkgs) >> test_report.txt 2>&1; \
		echo "Coverage:" >> test_report.txt; \
		(cd $$mod && go test -coverprofile=coverage.mod.out $$pkgs \
			&& go tool cover -func=coverage.mod.out) >> test_report.txt 2>&1; \
		rm -f $$mod/coverage.mod.out; \
	done
	@echo "Test report generated: test_report.txt"

test-unit: ## Run unit tests only (all modules)
	@echo "Running unit tests..."
	@set -e; for mod in $(GO_MODULES); do \
		pkgs="./..."; \
		if [ "$$mod" = "." ]; then pkgs="./... $(SATELLITE_PKGS)"; fi; \
		echo "==> go test -tags=unit $$mod"; \
		(cd $$mod && go test -tags=unit -v $$pkgs); \
	done

vet: ## Run go vet (examines Go source code and reports suspicious constructs)
	@echo "Running go vet..."
	go vet ./...

watch-test: ## Watch for changes and run tests
	@echo "Watching for changes..."
	find . -name "*.go" | entr -c make test-unit
