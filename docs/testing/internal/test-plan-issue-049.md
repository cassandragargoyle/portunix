# 🧪 Testovací plán pro Issue #049: Universal Virtualization Support

## Test Plan

**Issue**: #049 Universal Virtualization Support with QEMU/KVM and VirtualBox
**Feature Branch**: `feature/issue-049-qemu-full-support-implementation`
**Tester Role**: QA/Test Engineer
**Testing Environment**: Container-based isolation

### Fáze 1: Základní instalace a konfigurace
**Cíl**: Ověřit správnou instalaci virtualizačních backend-ů

#### TC001: Installation Package Configuration
**Given**: Čistý kontejner Ubuntu 22.04
**When**: Spustím `portunix install virt`
**Then**: Systém by měl automaticky vybrat a nainstalovat QEMU/KVM

#### TC002: Backend Auto-selection Linux
**Given**: Linux environment
**When**: Spustím `portunix virt --help`
**Then**: Měl by se zobrazit help pro univerzální virt command

#### TC003: Configuration File Loading
**Given**: Existující ~/.portunix/config.yaml
**When**: Načtu konfiguraci backend-u
**Then**: Systém by měl správně rozpoznat nastavený backend

### Fáze 2: Universal Command Interface
**Cíl**: Otestovat univerzální `virt` commands

#### TC004: VM Creation Universal Interface
**Given**: Nainstalovaný virtualizační backend
**When**: Spustím `portunix virt create test-vm --template ubuntu-24.04`
**Then**: VM by mělo být vytvořeno pomocí aktivního backend-u

#### TC005: VM Lifecycle Commands
**Given**: Existující VM 'test-vm'
**When**: Provedu lifecycle operace (start, stop, restart, suspend, resume)
**Then**: Všechny operace by měly fungovat konzistentně

#### TC006: VM State Detection and Smart Operations
**Given**: VM v různých stavech
**When**: Spustím command na VM v neočekávaném stavu
**Then**: Systém by měl inteligentně reagovat (např. "already running")

### Fáze 3: ISO Management System
**Cíl**: Otestovat ISO download a management funkce

#### TC007: ISO Download with Dry-run
**Given**: Internet connectivity
**When**: Spustím `portunix virt iso download ubuntu-24.04 --dry-run`
**Then**: Měl by se zobrazit preview bez skutečného stahování

#### TC008: ISO Download and Verification
**Given**: Sufficient disk space
**When**: Spustím `portunix virt iso download ubuntu-24.04`
**Then**: ISO by mělo být staženo a ověřeno checksumem

#### TC009: ISO List and Management
**Given**: Stažené ISO soubory
**When**: Spustím `portunix virt iso list`
**Then**: Měl by se zobrazit seznam s detaily

### Fáze 4: Template System
**Cíl**: Otestovat VM template systém

#### TC010: Template-based VM Creation
**Given**: Dostupné templates
**When**: Vytvořím VM s template `ubuntu-24.04`
**Then**: VM by mělo mít správné default nastavení z template

#### TC011: Template Override Parameters
**Given**: Template s defaults
**When**: Vytvořím VM s vlastními parametry (RAM, CPU)
**Then**: Vlastní parametry by měly override template defaults

### Fáze 5: SSH Integration
**Cíl**: Otestovat SSH funkce a smart connectivity

#### TC012: SSH with Boot Waiting
**Given**: Zastavené VM
**When**: Spustím `portunix virt ssh test-vm --start`
**Then**: VM by se mělo spustit a čekat na SSH dostupnost

#### TC013: SSH Ready Check
**Given**: Běžící VM
**When**: Spustím `portunix virt ssh test-vm --check`
**Then**: Měl by vrátit exit code 0 pokud je SSH ready

#### TC014: File Copy Operations
**Given**: VM s SSH přístupem
**When**: Spustím `portunix virt copy test-vm:/etc/hosts ./hosts`
**Then**: Soubor by měl být zkopírován z VM

### Fáze 6: Snapshot Management
**Cíl**: Otestovat snapshot systém

#### TC015: Snapshot Creation
**Given**: Běžící VM
**When**: Spustím `portunix virt snapshot create test-vm backup`
**Then**: Snapshot by měl být vytvořen

#### TC016: Snapshot Revert
**Given**: VM se změnami po snapshot
**When**: Spustím `portunix virt snapshot revert test-vm backup`
**Then**: VM by se mělo vrátit do stavu snapshot

### Fáze 7: Error Handling & Edge Cases
**Cíl**: Otestovat error handling a edge cases

#### TC017: Non-existent VM Operations
**Given**: Neexistující VM
**When**: Spustím operaci na neexistující VM
**Then**: Měla by se zobrazit jasná error zpráva

#### TC018: Invalid Templates
**Given**: Neexistující template
**When**: Pokusím se vytvořit VM s neplatným template
**Then**: Měla by se zobrazit error zpráva s dostupnými templates

#### TC019: Resource Validation
**Given**: Neplatné resource hodnoty (RAM, disk)
**When**: Pokusím se vytvořit VM s neplatnými hodnotami
**Then**: Systém by měl validovat a zamítnout neplatné hodnoty

## Test Cases (Given/When/Then)

### Critical Path Test Cases

#### TC001: Universal Installation
```gherkin
Given: Clean Ubuntu 22.04 container environment
  And: No virtualization software installed
When: I run "portunix install virt"
Then: System should detect Linux platform
  And: Auto-redirect to QEMU installation
  And: Install QEMU/KVM packages successfully
  And: Configure libvirt for current user
  And: Verify installation with "portunix virt check"
```

#### TC002: VM Lifecycle Complete Flow
```gherkin
Given: Successfully installed virtualization backend
  And: Downloaded Ubuntu 24.04 ISO
When: I create VM with "portunix virt create ubuntu-test --template ubuntu-24.04 --ram 4G"
  And: Start VM with "portunix virt start ubuntu-test"
  And: Wait for boot completion
  And: Create snapshot "portunix virt snapshot create ubuntu-test fresh-install"
  And: Stop VM with "portunix virt stop ubuntu-test"
  And: Restart VM with "portunix virt start ubuntu-test"
Then: All operations should complete successfully
  And: VM state should be correctly tracked
  And: Snapshot should be available for revert
```

#### TC003: SSH Smart Connection
```gherkin
Given: Created VM with SSH enabled
  And: VM is currently stopped
When: I run "portunix virt ssh ubuntu-test --start --wait-timeout 60s"
Then: System should start the VM automatically
  And: Wait for VM boot completion
  And: Wait for SSH service availability
  And: Establish SSH connection successfully
  And: Display "✅ Connected to ubuntu-test"
```

#### TC004: ISO Management Complete Workflow
```gherkin
Given: Clean system with no ISOs downloaded
When: I run "portunix virt iso download ubuntu-24.04 --dry-run"
Then: System should display download preview without downloading
When: I run "portunix virt iso download ubuntu-24.04"
Then: ISO should be downloaded to .cache/isos/
  And: Checksum should be verified
  And: ISO should appear in "portunix virt iso list"
When: I create VM using downloaded ISO
Then: VM creation should use local ISO file
```

#### TC005: Backend Selection Logic
```gherkin
Given: Linux environment with QEMU available
When: I run "portunix virt list"
Then: Command should use QEMU backend automatically
  And: Display "Backend: QEMU/KVM" in output
Given: Configuration file with virtualization_backend: "qemu"
When: I run any virt command
Then: System should respect configuration setting
  And: Use specified backend regardless of auto-detection
```

## Coverage

### Unit/Integration/E2E Test Matrix

#### Unit Tests (app/virt/ package)
| Component | Test Coverage | Critical Functions |
|-----------|---------------|-------------------|
| Config Loading | 95% | LoadConfig(), selectBackend() |
| Backend Interface | 90% | Manager creation, method delegation |
| Template System | 85% | Template loading, validation, application |
| State Management | 95% | State detection, smart operations |
| SSH Operations | 80% | Connection checking, timeout handling |

#### Integration Tests (cmd/ package)
| Command | Test Coverage | Scenarios |
|---------|---------------|-----------|
| virt create | 90% | Template usage, parameter validation |
| virt lifecycle | 95% | start/stop/restart in various states |
| virt ssh | 85% | Smart connection, timeout handling |
| virt snapshot | 80% | Create/revert/delete operations |
| virt iso | 75% | Download, verify, list operations |

#### E2E Tests (Full Workflow)
| Scenario | Duration | Environment |
|----------|----------|-------------|
| Complete VM Lifecycle | 15 min | Container isolation |
| Cross-backend Testing | 20 min | Multiple containers |
| ISO Download & Usage | 10 min | Network-enabled container |
| SSH Integration Flow | 12 min | Container with VM |
| Snapshot/Revert Flow | 8 min | Container with VM |

### Coverage Targets
- **Unit Tests**: 85%
- **Integration Tests**: 80%
- **Command Coverage**: 90%
- **Error Path Coverage**: 75%

## CI Notes

### Container-based Testing Requirements
```bash
# Required base images for testing
docker pull ubuntu:22.04
docker pull debian:bookworm
docker pull fedora:latest

# Test execution pattern
portunix docker run ubuntu    # Create test environment
# Copy portunix binary to container
# Run installation and VM tests inside container
# Verify no host contamination
```

### Performance Benchmarks
| Operation | Target Time | Failure Threshold |
|-----------|-------------|-------------------|
| VM Creation | < 60s | > 120s |
| VM Start | < 30s | > 60s |
| SSH Connection | < 15s | > 30s |
| Snapshot Creation | < 10s | > 20s |
| ISO Download (1GB) | < 300s | > 600s |

### Automated Testing Pipeline
1. **Setup Phase**: Create fresh containers for each test
2. **Installation Phase**: Test `portunix install virt` in isolation
3. **Functional Phase**: Run VM lifecycle tests
4. **Integration Phase**: Test SSH and snapshot features
5. **Cleanup Phase**: Verify no host system changes

### Failure Injection Ideas

#### Network Failures
- **Scenario**: ISO download interrupted
- **Injection**: Simulate network timeout during download
- **Expected**: Graceful failure with retry suggestion

#### Resource Constraints
- **Scenario**: Insufficient disk space for VM creation
- **Injection**: Fill container disk to 95% capacity
- **Expected**: Clear error message about disk space

#### Backend Failures
- **Scenario**: QEMU/libvirt not available
- **Injection**: Rename qemu-system-x86_64 binary
- **Expected**: Fallback to error message with installation instructions

#### SSH Connectivity Issues
- **Scenario**: VM boots but SSH not available
- **Injection**: Block port 22 in VM
- **Expected**: Timeout with clear error message

#### Invalid Configuration
- **Scenario**: Corrupted config file
- **Injection**: Create malformed ~/.portunix/config.yaml
- **Expected**: Default backend selection with warning

### Container Requirements
- **Standard container for QEMU testing**: `portunix docker run ubuntu`
- **Host System Protection**: ❌ NEVER install QEMU/VirtualBox on host during testing
- **Container Isolation**: ✅ ALWAYS use container isolation
- **Verification**: ✅ VERIFY no host system changes after tests
- **Cleanup**: ✅ CLEANUP containers after test completion

---

**Created**: 2025-09-17
**Version**: 1.0
**Author**: QA/Test Engineer
**Framework**: testframework package with verbose logging
**Methodology**: Container-based isolation testing