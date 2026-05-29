# edge

Package `edge` implements management of a VPS edge/bastion host — a reverse
proxy (Caddy), a WireGuard VPN tunnel back to internal services, and basic
security hardening (firewall, Fail2ban). It is the backing library for the
`portunix edge` CLI command (see `src/cmd/edge.go`).

## Scope

The package produces and manages an opinionated edge stack built from
containers. A typical deployment exposes public HTTPS on the VPS, terminates
TLS in Caddy, and forwards traffic over WireGuard to an upstream host on the
private network.

Supported platform: **Linux** (deployment is gated on `runtime.GOOS == "linux"`).
Container runtime: **Podman** (preferred) or **Docker** as a fallback.

## Files

| File | Purpose |
| ---- | ------- |
| `manager.go` | `Manager` type and top-level operations: `InitializeConfiguration`, `Deploy`, `Start`, `Stop`, `ShowStatus`, `ShowLogs`, `AddDomain`, `AddVPNClient`. |
| `config.go` | YAML configuration schema (`Config`, `EdgeConfig`, `DomainConfig`, `VPNConfig`, `SecurityConfig`, …) and `LoadConfig`. |
| `wireguard.go` | WireGuard key generation and client/server configuration helpers. |
| `helpers.go` | Container runtime shims (status, start/stop, logs) using `podman` with `docker` fallback. |

## Directory layout produced by `edge init`

```text
<config-dir>/
├── edge-config.yaml      # main configuration (generated with placeholders)
├── caddy/                # Caddyfile and runtime data/config
├── wireguard/            # wg0.conf and client configs
├── fail2ban/             # fail2ban jail configuration
├── logs/                 # service logs
└── backup/               # backup target
```

## Typical flow

1. `portunix edge init <name>` — scaffold `edge-config/<name>/` with default
   `edge-config.yaml` and subdirectories.
2. Edit `edge-config.yaml` — set `server.public_ip`, domains, upstreams, and
   VPN clients.
3. `portunix edge deploy <config-path>` — validate config, render
   `Caddyfile`, `wg0.conf`, and `docker-compose.yml`, then bring up the stack
   with `podman-compose` / `docker-compose`.
4. `portunix edge status` / `start` / `stop` / `logs` — runtime management.

## Configuration reference

The YAML schema is defined in `config.go`. Top-level keys:

- `edge.server` — public IP, SSH port, admin e-mail
- `edge.domains[]` — name, upstream host/port, TLS provider, optional paths
- `edge.vpn` — WireGuard network, server IP, port, peer list
- `edge.security.firewall` / `edge.security.fail2ban`
- `edge.containers` — runtime (`podman` | `docker`), network name, auto-update
- `edge.monitoring`, `edge.backup` — optional sections

## Extension points

- **Templates**: `generateCaddyfile` looks up `assets/templates/edge/Caddyfile.basic`
  first and falls back to an embedded minimal template. Ship new templates in
  `assets/templates/edge/` to override defaults.
- **Container runtime**: selected per `edge.containers.runtime`; helpers in
  `helpers.go` try Podman first and fall back to Docker transparently.

## Related

- CLI wiring: `src/cmd/edge.go`
- Templates: `assets/templates/edge/`
