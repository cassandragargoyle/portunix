#!/bin/bash
# github-publish.sh
# Interactive script for publishing to GitHub from local Gitea development
# Based on portunix-cleanup-public.ps1 strategy

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REMOTE="github"
GITHUB_REPO="https://github.com/cassandragargoyle/Portunix.git"
RELEASE_BRANCH="release-$(date +%Y%m%d-%H%M%S)"
TEMP_DIR="../portunix-github-staging"

# Files/directories to remove before GitHub publish (based on cleanup-public.ps1)
PRIVATE_FILES=(
    "CLAUDE.md"
    "GEMINI.md" 
    "NOTES.md"
    "bin/"
    "*.exe"
    "docs/private/"
    "config/dev/"
    "package.portunix.linux.bat"
    "package.portunix.windows.bat"
    "build.portunix.linux.arm.bat"
    "build.portunix.linux.bat"
    "build.portunix.linux.sh"
    "app/service_lnx.go"
    "cmd/login.go"
    "build.portunix.windows.bat"
    "scripts/package-win.ps1"
)

print_header() {
    echo -e "${BLUE}================================"
    echo -e "🚀 PORTUNIX GITHUB PUBLISHER"
    echo -e "================================${NC}"
    echo
}

print_step() {
    echo -e "${GREEN}📋 Step $1: $2${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

check_prerequisites() {
    print_step "1" "Checking prerequisites"
    
    # Check if we're in git repo
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
    
    # Check if GitHub remote exists
    if ! git remote get-url $GITHUB_REMOTE > /dev/null 2>&1; then
        print_warning "GitHub remote '$GITHUB_REMOTE' not found"
        echo -e "Adding GitHub remote: $GITHUB_REPO"
        git remote add $GITHUB_REMOTE $GITHUB_REPO
        echo -e "${GREEN}✓ GitHub remote added${NC}"
    else
        echo -e "${GREEN}✓ GitHub remote exists${NC}"
    fi
    
    # Check working directory status
    if [ -n "$(git status --porcelain)" ]; then
        print_warning "Working directory has uncommitted changes"
        echo "Please commit or stash changes before continuing"
        echo
        git status --short
        echo
        read -p "Continue anyway? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo -e "${GREEN}✓ Working directory clean${NC}"
    fi
    echo
}

show_changes_summary() {
    print_step "2" "Analyzing changes for publication"
    
    # Get last published commit (if any)
    LAST_GITHUB_COMMIT=""
    if git branch -r | grep -q "$GITHUB_REMOTE/main"; then
        git fetch $GITHUB_REMOTE main --quiet 2>/dev/null || true
        LAST_GITHUB_COMMIT=$(git rev-parse $GITHUB_REMOTE/main 2>/dev/null || echo "")
    fi
    
    if [ -n "$LAST_GITHUB_COMMIT" ]; then
        echo "📊 Changes since last GitHub publish:"
        echo "   Last GitHub commit: $(git log --oneline -1 $LAST_GITHUB_COMMIT 2>/dev/null || echo 'Not found')"
        echo "   Current commit: $(git log --oneline -1 HEAD)"
        echo
        echo "📝 Commits to be published:"
        if git rev-list $LAST_GITHUB_COMMIT..HEAD --count > /dev/null 2>&1; then
            COMMIT_COUNT=$(git rev-list $LAST_GITHUB_COMMIT..HEAD --count)
            echo "   Total commits: $COMMIT_COUNT"
            echo
            git log --oneline $LAST_GITHUB_COMMIT..HEAD | head -10
            if [ $COMMIT_COUNT -gt 10 ]; then
                echo "   ... and $(($COMMIT_COUNT - 10)) more commits"
            fi
        else
            echo "   Could not determine commit range"
            echo "   Will publish current state"
        fi
    else
        echo "📊 First time publishing to GitHub"
        echo "   Current commit: $(git log --oneline -1 HEAD)"
        echo "   Total commits to publish: $(git rev-list --count HEAD)"
    fi
    echo
}

create_release_commit() {
    print_step "3" "Creating release commit"
    
    # Create temporary staging area
    if [ -d "$TEMP_DIR" ]; then
        print_warning "Staging directory exists, removing..."
        rm -rf "$TEMP_DIR"
    fi
    
    echo "📁 Creating staging area: $TEMP_DIR"
    git clone . "$TEMP_DIR" --quiet
    cd "$TEMP_DIR"
    
    # Remove private files
    echo "🧹 Cleaning private files..."
    removed_count=0
    for pattern in "${PRIVATE_FILES[@]}"; do
        # Use find to handle patterns properly
        if [[ "$pattern" == *"*"* ]]; then
            # Handle glob patterns
            find . -name "$pattern" -type f -delete 2>/dev/null && removed_count=$((removed_count + 1)) || true
        else
            # Handle exact paths
            if [ -e "$pattern" ]; then
                rm -rf "$pattern"
                removed_count=$((removed_count + 1))
                echo "   ✓ Removed: $pattern"
            fi
        fi
    done
    
    if [ $removed_count -gt 0 ]; then
        echo -e "${GREEN}✓ Removed $removed_count private files/directories${NC}"
    else
        echo -e "${YELLOW}⚠️  No private files found to remove${NC}"
    fi
    
    # Check if there are changes after cleanup
    if [ -n "$(git status --porcelain)" ]; then
        git add -A
        git commit -m "cleanup: remove private files for GitHub publication"
        echo -e "${GREEN}✓ Cleanup commit created${NC}"
    fi
    
    echo
}

prepare_release_message() {
    print_step "4" "Preparing release commit message"
    
    # Default release message
    DEFAULT_TITLE="feat: publish development changes"
    DEFAULT_BODY="Summary of changes since last GitHub release:

$(git log --oneline HEAD~5..HEAD 2>/dev/null | sed 's/^/- /' || echo '- Development updates')
"
    
    echo "✏️  Enter release commit details:"
    echo
    read -p "Release title [$DEFAULT_TITLE]: " RELEASE_TITLE
    RELEASE_TITLE=${RELEASE_TITLE:-$DEFAULT_TITLE}
    
    echo
    echo "Release description (multi-line, end with empty line):"
    echo "Default: [press Enter to use default]"
    read -r first_line
    if [ -z "$first_line" ]; then
        RELEASE_BODY="$DEFAULT_BODY"
    else
        RELEASE_BODY="$first_line"$'\n'
        while IFS= read -r line; do
            if [ -z "$line" ]; then
                break
            fi
            RELEASE_BODY="$RELEASE_BODY$line"$'\n'
        done
        RELEASE_BODY="$RELEASE_BODY"$'\n'
    fi
    
    echo
    echo "📋 Release commit will be:"
    echo "Title: $RELEASE_TITLE"
    echo "Body:"
    echo "$RELEASE_BODY"
    echo
}

create_final_commit() {
    print_step "5" "Creating final release commit"
    
    # Create squashed commit with all changes
    FULL_MESSAGE="$RELEASE_TITLE

$RELEASE_BODY"
    
    # Reset to clean state and commit everything as one commit
    INITIAL_COMMIT=$(git rev-list --max-parents=0 HEAD)
    git reset --soft $INITIAL_COMMIT
    git commit -m "$FULL_MESSAGE"
    
    echo -e "${GREEN}✓ Release commit created${NC}"
    echo "   Commit: $(git log --oneline -1 HEAD)"
    echo
}

publish_to_github() {
    print_step "6" "Publishing to GitHub"
    
    echo "🚀 Ready to publish to GitHub"
    echo "   Target: $GITHUB_REPO"
    echo "   Branch: main"
    echo
    read -p "Proceed with publication? [y/N]: " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📤 Pushing to GitHub..."
        git push $GITHUB_REMOTE HEAD:main --force-with-lease
        echo -e "${GREEN}✅ Successfully published to GitHub!${NC}"
        echo
        echo "🔗 GitHub repository: $GITHUB_REPO"
        echo "📊 Latest commit: $(git log --oneline -1 HEAD)"
    else
        echo -e "${YELLOW}⏸️  Publication cancelled${NC}"
        echo "   Staging directory preserved: $TEMP_DIR"
        echo "   You can review and push manually if needed"
    fi
    echo
}

cleanup_staging() {
    print_step "7" "Cleanup"
    
    cd ..
    if [ -d "$TEMP_DIR" ]; then
        read -p "Remove staging directory? [Y/n]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            rm -rf "$TEMP_DIR"
            echo -e "${GREEN}✓ Staging directory removed${NC}"
        else
            echo -e "${YELLOW}⚠️  Staging directory preserved: $TEMP_DIR${NC}"
        fi
    fi
}

# Main execution
main() {
    print_header
    
    check_prerequisites
    show_changes_summary
    
    read -p "Continue with GitHub publication? [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Publication cancelled"
        exit 0
    fi
    
    create_release_commit
    prepare_release_message
    create_final_commit
    publish_to_github
    cleanup_staging
    
    echo -e "${GREEN}🎉 GitHub publication workflow completed!${NC}"
}

# Run main function
main "$@"