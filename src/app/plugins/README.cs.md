# plugins

Balíček `plugins` tvoří jádro pluginového systému Portunixu. Definuje
kontrakt pluginů, načítá a validuje jejich manifesty a komunikuje přes gRPC
s plugin binárkami běžícími v samostatných procesech. Zároveň poskytuje
registry službu, kterou hostující platformy (Synapse, Pack, Agent, …)
dotazují, aby zjistily, které nainstalované pluginy pro ně jsou určené.

CLI vstupní bod: `portunix plugin …` (viz `src/cmd/plugin.go`).

## Architektura

Pluginy jsou **samostatné procesy**, které Portunix spouští a komunikuje s
nimi přes **gRPC na localhostu**. Tím jsou izolované (vlastní paměť, vlastní
runtime — native, Java, nebo Python) a Portunix má přitom jednotné API.

Definovány jsou dvě gRPC služby:

- **`PluginService`** (`proto/plugin.proto`) — implementuje **každý plugin**;
  Portunix je klient. Metody: `Initialize`, `Execute`, `GetInfo`, `Health`,
  `Shutdown`, `ListCommands`.
- **`PluginRegistryService`** (`proto/plugin_registry.proto`) — implementuje
  **Portunix** (obsluhuje helper `ptx-plugin-registry`); klienti jsou
  hostující platformy. Metody: `ListPluginsForPlatform`, `HealthCheck`.
  Zdůvodnění viz issue #175 (DEC-1/DEC-2).

```text
           ┌──────────────┐       ┌────────────────┐
platformy  │   Synapse    │─────▶│ PluginRegistry │──▶ registry.json
 (klienti) │   Pack / …   │       │   Service      │
           └──────────────┘       └────────────────┘
                                         ▲
                                         │ obsluhuje ptx-plugin-registry
                                   ┌─────┴─────┐
                                   │ portunix  │
                                   │  plugin   │
                                   │  manager  │
                                   └─────┬─────┘
                                         │ PluginService (gRPC klient)
                     ┌───────────────────┼───────────────────┐
                     ▼                   ▼                   ▼
               ┌──────────┐       ┌──────────┐        ┌──────────┐
               │ plugin A │       │ plugin B │        │ plugin C │
               │ (native) │       │  (java)  │        │ (python) │
               └──────────┘       └──────────┘        └──────────┘
```

## Struktura

```text
src/app/plugins/
├── types.go            # rozhraní Plugin, PluginConfig, PluginInfo, PluginManifest, …
├── manifest.go         # LoadManifest / SaveManifest / ValidateManifest
├── grpc_client.go      # GRPCPlugin — klientská implementace Plugin přes gRPC
├── prerequisites.go    # kontrola dostupnosti runtime (Java/Python/native)
├── semver.go           # parser SemVer + match rozsahů pro supported_platforms
├── manager/            # lifecycle pluginů a persistence registry
│   ├── manager.go      # Manager: install/enable/start/stop/uninstall, health smyčka
│   └── registry.go     # Registry: registry.json
├── proto/              # gRPC definice a generovaný kód
│   ├── plugin.proto           # PluginService (plugin → ptx)
│   ├── plugin_registry.proto  # PluginRegistryService (ptx → platformy)
│   ├── generated.go
│   └── pluginregistry/        # generovaný kód pro registry službu
└── server/
    └── registry_server.go     # gRPC server pro PluginRegistryService
```

## Klíčové typy

- `Plugin` (rozhraní) — kontrakt, který plní každý plugin:
  `Initialize`, `Start`, `Stop`, `Execute`, `GetInfo`, `Health`, `IsRunning`.
- `PluginConfig` — runtime konfigurace pluginu (cesta k binárce, druh a
  verze runtime, port, oprávnění, env, working dir).
- `PluginInfo` — metadata vrácená z `GetInfo` (příkazy, schopnosti,
  požadovaná oprávnění, podporované OS, režim — `service` nebo `helper`).
- `PluginManifest` — podoba manifestu na disku
  (`plugin-manifest.schema.json v1.1.0` v repu `api/contract`).
- `GRPCPlugin` — výchozí implementace `Plugin`: spustí binárku pluginu,
  připojí se na gRPC port a přeposílá volání rozhraní.
- `manager.Manager` — orchestruje celý životní cyklus a řídí periodickou
  health-check smyčku.
- `manager.Registry` — drží seznam nainstalovaných pluginů v `registry.json`.

## Manifest

Pluginy dodávají JSON manifest, který popisuje identitu, příkazy, podporované
platformy, požadavky na runtime a oprávnění. `LoadManifest` soubor načte a
`ValidateManifest` vynucuje:

- SemVer tvar `version`
- `supported_platforms[].name` odpovídá `^[a-z][a-z0-9_-]*$`
- `supported_platforms[].features[]` odpovídá `^[a-z][a-z0-9._-]*$`

Schémata žijí ve sdíleném contract repu
(`../api/contract/schemas/`).

## Runtime prerekvizity

`prerequisites.go` ověřuje na hostiteli runtime deklarovaný v manifestu
(`native`, `java`, `python`) a kontroluje verzi proti `runtime_version`
(např. `">=21"`). Výsledky jsou dostupné přes `ptx plugin check` ještě před
instalací/spuštěním, takže se předejde neprůhledným selháním při startu.

## Související

- CLI napojení: `src/cmd/plugin.go`
- Platform-query helper: `ptx-plugin-registry` (viz issue #175)
- Schéma manifestu: `api/contract/schemas/plugin-manifest.schema.json`
- Přehled funkcí: `docs/FEATURES_OVERVIEW.md` (§ Plugin System)
