#!/usr/bin/env python3
"""
Portunix Release Notes Collector

Collects release notes from JSON files and generates Markdown documentation.
Designed to integrate with make dist workflow.

Usage:
    python3 collect_release_notes.py              # Generate RELEASE-NOTES.md
    python3 collect_release_notes.py --check      # Check for missing JSON files
    python3 collect_release_notes.py --warn-only  # With --check, only warn (don't fail)
    python3 collect_release_notes.py --list-missing  # List versions without JSON
    python3 collect_release_notes.py --version 1.8.0  # Generate for specific version
    python3 collect_release_notes.py --output dist/  # Output to specific directory
"""

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional


# Project root relative to this script
SCRIPT_DIR = Path(__file__).parent.resolve()
PROJECT_ROOT = SCRIPT_DIR.parent.parent
RELEASE_NOTES_DIR = PROJECT_ROOT / "release-notes"
SCHEMA_FILE = RELEASE_NOTES_DIR / "_schema.json"


def get_git_tags() -> list[str]:
    """Get all version tags from git, sorted by version number."""
    try:
        result = subprocess.run(
            ["git", "tag", "-l", "v*", "--sort=-v:refname"],
            capture_output=True,
            text=True,
            cwd=PROJECT_ROOT,
        )
        if result.returncode != 0:
            print(f"Warning: Could not get git tags: {result.stderr}", file=sys.stderr)
            return []

        tags = [t.strip() for t in result.stdout.strip().split("\n") if t.strip()]
        # Filter only valid version tags (vX.Y.Z)
        valid_tags = []
        for tag in tags:
            parts = tag[1:].split(".")  # Remove 'v' prefix
            if len(parts) == 3 and all(p.isdigit() for p in parts):
                valid_tags.append(tag)
        return valid_tags
    except FileNotFoundError:
        print("Warning: git not found", file=sys.stderr)
        return []


def get_existing_json_versions() -> set[str]:
    """Get set of versions that have JSON files."""
    if not RELEASE_NOTES_DIR.exists():
        return set()

    versions = set()
    for f in RELEASE_NOTES_DIR.glob("*.json"):
        if f.name.startswith("_"):  # Skip schema and other special files
            continue
        # Remove .json extension to get version
        version = f.stem
        versions.add(version)
    return versions


def tag_to_version(tag: str) -> str:
    """Convert git tag (v1.8.0) to version string (1.8.0)."""
    return tag[1:] if tag.startswith("v") else tag


def version_to_tag(version: str) -> str:
    """Convert version string (1.8.0) to git tag (v1.8.0)."""
    return f"v{version}" if not version.startswith("v") else version


def load_release_notes(version: str) -> Optional[dict]:
    """Load release notes JSON for a specific version."""
    json_file = RELEASE_NOTES_DIR / f"{version}.json"
    if not json_file.exists():
        return None

    try:
        with open(json_file, "r", encoding="utf-8") as f:
            return json.load(f)
    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON in {json_file}: {e}", file=sys.stderr)
        return None


def validate_json(data: dict, version: str) -> list[str]:
    """Basic validation of release notes JSON. Returns list of errors."""
    errors = []

    # Required fields
    for field in ["version", "date", "tag"]:
        if field not in data:
            errors.append(f"Missing required field: {field}")

    # Version consistency
    if data.get("version") != version:
        errors.append(f"Version mismatch: file is {version}.json but contains version {data.get('version')}")

    # Tag format
    if "tag" in data and not data["tag"].startswith("v"):
        errors.append(f"Tag should start with 'v': {data['tag']}")

    return errors


def generate_markdown_for_version(data: dict) -> str:
    """Generate Markdown content for a single version."""
    lines = []

    version = data.get("version", "Unknown")
    date = data.get("date", "Unknown")
    summary = data.get("summary", "")

    lines.append(f"## {version}")
    lines.append("")
    lines.append(f"**Release Date:** {date}")
    lines.append("")

    if summary:
        lines.append(summary)
        lines.append("")

    # Highlights
    highlights = data.get("highlights", [])
    if highlights:
        lines.append("### Highlights")
        lines.append("")
        for h in highlights:
            lines.append(f"- {h}")
        lines.append("")

    # Changes by category
    changes = data.get("changes", {})

    category_titles = {
        "breaking": "Breaking Changes",
        "security": "Security",
        "features": "New Features",
        "improvements": "Improvements",
        "fixes": "Bug Fixes",
        "docs": "Documentation",
    }

    # Order matters - breaking and security first
    category_order = ["breaking", "security", "features", "improvements", "fixes", "docs"]

    for category in category_order:
        items = changes.get(category, [])
        if items:
            title = category_titles.get(category, category.title())
            lines.append(f"### {title}")
            lines.append("")
            for item in items:
                desc = item.get("description", "")
                issue = item.get("issue", "")
                if issue:
                    lines.append(f"- {desc} ({issue})")
                else:
                    lines.append(f"- {desc}")
            lines.append("")

    # Components
    components = data.get("components", [])
    if components:
        lines.append("### Affected Components")
        lines.append("")
        lines.append(", ".join(components))
        lines.append("")

    # Additional notes
    notes = data.get("notes", "")
    if notes:
        lines.append("### Notes")
        lines.append("")
        lines.append(notes)
        lines.append("")

    return "\n".join(lines)


def generate_full_release_notes(versions: list[str] = None) -> str:
    """Generate complete RELEASE-NOTES.md content."""
    lines = []

    lines.append("# Portunix Release Notes")
    lines.append("")
    lines.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append("")
    lines.append("---")
    lines.append("")

    # Get all versions with JSON files, sorted by version number (newest first)
    existing = get_existing_json_versions()

    if versions:
        # Filter to requested versions
        to_generate = [v for v in versions if v in existing]
    else:
        to_generate = sorted(existing, key=lambda v: [int(x) for x in v.split(".")], reverse=True)

    if not to_generate:
        lines.append("No release notes available.")
        lines.append("")
        return "\n".join(lines)

    for version in to_generate:
        data = load_release_notes(version)
        if data:
            errors = validate_json(data, version)
            if errors:
                print(f"Warning: Validation errors in {version}.json:", file=sys.stderr)
                for e in errors:
                    print(f"  - {e}", file=sys.stderr)

            lines.append(generate_markdown_for_version(data))
            lines.append("---")
            lines.append("")

    return "\n".join(lines)


def check_missing_notes(warn_only: bool = False) -> int:
    """Check for missing release notes JSON files. Returns exit code."""
    tags = get_git_tags()
    existing = get_existing_json_versions()

    missing = []
    for tag in tags:
        version = tag_to_version(tag)
        if version not in existing:
            missing.append(version)

    if missing:
        print(f"{'Warning' if warn_only else 'Error'}: Missing release notes for {len(missing)} version(s):")
        for v in missing:
            print(f"  - {v} (tag: v{v})")
        print("")
        print("To generate missing release notes, run:")
        print("  /cs:generate-release-notes")
        print("")

        if not warn_only:
            return 1
    else:
        print("All versions have release notes JSON files.")

    return 0


def list_missing() -> None:
    """List versions without JSON files."""
    tags = get_git_tags()
    existing = get_existing_json_versions()

    print("Versions without release notes JSON:")
    found = False
    for tag in tags:
        version = tag_to_version(tag)
        if version not in existing:
            print(f"  {version}")
            found = True

    if not found:
        print("  (none - all versions have JSON files)")


def main():
    parser = argparse.ArgumentParser(
        description="Collect and generate release notes for Portunix",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    %(prog)s                    # Generate RELEASE-NOTES.md in current directory
    %(prog)s --check            # Check for missing JSON files (fails if missing)
    %(prog)s --check --warn-only  # Check but only warn
    %(prog)s --list-missing     # List versions without JSON
    %(prog)s --version 1.8.0    # Generate for specific version
    %(prog)s --output dist/     # Output to specific directory
        """
    )

    parser.add_argument(
        "--check",
        action="store_true",
        help="Check that all git tags have corresponding JSON files"
    )
    parser.add_argument(
        "--warn-only",
        action="store_true",
        help="With --check, only warn about missing files (don't fail)"
    )
    parser.add_argument(
        "--list-missing",
        action="store_true",
        help="List versions that don't have JSON files"
    )
    parser.add_argument(
        "--version",
        type=str,
        help="Generate release notes for specific version only"
    )
    parser.add_argument(
        "--output",
        type=str,
        default=".",
        help="Output directory for RELEASE-NOTES.md (default: current directory)"
    )
    parser.add_argument(
        "--filename",
        type=str,
        default="RELEASE-NOTES.md",
        help="Output filename (default: RELEASE-NOTES.md)"
    )

    args = parser.parse_args()

    # Ensure release-notes directory exists
    if not RELEASE_NOTES_DIR.exists():
        print(f"Creating release-notes directory: {RELEASE_NOTES_DIR}")
        RELEASE_NOTES_DIR.mkdir(parents=True)

    # Handle different modes
    if args.list_missing:
        list_missing()
        return 0

    if args.check:
        return check_missing_notes(warn_only=args.warn_only)

    # Generate release notes
    versions = [args.version] if args.version else None
    content = generate_full_release_notes(versions)

    # Write output
    output_dir = Path(args.output)
    if not output_dir.exists():
        output_dir.mkdir(parents=True)

    output_file = output_dir / args.filename
    with open(output_file, "w", encoding="utf-8") as f:
        f.write(content)

    print(f"Generated: {output_file}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
