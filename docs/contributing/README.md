# Contributing Documentation

> 🌐 **Language / Jazyk**: **English** | [Čeština](README.cs.md)

This directory contains guidelines and documentation for contributing to the Portunix project.

## Multilingual Team Communication

### Official Project Language

- **English** is the primary language for all project communication
- Code, comments, documentation, and commit messages must be in English
- Use simple, clear English to accommodate non-native speakers
- Avoid idioms, colloquialisms, and culture-specific references

### Communication Best Practices

- Use AI-powered translation tools (DeepL, Google Translate) when needed
- Leverage asynchronous communication to accommodate different time zones
- Foster open communication with regular check-ins for clarification
- Be patient and respectful of language barriers and cultural differences

### Documentation Standards

- All technical documentation in English
- Use clear, concise language with simple sentence structures
- Provide examples and visual aids when possible
- Maintain consistent terminology throughout the project

## Team Initials Assignment and Collision Resolution

Team members are assigned unique initials for identification purposes. When multiple members have the same initials (e.g., John Smith = JS), collision
resolution is applied by adding additional characters from the surname (e.g., John Smith = JSm if JS is already taken).

## Available Guidelines

### General

- [AI Assistants](AI-ASSISTANTS.md) — Guidelines and best practices for using AI assistants in development workflow
- [Terminology](TERMINOLOGY.md) — Common terminology, abbreviations, and technical terms used across the project
- [TODO Guidelines](TODO-GUIDELINES.md) — Standards for writing and managing TODO comments
- [Tools Recommendations](TOOLS-RECOMMENDATIONS.md) — Recommended development tools
- [Translation Workflow](TRANSLATION-WORKFLOW.md) — Translating documentation into team members' native languages
- [Markdown Style](MARKDOWN-STYLE.md) — Markdown style guide for documentation

### Issue & Project Management

- [Issue Management](ISSUE-MANAGEMENT.md) — How issues are tracked and synchronized between Gitea and GitHub
- [Issue Development Methodology](ISSUE-DEVELOPMENT-METHODOLOGY.md) — Mandatory development workflow with quality gates
- [Bug Reporting](BUG-REPORTING.md) — How to report and document bugs
- [Versioning](VERSIONING.md) — Semantic versioning and release strategy

### Git & Publishing Workflows

- [Git Workflow](GIT-WORKFLOW.md) — Branch naming, commit messages, and PR process
- [Gitea Internal Methodology](GITEA-INTERNAL-METHODOLOGY.md) — Internal development on self-hosted Gitea
- [GitHub Workflow](GITHUB-WORKFLOW.md) — Publishing from local Gitea to public GitHub mirror
- [AUR Workflow](AUR-WORKFLOW.md) — Workflow for Arch User Repository releases

### Code Style

Language-specific coding standards and conventions:

- [Go Code Style](CODE-STYLE-GO.md)
- [Java Code Style](CODE-STYLE-JAVA.md)
- [C++ Code Style](CODE-STYLE-CPP.md)
- [Python Code Style](CODE-STYLE-PYTHON.md)

### Testing

- [Testing Methodology](TESTING_METHODOLOGY.md) — Verbose test framework and container-based testing policy
- [Testing — Go](TESTING-GO.md) — Unit testing patterns for Go
- [Testing — Go Project](TESTING-GO-PROJECT.md) — Project-wide testing standards for Go
- [Testing — Java](TESTING-JAVA.md)
- [Testing — C++](TESTING-CPP.md)
- [Testing — Python](TESTING-PYTHON.md)

### Containers & Virtualization

- [Podman Guide](PODMAN-GUIDE.md) — Working with Podman in Portunix
- [VM/SSH Deployment](VM-SSH-DEPLOYMENT.md) — Deploying Portunix into virtual machines via SSH

### Documentation & Release Tooling

- [Manual Creation Methodology](MANUAL-CREATION-METHODOLOGY.md) — How project manuals are authored
- [PDF Generation Workflow](PDF-GENERATION-WORKFLOW.md) — Generating PDF documentation
- [README Dual System](README-DUAL-SYSTEM.md) — Maintaining the dual README system
- [Helper Binary Development](HELPER-BINARY-DEVELOPMENT.md) — Building auxiliary binaries

### Tutorials

- [Maven Basics for Juniors](tutorials/MAVEN-BASICS-FOR-JUNIORS.md)

## Translations

Czech translations are kept next to their English originals using the `.cs.md`
suffix (for example `README.cs.md`, `TERMINOLOGY.cs.md`). Translations are
optional — the English version is always authoritative.

---

**Note**: Portunix is published under the MIT license. Contributing guidelines apply to all
contributors of the public repository.
