# Issue #112: PTX-PFT Category Management for UC and Requirements

**Status**: 📋 Open
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, helper-binary, ptx-pft, categorization, organization

---

## Summary

Implement category management system for ptx-pft that allows organizing Use Cases (UC) and Requirements (REQ) into categories within each area (VoC, VoS, VoB, VoE). Each area maintains its own independent category registry, and feedback items can be either uncategorized or assigned to exactly one category.

## Motivation

As product feedback grows, organizing UC and requirements into logical categories becomes essential for:
- Better navigation and discovery
- Grouping related items for analysis
- Filtering by topic/domain
- Reporting by category
- Integration with external tools (category mapping)

## Requirements

### Functional Requirements

1. **Category Registry per Area**
   - Each area (voc/, vos/, vob/, voe/) has independent category registry
   - Categories stored in `categories.json` within each area directory
   - Categories have unique ID within their area

2. **Category Properties**
   - `id`: Unique identifier (slug format, e.g., "user-auth", "data-export")
   - `name`: Display name (e.g., "User Authentication", "Data Export")
   - `description`: Optional description
   - `color`: Optional color for UI (hex format, e.g., "#3B82F6")
   - `order`: Optional sort order for display

3. **Item-Category Assignment**
   - Each UC/REQ can be assigned to 0..N categories (multi-category support)
   - Uncategorized items are valid (empty categories = "uncategorized")
   - Categories stored in markdown as `## Categories` section (comma-separated)
   - Assignment persisted in both local markdown and sync cache

4. **CLI Commands**
   ```bash
   # Category management
   portunix pft category list [--area voc|vos|vob|voe]
   portunix pft category add <id> --name "Name" [--description "..."] [--color "#hex"] --area voc
   portunix pft category remove <id> --area voc [--force]
   portunix pft category rename <id> --name "New Name" --area voc
   portunix pft category show <id> --area voc

   # Item categorization (0..N categories per item)
   portunix pft assign <item-id> --category <category-id>    # Add category to item
   portunix pft unassign <item-id> --category <category-id>  # Remove specific category
   portunix pft unassign <item-id> --all                     # Remove all categories

   # Filtering by category
   portunix pft list --category <category-id>
   portunix pft list --uncategorized
   portunix pft list --all-categories
   ```

5. **Validation Rules**
   - Cannot remove category with assigned items (unless --force)
   - Category ID must be unique within area
   - Category ID format: alphanumeric with hyphens (case-insensitive input, stored as UPPERCASE)

6. **Reporting Integration**
   - `pft report` includes category breakdown
   - `pft export` includes category field
   - Category statistics in area summary

### Technical Requirements

1. **Data Structures**
   ```go
   // Category represents a category within an area
   type Category struct {
       ID          string `json:"id"`
       Name        string `json:"name"`
       Description string `json:"description,omitempty"`
       Color       string `json:"color,omitempty"`
       Order       int    `json:"order,omitempty"`
       CreatedAt   string `json:"created_at"`
       UpdatedAt   string `json:"updated_at"`
   }

   // CategoryRegistry holds all categories for an area
   type CategoryRegistry struct {
       Version    string     `json:"version"`
       Area       string     `json:"area"`
       Categories []Category `json:"categories"`
       UpdatedAt  string     `json:"updated_at"`
   }
   ```

2. **FeedbackItem Extension**
   ```go
   type FeedbackItem struct {
       // ... existing fields ...
       Categories []string `json:"categories,omitempty"` // NEW: 0..N Category IDs
   }
   ```

3. **Markdown Format Extension**
   ```markdown
   # UC001: User Login

   ## Categories
   user-auth, data-security

   ## Summary
   As a user, I want to log into the system...

   ## Priority
   High

   ## Status
   Open

   ## Description
   Users need a secure authentication mechanism...
   ```

4. **File Structure**
   ```
   project-root/
   ├── voc/
   │   ├── categories.json          # VoC category registry
   │   ├── UC001-user-login.md      # category: user-auth
   │   ├── UC002-data-export.md     # category: data-export
   │   └── UC003-notifications.md   # uncategorized
   ├── vos/
   │   ├── categories.json          # VoS category registry (independent)
   │   ├── REQ001-gdpr-compliance.md
   │   └── REQ002-performance-sla.md
   └── .pft-config.json
   ```

5. **Provider Synchronization**
   - Map categories to provider-specific concepts:
     - Fider: Tags (category as primary tag)
     - ClearFlask: Categories/Tags
     - Eververse: Custom field
   - Category sync is one-way (local → remote) by default
   - Optional: Import categories from remote system

## Implementation Phases

### Phase 1: Category Data Model
- [ ] Define `Category` and `CategoryRegistry` structs
- [ ] Add `Category` field to `FeedbackItem`
- [ ] Implement category storage (`categories.json`)
- [ ] CRUD operations for categories

### Phase 2: CLI Commands
- [ ] `pft category list` - list categories in area
- [ ] `pft category add` - create new category
- [ ] `pft category remove` - delete category
- [ ] `pft category rename` - rename category
- [ ] `pft category show` - show category details

### Phase 3: Item Categorization
- [ ] `pft assign` - assign item to category
- [ ] `pft unassign` - remove item from category
- [ ] Update `ParseMarkdownFile()` to read category
- [ ] Update markdown writer to include category section

### Phase 4: List Filtering
- [ ] `pft list --category <id>` filter
- [ ] `pft list --uncategorized` filter
- [ ] Category column in list output

### Phase 5: Reporting Integration
- [ ] Category breakdown in `pft report`
- [ ] Category field in `pft export`
- [ ] Category statistics

### Phase 6: Provider Sync
- [ ] Map category to Fider tags
- [ ] Map category to ClearFlask
- [ ] Map category to Eververse

## Architecture Diagram

```
+------------------+     +------------------+     +------------------+
|      VoC         |     |      VoS         |     |      VoB         |
|  categories.json |     |  categories.json |     |  categories.json |
+--------+---------+     +--------+---------+     +--------+---------+
         |                        |                        |
         v                        v                        v
+------------------+     +------------------+     +------------------+
| UC001.md         |     | REQ001.md        |     | BIZ001.md        |
| category: auth   |     | category: gdpr   |     | category: sales  |
+------------------+     +------------------+     +------------------+
         |                        |                        |
         +-----------+------------+-----------+------------+
                     |                        |
                     v                        v
              +--------------+        +---------------+
              | pft category |        | pft assign    |
              | list/add/rm  |        | pft unassign  |
              +--------------+        +---------------+
                     |                        |
                     v                        v
              +----------------------------------------+
              |           FeedbackProvider             |
              |  (category → tags/custom fields)       |
              +----------------------------------------+
```

## Example Workflow

```bash
# 1. Create categories in VoC area
pft category add user-auth --name "User Authentication" --color "#3B82F6" --area voc
pft category add data-export --name "Data Export" --color "#10B981" --area voc
pft category add notifications --name "Notifications" --color "#F59E0B" --area voc

# 2. List categories
pft category list --area voc
# ID             NAME                  ITEMS
# user-auth      User Authentication   0
# data-export    Data Export           0
# notifications  Notifications         0

# 3. Assign items to categories
pft assign UC001-user-login --category user-auth
pft assign UC002-data-export --category data-export

# 4. List items by category
pft list --category user-auth
# ID                TITLE          STATUS    CATEGORY
# UC001-user-login  User Login     Open      user-auth

# 5. List uncategorized items
pft list --uncategorized
# ID                        TITLE                STATUS    CATEGORY
# UC003-notification-pref   Notification Prefs   Open      -

# 6. Report with category breakdown
pft report
# Category Summary:
# - user-auth: 3 items (2 open, 1 planned)
# - data-export: 2 items (1 open, 1 completed)
# - uncategorized: 5 items
```

## categories.json Format

```json
{
  "version": "1.0",
  "area": "voc",
  "updated_at": "2025-12-24T10:00:00Z",
  "categories": [
    {
      "id": "user-auth",
      "name": "User Authentication",
      "description": "Login, logout, password reset, 2FA, SSO",
      "color": "#3B82F6",
      "order": 1,
      "created_at": "2025-12-24T10:00:00Z",
      "updated_at": "2025-12-24T10:00:00Z"
    },
    {
      "id": "data-export",
      "name": "Data Export",
      "description": "CSV, PDF, Excel export functionality",
      "color": "#10B981",
      "order": 2,
      "created_at": "2025-12-24T10:00:00Z",
      "updated_at": "2025-12-24T10:00:00Z"
    }
  ]
}
```

## Success Criteria

- [ ] Each area has independent category registry
- [ ] Categories can be created, renamed, deleted via CLI
- [ ] Items can be assigned/unassigned to categories
- [ ] `pft list` supports category filtering
- [ ] `pft report` includes category statistics
- [ ] `pft export` includes category field
- [ ] Categories sync to external providers as tags/categories
- [ ] All commands work on Linux and Windows

## Dependencies

- **Issue #107**: PTX-PFT core implementation (completed)
- **Issue #108**: Email notifications (completed)
- **Issue #109**: ClearFlask provider (completed)
- **Issue #110**: Eververse provider (completed)

## Notes

- Categories are multi-select (0..N categories per item)
- Categories are area-specific to allow different organizational schemes
- Color is optional but recommended for UI integration

---

## Per-Area Provider Configuration (Extended Scope)

### Motivation

Different areas (VoC, VoS, VoB, VoE) may require different providers:
- **VoC** (Voice of Customer): External feedback tool like Fider
- **VoS** (Voice of Stakeholder): Separate Fider instance or ClearFlask
- **VoB** (Voice of Business): Local only (no external sync)
- **VoE** (Voice of Engineer): Local only (no external sync)

Note: Fider doesn't support multi-project, so separate VoC/VoS requires two instances.

### CLI Configuration Commands

```bash
# Global settings (no --area)
portunix pft configure --name "MyProduct" --path /tmp/pft-test

# Per-area provider configuration
portunix pft configure --area voc --provider fider --url http://localhost:3100 --token xxx
portunix pft configure --area vos --provider fider --url http://localhost:3101 --token yyy
portunix pft configure --area vob --provider local
portunix pft configure --area voe --provider local

# SMTP configuration
portunix pft configure --smtp-host smtp.example.com --smtp-port 587 \
    --smtp-user user --smtp-pass secret --smtp-from noreply@example.com

# Show configuration
portunix pft configure --show
```

### Supported Providers

| Provider | Multi-project | Use Case |
|----------|---------------|----------|
| fider | ❌ No | Simple feedback, one instance per area |
| clearflask | ✅ Yes | Multi-project via project_id |
| eververse | ✅ Yes | Multi-product via product_id |
| local | N/A | No external sync, markdown only |

### Updated Config Structure

```json
{
  "name": "MyProduct",
  "path": "/tmp/pft-test",
  "smtp": {
    "host": "smtp.example.com",
    "port": 587,
    "username": "user",
    "password": "secret",
    "from": "noreply@example.com"
  },
  "voc": {
    "provider": "fider",
    "url": "http://localhost:3100",
    "api_token": "xxx"
  },
  "vos": {
    "provider": "fider",
    "url": "http://localhost:3101",
    "api_token": "yyy"
  },
  "vob": {
    "provider": "local"
  },
  "voe": {
    "provider": "local"
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

### Updated Go Structures

```go
// AreaConfig holds configuration for a single area (voc, vos, vob, voe)
type AreaConfig struct {
    Provider  string `json:"provider,omitempty"`   // fider, clearflask, eververse, local
    URL       string `json:"url,omitempty"`        // Provider endpoint URL
    APIToken  string `json:"api_token,omitempty"`  // API token for authentication
    ProjectID string `json:"project_id,omitempty"` // For ClearFlask multi-project
    ProductID string `json:"product_id,omitempty"` // For Eververse multi-product
}

// Config represents the .pft-config.json structure
type Config struct {
    Name     string      `json:"name"`
    Path     string      `json:"path"`
    SMTP     *SMTPConfig `json:"smtp,omitempty"`
    VoC      *AreaConfig `json:"voc,omitempty"`
    VoS      *AreaConfig `json:"vos,omitempty"`
    VoB      *AreaConfig `json:"vob,omitempty"`
    VoE      *AreaConfig `json:"voe,omitempty"`
    Sync     SyncConfig  `json:"sync"`
    Mappings Mappings    `json:"mappings"`
}
```

### Implementation Phases (Configuration)

#### Phase 7: Per-Area Configuration
- [ ] Add `AreaConfig` struct
- [ ] Update `Config` struct with VoB, VoE fields
- [ ] Implement `--area` flag parsing
- [ ] Implement per-area provider/url/token setting
- [ ] Update `--show` to display per-area config

#### Phase 8: SMTP Configuration
- [ ] Implement `--smtp-host`, `--smtp-port`, etc. flags
- [ ] Update SMTP config in configure command
- [ ] Validate SMTP configuration

#### Phase 9: MCP Configuration Tools
- [x] Add `pft_config_show` tool - show current configuration
- [x] Add `pft_config_global` tool - set global settings (name, path)
- [x] Add `pft_config_area` tool - configure area provider
- [x] Add `pft_config_smtp` tool - configure SMTP server
- [x] Register tools in MCP handlers

#### Phase 10: MCP Architecture Fix (partial)
- [x] Update `ptx-mcp` to use MCP server from `src/app/mcp/`
- [x] Update `ptx-mcp` to properly handle `mcp serve` command
- [x] Ensure dispatcher routes `mcp` commands to `ptx-mcp` helper
- [ ] Verify MCP tools work via `portunix mcp serve`

#### Phase 11: Move ALL MCP Commands to ptx-mcp
Move all MCP commands from main portunix binary to ptx-mcp helper:

- [x] Move `mcp configure` command (`src/cmd/mcp_configure.go`)
- [x] Move `mcp reconfigure` command (`src/cmd/mcp_reconfigure.go`)
- [x] Move `mcp remove` command (`src/cmd/mcp_remove.go`)
- [x] Move `mcp status` command (`src/cmd/mcp_status.go`)
- [x] Move `mcp init` command (`src/cmd/mcp_server_init.go`)
- [x] Move `mcp start` command (`src/cmd/mcp_server_start.go`)
- [x] Move `mcp stop` command (`src/cmd/mcp_server_stop.go`)
- [x] Move `mcp config` command (`src/cmd/mcp_server_config.go`)
- [x] Move `mcp test` command (`src/cmd/mcp_server_test_cmd.go`)
- [x] Move `mcp serve status` command
- [x] Update ptx-mcp help to show all available commands
- [x] Remove obsolete mcp_*.go files from src/cmd/
- [x] Build and test all MCP commands

#### Phase 12: MCP reconfigure --force fix
- [x] Add `--force` flag to `mcp reconfigure` command
- [x] Allow reconfigure to work even when no existing configuration exists

**Rationale**: All MCP functionality should be in the `ptx-mcp` helper binary, consistent with the dispatcher architecture. The dispatcher routes `mcp` commands to `ptx-mcp`, so ALL MCP commands must be implemented in `ptx-mcp`, not the main portunix binary.

---

## MCP Configuration Tools (Extended Scope)

### Motivation

AI assistants using MCP need to be able to configure PTX-PFT without requiring manual CLI usage. This enables fully automated setup of feedback collection infrastructure.

### MCP Tools for Configuration

#### pft_config_show
Shows current PTX-PFT configuration.

```json
{
  "name": "pft_config_show",
  "description": "Show current PTX-PFT configuration including global settings, per-area providers, and SMTP",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

#### pft_config_global
Set global configuration (product name, document path).

```json
{
  "name": "pft_config_global",
  "description": "Set global PTX-PFT configuration (product name and document path)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Product name"
      },
      "path": {
        "type": "string",
        "description": "Path to document directory"
      }
    },
    "required": ["path"]
  }
}
```

#### pft_config_area
Configure provider for a specific area (VoC, VoS, VoB, VoE).

```json
{
  "name": "pft_config_area",
  "description": "Configure provider for a specific area (voc, vos, vob, voe)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "area": {
        "type": "string",
        "enum": ["voc", "vos", "vob", "voe"],
        "description": "Target area"
      },
      "provider": {
        "type": "string",
        "enum": ["fider", "clearflask", "eververse", "local"],
        "description": "Provider type"
      },
      "url": {
        "type": "string",
        "description": "Provider endpoint URL"
      },
      "token": {
        "type": "string",
        "description": "API token for authentication"
      },
      "project_id": {
        "type": "string",
        "description": "Project ID (for ClearFlask)"
      }
    },
    "required": ["area", "provider"]
  }
}
```

#### pft_config_smtp
Configure SMTP server for notifications.

```json
{
  "name": "pft_config_smtp",
  "description": "Configure SMTP server for email notifications",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": {
        "type": "string",
        "description": "SMTP server hostname"
      },
      "port": {
        "type": "integer",
        "description": "SMTP server port (default: 587)"
      },
      "username": {
        "type": "string",
        "description": "SMTP username"
      },
      "password": {
        "type": "string",
        "description": "SMTP password"
      },
      "from": {
        "type": "string",
        "description": "Sender email address"
      }
    },
    "required": ["host"]
  }
}
```

### Example MCP Workflow

```json
// 1. Set global config
{"method": "tools/call", "params": {"name": "pft_config_global", "arguments": {"name": "MyProduct", "path": "/tmp/pft-test"}}}

// 2. Configure VoC with Fider
{"method": "tools/call", "params": {"name": "pft_config_area", "arguments": {"area": "voc", "provider": "fider", "url": "http://localhost:3100"}}}

// 3. Configure VoS with Fider
{"method": "tools/call", "params": {"name": "pft_config_area", "arguments": {"area": "vos", "provider": "fider", "url": "http://localhost:3101"}}}

// 4. Set VoB/VoE as local
{"method": "tools/call", "params": {"name": "pft_config_area", "arguments": {"area": "vob", "provider": "local"}}}
{"method": "tools/call", "params": {"name": "pft_config_area", "arguments": {"area": "voe", "provider": "local"}}}

// 5. Configure SMTP
{"method": "tools/call", "params": {"name": "pft_config_smtp", "arguments": {"host": "smtp.example.com", "port": 587, "from": "noreply@example.com"}}}

// 6. Verify configuration
{"method": "tools/call", "params": {"name": "pft_config_show", "arguments": {}}}
```

---

**Created**: 2025-12-24
**Updated**: 2025-12-25
**Author**: Software Architect
