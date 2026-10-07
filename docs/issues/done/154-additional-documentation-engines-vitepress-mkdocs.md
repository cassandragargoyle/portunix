# Issue #154: Additional Documentation Engines — VitePress, MkDocs

**Status:** ✅ Implemented
**Implemented:** 2026-05-04
**Acceptance Protocol:** [acceptance-154.md](../../testing/internal/acceptance-154.md) — 12/12 PASS

## Summary

Extend the documentation environment template system (Issue #153) with additional engines requested in GitHub Issue #22 as secondary targets. These are not required for the initial KNIFE project delivery but should be implemented as follow-up.

**Source**: GitHub Issue #22 — secondary documentation systems
**Depends on**: Issue #153 (primary delivery must be completed first)

## Desired State

```bash
# VitePress (Vue-based, Vite-powered)
portunix playbook init --template static-docs --engine vitepress --target container
portunix install vitepress

# MkDocs (Python-based, Material theme)
portunix playbook init --template static-docs --engine mkdocs --target container
portunix install mkdocs
```

## Requirements

### REQ-1: VitePress Support

- Create `vitepress.ptxbook.tmpl` playbook template (node-based container)
- Create `vitepress.json` package definition in ptx-installer
- Support project initialization, dev server (port 5173), and production build
- Add `vitepress` to `metadata.json` engine choices

### REQ-2: MkDocs Support

- Create `mkdocs.ptxbook.tmpl` playbook template (python-based container)
- Create `mkdocs.json` package definition in ptx-installer
- Include Material for MkDocs theme as default option
- Support project initialization, dev server (port 8000), and production build
- Add `mkdocs` to `metadata.json` engine choices

### REQ-3: Automatic OS Detection and Optimization

- Detect host OS and architecture automatically
- Select optimal container base image per platform
- Handle platform-specific quirks (file watching on Windows/macOS, permission mapping on Linux)

### REQ-4: Update User Guide

- Add VitePress and MkDocs to engine comparison table in `docs/guides/quickstart-documentation-site.md`
- Document any platform-specific differences

## Acceptance Criteria

- [ ] `portunix playbook init --template static-docs --engine vitepress` generates valid `.ptxbook`
- [ ] `portunix playbook init --template static-docs --engine mkdocs` generates valid `.ptxbook`
- [ ] VitePress dev server works with shared folder mount
- [ ] MkDocs dev server works with shared folder mount
- [ ] `portunix install vitepress` works as direct command
- [ ] `portunix install mkdocs` works as direct command
- [ ] Automatic OS detection selects correct container configuration
- [ ] User guide updated with new engines

## Priority

**Medium** — Secondary targets per GitHub Issue #22. Deliver after Issue #153 is complete.

## Labels

`feature`, `documentation`, `docker`, `enhancement`

## Related

- Issue #153 — Deliver Docker Documentation Environment for KNIFE Project (prerequisite)
- Issue #119 — PTX-Ansible Standalone Help & Template Examples
- GitHub Issue #22 — Docusaurus multiplatform support
- PFT: VC-V005, VoC S02
