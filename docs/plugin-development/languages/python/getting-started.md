# Python Plugin Development Guide

Python offers rapid development and access to a rich ecosystem of libraries for Portunix plugins. This guide will help you create powerful, feature-rich plugins using Python.

## Prerequisites

- Python 3.8 or later
- Protocol Buffers compiler (`protoc`)
- Protocol Buffers Python library
- Portunix development environment

## Quick Start

Create a new Python plugin using the Portunix CLI:

```bash
portunix plugin create my-python-plugin --language=python
cd my-python-plugin
```

This creates a complete Python project structure:

```
my-python-plugin/
├── plugin.yaml              # Plugin manifest
├── requirements.txt         # Python dependencies
├── setup.py                 # Package setup
├── src/                     # Source code
│   ├── __init__.py
│   ├── main.py             # Plugin entry point
│   ├── config/             # Configuration handling
│   ├── handlers/           # gRPC handlers
│   └── services/           # Business logic
├── proto/                  # Protocol Buffer definitions
├── tests/                  # Test files
├── scripts/                # Build and deployment scripts
└── README.md               # Plugin documentation
```

## Plugin Structure

### Main Entry Point (src/main.py)
```python
#!/usr/bin/env python3

import asyncio
import argparse
import logging
import signal
import sys
from concurrent import futures
from pathlib import Path

import grpc
from grpc_health.v1 import health_pb2_grpc
from grpc_health.v1.health_pb2 import HealthCheckResponse

from config.config import Config
from handlers.plugin_handler import PluginHandler
from proto import plugin_pb2_grpc


async def main():
    parser = argparse.ArgumentParser(description='Python Plugin for Portunix')
    parser.add_argument('--port', type=int, default=50051, help='gRPC server port')
    parser.add_argument('--health-port', type=int, default=50052, help='Health check port')
    parser.add_argument('--config', type=str, default='config.yaml', help='Configuration file path')
    args = parser.parse_args()

    # Load configuration
    try:
        config = Config.load(args.config)
    except Exception as e:
        logging.error(f"Failed to load configuration: {e}")
        sys.exit(1)

    # Configure logging
    log_level = getattr(logging, config.log_level.upper(), logging.INFO)
    logging.basicConfig(
        level=log_level,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.StreamHandler(),
            logging.FileHandler('plugin.log')
        ]
    )
    
    logger = logging.getLogger(__name__)
    logger.info(f"Starting {config.plugin.name} v{config.plugin.version}")

    # Create gRPC server
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Register plugin service
    plugin_handler = PluginHandler(config)
    plugin_pb2_grpc.add_PluginServiceServicer_to_server(plugin_handler, server)
    
    # Register health service
    health_servicer = HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    
    # Configure server address
    listen_addr = f'[::]:{args.port}'
    server.add_insecure_port(listen_addr)
    
    # Start server
    await server.start()
    logger.info(f"Plugin server listening on {listen_addr}")
    
    # Handle graceful shutdown
    def signal_handler(signum, frame):
        logger.info("Received shutdown signal, stopping server...")
        asyncio.create_task(server.stop(5))
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Server stopped by user")
    finally:
        plugin_handler.cleanup()
        logger.info("Plugin stopped")


class HealthServicer(health_pb2_grpc.HealthServicer):
    def Check(self, request, context):
        return HealthCheckResponse(status=HealthCheckResponse.SERVING)


if __name__ == '__main__':
    asyncio.run(main())
```

### Configuration Module (src/config/config.py)
```python
import yaml
from dataclasses import dataclass
from typing import Dict, Any, Optional
from pathlib import Path


@dataclass
class PluginConfig:
    name: str
    version: str
    description: str


@dataclass
class Config:
    plugin: PluginConfig
    log_level: str = "info"
    timeout: int = 30
    
    @classmethod
    def load(cls, config_path: str) -> 'Config':
        """Load configuration from YAML file."""
        path = Path(config_path)
        if not path.exists():
            raise FileNotFoundError(f"Configuration file not found: {config_path}")
        
        with open(path, 'r') as f:
            data = yaml.safe_load(f)
        
        plugin_data = data.get('plugin', {})
        plugin = PluginConfig(
            name=plugin_data.get('name', ''),
            version=plugin_data.get('version', ''),
            description=plugin_data.get('description', '')
        )
        
        config = cls(
            plugin=plugin,
            log_level=data.get('log_level', 'info'),
            timeout=data.get('timeout', 30)
        )
        
        config.validate()
        return config
    
    def validate(self):
        """Validate configuration."""
        if not self.plugin.name:
            raise ValueError("Plugin name is required")
        if not self.plugin.version:
            raise ValueError("Plugin version is required")
        if self.timeout <= 0:
            raise ValueError("Timeout must be positive")
```

### gRPC Handler (src/handlers/plugin_handler.py)
```python
import json
import logging
import time
from typing import Dict, Any

import grpc

from config.config import Config
from services.plugin_service import PluginService
from proto import plugin_pb2, plugin_pb2_grpc


class PluginHandler(plugin_pb2_grpc.PluginServiceServicer):
    def __init__(self, config: Config):
        self.config = config
        self.service = PluginService(config)
        self.logger = logging.getLogger(__name__)
    
    def GetInfo(self, request, context):
        """Get plugin information."""
        self.logger.debug("GetInfo called")
        
        return plugin_pb2.GetInfoResponse(
            info=plugin_pb2.PluginInfo(
                name=self.config.plugin.name,
                version=self.config.plugin.version,
                description=self.config.plugin.description,
                capabilities=["example-capability", "python-processing"]
            )
        )
    
    def HealthCheck(self, request, context):
        """Perform health check."""
        status = self.service.check_health()
        return plugin_pb2.HealthCheckResponse(status=status)
    
    def Execute(self, request, context):
        """Execute a command."""
        start_time = time.time()
        
        self.logger.info(f"Executing command: {request.command} with args: {list(request.args)}")
        
        try:
            result = self.service.execute(request.command, list(request.args))
            
            self.logger.info(f"Command executed successfully in {time.time() - start_time:.3f}s")
            
            return plugin_pb2.ExecuteResponse(
                result=result,
                status=plugin_pb2.ExecuteResponse.SUCCESS
            )
            
        except Exception as e:
            self.logger.error(f"Command execution failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            
            return plugin_pb2.ExecuteResponse(
                result="",
                status=plugin_pb2.ExecuteResponse.ERROR,
                error_message=str(e)
            )
    
    def ListTools(self, request, context):
        """List available MCP tools."""
        tools = [
            plugin_pb2.MCPTool(
                name="process_text",
                description="Process text with various operations",
                schema=json.dumps({
                    "type": "object",
                    "properties": {
                        "text": {"type": "string", "description": "Text to process"},
                        "operation": {
                            "type": "string", 
                            "enum": ["uppercase", "lowercase", "reverse", "word_count"],
                            "description": "Operation to perform"
                        }
                    },
                    "required": ["text", "operation"]
                })
            ),
            plugin_pb2.MCPTool(
                name="analyze_data",
                description="Analyze data and provide statistics",
                schema=json.dumps({
                    "type": "object",
                    "properties": {
                        "data": {"type": "array", "items": {"type": "number"}},
                        "analysis_type": {
                            "type": "string",
                            "enum": ["basic", "detailed"],
                            "default": "basic"
                        }
                    },
                    "required": ["data"]
                })
            )
        ]
        
        return plugin_pb2.ListToolsResponse(tools=tools)
    
    def CallTool(self, request, context):
        """Call an MCP tool."""
        self.logger.info(f"Calling tool: {request.tool_name}")
        
        try:
            if request.tool_name == "process_text":
                return self._handle_process_text(request.arguments)
            elif request.tool_name == "analyze_data":
                return self._handle_analyze_data(request.arguments)
            else:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Tool not found: {request.tool_name}")
                return plugin_pb2.CallToolResponse(
                    status=plugin_pb2.CallToolResponse.NOT_FOUND,
                    error_message=f"Tool not found: {request.tool_name}"
                )
                
        except Exception as e:
            self.logger.error(f"Tool execution failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            
            return plugin_pb2.CallToolResponse(
                status=plugin_pb2.CallToolResponse.ERROR,
                error_message=str(e)
            )
    
    def _handle_process_text(self, arguments: str) -> plugin_pb2.CallToolResponse:
        """Handle process_text tool."""
        try:
            params = json.loads(arguments)
        except json.JSONDecodeError as e:
            return plugin_pb2.CallToolResponse(
                status=plugin_pb2.CallToolResponse.INVALID_ARGS,
                error_message=f"Invalid JSON arguments: {e}"
            )
        
        text = params.get("text", "")
        operation = params.get("operation", "")
        
        result = self.service.process_text(text, operation)
        
        return plugin_pb2.CallToolResponse(
            result=result,
            status=plugin_pb2.CallToolResponse.SUCCESS
        )
    
    def _handle_analyze_data(self, arguments: str) -> plugin_pb2.CallToolResponse:
        """Handle analyze_data tool."""
        try:
            params = json.loads(arguments)
        except json.JSONDecodeError as e:
            return plugin_pb2.CallToolResponse(
                status=plugin_pb2.CallToolResponse.INVALID_ARGS,
                error_message=f"Invalid JSON arguments: {e}"
            )
        
        data = params.get("data", [])
        analysis_type = params.get("analysis_type", "basic")
        
        result = self.service.analyze_data(data, analysis_type)
        
        return plugin_pb2.CallToolResponse(
            result=json.dumps(result),
            status=plugin_pb2.CallToolResponse.SUCCESS
        )
    
    def Shutdown(self, request, context):
        """Shutdown the plugin."""
        self.logger.info("Shutdown requested")
        self.cleanup()
        
        return plugin_pb2.ShutdownResponse(
            status=plugin_pb2.ShutdownResponse.SUCCESS,
            message="Plugin shutdown successfully"
        )
    
    def cleanup(self):
        """Perform cleanup operations."""
        self.service.cleanup()
```

### Business Logic (src/services/plugin_service.py)
```python
import logging
import statistics
from typing import List, Dict, Any, Union


class PluginService:
    def __init__(self, config):
        self.config = config
        self.logger = logging.getLogger(__name__)
    
    def execute(self, command: str, args: List[str]) -> str:
        """Execute a plugin command."""
        self.logger.debug(f"Executing command: {command} with args: {args}")
        
        if command == "hello":
            return self._handle_hello(args)
        elif command == "echo":
            return self._handle_echo(args)
        elif command == "process":
            return self._handle_process(args)
        else:
            raise ValueError(f"Unknown command: {command}")
    
    def _handle_hello(self, args: List[str]) -> str:
        """Handle hello command."""
        name = args[0] if args else "World"
        return f"Hello, {name}! This is {self.config.plugin.name} v{self.config.plugin.version}"
    
    def _handle_echo(self, args: List[str]) -> str:
        """Handle echo command."""
        if not args:
            raise ValueError("Echo command requires at least one argument")
        return " ".join(args)
    
    def _handle_process(self, args: List[str]) -> str:
        """Handle process command."""
        if not args:
            raise ValueError("Process command requires at least one argument")
        
        # Example processing: count words and characters
        text = " ".join(args)
        word_count = len(text.split())
        char_count = len(text)
        
        return f"Processed text: {word_count} words, {char_count} characters"
    
    def process_text(self, text: str, operation: str) -> str:
        """Process text with various operations."""
        if not text:
            raise ValueError("Text cannot be empty")
        
        if operation == "uppercase":
            return text.upper()
        elif operation == "lowercase":
            return text.lower()
        elif operation == "reverse":
            return text[::-1]
        elif operation == "word_count":
            word_count = len(text.split())
            return f"Word count: {word_count}"
        else:
            raise ValueError(f"Unknown operation: {operation}")
    
    def analyze_data(self, data: List[float], analysis_type: str = "basic") -> Dict[str, Any]:
        """Analyze numerical data."""
        if not data:
            raise ValueError("Data cannot be empty")
        
        if not all(isinstance(x, (int, float)) for x in data):
            raise ValueError("All data points must be numbers")
        
        result = {
            "count": len(data),
            "min": min(data),
            "max": max(data),
            "mean": statistics.mean(data)
        }
        
        if analysis_type == "detailed" and len(data) > 1:
            result.update({
                "median": statistics.median(data),
                "stdev": statistics.stdev(data),
                "variance": statistics.variance(data)
            })
        
        return result
    
    def check_health(self) -> int:
        """Perform health check."""
        # Implement health check logic here
        # For this example, we'll always return healthy
        self.logger.debug("Health check performed")
        return 1  # SERVING
    
    def cleanup(self):
        """Perform cleanup operations."""
        self.logger.info("Performing cleanup operations")
        # Implement cleanup logic here:
        # - Close database connections
        # - Clean up temporary files
        # - Cancel background tasks
        # - Release resources
        self.logger.info("Cleanup completed")
```

## Development Setup

### Requirements (requirements.txt)
```txt
grpcio>=1.58.0
grpcio-tools>=1.58.0
grpcio-health-checking>=1.58.0
protobuf>=4.24.0
PyYAML>=6.0.1
pytest>=7.4.0
pytest-asyncio>=0.21.0
pytest-grpc>=0.8.0
black>=23.7.0
flake8>=6.0.0
mypy>=1.5.0
```

### Development Setup Script (scripts/setup.sh)
```bash
#!/bin/bash

set -e

echo "Setting up Python plugin development environment..."

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Upgrade pip
pip install --upgrade pip

# Install dependencies
pip install -r requirements.txt

# Generate protocol buffers
echo "Generating protocol buffers..."
python -m grpc_tools.protoc \
    --proto_path=proto \
    --python_out=src \
    --grpc_python_out=src \
    proto/*.proto

# Fix imports in generated files
find src -name "*_pb2_grpc.py" -exec sed -i 's/import.*_pb2/from . import &/' {} \;

echo "Setup completed! Activate the virtual environment with: source venv/bin/activate"
```

### Build Script (scripts/build.sh)
```bash
#!/bin/bash

set -e

echo "Building Python plugin..."

# Activate virtual environment
source venv/bin/activate

# Generate protocol buffers
python -m grpc_tools.protoc \
    --proto_path=proto \
    --python_out=src \
    --grpc_python_out=src \
    proto/*.proto

# Fix imports
find src -name "*_pb2_grpc.py" -exec sed -i 's/import.*_pb2/from . import &/' {} \;

# Run tests
echo "Running tests..."
python -m pytest tests/ -v

# Check code quality
echo "Checking code quality..."
black --check src/
flake8 src/
mypy src/

# Create distribution
echo "Creating distribution..."
python setup.py sdist bdist_wheel

echo "Build completed successfully!"
```

## Testing

### Test Example (tests/test_plugin.py)
```python
import pytest
import asyncio
from unittest.mock import Mock, patch

from src.config.config import Config, PluginConfig
from src.handlers.plugin_handler import PluginHandler
from src.services.plugin_service import PluginService
from src.proto import plugin_pb2


@pytest.fixture
def config():
    return Config(
        plugin=PluginConfig(
            name="test-plugin",
            version="1.0.0",
            description="Test plugin"
        ),
        log_level="debug",
        timeout=30
    )


@pytest.fixture
def plugin_service(config):
    return PluginService(config)


@pytest.fixture
def plugin_handler(config):
    return PluginHandler(config)


class TestPluginService:
    def test_hello_command(self, plugin_service):
        result = plugin_service.execute("hello", ["Alice"])
        assert "Hello, Alice!" in result
        assert plugin_service.config.plugin.name in result
    
    def test_echo_command(self, plugin_service):
        result = plugin_service.execute("echo", ["test", "message"])
        assert result == "test message"
    
    def test_echo_command_no_args(self, plugin_service):
        with pytest.raises(ValueError, match="requires at least one argument"):
            plugin_service.execute("echo", [])
    
    def test_process_text_uppercase(self, plugin_service):
        result = plugin_service.process_text("hello world", "uppercase")
        assert result == "HELLO WORLD"
    
    def test_process_text_reverse(self, plugin_service):
        result = plugin_service.process_text("hello", "reverse")
        assert result == "olleh"
    
    def test_analyze_data_basic(self, plugin_service):
        data = [1, 2, 3, 4, 5]
        result = plugin_service.analyze_data(data, "basic")
        
        assert result["count"] == 5
        assert result["min"] == 1
        assert result["max"] == 5
        assert result["mean"] == 3.0
    
    def test_analyze_data_detailed(self, plugin_service):
        data = [1, 2, 3, 4, 5]
        result = plugin_service.analyze_data(data, "detailed")
        
        assert "median" in result
        assert "stdev" in result
        assert "variance" in result


class TestPluginHandler:
    def test_get_info(self, plugin_handler):
        request = plugin_pb2.GetInfoRequest()
        response = plugin_handler.GetInfo(request, None)
        
        assert response.info.name == plugin_handler.config.plugin.name
        assert response.info.version == plugin_handler.config.plugin.version
    
    def test_health_check(self, plugin_handler):
        request = plugin_pb2.HealthCheckRequest()
        response = plugin_handler.HealthCheck(request, None)
        
        assert response.status == 1  # SERVING
    
    def test_execute_success(self, plugin_handler):
        request = plugin_pb2.ExecuteRequest(
            command="hello",
            args=["World"]
        )
        response = plugin_handler.Execute(request, None)
        
        assert response.status == plugin_pb2.ExecuteResponse.SUCCESS
        assert "Hello, World!" in response.result
    
    def test_list_tools(self, plugin_handler):
        request = plugin_pb2.ListToolsRequest()
        response = plugin_handler.ListTools(request, None)
        
        assert len(response.tools) > 0
        tool_names = [tool.name for tool in response.tools]
        assert "process_text" in tool_names
        assert "analyze_data" in tool_names


@pytest.mark.asyncio
class TestAsyncOperations:
    async def test_concurrent_requests(self, plugin_handler):
        """Test handling multiple concurrent requests."""
        tasks = []
        
        for i in range(10):
            request = plugin_pb2.ExecuteRequest(
                command="hello",
                args=[f"User{i}"]
            )
            task = asyncio.create_task(
                plugin_handler.Execute(request, None)
            )
            tasks.append(task)
        
        responses = await asyncio.gather(*tasks)
        
        for i, response in enumerate(responses):
            assert response.status == plugin_pb2.ExecuteResponse.SUCCESS
            assert f"User{i}" in response.result
```

## Advanced Features

### Async Operations
```python
import asyncio
import aiohttp
from typing import Optional


class AsyncPluginService(PluginService):
    def __init__(self, config):
        super().__init__(config)
        self.session: Optional[aiohttp.ClientSession] = None
    
    async def start(self):
        """Initialize async resources."""
        self.session = aiohttp.ClientSession()
    
    async def stop(self):
        """Cleanup async resources."""
        if self.session:
            await self.session.close()
    
    async def fetch_data(self, url: str) -> str:
        """Fetch data from external API."""
        if not self.session:
            raise RuntimeError("Service not started")
        
        async with self.session.get(url) as response:
            return await response.text()
    
    async def process_large_file(self, file_path: str) -> Dict[str, Any]:
        """Process large file asynchronously."""
        result = {"lines": 0, "words": 0, "chars": 0}
        
        with open(file_path, 'r') as f:
            async for line in self._async_file_reader(f):
                result["lines"] += 1
                result["words"] += len(line.split())
                result["chars"] += len(line)
                
                # Yield control to other tasks
                await asyncio.sleep(0)
        
        return result
    
    async def _async_file_reader(self, file_obj):
        """Async file reader generator."""
        while True:
            line = file_obj.readline()
            if not line:
                break
            yield line
```

### Data Processing with Pandas
```python
import pandas as pd
import numpy as np
from typing import Dict, Any, List


class DataAnalysisService:
    def analyze_csv(self, file_path: str) -> Dict[str, Any]:
        """Analyze CSV file and return statistics."""
        df = pd.read_csv(file_path)
        
        return {
            "shape": df.shape,
            "columns": df.columns.tolist(),
            "data_types": df.dtypes.to_dict(),
            "missing_values": df.isnull().sum().to_dict(),
            "numeric_summary": df.describe().to_dict(),
            "memory_usage": df.memory_usage(deep=True).sum()
        }
    
    def process_json_data(self, data: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Process JSON data with pandas."""
        df = pd.DataFrame(data)
        
        result = {
            "record_count": len(df),
            "unique_values": {},
            "correlations": {}
        }
        
        # Calculate unique values for each column
        for col in df.columns:
            result["unique_values"][col] = df[col].nunique()
        
        # Calculate correlations for numeric columns
        numeric_cols = df.select_dtypes(include=[np.number]).columns
        if len(numeric_cols) > 1:
            result["correlations"] = df[numeric_cols].corr().to_dict()
        
        return result
```

### Machine Learning Integration
```python
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.cluster import KMeans
import joblib
from typing import List, Dict, Any


class MLService:
    def __init__(self):
        self.models = {}
    
    def cluster_text(self, texts: List[str], n_clusters: int = 3) -> Dict[str, Any]:
        """Cluster text documents using TF-IDF and K-means."""
        if len(texts) < n_clusters:
            raise ValueError(f"Need at least {n_clusters} texts for clustering")
        
        # Vectorize texts
        vectorizer = TfidfVectorizer(stop_words='english', max_features=1000)
        X = vectorizer.fit_transform(texts)
        
        # Perform clustering
        kmeans = KMeans(n_clusters=n_clusters, random_state=42)
        clusters = kmeans.fit_predict(X)
        
        # Prepare results
        result = {
            "clusters": clusters.tolist(),
            "cluster_centers": kmeans.cluster_centers_.tolist(),
            "inertia": kmeans.inertia_,
            "feature_names": vectorizer.get_feature_names_out().tolist()
        }
        
        return result
    
    def save_model(self, model_name: str, model: Any) -> str:
        """Save trained model to disk."""
        file_path = f"models/{model_name}.joblib"
        joblib.dump(model, file_path)
        return file_path
    
    def load_model(self, model_name: str) -> Any:
        """Load trained model from disk."""
        file_path = f"models/{model_name}.joblib"
        return joblib.load(file_path)
```

## Deployment

### Dockerfile
```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY src/ src/
COPY proto/ proto/
COPY plugin.yaml .

# Generate protocol buffers
RUN python -m grpc_tools.protoc \
    --proto_path=proto \
    --python_out=src \
    --grpc_python_out=src \
    proto/*.proto

# Fix imports in generated files
RUN find src -name "*_pb2_grpc.py" -exec sed -i 's/import.*_pb2/from . import &/' {} \;

# Create non-root user
RUN useradd -m -u 1000 plugin
USER plugin

# Expose ports
EXPOSE 50051 50052 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import grpc; from grpc_health.v1 import health_pb2_grpc, health_pb2; \
                   channel = grpc.insecure_channel('localhost:50052'); \
                   stub = health_pb2_grpc.HealthStub(channel); \
                   response = stub.Check(health_pb2.HealthCheckRequest()); \
                   exit(0 if response.status == 1 else 1)"

# Start plugin
CMD ["python", "src/main.py"]
```

### Docker Compose for Development
```yaml
version: '3.8'

services:
  python-plugin:
    build: .
    ports:
      - "50051:50051"
      - "50052:50052"
      - "8080:8080"
    volumes:
      - ./src:/app/src:ro
      - ./config.dev.yaml:/app/config.yaml:ro
      - ./logs:/app/logs
    environment:
      - PYTHONPATH=/app/src
      - LOG_LEVEL=debug
    networks:
      - plugin-network

  plugin-test:
    build:
      context: .
      dockerfile: Dockerfile.test
    volumes:
      - ./tests:/app/tests:ro
      - ./src:/app/src:ro
    depends_on:
      - python-plugin
    networks:
      - plugin-network

networks:
  plugin-network:
    driver: bridge
```

## Best Practices

### Error Handling
```python
import functools
import logging
from typing import Callable, Any


def handle_grpc_errors(func: Callable) -> Callable:
    """Decorator to handle gRPC errors consistently."""
    @functools.wraps(func)
    def wrapper(self, request, context):
        try:
            return func(self, request, context)
        except ValueError as e:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(str(e))
            raise
        except FileNotFoundError as e:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(str(e))
            raise
        except PermissionError as e:
            context.set_code(grpc.StatusCode.PERMISSION_DENIED)
            context.set_details(str(e))
            raise
        except Exception as e:
            logging.error(f"Unexpected error in {func.__name__}: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details("Internal server error")
            raise
    
    return wrapper
```

### Configuration Validation
```python
from pydantic import BaseModel, validator
from typing import Optional, List


class PluginConfig(BaseModel):
    name: str
    version: str
    description: str
    
    @validator('name')
    def name_must_not_be_empty(cls, v):
        if not v.strip():
            raise ValueError('Plugin name cannot be empty')
        return v
    
    @validator('version')
    def version_must_be_semver(cls, v):
        import re
        pattern = r'^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$'
        if not re.match(pattern, v):
            raise ValueError('Version must follow semantic versioning')
        return v


class DatabaseConfig(BaseModel):
    url: str
    username: Optional[str] = None
    password: Optional[str] = None
    pool_size: int = 5
    
    @validator('pool_size')
    def pool_size_must_be_positive(cls, v):
        if v <= 0:
            raise ValueError('Pool size must be positive')
        return v
```

### Logging and Monitoring
```python
import logging
import time
from prometheus_client import Counter, Histogram, start_http_server


# Metrics
REQUEST_COUNT = Counter('plugin_requests_total', 'Total requests', ['method', 'status'])
REQUEST_DURATION = Histogram('plugin_request_duration_seconds', 'Request duration', ['method'])


def setup_monitoring(port: int = 8080):
    """Start Prometheus metrics server."""
    start_http_server(port)
    logging.info(f"Metrics server started on port {port}")


def log_and_monitor(method_name: str):
    """Decorator for logging and monitoring."""
    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            start_time = time.time()
            logger = logging.getLogger(func.__module__)
            
            logger.debug(f"Starting {method_name}")
            
            try:
                result = func(*args, **kwargs)
                REQUEST_COUNT.labels(method=method_name, status='success').inc()
                logger.info(f"Completed {method_name} successfully")
                return result
                
            except Exception as e:
                REQUEST_COUNT.labels(method=method_name, status='error').inc()
                logger.error(f"Error in {method_name}: {e}")
                raise
                
            finally:
                duration = time.time() - start_time
                REQUEST_DURATION.labels(method=method_name).observe(duration)
                logger.debug(f"Finished {method_name} in {duration:.3f}s")
        
        return wrapper
    return decorator
```

## Next Steps

- Study the [template code](template/) for a complete example
- Read [best practices](best-practices.md) for production deployment
- Explore [examples](examples/) for specific use cases
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents

## Resources

- [Python gRPC Documentation](https://grpc.io/docs/languages/python/)
- [Protocol Buffers Python Tutorial](https://developers.google.com/protocol-buffers/docs/pythontutorial)
- [Portunix Plugin API Reference](../../api-reference.md)
- [asyncio Documentation](https://docs.python.org/3/library/asyncio.html)