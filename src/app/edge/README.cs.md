# edge

Balíček `edge` implementuje správu VPS edge/bastion hostitele — reverzní proxy
(Caddy), WireGuard VPN tunel k interním službám a základní bezpečnostní
opatření (firewall, Fail2ban). Slouží jako podkladová knihovna pro CLI příkaz
`portunix edge` (viz `src/cmd/edge.go`).

## Rozsah

Balíček generuje a spravuje definovaný edge stack postavený na kontejnerech.
Typické nasazení vystavuje veřejné HTTPS na VPS, terminuje TLS v Caddy a
forwarduje provoz přes WireGuard na upstream hostitele v privátní síti.

Podporovaná platforma: **Linux** (nasazení je podmíněno `runtime.GOOS == "linux"`).
Kontejnerový runtime: **Podman** (preferovaný) nebo **Docker** jako fallback.

## Soubory

| Soubor | Účel |
| ------ | ---- |
| `manager.go` | Typ `Manager` a hlavní operace: `InitializeConfiguration`, `Deploy`, `Start`, `Stop`, `ShowStatus`, `ShowLogs`, `AddDomain`, `AddVPNClient`. |
| `config.go` | YAML schéma konfigurace (`Config`, `EdgeConfig`, `DomainConfig`, `VPNConfig`, `SecurityConfig`, …) a funkce `LoadConfig`. |
| `wireguard.go` | Generování WireGuard klíčů a konfigurace klienta/serveru. |
| `helpers.go` | Obálky kontejnerového runtime (status, start/stop, logy) — `podman` s fallbackem na `docker`. |

## Struktura adresáře po `edge init`

```text
<config-dir>/
├── edge-config.yaml      # hlavní konfigurace (vygenerována s placeholdery)
├── caddy/                # Caddyfile a runtime data/config
├── wireguard/            # wg0.conf a klientské konfigurace
├── fail2ban/             # konfigurace Fail2ban jailů
├── logs/                 # logy služeb
└── backup/               # cíl pro zálohy
```

## Typický průběh

1. `portunix edge init <name>` — vytvoří `edge-config/<name>/` s výchozím
   `edge-config.yaml` a podadresáři.
2. Úprava `edge-config.yaml` — nastavení `server.public_ip`, domén, upstreamů
   a VPN klientů.
3. `portunix edge deploy <config-path>` — validace konfigurace, vygenerování
   `Caddyfile`, `wg0.conf` a `docker-compose.yml`, spuštění stacku přes
   `podman-compose` / `docker-compose`.
4. `portunix edge status` / `start` / `stop` / `logs` — runtime správa.

## Reference konfigurace

YAML schéma je definováno v `config.go`. Hlavní klíče:

- `edge.server` — veřejná IP, SSH port, admin e-mail
- `edge.domains[]` — jméno, upstream host/port, TLS provider, volitelné cesty
- `edge.vpn` — WireGuard síť, IP serveru, port, seznam peerů
- `edge.security.firewall` / `edge.security.fail2ban`
- `edge.containers` — runtime (`podman` | `docker`), jméno sítě, auto-update
- `edge.monitoring`, `edge.backup` — volitelné sekce

## Rozšiřovací body

- **Šablony**: `generateCaddyfile` hledá nejprve
  `assets/templates/edge/Caddyfile.basic` a jako fallback použije vestavěnou
  minimální šablonu. Vlastní šablony přidejte do `assets/templates/edge/`.
- **Container runtime**: vybírá se podle `edge.containers.runtime`; helpery
  v `helpers.go` nejprve zkouší Podman a transparentně přecházejí na Docker.

## Související

- CLI napojení: `src/cmd/edge.go`
- Šablony: `assets/templates/edge/`
