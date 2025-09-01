#!/usr/bin/env python3
"""
Python Integration Test Runner for Portunix
Replaces the bash version with proper Python integration testing
"""

import sys
import subprocess
import argparse
from pathlib import Path


def main():
    """Main test runner entry point."""
    parser = argparse.ArgumentParser(
        description="Portunix Integration Test Runner",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --quick                    # Run quick test (Ubuntu 22.04 only)
  %(prog)s -q                         # Same as --quick
  %(prog)s --full-suite --parallel    # Run all tests in parallel
  %(prog)s -f -p                      # Same as above with short options
  %(prog)s --distribution ubuntu-22   # Run specific distribution
  %(prog)s -d ubuntu-22               # Same with short option
  %(prog)s --list-distributions       # List available distributions
  %(prog)s -l                         # Same with short option
        """
    )
    
    parser.add_argument("-q", "--quick", action="store_true",
                       help="Run quick test (Ubuntu 22.04 only)")
    parser.add_argument("-f", "--full-suite", action="store_true", 
                       help="Run complete test suite (all distributions)")
    parser.add_argument("-d", "--distribution", 
                       help="Run specific distribution test")
    parser.add_argument("-l", "--list-distributions", action="store_true",
                       help="List available distributions")
    parser.add_argument("-p", "--parallel", action="store_true",
                       help="Run tests in parallel")
    parser.add_argument("-c", "--cleanup", action="store_true",
                       help="Clean up test containers")
    parser.add_argument("-v", "--verbose", action="store_true",
                       help="Verbose output")
    parser.add_argument("--html-report", 
                       help="Generate HTML report to specified file")
    
    args = parser.parse_args()
    
    # Determine project root
    script_dir = Path(__file__).parent
    project_root = script_dir.parent.parent
    test_file = project_root / "test" / "integration" / "issue_012_powershell_installation_test.py"
    
    if not test_file.exists():
        print(f"❌ Test file not found: {test_file}")
        return 1
    
    # Handle special actions
    if args.list_distributions or args.cleanup:
        return subprocess.call([sys.executable, str(test_file), 
                              "--list-distributions" if args.list_distributions else "--cleanup"])
    
    # Validate arguments
    action_count = sum([args.quick, args.full_suite, bool(args.distribution)])
    if action_count == 0:
        print("❌ ERROR: No test suite type specified!")
        print("\nAvailable test modes:")
        print("  --quick              Run quick test (Ubuntu 22.04 only)")
        print("  --full-suite         Run complete test suite (all distributions)")
        print("  --distribution NAME  Run specific distribution test")
        print("\nFor complete help: python test/scripts/test-integration.py --help")
        return 1
    
    if action_count > 1:
        print("❌ ERROR: Multiple test modes specified. Choose only one.")
        return 1
    
    # Build pytest command
    cmd = [sys.executable, "-m", "pytest", str(test_file)]
    
    # Add verbosity
    if args.verbose:
        cmd.append("-v")
    elif args.full_suite and not args.parallel:
        # For full suite without parallel, show some progress
        cmd.append("-v")
    else:
        cmd.append("-q")
    
    # Add test selection
    if args.quick:
        cmd.extend(["-m", "quick"])
    elif args.distribution:
        cmd.extend(["-k", args.distribution])
    # For full-suite, no additional filters needed
    
    # Add parallel execution
    if args.parallel:
        cmd.extend(["-n", "auto"])  # Requires pytest-xdist
    
    # Add HTML report
    if args.html_report:
        cmd.extend(["--html", args.html_report, "--self-contained-html"])
    
    # Show test header
    print("=" * 64)
    print("  🧪 Issue #012 PowerShell Installation Test Suite")
    print(f"  📄 Test file: {test_file.name}")
    print("  🐍 Python Integration Testing Framework")
    print("=" * 64)
    print(f"Container Engine: Podman (rootless)")
    
    if args.quick:
        print("Suite Type: Quick Test (Ubuntu 22.04)")
        print("Expected duration: ~1 minute")
    elif args.full_suite:
        print("Suite Type: Full Suite (6 distributions)")
        if args.parallel:
            print("Expected duration: ~2-3 minutes (parallel execution)")
        else:
            print("Expected duration: ~6-8 minutes (sequential execution)")
            print("💡 Tip: Use -p/--parallel for faster execution")
    elif args.distribution:
        print(f"Suite Type: Single Distribution ({args.distribution})")
        print("Expected duration: ~1 minute")
    
    print(f"Parallel: {'Yes' if args.parallel else 'No'}")
    print("=" * 64)
    
    if args.full_suite and not args.parallel:
        print("⏳ Starting tests... This may take several minutes.")
        print("   Each distribution test takes about 1 minute to complete.")
        print("   Press Ctrl+C to cancel at any time.")
        print()
    
    print()
    
    # Run tests
    try:
        result = subprocess.run(cmd, cwd=project_root)
        return result.returncode
    except KeyboardInterrupt:
        print("\n⚠️  Test execution interrupted by user")
        return 130
    except Exception as e:
        print(f"❌ Error running tests: {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())