---
title: Brainstorming — metodika
description: Český překlad metodiky brainstormingových sessionů s AI asistentem — JSON verbatim journal, tok session, artefakty, integrace s use cases, ADR a PFT workflowy. Aplikuj při vedení session přes /brainstorm nebo při tvorbě materiálů v docs/architecture/brainstorming/.
category: methodology
ai_load: scoped
status: active
language: cs
translation_of: BRAINSTORMING.md
created: 2026-05-16
last_updated: 2026-05-16
related:
  - USE-CASE-METHODOLOGY.md
  - MARKDOWN-FRONTMATTER.md
---

# Brainstorming — metodika

> **Překladová poznámka:**
>
> Tento soubor je **český překlad** dokumentu
> [`BRAINSTORMING.md`](BRAINSTORMING.md) (anglický originál — zdroj pravdy).
> Při jakékoli úpravě obsahu je nutné synchronizovat obě verze; pokud se rozcházejí,
> **autoritativní je anglická verze**. Při překladu drž 1:1 strukturu sekcí a odkazů.

## Úvod

Tento dokument definuje, jak probíhají brainstormingové sessiony s AI asistentem
(Claude Code) v rámci ekosystému CassandraGargoyle / Portunix. Cílem je strukturovaně
zachytit syrové nápady, diskuze a rozhodnutí, které dále napájejí existující use case
a PFT workflowy.

## Struktura session

### 1. Aktivace

Brainstormingová session se zahajuje invokací skillu `/brainstorm` s tématem nebo
referencí na existující brainstormingový adresář.

### 2. Načtení kontextu

Při zahájení brainstormingové session MUSÍ asistent načíst relevantní architektonický
kontext, než se zapojí do diskuze:

- **Tento dokument** — `docs/contributing/BRAINSTORMING.md`
- **Use Case Methodology** — `docs/contributing/USE-CASE-METHODOLOGY.md`
- **Architecture Overview** — `docs/architecture/README.md`
- **Component Index** — `docs/architecture/components/README.md`
- **Existující brainstormingové materiály** v adresáři tématu (pokud existují)

Pro architektonicky zaměřené brainstormingy by asistent měl navíc načíst relevantní
specifikace a ADR z `docs/architecture/`.

### 3. Verbatim journal (JSON)

Během každé brainstormingové session asistent vytváří a udržuje **JSON journal soubor**,
který zaznamenává syrové verbatim záznamy konverzace. Tento soubor uchovává neupravený
tok myšlenek tak, jak vznikají.

#### Umístění souboru

```text
docs/architecture/brainstorming/<topic>/journal-<YYYYMMDD>-<seq>.json
```

- `<topic>` — adresář tématu brainstormingu (např. `config-packages`)
- `<YYYYMMDD>` — datum session
- `<seq>` — pořadové číslo, pokud na stejný den proběhne víc sessionů (`01`, `02`, ...)

#### JSON schéma

```json
{
  "session": {
    "id": "BS-<TOPIC>-<YYYYMMDD>-<SEQ>",
    "topic": "<topic name>",
    "date": "<YYYY-MM-DD>",
    "participants": ["<name>"],
    "medium": "claude-code",
    "status": "active | completed | paused",
    "related_docs": ["<path to related document>"]
  },
  "verbatims": [
    {
      "seq": 1,
      "timestamp": "<ISO 8601 or HH:MM>",
      "speaker": "<name or role>",
      "type": "idea | question | decision | concern | reference | clarification",
      "text": "<raw verbatim text — NEVER edited after recording>",
      "tags": ["<tag>"],
      "refs": ["<reference to doc, component, or external source>"]
    }
  ],
  "decisions": [
    {
      "seq": 1,
      "summary": "<short decision summary in English>",
      "rationale": "<why this was decided>",
      "verbatim_refs": [3, 5],
      "status": "proposed | accepted | rejected | deferred"
    }
  ],
  "open_questions": [
    {
      "seq": 1,
      "question": "<open question>",
      "verbatim_refs": [2],
      "status": "open | resolved | deferred"
    }
  ],
  "next_steps": [
    "<action item>"
  ]
}
```

#### Pravidla pro verbatim

1. **Neměnné** — jakmile je verbatim záznam zapsán, NIKDY se neupravuje
2. **Kompletní** — každý podstatný výrok uživatele i asistenta je zaznamenán
3. **Neupravený** — zachovávej původní formulaci včetně jazyka uživatele (čeština, angličtina, mix)
4. **Atribuovaný** — každý záznam identifikuje mluvčího
5. **Typovaný** — každý záznam je klasifikován (`idea`, `question`, `decision`, `concern`, `reference`, `clarification`)

### 4. Tok session

Typická brainstormingová session probíhá takto:

1. **Načti kontext** — asistent přečte relevantní dokumenty a existující materiály
2. **Zkontroluj existující stav** — pokud session navazuje, projdi předchozí journal záznamy
3. **Prozkoumávej** — volná diskuze, nápady, otázky, výzvy
4. **Zaznamenávej** — asistent průběžně zapisuje verbatimy do journalu
5. **Syntetizuj** — periodicky shrnuj rozhodnutí a otevřené otázky
6. **Uzavři** — finalizuj journal, aktualizuj status, uveď další kroky

### 5. Artefakty

Adresář brainstormingového tématu obsahuje:

```text
docs/architecture/brainstorming/<topic>/
├── README.md                              # Přehled tématu a aktuální stav
├── journal-<YYYYMMDD>-<SEQ>.json          # Verbatim journaly session
├── <YYYYMMDD>-<source>-<description>.md   # Externí vstupy (ChatGPT logy, poznámky atd.)
├── UC-<ID>-<slug>.md                      # Use cases vzniklé z brainstormingu
├── UC-<ID>-<slug>-sequence.puml           # PlantUML sequence diagramy pro use cases
├── UC-<ID>-<slug>-use-case.puml           # PlantUML use case diagramy (volitelné)
└── summary.md                             # Konsolidované závěry (až téma dozraje)
```

### 5a. Use cases a diagramy

Když z brainstormingu vznikne use case, vytvoř obojí:

1. **Use case dokument** (`UC-<ID>-<slug>.md`) — podle šablony z
   [USE-CASE-METHODOLOGY.md](USE-CASE-METHODOLOGY.md)
2. **PlantUML sequence diagram** (`UC-<ID>-<slug>-sequence.puml`) — vizualizuje
   základní tok i alternativní / výjimkové toky

Use case dokument MUSÍ odkazovat na své PlantUML diagramy. Nevkládej ASCII sequence
diagramy přímo do use case dokumentu — místo toho použij PlantUML soubory.

Další typy diagramů (use case diagram, activity diagram, component diagram) jsou
volitelné a měly by se vytvářet tam, kde přidají jasnost.

### 6. Integrace s ostatními workflowy

Výstupy brainstormingu napájejí:

| Výstup | Cílový workflow | Nástroj |
| ------- | --------------- | ------- |
| Ověřené nápady | PFT verbatim | `/add-pft-idea` |
| Formalizované požadavky | PFT requirement | `/add-pft-requirement` |
| Definované use cases | Use Case Methodology | `/add-pft-usecase` |
| Architektonická rozhodnutí | ADR | Manuální vytvoření ADR v `docs/adr/` |
| Technické specifikace | Specifications | `docs/architecture/specifications/` |

## Vodítka pro asistenta

### Během brainstormingu

- **Buď aktivním účastníkem** — navrhuj alternativy, zpochybňuj předpoklady, identifikuj mezery
- **Zaznamenávej průběžně** — nečekej se zápisem verbatimu na konec session
- **Zachovávej jazyk** — zaznamenávej slova uživatele v jeho původním jazyce
- **Taguj efektivně** — používej tagy, aby šly verbatimy později vyhledávat
- **Sleduj rozhodnutí** — kdykoli uživatel udělá rozhodnutí, zapiš ho do pole `decisions`
- **Vynášej otevřené otázky** — explicitně vyjmenovávej nevyřešené otázky

### Architektonické zaměření

Vzhledem k tomu, že `portunix-architecture` je primárně architektonický projekt,
brainstormingové sessiony se obvykle týkají:

- návrhu komponent a interakčních vzorů
- API kontraktů a rozhodnutí o protokolech
- volby technologií a trade-offů
- deploymentu a provozních záležitostí
- cross-project integrace (CassandraGargoyle ekosystém)

Asistent by měl do diskuze přinést architektonické myšlení — navrhovat varianty,
identifikovat trade-offy (latence, cena, komplexita) a odkazovat na relevantní vzory
a předchozí práci.

### Co NEDĚLAT

- Neupravuj a "neuhlazuj" zaznamenané verbatimy
- Nevynechávej záznam jen proto, že se něco jeví jako nedůležité
- Nedělej architektonická rozhodnutí bez prezentování alternativ
- Neignoruj existující materiály v adresáři tématu
