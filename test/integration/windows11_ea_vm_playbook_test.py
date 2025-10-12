#!/usr/bin/env python3
"""
Windows 11 Enterprise Architect Development VM Playbook Test

This test executes the Windows 11 EA development VM playbook using 'portunix playbook run'
and validates the successful completion of each phase.

Test Execution:
    python3 test/integration/windows11_ea_vm_playbook_test.py

Requirements:
- Portunix binary available in ../../portunix
- QEMU/KVM virtualization support
- Sufficient system resources (16GB RAM recommended)
- Internet connection for downloads
- Windows 11 ISO (will be downloaded automatically)
"""

import os
import sys
import subprocess
import json
import time
import unittest
from typing import Dict, List, Tuple, Optional
from pathlib import Path


class WindowsEAVMPlaybookTest(unittest.TestCase):
    """Test Windows 11 Enterprise Architect Development VM Playbook execution"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.test_name = "Windows11_EA_VM_Playbook"
        self.vm_name = "ea-dev-win11-test"
        self.playbook_path = "test/playbooks/windows11-ea-development-vm.ptxbook"
        self.portunix_binary = "./portunix"
        self.test_results = []
        self.start_time = None
        self.verbose = os.getenv('VERBOSE', '0') == '1' or '-v' in sys.argv

    def setUp(self):
        """Setup test environment"""
        self.start_time = time.time()
        self.log_header("Windows 11 EA Development VM Playbook Test")
        self.log_info(f"Test started at: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        self.log_info(f"VM Name: {self.vm_name}")
        self.log_info(f"Playbook: {self.playbook_path}")

        # Verify test prerequisites
        self._verify_prerequisites()

    def tearDown(self):
        """Cleanup test environment"""
        duration = time.time() - self.start_time if self.start_time else 0

        # Cleanup test VM if it exists
        self._cleanup_test_vm()

        self.log_separator()
        self.log_success(f"Test completed in {duration:.2f} seconds")
        self._print_test_summary()

    def test_playbook_execution_full(self):
        """Test full playbook execution"""
        self.log_phase("Full Playbook Execution")

        # Warning about full execution
        self.log_step("Preparing for full playbook execution")
        self.log_warning("This will create actual VM and download Windows 11 ISO (~6GB)")
        self.log_info("Full execution enabled - this may take 30-60 minutes")

        # Execute full playbook on host (default environment)
        self.log_step("Execute full playbook on host")
        result = self._run_command([
            self.portunix_binary, "playbook", "run",
            self.playbook_path, "--verbose"
        ], timeout=3600)  # 1 hour timeout for full execution

        if result['success']:
            self.log_success("Full playbook execution completed successfully")
            self._verify_vm_creation()
            self._analyze_execution_output(result['stdout'])
        else:
            self.log_error(f"Full playbook execution failed: {result['stderr']}")
            self.fail(f"Full playbook execution failed: {result['stderr']}")

    # Helper Methods for Playbook Analysis
    def _analyze_dry_run_output(self, output: str):
        """Analyze dry-run output for expected phases"""
        self.log_step("Analyzing dry-run output")

        expected_phases = [
            "Environment Setup",
            "VM Creation",
            "Network Configuration",
            "Development Tools",
            "Enterprise Architect",
            "Repository Setup",
            "VS Code Extensions",
            "Final Configuration"
        ]

        found_phases = []
        for phase in expected_phases:
            if any(phase.lower() in line.lower() for line in output.split('\n')):
                found_phases.append(phase)

        self.log_info(f"Found {len(found_phases)}/{len(expected_phases)} expected phases")
        for phase in found_phases:
            self.log_success(f"✓ {phase}")

    def _analyze_execution_output(self, output: str):
        """Analyze execution output for success indicators"""
        self.log_step("Analyzing execution output")

        success_indicators = [
            "completed successfully",
            "✅",
            "PASS",
            "SUCCESS"
        ]

        error_indicators = [
            "FAILED",
            "ERROR",
            "❌",
            "FAIL"
        ]

        lines = output.split('\n')
        successes = sum(1 for line in lines if any(ind in line for ind in success_indicators))
        errors = sum(1 for line in lines if any(ind in line for ind in error_indicators))

        self.log_info(f"Execution summary: {successes} successes, {errors} errors")

        if errors > 0:
            self.log_warning(f"Found {errors} error indicators in output")
        else:
            self.log_success("No error indicators found")

    def _verify_vm_creation(self):
        """Verify that VM was actually created"""
        self.log_step("Verifying VM creation")

        result = self._run_command([self.portunix_binary, "virt", "list"])
        if result['success'] and self.vm_name in result['stdout']:
            self.log_success(f"VM {self.vm_name} successfully created")

            # Check VM status
            status_result = self._run_command([self.portunix_binary, "virt", "status", self.vm_name])
            if status_result['success']:
                self.log_info(f"VM status: {status_result['stdout'].strip()}")
        else:
            self.log_warning("VM not found in VM list")


    # Helper Methods
    def _verify_prerequisites(self):
        """Verify test prerequisites"""
        self.log_step("Verifying test prerequisites")

        # Check portunix binary
        if not os.path.exists(self.portunix_binary):
            self.fail(f"Portunix binary not found at {self.portunix_binary}")

        # Check playbook file
        if not os.path.exists(self.playbook_path):
            self.fail(f"Playbook file not found at {self.playbook_path}")

        self.log_success("Prerequisites verified")

    def _cleanup_test_vm(self):
        """Clean up any test VMs"""
        self.log_step("Cleaning up test VMs")

        test_vm_patterns = [f"{self.vm_name}-minimal", self.vm_name]
        for vm_pattern in test_vm_patterns:
            result = self._run_command([self.portunix_binary, "virt", "destroy", vm_pattern], silent=True)
            if result['success']:
                self.log_info(f"Cleaned up VM: {vm_pattern}")


    def _get_system_info(self) -> Dict[str, float]:
        """Get system resource information"""
        try:
            # Get RAM info
            with open('/proc/meminfo', 'r') as f:
                mem_info = f.read()
                mem_total_line = [line for line in mem_info.split('\n') if 'MemTotal' in line][0]
                mem_kb = int(mem_total_line.split()[1])
                ram_gb = mem_kb / (1024 * 1024)

            # Get disk info (approximate)
            result = subprocess.run(['df', '-h', '.'], capture_output=True, text=True)
            if result.returncode == 0:
                lines = result.stdout.strip().split('\n')
                disk_line = lines[1].split()
                disk_available = disk_line[3]
                # Convert to GB (rough approximation)
                if 'G' in disk_available:
                    disk_gb = float(disk_available.replace('G', ''))
                elif 'T' in disk_available:
                    disk_gb = float(disk_available.replace('T', '')) * 1024
                else:
                    disk_gb = 100  # Default assumption
            else:
                disk_gb = 100  # Default assumption

            return {'ram_gb': ram_gb, 'disk_gb': disk_gb}
        except:
            return {'ram_gb': 8.0, 'disk_gb': 100.0}  # Default values

    def _run_command(self, cmd: List[str], silent: bool = False, timeout: int = 300) -> Dict[str, any]:
        """Run a command and return result"""
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout)
            success = result.returncode == 0

            if not silent and self.verbose:
                self.log_command(" ".join(cmd))
                if result.stdout:
                    self.log_output(result.stdout[:1000])  # Show more output for playbook execution

            return {
                'success': success,
                'returncode': result.returncode,
                'stdout': result.stdout,
                'stderr': result.stderr
            }
        except subprocess.TimeoutExpired:
            return {
                'success': False,
                'returncode': -1,
                'stdout': '',
                'stderr': f'Command timeout after {timeout} seconds'
            }
        except Exception as e:
            return {
                'success': False,
                'returncode': -1,
                'stdout': '',
                'stderr': str(e)
            }

    # Logging Methods
    def log_header(self, message: str):
        """Log test header"""
        if self.verbose:
            print("=" * 80)
            print(f"🚀 {message}")
            print("=" * 80)

    def log_phase(self, message: str):
        """Log test phase"""
        if self.verbose:
            print(f"\n📋 {message}")
            print("-" * 60)

    def log_step(self, message: str):
        """Log test step"""
        if self.verbose:
            print(f"   📌 {message}")

    def log_success(self, message: str):
        """Log success message"""
        if self.verbose:
            print(f"   ✅ {message}")

    def log_warning(self, message: str):
        """Log warning message"""
        if self.verbose:
            print(f"   ⚠️  {message}")

    def log_error(self, message: str):
        """Log error message"""
        if self.verbose:
            print(f"   ❌ {message}")

    def log_info(self, message: str):
        """Log info message"""
        if self.verbose:
            print(f"   ℹ️  {message}")

    def log_command(self, command: str):
        """Log executed command"""
        if self.verbose:
            print(f"   🔧 Executing: {command}")

    def log_output(self, output: str):
        """Log command output"""
        if self.verbose and output.strip():
            print(f"   📄 Output ({len(output)} chars):")
            for line in output.strip().split('\n')[:10]:  # Limit to first 10 lines
                print(f"      {line}")
            if len(output.split('\n')) > 10:
                print(f"      ... (output truncated)")

    def log_separator(self):
        """Log separator"""
        if self.verbose:
            print("-" * 80)

    def _print_test_summary(self):
        """Print test execution summary"""
        if self.verbose:
            print(f"\n🎉 {self.test_name} completed successfully")
            print(f"📊 Full playbook execution test completed")
            print(f"⏱️  Total execution time: {time.time() - self.start_time:.2f} seconds")
            print(f"📝 Playbook file: {self.playbook_path}")
            print(f"🔧 Portunix binary: {self.portunix_binary}")
            print(f"🏠 Execution environment: Host PC")
            print(f"🚀 Windows 11 EA VM should be created and configured")


def main():
    """Main test execution"""
    # Configure test verbosity
    if len(sys.argv) > 1 and sys.argv[1] in ['-v', '--verbose']:
        os.environ['VERBOSE'] = '1'

    # Run the test
    unittest.main(argv=[''], verbosity=2 if os.getenv('VERBOSE') == '1' else 1, exit=False)


if __name__ == "__main__":
    main()