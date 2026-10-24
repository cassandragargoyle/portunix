#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""
gosec scan of every Go module in the repository (issue #228)

`gosec ./...` stops at nested module boundaries, so each directory with its
own go.mod is scanned separately. Modules the root go.mod points at through
local replace directives are scanned inside a temporary Go workspace, because
several of them have stub go.mod files that do not resolve on their own.

Single source of truth for the gosec invocation - used by the Security Scan
job in GitHub Actions, `make security` and scripts/github-01-preflight-check.sh.
Excluded rules are documented in docs/contributing/SECURITY-SCAN.md.

Usage:
    python3 scripts/gosec-scan.py                      # scan the repository
    python3 scripts/gosec-scan.py --report-dir DIR     # keep one JSON report per module
    python3 scripts/gosec-scan.py --summary            # no per-finding lines
    python3 scripts/gosec-scan.py [options] REPO_DIR   # scan another checkout
"""

import argparse
import collections
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Dict, List, Optional

# Rules excluded by design, see docs/contributing/SECURITY-SCAN.md
GOSEC_EXCLUDE = "G204,G304,G702,G703"

# Directories that hold go.mod files of vendored or template code
# (same as GO_MODULES in the Makefile)
SKIPPED_TOP_DIRS = {"docs", "docs-site", ".git"}


def find_modules(repo: Path) -> List[Path]:
    """Return every module directory in the repository, relative to repo"""
    modules = []
    for dirpath, dirnames, filenames in os.walk(repo):
        rel = Path(dirpath).relative_to(repo)
        if rel == Path("."):
            dirnames[:] = [d for d in dirnames if d not in SKIPPED_TOP_DIRS]
        if "go.mod" in filenames:
            modules.append(rel)
    return sorted(modules, key=lambda p: p.as_posix())


def workspace_modules(repo: Path) -> List[Path]:
    """Return the root module and the local replace targets of its go.mod"""
    members = [Path(".")]
    pattern = re.compile(r"^(?:replace\s+)?\S+(?:\s+\S+)?\s+=>\s+(\.{1,2}/\S+)")
    for line in (repo / "go.mod").read_text(encoding="utf-8").splitlines():
        match = pattern.match(line.strip())
        if not match:
            continue
        target = Path(os.path.normpath(match.group(1)))
        if (repo / target / "go.mod").is_file() and target not in members:
            members.append(target)
    return members


def write_workspace(repo: Path, members: List[Path], work_dir: Path) -> Path:
    """Create a temporary go.work that joins the workspace members"""
    go_line = "go 1.25"
    for line in (repo / "go.mod").read_text(encoding="utf-8").splitlines():
        if re.match(r"^go \d", line):
            go_line = line.strip()
            break

    lines = [go_line, "", "use ("]
    for member in members:
        lines.append("\t" + (repo / member).as_posix())
    lines.append(")")

    go_work = work_dir / "go.work"
    go_work.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return go_work


def module_name(module: Path) -> str:
    return "(root)" if module == Path(".") else module.as_posix()


def module_slug(module: Path) -> str:
    return "root" if module == Path(".") else module.as_posix().replace("/", "_")


def run_in_module(args: List[str], module_dir: Path, gowork: str) -> subprocess.CompletedProcess:
    # Go trusts $PWD over getcwd, keep both in sync so paths in go.work match
    # even when the checkout is reached through a symlink
    env = dict(os.environ, GOWORK=gowork, PWD=str(module_dir))
    return subprocess.run(args, cwd=module_dir, env=env, capture_output=True, text=True)


def scan_module(repo: Path, module: Path, modules: List[Path], gowork: str, report: Path) -> Optional[dict]:
    """Run gosec in one module and return its parsed JSON report"""
    args = ["gosec", "-exclude-generated", f"-exclude={GOSEC_EXCLUDE}"]

    # Nested modules are scanned on their own
    for other in modules:
        if other != module and module in other.parents:
            args.append(f"-exclude-dir={other.relative_to(module).as_posix()}")

    # Template and vendored code under the skipped top-level directories is not product code
    if module == Path("."):
        for top in sorted(SKIPPED_TOP_DIRS):
            if (repo / top).is_dir() and top != ".git":
                args.append(f"-exclude-dir={top}")

    args += ["-fmt", "json", "-out", str(report), "./..."]

    result = run_in_module(args, repo / module, gowork)

    data = None
    if report.is_file() and report.stat().st_size > 0:
        try:
            data = json.loads(report.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            data = None

    if data is None and result.returncode != 0:
        # gosec failed before it could write a report
        tail = (result.stderr or result.stdout).strip().splitlines()[-5:]
        return {"fatal": tail or [f"gosec exited with code {result.returncode}"]}
    return data


def packages_load(repo: Path, module: Path, gowork: str) -> Optional[str]:
    """Return None when the module's packages load, the error text otherwise"""
    result = run_in_module(["go", "list", "./..."], repo / module, gowork)
    if result.returncode == 0:
        return None
    return result.stderr.strip() or f"go list exited with code {result.returncode}"


def relative_file(repo: Path, file_name: str) -> str:
    try:
        return Path(file_name).resolve().relative_to(repo).as_posix()
    except ValueError:
        return file_name


def print_findings(repo: Path, issues: List[dict], errors: Dict[str, list], summary: bool) -> None:
    by_rule = collections.Counter((i["rule_id"], i["severity"]) for i in issues)
    for (rule, severity), count in sorted(by_rule.items(), key=lambda item: (-item[1], item[0])):
        print(f"        {rule:6} {severity:6} {count}")

    for package, package_errors in errors.items():
        for error in package_errors[:3]:
            print(f"        build error in {relative_file(repo, package)}: {error.get('error', '')}")

    for issue in issues:
        # HIGH findings are always listed, they must be fixed in code
        if summary and issue["severity"] != "HIGH":
            continue
        location = f"{relative_file(repo, issue['file'])}:{issue['line']}"
        print(f"        {issue['severity']:6} {issue['rule_id']} {location} - {issue['details']}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Run gosec on every Go module in the repository")
    parser.add_argument("repo_dir", nargs="?", help="repository to scan (default: this checkout)")
    parser.add_argument("--report-dir", help="keep one gosec JSON report per module in this directory")
    parser.add_argument("--summary", action="store_true", help="print counts per rule, list only HIGH findings")
    opts = parser.parse_args()

    repo = Path(opts.repo_dir) if opts.repo_dir else Path(__file__).resolve().parent.parent
    repo = repo.resolve()
    if not (repo / "go.mod").is_file():
        print(f"error: no go.mod in {repo}", file=sys.stderr)
        return 2

    if shutil.which("gosec") is None:
        print("error: gosec is required (go install github.com/securego/gosec/v2/cmd/gosec@latest)", file=sys.stderr)
        return 2

    modules = find_modules(repo)
    members = workspace_modules(repo)

    with tempfile.TemporaryDirectory(prefix="gosec-scan-") as tmp:
        work_dir = Path(tmp)
        go_work = write_workspace(repo, members, work_dir)

        report_dir = Path(opts.report_dir).absolute() if opts.report_dir else work_dir / "reports"
        report_dir.mkdir(parents=True, exist_ok=True)

        failed = 0
        for module in modules:
            name = module_name(module)
            gowork = str(go_work) if module in members else "off"
            report = report_dir / f"gosec-{module_slug(module)}.json"
            if report.exists():
                report.unlink()

            data = scan_module(repo, module, modules, gowork, report)

            if data is not None and "fatal" in data:
                print(f"FAIL  {name}  - gosec could not run")
                for line in data["fatal"]:
                    print(f"        {line}")
                failed += 1
                continue

            issues = (data or {}).get("Issues") or []
            errors = (data or {}).get("Golang errors") or {}
            files = (data or {}).get("Stats", {}).get("files")

            if issues or errors:
                parts = []
                if issues:
                    parts.append(f"{len(issues)} finding(s)")
                if errors:
                    parts.append(f"{len(errors)} package(s) with build errors")
                print(f"FAIL  {name}  - {', '.join(parts)}")
                print_findings(repo, issues, errors, opts.summary)
                failed += 1
                continue

            # gosec also reports success when it could not load a single
            # package, so make sure the module really loads
            load_error = packages_load(repo, module, gowork)
            if load_error is not None:
                print(f"FAIL  {name}  - packages cannot be loaded")
                for line in load_error.splitlines()[-5:]:
                    print(f"        {line}")
                failed += 1
                continue

            print(f"PASS  {name}  ({files} files)" if files else f"PASS  {name}  (generated code only)")

    print()
    if failed:
        print(f"gosec: {failed} of {len(modules)} module(s) failed")
        return 1
    print(f"gosec: all {len(modules)} module(s) passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
