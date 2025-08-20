#!/bin/bash
# setup-github-remotes.sh
# One-time setup script for GitHub publishing workflow

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🔧 Setting up GitHub publishing workflow${NC}"
echo

# Check current remotes
echo "📍 Current git remotes:"
git remote -v
echo

# Add GitHub remote if not exists
GITHUB_REMOTE="github"
GITHUB_REPO="https://github.com/cassandragargoyle/Portunix.git"

if ! git remote get-url $GITHUB_REMOTE > /dev/null 2>&1; then
    echo -e "${YELLOW}Adding GitHub remote...${NC}"
    git remote add $GITHUB_REMOTE $GITHUB_REPO
    echo -e "${GREEN}✓ GitHub remote added${NC}"
else
    echo -e "${GREEN}✓ GitHub remote already exists${NC}"
fi

# Make scripts executable
chmod +x scripts/github-*.sh

echo
echo -e "${GREEN}🎉 Setup complete!${NC}"
echo
echo "📋 Available commands:"
echo "  ./scripts/github-02-sync-publish.sh   - Enhanced sync & publish workflow"
echo "  ./scripts/github-02-quick-publish.sh  - Quick squash publish (legacy)"
echo "  git remote -v                         - View all remotes"
echo
echo "💡 Enhanced Usage workflow:"
echo "  1. Develop locally, commit to your Gitea"
echo "  2. When ready to publish: ./scripts/github-02-sync-publish.sh"
echo "  3. Script will:"
echo "     • Fetch current GitHub state"
echo "     • Help create meaningful branch name"
echo "     • Sync your files (excluding private)"
echo "     • Publish as feature branch"
echo