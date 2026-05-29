# Terminologie a zkratky

> **Jazyk / Language**: **Čeština** | [English](TERMINOLOGY.md)

## Účel

Tento dokument definuje běžnou terminologii, zkratky a technické pojmy používané ve všech projektech CassandraGargoyle.
Slouží jako referenční materiál pro zajištění jednotné komunikace a porozumění v rámci vývojového týmu.

## Obecné zkratky

### Projektové řízení

- **ADR** – Architecture Decision Record (záznam architektonického rozhodnutí)
- **API** – Application Programming Interface (programové rozhraní)
- **CI/CD** – Continuous Integration / Continuous Deployment (průběžná integrace a nasazování)
- **CLI** – Command Line Interface (rozhraní příkazové řádky)
- **CRUD** – Create, Read, Update, Delete (vytvoř, čti, aktualizuj, smaž)
- **DDD** – Domain-Driven Design (návrh řízený doménou)
- **DevOps** – Development Operations (spolupráce vývoje a provozu)
- **GUI** – Graphical User Interface (grafické uživatelské rozhraní)
- **IDE** – Integrated Development Environment (integrované vývojové prostředí)
- **MVP** – Minimum Viable Product (minimální životaschopný produkt)
- **POC** – Proof of Concept (důkaz konceptu)
- **QA** – Quality Assurance (zajištění kvality)
- **REST** – Representational State Transfer
- **SDK** – Software Development Kit (sada vývojářských nástrojů)
- **SLA** – Service Level Agreement (dohoda o úrovni služeb)
- **SOP** – Standard Operating Procedure (standardní pracovní postup)
- **UI** – User Interface (uživatelské rozhraní)
- **UX** – User Experience (uživatelská zkušenost)

### Vývoj

- **AAA** – Arrange, Act, Assert (vzor pro psaní testů)
- **AOP** – Aspect-Oriented Programming (aspektově orientované programování)
- **DI** – Dependency Injection (vkládání závislostí)
- **DRY** – Don't Repeat Yourself (neopakuj se)
- **IoC** – Inversion of Control (inverze řízení)
- **KISS** – Keep It Simple, Stupid (drž se jednoduchosti)
- **MVC** – Model-View-Controller
- **OOP** – Object-Oriented Programming (objektově orientované programování)
- **ORM** – Object-Relational Mapping (objektově-relační mapování)
- **SOLID** – Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **TDD** – Test-Driven Development (vývoj řízený testy)
- **YAGNI** – You Aren't Gonna Need It (nebudeš to potřebovat)

### Infrastruktura a DevOps

- **CDN** – Content Delivery Network (síť pro doručování obsahu)
- **DNS** – Domain Name System (systém doménových jmen)
- **HTTP/HTTPS** – HyperText Transfer Protocol (Secure)
- **IaC** – Infrastructure as Code (infrastruktura jako kód)
- **JSON** – JavaScript Object Notation
- **K8s** – Kubernetes
- **LB** – Load Balancer (vyrovnávač zátěže)
- **RBAC** – Role-Based Access Control (řízení přístupu na základě rolí)
- **SSH** – Secure Shell
- **SSL/TLS** – Secure Sockets Layer / Transport Layer Security
- **VM** – Virtual Machine (virtuální stroj)
- **YAML** – YAML Ain't Markup Language

## Pojmy specifické pro CassandraGargoyle

### Názvy projektů

- **Bootstrap Scripts** – centrální inicializační skripty pro nastavení vývojového prostředí
- **Portunix** – multiplatformní nástroj pro správu vývojového prostředí
- **Dream** – [popis projektu bude doplněn]

### Týmové role

- **SD** – Software Developer (vývojář)
- **TL** – Tech Lead (technický vedoucí)
- **AR** – Architect (architekt)
- **TS** – Tester
- **US** – User (perspektiva koncového uživatele)

### Vývojové koncepty

- **CLAUDE.md** – soubor s pokyny pro AI asistenty specifický pro daný projekt
- **MCP** – Model Context Protocol (integrace s Claudem)
- **Self-deployment** – automatické nakopírování a nastavení Portunixu ve virtuálních prostředích
- **VM Environment** – izolované prostředí virtuálního stroje pro testování a vývoj

### Bezpečnost a řízení přístupu (Portunix)

- **RBAC role** – předdefinované sady oprávnění: `admin` (plný přístup), `developer` (vývojová prostředí), `operator` (produkce), `auditor` (jen pro čtení)
- **Permissions** – jemně členěné přístupové kontroly: `playbook:execute`, `env:local`, `env:container`, `env:virt`, `secret:read`, `secret:write`
- **Environment Isolation** – RBAC vynucené oddělení mezi lokálním, kontejnerovým a virtualizačním prostředím
- **Access Request** – RBAC validační požadavek obsahující uživatele, oprávnění, zdroj a kontext prostředí
- **Audit Trail** – kompletní logování všech RBAC rozhodnutí a spuštění playbooků pro účely compliance

## Technické pojmy podle kategorie

### Programovací jazyky

- **Go** – programovací jazyk používaný pro jádro Portunixu
- **PowerShell** – multiplatformní skriptovací jazyk a shell
- **Bash** – Unixový shell a příkazový jazyk
- **C++** – systémový programovací jazyk
- **Java** – objektově orientovaný programovací jazyk
- **Python** – vysokoúrovňový programovací jazyk

### Správci balíčků

- **apt** – Advanced Package Tool (Debian/Ubuntu)
- **chocolatey** – správce balíčků pro Windows
- **homebrew** – správce balíčků pro macOS
- **npm** – Node Package Manager
- **winget** – Windows Package Manager

### Kontejnerizace

- **Docker** – kontejnerová platforma s architekturou založenou na démonu
- **Podman** – Pod Manager, bezdémonový a rootless kontejnerový engine
- **Container** – lehké přenositelné runtime prostředí
- **Image** – šablona pro vytváření kontejnerů
- **Registry** – úložiště kontejnerových obrazů
- **Pod** – skupina jednoho nebo více kontejnerů se sdíleným úložištěm a sítí (koncept Kubernetes)

### Operační systémy

- **Distro** – linuxová distribuce
- **LTS** – Long Term Support (vydání s dlouhodobou podporou Ubuntu/Debian)
- **WSL** – Windows Subsystem for Linux

### Testování

- **E2E** – End-to-End testování
- **Integration Test** – testování interakcí komponent
- **Unit Test** – testování jednotlivých funkcí/metod
- **Mock** – simulovaný objekt pro testování
- **Fixture** – testovací data nebo příprava
- **Coverage** – procento pokrytí kódu testy
- **Suite** – sada souvisejících testů

### Verzování

- **PR** – Pull Request
- **MR** – Merge Request
- **Branch** – samostatná vývojová větev
- **Commit** – uložená změna v repozitáři
- **Tag** – pojmenovaný odkaz na konkrétní commit
- **HEAD** – odkaz na aktuální vrchol větve

### Build a nasazení

- **Artifact** – sestavený výstup (spustitelný soubor, knihovna apod.)
- **Pipeline** – automatizovaný proces sestavení/nasazení
- **Stage** – fáze v nasazovací pipeline
- **Release** – zabalená verze určená k distribuci
- **Rollback** – návrat k předchozí verzi

## Přípony a formáty souborů

### Konfigurace

- **.yaml/.yml** – konfigurační soubory ve formátu YAML
- **.json** – konfigurace nebo data ve formátu JSON
- **.toml** – konfigurační soubory TOML
- **.env** – soubor s proměnnými prostředí
- **.config** – obecný konfigurační soubor

### Dokumentace

- **.md** – dokumentace v Markdown
- **.rst** – dokumentace v reStructuredText
- **.adoc** – dokumentace v AsciiDoc

### Skripty a kód

- **.sh** – shellový skript (Linux/macOS)
- **.ps1** – PowerShell skript
- **.bat/.cmd** – Windows batch soubor
- **.go** – zdrojový soubor Go
- **.py** – zdrojový soubor Python
- **.js** – zdrojový soubor JavaScript

### Build a balíčky

- **Dockerfile** – definice Docker image
- **Makefile** – soubor pro automatizaci sestavení
- **go.mod** – definice Go modulu
- **package.json** – definice Node.js balíčku
- **requirements.txt** – závislosti pro Python

## Proměnné prostředí

### Běžné vzory

- **PATH** – cesta pro vyhledávání spustitelných souborů
- **HOME** – domovský adresář uživatele
- **USER** – aktuální uživatelské jméno
- **PWD** – aktuální pracovní adresář
- **TMPDIR** – dočasný adresář

### Specifické pro projekt

- **PORTUNIX_HOME** – instalační adresář Portunixu
- **PORTUNIX_CONFIG** – cesta ke konfiguračnímu souboru
- **DEBUG** – povolení režimu ladění
- **LOG_LEVEL** – úroveň podrobnosti logování

## Síť a bezpečnost

### Protokoly

- **gRPC** – framework pro vzdálená procedurální volání
- **WebSocket** – plně duplexní komunikační protokol
- **MQTT** – Message Queuing Telemetry Transport
- **TCP/UDP** – Transmission Control / User Datagram Protocol

### Bezpečnost

- **JWT** – JSON Web Token
- **OAuth** – Open Authorization
- **RBAC** – Role-Based Access Control: bezpečnostní model omezující přístup do systému na základě rolí a oprávnění uživatele.
V playbook systému Portunixu řídí, kdo může spouštět, číst, zapisovat či mazat playbooky v různých prostředích (local, container, virt).
- **SAML** – Security Assertion Markup Language
- **2FA/MFA** – Two-Factor / Multi-Factor Authentication (dvou-/vícefaktorové ověřování)
- **PKI** – Public Key Infrastructure (infrastruktura veřejných klíčů)

### Sítě

- **CIDR** – Classless Inter-Domain Routing
- **NAT** – Network Address Translation (překlad síťových adres)
- **VPN** – Virtual Private Network (virtuální privátní síť)
- **DHCP** – Dynamic Host Configuration Protocol

## Databáze a úložiště

### Typy databází

- **RDBMS** – Relational Database Management System (relační databázový systém)
- **NoSQL** – Not Only SQL databáze
- **ACID** – Atomicity, Consistency, Isolation, Durability
- **CAP** – Consistency, Availability, Partition tolerance

### Úložiště

- **BLOB** – Binary Large Object (velký binární objekt)
- **S3** – Simple Storage Service (AWS)
- **CDN** – Content Delivery Network
- **RAID** – Redundant Array of Independent Disks

## Monitoring a logování

### Metriky

- **SLI** – Service Level Indicator
- **SLO** – Service Level Objective
- **KPI** – Key Performance Indicator
- **APM** – Application Performance Monitoring

### Logování

- **ELK** – Elasticsearch, Logstash, Kibana
- **JSON** – strukturovaný formát logů
- **Syslog** – standardní logovací protokol

## Pravidla použití

### Pravidla konzistence

1. **Používejte standardní zkratky** z tohoto dokumentu
2. **Definujte projektově specifické pojmy** v projektové dokumentaci
3. **Nezavádějte nové zkratky** bez konsenzu týmu
4. **Aktualizujte tento dokument**, když se zavádí nová terminologie

### Standardy pro dokumentaci

1. **Pravidlo prvního použití**: napište plný termín a v závorce zkratku
   - Příklad: „Application Programming Interface (API)"
2. **Konzistentní psaní velkých písmen**: dodržujte zavedené vzory
3. **Citlivost na kontext**: zvažte technickou úroveň publika

### Komunikace

1. **Týmové diskuse**: zkratky lze používat volně v rámci týmu
2. **Externí dokumentace**: pojmy definujte pro širší publikum
3. **Komentáře v kódu**: pro srozumitelnost používejte plné termíny
4. **Commit messages**: zkratky jsou pro stručnost přípustné

## Reference a zdroje

### Oficiální dokumentace

- [Go Documentation](https://golang.org/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [GitHub Documentation](https://docs.github.com/)

### Průmyslové standardy

- [RFC standardy](https://www.ietf.org/rfc/)
- [ISO standardy](https://www.iso.org/standards.html)
- [NIST guidelines](https://www.nist.gov/)

### Zdroje CassandraGargoyle

- [Bootstrap Scripts dokumentace](../README.md)
- [Pokyny pro přispěvatele](README.md)
- [Pravidla kódového stylu](CODE-STYLE-GO.md)

---

**Poznámka**: Tento dokument je živý a měl by být aktualizován s tím, jak projekt roste
a vznikají nové pojmy. Členové týmu jsou vyzýváni k navrhování doplnění a upřesnění.

*Vytvořeno: 2025-08-23*
*Poslední aktualizace: 2025-08-23*
