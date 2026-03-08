#!/bin/bash
# github-01-preflight-check.sh
# Pre-flight check for sensitive data before GitHub publication

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Files that MUST NOT be published (only check these if scanning source directory)
# These should match PRIVATE_FILES in github-02-sync-publish.sh
FORBIDDEN_FILES=(
    "CLAUDE.md"
    "CLAUDE.local.md"
    "GEMINI.md"
    "NOTES.md"
    "TODO.md"
    ".claude/"
    "docs/adr/"
    "docs/private/"
    "docs/issues/internal/"
    "config/dev/"
    "install-from-server.ps1"
    "install-from-server.sh"
    ".translated/"
    ".vscode/"
    ".venv/"
)

# Directories excluded by sync script (don't report as issues, they won't be published)
ALREADY_EXCLUDED=(
    "dist/"
    "bin/"
    "test/e2e/"
    "test/venv/"
    "test/__pycache__/"
    "test/results/"
    ".cache/"
    ".pytest_cache/"
    "aur-package/"
    "aur-repo/"
)

# Patterns in content that indicate sensitive data
SENSITIVE_PATTERNS=(
    "gitea.cassandragargoyle"       # Internal Gitea server
    "cassandragargoyle.cz"          # Internal domain
    "192.168."                       # Local IP addresses
    "10.0."                          # Local IP addresses
    "password.*=.*['\"]"            # Hardcoded passwords
    "api_key.*=.*['\"]"             # API keys
    "secret.*=.*['\"]"              # Secrets
    "token.*=.*['\"]"               # Tokens (but allow generic token references)
    "PRIVATE"                        # Private markers
    "INTERNAL"                       # Internal markers
    "DO NOT PUBLISH"                 # Explicit markers
)

# Allowed exceptions (patterns that look sensitive but are OK)
ALLOWED_EXCEPTIONS=(
    "github.com/cassandragargoyle"  # Public GitHub URL is OK
    "ProductVersion"                 # Version strings OK
    "token string"                   # Go type definitions OK
    "token:"                         # YAML keys OK
    "api_key:"                       # YAML keys OK
    "internal_type"                  # JSON field name OK
    "v1.10.0"                        # Version number OK
    "v1.9.0"                         # Version number OK
    "internally"                     # English word OK
    "Internal"                       # Documentation word OK (capitalized)
)

# Paths to skip during content scanning (already excluded by sync script)
SKIP_PATHS=(
    "./dist/"
    "./.claude/"
    "./docs/adr/"
    "./docs/issues/internal/"
    "./docs/contributing/GITHUB-WORKFLOW.md"
    "./docs/contributing/GITEA-INTERNAL-METHODOLOGY.md"
    "./test/"
    "./.venv/"
    "./.translated/"
)

# Paths where examples are allowed (documentation with code examples)
DOCS_EXAMPLE_PATHS=(
    "./docs/commands/"
    "./docs/ai-assistants/"
    "./docs/contributing/"
    "./assets/templates/"
)

print_header() {
    echo -e "${CYAN}╔════════════════════════════════════════╗"
    echo -e "║     🔍 PRE-FLIGHT SECURITY CHECK       ║"
    echo -e "║   Checking for sensitive data leaks    ║"
    echo -e "╚════════════════════════════════════════╝${NC}"
    echo
}

check_forbidden_files() {
    echo -e "${YELLOW}📁 Checking for forbidden files...${NC}"
    local found=0
    local target_dir="${1:-.}"

    for file in "${FORBIDDEN_FILES[@]}"; do
        if [ -e "$target_dir/$file" ]; then
            echo -e "   ${RED}❌ FOUND: $file${NC}"
            found=$((found + 1))
        fi
    done

    if [ $found -eq 0 ]; then
        echo -e "   ${GREEN}✓ No forbidden files found${NC}"
        return 0
    else
        echo -e "   ${RED}⚠️  Found $found forbidden file(s)${NC}"
        return 1
    fi
}

check_sensitive_content() {
    echo -e "${YELLOW}🔎 Scanning file contents for sensitive patterns...${NC}"
    local found=0
    local target_dir="${1:-.}"
    local temp_file=$(mktemp)

    # Build exclude-dir arguments for paths that will be excluded anyway
    local exclude_args="--exclude-dir=.git"
    for skip_path in "${SKIP_PATHS[@]}"; do
        # Convert ./path/ to path for grep exclude-dir
        local clean_path="${skip_path#./}"
        clean_path="${clean_path%/}"
        exclude_args="$exclude_args --exclude-dir=$clean_path"
    done

    for pattern in "${SENSITIVE_PATTERNS[@]}"; do
        # Search in all text files, excluding .git, binaries, and already-excluded paths
        eval grep -r -l -i "\"$pattern\"" "\"$target_dir\"" \
            --include="*.go" \
            --include="*.md" \
            --include="*.json" \
            --include="*.yaml" \
            --include="*.yml" \
            --include="*.sh" \
            --include="*.ps1" \
            --include="*.py" \
            --include="*.txt" \
            $exclude_args \
            2>/dev/null | while read -r file; do
                # Skip files in excluded paths
                skip_file=false
                for skip_path in "${SKIP_PATHS[@]}"; do
                    if [[ "$file" == $skip_path* ]]; then
                        skip_file=true
                        break
                    fi
                done
                [ "$skip_file" = true ] && continue

                # Skip documentation example paths for non-critical patterns
                # (IP addresses and words like PRIVATE/INTERNAL are OK in docs)
                if [[ "$pattern" =~ ^(192\.168\.|10\.0\.|PRIVATE|INTERNAL)$ ]]; then
                    for doc_path in "${DOCS_EXAMPLE_PATHS[@]}"; do
                        if [[ "$file" == $doc_path* ]]; then
                            skip_file=true
                            break
                        fi
                    done
                    [ "$skip_file" = true ] && continue
                fi

                # Check if it's an allowed exception
                is_exception=false
                for exception in "${ALLOWED_EXCEPTIONS[@]}"; do
                    if grep -q "$exception" "$file" 2>/dev/null; then
                        # Check if the sensitive pattern is part of the exception
                        if grep -i "$pattern" "$file" 2>/dev/null | grep -q "$exception"; then
                            is_exception=true
                            break
                        fi
                    fi
                done

                if [ "$is_exception" = false ]; then
                    echo "$file|$pattern" >> "$temp_file"
                fi
            done
    done

    if [ -s "$temp_file" ]; then
        echo -e "   ${RED}⚠️  Potentially sensitive content found:${NC}"
        cat "$temp_file" | sort -u | while IFS='|' read -r file pattern; do
            echo -e "   ${RED}❌ $file${NC} (pattern: $pattern)"
            # Show the actual line
            grep -n -i "$pattern" "$file" 2>/dev/null | head -3 | sed 's/^/      /'
        done
        found=1
    else
        echo -e "   ${GREEN}✓ No sensitive patterns found${NC}"
    fi

    rm -f "$temp_file"
    return $found
}

check_binary_files() {
    echo -e "${YELLOW}📦 Checking for binary files...${NC}"
    local found=0
    local target_dir="${1:-.}"

    # Check for common binary extensions and executables
    binaries=$(find "$target_dir" -type f \( \
        -name "*.exe" -o \
        -name "*.dll" -o \
        -name "*.so" -o \
        -name "*.dylib" -o \
        -name "portunix" -o \
        -name "ptx-*" \
    \) -not -path "*/.git/*" 2>/dev/null)

    if [ -n "$binaries" ]; then
        echo -e "   ${YELLOW}⚠️  Binary files found:${NC}"
        echo "$binaries" | while read -r file; do
            echo -e "   ${YELLOW}⚠️  $file${NC}"
            found=$((found + 1))
        done
        return 1
    else
        echo -e "   ${GREEN}✓ No binary files found${NC}"
        return 0
    fi
}

check_large_files() {
    echo -e "${YELLOW}📏 Checking for large files (>1MB)...${NC}"
    local target_dir="${1:-.}"

    large_files=$(find "$target_dir" -type f -size +1M -not -path "*/.git/*" 2>/dev/null)

    if [ -n "$large_files" ]; then
        echo -e "   ${YELLOW}⚠️  Large files found:${NC}"
        echo "$large_files" | while read -r file; do
            size=$(du -h "$file" | cut -f1)
            echo -e "   ${YELLOW}⚠️  $file ($size)${NC}"
        done
        return 1
    else
        echo -e "   ${GREEN}✓ No large files found${NC}"
        return 0
    fi
}

check_go_lint() {
    echo -e "${YELLOW}🔍 Running Go lint check...${NC}"
    local target_dir="${1:-.}"

    # Check if golangci-lint is available
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "   ${YELLOW}⚠️  golangci-lint not installed, skipping lint check${NC}"
        echo -e "   ${CYAN}ℹ️  Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
        return 0
    fi

    # Check if .golangci.yml exists
    if [ ! -f "$target_dir/.golangci.yml" ]; then
        echo -e "   ${YELLOW}⚠️  No .golangci.yml found, skipping lint check${NC}"
        return 0
    fi

    # Run golangci-lint
    echo -e "   ${CYAN}ℹ️  Running golangci-lint...${NC}"

    cd "$target_dir"
    lint_output=$(golangci-lint run --timeout 5m 2>&1)
    lint_exit_code=$?
    cd - > /dev/null

    if [ $lint_exit_code -eq 0 ]; then
        echo -e "   ${GREEN}✓ Lint check passed${NC}"
        return 0
    else
        echo -e "   ${RED}❌ Lint check failed${NC}"
        # Show first 20 lines of errors
        echo "$lint_output" | head -20 | while read -r line; do
            echo -e "   ${RED}$line${NC}"
        done

        error_count=$(echo "$lint_output" | grep -c "^[^l]" || echo "0")
        if [ "$error_count" -gt 20 ]; then
            echo -e "   ${YELLOW}... and more errors (run 'golangci-lint run' for full output)${NC}"
        fi
        return 1
    fi
}

generate_report() {
    local target_dir="${1:-.}"
    local report_file="${2:-preflight-report.txt}"

    echo "Pre-flight Security Check Report" > "$report_file"
    echo "=================================" >> "$report_file"
    echo "Date: $(date)" >> "$report_file"
    echo "Directory: $target_dir" >> "$report_file"
    echo "" >> "$report_file"

    echo "Files to be published:" >> "$report_file"
    find "$target_dir" -type f -not -path "*/.git/*" | sort >> "$report_file"

    echo "" >> "$report_file"
    echo "Total files: $(find "$target_dir" -type f -not -path "*/.git/*" | wc -l)" >> "$report_file"

    echo -e "${GREEN}✓ Report saved to: $report_file${NC}"
}

# Main
main() {
    print_header

    local target_dir="${1:-.}"
    local errors=0

    echo "Checking directory: $target_dir"
    echo

    check_forbidden_files "$target_dir" || errors=$((errors + 1))
    echo

    check_sensitive_content "$target_dir" || errors=$((errors + 1))
    echo

    check_binary_files "$target_dir" || errors=$((errors + 1))
    echo

    check_large_files "$target_dir" || errors=$((errors + 1))
    echo

    check_go_lint "$target_dir" || errors=$((errors + 1))
    echo

    echo "════════════════════════════════════════"
    if [ $errors -eq 0 ]; then
        echo -e "${GREEN}✅ PRE-FLIGHT CHECK PASSED${NC}"
        echo "   Safe to publish to GitHub"
        exit 0
    else
        echo -e "${RED}❌ PRE-FLIGHT CHECK FAILED${NC}"
        echo "   Found $errors issue(s) that need review"
        echo
        echo -e "${YELLOW}Options:${NC}"
        echo "   1. Fix the issues and run check again"
        echo "   2. Add false positives to ALLOWED_EXCEPTIONS"
        echo "   3. Continue at your own risk"
        exit 1
    fi
}

# Run if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
