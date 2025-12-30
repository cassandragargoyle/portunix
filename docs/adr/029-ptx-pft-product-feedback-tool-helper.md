# ADR-029: PTX-PFT Product Feedback Tool Helper

**Status**: Proposed
**Date**: 2025-12-22
**Architect**: Kurc

## Context

### Current Situation

Portunix currently lacks integration with product feedback management systems. Development teams need to:
- Collect user feedback from multiple sources
- Synchronize feedback between local project tracking and external feedback platforms
- Process and prioritize feature requests
- Track feedback-to-implementation lifecycle

### Problem Statement

There is no standardized way to:
1. Connect Portunix-managed projects with external product feedback tools
2. Synchronize feedback bidirectionally (local ↔ external system)
3. Deploy and manage feedback tool infrastructure
4. Integrate feedback data into development workflow

### External System Selection

**Fider.io** selected as primary supported platform:
- Open-source product feedback tool
- Self-hosted deployment option
- REST API for integration
- Docker-based deployment
- Active community support

## Decision

We will create a new helper binary `ptx-pft` (Product Feedback Tool) for managing product feedback integration and synchronization.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    portunix (Main Dispatcher)               │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Dispatcher delegates to:
                        │
        ┌───────────────┼────────────────┬───────────────────┐
        │               │                │                   │
        ▼               ▼                ▼                   ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ ptx-container│ │ ptx-installer│ │   ptx-pft    │ │     ...      │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
       │                                   │
       │                                   │
       └──────────────┬────────────────────┘
                      │
                      ▼
        ┌─────────────────────────────┐
        │   Fider.io (Docker)         │
        │   - PostgreSQL database     │
        │   - Fider application       │
        └─────────────────────────────┘
```

### PTX-PFT Responsibilities

```
┌──────────────────────────────────────────────────────────────────────┐
│                         PTX-PFT Helper                               │
│                                                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐       │
│  │  Synchronization │  │  Configuration  │  │  User Registry  │       │
│  │                  │  │                  │  │                 │       │
│  │  - Pull feedback │  │  - API keys     │  │  - VoC users    │       │
│  │  - Push updates  │  │  - Endpoints    │  │  - VoS users    │       │
│  │  - Conflict res. │  │  - Mappings     │  │  - Role mgmt    │       │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘       │
│                                                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐       │
│  │  Local Storage   │  │  Reporting      │  │  External IDs   │       │
│  │                  │  │                  │  │                 │       │
│  │  - Feedback cache│  │  - Status       │  │  - Fider ID map │       │
│  │  - Issue mapping │  │  - Statistics   │  │  - Email→User   │       │
│  │  - History       │  │  - Export       │  │  - Auto-sync    │       │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘       │
└──────────────────────────────────────────────────────────────────────┘
```

### Command Structure

```bash
# Infrastructure management (via ptx-container)
portunix pft deploy              # Deploy Fider.io to Docker
portunix pft status              # Check Fider.io status
portunix pft destroy             # Remove Fider.io instance

# Synchronization
portunix pft sync                # Full bidirectional sync
portunix pft pull                # Pull feedback from external
portunix pft push                # Push local changes to external

# Configuration
portunix pft configure           # Interactive configuration
portunix pft configure --url     # Set Fider.io URL
portunix pft configure --token   # Set API token

# Feedback management
portunix pft list                # List all feedback items
portunix pft show <id>           # Show feedback details
portunix pft link <id> <issue>   # Link feedback to local issue

# Reporting
portunix pft report              # Generate feedback report
portunix pft export --format=md  # Export to markdown

# User/Customer Registry
portunix pft user list --voc     # List VoC users
portunix pft user list --vos     # List VoS users
portunix pft user add --voc --id "email" --name "Name" --role customer
portunix pft user update <id> --role <role>
portunix pft user link <id> --fider <fider-id>
portunix pft user remove <id> --voc

# Role Management
portunix pft role list --voc     # Show VoC roles (customer, proxy-customer)
portunix pft role list --vos     # Show VoS roles (cio, developer, dev-lead, ...)
```

### PTX-Installer Integration

PTX-Installer will be extended to support Fider.io deployment:

```bash
# New package definition in assets/packages/fider.json
portunix install fider           # Install Fider.io via Docker
```

**Fider.io Installation Flow:**

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  ptx-installer   │────▶│  ptx-container   │────▶│  Docker Engine   │
│                  │     │                  │     │                  │
│  install fider   │     │  - Pull images   │     │  - fider:latest  │
│                  │     │  - Create network│     │  - postgres:15   │
│                  │     │  - Start compose │     │                  │
└──────────────────┘     └──────────────────┘     └──────────────────┘
```

**Package Definition (assets/packages/fider.json):**

```json
{
  "name": "fider",
  "description": "Open-source product feedback tool",
  "category": "development-tools",
  "installer": {
    "type": "container",
    "provider": "ptx-container",
    "compose": {
      "services": {
        "db": {
          "image": "postgres:15",
          "environment": {
            "POSTGRES_DB": "fider",
            "POSTGRES_USER": "fider",
            "POSTGRES_PASSWORD": "${FIDER_DB_PASSWORD}"
          },
          "volumes": ["fider-db:/var/lib/postgresql/data"]
        },
        "fider": {
          "image": "getfider/fider:latest",
          "ports": ["3000:3000"],
          "environment": {
            "DATABASE_URL": "postgres://fider:${FIDER_DB_PASSWORD}@db:5432/fider?sslmode=disable"
          },
          "depends_on": ["db"]
        }
      }
    }
  }
}
```

### Synchronization Architecture

**Bidirectional Sync Model:**

```
┌─────────────────────────────────────────────────────────────┐
│                     Local Project                           │
│                                                             │
│  docs/issues/           docs/feedback/                      │
│  ├── 001-feature.md     ├── .pft-config.yaml               │
│  ├── 002-bugfix.md      ├── .pft-cache.json                │
│  └── README.md          └── mappings.yaml                   │
│                                                             │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              │ ptx-pft sync
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Fider.io API                            │
│                                                             │
│  /api/v1/posts          (feedback items)                    │
│  /api/v1/tags           (categories)                        │
│  /api/v1/users          (voters)                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Sync Operations:**

| Direction | Operation | Description |
|-----------|-----------|-------------|
| Pull | `pft pull` | Download new feedback from Fider.io |
| Push | `pft push` | Upload local status changes to Fider.io |
| Sync | `pft sync` | Full bidirectional synchronization |

**Conflict Resolution:**
- Timestamp-based: Latest change wins
- Manual: Prompt user for resolution
- Priority: External system takes precedence (configurable)

### User/Customer Registry Architecture

**ISO 16355 Alignment:**
Per ISO 16355 (Quality Function Deployment), feedback sources are categorized as:
- **VoC (Voice of Customer)**: End-user feedback, feature requests
- **VoS (Voice of Stakeholder)**: Internal stakeholder requirements
- **VoB (Voice of Business)**: Business requirements, market opportunities
- **VoE (Voice of Engineer)**: Technical constraints, implementation feedback

**Storage Structure:**

```
project-root/
├── users.json               # Unified user registry (all categories)
├── voc/
│   ├── roles.json           # VoC role definitions
│   └── UC001-feature.md
├── vos/
│   ├── roles.json           # VoS role definitions
│   └── REQ001-requirement.md
├── vob/
│   ├── roles.json           # VoB role definitions
│   └── BUS001-opportunity.md
└── voe/
    ├── roles.json           # VoE role definitions
    └── TECH001-constraint.md
```

**User Schema (users.json):**

```json
{
  "users": [
    {
      "id": "user@example.com",
      "name": "John Doe",
      "external_ids": {
        "fider": 42
      },
      "organization": "Acme Corp",
      "roles": {
        "voc": { "role": "customer", "proxy": false },
        "vos": { "role": "cio", "proxy": true },
        "vob": null,
        "voe": null
      },
      "created_at": "2025-12-22T10:00:00Z",
      "updated_at": "2025-12-22T10:00:00Z"
    }
  ]
}
```

**Role Assignment Rules:**
- User can have ONE role per category (VoC, VoS, VoB, VoE)
- `proxy: true` means user represents the role but is not actually in that position
  - Example: Assistant who speaks on behalf of CIO → `{"role": "cio", "proxy": true}`
- `proxy: false` means user holds the actual position
- User should be assigned to highest applicable role in each category

**Role Definitions by Category:**

```json
// voc/roles.json - Customer-facing roles
{
  "type": "voc",
  "roles": [
    {
      "id": "customer",
      "name": "Customer",
      "description": "Direct end-user providing feedback"
    },
    {
      "id": "proxy-customer",
      "name": "Proxy Customer",
      "description": "Representative speaking on behalf of customers"
    },
    {
      "id": "customer-admin",
      "name": "Customer Admin",
      "description": "Administrator at customer organization"
    },
    {
      "id": "customer-support",
      "name": "Customer Support",
      "description": "Support staff at customer organization"
    }
  ]
}

// vos/roles.json - Stakeholder roles
{
  "type": "vos",
  "roles": [
    {
      "id": "cio",
      "name": "CIO",
      "description": "Chief Information Officer"
    },
    {
      "id": "developer",
      "name": "Developer",
      "description": "Software developer"
    },
    {
      "id": "dev-lead",
      "name": "Dev Lead",
      "description": "Development team lead"
    },
    {
      "id": "tech-consultant",
      "name": "Technical Consultant",
      "description": "Technical consultant"
    },
    {
      "id": "ceo",
      "name": "CEO",
      "description": "Chief Executive Officer / Company owner"
    },
    {
      "id": "product-manager",
      "name": "Product Manager",
      "description": "Product management"
    },
    {
      "id": "architect",
      "name": "Architect",
      "description": "Software/Solution architect"
    },
    {
      "id": "facilitator",
      "name": "Facilitator",
      "description": "Person responsible for gathering, moderating and documenting requirements"
    },
    {
      "id": "tester",
      "name": "Tester",
      "description": "Software tester providing stakeholder perspective"
    },
    {
      "id": "support",
      "name": "Support",
      "description": "Support staff at software vendor"
    },
    {
      "id": "support-lead",
      "name": "Support Lead",
      "description": "Support team lead at software vendor"
    }
  ]
}

// vob/roles.json - Business roles
{
  "type": "vob",
  "roles": [
    {
      "id": "ceo",
      "name": "CEO",
      "description": "Chief Executive Officer / Company owner"
    },
    {
      "id": "dev-lead",
      "name": "Dev Lead",
      "description": "Development team lead"
    },
    {
      "id": "sales",
      "name": "Sales",
      "description": "Sales representative"
    },
    {
      "id": "marketing",
      "name": "Marketing",
      "description": "Marketing specialist"
    },
    {
      "id": "support",
      "name": "Support",
      "description": "Support staff at software vendor"
    }
  ]
}

// voe/roles.json - Engineering roles
{
  "type": "voe",
  "roles": [
    {
      "id": "architect",
      "name": "Architect",
      "description": "Software/Solution architect"
    },
    {
      "id": "senior-developer",
      "name": "Senior Developer",
      "description": "Senior software developer"
    },
    {
      "id": "developer",
      "name": "Developer",
      "description": "Software developer"
    },
    {
      "id": "devops",
      "name": "DevOps",
      "description": "DevOps engineer"
    },
    {
      "id": "qa",
      "name": "QA",
      "description": "Quality assurance engineer"
    },
    {
      "id": "support",
      "name": "Support",
      "description": "Support staff at software vendor"
    },
    {
      "id": "tester",
      "name": "Tester",
      "description": "Software tester"
    }
  ]
}
```

**User Registry Commands:**

```bash
# List users
portunix pft user list                  # List all users
portunix pft user list --voc            # List users with VoC roles
portunix pft user list --vos            # List users with VoS roles

# Add user
portunix pft user add \
  --id "user@example.com" \
  --name "John Doe" \
  --org "Acme Corp"

# Assign role to user (one role per category)
portunix pft user role user@example.com --voc customer
portunix pft user role user@example.com --vos cio --proxy    # Proxy CIO
portunix pft user role user@example.com --voe developer

# Remove role from category
portunix pft user role user@example.com --vos --remove

# Link to external system
portunix pft user link user@example.com --fider 42

# Remove user
portunix pft user remove user@example.com

# Show roles
portunix pft role list --voc            # Show available VoC roles
portunix pft role list --vos            # Show available VoS roles
portunix pft role list --vob            # Show available VoB roles
portunix pft role list --voe            # Show available VoE roles
```

**User-Feedback Association:**

Each feedback item can reference its author in metadata:

```markdown
## Metadata
- Fider ID: 42
- Author: user@example.com
- Author Role: customer (VoC)
- Proxy: false
- Synced: 2025-12-22
```

**External ID Synchronization:**

When syncing with Fider, user mappings are maintained:
1. Pull creates/updates user record with Fider ID
2. Push uses local user registry to attribute feedback
3. User registry can be pre-populated before first sync

### Configuration Storage

```yaml
# .pft-config.yaml (project root)
provider: fider
endpoint: https://feedback.example.com
api_token: ${PFT_API_TOKEN}  # Environment variable reference

sync:
  auto: false
  interval: 1h
  conflict_resolution: timestamp

mappings:
  status:
    open: pending
    planned: in_progress
    completed: implemented
    declined: rejected
```

## Trade-off Analysis

### Option A: Standalone Tool (No Integration)
**Pros:**
- Simpler implementation
- No dependencies on Portunix

**Cons:**
- No unified workflow
- Manual deployment
- No synchronization

### Option B: PTX-PFT Helper (Proposed)
**Pros:**
- Integrated with Portunix ecosystem
- Automated deployment via ptx-installer
- Bidirectional synchronization
- Consistent with helper pattern (ADR-014)

**Cons:**
- Additional helper binary
- Dependency on ptx-container for deployment
- Initial development effort

### Option C: Plugin Architecture
**Pros:**
- Extensible to multiple providers
- Plugin isolation

**Cons:**
- Overhead for single provider use case
- Complex plugin lifecycle
- Delayed implementation

**Decision**: **Option B** provides best balance of integration and simplicity.

## Implementation Phases

### Phase 1: Helper Foundation
- [ ] Create `src/helpers/ptx-pft/` directory structure
- [ ] Implement CLI with Cobra
- [ ] Add dispatcher routing in main binary
- [ ] Basic configuration management

### Phase 2: PTX-Installer Fider.io Support
- [ ] Create `assets/packages/fider.json` package definition
- [ ] Implement container-based installation in ptx-installer
- [ ] Integration with ptx-container for Docker Compose
- [ ] Installation verification

### Phase 3: Synchronization Engine
- [ ] Implement Fider.io API client
- [ ] Pull operation (external → local)
- [ ] Push operation (local → external)
- [ ] Conflict detection and resolution

### Phase 4: Advanced Features
- [ ] Issue linking (feedback ↔ local issues)
- [ ] Reporting and export
- [ ] Auto-sync scheduling
- [ ] Status webhooks

### Phase 5: User/Customer Registry
- [ ] JSON-based user storage (separate files for VoC/VoS)
- [ ] Role management with predefined role types
- [ ] External ID mapping (Fider ID, email)
- [ ] User CRUD operations via CLI

## Consequences

### Positive Consequences

1. **Unified Workflow**: Feedback management integrated into development process
2. **Automation**: Automated deployment and synchronization
3. **Consistency**: Follows established helper pattern
4. **Flexibility**: Extensible to other feedback tools in future

### Negative Consequences

1. **Complexity**: Additional helper binary to maintain
2. **Dependencies**: Requires ptx-container for deployment
3. **Learning Curve**: New commands and configuration

### Risk Mitigation

**Risk**: Fider.io API changes break integration
**Mitigation**: Version-specific API client, automated API compatibility testing

**Risk**: Sync conflicts cause data loss
**Mitigation**: Local backup before sync, manual conflict resolution option

## Related Decisions

- **ADR-014**: Git-like Dispatcher with Python Distribution Model
- **ADR-025**: PTX-Installer Helper Architecture
- **ADR-021**: Package Registry Architecture

## Success Criteria

- [ ] `portunix install fider` deploys working Fider.io instance
- [ ] `portunix pft sync` successfully synchronizes feedback
- [ ] No data loss during synchronization
- [ ] Cross-platform compatibility (Linux, Windows via WSL)

---

## Review and Approval

**Status**: Awaiting Product Owner approval
**Architect**: Kurc
**Date**: 2025-12-22
