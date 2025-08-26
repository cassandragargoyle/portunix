# ADR-003: Linux Distribution Version Support Strategy

**Status**: Accepted  
**Date**: 2025-08-23  
**Issue**: SPEC-012-001 Ubuntu Version Support

## Context

PowerShell installation fails on Ubuntu 25.04 due to restrictive version matching in `install-packages.json`. Current configuration only supports explicitly listed versions (20.04, 22.04, 24.04), causing installation failures on newer releases.

**Current Problems:**
- No forward compatibility for new Ubuntu releases  
- No fallback mechanism when variant fails
- Poor error handling without alternative suggestions
- High maintenance burden requiring manual updates

**Business Impact:**
- Poor user experience on newer OS versions
- Increased support burden
- Reduced adoption potential

## Decision

We implement a **Hybrid Forward Compatibility Strategy** combining version ranges with intelligent fallback mechanisms:

### 1. Version Range Support
```json
"ubuntu": {
  "version": "7.4.6",
  "distributions": ["ubuntu", "kubuntu"],
  "supported_version_ranges": [
    {"min": "20.04", "max": "30.99", "type": "lts_and_interim"}
  ],
  "fallback_variants": ["snap"],
  "fallback_strategy": "auto_with_confirmation"
}
```

### 2. Three-Tier Fallback Strategy
```
Ubuntu Variant → Snap Variant → Manual Installation Guide
     ↓               ↓                    ↓
Native repos    Universal         User guidance
  (fast)        (reliable)       (documentation)
```

### 3. Enhanced Error Handling
- Provide specific reasons for version rejection
- Suggest alternative installation methods
- Include links to documentation

### 4. Version Classification System
- **Supported**: Explicitly tested versions
- **Compatible**: Within range, not tested but likely to work  
- **Experimental**: Newer versions, with warnings
- **Unsupported**: Outside range, fallback required

## Implementation Architecture

### Component Changes

**1. Version Matching Logic** (`app/install/version_matcher.go`):
```go
type VersionRange struct {
    Min    string `json:"min"`
    Max    string `json:"max"` 
    Type   string `json:"type"`
}

func (vm *VersionMatcher) IsVersionSupported(version string, ranges []VersionRange) SupportLevel {
    // Returns: Supported, Compatible, Experimental, Unsupported
}
```

**2. Fallback Chain** (`app/install/fallback.go`):
```go
type FallbackStrategy struct {
    Variants []string `json:"fallback_variants"`
    Strategy string   `json:"fallback_strategy"` // auto, confirm, manual
}

func (fb *FallbackManager) ExecuteFallback(pkg Package, reason string) error {
    // Implements cascading fallback with user communication
}
```

**3. Enhanced Error Messages**:
```
❌ Installation FAILED!
Reason: Ubuntu 25.04 not explicitly supported for PowerShell ubuntu variant

✅ Alternative options:
1. Install via Snap (universal, recommended):
   ./portunix install powershell --variant snap
   
2. Use manual installation:
   See: https://docs.microsoft.com/powershell/install-ubuntu

Would you like to try Snap variant automatically? [Y/n]
```

### Configuration Schema Changes

**New Fields in install-packages.json**:
```json
{
  "supported_version_ranges": [
    {
      "min": "20.04",
      "max": "30.99", 
      "type": "lts_and_interim",
      "confidence": "high"
    }
  ],
  "fallback_variants": ["snap"],
  "fallback_strategy": "auto_with_confirmation",
  "version_support_policy": {
    "forward_compatibility": true,
    "testing_requirement": "none_for_interim",
    "maintenance_schedule": "quarterly"
  }
}
```

## Consequences

### ✅ **Positive**
- **Forward compatibility**: New Ubuntu versions work automatically
- **Better UX**: Clear guidance when primary method fails  
- **Reduced maintenance**: Version ranges eliminate constant updates
- **Reliability**: Snap fallback provides universal solution
- **Transparency**: Users understand what's happening

### ⚠️ **Trade-offs**
- **Complexity**: More sophisticated version matching logic
- **Testing burden**: Need to validate range boundaries
- **Configuration size**: Schema becomes more complex

### 🔄 **Migration Plan**
1. **Phase 1**: Implement version range parsing (backward compatible)
2. **Phase 2**: Add fallback mechanism 
3. **Phase 3**: Update PowerShell configuration
4. **Phase 4**: Apply pattern to other packages

## Implementation Sequence Diagram

```
User Request: install powershell --variant ubuntu
    ↓
┌─────────────────────────────────────────┐
│ 1. Detect OS Version (25.04)            │
├─────────────────────────────────────────┤  
│ 2. Check supported_versions []           │
│    → Not found in explicit list         │
├─────────────────────────────────────────┤
│ 3. Check supported_version_ranges        │
│    → 25.04 within 20.04-30.99 range    │
│    → Classification: COMPATIBLE          │
├─────────────────────────────────────────┤
│ 4. Attempt Ubuntu variant installation  │
│    → Repository setup fails             │
├─────────────────────────────────────────┤
│ 5. Trigger fallback_strategy            │
│    → Show error + alternatives          │
│    → Prompt for snap variant            │
├─────────────────────────────────────────┤
│ 6. User confirms → Install via snap     │
│    → Success + log decision              │
└─────────────────────────────────────────┘
```

## Success Metrics

**Immediate (Post-implementation):**
- Ubuntu 25.04 PowerShell installation success rate: >90%
- Support tickets about "unsupported version": -80%
- Average resolution time for version issues: <2 minutes

**Long-term (6 months):**
- New Ubuntu release (26.04) automatic compatibility: >85%
- Manual configuration updates required: <1 per quarter
- User satisfaction with error messages: >4/5

## Risk Mitigation

**Risk 1**: Version range too broad → Test fails
- **Mitigation**: Conservative initial ranges, gradual expansion

**Risk 2**: Snap variant unavailable
- **Mitigation**: Manual installation documentation as final fallback

**Risk 3**: Microsoft changes repository structure
- **Mitigation**: Version-specific repository URL patterns

## Future Evolution

**Planned Extensions:**
- Apply pattern to other Microsoft packages (VS Code, .NET)  
- Implement predictive compatibility scoring
- Add telemetry for version support analytics
- Create automated testing for new OS releases

**Version Support Policy:**
- **LTS versions**: Full support with testing
- **Interim versions**: Compatible support with warnings
- **Future versions**: Experimental with fallback
- **EOL versions**: Deprecation warnings, eventual removal

---

**Decision Owner**: Software Architect  
**Reviewers**: Development Team, QA Team  
**Implementation Target**: Sprint 2025-09