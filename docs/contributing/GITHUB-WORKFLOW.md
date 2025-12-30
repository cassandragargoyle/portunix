# GitHub Publishing Workflow

Portunix development is separated between local development (using local Gitea git server) and GitHub publication. This separation allows for internal development work while maintaining a clean public repository.

## 🎯 Purpose

User created this system for publishing from local Gitea to GitHub. Reason: wanted to separate internal development from public publication.

## 🔧 One-time Setup

```bash
# Run setup script to configure GitHub remote
./scripts/github-00-setup.sh
```

## 📋 Enhanced Workflow

### 1. Local Development
```bash
# Normal development on Gitea
git add .
git commit -m "wip: working on feature X"
git push origin feature-branch
```

### 2. Sync & Publish to GitHub
```bash
# When ready to publish
./scripts/github-02-sync-publish.sh
```

**The new enhanced workflow** will interactively guide you through:
1. 📥 **GitHub Sync** - downloads current state from GitHub to `../portunix-github-sync`
2. 📊 **Analysis** - analyzes local changes for publication
3. 🌿 **Branch Creation** - Claude helps create meaningful branch name
4. 📁 **File Sync** - copies files from local repo, skips private files
5. ✏️ **Commit** - creates descriptive commit
6. 🚀 **Publication** - pushes as feature branch to GitHub
7. 🧹 **Cleanup** - optional removal of staging directory

### 3. Alternative Quick Workflow
```bash
# For quick squash publications (original method)
./scripts/github-02-quick-publish.sh
```

## 📁 Files Removed Before Publication

The following files are automatically excluded from GitHub publication:

- **Private documentation**: `CLAUDE.md`, `GEMINI.md`, `NOTES.md`
- **Binary files**: `bin/`, `*.exe`
- **Private directories**: `docs/private/`, `config/dev/`
- **Build scripts**: Platform-specific `.bat`, `.sh` files
- **Service-specific files**: `app/service_lnx.go`, `cmd/login.go`
- **AI configuration**: `.claude/`
- **Architecture decisions**: `adr/` (internal only)

## 🎛️ Git Remotes Structure

```
origin  -> http://gitea.cassandragargoyle.cz:3000/CassandraGargoyle/portunix (development)
github  -> https://github.com/cassandragargoyle/Portunix/ (publication)
```

## ⚠️ Important Notes

- **Feature branches only**: This workflow creates feature branches, not direct push to main
- **Code review**: Allows code review through GitHub PR before merging
- **Clean history**: GitHub maintains clean, curated history
- **Full history**: Gitea preserves complete development history including WIP commits

## 💡 Tips

- **Development**: Commit often to Gitea, don't worry about WIP commits
- **Release**: Use script for clean GitHub commits
- **Security**: Private files are automatically removed
- **History**: GitHub will have clean history, Gitea preserves everything

## 🔍 Troubleshooting

### GitHub Remote Doesn't Exist
```bash
git remote add github https://github.com/CassandraGargoyle/bootstrap-scripts.git
```

### Push Conflict
The script uses `--force-with-lease` for safety. If someone committed to GitHub in the meantime, the script will stop.

### Wrong Private Files
Edit the list in `github-publish.sh` in the `PRIVATE_FILES` section.

## 🚀 Testing

Test with current changes:
```bash
./scripts/github-publish.sh
```