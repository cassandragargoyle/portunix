# Documentation Scripts

Portunix obsahuje dva skripty pro správu dokumentace:

## 📚 post-release-docs.py

**Účel**: Generování statické dokumentace z příkazů Portunix pomocí Hugo

**Použití**:
```bash
# Generování dokumentace
python3 scripts/post-release-docs.py v1.5.0

# Lokální server pro vývoj
python3 scripts/post-release-docs.py --serve

# Pouze build bez serveru
python3 scripts/post-release-docs.py --build-only
```

**Co dělá**:
- ✅ Kontroluje závislosti (Hugo, Portunix binary)
- ✅ Instaluje Hugo automaticky přes Portunix pokud chybí
- ✅ Vytváří Hugo site strukturu
- ✅ Generuje dokumentaci pro všechny core příkazy
- ✅ Objevuje plugin příkazy (částečně)
- ✅ Vytváří release notes
- ✅ Buildí statické HTML stránky do `docs-site/public/`

**Výstup**: `docs-site/public/` - připraveno k publikování

---

## 🚀 publish-docs-to-github.py

**Účel**: Publikování dokumentace na GitHub Pages

**Použití**:
```bash
# Publikování dokumentace
python3 scripts/publish-docs-to-github.py v1.5.0

# Dry run (test bez pushování)
python3 scripts/publish-docs-to-github.py v1.5.0 --dry-run

# S vlastní commit zprávou
python3 scripts/publish-docs-to-github.py v1.5.0 -m "Update docs for new features"

# Přeskočit kontroly (opatrně!)
python3 scripts/publish-docs-to-github.py v1.5.0 --skip-checks
```

**Co dělá**:
- ✅ Kontroluje GitHub CLI (instaluje automaticky přes Portunix)
- ✅ Ověřuje autentifikaci `gh auth status`
- ✅ Kontroluje git repository a GitHub remote
- ✅ Ověřuje existenci `docs-site/public/` (z post-release-docs.py)
- ✅ Vytváří/aktualizuje `gh-pages` branch
- ✅ Kopíruje dokumentaci a commituje změny
- ✅ Pushuje na GitHub Pages
- ✅ Zobrazuje URL finální dokumentace

**Požadavky**:
- Spuštěný `post-release-docs.py` (musí existovat `docs-site/public/`)
- GitHub CLI autentifikace: `gh auth login`
- Git repository s GitHub remote

---

## 🔄 Kompletní workflow

```bash
# 1. Generování dokumentace
python3 scripts/post-release-docs.py v1.5.0

# 2. Kontrola lokálně (volitelné)
python3 scripts/post-release-docs.py --serve
# Otevří http://localhost:1313

# 3. Publikování na GitHub Pages
python3 scripts/publish-docs-to-github.py v1.5.0
```

---

## ⚙️ Automatické závislosti

Oba skripty automaticky instalují své závislosti přes Portunix:

- **Hugo**: `portunix install hugo`
- **GitHub CLI**: `portunix install github-cli`

### První spuštění:

1. **Build Portunix**: `go build -o .`
2. **Autentifikace GitHub**: `gh auth login`
3. **Spustit workflow** výše

---

## 📂 Struktura souborů

```
docs-site/                 # Hugo site
├── content/               # Markdown content
│   ├── commands/          # Generované dokumentace příkazů
│   │   ├── core/          # Core příkazy
│   │   └── plugins/       # Plugin příkazy
│   ├── guides/            # Manuální guides
│   └── releases/          # Release notes
├── themes/portunix-docs/  # Hugo theme
├── public/                # Buildnuté HTML (gitignored)
└── hugo.toml              # Hugo konfigurace
```

---

## 🔧 Troubleshooting

### GitHub CLI není autentifikován
```bash
gh auth login
# Vyberte: Login with a web browser
```

### Hugo instalace selhala
```bash
portunix install hugo --variant extended
```

### Git remote chybí
```bash
git remote add origin https://github.com/cassandragargoyle/Portunix.git
```

### Dokumentace není builděná
```bash
python3 scripts/post-release-docs.py v1.5.0
```

---

## 📡 GitHub Pages URL

Po úspěšném publikování bude dokumentace dostupná na:
**https://cassandragargoyle.github.io/Portunix/**

GitHub Pages může trvat 2-3 minuty na aktualizaci.