---
title: Konvence Markdown Frontmatteru
description: Český překlad konvence YAML frontmatteru pro projektové `.md` soubory. AI asistenti (Claude Code, Gemini) čtou frontmatter jako první pro určení relevance, strategie načítání a routování na správnou metodiku. Použij u každého nového nebo podstatně upraveného `.md` souboru.
category: style-guide
ai_load: eager
status: active
language: cs
translation_of: MARKDOWN-FRONTMATTER.md
created: 2026-05-07
last_updated: 2026-05-07
related:
  - MARKDOWN-STYLE.md
  - README.md
---

# Konvence Markdown Frontmatteru

> **AI asistenti — kdy načítat tento soubor (Claude Code, Gemini):**
>
> **Načti eagerly, když:**
>
> - Vytváříš nový `.md` soubor v tomto repozitáři (musíš vyemitovat platný frontmatter)
> - Upravuješ existující `.md` soubor s frontmatterem (zachovej / aktualizuj pole jako `last_updated`)
> - Doplňuješ frontmatter do legacy `.md` souboru (backfill)
> - Validuješ, že frontmatter dokumentu odpovídá jeho obsahu (`status`, `category`, `ai_load`)
>
> **NENAČÍTEJ pro:**
>
> - Čtení obsahu `.md`, který nesouvisí s jeho frontmatterem (tělo má vlastní pravidla rozsahu)
> - Soubory osvobozené od frontmatteru — kořenový `README.md` repozitáře, soubory `CLAUDE.md`,
>   third-party `.md` (vendored, generované)
> - Single-question Q&A o projektu
>
> **Autoritativní rozsah:** YAML schéma, povolené hodnoty, lifecycle pravidla a procedura
> migrace pro legacy soubory. Formátování těla (nadpisy, tabulky, code blocks) řeší
> [MARKDOWN-STYLE.md](MARKDOWN-STYLE.md).

> **Překladová poznámka:**
>
> Tento soubor je **český překlad** dokumentu
> [`MARKDOWN-FRONTMATTER.md`](MARKDOWN-FRONTMATTER.md) (anglický originál — zdroj
> pravdy). Při jakékoli úpravě obsahu je nutné synchronizovat obě verze; pokud se
> rozcházejí, **autoritativní je anglická verze**. Při překladu drž 1:1 strukturu
> sekcí a odkazů.

## Účel

YAML frontmatter sedí úplně na začátku každého projektového `.md` souboru. Podporovaní
AI asistenti — Claude Code a Gemini — ho čtou jako první, aby se rozhodli:

1. Zda je dokument relevantní pro aktuální úkol (`description`, `category`)
2. Zda ho načíst eagerly, scoped, nebo jen on demand (`ai_load`)
3. Jak souvisí s jinými dokumenty (`translation_of`, `superseded_by`, `related`)

Frontmatter **doplňuje** — NIKOLI nenahrazuje — load-scope blockquote, který následuje
za H1. Frontmatter je strojově čitelný souhrn; blockquote je human-readable detail
s odůvodněním a příklady. Když se obojí rozchází, **vyhrává blockquote**.

## Podporovaní AI asistenti

| Asistent | Surface | Čte frontmatter | Čte load-scope blockquote |
| ------------------ | ------------------ | ---------------- | -------------------------- |
| Claude Code | CLI / IDE (Anthropic) | ano | ano |
| Gemini Code Assist | CLI / IDE (Google) | ano | ano |

Schéma je psané pro tyto dva asistenty. Ostatní agenti (Cursor built-in agent, Aider,
Cody, Continue atd.) ho mohou konzumovat, ale nejsou gating klienti. Při přidávání
nového asistenta připoj řádek a popiš jeho chování v sekci
[Per-assistant overrides](#per-assistant-overrides), pokud se odchyluje od defaultu.

## Schéma

### Povinná pole

| Pole | Typ | Povolené hodnoty | Účel |
| ------------- | ------ | ---------------- | ---- |
| `title` | string | libovolné | Název dokumentu; měl by odpovídat H1 |
| `description` | string | 1-2 věty | Signál relevance pro AI asistenty — veď s *kdy aplikovat*, ne *o čem to je* |
| `category` | enum | `style-guide`, `methodology`, `workflow`, `ai-guidance`, `reference`, `index`, `adr`, `specification` | Hrubá klasifikace |
| `ai_load` | enum | `eager`, `scoped`, `router`, `on-demand`, `never` | Kdy mají AI asistenti tento soubor načíst |
| `status` | enum | `draft`, `active`, `deprecated`, `superseded` | Lifecycle stav |

### Volitelná pole

| Pole | Typ | Účel |
| --------------------- | -------------- | ---- |
| `language` | `en` \| `cs` | Default `en`; **povinné** pro `.cs.md` soubory |
| `translation_of` | relativní path | Pro `.cs.md` soubory ukazuje na kanonický anglický originál — **povinné** při `language: cs` |
| `superseded_by` | relativní path | **Povinné** při `status: superseded` |
| `created` | `YYYY-MM-DD` | Datum vzniku |
| `last_updated` | `YYYY-MM-DD` | Datum poslední podstatné úpravy |
| `applies_to` | seznam globů | Pro metodiky — globy souborů, které metodika spravuje (např. `["**/*.bpmn"]`) |
| `related` | seznam path | Další dokumenty, které je vhodné konzultovat zároveň |
| `tags` | seznam stringů | Volné tagy |
| `ai_load_overrides` | mapa | Per-asistent override `ai_load` (viz [Per-assistant overrides](#per-assistant-overrides)) |

## Hodnoty `ai_load`

| Hodnota | Kdy AI asistent načte | Typická kategorie |
| ------------ | ---------------------- | ----------------- |
| `eager` | Při jakékoli kontextově relevantní práci dotýkající se `.md` souborů (style, naming, frontmatter, terminologie) | `style-guide` |
| `scoped` | Jen když load-scope blockquote dokumentu odpovídá aktuálnímu úkolu | `methodology` |
| `router` | Při navigaci z vyšší otázky; soubor většinou jen odkazuje dál | `index` |
| `on-demand` | Jen pokud je explicitně referencovaný uživatelem, jiným dokumentem nebo zadáním úkolu | `reference`, `adr` |
| `never` | Nikdy auto-loadovat (archivní, internal-only, work-in-progress) | `deprecated`, `draft` |

Když se `ai_load` rozchází s load-scope blockquotem, **vyhrává blockquote** — nese více
nuancí, než dokáže vyjádřit jeden enum.

### Per-assistant overrides

`ai_load` platí uniformně pro všechny podporované asistenty. Ve vzácných případech má
asistent jiné context-window nebo tooling omezení, které ospravedlňují odlišný load
direktiv. Použij volitelnou mapu `ai_load_overrides`:

```yaml
ai_load: scoped
ai_load_overrides:
  claude_code: scoped   # explicitní, totožné s defaultem
  gemini: on-demand     # menší context budget na tomto asistentovi
```

Kanonické klíče asistentů:

| Klíč | Odkazuje na |
| --------------- | ----------- |
| `claude_code` | Anthropic Claude Code (CLI / IDE) |
| `gemini` | Google Gemini Code Assist |

Override používej střídmě. Pokud většina souborů overriduje pro daného asistenta,
oprav základní hodnotu `ai_load` nebo aktualizuj řádek asistenta v
[Podporovaní AI asistenti](#podporovaní-ai-asistenti) — overrides jsou únikový poklop,
ne primární páka.

## Pravidla

### Umístění a formát

- Frontmatter MUSÍ být úplně první obsah souboru (žádné úvodní prázdné řádky, žádný BOM)
- Otevírej a zavírej `---` na vlastním řádku
- YAML 1.2 syntax; double-quote stringy jen když je třeba (úvodní dvojtečka, speciální znaky)
- Jeden prázdný řádek mezi zavírajícím `---` a H1 nadpisem

### Obsah

- Každý projektový `.md` BY MĚL mít frontmatter. **Osvobozené:** kořenový `README.md`
  repozitáře, soubory `CLAUDE.md` (mají jiný load mechanismus), generované dokumenty
  a vendored third-party `.md`.
- Všechna povinná pole MUSÍ být přítomná, i když je `description` krátký
- `description` je nejvíce load-bearing pole — AI asistenti podle něj rozhodují, zda
  číst dál. **Veď use casem** ("Použij když…", "Aplikuj když…"), ne tématem ("O čem X")
- Pro `.cs.md` soubory: `language: cs` A `translation_of: <originál>.md` jsou obě povinné
- Datumy používají ISO `YYYY-MM-DD`; nikdy přirozeně-jazyková data

### Lifecycle

- Nové dokumenty začínají jako `status: active`, nebo `status: draft`, pokud jsou
  explicitně work-in-progress
- Deprecated dokumenty si drží obsah, ale překlopí se na `status: deprecated`
  s deprecation poznámkou na začátku těla
- Superseded dokumenty přidají `superseded_by: <new-file>.md` a redirect poznámku v těle

### Synchronizace překladů

Když upravuješ anglický `.md`, který má `.cs.md` sourozence (nebo naopak), aktualizuj
oba ve stejném PR. Pokud se aktualizuje jen jedna strana, označ druhou
`tags: [out-of-sync]` a otevři issue — nedovol překladům tiše divergovat. Při kolizi
obsahu je autoritativní anglická verze.

### Konvence load-scope blockquote

Blockquote, který následuje za H1, by se měl obracet na AI asistenty obecně:

```markdown
> **AI assistant load scope (Claude Code, Gemini):**
>
> **Load when:** ...
>
> **Do NOT load for:** ...
>
> **Authoritative scope:** ...
```

Pokud se má konkrétní asistent chovat jinak, přidej sub-paragraf uvnitř blockquote
(`> **Gemini-specific:** …`) místo rozdělení vedení souboru přes několik top-level
callout bloků.

## Příklady

### Metodika s explicitním rozsahem

```yaml
---
title: BPMN Methodology
description: Modeling rules, naming conventions, and Camunda 8 integration for BPMN/DMN files. Apply when creating, editing, or reviewing .bpmn / .dmn files.
category: methodology
ai_load: scoped
status: active
language: en
created: 2026-03-15
last_updated: 2026-05-07
applies_to:
  - "**/*.bpmn"
  - "**/*.dmn"
related:
  - PROCESS-ANALYSIS-METHODOLOGY-ICT.cs.md
  - USE-CASE-METHODOLOGY.md
---
```

### Český překlad

```yaml
---
title: Capability Registry — Metodika
description: Český překlad metodiky Capability Registry. Načti při onboardingu nebo úpravě konvencí; pro tvorbu konkrétních YAML souborů použij capability-registry/INDEX.md.
category: methodology
ai_load: scoped
status: active
language: cs
translation_of: CAPABILITY-REGISTRY-METHODOLOGY.md
created: 2026-04-10
last_updated: 2026-05-07
---
```

### Index / router

```yaml
---
title: Contributing
description: Router for contributing guides, methodologies, and conventions. Load when locating which guide applies to a task.
category: index
ai_load: router
status: active
language: en
created: 2026-05-07
last_updated: 2026-05-07
---
```

### Style guide (eager-load)

```yaml
---
title: Markdown Style Guide
description: Markdown formatting rules — headings, tables, code blocks, links, linting. Apply to any .md edit.
category: style-guide
ai_load: eager
status: active
language: en
created: 2026-03-29
last_updated: 2026-05-07
---
```

### Per-assistant override

```yaml
---
title: Long Reference Catalog
description: Catalog of every external system; load when answering integration or vendor questions.
category: reference
ai_load: on-demand
ai_load_overrides:
  gemini: never   # příliš velké pro default Gemini context budget
status: active
language: en
created: 2026-04-15
last_updated: 2026-05-07
---
```

### Deprecated dokument

```yaml
---
title: Old Issue Format
description: Historical issue format used before INT-XXX numbering was introduced. Kept for reference; do not apply to new issues.
category: workflow
ai_load: never
status: deprecated
language: en
created: 2025-08-12
last_updated: 2026-02-20
---
```

### Superseded dokument

```yaml
---
title: Inbound Mail Triage (combined)
description: Original combined inbound-mail process. Superseded by per-channel processes after 2026-05-06 1:1 with J. Khýrová.
category: specification
ai_load: never
status: superseded
superseded_by: ../pft-infinitecare/processes/as-is/correspondence/paper-mail.md
language: cs
created: 2026-04-30
last_updated: 2026-05-06
---
```

## Migrace

Soubory bez frontmatteru ho dostanou při příští netriviální úpravě. Hromadná migrace
není nutná; oportunistický backfill drží churn nízký a review focused.

Postup backfillu:

1. Přečti soubor a urči `category` z umístění a obsahu
2. Urči `ai_load` z existujícího load-scope blockquote (pokud existuje), nebo z účelu souboru:
   - Style guide / konvence → `eager`
   - Metodika s deklarovaným rozsahem → `scoped`
   - Index / hub / table-of-contents → `router`
   - Reference, ADR, archivní → `on-demand` nebo `never`
3. Vytáhni `created` z gitu: `git log --diff-filter=A --follow --format=%ad --date=short -- <path>`
4. Nastav `last_updated` na dnešek
5. Přidej frontmatter; **nedotýkej se ostatního obsahu** ve stejném commitu (jeden topic per PR)

Pokud legacy soubor má v load-scope blockquote jen jednoho asistenta (např. "Claude Code —
when to load…"), generalizuj ho na "AI assistant load scope (Claude Code, Gemini)" ve
stejném PR — viz [Konvence load-scope blockquote](#konvence-load-scope-blockquote).

## Verzování schématu

Změny schématu probíhají v tomto dokumentu. Pravidla:

- **Aditivní změna** (nové volitelné pole, nový asistent klíč) — aktualizuj tento
  dokument; existující soubory zůstávají platné
- **Nové povinné pole** — aktualizuj tento dokument, přidej migration poznámku v této
  sekci, dolaď všechny soubory v jednom nebo několika následných PR
- **Přejmenované nebo odstraněné pole** — bumpuj `last_updated` zde a zmigruj všechny
  soubory v jediném PR (žádný mezistav, kdy jsou některé soubory platné a jiné ne)

Při aktualizaci schématu znovu zkontroluj [README.md](README.md) a
[MARKDOWN-STYLE.md](MARKDOWN-STYLE.md) na cross-reference, které mohou potřebovat následovat.

### Historie schématu

| Datum | Změna | Migrace |
| ---------- | ----- | ------- |
| 2026-05-07 | Initial schema published | — |
| 2026-05-07 | `claude_load` → `ai_load` (generalizace pro Claude Code + Gemini) | Přejmenované pole; povolené hodnoty beze změn. Soubory stále používající `claude_load` mají být migrovány při příštím editu. |

## Reference

- [MARKDOWN-STYLE.md](MARKDOWN-STYLE.md) — pravidla pro tělo (nadpisy, tabulky, code blocks, odkazy, linting)
- [README.md](README.md) — index contributing dokumentů a kategorizace
- [YAML 1.2 specification](https://yaml.org/spec/1.2.2/) — syntax reference

---

**Created**: 2026-05-07
**Last Updated**: 2026-05-07
