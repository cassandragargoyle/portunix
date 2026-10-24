# ptx-proxmox

Portunix helper binary for managing VMs and LXC containers on remote Proxmox VE hypervisors.

## Overview

`ptx-proxmox` talks to the **Proxmox VE REST API** and covers the whole life of a guest:
connection profiles, listing, lifecycle, creation from templates with cloud-init, snapshots
and SSH/SCP access with automatic IP resolution. It needs no agent on the Proxmox host and no
access to the web UI.

Typically invoked through the dispatcher as `portunix proxmox <command>`, or directly as
`ptx-proxmox proxmox <command>`. Guests are addressed by name or by numeric VMID.

Introduced by [issue #167](../../../docs/issues/done/167-proxmox-vm-ct-management.md).

## Quick Start

```bash
# Save a connection profile (API token, recommended for automation)
portunix proxmox auth login --host pve.example.com \
  --token-id 'automation@pve!portunix' --token-secret '<secret>'

# Check the profile and the credentials
portunix proxmox auth status

# List guests and create a VM from a template
portunix proxmox list
portunix proxmox create vm --name web-01 --node pve1 --template ubuntu-24.04-cloud

# Work with the guest
portunix proxmox ssh web-01
portunix proxmox snapshot create web-01 before-upgrade
```

## Connection Profiles

Profiles are stored in `proxmox.json` in the user config directory (`~/.config/portunix/` on
Linux). The file is written atomically with mode `0600`, its directory with `0700`. The
environment variable `PORTUNIX_PROXMOX_CONFIG` overrides the path.

| Command | Description |
| ------- | ----------- |
| `auth login` | Save a profile and make it the current one |
| `auth status` | Show the active profile and validate its credentials |
| `auth list` | List stored profiles |
| `auth logout` | Remove a stored profile |

Options of `auth login`:

| Option | Description |
| ------ | ----------- |
| `--host` | Proxmox VE hostname or IP (required) |
| `--port` | API port, default `8006` |
| `--profile` | Profile name, defaults to `default` or the host |
| `--token-id`, `--token-secret` | API token in `user@realm!tokenname` form and its secret |
| `--user` | Username in `user@realm` form for password authentication |
| `--insecure`, `--no-verify` | Skip TLS certificate verification |

Two authentication types are supported:

- **API token** — the token secret is stored in the profile in plain JSON, protected only by
  the file mode. Revoke the token in the Proxmox UI if the file leaks.
- **Password** — the password is prompted without echo and is never stored. A short-lived
  ticket is acquired for every session.

Most commands accept `--profile <name>` to use a profile other than the current one.

### TLS Verification

Certificates are verified by default. Proxmox VE ships with a self-signed certificate, so a
fresh installation needs `--insecure` at login. The choice is stored in the profile as
`verify_tls` and applies to every later command using that profile. Prefer installing a
trusted certificate on the Proxmox host over disabling verification.

## Commands

### Inspecting Guests

| Command | Description | Example |
| ------- | ----------- | ------- |
| `list` | List VMs and LXC containers | `portunix proxmox list --type vm --node pve1` |
| `info` | Show the configuration of a guest | `portunix proxmox info web-01` |
| `status` | Show the runtime status of a guest | `portunix proxmox status 105` |
| `resolve-ip` | Print the best-effort IP address of a guest | `portunix proxmox resolve-ip web-01` |

`list` filters with `--type vm|ct` and `--node <name>` and prints a table or, with
`--format json`, machine-readable output.

### Lifecycle

| Command | Description |
| ------- | ----------- |
| `start` | Start a guest |
| `shutdown` | Send a graceful shutdown signal |
| `stop` | Hard-stop a guest |
| `restart` | Reboot a guest |
| `delete` | Destroy a guest |

Lifecycle commands wait for the Proxmox task to finish by default. Use `--wait=false` to
return immediately and `--timeout <seconds>` to change the limit (default `300`).

`delete` is irreversible. `--force` stops a running guest first and `--purge` also removes it
from backup jobs and the HA configuration.

### Creating Guests

| Command | Description |
| ------- | ----------- |
| `create vm` | Create a QEMU VM from a template |
| `create ct` | Create an LXC container from a template |
| `template list` | List templates and ISO images on Proxmox storage |
| `cloud-init` | Configure cloud-init on an existing VM |

Options of `create vm` and `create ct`:

| Option | Default | Description |
| ------ | ------- | ----------- |
| `--name` | | Name of the new guest (required) |
| `--node` | | Target node (required) |
| `--template` | | Template volume ID or image name (required) |
| `--vmid` | `0` | VMID to assign, `0` allocates the next free one |
| `--cores` | `1` | Number of CPU cores |
| `--memory` | `1024` | Memory in MiB |
| `--disk` | `8` | Disk size in GiB |
| `--storage` | `local-lvm` | Storage pool for disks |
| `--ssh-key` | | Path to an SSH public key to install |
| `--password` | | Root or default user password, required for a CT unless `--ssh-key` is given |
| `--wait`, `--timeout` | `true`, `600` | Wait for the Proxmox task and its limit in seconds |

`cloud-init` sets the default user (`--user`, `--password`, `--ssh-key`) and the network
(`--ip dhcp` or a CIDR such as `10.0.0.5/24` together with `--gateway`, plus `--nameserver`).

```bash
portunix proxmox create vm --name web-01 --node pve1 --template ubuntu-24.04-cloud \
  --cores 2 --memory 4096 --disk 32 --ssh-key ~/.ssh/id_ed25519.pub
portunix proxmox cloud-init web-01 --user deploy --ip 10.0.0.5/24 --gateway 10.0.0.1
```

### Snapshots

| Command | Description |
| ------- | ----------- |
| `snapshot create <guest> <name>` | Create a snapshot, `--include-ram` also saves RAM state (QEMU only) |
| `snapshot list <guest>` | List snapshots of a guest |
| `snapshot revert <guest> <name>` | Revert a guest to a snapshot |
| `snapshot delete <guest> <name>` | Delete a snapshot |

### SSH and File Transfer

| Command | Description | Example |
| ------- | ----------- | ------- |
| `ssh` | Open an SSH session | `portunix proxmox ssh web-01 -u deploy` |
| `exec` | Run a command over SSH | `portunix proxmox exec web-01 -- systemctl status nginx` |
| `copy` | Copy files with SCP | `portunix proxmox copy ./app.tar.gz web-01:/tmp/` |

The guest IP is resolved automatically, so no address has to be known in advance:

- **QEMU VM** — through the QEMU guest agent. The `qemu-guest-agent` package must be installed
  inside the VM and the agent enabled in the VM configuration.
- **LXC container** — through the container's network interfaces reported by Proxmox.

Common options: `-u, --user` (default `root`), `-p, --port` (default `22`), `-i, --identity`
for a key file, `-o KEY=VALUE` for extra `ssh` options (repeatable) and `--ipv6` to prefer an
IPv6 address. Arguments after `--` are passed to `ssh` unchanged.

`copy` uses the form `<name|vmid>:/path` for the remote side, in either direction:

```bash
portunix proxmox copy ./local.txt my-vm:/tmp/
portunix proxmox copy my-vm:/var/log/app.log ./
portunix proxmox copy -r ./dist my-vm:/opt/app/
```

The system `ssh` and `scp` clients are used, they must be available on `PATH`.

## Playbook Integration

`.ptxbook` playbooks can target a Proxmox guest with `environment.type: proxmox`, including
`create_if_missing` to provision the guest on first run. The playbook side is handled by
`ptx-ansible`, which calls this helper.

## Source Layout

| File | Content |
| ---- | ------- |
| `main.go` | Command tree and dispatcher entry point |
| `auth.go` | `auth` commands, password prompt |
| `config.go` | Profile storage in `proxmox.json` |
| `client.go` | REST client, authentication headers, task waiting |
| `resources.go` | Guest lookup by name or VMID across the cluster |
| `vm.go` | `list`, `info`, `status` and lifecycle commands |
| `create.go`, `template.go`, `cloudinit.go` | Guest creation, template listing, cloud-init |
| `snapshot.go` | Snapshot commands |
| `ssh.go` | IP resolution, `ssh`, `exec`, `copy`, `resolve-ip` |
| `types.go` | Shared types (profile, config, resource) |

## Development

```bash
# Build the main binary and all helpers
make build

# Unit tests, they use a local HTTP test server and need no Proxmox host
cd src/helpers/ptx-proxmox && go test ./...
```

---

**Created**: 2026-10-11
**Last Updated**: 2026-10-11
