# Python Testing Guidelines

## Purpose
This document defines the testing strategy, structure, and best practices for Python projects in the CassandraGargoyle ecosystem.

## Testing Philosophy

### Testing Pyramid
1. **Unit Tests** (70%) - Fast, isolated tests for individual functions/classes
2. **Integration Tests** (20%) - Test component interactions and external services
3. **End-to-End Tests** (10%) - Full application workflow and system tests

### Key Principles
- Follow Test-Driven Development (TDD) when applicable
- Tests should be fast, reliable, and deterministic
- Each test should verify one specific behavior
- Use descriptive test names that explain the scenario
- Mock external dependencies and I/O operations
- **Container Testing**: Use Podman for all container-based testing (preferred over Docker)

## Testing Environment

### Virtual Environment
Python tests should be executed within a dedicated virtual environment to ensure isolation and reproducibility:
- **Location**: `test/venv` - Virtual environment for testing
- **Activation**: 
  - Linux/macOS: `source test/venv/bin/activate`
  - Windows: `test\venv\Scripts\activate`
- **Purpose**: Isolates test dependencies from system packages
- **Setup**: Create and configure before running tests

### Environment Setup
```bash
# Create virtual environment for testing
python -m venv test/venv

# Activate virtual environment
source test/venv/bin/activate  # Linux/macOS
# or
test\venv\Scripts\activate  # Windows

# Install test dependencies
pip install -e ".[test]"
```

## Project Structure

### Directory Layout
```
project/
├── src/
│   └── cassandragargoyle/
│       └── project/
│           ├── __init__.py
│           ├── config/
│           │   ├── __init__.py
│           │   ├── manager.py
│           │   └── loader.py
│           ├── install/
│           │   ├── __init__.py
│           │   ├── installer.py
│           │   └── detector.py
│           └── util/
│               ├── __init__.py
│               └── file_utils.py
├── test/
│   └── venv/              # Virtual environment for testing
├── tests/
│   ├── __init__.py
│   ├── unit/
│   │   ├── __init__.py
│   │   ├── config/
│   │   │   ├── __init__.py
│   │   │   ├── test_manager.py
│   │   │   └── test_loader.py
│   │   ├── install/
│   │   │   ├── __init__.py
│   │   │   ├── test_installer.py
│   │   │   └── test_detector.py
│   │   └── util/
│   │       ├── __init__.py
│   │       └── test_file_utils.py
│   ├── integration/
│   │   ├── __init__.py
│   │   ├── test_install_workflow.py
│   │   └── test_config_integration.py
│   ├── e2e/
│   │   ├── __init__.py
│   │   └── test_full_workflow.py
│   ├── fixtures/
│   │   ├── configs/
│   │   │   ├── valid_config.yaml
│   │   │   └── invalid_config.yaml
│   │   ├── data/
│   │   │   └── sample_packages.json
│   │   └── responses/
│   │       └── api_responses.json
│   ├── mocks/
│   │   ├── __init__.py
│   │   └── mock_services.py
│   └── utils/
│       ├── __init__.py
│       ├── test_helpers.py
│       └── fixtures.py
├── scripts/
│   ├── test.sh
│   ├── test-unit.sh
│   ├── test-integration.sh
│   └── coverage.sh
├── pyproject.toml
├── requirements-dev.txt
└── tox.ini
```

## Test Dependencies

### pyproject.toml Configuration
```toml
[project.optional-dependencies]
test = [
    \"pytest>=7.4.0\",
    \"pytest-cov>=4.1.0\",
    \"pytest-mock>=3.11.0\",
    \"pytest-asyncio>=0.21.0\",
    \"pytest-xdist>=3.3.0\",
    \"pytest-benchmark>=4.0.0\",
    \"hypothesis>=6.82.0\",
    \"factory-boy>=3.3.0\",
    \"responses>=0.23.0\",
    \"httpx>=0.24.0\",
    \"testcontainers>=3.7.0\",
    \"freezegun>=1.2.0\",
]

[tool.pytest.ini_options]
minversion = \"7.0\"
addopts = \"-ra -q --strict-markers --strict-config\"
testpaths = [\"tests\"]
markers = [
    \"unit: Unit tests\",
    \"integration: Integration tests\",
    \"e2e: End-to-end tests\",
    \"slow: Slow tests\",
    \"network: Tests requiring network access\",
    \"docker: Tests requiring Docker\",
]
python_files = [\"test_*.py\", \"*_test.py\"]
python_classes = [\"Test*\"]
python_functions = [\"test_*\"]

[tool.coverage.run]
source = [\"src\"]
omit = [
    \"tests/*\",
    \"*/test_*.py\",
    \"*/__pycache__/*\",
    \"*/venv/*\",
    \"*/virtualenv/*\",
]

[tool.coverage.report]
exclude_lines = [
    \"pragma: no cover\",
    \"def __repr__\",
    \"raise AssertionError\",
    \"raise NotImplementedError\",
    \"if __name__ == .__main__.:\",
    \"if TYPE_CHECKING:\",
]
show_missing = true
precision = 2

[tool.coverage.html]
directory = \"htmlcov\"
```

## Test Naming Conventions

### Test Files
- Unit tests: `test_module_name.py`
- Integration tests: `test_feature_integration.py`
- End-to-end tests: `test_workflow_e2e.py`

### Test Functions and Classes
Use descriptive names that explain the scenario:
```python
def test_should_install_package_when_valid_name_provided():
    # Test implementation
    pass

def test_should_raise_error_when_package_name_is_empty():
    # Test implementation
    pass

class TestPackageInstaller:
    def test_should_detect_os_correctly_on_linux(self):
        # Test implementation
        pass
    
    def test_should_handle_network_timeout_gracefully(self):
        # Test implementation
        pass
```

### Test Structure Pattern
Use **Arrange/Act/Assert** (AAA) pattern:
```python
def test_should_load_configuration_from_valid_file():
    # Arrange
    config_path = \"tests/fixtures/configs/valid_config.yaml\"
    loader = ConfigLoader()
    
    # Act
    config = loader.load_configuration(config_path)
    
    # Assert
    assert config is not None
    assert config.timeout == 300
    assert config.retry_count == 3
```

## Unit Testing with pytest

### Basic Unit Test Example
```python
import pytest
from unittest.mock import Mock, patch, MagicMock
from pathlib import Path

from cassandragargoyle.project.install.installer import PackageInstaller
from cassandragargoyle.project.install.exceptions import PackageInstallationError


class TestPackageInstaller:
    \"\"\"Unit tests for PackageInstaller class.\"\"\"
    
    @pytest.fixture
    def mock_package_manager(self):
        \"\"\"Create a mock package manager for testing.\"\"\"
        return Mock()
    
    @pytest.fixture
    def installer(self, mock_package_manager):
        \"\"\"Create PackageInstaller instance with mocked dependencies.\"\"\"
        return PackageInstaller(package_manager=mock_package_manager)
    
    def test_should_install_package_successfully_when_valid_name_provided(
        self, installer, mock_package_manager
    ):
        # Arrange
        package_name = \"python3\"
        mock_package_manager.install.return_value = True
        
        # Act
        result = installer.install_package(package_name)
        
        # Assert
        assert result is True
        mock_package_manager.install.assert_called_once_with(package_name)
    
    def test_should_raise_error_when_package_name_is_empty(self, installer):
        # Act & Assert
        with pytest.raises(ValueError, match=\"Package name cannot be empty\"):
            installer.install_package(\"\")
    
    def test_should_raise_error_when_package_name_is_none(self, installer):
        # Act & Assert
        with pytest.raises(ValueError, match=\"Package name cannot be None\"):
            installer.install_package(None)
    
    @pytest.mark.parametrize(\"invalid_name\", [
        \"\",
        \"   \",
        \"package;rm -rf /\",
        \"pkg with spaces\",
        \"a\" * 256,  # Too long
    ])
    def test_should_raise_error_for_invalid_package_names(self, installer, invalid_name):
        # Act & Assert
        with pytest.raises(ValueError):
            installer.install_package(invalid_name)
    
    def test_should_retry_installation_on_failure(self, installer, mock_package_manager):
        # Arrange
        package_name = \"unstable-package\"
        mock_package_manager.install.side_effect = [
            Exception(\"First attempt fails\"),
            Exception(\"Second attempt fails\"),
            True  # Third attempt succeeds
        ]
        
        # Act
        result = installer.install_package_with_retry(package_name, max_retries=3)
        
        # Assert
        assert result is True
        assert mock_package_manager.install.call_count == 3
    
    def test_should_handle_installation_timeout(self, installer, mock_package_manager):
        # Arrange
        import time
        def slow_install(package_name):
            time.sleep(0.1)  # Simulate slow operation
            return True
        
        mock_package_manager.install.side_effect = slow_install
        
        # Act & Assert
        with pytest.raises(TimeoutError):
            installer.install_package_with_timeout(\"slow-package\", timeout=0.05)
```

### Parameterized Tests
```python
import pytest

class TestOSDetection:
    \"\"\"Tests for operating system detection.\"\"\"
    
    @pytest.mark.parametrize(\"platform_system,expected_os\", [
        (\"Linux\", \"linux\"),
        (\"Windows\", \"windows\"),
        (\"Darwin\", \"macos\"),
    ])
    def test_should_detect_os_correctly(self, platform_system, expected_os):
        with patch(\"platform.system\", return_value=platform_system):
            from cassandragargoyle.project.install.detector import SystemDetector
            
            detector = SystemDetector()
            result = detector.detect_os()
            
            assert result == expected_os
    
    @pytest.mark.parametrize(\"os_name,expected_manager\", [
        (\"linux\", \"apt\"),
        (\"windows\", \"chocolatey\"),
        (\"macos\", \"homebrew\"),
    ])
    def test_should_select_correct_package_manager(self, os_name, expected_manager):
        from cassandragargoyle.project.install.detector import SystemDetector
        
        detector = SystemDetector()
        manager = detector.get_package_manager_for_os(os_name)
        
        assert manager == expected_manager
    
    @pytest.fixture(params=[\"ubuntu\", \"debian\", \"fedora\", \"centos\"])
    def linux_distribution(self, request):
        \"\"\"Parametrized fixture for Linux distributions.\"\"\"
        return request.param
    
    def test_should_handle_all_linux_distributions(self, linux_distribution):
        from cassandragargoyle.project.install.detector import SystemDetector
        
        with patch(\"platform.system\", return_value=\"Linux\"):
            with patch(\"distro.id\", return_value=linux_distribution):
                detector = SystemDetector()
                result = detector.detect_linux_distribution()
                
                assert result == linux_distribution
```

### Testing with Mocks and Patches
```python
import pytest
from unittest.mock import Mock, patch, mock_open, call
from pathlib import Path

class TestConfigurationLoader:
    \"\"\"Tests for configuration loading functionality.\"\"\"
    
    def test_should_load_configuration_from_yaml_file(self):
        # Arrange
        yaml_content = \"\"\"
        timeout: 300
        retry_count: 3
        package_manager: apt
        \"\"\"
        
        with patch(\"builtins.open\", mock_open(read_data=yaml_content)):
            with patch(\"yaml.safe_load\") as mock_yaml_load:
                mock_yaml_load.return_value = {
                    \"timeout\": 300,
                    \"retry_count\": 3,
                    \"package_manager\": \"apt\"
                }
                
                from cassandragargoyle.project.config.loader import ConfigLoader
                loader = ConfigLoader()
                
                # Act
                config = loader.load_from_file(\"config.yaml\")
                
                # Assert
                assert config.timeout == 300
                assert config.retry_count == 3
                assert config.package_manager == \"apt\"
    
    def test_should_handle_file_not_found_error(self):
        from cassandragargoyle.project.config.loader import ConfigLoader
        from cassandragargoyle.project.config.exceptions import ConfigurationError
        
        loader = ConfigLoader()
        
        with patch(\"builtins.open\", side_effect=FileNotFoundError()):
            with pytest.raises(ConfigurationError, match=\"Configuration file not found\"):
                loader.load_from_file(\"nonexistent.yaml\")
    
    @patch(\"cassandragargoyle.project.config.loader.requests.get\")
    def test_should_load_configuration_from_url(self, mock_get):
        # Arrange
        mock_response = Mock()
        mock_response.text = \"timeout: 600\\nretry_count: 5\"
        mock_response.status_code = 200
        mock_get.return_value = mock_response
        
        from cassandragargoyle.project.config.loader import ConfigLoader
        loader = ConfigLoader()
        
        # Act
        config = loader.load_from_url(\"https://example.com/config.yaml\")
        
        # Assert
        assert config.timeout == 600
        mock_get.assert_called_once_with(\"https://example.com/config.yaml\", timeout=30)
    
    def test_should_use_environment_variables_as_fallback(self):
        with patch.dict(os.environ, {
            \"PACKAGE_TIMEOUT\": \"450\",
            \"PACKAGE_RETRY_COUNT\": \"5\"
        }):
            from cassandragargoyle.project.config.loader import ConfigLoader
            loader = ConfigLoader()
            
            config = loader.load_from_environment()
            
            assert config.timeout == 450
            assert config.retry_count == 5
```

## Integration Testing

### Integration Test Structure
```python
import pytest
import tempfile
import subprocess
from pathlib import Path

class TestInstallationWorkflow:
    \"\"\"Integration tests for complete installation workflow.\"\"\"
    
    @pytest.fixture(scope=\"class\")
    def temp_config_file(self):
        \"\"\"Create a temporary configuration file for testing.\"\"\"
        config_content = \"\"\"
        timeout: 300
        retry_count: 3
        package_managers:
          linux: apt
          windows: chocolatey
          macos: homebrew
        \"\"\"
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write(config_content)
            temp_path = Path(f.name)
        
        yield temp_path
        
        # Cleanup
        temp_path.unlink(missing_ok=True)
    
    @pytest.mark.integration
    def test_should_complete_full_installation_workflow(self, temp_config_file):
        \"\"\"Test complete workflow from config loading to package installation.\"\"\"
        from cassandragargoyle.project.config.loader import ConfigLoader
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        # Load configuration
        loader = ConfigLoader()
        config = loader.load_from_file(str(temp_config_file))
        
        # Create installer with real dependencies
        installer = PackageInstaller(config=config)
        
        # Test with a safe, commonly available package
        result = installer.install_package(\"curl\")
        
        # Verify installation
        assert result is True
        assert installer.is_package_installed(\"curl\") is True
    
    @pytest.mark.integration
    @pytest.mark.slow
    def test_should_handle_network_package_installation(self):
        \"\"\"Test installation of packages that require network access.\"\"\"
        if not pytest.config.getoption(\"--run-integration\"):
            pytest.skip(\"Integration tests disabled\")
        
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        installer = PackageInstaller()
        
        # Test downloading and installing a package
        result = installer.install_from_url(
            \"https://example.com/packages/test-package.tar.gz\"
        )
        
        assert result is True
    
    @pytest.mark.integration
    def test_should_rollback_on_installation_failure(self, temp_config_file):
        \"\"\"Test rollback functionality when installation fails.\"\"\"
        from cassandragargoyle.project.config.loader import ConfigLoader
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        loader = ConfigLoader()
        config = loader.load_from_file(str(temp_config_file))
        installer = PackageInstaller(config=config)
        
        # Attempt to install non-existent package
        with pytest.raises(PackageInstallationError):
            installer.install_package(\"definitely-nonexistent-package-12345\")
        
        # Verify system state wasn't changed
        assert not installer.is_package_installed(\"definitely-nonexistent-package-12345\")
```

### Docker-based Integration Tests
```python
import pytest
import testcontainers
from testcontainers.general import DockerContainer

class TestDockerBasedInstallation:
    \"\"\"Integration tests using Docker containers.\"\"\"
    
    @pytest.fixture(scope=\"class\")
    def ubuntu_container(self):
        \"\"\"Start Ubuntu container for testing.\"\"\"
        with DockerContainer(\"ubuntu:22.04\").with_command(\"tail -f /dev/null\") as container:
            yield container
    
    @pytest.mark.integration
    @pytest.mark.docker
    def test_should_install_package_in_ubuntu_container(self, ubuntu_container):
        \"\"\"Test package installation in a clean Ubuntu environment.\"\"\"
        # Update package list
        result = ubuntu_container.exec(\"apt-get update -qq\")
        assert result.exit_code == 0
        
        # Install package
        result = ubuntu_container.exec(\"apt-get install -y python3\")
        assert result.exit_code == 0
        
        # Verify installation
        result = ubuntu_container.exec(\"python3 --version\")
        assert result.exit_code == 0
        assert \"Python 3\" in result.stdout
    
    @pytest.mark.integration
    @pytest.mark.docker
    def test_should_handle_package_conflicts_gracefully(self, ubuntu_container):
        \"\"\"Test handling of package conflicts.\"\"\"
        # Install conflicting packages and verify behavior
        result1 = ubuntu_container.exec(\"apt-get install -y python3.9\")
        result2 = ubuntu_container.exec(\"apt-get install -y python3.10\")
        
        # Both should succeed or handle conflict gracefully
        assert result1.exit_code in [0, 1]  # 0 = success, 1 = handled conflict
        assert result2.exit_code in [0, 1]
```

### API Integration Tests
```python
import pytest
import responses
import httpx
from unittest.mock import patch

class TestAPIIntegration:
    \"\"\"Integration tests for API-based functionality.\"\"\"
    
    @responses.activate
    def test_should_fetch_package_information_from_api(self):
        \"\"\"Test fetching package information from external API.\"\"\"
        # Mock API response
        responses.add(
            responses.GET,
            \"https://api.example.com/packages/python3\",
            json={
                \"name\": \"python3\",
                \"version\": \"3.11.0\",
                \"description\": \"Python programming language\",
                \"dependencies\": [\"python3-dev\", \"python3-pip\"]
            },
            status=200
        )
        
        from cassandragargoyle.project.api.client import PackageAPIClient
        
        client = PackageAPIClient(base_url=\"https://api.example.com\")
        package_info = client.get_package_info(\"python3\")
        
        assert package_info[\"name\"] == \"python3\"
        assert package_info[\"version\"] == \"3.11.0\"
        assert len(package_info[\"dependencies\"]) == 2
    
    @responses.activate
    def test_should_handle_api_timeout_gracefully(self):
        \"\"\"Test handling of API timeouts.\"\"\"
        # Mock timeout
        responses.add(
            responses.GET,
            \"https://api.example.com/packages/slow-package\",
            body=httpx.TimeoutException(\"Request timed out\")
        )
        
        from cassandragargoyle.project.api.client import PackageAPIClient
        from cassandragargoyle.project.api.exceptions import APITimeoutError
        
        client = PackageAPIClient(base_url=\"https://api.example.com\", timeout=5)
        
        with pytest.raises(APITimeoutError):
            client.get_package_info(\"slow-package\")
```

## Async Testing

### Testing Asyncio Code
```python
import pytest
import asyncio
from unittest.mock import AsyncMock

class TestAsyncInstaller:
    \"\"\"Tests for asynchronous installation functionality.\"\"\"
    
    @pytest.mark.asyncio
    async def test_should_install_package_asynchronously(self):
        \"\"\"Test asynchronous package installation.\"\"\"
        from cassandragargoyle.project.install.async_installer import AsyncPackageInstaller
        
        mock_manager = AsyncMock()
        mock_manager.install_async.return_value = True
        
        installer = AsyncPackageInstaller(manager=mock_manager)
        
        # Act
        result = await installer.install_package_async(\"python3\")
        
        # Assert
        assert result is True
        mock_manager.install_async.assert_called_once_with(\"python3\")
    
    @pytest.mark.asyncio
    async def test_should_handle_concurrent_installations(self):
        \"\"\"Test concurrent package installations.\"\"\"
        from cassandragargoyle.project.install.async_installer import AsyncPackageInstaller
        
        installer = AsyncPackageInstaller()
        packages = [\"python3\", \"git\", \"vim\", \"curl\", \"wget\"]
        
        # Install packages concurrently
        tasks = [installer.install_package_async(pkg) for pkg in packages]
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Check that most installations succeeded (some might fail in test environment)
        successful = sum(1 for r in results if r is True)
        assert successful >= len(packages) // 2  # At least half should succeed
    
    @pytest.mark.asyncio
    async def test_should_timeout_slow_async_operations(self):
        \"\"\"Test timeout handling for slow async operations.\"\"\"
        from cassandragargoyle.project.install.async_installer import AsyncPackageInstaller
        
        async def slow_install(package_name):
            await asyncio.sleep(10)  # Simulate very slow installation
            return True
        
        mock_manager = AsyncMock()
        mock_manager.install_async.side_effect = slow_install
        
        installer = AsyncPackageInstaller(manager=mock_manager)
        
        # Should timeout after 1 second
        with pytest.raises(asyncio.TimeoutError):
            await asyncio.wait_for(
                installer.install_package_async(\"slow-package\"),
                timeout=1.0
            )
```

## Property-based Testing with Hypothesis

### Using Hypothesis for Robust Testing
```python
import pytest
from hypothesis import given, strategies as st, assume, example
import string

class TestPackageNameValidation:
    \"\"\"Property-based tests for package name validation.\"\"\"
    
    @given(st.text(alphabet=string.ascii_lowercase + string.digits + \"-\", min_size=1, max_size=50))
    def test_valid_package_names_should_be_accepted(self, package_name):
        \"\"\"Test that valid package names are accepted.\"\"\"
        from cassandragargoyle.project.install.validator import validate_package_name
        
        # Ensure we don't start or end with hyphens
        assume(not package_name.startswith(\"-\"))
        assume(not package_name.endswith(\"-\"))
        assume(\"--\" not in package_name)  # No double hyphens
        
        # Should not raise any exception
        validate_package_name(package_name)
    
    @given(st.text(alphabet=string.ascii_letters + string.digits + \"!@#$%^&*()_+{}|:<>?[];',./\"))
    def test_invalid_package_names_should_be_rejected(self, package_name):
        \"\"\"Test that invalid package names are rejected.\"\"\"
        from cassandragargoyle.project.install.validator import validate_package_name
        
        # Skip valid names that might be generated
        assume(any(char in package_name for char in \"!@#$%^&*()_+{}|:<>?[];',./\"))
        
        with pytest.raises(ValueError):
            validate_package_name(package_name)
    
    @given(st.lists(st.text(min_size=1, max_size=20), min_size=0, max_size=100))
    def test_batch_installation_should_handle_any_list_size(self, package_list):
        \"\"\"Test batch installation with various list sizes.\"\"\"
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        installer = PackageInstaller()
        
        # Should handle empty lists, single items, and large batches
        result = installer.install_packages(package_list)
        
        # Result should be a list of the same length
        assert len(result) == len(package_list)
        # All results should be boolean
        assert all(isinstance(r, bool) for r in result)
    
    @example(package_name=\"\")  # Explicit edge case
    @example(package_name=\"a\" * 256)  # Too long
    @given(st.text())
    def test_edge_cases_in_package_names(self, package_name):
        \"\"\"Test edge cases in package name validation.\"\"\"
        from cassandragargoyle.project.install.validator import validate_package_name
        
        if len(package_name) == 0 or len(package_name) > 255:
            with pytest.raises(ValueError):
                validate_package_name(package_name)
        elif any(char in package_name for char in [\";\", \"|\", \"&\", \"$\", \"`\"]):
            with pytest.raises(ValueError):
                validate_package_name(package_name)
```

## Testing Utilities and Fixtures

### Custom Fixtures
```python
# tests/utils/fixtures.py
import pytest
import tempfile
import shutil
from pathlib import Path
from unittest.mock import Mock

@pytest.fixture
def temp_directory():
    \"\"\"Create a temporary directory for testing.\"\"\"
    temp_dir = Path(tempfile.mkdtemp())
    yield temp_dir
    shutil.rmtree(temp_dir, ignore_errors=True)

@pytest.fixture
def sample_config_file(temp_directory):
    \"\"\"Create a sample configuration file.\"\"\"
    config_content = \"\"\"
    timeout: 300
    retry_count: 3
    package_manager: auto
    log_level: INFO
    \"\"\"
    
    config_file = temp_directory / \"config.yaml\"
    config_file.write_text(config_content)
    return config_file

@pytest.fixture
def mock_package_manager():
    \"\"\"Create a mock package manager with common methods.\"\"\"
    manager = Mock()
    manager.install.return_value = True
    manager.remove.return_value = True
    manager.is_installed.return_value = False
    manager.list_installed.return_value = []
    return manager

@pytest.fixture(params=[\"apt\", \"yum\", \"chocolatey\", \"homebrew\"])
def package_manager_type(request):
    \"\"\"Parametrized fixture for different package manager types.\"\"\"
    return request.param

@pytest.fixture(scope=\"session\")
def docker_available():
    \"\"\"Check if Docker is available for testing.\"\"\"
    import subprocess
    try:
        subprocess.run([\"docker\", \"--version\"], check=True, capture_output=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False
```

### Test Data Factories
```python
# tests/utils/factories.py
import factory
from datetime import datetime
from cassandragargoyle.project.config.models import Configuration
from cassandragargoyle.project.install.models import PackageInfo

class ConfigurationFactory(factory.Factory):
    class Meta:
        model = Configuration
    
    timeout = 300
    retry_count = 3
    package_manager = \"auto\"
    log_level = \"INFO\"
    created_at = factory.LazyFunction(datetime.now)

class PackageInfoFactory(factory.Factory):
    class Meta:
        model = PackageInfo
    
    name = factory.Sequence(lambda n: f\"package-{n}\")
    version = \"1.0.0\"
    description = factory.Faker(\"sentence\")
    dependencies = factory.List([\"dep1\", \"dep2\"])
    installed = False

# Usage in tests
def test_configuration_with_factory():
    config = ConfigurationFactory(timeout=600, retry_count=5)
    assert config.timeout == 600
    assert config.retry_count == 5

def test_batch_package_creation():
    packages = PackageInfoFactory.build_batch(10)
    assert len(packages) == 10
    assert all(pkg.version == \"1.0.0\" for pkg in packages)
```

### Test Helpers
```python
# tests/utils/test_helpers.py
import os
import subprocess
import contextlib
from pathlib import Path
from typing import Dict, Any, Optional

def run_command(command: list, cwd: Optional[Path] = None, env: Optional[Dict[str, str]] = None) -> subprocess.CompletedProcess:
    \"\"\"Run a command and return the result.\"\"\"
    full_env = os.environ.copy()
    if env:
        full_env.update(env)
    
    return subprocess.run(
        command,
        cwd=cwd,
        env=full_env,
        capture_output=True,
        text=True
    )

def create_test_file(path: Path, content: str) -> Path:
    \"\"\"Create a test file with given content.\"\"\"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)
    return path

@contextlib.contextmanager
def environment_variable(name: str, value: str):
    \"\"\"Temporarily set an environment variable.\"\"\"
    old_value = os.environ.get(name)
    os.environ[name] = value
    try:
        yield
    finally:
        if old_value is None:
            os.environ.pop(name, None)
        else:
            os.environ[name] = old_value

def assert_command_successful(result: subprocess.CompletedProcess, message: str = \"\"):
    \"\"\"Assert that a command completed successfully.\"\"\"
    if result.returncode != 0:
        error_msg = f\"Command failed: {result.args}\\n\"
        error_msg += f\"Return code: {result.returncode}\\n\"
        error_msg += f\"Stdout: {result.stdout}\\n\"
        error_msg += f\"Stderr: {result.stderr}\"
        if message:
            error_msg = f\"{message}\\n{error_msg}\"
        raise AssertionError(error_msg)

def wait_for_condition(condition_func, timeout: int = 10, interval: float = 0.1):
    \"\"\"Wait for a condition to become true.\"\"\"
    import time
    
    start_time = time.time()
    while time.time() - start_time < timeout:
        if condition_func():
            return True
        time.sleep(interval)
    
    raise TimeoutError(f\"Condition not met within {timeout} seconds\")
```

## Performance and Load Testing

### Performance Tests
```python
import pytest
import time
from unittest.mock import patch

class TestPerformance:
    \"\"\"Performance-related tests.\"\"\"
    
    def test_installation_should_complete_within_time_limit(self):
        \"\"\"Test that installation completes within acceptable time.\"\"\"
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        installer = PackageInstaller()
        
        start_time = time.time()
        result = installer.install_package(\"quick-package\")
        end_time = time.time()
        
        execution_time = end_time - start_time
        
        assert result is True
        assert execution_time < 5.0, f\"Installation took {execution_time:.2f} seconds (limit: 5.0s)\"
    
    @pytest.mark.benchmark
    def test_package_validation_performance(self, benchmark):
        \"\"\"Benchmark package name validation performance.\"\"\"
        from cassandragargoyle.project.install.validator import validate_package_name
        
        def validation_operation():
            return validate_package_name(\"test-package-name\")
        
        result = benchmark(validation_operation)
        assert result is None  # validate_package_name returns None on success
    
    def test_concurrent_installation_performance(self):
        \"\"\"Test performance with concurrent installations.\"\"\"
        import concurrent.futures
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        installer = PackageInstaller()
        packages = [f\"package-{i}\" for i in range(10)]
        
        start_time = time.time()
        
        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(installer.install_package, pkg) for pkg in packages]
            results = [future.result() for future in concurrent.futures.as_completed(futures)]
        
        end_time = time.time()
        execution_time = end_time - start_time
        
        assert len(results) == len(packages)
        assert execution_time < 15.0, f\"Concurrent installation took {execution_time:.2f} seconds\"
```

### Load Testing with pytest-benchmark
```python
import pytest

class TestLoadPerformance:
    \"\"\"Load and stress testing.\"\"\"
    
    @pytest.mark.parametrize(\"package_count\", [10, 50, 100, 500])
    def test_batch_installation_scalability(self, package_count, benchmark):
        \"\"\"Test scalability of batch package installation.\"\"\"
        from cassandragargoyle.project.install.installer import PackageInstaller
        
        installer = PackageInstaller()
        packages = [f\"package-{i}\" for i in range(package_count)]
        
        def batch_install():
            return installer.install_packages(packages)
        
        result = benchmark(batch_install)
        assert len(result) == package_count
    
    def test_memory_usage_with_large_datasets(self):
        \"\"\"Test memory usage with large package lists.\"\"\"
        import psutil
        import os
        
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB
        
        # Create large dataset
        from cassandragargoyle.project.install.installer import PackageInstaller
        installer = PackageInstaller()
        
        large_package_list = [f\"package-{i}\" for i in range(10000)]
        installer.validate_package_list(large_package_list)
        
        final_memory = process.memory_info().rss / 1024 / 1024  # MB
        memory_increase = final_memory - initial_memory
        
        # Memory increase should be reasonable (less than 100MB for 10k packages)
        assert memory_increase < 100, f\"Memory increased by {memory_increase:.2f} MB\"
```

## Testing Configuration

### pytest Configuration
```python
# conftest.py
import pytest
import os
import sys
from pathlib import Path

# Add src to Python path
sys.path.insert(0, str(Path(__file__).parent / \"src\"))

def pytest_configure(config):
    \"\"\"Configure pytest markers.\"\"\"
    config.addinivalue_line(\"markers\", \"unit: Unit tests\")
    config.addinivalue_line(\"markers\", \"integration: Integration tests\")
    config.addinivalue_line(\"markers\", \"e2e: End-to-end tests\")
    config.addinivalue_line(\"markers\", \"slow: Slow tests (> 1 second)\")
    config.addinivalue_line(\"markers\", \"network: Tests requiring network access\")
    config.addinivalue_line(\"markers\", \"docker: Tests requiring Docker\")
    config.addinivalue_line(\"markers\", \"benchmark: Performance benchmark tests\")

def pytest_addoption(parser):
    \"\"\"Add custom command line options.\"\"\"
    parser.addoption(
        \"--run-integration\",
        action=\"store_true\",
        default=False,
        help=\"Run integration tests\"
    )
    parser.addoption(
        \"--run-e2e\",
        action=\"store_true\",
        default=False,
        help=\"Run end-to-end tests\"
    )
    parser.addoption(
        \"--docker\",
        action=\"store_true\",
        default=False,
        help=\"Run tests that require Docker\"
    )

def pytest_collection_modifyitems(config, items):
    \"\"\"Modify test collection based on options.\"\"\"
    if not config.getoption(\"--run-integration\"):
        skip_integration = pytest.mark.skip(reason=\"need --run-integration option to run\")
        for item in items:
            if \"integration\" in item.keywords:
                item.add_marker(skip_integration)
    
    if not config.getoption(\"--run-e2e\"):
        skip_e2e = pytest.mark.skip(reason=\"need --run-e2e option to run\")
        for item in items:
            if \"e2e\" in item.keywords:
                item.add_marker(skip_e2e)
    
    if not config.getoption(\"--docker\"):
        skip_docker = pytest.mark.skip(reason=\"need --docker option to run\")
        for item in items:
            if \"docker\" in item.keywords:
                item.add_marker(skip_docker)

@pytest.fixture(autouse=True)
def reset_environment():
    \"\"\"Reset environment for each test.\"\"\"
    # Store original environment
    original_env = os.environ.copy()
    
    yield
    
    # Restore original environment
    os.environ.clear()
    os.environ.update(original_env)
```

## Coverage and Quality

### Running Tests with Coverage
```bash
# Run all tests with coverage
pytest --cov=src --cov-report=html --cov-report=term-missing

# Run specific test categories
pytest -m unit --cov=src
pytest -m integration --run-integration --cov=src
pytest -m \"not slow\" --cov=src

# Run tests in parallel
pytest -n auto --cov=src

# Generate coverage report
pytest --cov=src --cov-report=html
# Report will be in htmlcov/index.html
```

### Coverage Requirements
- **Minimum coverage**: 85% for unit tests
- **Critical modules**: 95% coverage required
- **Integration coverage**: Additional coverage beyond unit tests

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Python Test Suite

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        python-version: [\"3.8\", \"3.9\", \"3.10\", \"3.11\"]
        os: [ubuntu-latest, windows-latest, macos-latest]
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -e \".[test]\"
    
    - name: Run unit tests
      run: |
        pytest -m unit --cov=src --cov-report=xml
    
    - name: Run integration tests
      run: |
        pytest -m integration --run-integration
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml
        
  quality:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-python@v4
      with:
        python-version: \"3.11\"
    
    - name: Install dependencies
      run: |
        pip install -e \".[test]\"
        pip install black isort flake8 mypy
    
    - name: Code formatting
      run: |
        black --check src tests
        isort --check-only src tests
    
    - name: Linting
      run: flake8 src tests
    
    - name: Type checking
      run: mypy src
```

## Testing Scripts

### Main Test Script
```bash
#!/bin/bash
# scripts/test.sh

set -e

# Colors
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
NC='\\033[0m'

echo -e \"${GREEN}Running Python Test Suite${NC}\"

# Install dependencies
echo -e \"${YELLOW}Installing test dependencies...${NC}\"
pip install -e \".[test]\"

# Code quality checks
echo -e \"${YELLOW}Running code quality checks...${NC}\"
black --check src tests
isort --check-only src tests
flake8 src tests
mypy src

# Unit tests
echo -e \"${YELLOW}Running unit tests...${NC}\"
pytest -m unit --cov=src --cov-report=term-missing

# Integration tests
echo -e \"${YELLOW}Running integration tests...${NC}\"
pytest -m integration --run-integration

# Coverage report
echo -e \"${YELLOW}Generating coverage report...${NC}\"
pytest --cov=src --cov-report=html
echo \"Coverage report: htmlcov/index.html\"

echo -e \"${GREEN}All tests passed!${NC}\"
```

## Best Practices Summary

### Test Organization
1. Follow pytest conventions for test discovery
2. Use descriptive test names that explain the scenario
3. Organize tests by functionality, not by test type
4. Use appropriate markers to categorize tests

### Test Quality
1. Each test should verify one specific behavior
2. Use AAA pattern (Arrange, Act, Assert)
3. Mock external dependencies and I/O operations
4. Make tests independent and deterministic

### Performance and Maintainability
1. Keep unit tests fast (< 100ms each)
2. Use fixtures to reduce code duplication
3. Implement property-based testing for robust validation
4. Monitor test execution times and coverage

### Modern Python Practices
1. Use type hints in test code
2. Leverage async/await for testing async code
3. Use context managers for resource cleanup
4. Follow PEP 8 guidelines in test code

---

**Note**: These guidelines should be adapted based on specific project requirements and Python version used. Regular review ensures tests remain effective and maintainable.

*Created: 2025-08-23*
*Last updated: 2025-08-23*