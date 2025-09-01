#!/bin/bash
# github-sync-publish.sh
# Enhanced workflow: GitHub sync → branch creation → file sync → publish

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
GITHUB_REMOTE="github"
GITHUB_REPO="https://github.com/cassandragargoyle/Portunix.git"
LOCAL_REPO_PATH="$(pwd)"
GITHUB_WORK_DIR="../portunix-github-sync"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Private files to exclude (same as cleanup script)
PRIVATE_FILES=(
    "CLAUDE.md"
    "GEMINI.md" 
    "NOTES.md"
    "bin/"
    "*.exe"
    "docs/private/"
    "docs/issues/internal/"
    "config/dev/"
    "./scripts/"
    ".claude/"
    "internal/testutils/"
    "package.portunix.linux.bat"
    "package.portunix.windows.bat"
    "build.portunix.linux.arm.bat"
    "build.portunix.linux.bat"
    "build.portunix.linux.sh"
    "app/service_lnx.go"
    "cmd/login.go"
    "build.portunix.windows.bat"
    "scripts/package-win.ps1"
    "test/venv/"
    "test/__pycache__/"
    "test/integration/__pycache__/"
    "test/results/"
    "*.pyc"
    "*.pyo"
    ".pytest_cache/"
    ".cache/"
    "dist/"
    ".tmp/"
)

print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════╗"
    echo -e "║       🚀 PORTUNIX SYNC & PUBLISH       ║"
    echo -e "║     GitHub Sync → Branch → Publish     ║"
    echo -e "╚════════════════════════════════════════╝${NC}"
    echo
}

print_step() {
    echo -e "${GREEN}📋 Step $1: $2${NC}"
    echo
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

step1_fetch_github() {
    print_step "1" "Fetching current GitHub state"
    
    # Setup GitHub remote if needed
    if ! git remote get-url $GITHUB_REMOTE > /dev/null 2>&1; then
        print_info "Adding GitHub remote: $GITHUB_REPO"
        git remote add $GITHUB_REMOTE $GITHUB_REPO
    fi
    
    # Create clean working directory
    if [ -d "$GITHUB_WORK_DIR" ]; then
        print_warning "Removing existing sync directory..."
        rm -rf "$GITHUB_WORK_DIR"
    fi
    
    print_info "Cloning GitHub repository to: $GITHUB_WORK_DIR"
    git clone $GITHUB_REPO "$GITHUB_WORK_DIR" --quiet || { print_error "Failed to clone GitHub repository"; exit 1; }
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    echo -e "${GREEN}✓ GitHub repository cloned${NC}"
    echo -e "   Latest commit: $(git log --oneline -1 HEAD)"
    echo
}

step2_analyze_changes() {
    print_step "2" "Analyzing local changes for publication"
    
    cd "$LOCAL_REPO_PATH"
    
    echo "📊 Local repository analysis:"
    echo -e "   Current branch: ${CYAN}$(git branch --show-current)${NC}"
    echo -e "   Latest commit: $(git log --oneline -1 HEAD)"
    echo -e "   Total commits: $(git rev-list --count HEAD)"
    echo
    
    # Get recent commits for branch naming inspiration
    echo "📝 Recent local commits (for branch naming):"
    git log --oneline -5 HEAD | sed 's/^/   /'
    echo
    
    # Count files that will be copied
    echo "📁 Files analysis:"
    total_files=$(find . -type f -not -path './.git/*' | wc -l)
    echo "   Total files: $total_files"
    
    # Estimate private files
    private_count=0
    for pattern in "${PRIVATE_FILES[@]}"; do
        if [[ "$pattern" == *"*"* ]]; then
            count=$(find . -name "$pattern" -type f 2>/dev/null | wc -l)
        else
            [ -e "$pattern" ] && count=1 || count=0
        fi
        private_count=$((private_count + count))
    done
    echo "   Private files (will be excluded): $private_count"
    echo "   Files to publish: $((total_files - private_count))"
    echo
}

process_public_issues() {
    # This function processes public issues based on mapping.json
    # It copies only issues marked for publication from internal/ to public/
    
    if [ ! -f "$GITHUB_WORK_DIR/docs/issues/public/mapping.json" ]; then
        return
    fi
    
    # Read mapping and process public issues
    if command -v jq &> /dev/null; then
        # If jq is available, use it for parsing JSON
        for pub_id in $(jq -r '.mappings | keys[]' "$GITHUB_WORK_DIR/docs/issues/public/mapping.json" 2>/dev/null); do
            internal_id=$(jq -r ".mappings.\"$pub_id\".internal" "$GITHUB_WORK_DIR/docs/issues/public/mapping.json")
            published=$(jq -r ".mappings.\"$pub_id\".published" "$GITHUB_WORK_DIR/docs/issues/public/mapping.json")
            
            if [ "$published" = "true" ] && [ -f "$LOCAL_REPO_PATH/docs/issues/internal/${internal_id}-"*.md ]; then
                # Copy internal issue to public with PUB number
                source_file=$(ls "$LOCAL_REPO_PATH/docs/issues/internal/${internal_id}-"*.md 2>/dev/null | head -1)
                if [ -f "$source_file" ]; then
                    target_file="$GITHUB_WORK_DIR/docs/issues/${pub_id}-$(basename "$source_file" | sed "s/^${internal_id}-//")"
                    cp "$source_file" "$target_file"
                    print_info "Published: $pub_id (from internal #$internal_id)"
                fi
            fi
        done
    else
        print_warning "jq not installed - skipping automatic issue processing"
        print_info "Install jq for automatic public issue processing: sudo apt-get install jq"
    fi
    
    # Update README to show only public issues
    if [ -f "$GITHUB_WORK_DIR/docs/issues/README.md" ]; then
        # Create public-facing README with only published issues
        sed -i '/| #[0-9]* | - |/d' "$GITHUB_WORK_DIR/docs/issues/README.md" 2>/dev/null || true
        sed -i 's|internal/||g' "$GITHUB_WORK_DIR/docs/issues/README.md" 2>/dev/null || true
    fi
}

step3_create_branch() {
    print_step "3" "Creating feature branch with help"
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    echo "🤖 Let's create a meaningful branch name based on your changes..."
    echo
    echo "📋 Recent changes summary:"
    cd "$LOCAL_REPO_PATH"
    echo "$(git log --oneline -5 HEAD | sed 's/^/   /')"
    echo
    
    # Interactive branch naming with suggestions
    echo -e "${CYAN}🎯 Branch naming suggestions:${NC}"
    
    # Analyze commits to suggest branch names
    recent_messages=$(git log --oneline -5 HEAD | cut -d' ' -f2- | tr '\n' ' ')
    
    # Generate suggestions based on patterns
    suggestions=()
    if [[ $recent_messages =~ [Dd]ocker ]]; then
        suggestions+=("feature/docker-management-improvements")
        suggestions+=("feature/docker-auto-install")
    fi
    if [[ $recent_messages =~ [Tt]est ]]; then
        suggestions+=("feature/testing-enhancements")
    fi
    if [[ $recent_messages =~ [Ff]ix ]]; then
        suggestions+=("fix/bug-fixes-${TIMESTAMP}")
    fi
    if [[ $recent_messages =~ [Ff]eat ]]; then
        suggestions+=("feature/new-features-${TIMESTAMP}")
    fi
    
    # Default suggestions
    suggestions+=("feature/development-sync-${TIMESTAMP}")
    suggestions+=("update/codebase-improvements")
    
    echo "Generated suggestions:"
    for i in "${!suggestions[@]}"; do
        echo "   $((i+1)). ${suggestions[i]}"
    done
    echo "   0. Custom branch name"
    echo
    
    read -p "Select branch name [1]: " choice
    choice=${choice:-1}
    
    if [ "$choice" -eq 0 ] 2>/dev/null; then
        read -p "Enter custom branch name: " BRANCH_NAME
    elif [ "$choice" -ge 1 ] && [ "$choice" -le "${#suggestions[@]}" ] 2>/dev/null; then
        BRANCH_NAME="${suggestions[$((choice-1))]}"
    else
        BRANCH_NAME="${suggestions[0]}"
    fi
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    print_info "Creating branch: $BRANCH_NAME"
    git checkout -b "$BRANCH_NAME"
    echo -e "${GREEN}✓ Branch created: $BRANCH_NAME${NC}"
    echo
}

step4_sync_files() {
    print_step "4" "Syncing files from local repository"
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    print_info "Copying files from local repository..."
    print_info "Source: $LOCAL_REPO_PATH"
    print_info "Target: $GITHUB_WORK_DIR"
    
    # Clear existing files (except .git)
    find . -maxdepth 1 -not -name '.' -not -name '.git' -exec rm -rf {} + 2>/dev/null || true
    
    # Copy all files from local repo
    cd "$LOCAL_REPO_PATH"
    
    # Use rsync for efficient copying, excluding .git and private files
    exclude_args=()
    for pattern in "${PRIVATE_FILES[@]}"; do
        exclude_args+=(--exclude="$pattern")
    done
    
    rsync -av --exclude='.git/' "${exclude_args[@]}" ./ "$GITHUB_WORK_DIR/" || { print_error "Failed to sync files"; exit 1; }
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    # Process public issues if mapping exists
    if [ -f "$GITHUB_WORK_DIR/docs/issues/public/mapping.json" ]; then
        print_info "Processing public issues..."
        process_public_issues
    fi
    
    # Show what was copied
    echo -e "${GREEN}✓ Files synchronized${NC}"
    
    # Check for changes
    if [ -n "$(git status --porcelain)" ]; then
        echo
        echo "📋 Changes detected:"
        git status --short | head -20
        if [ $(git status --porcelain | wc -l) -gt 20 ]; then
            echo "   ... and $(($(git status --porcelain | wc -l) - 20)) more files"
        fi
    else
        print_warning "No changes detected - files are already in sync"
    fi
    echo
}

step5_create_commit() {
    print_step "5" "Creating commit with detailed description"
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    if [ -z "$(git status --porcelain)" ]; then
        print_warning "No changes to commit"
        return
    fi
    
    echo "✏️  Creating commit message..."
    echo
    
    # Generate default commit message based on local commits
    cd "$LOCAL_REPO_PATH"
    recent_commits=$(git log --oneline -10 HEAD | cut -d' ' -f2-)
    
    # Create smart default message
    DEFAULT_TITLE="feat: sync development changes from local repository"
    DEFAULT_BODY="Summary of integrated changes:

$(echo "$recent_commits" | head -5 | sed 's/^/- /')

Synchronized from local development repository
Branch: $(git branch --show-current)
Total local commits: $(git rev-list --count HEAD)

"
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    echo "📝 Commit details:"
    echo
    read -p "Commit title [$DEFAULT_TITLE]: " COMMIT_TITLE
    COMMIT_TITLE=${COMMIT_TITLE:-$DEFAULT_TITLE}
    
    echo
    echo "Commit description (enter lines, empty line to finish):"
    echo "[Press Enter to use default description]"
    read -r first_line
    if [ -z "$first_line" ]; then
        COMMIT_BODY="$DEFAULT_BODY"
    else
        COMMIT_BODY="$first_line"$'\n'
        while IFS= read -r line; do
            [ -z "$line" ] && break
            COMMIT_BODY="$COMMIT_BODY$line"$'\n'
        done
        COMMIT_BODY="$COMMIT_BODY"$'\n'
    fi
    
    # Create commit
    git add -A
    FULL_MESSAGE="$COMMIT_TITLE

$COMMIT_BODY"
    
    git commit -m "$FULL_MESSAGE"
    
    echo -e "${GREEN}✓ Commit created${NC}"
    echo -e "   Commit: $(git log --oneline -1 HEAD)"
    echo
}

step6_publish() {
    print_step "6" "Publishing to GitHub"
    
    cd "$GITHUB_WORK_DIR" || { print_error "Failed to enter GitHub work directory: $GITHUB_WORK_DIR"; exit 1; }
    
    echo "🚀 Ready to publish to GitHub"
    echo -e "   Repository: ${CYAN}$GITHUB_REPO${NC}"
    echo -e "   Branch: ${CYAN}$BRANCH_NAME${NC}"
    echo -e "   Commit: $(git log --oneline -1 HEAD)"
    echo
    
    read -p "Publish to GitHub? [Y/n]: " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        print_info "Pushing to GitHub..."
        git push origin "$BRANCH_NAME"
        
        echo -e "${GREEN}✅ Successfully published to GitHub!${NC}"
        echo
        echo "🔗 GitHub repository: $GITHUB_REPO"
        echo "🌟 Branch: $BRANCH_NAME"
        echo "📊 Commit: $(git log --oneline -1 HEAD)"
        echo
        echo -e "${CYAN}💡 Next steps:${NC}"
        echo "   • Visit GitHub to create a Pull Request"
        echo "   • Review changes before merging to main"
        echo "   • Delete this branch after merging"
    else
        print_warning "Publication cancelled"
        echo "   Working directory preserved: $GITHUB_WORK_DIR"
        echo "   You can review and push manually"
    fi
    echo
}

step7_cleanup() {
    print_step "7" "Cleanup"
    
    echo "🧹 Cleanup options:"
    echo "   1. Keep working directory for review"
    echo "   2. Remove working directory"
    echo
    read -p "Choose [1]: " cleanup_choice
    cleanup_choice=${cleanup_choice:-1}
    
    if [ "$cleanup_choice" -eq 2 ]; then
        cd "$LOCAL_REPO_PATH"
        rm -rf "$GITHUB_WORK_DIR"
        echo -e "${GREEN}✓ Working directory removed${NC}"
    else
        echo -e "${YELLOW}⚠️  Working directory preserved: $GITHUB_WORK_DIR${NC}"
        echo "   You can continue working or remove it manually"
    fi
}

# Main execution
main() {
    print_header
    
    echo "🎯 This workflow will:"
    echo "   1. Fetch current GitHub state"
    echo "   2. Analyze your local changes"  
    echo "   3. Create a feature branch (with Claude's help)"
    echo "   4. Sync files from local repo (excluding private files)"
    echo "   5. Create a descriptive commit"
    echo "   6. Publish to GitHub"
    echo "   7. Cleanup"
    echo
    
    read -p "Continue? [Y/n]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "Operation cancelled"
        exit 0
    fi
    
    step1_fetch_github
    step2_analyze_changes
    step3_create_branch
    step4_sync_files
    step5_create_commit
    step6_publish
    step7_cleanup
    
    echo -e "${GREEN}🎉 GitHub sync & publish workflow completed!${NC}"
    echo
    echo "📋 Summary:"
    echo "   • Local changes synced to GitHub"
    echo "   • Branch: $BRANCH_NAME"
    echo "   • Ready for Pull Request"
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository"
    exit 1
fi

# Run main function
main "$@"