#!/usr/bin/env python3
"""
Integration tests for Issue #012: PowerShell Installation on Linux Distributions

This test suite validates PowerShell installation across multiple Linux distributions
using Podman containers and the Portunix package manager.
"""

import pytest
import subprocess
import time
import json
import os
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from datetime import datetime
import tempfile
import shutil


class TestPowerShellInstallation:
    """Integration tests for PowerShell installation across Linux distributions."""
    
    # Test configuration
    DISTRIBUTIONS = [
        ("ubuntu-22", "ubuntu", "ubuntu:22.04"),
        ("ubuntu-24", "ubuntu", "ubuntu:24.04"),
        # Note: Debian 11 excluded - GLIBC too old for current portunix binary
        # ("debian-11", "debian", "debian:bullseye"),  # GLIBC 2.31 - too old
        ("debian-12", "debian", "debian:bookworm"),
        ("fedora-39", "fedora", "fedora:39"),
        ("fedora-40", "fedora", "fedora:40"),
        ("rocky-9", "rocky", "rockylinux:9"),
        ("mint-21", "ubuntu", "docker.io/vcatechnology/linux-mint:21.2"),  # Mint uses Ubuntu variant
    ]
    
    CONTAINER_PREFIX = "portunix-test-ps"
    PORTUNIX_BIN = None
    
    @classmethod
    def setup_class(cls):
        """Setup test environment and validate prerequisites."""
        cls._setup_environment()
        cls._check_prerequisites()
        cls._setup_results_directory()
    
    @classmethod
    def teardown_class(cls):
        """Cleanup test environment."""
        cls._cleanup_all_containers()
    
    @classmethod
    def _setup_environment(cls):
        """Setup test environment paths."""
        # Determine project root and portunix binary location
        current_dir = Path(__file__).parent
        project_root = current_dir.parent.parent
        cls.PORTUNIX_BIN = project_root / "portunix"
        
        # Setup results directory
        cls.results_dir = project_root / "test" / "results"
        cls.timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        cls.log_dir = cls.results_dir / f"logs-{cls.timestamp}"
        cls.log_dir.mkdir(parents=True, exist_ok=True)
    
    @classmethod
    def _check_prerequisites(cls):
        """Check if all prerequisites are available."""
        # Check if portunix binary exists
        if not cls.PORTUNIX_BIN.exists():
            pytest.fail(f"Portunix binary not found at: {cls.PORTUNIX_BIN}")
        
        # Check if Podman is installed and working
        try:
            result = subprocess.run(
                ["podman", "info"], 
                capture_output=True, 
                text=True, 
                check=True
            )
            print(f"✅ Podman is available and working")
        except (subprocess.CalledProcessError, FileNotFoundError):
            pytest.fail("Podman is not installed or not working properly")
    
    @classmethod
    def _setup_results_directory(cls):
        """Setup results directory structure."""
        cls.results_dir.mkdir(parents=True, exist_ok=True)
        print(f"📁 Results directory: {cls.results_dir}")
        print(f"📁 Log directory: {cls.log_dir}")
    
    @classmethod
    def _cleanup_all_containers(cls):
        """Remove all test containers."""
        try:
            result = subprocess.run(
                ["podman", "ps", "-a", "--format", "{{.Names}}"],
                capture_output=True,
                text=True
            )
            
            containers = [
                name for name in result.stdout.strip().split('\n')
                if name.startswith(cls.CONTAINER_PREFIX)
            ]
            
            if containers:
                print(f"🧹 Cleaning up {len(containers)} test containers...")
                for container in containers:
                    subprocess.run(
                        ["podman", "rm", "-f", container],
                        capture_output=True
                    )
                print("✅ All test containers cleaned up")
        except Exception as e:
            print(f"⚠️  Warning: Could not clean up containers: {e}")
    
    def setup_method(self):
        """Setup for each test method."""
        self.container_id = None
        self.container_name = None
        self.test_start_time = time.time()
    
    def teardown_method(self):
        """Cleanup after each test method."""
        if self.container_name and not self._should_keep_container():
            self._remove_container(self.container_name)
    
    def _should_keep_container(self) -> bool:
        """Determine if container should be kept for debugging."""
        # Keep failed containers if pytest fails
        return hasattr(self, '_test_failed') and self._test_failed
    
    @pytest.mark.parametrize("name,variant,image", DISTRIBUTIONS)
    def test_powershell_installation_on_distribution(self, name: str, variant: str, image: str):
        """Test PowerShell installation on specific Linux distribution."""
        print(f"\n🧪 Testing PowerShell installation on {name} ({image})")
        print(f"  📅 Started at: {datetime.now().strftime('%H:%M:%S')}")
        
        log_file = self.log_dir / f"{name}.log"
        
        try:
            with open(log_file, 'w') as log:
                self._log_test_header(log, name, variant, image)
                
                # Step 1: Create container
                self._log_step(log, 1, 5, "Creating container")
                container_name = self._create_container(log, name, image)
                self.container_name = container_name
                
                # Step 2: Copy portunix binary
                self._log_step(log, 2, 5, "Copying portunix binary")
                self._copy_portunix_binary(log, container_name)
                
                # Step 3: Install PowerShell
                self._log_step(log, 3, 5, "Installing PowerShell")
                install_success = self._install_powershell(log, container_name, variant)
                
                if not install_success:
                    # Try fallback
                    self._log_fallback(log, container_name)
                
                # Step 4: Verify installation
                self._log_step(log, 4, 5, "Verifying PowerShell installation")
                self._verify_powershell(log, container_name)
                
                # Step 5: Final system check
                self._log_step(log, 5, 5, "Final system state check")
                self._final_system_check(log, container_name)
                
                self._log_test_success(log, name)
                
                # Show completion info
                duration = time.time() - self.test_start_time
                print(f"  🎉 Test completed successfully in {duration:.1f}s")
                
        except Exception as e:
            self._test_failed = True
            duration = time.time() - self.test_start_time
            # Truncate error message for console output
            error_msg = str(e)
            if len(error_msg) > 150:
                error_msg = error_msg[:150] + "... (see log file for full details)"
            print(f"  💥 Test failed after {duration:.1f}s: {error_msg}")
            with open(log_file, 'a') as log:
                self._log_test_failure(log, name, str(e))
            pytest.fail(f"PowerShell installation failed on {name}: {error_msg}")
    
    def _log_test_header(self, log, name: str, variant: str, image: str):
        """Log test header with emojis and formatting."""
        log.write("=" * 64 + "\n")
        log.write(f"🧪 DISTRIBUTION TEST: {name}\n")
        log.write("=" * 64 + "\n")
        log.write(f"📦 Container: {self.CONTAINER_PREFIX}-{name}\n")
        log.write(f"🐧 Image: {image}\n")
        log.write(f"🔧 Variant: {variant}\n")
        log.write(f"⏰ Started: {datetime.now()}\n")
        log.write(f"🌍 Test Environment: {os.getenv('USER', 'unknown')}@{os.uname().nodename}\n")
        log.write("=" * 64 + "\n\n")
        log.flush()
    
    def _log_step(self, log, step: int, total: int, description: str):
        """Log test step."""
        log.write(f"🔨 STEP {step}/{total}: {description}...\n")
        log.flush()
        elapsed = time.time() - self.test_start_time
        print(f"  🔨 Step {step}/{total}: {description}... ({elapsed:.1f}s elapsed)")
    
    def _log_fallback(self, log, container_name: str):
        """Log fallback attempt."""
        log.write("\n🔄 FALLBACK: Attempting snap installation...\n")
        log.write("Command: /usr/local/bin/portunix install powershell --variant snap\n")
        log.write("--- Fallback Installation Output ---\n")
        log.flush()
        
        if not self._install_powershell_fallback(log, container_name):
            raise Exception("Both primary and fallback installations failed")
        
        log.write("--- End Fallback Installation Output ---\n")
        log.write("✅ Fallback installation succeeded\n")
    
    def _log_test_success(self, log, name: str):
        """Log successful test completion."""
        duration = time.time() - self.test_start_time
        log.write("\n" + "=" * 64 + "\n")
        log.write("✅ TEST COMPLETED SUCCESSFULLY\n")
        log.write(f"🎉 Distribution: {name}\n")
        log.write(f"⏰ Completed: {datetime.now()}\n")
        log.write(f"⏱️  Duration: {duration:.1f} seconds\n")
        log.write("=" * 64 + "\n")
    
    def _log_test_failure(self, log, name: str, error: str):
        """Log failed test."""
        duration = time.time() - self.test_start_time
        log.write("\n" + "=" * 64 + "\n")
        log.write("❌ TEST FAILED\n")
        log.write(f"💥 Distribution: {name}\n")
        log.write(f"🔥 Error: {error}\n")
        log.write(f"⏰ Failed at: {datetime.now()}\n")
        log.write(f"⏱️  Duration: {duration:.1f} seconds\n")
        log.write("=" * 64 + "\n")
    
    def _create_container(self, log, name: str, image: str) -> str:
        """Create and setup test container."""
        container_name = f"{self.CONTAINER_PREFIX}-{name}"
        
        print(f"    🔨 Creating container {container_name}...")
        log.write(f"Command: podman run -d --name {container_name} {image}\n")
        
        # Remove existing container if it exists
        print(f"    🧹 Removing existing container (if any)...")
        subprocess.run(
            ["podman", "rm", "-f", container_name],
            capture_output=True
        )
        
        # Create new container with proper network configuration
        print(f"    🐧 Starting {image} container...")
        result = subprocess.run(
            [
                "podman", "run", "-d",
                "--name", container_name,
                "--hostname", name,
                "--network", "host",  # Use host networking to avoid DNS issues
                "-e", "DEBIAN_FRONTEND=noninteractive",
                image,
                "/bin/sh", "-c", "tail -f /dev/null"
            ],
            capture_output=True,
            text=True,
            check=True
        )
        
        container_id = result.stdout.strip()
        print(f"    ✅ Container created: {container_id[:12]}")
        log.write(f"✅ Container created with ID: {container_id[:12]}\n\n")
        
        # NOTE: We do NOT install any dependencies here!
        # Portunix should handle all necessary dependencies itself
        print(f"    📋 Container is clean - no pre-installed dependencies")
        log.write("📋 Container is clean - portunix will handle all dependencies\n\n")
        
        # Log container info
        self._log_container_info(log, container_name)
        
        return container_name
    
    def _log_container_info(self, log, container_name: str):
        """Log container information."""
        log.write("📊 Container Information:\n")
        
        # Container ID and status
        try:
            id_result = subprocess.run(
                ["podman", "inspect", container_name, "--format", "{{.Id}}"],
                capture_output=True, text=True
            )
            if id_result.returncode == 0:
                log.write(f"ID: {id_result.stdout.strip()}\n")
        except:
            log.write("Could not get container ID\n")
        
        # OS release info
        try:
            os_result = subprocess.run(
                ["podman", "exec", container_name, "cat", "/etc/os-release"],
                capture_output=True, text=True
            )
            if os_result.returncode == 0:
                for line in os_result.stdout.split('\n')[:5]:
                    if line.strip():
                        log.write(f"{line}\n")
        except:
            log.write("Could not get OS info\n")
        
        log.write("\n")
    
    def _copy_portunix_binary(self, log, container_name: str):
        """Copy portunix binary to container."""
        print(f"    📋 Copying portunix binary to container...")
        log.write(f"Source: {self.PORTUNIX_BIN}\n")
        log.write(f"Target container: {container_name}:/usr/local/bin/portunix\n")
        
        result = subprocess.run([
            "podman", "cp",
            str(self.PORTUNIX_BIN),
            f"{container_name}:/usr/local/bin/portunix"
        ], capture_output=True, text=True)
        
        if result.returncode != 0:
            error_msg = result.stderr if result.stderr else "Unknown error"
            if len(error_msg) > 100:
                error_msg = error_msg[:100] + "..."
            raise Exception(f"Failed to copy portunix binary: {error_msg}")
        
        # Make executable
        subprocess.run([
            "podman", "exec", container_name,
            "chmod", "+x", "/usr/local/bin/portunix"
        ], check=True)
        
        # Verify binary
        verify_result = subprocess.run([
            "podman", "exec", container_name,
            "ls", "-la", "/usr/local/bin/portunix"
        ], capture_output=True, text=True)
        
        print(f"    ✅ Binary copied successfully")
        log.write(f"✅ Binary copied successfully: {verify_result.stdout.strip()}\n\n")
    
    def _install_powershell(self, log, container_name: str, variant: str) -> bool:
        """Install PowerShell using portunix."""
        cmd = f"/usr/local/bin/portunix install powershell --variant {variant}"
        print(f"    🔧 Installing PowerShell via {variant} variant: {cmd}")
        log.write(f"Command: {cmd}\n")
        log.write(f"Installation started at: {datetime.now()}\n")
        log.write("--- Installation Output ---\n")
        
        result = subprocess.run([
            "podman", "exec", container_name,
            "/usr/local/bin/portunix", "install", "powershell", "--variant", variant
        ], capture_output=True, text=True)
        
        # Log installation output
        if result.stdout:
            log.write(result.stdout)
        if result.stderr:
            log.write(result.stderr)
        
        log.write("--- End Installation Output ---\n")
        
        if result.returncode == 0:
            print(f"    ✅ PowerShell installation succeeded")
            log.write(f"✅ PowerShell installation succeeded via {variant} variant\n\n")
            return True
        else:
            print(f"    ❌ PowerShell installation failed via {variant}")
            log.write(f"❌ PowerShell installation failed via {variant} variant\n")
            return False
    
    def _install_powershell_fallback(self, log, container_name: str) -> bool:
        """Install PowerShell using snap fallback."""
        print(f"    🔄 Trying snap fallback installation...")
        result = subprocess.run([
            "podman", "exec", container_name,
            "/usr/local/bin/portunix", "install", "powershell", "--variant", "snap"
        ], capture_output=True, text=True)
        
        # Log fallback output
        if result.stdout:
            log.write(result.stdout)
        if result.stderr:
            log.write(result.stderr)
        
        if result.returncode == 0:
            print(f"    ✅ Snap fallback succeeded")
        else:
            print(f"    ❌ Snap fallback also failed")
        return result.returncode == 0
    
    def _verify_powershell(self, log, container_name: str):
        """Verify PowerShell installation."""
        print(f"    🔍 Verifying PowerShell installation...")
        log.write(f"Verification started at: {datetime.now()}\n")
        log.write("--- Verification Output ---\n")
        
        # Run comprehensive verification
        result = subprocess.run([
            "podman", "exec", container_name, "sh", "-c", """
            echo 'Checking PowerShell binary location:'
            command -v pwsh
            echo ''
            echo 'PowerShell version:'
            pwsh --version 2>&1
            echo ''
            echo 'PowerShell test command:'
            pwsh -c 'Write-Host "PowerShell is working: $($PSVersionTable.PSVersion)"' 2>&1
            """
        ], capture_output=True, text=True)
        
        log.write(result.stdout)
        if result.stderr:
            log.write(result.stderr)
        
        log.write("--- End Verification Output ---\n")
        
        # Check if verification was successful
        if result.returncode != 0 or "PowerShell" not in result.stdout:
            print(f"    ❌ PowerShell verification failed")
            # Create brief error message for exception
            error_details = result.stderr if result.stderr else "No PowerShell found in output"
            if len(error_details) > 100:
                error_details = error_details[:100] + "..."
            raise Exception(f"PowerShell verification failed: {error_details}")
        
        print(f"    ✅ PowerShell verification successful")
        log.write("✅ PowerShell verification successful\n\n")
    
    def _final_system_check(self, log, container_name: str):
        """Perform final system state check."""
        print(f"    📊 Final system state check...")
        log.write("Checking installed packages and system state...\n")
        
        result = subprocess.run([
            "podman", "exec", container_name, "sh", "-c", """
            echo 'Installed PowerShell packages:'
            if command -v dpkg &> /dev/null; then
                dpkg -l | grep -i powershell || echo 'No PowerShell packages found via dpkg'
            elif command -v rpm &> /dev/null; then
                rpm -qa | grep -i powershell || echo 'No PowerShell packages found via rpm'
            fi
            echo ''
            echo 'PowerShell executable:'
            which pwsh || echo 'pwsh not found in PATH'
            echo ''
            echo 'Available PowerShell commands:'
            pwsh -c 'Get-Command | Select-Object -First 5 Name' 2>/dev/null || echo 'Could not list PowerShell commands'
            """
        ], capture_output=True, text=True)
        
        log.write(result.stdout)
        if result.stderr:
            log.write(result.stderr)
        
        log.write("\n")
    
    def _remove_container(self, container_name: str):
        """Remove test container."""
        subprocess.run(
            ["podman", "rm", "-f", container_name],
            capture_output=True
        )
    
    @pytest.mark.quick
    def test_quick_ubuntu_22_installation(self):
        """Quick test for Ubuntu 22.04 only."""
        self.test_powershell_installation_on_distribution("ubuntu-22", "ubuntu", "ubuntu:22.04")
    
    def test_generate_html_report(self):
        """Generate HTML test report (runs after all distribution tests)."""
        report_file = self.results_dir / f"issue-012-test-report-{self.timestamp}.html"
        
        # Generate HTML report
        self._generate_html_report(report_file)
        
        print(f"📊 HTML report generated: file://{report_file}")
    
    def _generate_html_report(self, report_file: Path):
        """Generate HTML test report."""
        
        # Collect test results
        test_results = []
        total_tests = 0
        passed_tests = 0
        
        for name, variant, image in self.DISTRIBUTIONS:
            log_file = self.log_dir / f"{name}.log"
            status = "Not Run"
            
            if log_file.exists():
                total_tests += 1
                log_content = log_file.read_text()
                if "TEST COMPLETED SUCCESSFULLY" in log_content:
                    status = "Passed"
                    passed_tests += 1
                else:
                    status = "Failed"
            
            test_results.append({
                'name': name,
                'variant': variant,
                'image': image,
                'status': status,
                'log_file': f"logs-{self.timestamp}/{name}.log"
            })
        
        pass_rate = (passed_tests * 100 // total_tests) if total_tests > 0 else 0
        
        # Generate HTML content
        html_content = f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Issue #012 Test Report - {self.timestamp}</title>
    <style>
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }}
        .header {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
        }}
        h1 {{ margin: 0; font-size: 2em; }}
        .subtitle {{ opacity: 0.9; margin-top: 10px; }}
        .summary {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }}
        .card {{
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }}
        .card h3 {{ margin-top: 0; color: #667eea; }}
        .stat {{ font-size: 2em; font-weight: bold; }}
        .passed {{ color: #28a745; }}
        .failed {{ color: #dc3545; }}
        .total {{ color: #007bff; }}
        .details {{
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }}
        table {{ width: 100%; border-collapse: collapse; margin-top: 20px; }}
        th, td {{ padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }}
        th {{ background: #f8f9fa; font-weight: 600; }}
        .status-pass {{
            background: #d4edda; color: #155724;
            padding: 4px 8px; border-radius: 4px;
        }}
        .status-fail {{
            background: #f8d7da; color: #721c24;
            padding: 4px 8px; border-radius: 4px;
        }}
        .footer {{
            margin-top: 30px; text-align: center;
            color: #666; font-size: 0.9em;
        }}
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 Issue #012: PowerShell Linux Installation</h1>
        <div class="subtitle">Python Integration Test Report - {self.timestamp}</div>
        <div class="subtitle">Container Engine: Podman (rootless mode)</div>
    </div>
    
    <div class="summary">
        <div class="card">
            <h3>Total Tests</h3>
            <div class="stat total">{total_tests}</div>
        </div>
        <div class="card">
            <h3>Passed</h3>
            <div class="stat passed">{passed_tests}</div>
        </div>
        <div class="card">
            <h3>Failed</h3>
            <div class="stat failed">{total_tests - passed_tests}</div>
        </div>
        <div class="card">
            <h3>Pass Rate</h3>
            <div class="stat {'passed' if pass_rate >= 80 else 'failed'}">{pass_rate}%</div>
        </div>
    </div>
    
    <div class="details">
        <h2>Test Results by Distribution</h2>
        <table>
            <thead>
                <tr>
                    <th>Distribution</th>
                    <th>Variant</th>
                    <th>Image</th>
                    <th>Status</th>
                    <th>Log File</th>
                </tr>
            </thead>
            <tbody>"""
        
        for result in test_results:
            status_class = "status-pass" if result['status'] == "Passed" else "status-fail"
            html_content += f"""
                <tr>
                    <td>{result['name']}</td>
                    <td>{result['variant']}</td>
                    <td>{result['image']}</td>
                    <td><span class="{status_class}">{result['status']}</span></td>
                    <td><a href="{result['log_file']}">{result['name']}.log</a></td>
                </tr>"""
        
        html_content += f"""
            </tbody>
        </table>
    </div>
    
    <div class="footer">
        <p>Generated by Portunix Python Integration Test Suite</p>
        <p>Report created at {datetime.now()}</p>
    </div>
</body>
</html>"""
        
        report_file.write_text(html_content)


# Command line interface for running individual tests
if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Issue #012 PowerShell Installation Tests")
    parser.add_argument("--quick", action="store_true", help="Run quick test (Ubuntu 22.04 only)")
    parser.add_argument("--distribution", help="Run specific distribution test")
    parser.add_argument("--list-distributions", action="store_true", help="List available distributions")
    parser.add_argument("--cleanup", action="store_true", help="Clean up test containers")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")
    
    args = parser.parse_args()
    
    if args.list_distributions:
        print("Available distributions for testing:")
        print("Name          Variant       Image")
        print("-" * 50)
        for name, variant, image in TestPowerShellInstallation.DISTRIBUTIONS:
            print(f"{name:<12} {variant:<12} {image}")
        exit(0)
    
    if args.cleanup:
        TestPowerShellInstallation._cleanup_all_containers()
        exit(0)
    
    # Run pytest programmatically
    pytest_args = [__file__]
    
    # Add verbosity
    if args.verbose:
        pytest_args.extend(["-v", "-s"])  # -s for print outputs
    else:
        pytest_args.append("-q")
    
    if args.quick:
        pytest_args.extend(["-m", "quick"])
    elif args.distribution:
        # Find the specific distribution test
        found = False
        for name, variant, image in TestPowerShellInstallation.DISTRIBUTIONS:
            if name == args.distribution:
                pytest_args.extend(["-k", f"test_powershell_installation_on_distribution[{name}-{variant}-{image}]"])
                found = True
                break
        if not found:
            print(f"Distribution '{args.distribution}' not found. Use --list-distributions to see available options.")
            exit(1)
    
    exit(pytest.main(pytest_args))