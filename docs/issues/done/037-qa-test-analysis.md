# QA Test Analysis: Issue #037 - MCP Serve Implementation

## Plan

### Test Strategy Overview
Issue #037 představuje významnou architekturní změnu s potenciálem breaking changes. Test strategy zahrnuje:
- **Regression testing** původní MCP funkcionality
- **Migration path testing** pro přechod z legacy příkazů
- **Multi-mode communication testing** (stdio, TCP, unix socket)
- **Integration testing** s AI assistenty
- **Help system validation** pro CLI UX

### Test Scope
- Core CLI argument parsing behavior
- MCP serve command s různými parametry
- Komunikační módy a jejich funkcionalita
- Zpětná kompatibilita během migrace
- AI assistant integration workflows

## Cases

### TC001: Default Help Display
**Given**: Portunix bez parametrů  
**When**: User spustí `./portunix`  
**Then**: Zobrazí se help text (ne stdio mode)

### TC002: Basic MCP Serve - stdio Mode
**Given**: Nový mcp serve příkaz  
**When**: User spustí `portunix mcp serve`  
**Then**: Application vstoupí do stdio módu pro MCP komunikaci

### TC003: Explicit stdio Mode
**Given**: Explicitní stdio parametr  
**When**: User spustí `portunix mcp serve --mode stdio`  
**Then**: Application vstoupí do stdio módu

### TC004: TCP Mode with Port
**Given**: TCP komunikační mód  
**When**: User spustí `portunix mcp serve --mode tcp --port 8080`  
**Then**: MCP server naslouchá na portu 8080

### TC005: Unix Socket Mode
**Given**: Unix socket mód  
**When**: User spustí `portunix mcp serve --mode unix --socket /tmp/portunix.sock`  
**Then**: MCP server vytvoří unix socket na specifikované cestě

### TC006: Legacy Command Deprecation
**Given**: Starý mcp-server příkaz  
**When**: User spustí `portunix mcp serve`  
**Then**: Zobrazí se error s návoden k novému příkazu nebo deprecation warning

### TC007: MCP Configure Integration
**Given**: MCP configure příkaz  
**When**: User spustí `portunix mcp configure`  
**Then**: Konfigurace používá novou `mcp serve` strukturu

### TC008: Invalid Mode Parameter
**Given**: Neplatný mode parametr  
**When**: User spustí `portunix mcp serve --mode invalid`  
**Then**: Zobrazí se error s platnými možnostmi

### TC009: TCP Mode without Port
**Given**: TCP mód bez portu  
**When**: User spustí `portunix mcp serve --mode tcp`  
**Then**: Použije se default port nebo zobrazí error

### TC010: Socket File Permissions
**Given**: Unix socket s neplatnými permissions  
**When**: User spustí s neplatnou socket cestou  
**Then**: Zobrazí se error s informací o permissions

### TC011: Concurrent MCP Servers
**Given**: Více MCP server instancí  
**When**: User spustí druhý `portunix mcp serve`  
**Then**: Correct error handling pro port conflicts

### TC012: AI Assistant Integration (Claude Code)
**Given**: Claude Code konfigurace  
**When**: Claude Code se pokusí připojit k MCP serveru  
**Then**: Úspěšné připojení přes novou command strukturu

## Coverage

### Unit Test Coverage Targets
- **CLI Argument Parsing**: 100% - kritické pro UX
- **MCP Serve Command**: 95% - všechny parametry a módy
- **Communication Modes**: 90% - stdio, tcp, unix socket
- **Error Handling**: 85% - invalid parameters, conflicts

### Integration Test Matrix
| Component | stdio | TCP | Unix Socket | Legacy Compat |
|-----------|-------|-----|-------------|---------------|
| CLI Args  | ✓     | ✓   | ✓           | ✓             |
| MCP Core  | ✓     | ✓   | ✓           | N/A           |
| AI Tools  | ✓     | ✓   | ✓           | ✓             |
| Configure | ✓     | ✓   | ✓           | ✓             |

### E2E Test Coverage
- **Complete AI Assistant Workflows**: Claude Code, VS Code extension
- **Migration Scenarios**: Legacy -> New command structure
- **Cross-platform Testing**: Windows, Linux behavior consistency
- **Performance Testing**: Latency pro různé communication modes

## CI Notes

### Pipeline Integration
- **Pre-merge**: Unit + Integration tests (Required)
- **Post-merge**: E2E tests včetně AI assistant integration
- **Regression Suite**: Legacy compatibility tests během migrace
- **Performance Benchmarks**: Communication latency testing

### Test Environment Requirements
- Multi-platform runners (Windows, Linux)
- Network testing capabilities pro TCP mode
- File system permissions testing pro unix sockets
- AI assistant mock services pro integration testing

### Critical Path Testing
1. Default help display (blocker pro release)
2. Basic stdio mode functionality
3. AI assistant integration workflows
4. Migration path from legacy commands

## Acceptance Protocol

### Pre-Implementation Validation
- [ ] Architectural review ADR-005 compliance
- [ ] Breaking change impact assessment
- [ ] Migration strategy documentation

### Implementation Validation
- [ ] All unit tests pass (100% pro kritické komponenty)
- [ ] Integration test suite completion
- [ ] Cross-platform compatibility verified
- [ ] Performance benchmarks meet requirements

### Pre-Merge Requirements
- [ ] All acceptance criteria from issue #037 fulfilled
- [ ] Migration guide dokumentace completed
- [ ] AI assistant integration tested s real tools
- [ ] Backward compatibility strategy implemented
- [ ] Help system UX validated

### Post-Merge Monitoring
- [ ] User feedback collection na breaking changes
- [ ] Performance monitoring pro různé communication modes
- [ ] AI assistant integration success rate tracking
- [ ] Legacy command usage deprecation metrics

### Release Readiness Criteria
1. **Zero regressions** v existing MCP functionality
2. **Complete migration path** s clear documentation
3. **AI assistant compatibility** verified with major tools
4. **Help system improvement** confirmed by UX testing
5. **Performance parity** nebo improvement across all modes

---

**Test Priority**: HIGH - Breaking change s significant user impact  
**Risk Level**: MEDIUM - Well-planned architectural change s clear migration path  
**Review Required**: Architecture team + AI integration specialists

**Created**: 2025-09-11  
**QA Engineer**: Claude Code QA Role