# Quickstart: Documentation Site with Portunix

Set up a fully working documentation site in minutes using Portunix and Docker containers. No need to install Node.js, Hugo, or Go
on your machine — everything runs inside a container with your project files shared via a mounted folder.

## Prerequisites

- **Portunix** installed ([download latest release](https://github.com/CassandraGargoyle/Portunix/releases/latest))
- **Git** (for cloning your project repository)

> **Note**: Docker or Podman is required for container-based workflows, but Portunix will automatically detect if a container runtime is missing
> and offer to install it for you. No manual Docker installation needed.

## Quick Start

### Option A: One-liner (Windows PowerShell)

```powershell
irm https://github.com/CassandraGargoyle/Portunix/releases/latest/download/quickstart-docusaurus.ps1 | iex
```

### Option B: Step-by-step (Any OS)

```bash
# 1. Initialize a documentation project
portunix playbook init my-docs --template static-docs --engine docusaurus --target container

# 2. Create the documentation site inside a container
portunix playbook run my-docs.ptxbook --script create

# 3. Start the development server with live-reload
portunix playbook run my-docs.ptxbook --script dev
# -> Open http://localhost:3000 in your browser
```

### Option C: Direct Install (local, no container)

```bash
# Install Docusaurus directly (auto-installs Node.js if missing)
portunix install docusaurus

# Or install Hugo
portunix install hugo
```

## Engine Comparison

| Feature | Docusaurus | Hugo | Docsy | Docsify |
| ------- | ---------- | ---- | ----- | ------- |
| **Language** | JavaScript (React) | Go | Go (Hugo + theme) | JavaScript |
| **Best for** | Product docs, API docs | Blogs, corporate sites | Technical project docs | Simple docs |
| **Build speed** | Medium | Very fast | Fast | No build step |
| **Live-reload** | Yes | Yes | Yes | Yes |
| **Search** | Built-in (Algolia) | Plugin-based | Built-in | Plugin-based |
| **Versioning** | Built-in | Manual | Manual | Manual |
| **i18n** | Built-in | Built-in | Built-in | Plugin-based |
| **Port** | 3000 | 1313 | 1313 | 3000 |
| **Requirements** | Node.js | Hugo binary | Hugo + Go | Node.js |

### When to Use What

- **Docusaurus**: You need versioned docs, API documentation, or a React-based site
- **Hugo**: You want the fastest build times and a lightweight setup
- **Docsy**: You're building technical documentation for an open-source project (Google-style)
- **Docsify**: You want zero-build simplicity for a small project

## Step-by-Step Walkthrough

### 1. Choose Your Engine

List available templates and engines:

```bash
portunix playbook template list
portunix playbook template show static-docs
```

### 2. Initialize the Project

```bash
# Docusaurus (default)
portunix playbook init my-docs --template static-docs --engine docusaurus --target container

# Hugo
portunix playbook init my-docs --template static-docs --engine hugo --target container

# Docsy (Hugo + Docsy theme)
portunix playbook init my-docs --template static-docs --engine docsy --target container

# Docsify
portunix playbook init my-docs --template static-docs --engine docsify --target container
```

This generates a `my-docs.ptxbook` file — an infrastructure-as-code definition for your documentation environment.

### 3. Create the Site

```bash
portunix playbook run my-docs.ptxbook --script create
```

This starts a container, installs the documentation engine, and scaffolds a new site in the `./site` (Docusaurus) or `./` (Hugo/Docsy) directory.

### 4. Start Development Server

```bash
portunix playbook run my-docs.ptxbook --script dev
```

Open your browser:

- Docusaurus / Docsify: `http://localhost:3000`
- Hugo / Docsy: `http://localhost:1313`

### 5. Edit Content

Edit files in your local project directory. Changes are reflected immediately in the browser via live-reload — no need to restart the container.

- **Docusaurus**: Edit files in `./site/docs/`
- **Hugo / Docsy**: Edit files in `./content/`
- **Docsify**: Edit files in `./docs/`

### 6. Build for Production

```bash
portunix playbook run my-docs.ptxbook --script build
```

Output is generated in:

- Docusaurus: `./site/build/`
- Hugo / Docsy: `./public/`

## Shared Folder Workflow

Portunix mounts your local project directory into the container as a shared volume. This means:

1. **Local edits are instant** — save a file locally, see the change in the browser
2. **No file copying** — the container reads directly from your disk
3. **Persistent data** — container restarts don't lose your content
4. **Node modules isolated** — Docusaurus uses named volumes for `node_modules` and npm cache, keeping your local directory clean

```text
Local machine          Container
┌──────────────┐      ┌──────────────┐
│ ./site/      │ <--> │ /workspace/  │
│  docs/       │      │  site/       │
│  blog/       │      │  docs/       │
│  ...         │      │  ...         │
└──────────────┘      └──────────────┘
     Edit here           Served here
```

## Available Scripts

List all scripts in your playbook:

```bash
portunix playbook run my-docs.ptxbook --list-scripts
```

| Script | Purpose |
| ------ | ------- |
| `create` | One-time project initialization (run once) |
| `dev` | Start development server with hot reload |
| `build` | Production build (for CI/CD or deployment) |
| `serve` | Serve the production build locally |

## Troubleshooting

### Docker/Podman is not running

Portunix will auto-detect missing container runtimes and offer to install Docker or Podman for you. If Docker is installed but not running, Portunix
will attempt to start Docker Desktop automatically. If it still fails, start Docker Desktop manually and try again.

### Port already in use

```text
Error: port 3000 is already in use
```

Stop the process using the port, or edit the `ports` section in your `.ptxbook` file to use a different port (e.g., `"3001:3000"`).

### Permission errors on Linux

```text
Error: permission denied
```

Ensure your user is in the `docker` group:

```bash
sudo usermod -aG docker $USER
# Log out and back in
```

### Slow first run (Docusaurus)

The first `create` or `dev` run downloads Node.js packages (~200MB). Subsequent runs use cached named volumes and start in ~30 seconds.
If you need to reset the cache:

```bash
docker volume rm my-docs_node_modules my-docs_npm_cache
```

### Hugo module download fails (Docsy)

Docsy uses Hugo Modules which require Go. Ensure the container has internet access. If behind a proxy, configure it in the `.ptxbook` environment section.

### Container won't start

```bash
# Check container status
portunix container list

# Remove stale container and retry
portunix container rm my-docs-docusaurus
portunix playbook run my-docs.ptxbook --script dev
```

## Reference

### Playbook Commands

```bash
portunix playbook init <name> --template static-docs --engine <engine>  # Generate .ptxbook
portunix playbook run <file> --script <name>                            # Run a script
portunix playbook run <file> --list-scripts                             # List scripts
portunix playbook validate <file>                                       # Validate syntax
portunix playbook template list                                         # List templates
portunix playbook template show <name>                                  # Template details
```

### Direct Install Commands

```bash
portunix install docusaurus          # Install Docusaurus (auto-installs Node.js)
portunix install hugo                # Install Hugo
portunix install hugo --variant extended  # Install Hugo Extended
portunix install nodejs              # Install Node.js
```

### Useful Container Commands

```bash
portunix container list              # List running containers
portunix container stop <name>       # Stop a container
portunix container rm <name>         # Remove a container
```
