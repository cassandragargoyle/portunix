# Architecture Decision Records (ADR)

## What is an ADR?

An **Architecture Decision Record (ADR)** is a document that captures an important architectural decision made along with its context and consequences. ADRs are used to document the reasoning behind significant technical decisions in a project.

## Purpose

ADRs serve several important purposes:
- **Documentation**: Preserve the reasoning behind technical decisions for future reference
- **Communication**: Share architectural decisions with the team and stakeholders
- **History**: Track the evolution of the system architecture over time
- **Onboarding**: Help new team members understand why certain decisions were made
- **Accountability**: Record who made decisions and when
- **Review**: Enable revisiting decisions when context or requirements change

## ADR Structure

Each ADR typically contains:
1. **Title**: Brief description of the decision
2. **Status**: Active, Deprecated, Superseded, etc.
3. **Context**: The issue or problem that motivated this decision
4. **Decision**: The chosen solution or approach
5. **Consequences**: Both positive and negative outcomes of the decision
6. **References**: Related documents, issues, or other ADRs

## Portunix ADR Index

| Number | Title | Status | Date | Author |
|--------|-------|--------|------|--------|
| 001 | Use PostgreSQL | Unknown | - | - |
| 002 | PowerShell Linux Architecture | Unknown | - | - |
| 003 | Linux Distribution Version Support Strategy | Unknown | - | - |
| 004 | Default STDIO Mode for MCP | Unknown | - | - |
| 005 | Revert Default STDIO Mode for MCP | Unknown | - | - |
| 006 | Dynamic Package List Generation | Unknown | - | - |
| 007 | Prerequisite Package Handling System | Unknown | - | - |
| 008 | Dynamic Sudo Handling Post-Install Commands | Unknown | - | - |
| 009 | Officially Supported Linux Distributions | Unknown | - | - |
| 010 | Temporary Virtualization Priority | Active | 2025-01-16 | Zdenek |
| 011 | Multi-Level Help System | Active | 2025-01-17 | Architect |
| 012 | Development Workflow and Contribution Model | Active | 2025-09-18 | Architect |
| 013 | Software Manifests System | Unknown | - | - |
| 014 | Git-like Dispatcher with Python Distribution Model | Proposed | 2025-09-19 | Architect |
| 015 | Logging System Architecture | Proposed | 2025-01-20 | Architect |
| 016 | Ansible Infrastructure as Code Integration | Proposed | 2025-09-23 | Architect |
| 017 | PTX-Prompting Helper for Template-Based Prompt Generation | Active | 2025-09-26 | Architect |
| 018 | Post-Release Documentation Automation and Static Site Generation | Proposed | 2025-09-26 | Architect |
| 019 | Package Metadata URL Tracking | Proposed | 2025-09-27 | Architect |
| 020 | AI Prompts for Package Discovery | Proposed | 2025-09-27 | Architect |
| 021 | Package Registry Architecture | Proposed | 2025-09-27 | Architect |
| 022 | Debtap Package Installation Support | Proposed | 2025-09-28 | Architect |
| 023 | AUR (Arch User Repository) Support | Proposed | 2025-09-28 | Architect |
| 024 | AI Parameter Autocorrection System | Proposed | 2025-09-30 | Architect |
| 025 | PTX-Installer Helper Architecture | Proposed | - | Architect |
| 026 | Shared Platform Utilities | Proposed | - | Architect |
| 027 | Compose Command Architecture | Proposed | 2025-12-01 | Architect |

## Creating New ADRs

### Naming Convention
- Format: `NNN-brief-description.md`
- Example: `011-container-management-strategy.md`
- Numbers are sequential, zero-padded to 3 digits

### Who Can Create ADRs
According to project guidelines (CLAUDE.md):
- **Only the Architect role can write to ADR**
- Developers can propose decisions for Architect review
- All ADRs must be approved before merging

### Process
1. Identify a significant architectural decision that needs documentation
2. Create a new ADR file with the next sequential number
3. Follow the standard ADR structure
4. Submit for Architect review and approval
5. Update this README index once approved

## ADR Lifecycle

### Status Values
- **Draft**: Under development, not yet approved
- **Proposed**: Ready for review
- **Active**: Approved and in effect
- **Deprecated**: No longer relevant but kept for history
- **Superseded**: Replaced by another ADR (reference the new one)

### Updating ADRs
- ADRs are generally immutable once approved
- If a decision needs to change, create a new ADR that supersedes the old one
- Update the status of the old ADR to "Superseded by ADR-NNN"

## Best Practices

1. **Be Concise**: ADRs should be brief but complete
2. **Focus on "Why"**: Explain the reasoning, not just the what
3. **Consider Alternatives**: Document options that were considered but rejected
4. **Include Trade-offs**: Be honest about both benefits and drawbacks
5. **Link Context**: Reference relevant issues, discussions, or documentation
6. **Use Clear Language**: Write for future readers who may lack current context

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions by Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- Portunix project guidelines in `/CLAUDE.md`

---

**Note**: This directory contains critical architectural decisions for the Portunix project. Only authorized Architects should modify these records according to project governance rules.