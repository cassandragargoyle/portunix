# Quick Test Guide

Rychlý návod pro spuštění testů s novou testovací architekturou.

## 🧪 **Spuštění testů s novou architekturou:**

### **1. Task (moderní, doporučeno):**

```bash
# Instalace Task (jednou)
go install github.com/go-task/task/v3/cmd/task@latest

# Základní příkazy
task test               # Všechny testy
task test-unit          # Rychlé unit testy  
task test-integration   # Integration testy (potřebuje Docker)
task test-coverage      # Testy s coverage reportem
task lint              # Code quality
task setup             # Setup environment
```

### **2. Makefile (alternativa):**

```bash
make test               # Všechny testy
make test-unit          # Rychlé unit testy
make test-integration   # Integration testy (potřebuje Docker)
make test-coverage      # Testy s coverage reportem
```

### **3. Manuální spuštění:**

```bash
# Unit testy (rychlé, bez Docker)
go test -tags=unit ./pkg/docker/...

# Integration testy (potřebuje Docker)
go test -tags=integration ./pkg/docker/...

# Všechny testy
go test ./...

# S verbose výstupem
go test -v ./pkg/docker/...
```

### **4. Coverage analýza:**

```bash
# S Task (doporučeno)
task test-coverage
open coverage.html

# S Make
make test-coverage

# Manuálně:
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### **5. Linting a kvalita:**

```bash
# S Task (doporučeno)
task setup              # Nainstaluje dependencies
task lint              # Spustí linting

# S Make
make deps
make lint

# Manuálně:
go vet ./...
gofmt -s -l .
```

### **6. Pokud Docker není k dispozici:**

```bash
# S Task
task test-unit

# S Make
make test-unit

# Manuálně
go test -tags=unit -v ./pkg/docker/
```

### **7. Debug konkrétního testu:**

```bash
# Konkrétní test s verbose
go test -v -run TestDockerInstall ./pkg/docker/

# S debugging informacemi
go test -v -tags=unit -run TestPackageManagerDetection ./pkg/docker/
```

## ⚠️ **Očekávané chování:**

1. **Unit testy** - Měly by projít i bez Dockeru (používají mocky)
2. **Integration testy** - Potřebují Docker daemon, jinak se skipnou
3. **Linting** - Může najít problémy (což je v pořádku pro začátek)

## 🔧 **Setup před prvním spuštěním:**

### **S Task (doporučeno):**
```bash
# 1. Nainstaluj Task
go install github.com/go-task/task/v3/cmd/task@latest

# 2. Setup všeho jedním příkazem
task setup

# 3. Zkontroluj status
task status

# 4. Spusť testy
task test-unit
```

### **S Make (alternativa):**
```bash
# 1. Nainstaluj testing dependencies
make deps

# 2. Setup test environment  
make dev-setup

# 3. Zkontroluj status
make status

# 4. Spusť testy
make test-unit
```

## 🚀 **Doporučený workflow:**

### Pro rychlý vývoj:
```bash
# S Task
task dev          # Unit testy + lint

# S Make  
make test-unit    # Během psaní kódu
make lint         # Před commitnutím
```

### Pro kompletní validaci:
```bash
# S Task
task full         # Všechny testy + quality

# S Make
make test         # Všechny testy
make ci-test      # Celý CI pipeline
```

### Pro debugging:
```bash
go test -v -tags=unit -run TestSpecificFunction ./pkg/docker/
```

## 📁 **Struktura testů:**

```
pkg/docker/
├── docker.go                     # Implementace
├── interfaces.go                 # Mockable interfaces  
├── docker_test.go                # Unit testy (//go:build unit)
└── docker_integration_test.go    # Integration testy (//go:build integration)
```

## 🎯 **Co testovat:**

- **Unit testy**: Logika bez external dependencies
- **Integration testy**: Reálné Docker operace
- **CLI testy**: Testování command-line interface
- **Performance testy**: Benchmarky pro kritické operace

---

**💡 Tip**: Začni s `task test-unit` (nebo `make test-unit`) pro rychlou zpětnou vazbu, pak přejdi na `task test` pro kompletní validaci.

## 🆕 **Task vs Make:**

**Task** (moderní):
- ✅ Go-friendly syntax (YAML)
- ✅ Built-in dependency management
- ✅ Better error handling
- ✅ Cross-platform
- ✅ Parallel execution

**Make** (klasický):
- ✅ Univerzálně dostupný
- ✅ Známý všem
- ✅ CI/CD integrace
- ✅ Jednoduchý syntax