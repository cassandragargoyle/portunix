# Python Code Style Guidelines

## Purpose
This document defines the Python coding standards for CassandraGargoyle projects, based on PEP 8 and modern Python best practices.

## General Principles

### 1. Follow Python Standards
- Adhere to PEP 8 for style conventions
- Use PEP 257 for docstring conventions
- Embrace Python idioms: \"Pythonic\" code
- Follow the Zen of Python principles

### 2. Code Quality
- Write readable, self-documenting code
- Use type hints for better code clarity
- Prefer composition over inheritance
- Follow DRY (Don't Repeat Yourself) principle

## File and Project Structure

### File Naming
Use snake_case for all file and directory names:
```
✅ package_installer.py
✅ config_manager.py
✅ system_detector.py
❌ PackageInstaller.py
❌ configManager.py
```

### Project Structure
```
project/
├── src/
│   └── cassandragargoyle/
│       └── project/
│           ├── __init__.py
│           ├── config/
│           │   ├── __init__.py
│           │   ├── configuration_manager.py
│           │   └── database_config.py
│           ├── install/
│           │   ├── __init__.py
│           │   ├── package_installer.py
│           │   └── system_detector.py
│           └── util/
│               ├── __init__.py
│               └── file_utils.py
├── tests/
│   ├── __init__.py
│   ├── config/
│   │   └── test_configuration_manager.py
│   └── install/
│       └── test_package_installer.py
├── docs/
├── requirements.txt
├── pyproject.toml
└── README.md
```

### Module Organization
```python
#!/usr/bin/env python3
\"\"\"
Package installer module for cross-platform software installation.

This module provides utilities for installing packages across different
operating systems using their respective package managers.
\"\"\"

# Standard library imports
import os
import sys
from pathlib import Path
from typing import Dict, List, Optional, Union

# Third-party imports
import click
import requests

# Local imports
from cassandragargoyle.project.config import ConfigurationManager
from cassandragargoyle.project.util import FileUtils
```

## Naming Conventions

### Variables and Functions
Use snake_case for variables and functions:
```python
# Variables
package_name = \"python3\"
is_installed = False
config_path = Path(\"/etc/config.yaml\")

# Functions
def install_package(package_name: str) -> bool:
    \"\"\"Install a package using the system package manager.\"\"\"
    pass

def is_package_installed(package_name: str) -> bool:
    \"\"\"Check if a package is currently installed.\"\"\"
    pass
```

### Classes
Use PascalCase for class names:
```python
class PackageInstaller:
    \"\"\"Cross-platform package installer.\"\"\"
    pass

class ConfigurationManager:
    \"\"\"Manages application configuration.\"\"\"
    pass

class SystemDetector:
    \"\"\"Detects operating system and environment details.\"\"\"
    pass
```

### Constants
Use UPPER_CASE with underscores for constants:
```python
# Module-level constants
DEFAULT_CONFIG_PATH = \"/etc/cassandragargoyle/config.yaml\"
MAX_RETRY_ATTEMPTS = 3
SUPPORTED_PACKAGE_MANAGERS = [\"apt\", \"yum\", \"chocolatey\", \"homebrew\"]

# Class constants
class PackageInstaller:
    DEFAULT_TIMEOUT = 300
    PACKAGE_CACHE_SIZE = 1000
```

### Private Members
Use single underscore prefix for internal use:
```python
class PackageInstaller:
    def __init__(self):
        self._config_path = DEFAULT_CONFIG_PATH
        self._is_initialized = False
        self.__internal_state = {}  # Double underscore for name mangling
    
    def _load_configuration(self) -> None:
        \"\"\"Load configuration from file (internal method).\"\"\"
        pass
```

## Type Hints

### Basic Type Hints
Use type hints for all public interfaces:
```python
from typing import Dict, List, Optional, Union, Any, Callable
from pathlib import Path

def install_package(
    package_name: str, 
    version: Optional[str] = None,
    force: bool = False
) -> bool:
    \"\"\"Install a package with optional version specification.\"\"\"
    pass

def get_installed_packages() -> List[str]:
    \"\"\"Get list of installed package names.\"\"\"
    return []

def load_configuration(config_path: Path) -> Dict[str, Any]:
    \"\"\"Load configuration from YAML file.\"\"\"
    return {}
```

### Complex Type Hints
```python
from typing import Dict, List, Optional, Protocol, TypedDict, Union
from dataclasses import dataclass
from enum import Enum

class PackageStatus(Enum):
    \"\"\"Package installation status.\"\"\"
    NOT_INSTALLED = \"not_installed\"
    INSTALLED = \"installed\"
    UPDATE_AVAILABLE = \"update_available\"
    ERROR = \"error\"

@dataclass
class PackageInfo:
    \"\"\"Information about a package.\"\"\"
    name: str
    version: str
    status: PackageStatus
    dependencies: List[str]

class PackageManagerProtocol(Protocol):
    \"\"\"Protocol for package managers.\"\"\"
    
    def install(self, package_name: str) -> bool:
        \"\"\"Install a package.\"\"\"
        ...
    
    def is_installed(self, package_name: str) -> bool:
        \"\"\"Check if package is installed.\"\"\"
        ...

PackageRegistry = Dict[str, PackageInfo]  # Type alias
```

## Code Formatting

### Line Length and Indentation
- Maximum line length: 88 characters (Black formatter default)
- Use 4 spaces for indentation
- No trailing whitespace

### String Formatting
Prefer f-strings for string formatting:
```python
# Good
package_name = \"python3\"
version = \"3.9.0\"
message = f\"Installing {package_name} version {version}\"

# For complex formatting or when f-strings aren't suitable
template = \"Package {name} (v{version}) installed successfully\"
message = template.format(name=package_name, version=version)

# Avoid (old style)
message = \"Installing %s version %s\" % (package_name, version)
```

### List and Dictionary Formatting
```python
# Short lists/dicts on single line
packages = [\"python3\", \"git\", \"vim\"]
config = {\"timeout\": 300, \"retry_count\": 3}

# Long lists/dicts with trailing comma
supported_managers = [
    \"apt\",
    \"yum\",
    \"dnf\",
    \"chocolatey\",
    \"homebrew\",
    \"pip\",
]

default_config = {
    \"package_manager\": \"auto\",
    \"timeout\": 300,
    \"retry_count\": 3,
    \"enable_logging\": True,
    \"log_level\": \"INFO\",
}
```

## Class Design

### Class Structure
Organize class members in this order:
1. Class docstring
2. Class variables and constants
3. `__init__` method
4. Special methods (`__str__`, `__repr__`, etc.)
5. Public methods
6. Private methods

```python
class PackageInstaller:
    \"\"\"
    Cross-platform package installer.
    
    This class provides a unified interface for installing software packages
    across different operating systems. It automatically detects the appropriate
    package manager and handles platform-specific installation procedures.
    
    Example:
        installer = PackageInstaller()
        installer.install_package(\"python3\")
    \"\"\"
    
    # Class constants
    DEFAULT_TIMEOUT = 300
    SUPPORTED_PLATFORMS = [\"linux\", \"darwin\", \"win32\"]
    
    def __init__(self, config_path: Optional[Path] = None):
        \"\"\"Initialize the package installer.
        
        Args:
            config_path: Optional path to configuration file.
        \"\"\"
        self._config_path = config_path or Path(DEFAULT_CONFIG_PATH)
        self._config: Dict[str, Any] = {}
        self._is_initialized = False
        self._package_managers: Dict[str, PackageManagerProtocol] = {}
    
    def __repr__(self) -> str:
        \"\"\"Return string representation of the installer.\"\"\"
        return f\"PackageInstaller(config_path={self._config_path})\"
    
    def __str__(self) -> str:
        \"\"\"Return human-readable string representation.\"\"\"
        status = \"initialized\" if self._is_initialized else \"not initialized\"
        return f\"PackageInstaller ({status})\"
    
    def install_package(self, package_name: str, version: Optional[str] = None) -> bool:
        \"\"\"Install a package using the appropriate package manager.
        
        Args:
            package_name: Name of the package to install.
            version: Optional specific version to install.
            
        Returns:
            True if installation was successful, False otherwise.
            
        Raises:
            PackageInstallationError: If installation fails.
            ValueError: If package_name is invalid.
        \"\"\"
        if not self._is_initialized:
            self._initialize()
        
        return self._install_package_impl(package_name, version)
    
    def _initialize(self) -> None:
        \"\"\"Initialize the installer (internal method).\"\"\"
        self._load_configuration()
        self._detect_package_managers()
        self._is_initialized = True
    
    def _install_package_impl(self, package_name: str, version: Optional[str]) -> bool:
        \"\"\"Internal implementation of package installation.\"\"\"
        # Implementation details
        return True
```

### Property Usage
Use properties for getter/setter behavior:
```python
class ConfigurationManager:
    \"\"\"Manages application configuration.\"\"\"
    
    def __init__(self, config_path: Path):
        self._config_path = config_path
        self._config: Dict[str, Any] = {}
        self._is_loaded = False
    
    @property
    def config_path(self) -> Path:
        \"\"\"Get the configuration file path.\"\"\"
        return self._config_path
    
    @config_path.setter
    def config_path(self, path: Path) -> None:
        \"\"\"Set the configuration file path.\"\"\"
        self._config_path = path
        self._is_loaded = False  # Reset loaded state
    
    @property
    def is_loaded(self) -> bool:
        \"\"\"Check if configuration is loaded.\"\"\"
        return self._is_loaded
```

## Error Handling

### Exception Hierarchy
Create domain-specific exceptions:
```python
class CassandraGargoyleError(Exception):
    \"\"\"Base exception for CassandraGargoyle project.\"\"\"
    pass

class ConfigurationError(CassandraGargoyleError):
    \"\"\"Raised when configuration is invalid or cannot be loaded.\"\"\"
    pass

class PackageInstallationError(CassandraGargoyleError):
    \"\"\"Raised when package installation fails.\"\"\"
    
    def __init__(self, package_name: str, reason: str, exit_code: Optional[int] = None):
        self.package_name = package_name
        self.reason = reason
        self.exit_code = exit_code
        super().__init__(f\"Failed to install '{package_name}': {reason}\")

class UnsupportedPlatformError(CassandraGargoyleError):
    \"\"\"Raised when operation is not supported on current platform.\"\"\"
    pass
```

### Exception Handling Patterns
```python
import logging
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)

def load_configuration(config_path: Path) -> Dict[str, Any]:
    \"\"\"Load configuration from YAML file.
    
    Args:
        config_path: Path to the configuration file.
        
    Returns:
        Dictionary containing configuration data.
        
    Raises:
        ConfigurationError: If file cannot be read or parsed.
    \"\"\"
    try:
        with open(config_path, 'r', encoding='utf-8') as file:
            import yaml
            return yaml.safe_load(file)
    except FileNotFoundError:
        raise ConfigurationError(f\"Configuration file not found: {config_path}\")
    except yaml.YAMLError as e:
        raise ConfigurationError(f\"Invalid YAML in {config_path}: {e}\") from e
    except Exception as e:
        logger.exception(\"Unexpected error loading configuration\")
        raise ConfigurationError(f\"Failed to load configuration: {e}\") from e

def install_package_safe(package_name: str) -> bool:
    \"\"\"Install package with error handling and logging.\"\"\"
    try:
        installer = PackageInstaller()
        result = installer.install_package(package_name)
        logger.info(f\"Successfully installed package: {package_name}\")
        return result
    except PackageInstallationError as e:
        logger.error(f\"Package installation failed: {e}\")
        return False
    except Exception as e:
        logger.exception(\"Unexpected error during package installation\")
        return False
```

## Documentation

### Module Docstrings
```python
\"\"\"
Package installer module for cross-platform software installation.

This module provides utilities for installing packages across different
operating systems using their respective package managers (APT, Chocolatey,
Homebrew, etc.).

The main class `PackageInstaller` automatically detects the platform and
uses the appropriate package manager. It supports:

- Linux: APT, YUM, DNF
- macOS: Homebrew, MacPorts
- Windows: Chocolatey, Winget

Example:
    Basic package installation:
    
    >>> from cassandragargoyle.project.install import PackageInstaller
    >>> installer = PackageInstaller()
    >>> installer.install_package(\"python3\")
    True

Attributes:
    DEFAULT_CONFIG_PATH: Default path to configuration file.
    SUPPORTED_PLATFORMS: List of supported operating systems.
\"\"\"
```

### Function and Method Docstrings
Use Google-style docstrings:
```python
def install_package(
    package_name: str, 
    version: Optional[str] = None,
    force: bool = False,
    dry_run: bool = False
) -> bool:
    \"\"\"Install a package using the system package manager.
    
    This method attempts to install the specified package using the most
    appropriate package manager for the current platform. It supports
    version specification and forced installation.
    
    Args:
        package_name: The name of the package to install. Must be a valid
            package name for the target package manager.
        version: Optional version specification. Format depends on the
            package manager (e.g., \"3.9.0\" for exact version, \">=3.8\"
            for minimum version).
        force: If True, force reinstallation even if package is already
            installed. Defaults to False.
        dry_run: If True, simulate installation without making changes.
            Defaults to False.
    
    Returns:
        True if installation was successful or simulated successfully,
        False otherwise.
    
    Raises:
        ValueError: If package_name is empty or contains invalid characters.
        PackageInstallationError: If installation fails due to package
            manager errors, network issues, or permission problems.
        UnsupportedPlatformError: If current platform is not supported.
    
    Example:
        >>> installer = PackageInstaller()
        >>> installer.install_package(\"python3\", version=\"3.9.0\")
        True
        >>> installer.install_package(\"nonexistent-package\")
        False
    
    Note:
        This method requires appropriate permissions (sudo on Unix-like
        systems, administrator on Windows) for system-wide package
        installation.
    \"\"\"
    pass
```

## Testing

### Test Structure
Use pytest for testing with descriptive test names:
```python
import pytest
from pathlib import Path
from unittest.mock import Mock, patch

from cassandragargoyle.project.install import PackageInstaller
from cassandragargoyle.project.install.exceptions import (
    PackageInstallationError,
    UnsupportedPlatformError
)

class TestPackageInstaller:
    \"\"\"Test cases for PackageInstaller class.\"\"\"
    
    @pytest.fixture
    def installer(self):
        \"\"\"Create a PackageInstaller instance for testing.\"\"\"
        return PackageInstaller()
    
    @pytest.fixture
    def mock_config_file(self, tmp_path):
        \"\"\"Create a temporary configuration file.\"\"\"
        config_file = tmp_path / \"config.yaml\"
        config_file.write_text(\"timeout: 300\\nretry_count: 3\\n\")
        return config_file
    
    def test_install_package_success(self, installer):
        \"\"\"Test successful package installation.\"\"\"
        # Arrange
        package_name = \"test-package\"
        
        # Act
        with patch.object(installer, '_install_package_impl', return_value=True):
            result = installer.install_package(package_name)
        
        # Assert
        assert result is True
    
    def test_install_package_with_invalid_name_raises_error(self, installer):
        \"\"\"Test that invalid package name raises ValueError.\"\"\"
        # Arrange
        invalid_names = [\"\", \" \", \"package;rm -rf /\"]
        
        # Act & Assert
        for invalid_name in invalid_names:
            with pytest.raises(ValueError, match=\"Invalid package name\"):
                installer.install_package(invalid_name)
    
    def test_install_package_failure_raises_exception(self, installer):
        \"\"\"Test that installation failure raises PackageInstallationError.\"\"\"
        # Arrange
        package_name = \"nonexistent-package\"
        
        # Act & Assert
        with patch.object(installer, '_install_package_impl', 
                         side_effect=PackageInstallationError(package_name, \"Not found\")):
            with pytest.raises(PackageInstallationError):
                installer.install_package(package_name)
    
    @pytest.mark.parametrize(\"platform,expected_manager\", [
        (\"linux\", \"apt\"),
        (\"darwin\", \"homebrew\"),
        (\"win32\", \"chocolatey\"),
    ])
    def test_detect_package_manager_by_platform(self, installer, platform, expected_manager):
        \"\"\"Test package manager detection based on platform.\"\"\"
        # Arrange & Act
        with patch('sys.platform', platform):
            manager = installer._detect_package_manager()
        
        # Assert
        assert manager == expected_manager
    
    def test_configuration_loading(self, installer, mock_config_file):
        \"\"\"Test configuration loading from file.\"\"\"
        # Arrange
        installer._config_path = mock_config_file
        
        # Act
        installer._load_configuration()
        
        # Assert
        assert installer._config[\"timeout\"] == 300
        assert installer._config[\"retry_count\"] == 3
```

### Test Fixtures and Utilities
```python
# conftest.py
import pytest
import tempfile
from pathlib import Path
from unittest.mock import Mock

@pytest.fixture
def temp_directory():
    \"\"\"Create a temporary directory for testing.\"\"\"
    with tempfile.TemporaryDirectory() as temp_dir:
        yield Path(temp_dir)

@pytest.fixture
def mock_package_manager():
    \"\"\"Create a mock package manager for testing.\"\"\"
    manager = Mock()
    manager.install.return_value = True
    manager.is_installed.return_value = False
    manager.list_installed.return_value = []
    return manager

@pytest.fixture(autouse=True)
def reset_logging():
    \"\"\"Reset logging configuration for each test.\"\"\"
    import logging
    # Reset handlers and level
    logging.getLogger().handlers = []
    logging.getLogger().setLevel(logging.WARNING)
```

## Logging

### Logging Configuration
```python
import logging
import sys
from pathlib import Path

def setup_logging(log_level: str = \"INFO\", log_file: Optional[Path] = None) -> None:
    \"\"\"Setup logging configuration for the application.
    
    Args:
        log_level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL).
        log_file: Optional file path for log output.
    \"\"\"
    # Create formatter
    formatter = logging.Formatter(
        fmt=\"%(asctime)s - %(name)s - %(levelname)s - %(message)s\",
        datefmt=\"%Y-%m-%d %H:%M:%S\"
    )
    
    # Setup console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setFormatter(formatter)
    
    # Setup file handler if specified
    handlers = [console_handler]
    if log_file:
        file_handler = logging.FileHandler(log_file)
        file_handler.setFormatter(formatter)
        handlers.append(file_handler)
    
    # Configure root logger
    logging.basicConfig(
        level=getattr(logging, log_level.upper()),
        handlers=handlers
    )

# In modules, use module-level loggers
logger = logging.getLogger(__name__)

class PackageInstaller:
    \"\"\"Package installer with logging.\"\"\"
    
    def __init__(self):
        self.logger = logging.getLogger(f\"{__name__}.{self.__class__.__name__}\")
    
    def install_package(self, package_name: str) -> bool:
        \"\"\"Install package with logging.\"\"\"
        self.logger.info(f\"Starting installation of package: {package_name}\")
        
        try:
            result = self._install_package_impl(package_name)
            if result:
                self.logger.info(f\"Successfully installed package: {package_name}\")
            else:
                self.logger.warning(f\"Failed to install package: {package_name}\")
            return result
        except Exception as e:
            self.logger.error(f\"Exception during installation of {package_name}: {e}\", 
                            exc_info=True)
            raise
```

## Configuration Management

### Configuration with dataclasses
```python
from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, List, Optional
import yaml

@dataclass
class PackageManagerConfig:
    \"\"\"Configuration for a package manager.\"\"\"
    name: str
    command: str
    install_args: List[str] = field(default_factory=list)
    timeout: int = 300

@dataclass
class ApplicationConfig:
    \"\"\"Main application configuration.\"\"\"
    package_managers: Dict[str, PackageManagerConfig] = field(default_factory=dict)
    default_timeout: int = 300
    retry_count: int = 3
    log_level: str = \"INFO\"
    log_file: Optional[Path] = None
    
    @classmethod
    def from_yaml(cls, config_path: Path) -> \"ApplicationConfig\":
        \"\"\"Load configuration from YAML file.\"\"\"
        with open(config_path, 'r', encoding='utf-8') as file:
            data = yaml.safe_load(file)
        
        # Convert package manager configs
        package_managers = {}
        for name, pm_data in data.get(\"package_managers\", {}).items():
            package_managers[name] = PackageManagerConfig(
                name=name,
                command=pm_data[\"command\"],
                install_args=pm_data.get(\"install_args\", []),
                timeout=pm_data.get(\"timeout\", 300)
            )
        
        return cls(
            package_managers=package_managers,
            default_timeout=data.get(\"default_timeout\", 300),
            retry_count=data.get(\"retry_count\", 3),
            log_level=data.get(\"log_level\", \"INFO\"),
            log_file=Path(data[\"log_file\"]) if data.get(\"log_file\") else None
        )
    
    def to_yaml(self, config_path: Path) -> None:
        \"\"\"Save configuration to YAML file.\"\"\"
        data = {
            \"default_timeout\": self.default_timeout,
            \"retry_count\": self.retry_count,
            \"log_level\": self.log_level,
            \"package_managers\": {
                name: {
                    \"command\": pm.command,
                    \"install_args\": pm.install_args,
                    \"timeout\": pm.timeout
                }
                for name, pm in self.package_managers.items()
            }
        }
        
        if self.log_file:
            data[\"log_file\"] = str(self.log_file)
        
        with open(config_path, 'w', encoding='utf-8') as file:
            yaml.dump(data, file, default_flow_style=False, indent=2)
```

## Async Programming

### Async/Await Patterns
```python
import asyncio
import aiohttp
from typing import List, Optional
import logging

logger = logging.getLogger(__name__)

class AsyncPackageInstaller:
    \"\"\"Asynchronous package installer for concurrent operations.\"\"\"
    
    def __init__(self, max_concurrent: int = 5):
        self.max_concurrent = max_concurrent
        self._semaphore = asyncio.Semaphore(max_concurrent)
    
    async def install_package(self, package_name: str) -> bool:
        \"\"\"Install a single package asynchronously.\"\"\"
        async with self._semaphore:
            logger.info(f\"Starting async installation: {package_name}\")
            
            # Simulate package installation
            await asyncio.sleep(1)  # Replace with actual installation logic
            
            logger.info(f\"Completed installation: {package_name}\")
            return True
    
    async def install_packages(self, package_names: List[str]) -> Dict[str, bool]:
        \"\"\"Install multiple packages concurrently.\"\"\"
        tasks = [
            asyncio.create_task(self.install_package(name), name=f\"install_{name}\")
            for name in package_names
        ]
        
        results = {}
        for task in asyncio.as_completed(tasks):
            try:
                package_name = task.get_name().replace(\"install_\", \"\")
                results[package_name] = await task
            except Exception as e:
                logger.error(f\"Task failed: {e}\")
                results[task.get_name().replace(\"install_\", \"\")] = False
        
        return results
    
    async def check_package_availability(self, package_name: str) -> bool:
        \"\"\"Check if package is available for installation.\"\"\"
        async with aiohttp.ClientSession() as session:
            # Example: Check package repository
            url = f\"https://api.example.com/packages/{package_name}\"
            try:
                async with session.get(url, timeout=10) as response:
                    return response.status == 200
            except asyncio.TimeoutError:
                logger.warning(f\"Timeout checking availability: {package_name}\")
                return False
            except Exception as e:
                logger.error(f\"Error checking availability: {e}\")
                return False

# Usage example
async def main():
    \"\"\"Main async function demonstrating usage.\"\"\"
    installer = AsyncPackageInstaller(max_concurrent=3)
    
    packages = [\"python3\", \"git\", \"vim\", \"nodejs\", \"docker\"]
    
    # Install packages concurrently
    results = await installer.install_packages(packages)
    
    # Print results
    for package, success in results.items():
        status = \"✓\" if success else \"✗\"
        print(f\"{status} {package}\")

if __name__ == \"__main__\":
    asyncio.run(main())
```

## CLI Development

### Click Framework Usage
```python
import click
from pathlib import Path
from typing import Optional

from cassandragargoyle.project.install import PackageInstaller
from cassandragargoyle.project.config import ApplicationConfig

@click.group()
@click.option('--config', '-c', 'config_path', type=click.Path(exists=True, path_type=Path),
              help='Path to configuration file')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose logging')
@click.pass_context
def cli(ctx: click.Context, config_path: Optional[Path], verbose: bool):
    \"\"\"CassandraGargoyle Project CLI Tool.\"\"\"
    # Setup logging
    log_level = \"DEBUG\" if verbose else \"INFO\"
    setup_logging(log_level)
    
    # Load configuration
    config = ApplicationConfig.from_yaml(config_path) if config_path else ApplicationConfig()
    
    # Store in context for subcommands
    ctx.ensure_object(dict)
    ctx.obj['config'] = config
    ctx.obj['installer'] = PackageInstaller(config)

@cli.command()
@click.argument('package_name')
@click.option('--version', help='Specific version to install')
@click.option('--force', is_flag=True, help='Force reinstallation')
@click.pass_context
def install(ctx: click.Context, package_name: str, version: Optional[str], force: bool):
    \"\"\"Install a package.\"\"\"
    installer: PackageInstaller = ctx.obj['installer']
    
    try:
        click.echo(f\"Installing {package_name}{'=' + version if version else ''}...\")
        
        success = installer.install_package(package_name, version=version, force=force)
        
        if success:
            click.echo(click.style(f\"✓ Successfully installed {package_name}\", fg='green'))
        else:
            click.echo(click.style(f\"✗ Failed to install {package_name}\", fg='red'))
            ctx.exit(1)
            
    except Exception as e:
        click.echo(click.style(f\"Error: {e}\", fg='red'))
        ctx.exit(1)

@cli.command()
@click.pass_context
def list_installed(ctx: click.Context):
    \"\"\"List installed packages.\"\"\"
    installer: PackageInstaller = ctx.obj['installer']
    
    packages = installer.get_installed_packages()
    
    if packages:
        click.echo(\"Installed packages:\")
        for package in sorted(packages):
            click.echo(f\"  • {package}\")
    else:
        click.echo(\"No packages installed.\")

if __name__ == '__main__':
    cli()
```

## Package Distribution

### pyproject.toml Configuration
```toml
[build-system]
requires = [\"hatchling\"]
build-backend = \"hatchling.build\"

[project]
name = \"cassandragargoyle-project\"
version = \"1.0.0\"
description = \"Cross-platform package installer for CassandraGargoyle projects\"
readme = \"README.md\"
license = {file = \"LICENSE\"}
authors = [
    {name = \"CassandraGargoyle Team\", email = \"team@cassandragargoyle.com\"},
]
maintainers = [
    {name = \"CassandraGargoyle Team\", email = \"team@cassandragargoyle.com\"},
]
classifiers = [
    \"Development Status :: 4 - Beta\",
    \"Environment :: Console\",
    \"Intended Audience :: Developers\",
    \"License :: OSI Approved :: MIT License\",
    \"Operating System :: OS Independent\",
    \"Programming Language :: Python :: 3\",
    \"Programming Language :: Python :: 3.8\",
    \"Programming Language :: Python :: 3.9\",
    \"Programming Language :: Python :: 3.10\",
    \"Programming Language :: Python :: 3.11\",
    \"Topic :: Software Development :: Libraries :: Python Modules\",
    \"Topic :: System :: Installation/Setup\",
]
keywords = [\"package\", \"installer\", \"cross-platform\", \"automation\"]
dependencies = [
    \"click>=8.0.0\",
    \"pyyaml>=6.0\",
    \"requests>=2.25.0\",
]
requires-python = \">=3.8\"

[project.optional-dependencies]
dev = [
    \"pytest>=7.0.0\",
    \"pytest-cov>=4.0.0\",
    \"pytest-asyncio>=0.21.0\",
    \"black>=22.0.0\",
    \"isort>=5.10.0\",
    \"mypy>=1.0.0\",
    \"flake8>=5.0.0\",
    \"pre-commit>=2.20.0\",
]
docs = [
    \"sphinx>=5.0.0\",
    \"sphinx-rtd-theme>=1.0.0\",
]

[project.scripts]
cg-installer = \"cassandragargoyle.project.cli:cli\"

[project.urls]
Homepage = \"https://github.com/cassandragargoyle/project\"
Repository = \"https://github.com/cassandragargoyle/project.git\"
Issues = \"https://github.com/cassandragargoyle/project/issues\"
Documentation = \"https://cassandragargoyle-project.readthedocs.io\"

[tool.black]
line-length = 88
target-version = ['py38', 'py39', 'py310', 'py311']
include = '\\.pyi?$'

[tool.isort]
profile = \"black\"
line_length = 88
multi_line_output = 3

[tool.mypy]
python_version = \"3.8\"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
disallow_incomplete_defs = true
check_untyped_defs = true
disallow_untyped_decorators = true
no_implicit_optional = true
warn_redundant_casts = true
warn_unused_ignores = true
warn_no_return = true
warn_unreachable = true
strict_equality = true

[tool.pytest.ini_options]
minversion = \"7.0\"
addopts = \"-ra -q --cov=src --cov-report=term-missing --cov-report=html\"
testpaths = [\"tests\"]
python_files = [\"test_*.py\", \"*_test.py\"]
python_classes = [\"Test*\"]
python_functions = [\"test_*\"]

[tool.coverage.run]
source = [\"src\"]
omit = [\"tests/*\", \"*/test_*\"]

[tool.coverage.report]
exclude_lines = [
    \"pragma: no cover\",
    \"def __repr__\",
    \"raise AssertionError\",
    \"raise NotImplementedError\",
    \"if __name__ == .__main__.:\",
]
```

## Code Quality Tools

### Pre-commit Configuration
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      - id: check-merge-conflict
      - id: check-case-conflict

  - repo: https://github.com/psf/black
    rev: 23.3.0
    hooks:
      - id: black

  - repo: https://github.com/pycqa/isort
    rev: 5.12.0
    hooks:
      - id: isort

  - repo: https://github.com/pycqa/flake8
    rev: 6.0.0
    hooks:
      - id: flake8
        additional_dependencies: [flake8-docstrings, flake8-typing-imports]

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.3.0
    hooks:
      - id: mypy
        additional_dependencies: [types-PyYAML, types-requests]
```

### Makefile for Development
```makefile
.PHONY: install dev-install test lint format type-check clean docs

# Install package
install:
\tpip install .

# Install development dependencies
dev-install:
\tpip install -e \".[dev]\"
\tpre-commit install

# Run tests
test:
\tpytest

# Run linting
lint:
\tflake8 src tests
\tisort --check-only src tests
\tblack --check src tests

# Format code
format:
\tisort src tests
\tblack src tests

# Type checking
type-check:
\tmypy src

# Clean build artifacts
clean:
\trm -rf build/
\trm -rf dist/
\trm -rf *.egg-info/
\tfind . -type d -name __pycache__ -delete
\tfind . -type f -name \"*.pyc\" -delete

# Build documentation
docs:
\tcd docs && make html

# Run all quality checks
check: lint type-check test

# Prepare for commit
prepare: format check
```

---

**Note**: These guidelines should be adapted based on specific project requirements and team preferences. Regular updates ensure alignment with evolving Python best practices.

*Created: 2025-08-23*
*Last updated: 2025-08-23*