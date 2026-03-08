# Acceptance Protocol - Issue #119

**Issue**: PTX-Ansible Standalone Help and Template Examples System
**Branch**: `feature/119-ptx-ansible-standalone-help`
**Tester**: _________________
**Date**: _________________

## Test Summary

- Total test scenarios: 12
- Passed: ___
- Failed: ___
- Skipped: ___

---

# PART A: Local Tests (no installation)

These tests can be safely run on the development machine - they do not install any software.

**Testing OS**: _________________ (host system)

---

### TC001: Standalone Help - ptx-ansible

**Description**: Verify ptx-ansible standalone help works

**Steps**:

```bash
./ptx-ansible --help
./ptx-ansible -h
```

**Expected Result**:

- Displays comprehensive help with commands list
- Shows: playbook, mcp, secrets, audit, rbac, cicd, enterprise, security, compliance
- Shows version information

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC002: Standalone Help - portunix playbook

**Description**: Verify portunix playbook help includes all subcommands

**Steps**:

```bash
./portunix playbook --help
```

**Expected Result**:

- Shows subcommands: run, validate, check, list, init, template, help
- Shows examples for each command
- Shows environment options (local, container, virt)

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC003: Template List

**Description**: Verify template listing works

**Steps**:

```bash
./portunix playbook template list
```

**Expected Result**:

- Shows `static-docs` template
- Shows description: "Static documentation site generator"
- Shows engines: docsify, docusaurus, hugo

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC004: Template Show

**Description**: Verify template details display

**Steps**:

```bash
./portunix playbook template show static-docs
```

**Expected Result**:

- Shows template name, version, description
- Shows parameters: --engine, --target
- Shows supported OS: linux, windows, darwin
- Shows usage example

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC005: Playbook Init - Hugo

**Description**: Generate playbook from Hugo template

**Steps**:

```bash
cd /tmp
./portunix playbook init test-hugo --template static-docs --engine hugo --target local
cat test-hugo.ptxbook
```

**Expected Result**:

- Creates `test-hugo.ptxbook` file
- Contains valid YAML with apiVersion, kind, metadata, spec
- Contains hugo package with extended variant
- Contains scripts: init, build, serve

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC006: Playbook Init - Docusaurus

**Description**: Generate playbook from Docusaurus template

**Steps**:

```bash
cd /tmp
./portunix playbook init test-docusaurus --template static-docs --engine docusaurus --target local
cat test-docusaurus.ptxbook
```

**Expected Result**:

- Creates `test-docusaurus.ptxbook` file
- Contains nodejs package
- Contains npm-based scripts

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC007: Playbook Validate

**Description**: Verify playbook validation works

**Steps**:

```bash
./portunix playbook validate /tmp/test-hugo.ptxbook
```

**Expected Result**:

- Shows "Playbook validation successful"
- Shows playbook name and description
- Shows package count

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC008: Playbook Run - Dry Run

**Description**: Verify dry-run mode works without RBAC issues

**Steps**:

```bash
./portunix playbook run /tmp/test-hugo.ptxbook --dry-run
```

**Expected Result**:

- Shows "Dry-run mode: Validating playbook"
- Shows package installation preview: "[DRY-RUN] Would install: hugo"
- Shows script execution preview
- Completes with "Dry-run completed successfully"
- NO "access denied" or "User not found" errors

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC010: Dispatcher Routing

**Description**: Verify main portunix correctly dispatches to ptx-ansible

**Steps**:

```bash
# All these should work via dispatcher
./portunix playbook --help
./portunix playbook template list
./portunix playbook validate /tmp/test-hugo.ptxbook
```

**Expected Result**:

- All commands execute without "command not found" errors
- Output matches direct ptx-ansible execution

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC011: RBAC Disabled by Default

**Description**: Verify RBAC doesn't block standalone usage

**Steps**:

```bash
# Should work without any RBAC setup
./portunix playbook run /tmp/test-hugo.ptxbook --dry-run

# Check no RBAC directory was required
ls -la ~/.portunix/rbac/ 2>/dev/null || echo "RBAC dir not required - OK"
```

**Expected Result**:

- Playbook runs without RBAC errors
- No "User not found" or "access denied" messages
- RBAC directory creation is optional

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### TC012: Error Handling

**Description**: Verify proper error messages

**Steps**:

```bash
# Missing template
./portunix playbook init test --template nonexistent

# Missing file
./portunix playbook run /tmp/nonexistent.ptxbook

# Invalid YAML
echo "invalid: yaml: content:" > /tmp/bad.ptxbook
./portunix playbook validate /tmp/bad.ptxbook
```

**Expected Result**:

- Clear error messages for each case
- No panics or stack traces
- Helpful suggestions where applicable

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### Local cleanup

```bash
rm -f /tmp/test-hugo.ptxbook /tmp/test-docusaurus.ptxbook /tmp/bad.ptxbook
```

---

# PART B: VM Tests (with installation)

These tests MUST run in an isolated environment (VM) - they install software.

**Testing OS**: _________________ (VM - Ubuntu/Debian recommended)

---

## VM Setup

### 1. Start file server on host machine

```bash
cd /media/zdenek/DevDisk/DEV/CassandraGargoyle/portunix/portunix
python3 scripts/file-server.py --port 8080
```

Note the displayed IP address: _________________

### 2. Connect to VM

```bash
# Linux VM
portunix virt ssh <vm-name> --start

# Windows VM (use name: win11)
portunix virt ssh win11 --start
```

### 3. Install Portunix in VM

```bash
# In VM - use auto-generated install script
curl -fsSL http://<host-ip>:8080/install-from-server.sh | sudo bash

# Verify installation
portunix --version
```

---

### TC009: Playbook Run - Local Execution (Hugo)

**Description**: Full E2E test - install Hugo and run scripts

**Prerequisites**:

- VM has internet access
- sudo/root access available

**Steps**:

```bash
# In VM
cd /tmp
mkdir hugo-test && cd hugo-test

# Generate playbook
portunix playbook init my-docs --template static-docs --engine hugo --target local

# Run playbook (actual installation)
portunix playbook run my-docs.ptxbook

# Verify Hugo installed
hugo version

# Verify site structure created
ls -la
ls -la public/ 2>/dev/null || echo "Note: public/ created after hugo build"
```

**Expected Result**:

- Hugo package installed successfully
- `hugo version` shows installed version
- Init script creates Hugo site structure (config.toml, content/, themes/, etc.)
- No errors during execution

**Result**: [ ] PASS / [ ] FAIL

**Notes**: _______________

---

### VM Cleanup

```bash
# In VM
rm -rf /tmp/hugo-test
# Optionally remove Hugo if needed
```

---

# Final Decision

## Part A (Local Tests)

**Tests Passed**: ___ / 10
**Status**: [ ] PASS / [ ] FAIL

## Part B (VM Tests)

**Tests Passed**: ___ / 1
**Status**: [ ] PASS / [ ] FAIL

---

## Overall Status

**FINAL STATUS**: [ ] PASS / [ ] FAIL / [ ] CONDITIONAL

**Blocking Issues** (if any):
1. _______________
2. _______________

**Recommendations**:
1. _______________
2. _______________

**Approval for merge**: [ ] YES / [ ] NO

**Date**: _________________

**Tester signature**: _________________

---

## Test Environment Details

**Local Machine**:

- OS: _________________
- Portunix Version: _________________

**VM**:

- OS: _________________
- Portunix Version: _________________
- Host IP: _________________
