# Issue #108: PTX-PFT E-mail Notifications for User Actions

**Status**: ✅ Implemented
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, helper-binary, ptx-pft, email, notifications, smtp
**Created**: 2025-12-23
**Implemented**: 2025-12-23
**Depends on**: Issue #107 (PTX-PFT Helper)

## Summary

Implement e-mail notification functionality for PTX-PFT to request user actions on feedback items (requirements, feature requests).

## Problem Statement

Currently there is no way to:
1. Request users to vote on a requirement
2. Ask users to provide additional description for a feedback item
3. Request users to define acceptance criteria

## Proposed Solution

Add `pft notify` command to send e-mail notifications with action requests.

### Command Structure

```bash
# Basic usage
portunix pft notify <item-id> --user <user-id> --type <type>

# Examples
portunix pft notify UC001 --user cassandragargoyle@gmail.com --type vote
portunix pft notify REQ001 --user cassandragargoyle@gmail.com --type description
portunix pft notify REQ001 --all-voc --type acceptance

# Preview without sending
portunix pft notify UC001 --user cassandragargoyle@gmail.com --type vote --dry-run
```

### Notification Types

| Type | Description |
|------|-------------|
| `vote` | Request user to vote for/against requirement |
| `description` | Request user to provide more details |
| `acceptance` | Request user to define acceptance criteria |

### Flags

| Flag | Description |
|------|-------------|
| `--user <id>` | Send to specific user (by email/ID) |
| `--all-voc` | Send to all users with VoC role |
| `--all-vos` | Send to all users with VoS role |
| `--type <type>` | Notification type (vote/description/acceptance) |
| `--dry-run` | Show email without sending |

## Technical Implementation

### Phase 1: SMTP Client
**File:** `src/helpers/ptx-pft/email.go` (new)

- SMTPClient struct with host, port, credentials
- SendEmail(to, subject, body) method
- Support for Mailhog in development (ports 3200/3201)

### Phase 2: E-mail Templates

Three template types:
1. **Vote request** - Link to Fider post for voting
2. **Description request** - Request for more details
3. **Acceptance criteria request** - Request to define AC

### Phase 3: CLI Command
**File:** `src/helpers/ptx-pft/main.go`

- Add `notify` subcommand handler
- Parse flags and validate inputs
- Load feedback item from local files or Fider
- Send e-mail via SMTP

### Phase 4: Configuration
**File:** `src/helpers/ptx-pft/config.go`

Add SMTPConfig:
```go
type SMTPConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username,omitempty"`
    Password string `json:"password,omitempty"`
    From     string `json:"from"`
}
```

Default for development (Mailhog):
```json
{
  "smtp": {
    "host": "localhost",
    "port": 3200,
    "from": "noreply@localhost"
  }
}
```

## E-mail Templates

### Vote Request
```
Subject: [Product Name] Vote request: {Title}

Hello {UserName},

We would like your opinion on this requirement:

{Title}
{Description preview...}

Please vote here:
{FiderURL}/posts/{Number}

Regards,
Product Team
```

### Description Request
```
Subject: [Product Name] Please clarify: {Title}

Hello {UserName},

We need more details for this requirement:

{Title}

Current description:
{Description}

Please add details here:
{FiderURL}/posts/{Number}

Regards,
Product Team
```

### Acceptance Criteria Request
```
Subject: [Product Name] Define acceptance criteria: {Title}

Hello {UserName},

Please define acceptance criteria for:

{Title}
{Description}

Add acceptance criteria as comment:
{FiderURL}/posts/{Number}

Regards,
Product Team
```

## Testing

1. Start Fider with Mailhog: `portunix pft example`
2. Send test email: `portunix pft notify UC001 --user test@test.com --type vote`
3. Check Mailhog UI: http://localhost:3200

## Acceptance Criteria

- [x] `pft notify` command implemented
- [x] Vote notification type works
- [x] Description notification type works
- [x] Acceptance criteria notification type works
- [x] `--dry-run` shows email without sending
- [x] `--all-voc` sends to all VoC users
- [x] `--all-vos` sends to all VoS users
- [x] SMTP configuration in config file
- [x] Mailhog integration for development
- [x] E-mail templates are customizable (file-based in assets/templates/)

## Email-Only Mode (Provider: email)

### Overview

PFT can operate in **email-only mode** where the provider is set to `email` instead of `fider`. In this mode:
- No Fider instance is deployed
- Only Mailhog is deployed for email capture/sending (port 3200)
- Users vote and respond by replying to emails
- Synchronization commands are disabled (no external system to sync with)

### Configuration

```json
{
  "provider": "email",
  "smtp": {
    "host": "localhost",
    "port": 3200,
    "from": "noreply@localhost"
  }
}
```

### Deploy Behavior

When `provider: "email"`:
```bash
portunix pft deploy
# Only deploys Mailhog container on port 3200
# No Fider, no PostgreSQL
```

Comparison:
| Provider | Containers | Ports |
|----------|------------|-------|
| `fider` | Fider + PostgreSQL + Mailhog (×2 for VoC/VoS) | 3100, 3101, 3200, 3201 |
| `email` | Mailhog only | 3200 |

### Email-based Voting

In email-only mode, voting is done by replying to the notification email:

**Vote Request Email:**
```
Subject: [Product Name] Vote request: {Title}

Hello {UserName},

We would like your opinion on this requirement:

{Title}
{Description preview...}

To vote, reply to this email with:
  +1  = I support this requirement
  -1  = I do not support this requirement
   0  = I abstain

Regards,
Product Team
```

**Processing Replies:**
- PFT parses incoming emails from Mailhog
- Extracts vote from reply body (+1, -1, 0)
- Updates local markdown file with vote count
- `pft votes <item-id>` shows current votes

### Disabled Commands

When `provider: "email"`, these commands are disabled:
- `pft sync` - No external system to sync with
- `pft pull` - No external system to pull from
- `pft push` - No external system to push to

Error message:
```
Sync commands are not available in email-only mode.
Set provider to 'fider' to enable synchronization.
```

### Use Cases

1. **Simple feedback collection** - Small teams that don't need full Fider UI
2. **Offline-first workflows** - All data stays in local markdown files
3. **Minimal infrastructure** - Only Mailhog container needed
4. **Email-centric teams** - Teams that prefer email over web UI

### Implementation Tasks

- [x] Add `email` provider type validation
- [x] Modify `pft deploy` to only start Mailhog when `provider: email`
- [x] Update email templates for email-based voting (file-based templates)
- [ ] Implement `pft votes <item-id>` command to show votes (future)
- [ ] Implement Mailhog reply parsing for vote collection (future)
- [x] Disable sync/pull/push commands when `provider: email`
- [ ] Update `pft example` to support `--provider email` flag (future)

## Future Enhancements (Out of Scope)

- [ ] HTML email templates
- [ ] Delivery tracking
- [ ] External template files
- [ ] Bulk notification scheduling
- [ ] Integration with Fider native notifications

## Related Issues

- Issue #107: PTX-PFT Product Feedback Tool Helper (parent)

## Files to Modify

| File | Action |
|------|--------|
| `src/helpers/ptx-pft/email.go` | NEW - SMTP client + templates |
| `src/helpers/ptx-pft/config.go` | Add SMTPConfig |
| `src/helpers/ptx-pft/main.go` | Add `notify` command |
