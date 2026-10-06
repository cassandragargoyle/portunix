# Issue #54: GUID Generation Module for Portunix Core

## Status
📋 Open

## Priority
Medium

## Type
Enhancement

## Labels
enhancement, core, utilities, cli

## Summary
Implement a Go module in Portunix core that provides GUID generation functionality with two modes: random GUID generation and deterministic GUID generation based on two string values.

## Description

### Problem Statement
Portunix core needs a built-in GUID generation capability for various internal operations and as a utility for users. The requirement is for two distinct generation modes:

1. **Random GUID Generation**: Standard UUID v4 generation for unique identifiers
2. **Deterministic GUID Generation**: Generate consistent GUIDs based on two input string values (same inputs always produce the same GUID)

### Requirements

#### Functional Requirements

1. **Go Module Implementation**
   - Create new Go module/package within Portunix core
   - Implement UUID v4 generation for random GUIDs
   - Implement deterministic GUID generation using hash-based approach
   - Ensure thread-safety and performance

2. **CLI Integration**
   - Add `portunix guid` command with subcommands
   - Support for random GUID generation: `portunix guid random`
   - Support for deterministic generation: `portunix guid from <string1> <string2>`
   - Output in standard UUID format (8-4-4-4-12)

3. **Help System Integration**
   - Command help visible only in expert mode (`--help-expert`)
   - Command help visible for AI assistants
   - NOT visible in standard help output (`--help`)

#### Technical Requirements

1. **Algorithm Specifications**
   - Random GUIDs: Use crypto/rand for secure random generation
   - Deterministic GUIDs: Use SHA-256 hash of concatenated input strings
   - Follow RFC 4122 UUID format
   - Handle edge cases (empty strings, special characters)

2. **API Design**
   ```go
   package guid

   // GenerateRandom generates a random UUID v4
   func GenerateRandom() (string, error)

   // GenerateFromStrings generates deterministic UUID from two strings
   func GenerateFromStrings(str1, str2 string) (string, error)

   // Validate checks if string is valid UUID format
   func Validate(uuid string) bool
   ```

3. **CLI Command Structure**
   ```
   portunix guid random
   portunix guid from <string1> <string2>
   portunix guid validate <uuid-string>
   ```

#### Non-Functional Requirements

1. **Performance**: Sub-millisecond generation time
2. **Security**: Use cryptographically secure random number generation
3. **Consistency**: Deterministic function must always return same result for same inputs
4. **Standards Compliance**: Follow RFC 4122 UUID standards

### Use Cases

#### Use Case 1: Random GUID Generation
```bash
$ portunix guid random
550e8400-e29b-41d4-a716-446655440000
```

#### Use Case 2: Deterministic GUID Generation
```bash
$ portunix guid from "project-name" "environment-prod"
3f2504e0-4f89-11d3-9a0c-0305e82c3301

$ portunix guid from "project-name" "environment-prod"
3f2504e0-4f89-11d3-9a0c-0305e82c3301  # Same result every time
```

#### Use Case 3: GUID Validation
```bash
$ portunix guid validate "550e8400-e29b-41d4-a716-446655440000"
Valid UUID

$ portunix guid validate "invalid-uuid"
Invalid UUID format
```

### Implementation Plan

#### Phase 1: Core Module Implementation
1. Create `app/guid/` package
2. Implement random GUID generation
3. Implement deterministic GUID generation
4. Add comprehensive unit tests
5. Add benchmarks for performance validation

#### Phase 2: CLI Integration
1. Add `cmd/guid.go` command file
2. Implement subcommands (random, from, validate)
3. Add command parsing and validation
4. Integrate with help system (expert/AI only)

#### Phase 3: Testing & Documentation
1. Integration tests for CLI commands
2. Cross-platform testing (Windows/Linux)
3. Performance benchmarks
4. API documentation
5. Usage examples

### Architecture Considerations

#### Package Structure
```
app/
├── guid/
│   ├── generator.go     # Core GUID generation logic
│   ├── generator_test.go # Unit tests
│   ├── benchmark_test.go # Performance tests
│   └── validator.go     # UUID validation utilities

cmd/
├── guid.go             # CLI command implementation

parser/
├── guid_parser.go      # Command parsing logic
```

#### Dependencies
- `crypto/rand` - Secure random number generation
- `crypto/sha256` - Hash generation for deterministic mode
- `encoding/hex` - UUID formatting
- `regexp` - UUID validation

#### Error Handling
- Graceful handling of invalid inputs
- Proper error messages for CLI users
- Logging integration for debugging

### Testing Strategy

#### Unit Tests
- Test random GUID generation (format, uniqueness)
- Test deterministic generation (consistency, collision resistance)
- Test edge cases (empty strings, special characters)
- Test validation function accuracy

#### Integration Tests
- CLI command execution tests
- Cross-platform compatibility tests
- Performance benchmarks
- Help system integration tests

#### Container-Based Testing
- Test GUID generation in different container environments
- Verify consistency across platforms
- Test CLI commands in isolated environments

### Acceptance Criteria

1. ✅ Random GUID generation produces valid UUID v4 format
2. ✅ Deterministic generation produces consistent results for same inputs
3. ✅ CLI commands work correctly with proper error handling
4. ✅ Help integration works (expert/AI only, not standard help)
5. ✅ Cross-platform compatibility (Windows/Linux)
6. ✅ Performance meets sub-millisecond requirement
7. ✅ Full test coverage (unit + integration)
8. ✅ Documentation and examples provided

### Risk Assessment

#### Low Risk
- GUID generation is well-established technology
- No external dependencies required
- Limited scope and complexity

#### Potential Issues
- UUID collision probability (extremely low but theoretically possible)
- Performance on resource-constrained systems
- CLI integration complexity with help system

#### Mitigation Strategies
- Use crypto/rand for maximum randomness
- Implement comprehensive testing
- Follow existing CLI patterns for help integration

## Related Issues
- #050: Multi-Level Help System (help integration dependency)

## Implementation Notes
- Integrate with existing help system architecture from Issue #050
- Follow existing code patterns in `app/` and `cmd/` directories
- Ensure consistent error handling patterns with other Portunix commands
- Consider future extension for other UUID versions (v1, v3, v5)

---

**Created**: 2025-09-22
**Last Updated**: 2025-09-22
**Assigned**: Developer
**Epic**: Core Utilities Enhancement