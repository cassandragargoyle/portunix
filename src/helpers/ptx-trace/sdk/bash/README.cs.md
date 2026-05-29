# PTX-TRACE Bash SDK

Lehké shellové funkce nad CLI `portunix trace` pro použití v Bash skriptech,
CI pipeline a operačních nástrojích.

> Anglická verze: [`README.md`](README.md)

## Požadavky

- Bash 4.0+ (používá `[[ ... ]]` a `local`)
- Binárka `portunix` v `PATH` (nebo nastavte `PTX_TRACE_BIN`)

## Instalace

SDK je jediný shellový skript určený k načtení přes `source`:

```bash
source /cesta/k/portunix/src/helpers/ptx-trace/sdk/bash/ptx-trace.sh
```

Pro pohodlí si jej zkopírujte vedle vašich skriptů, nebo umístěte do `PATH`.

## Konfigurace

| Proměnná | Popis | Výchozí hodnota |
| -------- | ----- | --------------- |
| `PTX_TRACE_BIN` | Cesta nebo název portunix binárky | `portunix` |

## Funkce

| Funkce | Volá |
| ------ | ---- |
| `ptx_trace_start <jméno> [flagy...]` | `portunix trace start` |
| `ptx_trace_end [flagy...]` | `portunix trace end` |
| `ptx_trace_event <operace> [flagy...]` | `portunix trace event` |
| `ptx_trace_pipe <operace> [flagy...]` | `portunix trace pipe` |
| `ptx_trace_exec <operace> [flagy...] -- <cmd>` | `portunix trace exec` |
| `ptx_trace_sessions [flagy...]` | `portunix trace sessions` |
| `ptx_trace_view [session] [flagy...]` | `portunix trace view` |
| `ptx_trace_stats [session] [flagy...]` | `portunix trace stats` |
| `ptx_trace_export <kind> [flagy...]` | `portunix trace export <kind>` |
| `ptx_trace_session_id` | vypíše ID aktivní session (jinak nenulový exit) |

Všechny argumenty s flagy jsou předány beze změny — kompletní seznam viz
`portunix trace <cmd> --help`.

## Rychlý start

```bash
#!/usr/bin/env bash
set -euo pipefail
source /cesta/k/sdk/bash/ptx-trace.sh

ptx_trace_start "nightly-import" --tag etl --tag prod

# Protlačí CSV přes recorder a zároveň ho streamuje do další fáze.
cat customers.csv | ptx_trace_pipe "ingest" --format csv > staged.csv

# Spustí externí nástroj s plným tracingem (exit code je propagován).
ptx_trace_exec "load" --capture-stderr -- \
  psql -h db.example.com -d shop -f load.sql

# Zaznamená ad-hoc událost odvozenou ze shellové logiky.
rows="$(wc -l < staged.csv)"
ptx_trace_event "summary" --status success --output "rows=${rows}"

ptx_trace_end --summary
```

## Vzory použití

### Podmíněný tracing

Načtení SDK obalíte ochrannou podmínkou, aby skript fungoval i bez nainstalovaného
portunixu:

```bash
if command -v portunix >/dev/null 2>&1; then
  source /cesta/k/sdk/bash/ptx-trace.sh
else
  ptx_trace_start() { :; }
  ptx_trace_end()   { :; }
  ptx_trace_event() { :; }
  ptx_trace_pipe()  { cat; }
  ptx_trace_exec() {
    shift                                                # zahodí jméno operace
    while [[ $# -gt 0 && "$1" != "--" ]]; do shift; done # zahodí volitelné flagy
    [[ "${1:-}" == "--" ]] && shift                      # zahodí `--`
    "$@"
  }
fi
```

### Ad-hoc session

`ptx_trace_pipe` i `ptx_trace_exec` automaticky vytvoří jednorázovou session,
pokud žádná aktivní neexistuje — můžete je tedy používat ve volně stojících
skriptech bez ručního `start`/`end`:

```bash
some-producer | ptx_trace_pipe "filter" --tag oneshot > out.txt
ptx_trace_exec "build" --tag ci -- make build
```

Stabilní jméno automaticky vytvořené session lze zadat pomocí `--session-name`.

### Selhání a návratové kódy

`ptx_trace_exec` propaguje návratový kód obaleného příkazu, takže lze bezpečně
použít pod `set -e`:

```bash
set -euo pipefail
ptx_trace_exec "deploy" -- ./deploy.sh prod
# skript končí, pokud deploy.sh vrátil nenulový kód
```

### Zachycení stderr při selhání

```bash
ptx_trace_exec "migrate" --capture-stderr -- ./run-migrations.sh
```

Pokud příkaz selže, posledních 4 KB jeho stderr se připojí k trace události
do `context.stderr_tail`.

## Související

- [Hlavní README PTX-TRACE](../../README.md) — přehled a CLI reference
- [Srovnání SDK](../README.md) — Python / Java / TypeScript / Go SDK
- Příklad: [`../examples/bash/etl-pipeline.sh`](../examples/bash/etl-pipeline.sh)
