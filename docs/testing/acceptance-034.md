# Acceptance Protocol - Issue #034

**Issue**: MCP Server Wizard — Advanced Features Completion
**Branch**: `feature/034-mcp-wizard-advanced` (commit `9fcd40e`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-05-12
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — PowerShell + Git Bash

## Test Summary

- Total test scenarios: 30
- Passed: 29
- Failed: 0
- Conditional: 1 (T8d — exit-code semantics for malformed existing Claude Desktop config)

## Test Environment

- Build command: `make build`
- Binaries: `portunix.exe` (24,7 MB) + `ptx-mcp.exe` (14,6 MB) + 17 dalších helperů
- Claude Code CLI: ověřená (responds to `claude mcp list`)
- MCP server: lokální stdio, JSON-RPC 2.0, protokol verze `2024-11-05`

## Test Results

### T1 — Build Sanity

- [x] `make build` PASS, žádné fatální chyby
- [x] Commit `9fcd40e` je HEAD feature větve
- [x] `portunix.exe` + `ptx-mcp.exe` přítomné

### T2 — Help / Flag Visibility

- [x] `portunix mcp init --help` zobrazuje 7 nových flagů:
  - `--scope`, `--env`, `--timeout`, `--from-json`, `--bind-address`, `--tls-cert`, `--tls-key`
- [x] `portunix mcp test --help` zobrazuje aktualizovaný description s "real JSON-RPC initialize handshake"

### T3 — `--from-json` Happy Path

- [x] Načte validní JSON, validuje schéma
- [x] Zapíše do `~/.portunix/mcp-server.json`
- [x] Propaguje konfiguraci do `claude mcp add` s correct flagy
- [x] Vrátí exit=0 a "✅ Import complete"

### T4 — `--from-json` Validation (Negative Cases)

| # | Případ | Očekáváno | Skutečné | Status |
| - | ------ | --------- | -------- | ------ |
| T4a | invalid `scope` (`INVALID`) | exit=1, popisná chyba | exit=1, "scope must be local, project or user" | PASS |
| T4b | `https` bez TLS cert/key | exit=1 | "tls_cert and tls_key are required for https" | PASS |
| T4c | port `99999` | exit=1 | "port must be in range 1-65535" | PASS |
| T4d | invalid `server_type` (`weirdtype`) | exit=1 | "server_type must be stdio or remote" | PASS |
| T4e | invalid `protocol` (`ftp`) | exit=1 | "protocol must be one of [http https ws wss]" | PASS |
| T4f | invalid `bind_address` (`not_an_ip`) | exit=1 | "invalid bind_address" | PASS |
| T4g | malformed JSON | exit=1 | "invalid JSON in ..." | PASS |
| T4h | non-existent file | exit=1 | "failed to read ..." | PASS |
| T4i | invalid `timeout` (`forever`) | exit=1 | "invalid timeout ... time: invalid duration" | PASS |
| T4j | invalid `security_profile` (`insane`) | exit=1 | "security_profile must be development, standard or restricted" | PASS |

### T5 — Real MCP Handshake

- [x] **T5a** `mcp test` — handshake přes stdio → `protocol=2024-11-05, server=Portunix/v2.0.0`
- [x] **T5b** `mcp test --verbose` — navíc ukazuje `claude mcp get portunix` výstup včetně env vars
- [x] **T5c** `mcp test --assistant claude-code` — explicit assistant works
- [⚠️] **T5d** `mcp test --assistant nonexistent-assistant` — projde díky fallbacku na default stdio. **UX nuance**, ne bug. Doporučení: pro neznámý assistant raději hlásit "assistant not configured" a vrátit non-zero.

### T6 — Non-Interactive Configuration

- [x] **T6a** `--assistant claude-code --scope user --env "T6=value,FOO=bar" --timeout 30s --force` — všechno persistované, `claude mcp list` ukazuje "✓ Connected"
- [x] **T6b** Bez `--force` při existující konfiguraci — exit=1, "MCP server already configured. Use --force to reconfigure"

### T7 — Env / Timeout Edge Cases

- [x] **T7a** `--env "NO_EQUALS_SIGN"` — exit=1, "invalid env entry"
- [x] **T7b** `--env "=value_only"` (prázdný key) — exit=1
- [x] **T7c** `--timeout "abc"` — exit=1
- [x] **T7d** `--env ""` (prázdný) — projde (no-op)
- [x] **T7e** `--env "KEY1=val1,KEY2=val2,KEY3=val with spaces,KEY4=val=with=equals"` — všech 4 hodnoty správně parsovány:
  - `KEY3` = `val with spaces` — mezery zachovány
  - `KEY4` = `val=with=equals` — pouze první `=` je separator
- [x] Env vars se **kumulují** přes opakované volání (záměrný merge per-key, ne reset)

### T8 — Claude Desktop Auto-Configure

- [x] **T8a** Vytvoření `mcp_servers.json` z nuly — portunix entry vytvořen, dir `$APPDATA/Claude/` vytvořen
- [x] **T8b** Merge s existujícími entries (`another-server`, `yet-another`) — všechny 3 entries zachovány, portunix přidán
- [x] **T8c** Re-merge (running druhý import) — počet serverů stabilní (3), portunix entry aktualizován
- [⚠️] **T8d** Malformed existující JSON — chyba zobrazena uživateli ("⚠️  claude-desktop: existing ... is not valid JSON"), ALE následuje "✅ Import complete" a exit=0. **Doporučení:** vrátit non-zero exit code, když všechny per-assistant aplikace selžou, nebo aspoň přejmenovat "✅ Import complete" na "Import finished with warnings".

### T9 — End-to-End Claude Registration

- [x] Po `mcp init --from-json` ukazuje `claude mcp list`:
  ```
  portunix: C:\DEV\CassandraGargoyle\portunix\portunix\portunix.exe mcp serve - ✓ Connected
  ```

### T10 — Persistence Schema

- [x] `mcp-server.json` obsahuje všechna nová pole (scope, env, timeout, bind_address, tls_cert, tls_key) s `omitempty`
- [x] JSON struktura validní, čitelná zpět

### T11 — Regression Tests (Existing Commands)

- [x] **T11a** `mcp config` (text) — funguje, ukazuje konfiguraci
- [x] **T11b** `mcp config --json` — funguje, validní JSON
- [x] **T11c** `mcp status` — "🎉 Portunix MCP: OPERATIONAL"
- [x] **T11d** `mcp configure --help` — funguje
- [x] **T11e** `mcp serve --help` — funguje

### T12 — Backward Compatibility

- [x] Načtení starého `mcp-server.json` (bez nových polí) — funguje
- [x] Po save jsou nová pole **stále omitted** v JSON (díky `omitempty`)
- [x] Tester ověřil: starý formát se nezmění bez explicitního přidání nových hodnot

### T13 — Dispatcher Pattern

- [x] `./ptx-mcp.exe mcp --help` — funguje (direct invocation), help text identický s `portunix mcp --help`

### T14 — End-to-End Env Propagation

- [x] `--env "TC3_KEY=tc3_value"` v `--from-json` → `claude mcp get portunix` (přes `mcp test --verbose`) ukazuje `Environment: TC3_KEY=tc3_value`
- [x] Důkaz, že env vars proputují přes `claude mcp add -e KEY=VAL` až do `claude mcp list/get`

## Functional Tests

### Funkční pokrytí

- [x] Bod 1 (issue #034): Real Connection Testing — **implementováno a ověřeno**
- [x] Bod 2: Advanced Claude Code Options (`--scope`, `--env`, `--timeout`, `--from-json`) — **implementováno a ověřeno**
- [x] Bod 3: Auto-configure Claude Desktop — **implementováno a ověřeno** (incl. merge)
- [x] Bod 4: Network Configuration for Remote Servers (validace protokolu, bind address, TLS) — **implementováno a ověřeno** (TLS prompty interactive cestou, validace via `--from-json`)
- [⏸️] Bod 5: Resolve Dependency on #035 — **vědomě vynecháno** (blokováno issuem #035)

### Acceptance criteria (z issue #034)

- [x] `portunix mcp test` provádí reálný JSON-RPC `initialize` handshake — **OK**, vrací `protocol=2024-11-05`
- [x] `portunix mcp init --assistant claude-code --scope project|user|local` — **OK**, propaguje do `claude mcp add --scope`
- [x] `portunix mcp init --assistant claude-code --env KEY=VAL` — **OK**, propaguje do `claude mcp add -e KEY=VAL`
- [x] `portunix mcp init --assistant claude-code --timeout 60s` — **OK**, zapisuje `MCP_TIMEOUT` env před `claude mcp add`
- [x] `portunix mcp init --from-json <file>` — **OK**, importuje validovaný konfigurák bez interaktivních promptů
- [x] `portunix mcp init --assistant claude-desktop` automaticky zapisuje `mcp_servers.json` bez clobberingu — **OK** (na Windows)
- [x] Remote-server wizard nabízí validovaný protocol / bind / TLS — **OK** (přes flagy a `--from-json` validace; interaktivní část neotestována pro chybějící interactive testing harness, ale logika je sdílená)
- [⏸️] After #035 — N/A (vědomě out-of-scope)

## Regression Tests

- [x] Existing functionality unaffected — `mcp config`, `mcp configure`, `mcp status`, `mcp serve --help` všechno funguje
- [x] Cross-platform compatibility — Windows ověřeno; Linux/macOS path resolution v kódu (`getClaudeDesktopConfigPath()` switch on `runtime.GOOS`), nutno otestovat na Linux/macOS testerem
- [x] Backward compatibility se starým `mcp-server.json` ověřena

## Issues Found

### Minor (non-blocking)

1. **T5d — fallback handshake pro neznámý assistant**
   - **Co**: `mcp test --assistant nonexistent-foo` projde, protože `handshakeAssistant` použije fallback na default stdio.
   - **Závažnost**: UX — info pro uživatele zavádějící.
   - **Doporučení**: Hlásit "assistant <X> is not configured" a vrátit non-zero exit kód.

2. **T8d — malformed existující JSON v `mcp_servers.json`**
   - **Co**: Když existující Claude Desktop config je rozbitý JSON, kód vypíše warning, ale return exit=0 a hlásí "✅ Import complete".
   - **Závažnost**: Mid — uživatel může věřit, že vše proběhlo OK, ale Claude Desktop entry chybí.
   - **Doporučení**: `runFromJSONConfiguration` by mělo počítat per-assistant errors a vrátit non-zero exit když ≥1 selže; nebo přejmenovat finální zprávu na "Import finished with warnings".

### Recommendations (nice-to-have, ne blocking)

- Zvážit, zda env vars by se měly **kumulovat** nebo **reset**ovat na každý `mcp init` volání. Aktuální chování je per-key merge (kumulace). Doporučuji přidat `--env-reset` flag pro explicitní reset, pokud uživatel preferuje.
- Funkce `addMCPServerToClaudeCode` (BC wrapper) je `unused` podle linteru — buď ji odstranit, nebo přidat komentář o BC accountability.
- Linux/macOS testing nutné pro plný cross-platform certificate, protože self-signed cert generování a path resolution se liší (path `~/Library/Application Support/Claude/` na macOS, `~/.config/claude/` na Linux).

## Final Decision

**STATUS**: **CONDITIONAL PASS**

**Důvod**: Hlavní funkcionalita (4 ze 4 implementovaných bodů) funguje **end-to-end ověřeno**. Real JSON-RPC handshake prokazatelně vrací správnou protocol verzi. Validace `--from-json` zachycuje 10 různých negative cases. Claude Desktop merge zachovává existující entries. Backward compatibility se starým JSON formátem zachována. Dispatcher pattern není narušen.

**Conditional**: T8d (malformed JSON → exit=0 místo non-zero) **doporučujeme opravit před merge do main**, ale nejde o blocking bug pro core funkcionalitu — uživateli je chyba zobrazena na obrazovce.

**Approval for merge**: **YES** s podmínkou doplnit fix pro T8d nebo akceptovat jako známou minor issue (commit message v poznámkách).

**Acceptance protocol date**: 2026-05-12
**Tester signature**: zdendaku

---

## Untested Areas (pro úplnost)

- **Interaktivní prompty** (`promptClaudeCodeOptions`, `promptRemoteOptions`) — kód sdílí logiku s flagy, ale plný E2E interactive testing vyžaduje terminal harness (mimo scope pro tento protokol).
- **Self-signed cert generování** přes interaktivní wizard — kód jsme reviewed, ale runtime test interaktivně nedělaný.
- **Linux/macOS path resolution** pro `mcp_servers.json` — vyžaduje cross-platform tester (`tester-linux`, `tester-macos`).
- **Remote TCP handshake** přes `performNetworkHandshake` — vyžaduje běžící TCP MCP server, neotestováno (stdio handshake plně ověřeno).
- **`wss`/`ws` protokoly** — kód mapuje na http/https, nikoli skutečný websocket handshake. Pro plné MCP přes WebSocket bude pravděpodobně nutná samostatná implementace.

## Files Tested

```
src/helpers/ptx-mcp/common.go     — schema + validation helpers
src/helpers/ptx-mcp/configure.go  — claude mcp add propagation
src/helpers/ptx-mcp/init.go       — wizard + flags + Claude Desktop merge
src/helpers/ptx-mcp/test.go       — JSON-RPC handshake
```

## Test Commands Used (subset)

```bash
make build
./portunix.exe mcp init --help
./portunix.exe mcp init --from-json <path> --force
./portunix.exe mcp init --assistant claude-code --scope user --env "KEY=VAL" --timeout 30s --force
./portunix.exe mcp test [--assistant <name>] [--verbose]
./portunix.exe mcp config [--json]
./portunix.exe mcp status
claude mcp list
claude mcp get portunix
```
