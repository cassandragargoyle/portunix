# 🏗️ **Specification Issue Report - Issue #012 PowerShell Linux Installation**

---

## **Issue Summary**
PowerShell installation fails on Ubuntu 25.04 due to incomplete version support specification in `install-packages.json`, revealing gaps in forward compatibility strategy for new Ubuntu releases.

---

## **Issue Details**

| **Field** | **Value** |
|-----------|-----------|
| **Issue ID** | SPEC-012-001 |
| **Reporter** | QA Tester |
| **Date** | 2025-08-23 |
| **Type** | Specification Gap |
| **Severity** | Medium |
| **Priority** | High |
| **Component** | Package Configuration |
| **Affected Feature** | Issue #012 - PowerShell Linux Installation |
| **Files Affected** | `assets/install-packages.json` |

---

## **Problem Statement**

### **Current Behavior:**
```bash
./portunix install powershell --variant ubuntu --dry-run
# Result: ❌ Installation FAILED!
# Error: this variant does not support ubuntu version 25.04
```

### **Root Cause:**
In `install-packages.json`, PowerShell ubuntu variant specifies:
```json
"supported_versions": ["20.04", "22.04", "24.04"]
```

**But Ubuntu 25.04 exists and should be supportable.**

---

## **Specification Gaps Identified**

### **1. Forward Compatibility Strategy Missing**
- **Problem**: No strategy for new Ubuntu versions (25.04, 26.04, etc.)
- **Impact**: Each new Ubuntu release breaks PowerShell installation
- **Question**: Should new versions be auto-supported or explicitly added?

### **2. Fallback Mechanism Undefined**
- **Problem**: No fallback when specific variant fails
- **Impact**: Installation completely fails instead of trying alternatives
- **Question**: Should unsupported versions fallback to `snap` variant?

### **3. Version Matching Logic Unclear**
- **Problem**: Exact version matching is too restrictive
- **Impact**: Minor version differences (25.04 vs 24.04) cause failures
- **Question**: Should we support version ranges or family matching?

### **4. Error Handling Specification Incomplete**
- **Problem**: Error message doesn't suggest alternatives
- **Impact**: Poor user experience - no guidance on next steps
- **Question**: Should error include suggested fallback commands?

---

## **Business Impact**
- **User Experience**: Poor - users on newer Ubuntu versions cannot install PowerShell
- **Maintenance**: High - requires manual updates for each Ubuntu release
- **Support Burden**: Increases - users will report "broken" functionality
- **Adoption**: Reduced - users may abandon portunix if it doesn't work on latest OS

---

## **Proposed Architecture Solutions**

### **Option 1: Version Range Support**
```json
"ubuntu": {
  "version": "7.4.6",
  "distributions": ["ubuntu", "kubuntu"],
  "supported_version_ranges": [
    {"min": "20.04", "max": "25.04"},
    {"min": "26.04", "max": "*"}  // future versions
  ]
}
```

### **Option 2: Automatic Fallback Chain**
```json
"ubuntu": {
  "version": "7.4.6",
  "supported_versions": ["20.04", "22.04", "24.04"],
  "fallback_variants": ["snap"],
  "fallback_strategy": "auto"
}
```

### **Option 3: Version Family Matching**
```json
"ubuntu": {
  "version": "7.4.6", 
  "version_family": "ubuntu-lts", // supports all LTS versions
  "version_compatibility": "forward" // newer versions allowed
}
```

### **Option 4: Smart Auto-Detection**
- Try ubuntu variant first
- If unsupported version detected → auto-switch to snap
- Log decision and continue installation

---

## **Recommended Solution (Architect Decision Required)**

### **Immediate Fix:**
1. **Add Ubuntu 25.04** to supported_versions
2. **Implement fallback** to snap when ubuntu variant unsupported
3. **Improve error message** with suggested alternatives

### **Long-term Strategy:**
Define **version support policy**:
- How to handle new Ubuntu releases?
- When to deprecate old versions?
- Fallback chain priority order?
- Testing strategy for new versions?

---

## **Test Cases Needed (Post-Fix)**
- [ ] Ubuntu 20.04, 22.04, 24.04, 25.04 installation
- [ ] Fallback mechanism when unsupported version detected  
- [ ] Error messages provide helpful guidance
- [ ] Forward compatibility with hypothetical Ubuntu 26.04
- [ ] Behavior on non-Ubuntu systems (should not use ubuntu variant)

---

## **Questions for Architect**

1. **Version Support Policy**: How should we handle new Ubuntu releases going forward?

2. **Fallback Strategy**: Should unsupported versions automatically try snap/elementary variants?

3. **Maintenance Strategy**: Who updates supported_versions when new Ubuntu releases?

4. **Testing Strategy**: How do we test forward compatibility without future OS versions?

5. **User Experience**: What's the acceptable behavior when user's OS version isn't explicitly supported?

---

## **Specification Requirements (for ADR)**

This issue requires an **Architecture Decision Record** addressing:
- Forward compatibility strategy for Linux distributions
- Version matching and fallback policies  
- Error handling and user guidance standards
- Maintenance procedures for OS support matrix

---

**Status**: 🔴 **OPEN - REQUIRES ARCHITECT DECISION**

**Next Steps**: 
1. Architect reviews and creates ADR for version support strategy
2. Development implements chosen approach
3. QA validates solution across multiple Ubuntu versions