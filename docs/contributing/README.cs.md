# Dokumentace pro přispěvatele

> 🌐 **Jazyk / Language**: **Čeština** | [English](README.md)

Tento adresář obsahuje pokyny a dokumentaci pro přispívání do projektu Portunix.

## Vícejazyčná týmová komunikace

### Oficiální jazyk projektu

- Hlavním jazykem projektu je **angličtina**
- Kód, komentáře, dokumentace i commit messages musí být v angličtině
- Používejte jednoduchou a srozumitelnou angličtinu kvůli nerodilým mluvčím
- Vyhněte se idiomům, hovorovým výrazům a kulturně specifickým odkazům

### Doporučené postupy pro komunikaci

- Pokud je to potřeba, používejte AI překladače (DeepL, Google Translate)
- Využívejte asynchronní komunikaci kvůli různým časovým pásmům
- Pěstujte otevřenou komunikaci s pravidelnými kontrolními schůzkami
- Buďte trpěliví a respektujte jazykové i kulturní rozdíly

### Standardy dokumentace

- Veškerá technická dokumentace v angličtině
- Používejte jasný a stručný jazyk s jednoduchou stavbou vět
- Pokud to dává smysl, doplňujte příklady a obrázky
- V celém projektu udržujte konzistentní terminologii

## Přidělování iniciál v týmu a řešení kolizí

Členové týmu mají přidělené unikátní iniciály pro identifikaci. Pokud má více členů
stejné iniciály (např. John Smith = JS), kolize se řeší přidáním dalšího znaku
z příjmení (např. John Smith = JSm, pokud je JS už obsazené).

## Dostupné pokyny

### Obecné

- [AI asistenti](AI-ASSISTANTS.md) — pokyny a osvědčené postupy pro práci s AI asistenty
- [Terminologie](TERMINOLOGY.md) — společné pojmy, zkratky a technické termíny používané v projektu
- [Pravidla pro TODO](TODO-GUIDELINES.md) — standardy pro zápis a údržbu TODO komentářů
- [Doporučené nástroje](TOOLS-RECOMMENDATIONS.md) — doporučené vývojářské nástroje
- [Postup překladu](TRANSLATION-WORKFLOW.md) — překlady dokumentace do mateřských jazyků členů týmu
- [Markdown style](MARKDOWN-STYLE.md) — pravidla pro psaní Markdown dokumentace

### Správa issues a projektu

- [Správa issues](ISSUE-MANAGEMENT.md) — jak se evidují a synchronizují issues mezi Gitea a GitHub
- [Metodika vývoje issue](ISSUE-DEVELOPMENT-METHODOLOGY.md) — povinný vývojový proces s kontrolními body kvality
- [Hlášení chyb](BUG-REPORTING.md) — jak hlásit a dokumentovat chyby
- [Versioning](VERSIONING.md) — sémantické verzování a strategie vydávání

### Git a publikační workflow

- [Git workflow](GIT-WORKFLOW.md) — pojmenování větví, commit messages a PR proces
- [Interní metodika Gitea](GITEA-INTERNAL-METHODOLOGY.md) — interní vývoj na self-hosted Gitea
- [GitHub workflow](GITHUB-WORKFLOW.md) — publikace z lokální Gitea na veřejný GitHub mirror
- [AUR workflow](AUR-WORKFLOW.md) — proces vydávání pro Arch User Repository

### Kódový styl

Konvence a standardy podle programovacího jazyka:

- [Kódový styl Go](CODE-STYLE-GO.md)
- [Kódový styl Java](CODE-STYLE-JAVA.md)
- [Kódový styl C++](CODE-STYLE-CPP.md)
- [Kódový styl Python](CODE-STYLE-PYTHON.md)

### Testování

- [Metodika testování](TESTING_METHODOLOGY.md) — verbose test framework a kontejnerová politika testování
- [Testování — Go](TESTING-GO.md) — vzory unit testů v Go
- [Testování — Go projekt](TESTING-GO-PROJECT.md) — celoprojektové standardy testování v Go
- [Testování — Java](TESTING-JAVA.md)
- [Testování — C++](TESTING-CPP.md)
- [Testování — Python](TESTING-PYTHON.md)

### Kontejnery a virtualizace

- [Podman guide](PODMAN-GUIDE.md) — práce s Podmanem v Portunixu
- [VM/SSH deployment](VM-SSH-DEPLOYMENT.md) — nasazení Portunixu do virtuálních strojů přes SSH

### Dokumentace a release tooling

- [Metodika tvorby manuálů](MANUAL-CREATION-METHODOLOGY.md) — jak se píší projektové manuály
- [Generování PDF](PDF-GENERATION-WORKFLOW.md) — generování PDF dokumentace
- [Duální README systém](README-DUAL-SYSTEM.md) — údržba duálního systému README
- [Vývoj pomocných binárek](HELPER-BINARY-DEVELOPMENT.md) — sestavování doprovodných binárek

### Tutoriály

- [Maven základy pro juniory](tutorials/MAVEN-BASICS-FOR-JUNIORS.md)

## Překlady

České překlady leží vedle anglických originálů s příponou `.cs.md`
(například `README.cs.md`, `TERMINOLOGY.cs.md`). Překlady jsou volitelné —
závazná je vždy anglická verze.

---

**Poznámka**: Portunix je vydáván pod licencí MIT. Pravidla pro přispěvatele se vztahují
na všechny přispěvatele veřejného repozitáře.
