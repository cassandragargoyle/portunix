---
title: "pft"
description: "Product feedback tool integration"
---

# pft

Product feedback tool integration

## Usage

```bash
portunix pft [options] [arguments]
```

## Full Help

```
Usage: portunix pft [subcommand]

Product Feedback Tool Commands:

Project Management:
  project create <name>    - Create new PFT project (default: qfd template)
  project create <name> --template <tpl>
                           - Create project with specific template (qfd, basic)
  info                     - Show methodology documentation
  info --json              - Output as JSON (for MCP integration)

Configuration:
  configure                              - Interactive configuration wizard
  configure --name <name> --path <path>  - Set global settings
  configure --area <voc|vos|vob|voe> ... - Configure per-area provider
  configure --smtp-host <host> ...       - Configure SMTP server
  configure --show                       - Show current configuration

Infrastructure:
  deploy                   - Deploy feedback tool to container
  status                   - Check feedback tool status
  destroy                  - Remove feedback tool instance

Synchronization:
  sync                     - Full bidirectional sync
  pull                     - Pull from external system
  push                     - Push to external system

User/Customer Registry:
  user list                - List all users
  user add                 - Add new user
  user role <id>           - Assign role to user
  user link <id>           - Link user to external ID
  user remove <id>         - Remove user
  role list                - List available roles
  role init                - Initialize default role files

Feedback Management:
  list                     - List all feedback items
  add                      - Add new feedback item
  show <id>                - Show feedback details
  link <id> <issue>        - Link feedback to local issue

Category Management:
  category list            - List categories in area
  category add <id>        - Create new category
  category remove <id>     - Delete category
  category rename <id>     - Rename category
  category show <id>       - Show category details

Item Categorization:
  assign <item-id> --category <cat-id>
                           - Add category to item
  unassign <item-id> --category <cat-id>
                           - Remove category from item
  unassign <item-id> --all - Remove all categories

Reporting:
  report                   - Generate feedback report
  export --format=md       - Export to markdown

Notifications:
  notify <id> --user <email> --type <type>
                           - Send notification to user
  notify <id> --all-voc --type <type>
                           - Notify all VoC users
  notify <id> --all-vos --type <type>
                           - Notify all VoS users

Available providers: clearflask, email, eververse, fider

```

