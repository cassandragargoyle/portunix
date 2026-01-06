# GitHub Issue #22: Implementácia podpory pre Docusaurus na multiplatformnom prostredí

**Source**: https://github.com/cassandragargoyle/portunix/issues/22
**Author**: Roman-Kazicka
**Created**: 2025-09-27
**State**: OPEN
**Type**: Feature Request
**Priority**: Medium

---

## Popis

Implementovať infraštruktúru pre Docusaurus dokumentačný systém s podporou pre všetky hlavné operačné systémy (Windows, macOS, Linux) podľa OS Agnostic System architektúry.

## Požiadavky

### 1. Multiplatformová podpora
- Windows (obrázok 301)
- macOS (obrázok 302)
- Linux (obrázok 303)

### 2. Dokumentačné systémy na implementáciu

Podľa obrázka implementovať podporu pre:
- **Docusaurus** (primárny cieľ)
- Docsy (sekundárny)
- VitePress (sekundárny)
- MkDocs (sekundárny)

### 3. OS Agnostic System Integration
- Implementovať podľa architektúry OS Agnostic System (obrázok 300)
- Zabezpečiť jednotné správanie naprieč platformami
- Automatická detekcia operačného systému a optimalizácia

## Funkčnosti

- Automatická inštalácia Node.js dependencies
- Inicializácia nového Docusaurus projektu
- Podpora pre rôzne templates
- Integrácia s build systémom
- Hot-reload development server

## Akceptačné kritériá

### 1. Inštalácia funguje na všetkých 3 platformách
- Windows 11
- macOS (Intel + Apple Silicon)
- Linux (Ubuntu, Debian, CentOS/RHEL)

### 2. Kompletná funkcionalita
- `portunix install docusaurus`
- Automatická inštalácia prerequisites (Node.js)
- Vytvorenie nového projektu
- Spustenie dev servera

### 3. Dokumentácia a examples
- Príklady použitia
- Troubleshooting guide
- Platform-specific notes

## Odkazy

- Docusaurus dokumentácia: https://docusaurus.io/docs
- OS Agnostic System: Referencia obrázky 300-303

## Notes

> **Poznámka (2026-01-04):** Obrázky 300-303 zmíněné v issue nebyly dodány. Proběhla komunikace v Teams ohledně upřesnění požadavků.

---

## Related Internal Issues

- **#119**: PTX-Ansible Standalone Help and Template Examples System (current work)
