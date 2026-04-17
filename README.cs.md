# Portunix

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org)
[![Coverage](https://img.shields.io/badge/Coverage-0%25-red.svg)](https://codecov.io/gh/cassandragargoyle/portunix)
[![Build Status](https://github.com/cassandragargoyle/portunix/workflows/Test%20Suite/badge.svg)](https://github.com/cassandragargoyle/portunix/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cassandragargoyle/portunix)](https://goreportcard.com/report/github.com/cassandragargoyle/portunix)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 🌐 **Jazyk / Language**: **Čeština** | [English](README.md)

**Jednotná AI platforma pluginů a úloh pro vývojová prostředí** — s inteligentní detekcí OS, podporou Docker kontejnerů a automatizovanou instalací softwaru.

> Zjednodušte svůj vývojový workflow napříč Windows, Linux a macOS jediným, mocným CLI nástrojem.

## Proč Portunix?

- **Jeden nástroj, všechny platformy**: Konzistentní práce napříč Windows, Linux a macOS
- **Chytrá automatizace**: Inteligentní detekce OS a automatická konfigurace
- **Container-First**: Prvotřídní podpora Dockeru a Podmana se SSH-enabled kontejnery
- **Zaměřeno na vývojáře**: Vytvořeno vývojáři pro vývojáře

## Rychlá instalace

### Z Releases (doporučeno)

```bash
# Linux (amd64)
wget https://github.com/cassandragargoyle/portunix/releases/latest/download/portunix_linux_amd64.tar.gz
tar -xzf portunix_linux_amd64.tar.gz
sudo mv portunix /usr/local/bin/

# Ověření instalace
portunix version
```

### Ze zdrojových kódů

```bash
git clone https://github.com/cassandragargoyle/portunix.git
cd Portunix
make build
```

## Funkce

### Základní schopnosti

- **Univerzální instalační systém**: Instalace vývojových nástrojů napříč Windows, Linux a macOS
- **Správa Dockeru**: Kompletní správa životního cyklu Docker kontejnerů s podporou SSH
- **Správa certifikátů**: Automatická instalace CA certifikátů a ověření HTTPS konektivity
- **Integrace s Windows Sandbox**: Izolovaná vývojová prostředí na Windows
- **Inteligentní detekce OS**: Automatická detekce platformy a optimalizace včetně detekce certifikačních balíčků
- **Multiplatformní podpora**: Nativní podpora pro Windows, Linux a macOS
- **Správa VM**: Virtualizace QEMU/KVM s podporou Windows 11 a snapshotů

### Správa Dockeru

- **Inteligentní instalace Dockeru**: OS-specifická instalace Dockeru s optimalizací úložiště
- **Multiplatformní kontejnery**: Podpora Ubuntu, Alpine, CentOS, Debian a vlastních obrazů
- **SSH-enabled kontejnery**: Automatické nastavení SSH serveru s vygenerovanými přihlašovacími údaji
- **Detekce správce balíčků**: Automatická detekce apt-get, yum, dnf, apk
- **Mountování cache adresářů**: Perzistentní úložiště pro stahované soubory a balíčky
- **Flexibilní základní obrazy**: Výběr z různých distribucí Linuxu

### Správa certifikátů

- **Automatické nastavení CA certifikátů**: Instaluje CA certifikáty do kontejnerů před instalací softwaru
- **Ověření HTTPS konektivity**: Testuje HTTPS připojení po instalaci certifikátů
- **Podpora více distribucí**: Funguje se správci balíčků apt, yum, dnf, apk, pacman, zypper
- **Detekce systémových certifikátů**: Zobrazuje stav certifikačního balíčku v systémových informacích
- **Samostatná instalace certifikátů**: Příkaz `portunix install ca-certificates`

### Prostředí pro dokumentaci

- **Systém šablon Playbook**: Generování dokumentačních stránek jediným příkazem
- **Podporované enginy**: Docusaurus, Hugo, Docsy (Hugo + Google Docsy theme), Docsify
- **Založeno na kontejnerech**: Spouštění dokumentačních prostředí v Docker/Podman kontejnerech
- **Workflow se sdílenou složkou**: Editace lokálně, okamžité změny přes live-reload
- **Automatické řešení závislostí**: `portunix install docusaurus` automaticky instaluje Node.js
- **Quickstart skript**: PowerShell skript na jeden řádek pro Windows uživatele

### Systém pluginů

- **Architektura založená na gRPC**: Vysokovýkonná komunikace pluginů
- **Životní cyklus pluginu**: Install, enable, disable, start, stop, uninstall
- **Vytváření pluginů**: Generování šablon nových pluginů příkazem `portunix plugin create`
- **Registr pluginů**: Centralizované objevování a správa pluginů

### Systém samoaktualizace

- **Automatické aktualizace**: `portunix update` kontroluje a instaluje nejnovější verzi
- **SHA256 verifikace**: Bezpečná kontrola kontrolních součtů stahovaných souborů
- **Záloha a rollback**: Automatická záloha před aktualizací s rollbackem při selhání

### Instalační typy

- **`default`**: Python + Java + VSCode (doporučeno)
- **`empty`**: Čisté prostředí bez balíčků
- **`python`**: Vývojové prostředí pro Python
- **`java`**: Vývojové prostředí pro Javu
- **`vscode`**: Nastavení Visual Studio Code

## Rychlý start

### Základní použití

```bash
# Zobrazit nápovědu
portunix --help

# Instalace Dockeru s inteligentní detekcí OS
portunix install docker

# Instalace Dockeru s automatickým přijetím doporučeného úložiště
portunix install docker -y

# Instalace dalšího softwaru
portunix install python java vscode

# Instalace CA certifikátů pro HTTPS konektivitu
portunix install ca-certificates

# Zobrazit informace o systému včetně stavu certifikátů
portunix system info
```

### Správa Docker kontejnerů

```bash
# Spustit Python prostředí v Ubuntu kontejneru
portunix docker run-in-container python

# Spustit Java prostředí v Alpine kontejneru
portunix docker run-in-container java --image alpine:3.18

# Spustit vývojové prostředí s vlastním nastavením
portunix docker run-in-container default \
  --image ubuntu:20.04 \
  --name my-dev-env \
  --port 8080:8080 \
  --keep-running

# Správa kontejnerů
portunix docker list
portunix docker logs <container-id>
portunix docker stop <container-id>
portunix docker remove <container-id>
```

### Nastavení dokumentační stránky

```bash
# Vytvoření Docusaurus dokumentační stránky v kontejneru
portunix playbook init my-docs --template static-docs --engine docusaurus --target container
portunix playbook run my-docs.ptxbook --script create
portunix playbook run my-docs.ptxbook --script dev
# -> Otevřít http://localhost:3000

# Nebo použít Hugo / Docsy / Docsify
portunix playbook init my-docs --template static-docs --engine hugo --target container

# Přímá instalace (automaticky instaluje závislosti)
portunix install docusaurus
portunix install hugo
```

### Windows Sandbox

```bash
# Spustit ve Windows Sandbox s SSH
portunix sandbox run-in-sandbox python

# Vygenerovat vlastní konfiguraci sandboxu
portunix sandbox generate --enable-ssh
```

### Virtuální stroje (QEMU/KVM a VirtualBox)

```bash
# Instalace virtualizačního stacku QEMU/KVM
portunix vm install-qemu

# Kontrola podpory virtualizace
portunix vm check

# Vytvoření Windows 11 VM
portunix vm create win11-vm \
  --iso ~/Downloads/Win11.iso \
  --disk-size 80G \
  --ram 8G \
  --cpus 4 \
  --os windows11

# Správa životního cyklu VM
portunix vm start win11-vm
portunix vm list --all
portunix vm info win11-vm
portunix vm console win11-vm
portunix vm stop win11-vm

# Správa snapshotů pro testování trial softwaru
portunix vm snapshot create win11-vm clean-install \
  --description "Fresh Windows 11 after updates"
portunix vm snapshot list win11-vm
portunix vm snapshot revert win11-vm clean-install
```

## Správa VM (QEMU/KVM)

### Hlavní funkce

- **Podpora dvou VM backendů**: Podpora QEMU/KVM i VirtualBox
- **Multiplatformní virtualizace**: QEMU/KVM pro Linux hostitele, VirtualBox pro Windows/macOS
- **Podpora Windows 11**: Plná podpora Windows 11 s TPM 2.0 a Secure Boot
- **Správa snapshotů**: Vytváření, obnovení a správa VM snapshotů
- **Testování trial softwaru**: Ideální pro testování 30denního trial softwaru
- **Podpora více OS**: Windows, Linux a vlastní instalace OS
- **Konfigurace zdrojů**: Flexibilní konfigurace CPU, RAM a disku
- **Jednotné rozhraní**: Dva způsoby vytváření VM - dedikované příkazy nebo jednotné create rozhraní

### Podporované hostované operační systémy

- **Windows**: Windows 11, Windows 10, Windows Server 2022
- **Linux**: Ubuntu, Debian, CentOS, Fedora, Arch a další
- **Vlastní**: Jakýkoli OS, který podporuje QEMU/KVM

## Funkce Dockeru

### Podporované základní obrazy

- **Ubuntu**: `ubuntu:22.04`, `ubuntu:20.04` (výchozí)
- **Alpine**: `alpine:3.18`, `alpine:latest` (odlehčený)
- **Debian**: `debian:bullseye`, `debian:buster`
- **CentOS**: `centos:8`, `centos:7`
- **Fedora**: `fedora:38`, `fedora:37`
- **Rocky Linux**: `rockylinux:8`, `rockylinux:9`
- **Vlastní**: Jakýkoli Docker obraz z registrů

### Workflow kontejnerů

1. **Výběr obrazu**: Volba základního obrazu nebo použití výchozího Ubuntu 22.04
2. **Detekce správce balíčků**: Automatická detekce apt-get/yum/dnf/apk
3. **Vytvoření kontejneru**: Vytvoření se správnými volume a port mappingy
4. **Nastavení SSH**: Instalace OpenSSH serveru s vygenerovanými přihlašovacími údaji
5. **Instalace softwaru**: Instalace požadovaných balíčků pomocí detekovaného správce balíčků
6. **Připraveno pro vývoj**: SSH přístup se sdíleným pracovním prostorem a cache

## Vývoj a testování

### Rychlé testování

```bash
# Nastavení vývojového prostředí
make dev-setup

# Spustit všechny testy
make test

# Pouze unit testy (rychlé)
make test-unit

# Integrační testy (vyžaduje Docker)
make test-integration

# Pokrytí testy
make test-coverage

# Lintování a kvalita
make lint
```

### Lokální nasazení

```bash
# Sestavení a instalace do lokálního systému (automatická detekce existující instalace)
make deploy-local

# Odstranění z lokálního systému (automatická detekce instalační cesty)
make undeploy-local
```

## Dokumentace

- **[Windows Setup Guide](docs/WINDOWS-SETUP.md)**: Windows-specifické nastavení a UTF-8 konfigurace
- **[TEST_GUIDE.md](TEST_GUIDE.md)**: Kompletní průvodce testováním pro vývojáře
- **[TESTING.md](TESTING.md)**: Architektura testování a standardy
- **[Issues Documentation](docs/issues/README.md)**: Mirror a sledování GitHub issues

## Konfigurace

### Proměnné prostředí

```bash
# Konfigurace Dockeru
export DOCKER_HOST=unix:///var/run/docker.sock

# Vývojový režim
export PORTUNIX_DEBUG=true

# Vlastní adresář cache
export PORTUNIX_CACHE_DIR=/custom/cache/path
```

### Konfigurační soubory

- **Instalační balíčky**: `assets/install-packages.json`
- **Uživatelská konfigurace**: `examples/user-install-config.json`

## Roadmapa

### Aktuální stav

- Multiplatformní systém detekce OS
- Správa Docker/Podman s inteligentní instalací
- Multiplatformní podpora kontejnerů se SSH-enabled vývojovými kontejnery
- Orchestrace kontejnerů s docker-compose/podman-compose
- MCP server pro integraci s AI asistenty
- Systém registru balíčků s automatickým objevováním
- Víceúrovňový systém nápovědy (basic, expert, AI)
- Virtualizace QEMU/KVM s podporou Windows 11 a snapshotů
- Nástroj pro zpětnou vazbu k produktu (ptx-pft) s providery Fider/ClearFlask/Eververse
- AIOps helper pro GPU/AI kontejnerové úlohy
- Make helper pro multiplatformní buildy
- Integrace Ansible infrastructure as code
- Systém samoaktualizace s možností rollbacku
- Komplexní testovací architektura a CI/CD pipeline
- Python development helper s podporou projektově lokálního venv
- Systém pluginů s gRPC architekturou (#7)
- Systém šablon Playbook pro dokumentační prostředí
- Automatické řešení závislostí pro instalaci balíčků
- Podpora Docusaurus, Hugo, Docsy, Docsify

### Plánované funkce

- Správa virtuálního vývojového disku (#8)
- Konfigurovatelné datastore backendy (#9)
- Framework interaktivního průvodce (#14)
- Podpora instalace AI asistentů (#35)
- Integrace VSCode development containers

## Přispívání

```bash
# Klonování a nastavení
git clone https://github.com/cassandragargoyle/portunix.git
cd Portunix
make dev-setup

# Spustit testy
make test

# Kontrola stavu
make status
```

### Pravidla

1. Následujte existující konvence kódu
2. Pište testy pro nové funkce
3. Aktualizujte dokumentaci
4. Spouštějte kontroly kvality: `make lint`
5. Zajistěte, že všechny testy projdou: `make test`

### Proces Pull Requestu

1. Vytvořte feature branch: `git checkout -b feature/my-feature`
2. Implementujte změny s testy
3. Spusťte kontroly kvality: `make ci-test`
4. Odešlete pull request
5. Automatizovaná CI/CD pipeline ověří změny

## Vytvoření releasu

```bash
# Vytvoření releasu (sestaví všechny platformy, vygeneruje poznámky, checksumy)
python3 scripts/make-release.py v1.10.7

# Nahrání na GitHub
python3 scripts/upload-release-to-github.py v1.10.7
```

Release skript automaticky:

- Aktualizuje verzi ve zdrojových souborech a `portunix.rc`
- Vytvoří git tag
- Sestaví multiplatformní binárky pomocí GoReleaser
- Vytvoří platform-specific archivy (Linux, Windows, macOS)
- Vygeneruje release notes a kontrolní součty

## Externí partnerství

### act - lokální runner pro GitHub Actions

Portunix se integruje a přispívá do projektu **[nektos/act](https://github.com/nektos/act)** pro lokální testování GitHub Actions.

- **Projekt**: [nektos/act](https://github.com/nektos/act) - Spouštění GitHub Actions lokálně
- **Web**: [nektosact.com](https://nektosact.com/)
- **Integrace**: Portunix poskytuje bezproblémovou instalaci act a možnosti testování workflow GitHub Actions

## Licence

MIT License - podrobnosti viz soubor [LICENSE](LICENSE).

## Odkazy

- **GitHub**: [cassandragargoyle/portunix](https://github.com/cassandragargoyle/portunix)
- **Issues**: [GitHub Issues](https://github.com/cassandragargoyle/portunix/issues)
- **Dokumentace**: [docs/](docs/)

---

**Univerzální správa vývojového prostředí zjednodušeně.**
