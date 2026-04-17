# Documentation Scripts

Portunix includes two scripts for managing documentation:

## 📚 post-release-docs.py

**Purpose**: Generate static documentation from Portunix commands using Hugo

**Usage**:

```bash
# Generate documentation
python3 scripts/post-release-docs.py v1.5.0

# Local dev server
python3 scripts/post-release-docs.py --serve

# Build only without server
python3 scripts/post-release-docs.py --build-only
```

**What it does**:

- ✅ Checks dependencies (Hugo, Portunix binary)
- ✅ Installs Hugo automatically via Portunix if missing
- ✅ Creates the Hugo site structure
- ✅ Generates documentation for all core commands
- ✅ Discovers plugin commands (partial)
- ✅ Creates release notes
- ✅ Builds static HTML pages into `docs-site/public/`

**Output**: `docs-site/public/` – ready for publishing

---

## 🚀 publish-docs-to-github.py

**Purpose**: Publish documentation to GitHub Pages

**Usage**:

```bash
# Publish documentation
python3 scripts/publish-docs-to-github.py v1.5.0

# Dry run (test without pushing)
python3 scripts/publish-docs-to-github.py v1.5.0 --dry-run

# Custom commit message
python3 scripts/publish-docs-to-github.py v1.5.0 -m "Update docs for new features"

# Skip checks (careful!)
python3 scripts/publish-docs-to-github.py v1.5.0 --skip-checks
```

**What it does**:

- ✅ Checks GitHub CLI (installs automatically via Portunix)
- ✅ Verifies authentication with `gh auth status`
- ✅ Checks the git repository and GitHub remote
- ✅ Ensures `docs-site/public/` exists (from post-release-docs.py)
- ✅ Creates/updates the `gh-pages` branch
- ✅ Copies documentation and commits changes
- ✅ Pushes to GitHub Pages
- ✅ Shows the final documentation URL

**Requirements**:

- `post-release-docs.py` must be run (needs `docs-site/public/`)
- GitHub CLI authentication: `gh auth login`
- Git repository with a GitHub remote

---

## 🔄 Complete workflow

```bash
# 1. Generate documentation
python3 scripts/post-release-docs.py v1.5.0

# 2. Check locally (optional)
python3 scripts/post-release-docs.py --serve
# Open http://localhost:1313

# 3. Publish to GitHub Pages
python3 scripts/publish-docs-to-github.py v1.5.0
```

---

## ⚙️ Automatic dependencies

Both scripts automatically install their dependencies via Portunix:

- **Hugo**: `portunix install hugo`
- **GitHub CLI**: `portunix install github-cli`

### First run

1. **Build Portunix**: `go build -o .`
2. **Authenticate GitHub**: `gh auth login`
3. **Run the workflow** above

---

## 📂 File structure

```text
docs-site/                 # Hugo site
├── content/               # Markdown content
│   ├── commands/          # Generated command docs
│   │   ├── core/          # Core commands
│   │   └── plugins/       # Plugin commands
│   ├── guides/            # Manual guides
│   └── releases/          # Release notes
├── themes/portunix-docs/  # Hugo theme
├── public/                # Built HTML (gitignored)
└── hugo.toml              # Hugo configuration
```

---

## 🔧 Troubleshooting

### GitHub CLI not authenticated

```bash
gh auth login
# Choose: Login with a web browser
```

### Hugo installation failed

```bash
portunix install hugo --variant extended
```

### Git remote missing

```bash
git remote add origin https://github.com/cassandragargoyle/portunix.git
```

### Documentation not built

```bash
python3 scripts/post-release-docs.py v1.5.0
```

---

## 🖥️ Local documentation server

For quick local preview:

```bash
# Linux/macOS
./scripts/docs-serve.sh

# Windows CMD
scripts\docs-serve.cmd

# Windows PowerShell
.\scripts\docs-serve.ps1
```

**Parameters:**

- No parameters: Hugo server at `http://localhost:1313` (hot reload)
- `--static` / `-Static`: Python HTTP server at `http://localhost:8080` (serves `public/` only)

---

## 📡 GitHub Pages URL

After successful publishing, the documentation is available at:
**<https://cassandragargoyle.github.io/portunix/>**

GitHub Pages may take 2–3 minutes to update.
