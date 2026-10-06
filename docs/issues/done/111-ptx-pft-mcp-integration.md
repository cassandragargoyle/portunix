# Issue #111: PTX-PFT MCP Integration

**Status**: ✅ Implemented
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, helper-binary, ptx-pft, ptx-mcp, ai-integration, mcp
**Closed**: 2026-06-30

---

## Summary

Integrate Product Feedback Tool (ptx-pft) functionality into the MCP Server (ptx-mcp) to enable AI assistants (Claude Code, Cursor, etc.) to interact with product feedback systems through the Model Context Protocol.

## Motivation

Currently:
- **ptx-pft** provides comprehensive product feedback management (sync, deploy, user management, notifications, etc.)
- **ptx-mcp** is a stub with TODO placeholders for MCP functionality

By integrating ptx-pft into ptx-mcp, AI assistants will be able to:
- Create, read, update feedback items through natural language
- Synchronize local documentation with external feedback tools
- Manage user feedback workflows
- Generate reports and analyze product feedback trends

## Architecture Decision

### Option A: MCP Tools in ptx-mcp calling ptx-pft (Recommended)
```
AI Assistant → MCP Protocol → ptx-mcp → exec(ptx-pft) → Fider/ClearFlask/Eververse
```

**Pros:**
- Clean separation of concerns
- ptx-pft remains usable standalone
- Easier testing and maintenance
- MCP tools are thin wrappers

**Cons:**
- Process spawn overhead

### Option B: Shared library between ptx-pft and ptx-mcp
```
AI Assistant → MCP Protocol → ptx-mcp → pftlib → Fider/ClearFlask/Eververse
```

**Pros:**
- No process spawn overhead
- Direct function calls

**Cons:**
- Requires refactoring ptx-pft into library
- Duplicate code or complex build dependencies
- Higher coupling

### Decision
**Option A** - MCP tools in ptx-mcp will invoke ptx-pft commands via exec. This maintains modularity and allows ptx-pft to evolve independently.

## MCP Tool Specifications

### Core Tools

```json
{
  "tools": [
    {
      "name": "pft_list",
      "description": "List all feedback items from product feedback system",
      "inputSchema": {
        "type": "object",
        "properties": {
          "voice": {
            "type": "string",
            "enum": ["voc", "vos", "all"],
            "description": "Voice type: voc (customer), vos (stakeholder), or all"
          },
          "status": {
            "type": "string",
            "description": "Filter by status (open, planned, started, completed, declined)"
          },
          "limit": {
            "type": "integer",
            "description": "Maximum number of items to return"
          }
        }
      }
    },
    {
      "name": "pft_show",
      "description": "Show details of a specific feedback item",
      "inputSchema": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "Feedback item ID"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "pft_sync",
      "description": "Synchronize local documentation with external feedback system",
      "inputSchema": {
        "type": "object",
        "properties": {
          "direction": {
            "type": "string",
            "enum": ["pull", "push", "bidirectional"],
            "description": "Sync direction"
          },
          "voice": {
            "type": "string",
            "enum": ["voc", "vos", "all"],
            "description": "Which voice to sync"
          }
        }
      }
    },
    {
      "name": "pft_link",
      "description": "Link a feedback item to a local issue",
      "inputSchema": {
        "type": "object",
        "properties": {
          "feedback_id": {
            "type": "string",
            "description": "Feedback item ID"
          },
          "issue_path": {
            "type": "string",
            "description": "Path to local issue file (e.g., docs/issues/042-feature.md)"
          }
        },
        "required": ["feedback_id", "issue_path"]
      }
    },
    {
      "name": "pft_report",
      "description": "Generate feedback report",
      "inputSchema": {
        "type": "object",
        "properties": {
          "format": {
            "type": "string",
            "enum": ["summary", "detailed", "status"],
            "description": "Report format"
          },
          "voice": {
            "type": "string",
            "enum": ["voc", "vos", "all"],
            "description": "Voice type to include"
          }
        }
      }
    },
    {
      "name": "pft_user_list",
      "description": "List users in the feedback system",
      "inputSchema": {
        "type": "object",
        "properties": {
          "category": {
            "type": "string",
            "enum": ["customer", "partner", "employee", "prospect", "all"],
            "description": "User category filter"
          },
          "role": {
            "type": "string",
            "enum": ["user", "admin", "moderator", "observer", "all"],
            "description": "Filter by role"
          }
        }
      }
    },
    {
      "name": "pft_user_add",
      "description": "Add a new user to the feedback system",
      "inputSchema": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "description": "User's display name"
          },
          "email": {
            "type": "string",
            "description": "User's email address"
          },
          "category": {
            "type": "string",
            "enum": ["customer", "partner", "employee", "prospect"],
            "description": "User category"
          },
          "role": {
            "type": "string",
            "enum": ["user", "admin", "moderator", "observer"],
            "description": "User role (default: user)"
          },
          "company": {
            "type": "string",
            "description": "User's company/organization"
          }
        },
        "required": ["name", "email", "category"]
      }
    },
    {
      "name": "pft_user_show",
      "description": "Show details of a specific user",
      "inputSchema": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "User ID or email"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "pft_user_update",
      "description": "Update user information",
      "inputSchema": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "User ID or email"
          },
          "name": {
            "type": "string",
            "description": "New display name"
          },
          "category": {
            "type": "string",
            "enum": ["customer", "partner", "employee", "prospect"],
            "description": "New category"
          },
          "company": {
            "type": "string",
            "description": "New company/organization"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "pft_user_role",
      "description": "Assign or change role for a user",
      "inputSchema": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "User ID or email"
          },
          "role": {
            "type": "string",
            "enum": ["user", "admin", "moderator", "observer"],
            "description": "Role to assign"
          }
        },
        "required": ["id", "role"]
      }
    },
    {
      "name": "pft_user_link",
      "description": "Link local user to external feedback system user ID",
      "inputSchema": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "Local user ID or email"
          },
          "external_id": {
            "type": "string",
            "description": "External system user ID"
          },
          "provider": {
            "type": "string",
            "enum": ["fider", "clearflask", "eververse"],
            "description": "External provider name"
          }
        },
        "required": ["id", "external_id"]
      }
    },
    {
      "name": "pft_user_remove",
      "description": "Remove a user from the feedback system",
      "inputSchema": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "User ID or email"
          },
          "keep_feedback": {
            "type": "boolean",
            "description": "Keep user's feedback items (default: true)"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "pft_role_list",
      "description": "List available roles and their permissions",
      "inputSchema": {
        "type": "object",
        "properties": {}
      }
    },
    {
      "name": "pft_notify",
      "description": "Send notification to users about feedback item",
      "inputSchema": {
        "type": "object",
        "properties": {
          "feedback_id": {
            "type": "string",
            "description": "Feedback item ID"
          },
          "notification_type": {
            "type": "string",
            "enum": ["vote", "description", "acceptance", "status_change"],
            "description": "Type of notification"
          },
          "recipients": {
            "type": "string",
            "enum": ["all-voc", "all-vos", "specific"],
            "description": "Recipient group"
          },
          "email": {
            "type": "string",
            "description": "Specific email if recipients is 'specific'"
          }
        },
        "required": ["feedback_id", "notification_type"]
      }
    },
    {
      "name": "pft_status",
      "description": "Check status of feedback tool deployment",
      "inputSchema": {
        "type": "object",
        "properties": {}
      }
    },
    {
      "name": "pft_export",
      "description": "Export feedback data to various formats",
      "inputSchema": {
        "type": "object",
        "properties": {
          "format": {
            "type": "string",
            "enum": ["md", "json", "csv"],
            "description": "Export format"
          },
          "output": {
            "type": "string",
            "description": "Output file path"
          }
        },
        "required": ["format"]
      }
    }
  ]
}
```

### MCP Resources

```json
{
  "resources": [
    {
      "name": "pft://config",
      "description": "Current PFT configuration",
      "mimeType": "application/json"
    },
    {
      "name": "pft://cache",
      "description": "Local cache of feedback items",
      "mimeType": "application/json"
    },
    {
      "name": "pft://voc/{id}",
      "description": "Voice of Customer feedback item",
      "mimeType": "text/markdown"
    },
    {
      "name": "pft://vos/{id}",
      "description": "Voice of Stakeholder requirement",
      "mimeType": "text/markdown"
    },
    {
      "name": "pft://users",
      "description": "List of all users in the feedback system",
      "mimeType": "application/json"
    },
    {
      "name": "pft://users/{id}",
      "description": "Specific user details and feedback history",
      "mimeType": "application/json"
    },
    {
      "name": "pft://roles",
      "description": "Available roles and their permissions",
      "mimeType": "application/json"
    }
  ]
}
```

## Implementation Phases

### Phase 1: MCP Server Foundation
- [ ] Implement MCP protocol handler in ptx-mcp
- [ ] Add stdio transport (standard for MCP)
- [ ] Implement tool discovery endpoint
- [ ] Implement resource discovery endpoint

### Phase 2: PFT Tool Wrappers
- [ ] Implement `pft_list` tool
- [ ] Implement `pft_show` tool
- [ ] Implement `pft_sync` tool
- [ ] Implement `pft_link` tool
- [ ] Implement `pft_status` tool

### Phase 3: User Management Tools
- [ ] Implement `pft_user_list` tool
- [ ] Implement `pft_user_add` tool
- [ ] Implement `pft_user_show` tool
- [ ] Implement `pft_user_update` tool
- [ ] Implement `pft_user_role` tool
- [ ] Implement `pft_user_link` tool
- [ ] Implement `pft_user_remove` tool
- [ ] Implement `pft_role_list` tool

### Phase 4: Advanced Tools
- [ ] Implement `pft_report` tool
- [ ] Implement `pft_export` tool
- [ ] Implement `pft_notify` tool

### Phase 5: MCP Resources
- [ ] Implement config resource
- [ ] Implement cache resource
- [ ] Implement feedback item resources (voc/vos)
- [ ] Implement user resources (pft://users/{id})

### Phase 6: Testing & Documentation
- [ ] Unit tests for MCP tools
- [ ] Integration tests with Claude Code
- [ ] Documentation for AI assistant usage
- [ ] Example prompts for common workflows

## Technical Requirements

1. **MCP Protocol Compliance**
   - Follow Model Context Protocol specification
   - Support stdio transport
   - Implement proper error handling and responses

2. **ptx-pft Integration**
   - Use `exec.Command("ptx-pft", ...)` for all operations
   - Parse JSON output from ptx-pft commands
   - Handle errors and timeouts gracefully

3. **Configuration Discovery**
   - Auto-detect .pft-config.json in current/parent directories
   - Support explicit config path via MCP initialization

## Dependencies

- **ptx-pft**: Product Feedback Tool helper (Issue #107)
- **MCP SDK**: Model Context Protocol Go SDK (if available) or custom implementation

## Success Criteria

### Feedback Operations
- [ ] Claude Code can list feedback items via MCP
- [ ] Claude Code can sync feedback via MCP
- [ ] Claude Code can generate reports via MCP

### User Management
- [ ] Claude Code can list users with category/role filters
- [ ] Claude Code can add new users via MCP
- [ ] Claude Code can update user roles via MCP
- [ ] Claude Code can link users to external systems
- [ ] Claude Code can view user details and feedback history

### General
- [ ] All MCP tools return proper JSON responses
- [ ] Error handling provides useful messages to AI
- [ ] User management operations are properly validated

## Example AI Interactions

### Feedback Management
```
User: "What feedback items do we have from customers?"
AI: [Uses pft_list tool with voice="voc"]
    "You have 12 customer feedback items. 5 are open, 3 are planned..."

User: "Sync the latest feedback from Fider"
AI: [Uses pft_sync tool with direction="pull"]
    "Synchronized 3 new items and updated 2 existing items."

User: "Generate a summary report of all feedback"
AI: [Uses pft_report tool with format="summary"]
    "Here's the feedback summary: ..."

User: "Link feedback #42 to issue docs/issues/110-eververse.md"
AI: [Uses pft_link tool]
    "Successfully linked feedback #42 to issue #110."
```

### User Management
```
User: "Show me all customer users"
AI: [Uses pft_user_list tool with category="customer"]
    "Found 15 customer users: John Doe (john@example.com), ..."

User: "Add a new partner user named Jane Smith from Acme Corp"
AI: [Uses pft_user_add tool with name="Jane Smith", email="jane@acme.com", category="partner", company="Acme Corp"]
    "Successfully added Jane Smith as a partner user."

User: "Make user john@example.com a moderator"
AI: [Uses pft_user_role tool with id="john@example.com", role="moderator"]
    "Updated John Doe's role to moderator."

User: "Link our user jane@acme.com to Fider user ID 42"
AI: [Uses pft_user_link tool with id="jane@acme.com", external_id="42", provider="fider"]
    "Linked Jane Smith to Fider user #42."

User: "What roles are available in the system?"
AI: [Uses pft_role_list tool]
    "Available roles: user (basic access), moderator (can edit), admin (full access), observer (read-only)."

User: "Show details for user john@example.com"
AI: [Uses pft_user_show tool with id="john@example.com"]
    "John Doe (john@example.com) - Customer, Moderator role, 8 feedback items submitted..."
```

## Notes

- MCP is designed for AI assistant integration, not direct user interaction
- All tools should return structured JSON for AI parsing
- Consider rate limiting for sync operations
- ptx-pft must be available in PATH or co-located with ptx-mcp

---

**Created**: 2025-12-24
**Related Issues**: #107 (ptx-pft), #4 (MCP Server)
