#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""
govulncheck scan of every Go module in the repository (issue #232)

Replaces nancy, which needs an OSS Index account. govulncheck queries the Go
vulnerability database (vuln.go.dev) without credentials and reports only
vulnerabilities whose vulnerable functions the code calls.

Modules are discovered and joined into a temporary Go workspace the same way
as in scripts/gosec-scan.py. The script stays self-contained because the
preflight check copies scripts into containers one by one.

Used by the Security Scan job in GitHub Actions and `make security`.
See docs/contributing/SECURITY-SCAN.md.

Usage:
    python3 scripts/govulncheck-scan.py                      # scan the repository
    python3 scripts/govulncheck-scan.py --report-dir DIR     # keep one JSON report per module
    python3 scripts/govulncheck-scan.py [options] REPO_DIR   # scan another checkout
"""

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Dict, List, Tuple

# Directories that hold go.mod files of vendored or template code
# (same as GO_MODULES in the Makefile and SKIPPED_TOP_DIRS in gosec-scan.py)
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


def parse_stream(text: str) -> List[dict]:
    """Split the govulncheck JSON output, a stream of concatenated objects"""
    decoder = json.JSONDecoder()
    messages = []
    pos = 0
    while True:
        while pos < len(text) and text[pos].isspace():
            pos += 1
        if pos >= len(text):
            break
        message, pos = decoder.raw_decode(text, pos)
        messages.append(message)
    return messages


def called_vulns(messages: List[dict]) -> Dict[Tuple[str, str], dict]:
    """Return the vulnerabilities the code calls, keyed by OSV id and module"""
    summaries = {m["osv"]["id"]: m["osv"].get("summary", "") for m in messages if "osv" in m}

    vulns: Dict[Tuple[str, str], dict] = {}
    for message in messages:
        finding = message.get("finding")
        if not finding:
            continue
        trace = finding.get("trace") or []
        # Only symbol-level findings carry a function in the first frame,
        # package and module level findings are not reachable from the code
        if not trace or not trace[0].get("function"):
            continue
        osv_id = finding["osv"]
        frame = trace[0]
        # One OSV entry may cover several modules (net/http and x/net/http2)
        module = frame.get("module", "")
        vulns.setdefault((osv_id, module), {
            "id": osv_id,
            "module": module,
            "found": frame.get("version", ""),
            "fixed": finding.get("fixed_version", "") or "none",
            "summary": summaries.get(osv_id, ""),
        })
    return vulns


def scan_module(repo: Path, module: Path, gowork: str) -> Tuple[Dict[Tuple[str, str], dict], List[str], str]:
    """Run govulncheck in one module, return called vulns, fatal error lines and raw output"""
    result = run_in_module(["govulncheck", "-format", "json", "./..."], repo / module, gowork)

    # In JSON mode govulncheck exits 0 also when it finds vulnerabilities,
    # a non-zero exit means the scan itself failed
    if result.returncode != 0:
        tail = (result.stderr or result.stdout).strip().splitlines()[-5:]
        return {}, tail or [f"govulncheck exited with code {result.returncode}"], result.stdout

    try:
        messages = parse_stream(result.stdout)
    except json.JSONDecodeError as exc:
        return {}, [f"cannot parse govulncheck output: {exc}"], result.stdout
    return called_vulns(messages), [], result.stdout


def main() -> int:
    parser = argparse.ArgumentParser(description="Run govulncheck on every Go module in the repository")
    parser.add_argument("repo_dir", nargs="?", help="repository to scan (default: this checkout)")
    parser.add_argument("--report-dir", help="keep one govulncheck JSON report per module in this directory")
    opts = parser.parse_args()

    repo = Path(opts.repo_dir) if opts.repo_dir else Path(__file__).resolve().parent.parent
    repo = repo.resolve()
    if not (repo / "go.mod").is_file():
        print(f"error: no go.mod in {repo}", file=sys.stderr)
        return 2

    if shutil.which("govulncheck") is None:
        print("error: govulncheck is required (go install golang.org/x/vuln/cmd/govulncheck@latest)", file=sys.stderr)
        return 2

    modules = find_modules(repo)
    members = workspace_modules(repo)

    report_dir = Path(opts.report_dir).absolute() if opts.report_dir else None
    if report_dir:
        report_dir.mkdir(parents=True, exist_ok=True)

    with tempfile.TemporaryDirectory(prefix="govulncheck-scan-") as tmp:
        go_work = write_workspace(repo, members, Path(tmp))

        failed = 0
        stdlib_fixes = set()
        for module in modules:
            name = module_name(module)
            gowork = str(go_work) if module in members else "off"

            vulns, fatal, raw = scan_module(repo, module, gowork)

            if report_dir:
                (report_dir / f"govulncheck-{module_slug(module)}.json").write_text(raw, encoding="utf-8")

            if fatal:
                print(f"FAIL  {name}  - govulncheck could not run")
                for line in fatal:
                    print(f"        {line}")
                failed += 1
                continue

            if vulns:
                print(f"FAIL  {name}  - {len(vulns)} called vulnerability(ies)")
                for _, vuln in sorted(vulns.items()):
                    print(f"        {vuln['id']}  {vuln['module']} {vuln['found']} -> {vuln['fixed']}  {vuln['summary']}")
                    if vuln["module"] == "stdlib":
                        stdlib_fixes.add(vuln["fixed"])
                failed += 1
                continue

            print(f"PASS  {name}")

    print()
    if failed:
        print(f"govulncheck: {failed} of {len(modules)} module(s) failed")
        if stdlib_fixes:
            newest = max(stdlib_fixes, key=lambda v: [int(x) for x in re.findall(r"\d+", v)])
            print(f"stdlib findings are fixed by building with Go {newest.lstrip('v')} or newer")
        print("details: run `govulncheck ./...` in the failing module")
        return 1
    print(f"govulncheck: all {len(modules)} module(s) passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
