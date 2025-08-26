"""
pytest configuration and shared fixtures for Portunix tests.
"""

import pytest
import subprocess
from pathlib import Path


@pytest.fixture(scope="session")
def project_root():
    """Project root directory fixture."""
    return Path(__file__).parent.parent


@pytest.fixture(scope="session") 
def portunix_binary(project_root):
    """Portunix binary path fixture."""
    binary_path = project_root / "portunix"
    if not binary_path.exists():
        pytest.skip(f"Portunix binary not found at {binary_path}")
    return binary_path


@pytest.fixture(scope="session")
def podman_available():
    """Check if Podman is available."""
    try:
        subprocess.run(["podman", "info"], capture_output=True, check=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        pytest.skip("Podman is not available")


def pytest_configure(config):
    """Configure pytest with custom markers."""
    config.addinivalue_line("markers", "quick: Quick tests for development")
    config.addinivalue_line("markers", "slow: Slow tests that take time")
    config.addinivalue_line("markers", "integration: Integration tests")
    config.addinivalue_line("markers", "e2e: End-to-end tests")
    config.addinivalue_line("markers", "podman: Tests requiring Podman")


def pytest_collection_modifyitems(config, items):
    """Add markers to tests based on their location."""
    for item in items:
        # Add integration marker to tests in integration directory
        if "integration" in str(item.fspath):
            item.add_marker(pytest.mark.integration)
            item.add_marker(pytest.mark.podman)
        
        # Add e2e marker to tests in e2e directory  
        if "e2e" in str(item.fspath):
            item.add_marker(pytest.mark.e2e)
            item.add_marker(pytest.mark.podman)
            item.add_marker(pytest.mark.slow)


def pytest_report_header(config):
    """Add custom header to pytest output."""
    return [
        "Portunix Integration Test Suite",
        "Container Engine: Podman (rootless mode)",
        "Test Framework: pytest + Python"
    ]