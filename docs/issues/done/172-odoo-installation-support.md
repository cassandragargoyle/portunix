# Issue #172: Odoo Installation Support

**Type**: Feature
**Priority**: High
**Status**: ✅ Implemented
**Closed**: 2026-04-17
**Labels**: enhancement, package-management, odoo, erp, python, postgresql, container, cross-platform, ptx-installer
**Author**: Architect
**Created**: 2026-04-17

---

## Overview

Add Odoo (open-source ERP / business application suite) as an installable package in
ptx-installer. Odoo installation must support multiple strategies — native package
install from nightly APT repository, build from source (Python + PostgreSQL), and
container-based deployment using the official Odoo Docker images from
<https://github.com/odoo/docker-official-images> (`odoo:19`, `odoo:18`, etc.).

The primary scope is the **Odoo Community core**; the package must be designed so that
additional extensions (custom/OCA addons, Enterprise add-ons) can be layered on top in
follow-up issues without redesigning the core package definition.

## Motivation

- Odoo is the leading open-source ERP platform (50k+ GitHub stars, 13M+ users worldwide,
  4,600+ partners) covering CRM, accounting, inventory, manufacturing, HR, e-commerce,
  and more in a single integrated suite
- Frequently requested by users running SMB business workloads, internal finance/HR
  systems, and e-commerce platforms — commonly deployed alongside other Portunix
  packages (PostgreSQL, Caddy/Nginx reverse proxy, MinIO for attachments)
- Three distinct deployment patterns are in common use, each valid in different
  contexts — Portunix should support all of them:
  1. **Native package install** (`nightly.odoo.com` APT repo) — production Linux
     servers, systemd-managed service
  2. **Build from source** — developer workstations, custom addon development,
     testing against unreleased Odoo branches
  3. **Container deployment** (`odoo:19`) — dev/staging environments, quick demos,
     CI pipelines, isolation from host Python
- Provides a foundation for future Odoo ecosystem packages: Enterprise edition,
  OCA (Odoo Community Association) modules, custom addon scaffolding
- Complements existing Portunix infrastructure packages (PostgreSQL, Docker/Podman,
  MinIO #148) and developer tooling (Python via uv #171, Node.js #041)

## Requirements

### Functional Requirements

- [ ] FR-1: Install Odoo Community via `portunix install odoo` command
- [ ] FR-2: Support three installation variants:
  - `default` — native package install from `nightly.odoo.com` APT repository
    (Debian/Ubuntu)
  - `source` — build from source via git clone + `pip install -r requirements.txt`
    (cross-platform: Linux, Windows, macOS)
  - `container` — official Odoo Docker image (`odoo:19` from
    <https://github.com/odoo/docker-official-images>)
- [ ] FR-3: Support pinning to specific major version (`19`, `18`, `17`) — default to
  latest stable (19 as of April 2026)
- [ ] FR-4: PostgreSQL prerequisite detection and guidance — Odoo requires PostgreSQL
  13+ (14+ recommended); install must verify availability or guide user to install it
  via `portunix install postgresql`
- [ ] FR-5: For `source` variant: detect/install Python 3.10+, Node.js 16+, and
  wkhtmltopdf 0.12.6+ prerequisites
- [ ] FR-6: For `container` variant: provide full lifecycle (pull image, create
  container with linked PostgreSQL, persistent volumes, expose port 8069)
- [ ] FR-7: Post-installation verification — `odoo --version` (package/source) or
  `docker exec <container> odoo --version` (container); HTTP probe against
  `http://localhost:8069/web/database/list`
- [ ] FR-8: Cross-platform support where Odoo itself supports it:
  - Linux (all three variants) — primary target
  - Windows (`source` and `container` variants) — development/testing only
  - macOS (`source` and `container` variants) — development only

### Non-Functional Requirements

- [ ] NFR-1: Follow existing ptx-installer package definition pattern (see
  `minio.json`, `elasticsearch.json`)
- [ ] NFR-2: Auto-discovery by package registry (no manual index updates)
- [ ] NFR-3: Clearly communicate the three variants and their trade-offs in `aiPrompts`
  and help output
- [ ] NFR-4: Security defaults — strong master password guidance, `list_db = False`
  recommendation for production, reverse-proxy reminder
- [ ] NFR-5: Design package definition to allow future layering of addons/extensions
  (custom addons path, OCA modules) without breaking the core install

## Technical Design

### Package Definition Files

Create one primary file in the initial scope:

1. `src/helpers/ptx-installer/assets/packages/odoo.json` — Odoo Community core

Future issues (out of scope here) will add:

- `odoo-enterprise.json` — Enterprise add-ons (requires subscription credentials)
- `odoo-oca-<module-set>.json` — curated OCA module bundles (web, accounting,
  server-tools, etc.)

### Installation Methods — Odoo Community

| Platform | Method | Variant | Command/URL |
| -------- | ------ | ------- | ----------- |
| Debian/Ubuntu | Native APT package | `default` | `nightly.odoo.com/19.0/nightly/deb/` |
| Linux/Windows/macOS | Build from source | `source` | `git clone https://github.com/odoo/odoo.git --branch 19.0 --depth 1` + `pip install -r requirements.txt` |
| Linux/Windows/macOS | Container | `container` | `docker pull odoo:19` (official image from <https://github.com/odoo/docker-official-images>) |

### Source URLs / References

```text
# APT repository (Debian/Ubuntu)
https://nightly.odoo.com/odoo.key
https://nightly.odoo.com/19.0/nightly/deb/

# Source repository
https://github.com/odoo/odoo

# Official Docker images repository
https://github.com/odoo/docker-official-images
docker.io/library/odoo:19    (canonical image ref)
```

### Key Technical Details

- **Current version**: 19.0 (as of April 2026); Odoo 20 expected September 2026
- **License**: LGPL v3 (Community Edition)
- **Runtime stack**:
  - Python 3.10+ (3.12 recommended)
  - PostgreSQL 13+ (14+ recommended) — **exclusive**, no other RDBMS supported
  - Node.js 16+ (for SCSS/JS asset compilation)
  - wkhtmltopdf 0.12.6+ (for PDF report generation)
- **Default port**: 8069 (HTTP), 8072 (longpolling/websocket)
- **Default data paths** (package install):
  - Binary: `/usr/bin/odoo`
  - Config: `/etc/odoo/odoo.conf`
  - Addons: `/usr/lib/python3/dist-packages/odoo/addons/`
  - Data: `/var/lib/odoo/`
  - Systemd unit: `odoo.service`
- **Binary size**: ~400MB (installed, including dependencies); container image ~800MB

### Variant Support

#### `default` — Native APT Package (Debian/Ubuntu)

- Installs Odoo as a systemd-managed service
- Auto-configures `/etc/odoo/odoo.conf` with sensible defaults
- Adds `odoo` system user, creates data directories
- Production-ready on Linux servers
- **Requires**: Debian 11+/Ubuntu 22.04+; PostgreSQL running on localhost or reachable

Install steps (the package definition must orchestrate these):

```bash
wget -O - https://nightly.odoo.com/odoo.key | \
  sudo gpg --dearmor -o /usr/share/keyrings/odoo-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/odoo-archive-keyring.gpg] \
  https://nightly.odoo.com/19.0/nightly/deb/ ./" | \
  sudo tee /etc/apt/sources.list.d/odoo.list
sudo apt update && sudo apt install -y odoo
```

#### `source` — Build from Source

- Git clone of `odoo/odoo` at a pinned branch (default `19.0`)
- Creates a project-local Python virtualenv (via `uv` per ADR-039 / #171)
- Installs Python deps from `requirements.txt`
- Target directory: `$PORTUNIX_DATA/odoo/` (user-scoped) or user-specified path
- Produces a runnable `odoo-bin` entrypoint

Install steps:

```bash
git clone https://github.com/odoo/odoo.git --branch 19.0 --depth 1 <target>
cd <target>
# Linux: install system build deps
sudo apt install -y python3-pip python3-dev libxml2-dev libxslt1-dev \
  libldap2-dev libpq-dev libsasl2-dev libjpeg-dev zlib1g-dev node-less npm
# Create venv + install Python deps
uv venv
uv pip install -r requirements.txt
# Run: python3 odoo-bin --addons-path=addons -d mydb
```

#### `container` — Official Docker Image

- Pulls `odoo:19` from Docker Hub (official image maintained at
  <https://github.com/odoo/docker-official-images>)
- Creates a named container `portunix-odoo` linked to a PostgreSQL container
- Persistent volumes for `/var/lib/odoo` (filestore) and `/mnt/extra-addons`
  (custom addons mount point)
- Exposes port 8069 to host
- Compatible with Docker and Podman via Portunix container abstraction

Example container config (package definition will encode this):

```yaml
image: odoo:19
name: portunix-odoo
ports:
  - "8069:8069"     # HTTP
  - "8072:8072"     # longpolling
environment:
  HOST: portunix-odoo-db       # name of linked PostgreSQL container
  USER: odoo
  PASSWORD: odoo
volumes:
  - odoo-web-data:/var/lib/odoo
  - ./addons:/mnt/extra-addons  (optional, for custom addons)
depends_on:
  - portunix-odoo-db           # PostgreSQL container
detach: true
healthCheck:
  endpoint: http://localhost:8069/web/database/list
  timeout: 120
  interval: 10
  retries: 12
```

**Container variant decision — two approaches to PostgreSQL dependency:**

1. **Compose-style** — create both `portunix-odoo-db` (postgres:16) and
   `portunix-odoo` containers, with Odoo depending on the db container. Simpler UX,
   self-contained.
2. **External-db** — allow user to point at an existing PostgreSQL (`--db-host`,
   `--db-port` flags during install). More flexible, shares db with other services.

The initial implementation should do #1 (compose-style) by default and document the
path to #2 for advanced users.

### Prerequisites

```json
"dependencies": [
  {
    "name": "postgresql",
    "minVersion": "13",
    "purpose": "Odoo exclusive RDBMS — required for all variants",
    "variantScope": ["default", "source"]
  },
  {
    "name": "python",
    "minVersion": "3.10",
    "purpose": "Odoo runtime (source variant only)",
    "variantScope": ["source"]
  },
  {
    "name": "nodejs",
    "minVersion": "16",
    "purpose": "Asset pipeline — SCSS/JS compilation (source variant only)",
    "variantScope": ["source"]
  }
]
```

**Container variant** depends only on Docker/Podman (resolved via Portunix
container abstraction) — the Odoo container image bundles all runtime dependencies.

### Default Configuration (Generated)

For `default` and `source` variants, the package install should generate a baseline
`odoo.conf` matching the manifest recommendation — with a placeholder master
password that the user is prompted to change:

```ini
[options]
admin_passwd = CHANGE_ME_ON_FIRST_RUN
db_host = localhost
db_port = 5432
db_user = odoo
db_password = odoo
addons_path = /usr/lib/python3/dist-packages/odoo/addons
data_dir = /var/lib/odoo
http_port = 8069
workers = 4
max_cron_threads = 2
limit_memory_hard = 2684354560
limit_memory_soft = 2147483648
limit_time_cpu = 600
limit_time_real = 1200
proxy_mode = True
list_db = False
```

### Verification

```bash
# Package install (default variant)
odoo --version
# Expected: Odoo Server 19.0
systemctl status odoo
curl -s http://localhost:8069/web/database/list

# Source install
cd <install-dir>
uv run python3 odoo-bin --version
# Expected: Odoo Server 19.0

# Container install
portunix container list | grep portunix-odoo
docker exec portunix-odoo odoo --version
curl -s http://localhost:8069/web/database/list
```

## Architecture Considerations

### Why Three Variants Are Necessary

Each variant addresses a fundamentally different use case; collapsing them would
make the package unusable for one of the primary audiences:

| Variant | Primary audience | Why this variant |
| ------- | ---------------- | ---------------- |
| `default` | Production sysadmins | systemd-managed service, package updates via apt, standard Linux server workflow |
| `source` | Addon developers | Read/modify Odoo source, run against unreleased branches, scaffold new modules |
| `container` | Dev/staging, demos, CI | Zero-touch deploy, isolated, quick teardown, PostgreSQL bundled |

The official Odoo documentation itself lists these three methods side-by-side
(<https://www.odoo.com/documentation/19.0/administration/on_premise.html>), reinforcing
that all three are first-class.

### Extension/Addon Architecture (Out of Scope, But Design for It)

The Odoo ecosystem has three layers of addons on top of the Community core:

1. **Core addons** — shipped with `odoo/odoo` repo (`addons/` directory)
2. **OCA modules** — Odoo Community Association's 38,000+ community modules
3. **Enterprise addons** — proprietary modules in `odoo/enterprise` repo
   (requires subscription)

The core `odoo.json` package definition must expose a mechanism (e.g. a
configurable `addons_path` and a well-defined `/mnt/extra-addons` mount for the
container variant) so follow-up packages can deposit additional addons into the
correct location without the core needing to know about them. Explicit extension
packages are out of scope for this issue but must not be blocked by the design.

### PostgreSQL Dependency — Critical Constraint

Odoo **only supports PostgreSQL** — no MySQL, SQLite, or other RDBMS. This is
architecturally fundamental (the ORM uses PostgreSQL-specific features: JSONB,
array types, advisory locks, ltree in some modules). The installer must:

1. For `default`/`source` variants: verify PostgreSQL is installed and reachable
   before proceeding; on failure, emit a clear message pointing to
   `portunix install postgresql`
2. For `container` variant: spin up a bundled `postgres:16` container unless the
   user explicitly points at an external instance

### Version Management

Odoo has annual major releases with overlapping support for the last 3 versions.
The package definition should support explicit version pinning per variant:

- `default`: APT repo subpath selects version (`nightly.odoo.com/19.0/` vs
  `nightly.odoo.com/18.0/`)
- `source`: git branch selects version (`--branch 19.0`, `--branch 18.0`, `master`)
- `container`: image tag selects version (`odoo:19`, `odoo:18`, `odoo:latest`)

Default to latest stable (`19` as of April 2026). User can override via
installer flags.

### Security Considerations (Surface in aiPrompts / Help)

- **Master password** (`admin_passwd`) must be changed from default on first use —
  the default config uses a placeholder that blocks until replaced
- Enable `proxy_mode = True` when fronting with a reverse proxy (Caddy, Nginx)
- Set `list_db = False` in production to hide database selector
- Use PostgreSQL roles with minimal privileges (not the superuser)
- Container variant should publish only `127.0.0.1:8069` by default, not `0.0.0.0`,
  unless explicitly overridden

### Windows / macOS Support Scope

- `default` (APT) variant is Linux-only by design
- `source` variant works on Windows/macOS for development, but Odoo explicitly
  does not recommend Windows for production (per manifest)
- `container` variant works anywhere Docker/Podman runs — this is the recommended
  path for Windows/macOS users who need Odoo

Package definition should reflect this: `default` limited to linux; `source` and
`container` available on all three platforms.

## Implementation Scope

### Files to Create

| File | Description |
| ---- | ----------- |
| `src/helpers/ptx-installer/assets/packages/odoo.json` | Odoo Community core package definition (all three variants) |

### Files to Modify

None required — registry auto-discovers packages from `assets/packages/`.

### Potentially Affected / Referenced

- `postgresql.json` — referenced as prerequisite (may need to verify/create if
  not yet present)
- `python.json` / `nodejs.json` — referenced as prerequisites for `source` variant
- `docs/issues/README.md` — add row for issue #172

## Acceptance Criteria

### `default` variant (APT package on Debian/Ubuntu)

1. `portunix install odoo` (on Ubuntu 22.04+/Debian 12+) adds Odoo APT repo, imports
   GPG key, and installs `odoo` package
2. `odoo.service` systemd unit is installed and enabled
3. `/etc/odoo/odoo.conf` exists with generated baseline config
4. `odoo --version` returns `Odoo Server 19.0` (or the installed version)
5. HTTP probe against `http://localhost:8069/web/database/list` returns JSON
6. PostgreSQL prerequisite is detected — if missing, install fails with clear
   error pointing to `portunix install postgresql`

### `source` variant (build from source)

7. `portunix install odoo --variant source` clones `odoo/odoo` at branch `19.0`
   into the configured target directory
8. Python venv created via `uv`, deps from `requirements.txt` installed
9. System-level build dependencies (libxml2-dev, libpq-dev, etc.) are installed
   on Linux automatically
10. `uv run python3 odoo-bin --version` returns expected version
11. Works cross-platform: Ubuntu, Fedora, Windows, macOS (validated in at least
    two platforms via containers or VMs)

### `container` variant (official Docker image)

12. `portunix install odoo --variant container` pulls `odoo:19` image
13. Bundled PostgreSQL container `portunix-odoo-db` is created and linked
14. Odoo container `portunix-odoo` starts with persistent volume for filestore
15. Port 8069 is accessible, HTTP probe against `/web/database/list` succeeds
16. Health check passes before returning success
17. Works with both Docker and Podman (via Portunix container abstraction)
18. `portunix container list` shows both `portunix-odoo` and `portunix-odoo-db`
19. Data persists across container restarts (named volumes survive)

### General

20. `odoo` appears in `portunix install --list` output
21. `portunix install odoo --help` documents all three variants clearly
22. Package definition passes registry validation (no auto-discovery errors)
23. `aiPrompts` section describes the three variants, when to use each, and
    common pitfalls (master password, PostgreSQL requirement, reverse proxy)
24. Package design allows future extension packages (OCA, Enterprise) to register
    additional addons paths without modifying core `odoo.json`

## Testing

All installation testing MUST be performed in containers per project methodology
(`docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` — Container-Based Testing
Policy).

### `default` variant (APT package)

```bash
portunix container run --image ubuntu:22.04
# inside container:
apt update && apt install -y curl wget gnupg postgresql
service postgresql start
sudo -u postgres createuser -s odoo
./portunix install odoo
systemctl status odoo || service odoo status
curl -s http://localhost:8069/web/database/list
odoo --version
```

### `source` variant (build from source)

```bash
portunix container run --image ubuntu:22.04
# inside container:
apt update && apt install -y git curl postgresql
service postgresql start
sudo -u postgres createuser -s odoo

./portunix install python          # prerequisite (uv from #171)
./portunix install nodejs          # prerequisite
./portunix install odoo --variant source

cd /opt/odoo     # or wherever source is installed
uv run python3 odoo-bin --version
# quick smoke test
sudo -u postgres createdb odoo-test
uv run python3 odoo-bin -d odoo-test --stop-after-init
```

### `container` variant (official Docker image)

```bash
# Host with Docker or Podman available
./portunix install odoo --variant container
# Expected: pulls odoo:19 + postgres:16, creates portunix-odoo-db and
# portunix-odoo, waits for health check

portunix container list
# Expected: portunix-odoo (running), portunix-odoo-db (running)

curl -s http://localhost:8069/web/database/list
# Expected: HTTP 200, JSON response

# Data persistence test
portunix container stop portunix-odoo
portunix container start portunix-odoo
# filestore and database should survive

# Cleanup
portunix container rm portunix-odoo
portunix container rm portunix-odoo-db
```

### Version pinning test

```bash
# Install older major version
./portunix install odoo --version 18 --variant container
docker exec portunix-odoo odoo --version
# Expected: Odoo Server 18.0
```

### Negative tests

```bash
# default variant on unsupported distro (e.g. Alpine)
portunix container run --image alpine:latest
./portunix install odoo
# Expected: clear error — APT-based install requires Debian/Ubuntu; suggest
# --variant source or --variant container

# Missing PostgreSQL
portunix container run --image ubuntu:22.04
./portunix install odoo
# Expected: clear error pointing to `portunix install postgresql`

# Container variant without container runtime
./portunix install odoo --variant container
# (on system without Docker/Podman) — clear error about missing container runtime
```

### Cross-platform (container-based validation)

| Variant | Ubuntu 22.04 | Debian 12 | Fedora 42 | Windows 11 | macOS |
| ------- | ------------ | --------- | --------- | ---------- | ----- |
| default | ✅ test | ✅ test | ❌ N/A | ❌ N/A | ❌ N/A |
| source | ✅ test | ✅ test | ✅ test | ✅ test | 🔹 best-effort |
| container | ✅ test | ✅ test | ✅ test | ✅ test | 🔹 best-effort |

## Reference

- Manifest: `portunix-architecture/docs/manifests/manifest-odoo.md`
- Odoo website: <https://www.odoo.com/>
- Source code: <https://github.com/odoo/odoo>
- Official Docker images: <https://github.com/odoo/docker-official-images>
- APT repository: <https://nightly.odoo.com/>
- Documentation: <https://www.odoo.com/documentation/19.0/>
- Installation guide: <https://www.odoo.com/documentation/19.0/administration/on_premise.html>
- Related issues:
  - #148 (MinIO) — template for multi-variant pattern (source + container)
  - #142 (Elasticsearch) — template for container service with health check
  - #171 (uv Python tooling) — `source` variant uses uv for venv management
  - #041 (Node.js) — prerequisite for `source` variant
  - PostgreSQL package — mandatory prerequisite (verify availability or create)

## Complexity

**High** — Combines three distinct installation patterns in a single package: (1) APT
repository management with GPG key import and systemd service setup, (2) multi-stage
source build with system deps + Python venv (via uv) + Node.js, and (3) container
orchestration with a bundled PostgreSQL sidecar, persistent volumes, and health checks.
Additional complexity comes from a mandatory non-trivial prerequisite (PostgreSQL) that
must be detected across all variants, cross-platform constraints (default variant is
Linux-only), security defaults that need surfacing (master password, reverse proxy,
list_db), and a design constraint to support future extension packages without rework.

---

## Addendum — 2026-04-17 (implementation scoping)

During implementation analysis it became clear that the original *Files to Modify:
None required* claim conflicts with the compose-style bundled-PostgreSQL container
variant (section "Container variant decision"). The current `ContainerSpec` schema
(`src/helpers/ptx-installer/registry/registry.go:128`) and the installer engine
(`src/helpers/ptx-installer/engine/container_service.go:82`) support exactly one
container per variant — no sidecars, no shared network, no startup ordering. The
main `portunix container run` (`src/cmd/container.go:511`) and its underlying
`docker.ContainerRunOptions` / `podman.ContainerRunOptions` likewise have no
`--network` passthrough.

To fulfil the original intent (self-contained container variant with bundled
PostgreSQL), the scope of this issue is expanded to cover the minimal infrastructure
needed. All additions are backwards-compatible (new fields are `omitempty`; existing
packages like MinIO, Elasticsearch are untouched).

### Expanded Files to Modify

| File | Change |
| ---- | ------ |
| `src/helpers/ptx-installer/registry/registry.go` | Add `Network string`, `Sidecars []ContainerSpec`, `DependsOn []string` to `ContainerSpec` |
| `src/helpers/ptx-installer/engine/container_service.go` | Create network if set; launch sidecars (await healthcheck each) before primary; pass `--network` to primary and sidecars |
| `src/app/docker/docker.go` | Add `Network string` to `ContainerRunOptions`; pass as `--network` |
| `src/app/podman/podman.go` | Add `Network string` to `ContainerRunOptions`; pass as `--network` |
| `src/cmd/container.go` | Add `--network` string flag to `containerRunCmd`; forward to docker/podman `RunContainer` |

### Network Creation Strategy

Network create/delete is performed by `container_service.go` via direct
`docker network create` / `podman network create` shell invocation (runtime
detected via `exec.LookPath` — podman first, docker fallback, matching
`ptx-container` convention). Idempotent: pre-existing network is ignored.
This avoids introducing a new `portunix container network` subcommand in the
scope of this issue; that can follow if a broader need emerges.

### Sidecar Orchestration

1. If primary `container.Network != ""`, ensure network exists (create if missing).
2. For each `container.Sidecars[i]`:
   - Launch via the same code path as the primary (re-use `installContainer`
     logic for name/ports/volumes/env/network).
   - If sidecar has `HealthCheck.Endpoint`, wait for it before proceeding.
   - If sidecar has only a `HealthCheck.Command` or neither, fall back to a
     short fixed delay (3s) so downstream containers see the service.
3. Launch primary with the same network.
4. Await primary healthcheck (existing logic).

Teardown is NOT added in this issue — `portunix container rm <name>` remains
manual for both primary and sidecars. A follow-up may introduce a coupling that
removes the network and sidecars when the primary is removed via a package-aware
uninstall command.

### Updated Acceptance Criteria Delta

- AC-12 through AC-19 (container variant) remain valid; the installer now truly
  starts BOTH `portunix-odoo` AND `portunix-odoo-db` on shared network
  `portunix-odoo-net` without requiring any pre-existing PostgreSQL.
- New AC-25: Other container-based packages (`minio`, `elasticsearch`) continue
  to install successfully — no regression from schema additions.

*Addendum added: 2026-04-17*

## Addendum — 2026-04-17 (external-db container variant)

The original "Container variant decision" section listed two approaches —
compose-style (bundled PostgreSQL sidecar) and external-db (user's own
PostgreSQL elsewhere). The first is implemented as the `container` variant.
This addendum adds the second.

### New variant `container-external-db`

Same as `container`, but **without** the bundled PostgreSQL sidecar. The
installer only creates the Odoo container and attaches it to the shared network
`portunix-odoo-net`. The user is responsible for:

- Running a PostgreSQL instance reachable from the Odoo container (on the same
  network, on the host via `host.containers.internal`, or a remote server)
- Providing correct connection parameters via the new `--db-*` install flags

### New install flags

Extends `ptx-installer` with four optional flags that override the container
environment for the installed package:

| Flag | Env key overridden | Default (from odoo.json) |
| ---- | ------------------ | ------------------------ |
| `--db-host=<host>` | `HOST` | `portunix-odoo-db` |
| `--db-port=<port>` | `PORT` | (unset) |
| `--db-user=<user>` | `USER` | `odoo` |
| `--db-password=<pass>` | `PASSWORD` | `odoo` |

The flags apply to any container-type variant, but are only meaningful for
variants that read these environment keys (currently: `container`,
`container-external-db`). Unspecified flags fall back to whatever the
variant's JSON defines.

### Expanded Files to Modify (additive)

| File | Change |
| ---- | ------ |
| `src/helpers/ptx-installer/engine/installer.go` | Add `DBHost`, `DBPort`, `DBUser`, `DBPassword` to `InstallOptions` |
| `src/helpers/ptx-installer/main.go` | Parse `--db-host` / `--db-port` / `--db-user` / `--db-password`; document in help |
| `src/helpers/ptx-installer/engine/container_service.go` | Override `container.Environment` HOST/PORT/USER/PASSWORD from `InstallOptions` before building run args |
| `src/helpers/ptx-installer/assets/packages/odoo.json` | Add `container-external-db` variant (no sidecars, same network); update `aiPrompts` |

### Usage examples

```bash
# PostgreSQL running in another container on portunix-odoo-net (default hostname)
portunix install odoo --variant=container-external-db

# PostgreSQL on host (accessible via host.containers.internal in rootless podman)
portunix install odoo --variant=container-external-db \
  --db-host=host.containers.internal \
  --db-user=odoo --db-password=secret

# PostgreSQL on a fully external server
portunix install odoo --variant=container-external-db \
  --db-host=db.example.com --db-port=5432 \
  --db-user=odoo --db-password=mypass
```

### New Acceptance Criteria

- AC-26: `portunix install odoo --variant=container-external-db` creates only
  `portunix-odoo` on network `portunix-odoo-net`. No `portunix-odoo-db`
  container or `odoo-db-data` volume is created.
- AC-27: `--db-host=<value>` overrides the container's `HOST` env var; verified
  via `podman inspect` output.
- AC-28: All four `--db-*` flags work independently and in combination. Omitted
  flags keep the JSON defaults.
- AC-29: Providing `--db-*` flags to the existing `container` variant does NOT
  break it — the flags override the defaults but the sidecar still starts.
  (Edge case: user opts to run the bundled sidecar with different credentials —
  supported as a side-effect of the generic override mechanism.)
- AC-30: Odoo container reports healthy via `/web/health` when a matching
  PostgreSQL is reachable at the specified host.

*Addendum added: 2026-04-17*

---

*Created: 2026-04-17*
