# GitHub Publishing Workflow

Tento systém ti umožňuje oddělit lokální development (Gitea) od publikace na GitHub.

## 🔧 Jednorázové nastavení

```bash
# Spustit setup script
./scripts/github-00-setup.sh
```

## 📋 Enhanced Workflow

### 1. Lokální development

```bash
# Normální development na Gitea
git add .
git commit -m "wip: working on feature X"
git push origin feature-branch
```

### 2. Sync & Publikace na GitHub

```bash
# Když jsi připraven publikovat
./scripts/github-02-sync-publish.sh
```

**Nový enhanced workflow** ti interaktivně projde:

1. 📥 **GitHub Sync** - stáhne aktuální stav z GitHubu
2. 📊 **Analýza změn** - analyzuje lokální změny
3. 🌿 **Vytvoření větve** - s pomocí Claude vytvoří větev s dobrým názvem
4. 📁 **Sync souborů** - zkopíruje soubory z lokálního repo (bez privátních)
5. ✏️ **Commit zpráva** - vytvoří popisný commit
6. 🚀 **Publikace** - pošle větev na GitHub
7. 🧹 **Cleanup** - volitelné vyčištění

### 3. Alternativní rychlý workflow

```bash
# Pro rychlé squash publikace (původní způsob)
./scripts/github-02-quick-publish.sh
```

## 📁 Co se odstraňuje před publikací

Založeno na `portunix-cleanup-public.ps1`:

- `CLAUDE.md`, `GEMINI.md`, `NOTES.md`
- `bin/`, `*.exe`
- `docs/private/`, `config/dev/`
- Build scripty (`.bat`, `.sh`)
- `app/service_lnx.go`, `cmd/login.go`
- Packaging scripty

## 🎛️ Git remotes struktura

```text
origin  -> tvůj-gitea-server (development)
github  -> github.com/cassandragargoyle/portunix (publikace)
```

## 💡 Tips

- **Development**: Commituj často na Gitea, neboj se WIP commitů
- **Release**: Použij script pro čisté GitHub commity
- **Bezpečnost**: Privátní soubory se automaticky odstraní
- **Historie**: GitHub bude mít čistou historii, Gitea zachová vše

## 🔍 Troubleshooting

### GitHub remote neexistuje

```bash
git remote add github https://github.com/cassandragargoyle/portunix.git
```

### Konflikt při push

Script používá `--force-with-lease` pro bezpečnost. Pokud někdo mezitím commitl na GitHub, script se zastaví.

### Chybné privátní soubory

Uprav seznam v `github-publish.sh` v sekci `PRIVATE_FILES`.

## 🚀 Aktuální test

Můžeme otestovat na současných Docker změnách:

```bash
./scripts/github-publish.sh
```
