#!/bin/bash
# Run PTX-TRACE Java SDK Example
# Usage: ./run.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$SCRIPT_DIR/../../java"
PROJECT_ROOT="$SCRIPT_DIR/../../../../../.."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "PTX-TRACE Java SDK Example"
echo "=========================="

# Check Java
if ! command -v java &> /dev/null; then
    echo -e "${RED}Error: Java not found. Install Java 21+${NC}"
    exit 1
fi

JAVA_VERSION=$(java -version 2>&1 | head -1 | cut -d'"' -f2 | cut -d'.' -f1)
echo "Java version: $JAVA_VERSION"

# Check Maven
if ! command -v mvn &> /dev/null; then
    echo -e "${RED}Error: Maven not found. Install Maven 3.6+${NC}"
    exit 1
fi

# Check portunix binary
if command -v portunix &> /dev/null; then
    echo -e "${GREEN}portunix found in PATH${NC}"
elif [ -f "$PROJECT_ROOT/portunix" ]; then
    echo -e "${YELLOW}Using portunix from project root${NC}"
    export PATH="$PROJECT_ROOT:$PATH"
else
    echo -e "${RED}Warning: portunix binary not found${NC}"
    echo "Build it first: cd $PROJECT_ROOT && make build"
fi

# Build SDK and copy dependencies
JAR_FILE="$SDK_DIR/target/ptx-trace-1.0.0.jar"
DEPS_DIR="$SDK_DIR/target/dependency"

if [ ! -f "$JAR_FILE" ] || [ ! -d "$DEPS_DIR" ]; then
    echo ""
    echo "Building Java SDK and downloading dependencies..."
    cd "$SDK_DIR"
    mvn clean package dependency:copy-dependencies -q -DskipTests
    cd "$SCRIPT_DIR"
fi

# Build classpath with all dependencies
CLASSPATH="$JAR_FILE"
for jar in "$DEPS_DIR"/*.jar; do
    [ -f "$jar" ] && CLASSPATH="$CLASSPATH:$jar"
done

echo ""
echo "Compiling example..."
javac -cp "$CLASSPATH" -d "$SCRIPT_DIR" "$SCRIPT_DIR/EtlPipeline.java"

echo ""
echo "Running example..."
echo "----------------------------------------"
java -cp "$SCRIPT_DIR:$CLASSPATH" ai.portunix.trace.examples.EtlPipeline

echo ""
echo "----------------------------------------"
echo -e "${GREEN}Example completed${NC}"

# Cleanup
rm -f "$SCRIPT_DIR"/*.class
