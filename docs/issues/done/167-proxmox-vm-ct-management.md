# Issue #167: Proxmox VM/CT Management Commands

**Type**: Feature
**Priority**: High
**Labels**: feature, proxmox, virtualization, infrastructure, deployment
**Status**: ✅ Implemented
**Created**: 2026-04-03
**Implemented**: 2026-04-20 (Phases 1–4, see acceptance-167.md + acceptance-167-phases-2-4.md)

---

## Summary

Add `portunix proxmox` command group for managing VMs and LXC containers (CTs) on Proxmox VE servers via the Proxmox REST API. This enables automated provisioning and deployment workflows without requiring manual Proxmox web UI interaction.

## Motivation

The `portunix virt` commands support local QEMU/KVM and VirtualBox VMs, but there is no way to manage VMs or CTs on remote Proxmox VE hypervisors. Projects like portunix-synapse need to deploy application stacks to Proxmox infrastructure, and currently this requires manual steps or custom Python scripts with direct Proxmox API calls.

Adding native Proxmox support to portunix would:
- Enable end-to-end deployment automation from build to production
- Reuse existing `portunix virt ssh/exec/copy` for post-provisioning
- Allow `.ptxbook` playbooks to target Proxmox VMs/CTs
- Follow the same UX patterns as existing `portunix virt` and `portunix container` commands

## Design

### Authentication

```bash
# Configure Proxmox connection (stored in ~/.config/portunix/proxmox.json)
portunix proxmox auth --host pve.example.com --user root@pam
portunix proxmox auth --host pve.example.com --token-id user@pam!tokenid --token-secret <secret>
portunix proxmox auth status
```

Support both password-based (interactive) and API token authentication. Token-based is preferred for automation.

### VM Management

```bash
# List VMs and CTs on all nodes
portunix proxmox list
portunix proxmox list --node pve1
portunix proxmox list --type vm        # Only VMs (QEMU)
portunix proxmox list --type ct        # Only CTs (LXC)

# Create VM from template or ISO
portunix proxmox create vm --name synapse-registry \
    --template ubuntu-22.04-cloud --cores 2 --memory 4096 --disk 32 \
    --node pve1 --storage local-zfs

# Create LXC CT from template
portunix proxmox create ct --name synapse-web \
    --template ubuntu-22.04-standard --cores 1 --memory 2048 --disk 16 \
    --node pve1 --storage local-zfs

# Lifecycle
portunix proxmox start <vmid|name>
portunix proxmox stop <vmid|name>
portunix proxmox restart <vmid|name>
portunix proxmox delete <vmid|name>

# Information
portunix proxmox info <vmid|name>
portunix proxmox status <vmid|name>
portunix proxmox console <vmid|name>    # Open noVNC or SPICE console

# Snapshots
portunix proxmox snapshot create <vmid|name> <snapshot-name>
portunix proxmox snapshot list <vmid|name>
portunix proxmox snapshot revert <vmid|name> <snapshot-name>
portunix proxmox snapshot delete <vmid|name> <snapshot-name>
```

### SSH Integration

Reuse `portunix virt ssh/exec/copy` patterns but with Proxmox IP resolution:

```bash
# SSH into Proxmox VM/CT (resolves IP from Proxmox QEMU agent or CT config)
portunix proxmox ssh <vmid|name>
portunix proxmox exec <vmid|name> "command"
portunix proxmox copy ./local-file <vmid|name>:/remote/path
```

### Template Management

```bash
# List available templates on Proxmox storage
portunix proxmox template list
portunix proxmox template download <template-name> --storage local

# Cloud-init configuration
portunix proxmox cloud-init <vmid|name> \
    --user deploy --ssh-key ~/.ssh/id_rsa.pub \
    --ip dhcp --nameserver 8.8.8.8
```

### Playbook Integration

Extend `.ptxbook` environment options:

```yaml
spec:
  environment:
    type: proxmox
    host: pve.example.com
    target: synapse-registry     # VM/CT name or ID
    create_if_missing: true
    template: ubuntu-22.04-cloud
    resources:
      cores: 2
      memory: 4096
      disk: 32
```

```bash
# Deploy playbook to Proxmox VM
portunix playbook run deploy.ptxbook --env proxmox --target synapse-registry
```

## Implementation

### Phase 1: Core API Client

- Proxmox REST API client in Go (`src/app/proxmox/client.go`)
- Authentication: password + API token support
- Configuration storage in `~/.config/portunix/proxmox.json`
- Add to `ptx-virt` helper or create new `ptx-proxmox` helper

### Phase 2: VM/CT Lifecycle

- `portunix proxmox list/create/start/stop/delete`
- `portunix proxmox info/status`
- QEMU VM and LXC CT support
- Template-based creation with cloud-init

### Phase 3: SSH + Snapshot

- `portunix proxmox ssh/exec/copy` with automatic IP resolution
- `portunix proxmox snapshot create/list/revert/delete`
- Integration with existing `portunix virt ssh` code

### Phase 4: Playbook Integration

- Extend `.ptxbook` parser for `proxmox` environment type
- Auto-create VM/CT if `create_if_missing: true`
- Post-creation provisioning via SSH

---

## Proxmox REST API Reference

- Base URL: `https://<host>:8006/api2/json/`
- Authentication: `POST /access/ticket` or `PVEAPIToken=<tokenid>=<secret>` header
- VMs: `/nodes/{node}/qemu/{vmid}/...`
- CTs: `/nodes/{node}/lxc/{vmid}/...`
- Templates: `/nodes/{node}/storage/{storage}/content`

---

## Acceptance Criteria

### AC-1: Authentication
- [ ] API token and password authentication work
- [ ] Credentials stored securely in config file
- [ ] `portunix proxmox auth status` shows connection info

### AC-2: VM/CT Management
- [ ] List, create, start, stop, delete VMs and CTs
- [ ] Template-based creation with resource configuration
- [ ] Cloud-init support for initial setup

### AC-3: SSH Integration
- [ ] `portunix proxmox ssh` resolves IP and connects
- [ ] `portunix proxmox exec` runs remote commands
- [ ] `portunix proxmox copy` transfers files

### AC-4: Playbook Integration
- [ ] `.ptxbook` files support `proxmox` environment type
- [ ] Automatic VM/CT creation from playbook spec

---

## Related

- [Proxmox Server Specification](../../infrastructure/001-proxmox-server-specification.md) — planned hardware
- [ADR-010: Temporary Virtualization Priority](../../adr/010-temporary-virtualization-priority.md) — current QEMU/VBox strategy
- portunix-synapse `scripts/deploy-proxmox.py` — consumer of these commands
