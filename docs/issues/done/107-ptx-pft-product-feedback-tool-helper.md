# Issue #107: PTX-PFT Product Feedback Tool Helper Implementation

**Status**: ✅ Implemented (pending integration tests)
**Priority**: High
**Type**: Feature
**Labels**: enhancement, helper-binary, product-feedback, fider, synchronization

---

## ⚠️ CRITICAL: Container Usage Rules

**NEVER use `docker` or `podman` directly in ptx-pft code!**

All container operations MUST go through `portunix container` commands:

- `portunix container compose up -d`
- `portunix container compose down`
- `portunix container compose ps`
- etc.

The `ptx-container` helper automatically detects whether Docker or Podman is available and uses the correct runtime. Direct docker/podman calls bypass this abstraction and break cross-platform compatibility.

**In Go code**: Use `exec.Command("portunix", "container", "compose", ...)`

**In user instructions**: Always say `portunix container compose`, NEVER `docker compose`

### Bug Fix Required: ptx-container Runtime Detection

The `detectComposeRuntime()` function in `ptx-container` incorrectly detects Docker even when daemon is not running:

```go
// WRONG: docker compose version succeeds even without running daemon
if cmd := exec.Command("docker", "compose", "version", "--short"); cmd.Run() == nil {

// CORRECT: docker info fails if daemon is not running
if cmd := exec.Command("docker", "info"); cmd.Run() == nil {
```

This causes ptx-pft to fail on systems with Podman when Docker CLI is installed but daemon is not running.

---

## Summary

Implement a new helper binary `ptx-pft` for managing integration with external Product Feedback Tools, primarily Fider.io. The helper will provide bidirectional synchronization between local project documentation (markdown files) and external feedback systems.

**Standards Compliance**: Tool aims to be compliant with **ISO 16355** (Quality Function Deployment - QFD). See [docs/ISO-16355-QFD.md](../../ISO-16355-QFD.md) for details.

### VoC/VoS Separation (CRITICAL)

The tool MUST support separate handling of:

- **VoC (Voice of Customer)** - public customer feedback, visible to users
- **VoS (Voice of Stakeholder)** - internal requirements (regulatory, business, technical), NOT visible to customers

Each voice type has its own:

- Local directory (`voc/`, `vos/`)
- Feedback tool instance (separate Fider deployments)
- Synchronization configuration

## Use Case Example

**Product**: "Example AI Assistant MVP"
**Local path**: `/path/to/project/docs/strategy`

The workflow:

1. Configure product with `pft configure --name "Example AI Assistant MVP" --path /path/to/docs`
2. Deploy Fider.io instance for collecting user feedback
3. Synchronize between local markdown documents and Fider.io posts (bidirectional)
4. Use AI assistants to process feedback and generate product documentation

The synchronization strategy (individual use cases vs. aggregated documents) will be refined during implementation.

## Related Architecture Decision

- **ADR-029**: PTX-PFT Product Feedback Tool Helper

## Requirements

### Functional Requirements

1. **Quick Example / Demo**
   - Single command to demonstrate full workflow
   - Creates demo directory structure:
     ```
     pft-demo/
     ├── voc/                          # Voice of Customer (public)
     │   ├── UC001-user-login.md
     │   ├── UC002-data-export.md
     │   └── UC003-notifications.md
     ├── vos/                          # Voice of Stakeholder (internal)
     │   ├── REQ001-gdpr-compliance.md
     │   ├── REQ002-performance-sla.md
     │   └── REQ003-security-audit.md
     └── .pft-config.json
     ```
   - Deploys 2x Fider instances:
     - **VoC Fider** on port 3000 (public customer feedback)
     - **VoS Fider** on port 3001 (internal stakeholder requirements)
   - Configures ptx-pft automatically with VoC/VoS separation
   - Pushes sample items to respective Fider instances

2. **Infrastructure Management**
   - Deploy feedback tool via ptx-container
   - Check feedback tool status
   - Destroy/cleanup feedback tool instance

3. **Synchronization**
   - Pull feedback from external system to local
   - Push local changes to external system
   - Full bidirectional sync with conflict resolution

4. **Configuration**
   - Interactive configuration wizard
   - API endpoint and token management
   - Mapping between local and external statuses

5. **Feedback Management**
   - List all feedback items
   - Show feedback details
   - Link feedback to local issues

6. **Reporting**
   - Generate feedback reports
   - Export to markdown/other formats

### Technical Requirements

1. **Helper Binary Structure**
   - Location: `src/helpers/ptx-pft/`
   - CLI framework: Cobra
   - Dispatcher integration in main binary

2. **Provider Abstraction (CRITICAL)**
   - Implementation MUST be independent of the specific feedback tool (Fider.io)
   - Use provider/adapter pattern for external systems
   - Common interface for all providers (sync, pull, push, list, etc.)
   - Easy to add new providers (Canny, ProductBoard, etc.)
   - Support for multiple providers simultaneously
   - Shared synchronization logic across all providers

3. **PTX-Container Integration (CRITICAL)**
   - MUST use `portunix container compose` for all container operations
   - NEVER call docker/podman directly - ptx-container handles runtime detection
   - Create `assets/packages/fider.json` package definition
   - Container-based deployment via ptx-container compose commands

4. **Configuration Storage**
   - Project-level: `.pft-config.json`
   - Cache: `.pft-cache.json`

## Command Structure

```bash
# Quick Start / Demo
portunix pft example             # Full example: configure + deploy + sample data
portunix pft example --path /tmp/pft-demo  # Custom path for demo

# Infrastructure
portunix pft deploy              # Deploy feedback tool to container
portunix pft status              # Check feedback tool status
portunix pft destroy             # Remove feedback tool instance

# Synchronization
portunix pft sync                # Full bidirectional sync
portunix pft pull                # Pull from external
portunix pft push                # Push to external

# Configuration
portunix pft configure           # Interactive configuration
portunix pft configure --name    # Set product name
portunix pft configure --path    # Set path to local documents
portunix pft configure --url     # Set Fider.io URL
portunix pft configure --token   # Set API token

# Feedback management
portunix pft list                # List all feedback items
portunix pft show <id>           # Show feedback details
portunix pft link <id> <issue>   # Link feedback to local issue

# Reporting
portunix pft report              # Generate feedback report
portunix pft export --format=md  # Export to markdown
```

## Implementation Phases

### Phase 1: Helper Foundation ✅

- [x] Create `src/helpers/ptx-pft/` directory structure
- [x] Implement CLI skeleton with Cobra
- [x] Add dispatcher routing in main binary (`pft` command)
- [x] Implement `pft --version` and `pft --help`
- [x] Basic configuration management (read/write .pft-config.json)
- [x] Define `FeedbackProvider` interface (provider abstraction)

### Phase 2: Fider.io Installer Integration ✅

- [x] Create `assets/packages/fider.json` package definition
- [x] Implement container-based installation in ptx-installer
- [x] Docker Compose template for Fider.io + PostgreSQL
- [x] Installation verification (`pft status`)
- [x] Destruction command (`pft destroy`)

### Phase 3: Fider.io API Client ✅

- [x] Implement Fider.io REST API client
- [x] Authentication with API token
- [x] CRUD operations for posts (feedback items)
- [x] Tag management
- [x] User/voter information

### Phase 4: Synchronization Engine ✅

- [x] Pull operation (external → local)
- [x] Push operation (local → external)
- [x] Full sync with conflict detection
- [x] Conflict resolution strategies (timestamp, manual, priority)
- [x] Local cache management

### Phase 5: User/Customer Registry ✅

- [x] User registry with 4 categories (customer, partner, employee, prospect)
- [x] Role management (user, admin, moderator, observer)
- [x] Proxy attribute for representing others
- [x] User sync with Fider
- [x] User update functionality

### Phase 6: Issue Linking & Reporting ✅

- [x] Link feedback to local issues (docs/issues/)
- [x] Status mapping configuration
- [x] Report generation (summary, detailed, status)
- [x] Export functionality (markdown, JSON, CSV)

### Phase 7: Email Notifications ✅

- [x] Template-based email system
- [x] Email-only mode support
- [x] SMTP configuration
- [x] Notification types: vote, description, acceptance

### Phase 8: Unit Tests ✅

- [x] Unit tests for API client (8 tests)
- [x] Unit tests for sync engine (9 tests)
- [x] Unit tests for cache management (14 tests)
- [x] Unit tests for conflict detection (14 tests)

### Remaining Tasks

- [ ] Integration tests with mock Fider.io
- [ ] Container deployment tests
- [ ] Cross-platform tests (Linux, Windows via WSL)

## Fider.io Package Definition

```json
{
  "name": "fider",
  "description": "Open-source product feedback tool",
  "category": "development-tools",
  "homepage": "https://fider.io",
  "installer": {
    "type": "container",
    "provider": "ptx-container",
    "compose": {
      "version": "3.8",
      "services": {
        "db": {
          "image": "postgres:15",
          "environment": {
            "POSTGRES_DB": "fider",
            "POSTGRES_USER": "fider",
            "POSTGRES_PASSWORD": "${FIDER_DB_PASSWORD}"
          },
          "volumes": ["fider-db:/var/lib/postgresql/data"],
          "restart": "unless-stopped"
        },
        "fider": {
          "image": "getfider/fider:latest",
          "ports": ["3000:3000"],
          "environment": {
            "DATABASE_URL": "postgres://fider:${FIDER_DB_PASSWORD}@db:5432/fider?sslmode=disable",
            "JWT_SECRET": "${FIDER_JWT_SECRET}",
            "EMAIL_NOREPLY": "noreply@example.com"
          },
          "depends_on": ["db"],
          "restart": "unless-stopped"
        }
      },
      "volumes": {
        "fider-db": {}
      }
    }
  }
}
```

## Configuration File Format

```json
{
  "name": "Example AI Assistant MVP",
  "voc": {
    "path": "./voc",
    "provider": "fider",
    "endpoint": "http://localhost:3000",
    "api_token": "${PFT_VOC_TOKEN}",
    "visibility": "public"
  },
  "vos": {
    "path": "./vos",
    "provider": "fider",
    "endpoint": "http://localhost:3001",
    "api_token": "${PFT_VOS_TOKEN}",
    "visibility": "internal"
  },
  "sync": {
    "auto": false,
    "interval": "1h",
    "conflict_resolution": "timestamp"
  },
  "mappings": {
    "status": {
      "open": "pending",
      "planned": "in_progress",
      "started": "in_progress",
      "completed": "implemented",
      "declined": "rejected"
    }
  }
}
```

## Testing Requirements

- [ ] Unit tests for API client
- [ ] Unit tests for sync engine
- [ ] Integration tests with mock Fider.io
- [ ] Container deployment tests
- [ ] Cross-platform tests (Linux, Windows via WSL)

## Success Criteria

- [ ] `portunix install fider` successfully deploys Fider.io
- [ ] `portunix pft configure` creates valid configuration
- [ ] `portunix pft sync` synchronizes without data loss
- [ ] `portunix pft link` correctly maps feedback to issues
- [ ] All commands work on Linux and Windows (WSL)

## Dependencies

- **ptx-container**: For Docker/Podman operations
- **ptx-installer**: For package installation framework
- **Fider.io**: External product feedback tool

## Architecture Notes

### Provider Interface

```go
type FeedbackProvider interface {
    Name() string
    Connect(config ProviderConfig) error
    List() ([]FeedbackItem, error)
    Get(id string) (*FeedbackItem, error)
    Create(item FeedbackItem) error
    Update(item FeedbackItem) error
    Delete(id string) error
    Close() error
}
```

All synchronization logic works with this interface, not with specific provider implementations.

## Notes

- Fider.io requires PostgreSQL database
- API token obtained from Fider.io admin settings
- **Provider abstraction is critical** - sync code must be reusable across providers
- Future providers: Canny, ProductBoard, GitHub Discussions, custom REST API

---

**Created**: 2025-12-22
**Last Updated**: 2025-12-23

## Implementation History

- 2025-12-22: Initial issue creation, ISO 16355 QFD compliance, VoC/VoS separation
- 2025-12-23: Completed Phases 1-8
  - FeedbackProvider interface with Fider and Email providers
  - Full command set: list, show, link, report, export, cache
  - Conflict detection with resolution strategies
  - Local cache management (.pft-cache.json)
  - User registry with 4 categories and role management
  - Template-based email notifications
  - Unit tests: 57 tests (API client, sync engine, cache, conflicts)
