#!/bin/bash

echo "🧪 Testing Portunix Docker Implementation"
echo "========================================"

# Build the project
echo "📦 Building Portunix..."
go build -o .
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"
echo

# Test help commands
echo "📚 Testing help commands..."
echo "1. Docker main help:"
./portunix docker --help | head -10
echo

echo "2. Docker install help:"
./portunix docker install --help | head -5
echo

echo "3. Docker run-in-container help:"
./portunix docker run-in-container --help | head -5
echo

# Test install command integration
echo "🔧 Testing install command integration..."
echo "Checking if 'docker' appears in install help:"
./portunix install --help | grep docker
if [ $? -eq 0 ]; then
    echo "✅ Docker found in install help"
else
    echo "❌ Docker not found in install help"
fi
echo

# Test OS detection (dry run)
echo "🖥️  Testing OS detection..."
echo "Running: ./portunix install docker --help"
./portunix install docker --help | head -3
echo

# Test Docker commands (will fail gracefully if Docker not installed)
echo "🐳 Testing Docker commands (may fail if Docker not installed)..."

echo "Testing docker list:"
./portunix docker list 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✅ Docker list command works"
else
    echo "ℹ️  Docker list failed (Docker probably not installed)"
fi

echo "Testing docker build help:"
./portunix docker build --help | head -3

echo "Testing invalid installation type:"
./portunix docker run-in-container invalid-type 2>&1 | head -2

echo
echo "🎉 Basic testing completed!"
echo "💡 For full testing, install Docker and run:"
echo "   ./portunix docker run-in-container empty --image alpine:latest"