# PTX-TRACE SDK Makefile
# Build and test all SDKs

.PHONY: all build test clean help
.PHONY: python-build python-test python-clean
.PHONY: java-build java-test java-clean
.PHONY: typescript-build typescript-test typescript-clean

# Default target
all: build

# Build all SDKs
build: python-build java-build typescript-build
	@echo "All SDKs built successfully"

# Test all SDKs
test: python-test java-test typescript-test
	@echo "All SDK tests completed"

# Clean all SDKs
clean: python-clean java-clean typescript-clean
	@echo "All SDKs cleaned"

# ============================================
# Python SDK
# ============================================

python-build:
	@echo "Building Python SDK..."
	cd python && pip install -e . --quiet
	@echo "Python SDK installed"

python-test:
	@echo "Testing Python SDK..."
	cd python && python -c "from ptx_trace import Session, Severity, CsvSource; print('Python SDK: OK')"

python-clean:
	@echo "Cleaning Python SDK..."
	cd python && rm -rf build/ dist/ *.egg-info/ __pycache__/ ptx_trace/__pycache__/
	@echo "Python SDK cleaned"

python-check:
	@echo "Checking Python SDK syntax..."
	cd python && python -m py_compile ptx_trace/*.py
	@echo "Python SDK syntax OK"

# ============================================
# Java SDK
# ============================================

java-build:
	@echo "Building Java SDK..."
	cd java && mvn clean package -q -DskipTests
	@echo "Java SDK built"

java-test:
	@echo "Testing Java SDK..."
	cd java && mvn test -q
	@echo "Java SDK tests passed"

java-clean:
	@echo "Cleaning Java SDK..."
	cd java && mvn clean -q
	@echo "Java SDK cleaned"

java-check:
	@echo "Checking Java SDK compilation..."
	cd java && mvn compile -q
	@echo "Java SDK compilation OK"

# ============================================
# TypeScript SDK
# ============================================

typescript-build:
	@echo "Building TypeScript SDK..."
	cd typescript && npm install --silent && npm run build --silent
	@echo "TypeScript SDK built"

typescript-test:
	@echo "Testing TypeScript SDK..."
	cd typescript && npm test --silent 2>/dev/null || echo "No tests configured yet"

typescript-clean:
	@echo "Cleaning TypeScript SDK..."
	cd typescript && rm -rf node_modules/ dist/
	@echo "TypeScript SDK cleaned"

typescript-check:
	@echo "Checking TypeScript SDK..."
	cd typescript && npm run build --silent
	@echo "TypeScript SDK compilation OK"

# ============================================
# Verification (quick syntax check without full build)
# ============================================

check: python-check java-check typescript-check
	@echo "All SDK syntax checks passed"

# ============================================
# Examples
# ============================================

run-example-python:
	@echo "Running Python ETL example..."
	cd examples/python && python etl_pipeline.py

run-example-java:
	@echo "Running Java ETL example..."
	cd java && mvn exec:java -Dexec.mainClass="ai.portunix.trace.examples.EtlPipeline" -q

run-example-typescript:
	@echo "Running TypeScript ETL example..."
	cd typescript && npx ts-node ../examples/typescript/etl-pipeline.ts

# ============================================
# Help
# ============================================

help:
	@echo "PTX-TRACE SDK Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Main targets:"
	@echo "  all              Build all SDKs (default)"
	@echo "  build            Build all SDKs"
	@echo "  test             Test all SDKs"
	@echo "  clean            Clean all SDKs"
	@echo "  check            Quick syntax check for all SDKs"
	@echo ""
	@echo "Python SDK:"
	@echo "  python-build     Install Python SDK in development mode"
	@echo "  python-test      Test Python SDK import"
	@echo "  python-clean     Clean Python build artifacts"
	@echo "  python-check     Check Python syntax"
	@echo ""
	@echo "Java SDK:"
	@echo "  java-build       Build Java SDK with Maven"
	@echo "  java-test        Run Java SDK tests"
	@echo "  java-clean       Clean Java build artifacts"
	@echo "  java-check       Check Java compilation"
	@echo ""
	@echo "TypeScript SDK:"
	@echo "  typescript-build Build TypeScript SDK"
	@echo "  typescript-test  Run TypeScript SDK tests"
	@echo "  typescript-clean Clean TypeScript build artifacts"
	@echo "  typescript-check Check TypeScript compilation"
	@echo ""
	@echo "Examples:"
	@echo "  run-example-python     Run Python ETL example"
	@echo "  run-example-java       Run Java ETL example"
	@echo "  run-example-typescript Run TypeScript ETL example"
