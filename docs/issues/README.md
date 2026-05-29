# Issues Documentation & Tracking

This directory contains detailed documentation for all issues, feature requests, and development planning.

## Directory Layout

Active issues live in `internal/`; completed (done) issues are archived in `internal/done/`.
When an issue is finalized (✅ Implemented or ❌ Closed), move its file to `internal/done/`
and migrate its row from the **Active** table to the **Done** table below.

## Dual Numbering System

We use a dual numbering system to separate internal development tracking from public GitHub issues:

- **Internal**: All issues (bugs, security, features) tracked in `internal/` with sequential numbering (#001, #002, etc.)
- **Public**: Selected features and enhancements published to GitHub with PUB- prefix (PUB-001, PUB-002, etc.)

## Internal Issues — Active

| Internal | Public | Title | Status | Priority | Type | Labels |
| -------- | ------ | ----- | ------ | -------- | ---- | ------ |
| [#009](internal/009-configurable-datastore.md) | - | Configurable Datastore System | 📋 Open | High | Feature | enhancement, datastore, enterprise |
| [#023](internal/023-arch-linux-distribution-support.md) | - | Arch Linux Distribution Support Integration | 📋 Open | Medium | Feature | enhancement, package-management, linux, arch-linux, aur |
| [#035](internal/035-ai-assistant-installation-support.md) | - | AI Assistant Installation Support | 🔄 Partially Implemented (bundles + detection + macOS install + recommend-ai/MCP-hook/version) | High | Enhancement | package-management, ai-integration, installation, cross-platform, mcp |
| [#051](internal/051-git-dispatcher-python-distribution-architecture.md) | - | Git-like Dispatcher with Python Distribution Architecture | 🔄 In Progress (Phase 2 Complete) | High | Architecture | architecture, dispatcher, helper-binaries, version-1.6 |
| [#067](internal/067-disk-image-files-helper.md) | - | Disk Image Files Helper for Multiple Formats | 📋 Open | High | Enhancement | enhancement, virtualization, disk-management, cross-platform, vdi, vmdk, vhd, qcow2, image-processing |
| [#068](internal/068-main-binary-ptx-virt-helper-integration.md) | - | Main Binary ptx-virt Helper Integration | 📋 Open | High | Enhancement | enhancement, virtualization, dispatcher, helper-binary, integration, consistency |
| [#073](internal/073-ptx-prompting-helper-implementation.md) | - | PTX-Prompting Helper Implementation | 📋 Open | High | Feature | enhancement, helper-system, ai-integration, template-system |
| [#098](internal/098-ptx-vocalio-helper-implementation.md) | - | PTX-Vocalio Helper Implementation | 📋 Open | High | Feature | enhancement, helper-binary, speech-recognition, text-to-speech, ai-integration, accessibility |
| [#105](internal/105-ptx-make-gobuild-cross-platform-compilation.md) | - | PTX-Make GoBuild Cross-Platform Compilation | 📋 Open | High | Enhancement | enhancement, helper-binary, ptx-make, cross-platform, go-compilation |
| [#111](internal/111-ptx-pft-mcp-integration.md) | - | PTX-PFT MCP Integration | 📋 Open | High | Enhancement | enhancement, helper-binary, ptx-pft, ptx-mcp, ai-integration, mcp |
| [#116](internal/116-ptx-pft-iso16355-qfd-project-structure.md) | - | PTX-PFT ISO 16355 QFD Project Structure | 📋 Open | High | Enhancement | enhancement, helper-binary, ptx-pft, iso-16355, qfd, requirements-management |
| [#125](internal/125-cross-platform-binary-distribution.md) | - | Cross-Platform Binary Distribution | 📋 Open | High | Enhancement | enhancement, architecture, distribution, cross-platform, container, vm |
| [#132](internal/132-text-extractor-plugin-integration.md) | - | Text Extractor Plugin Integration | 📋 Open | High | Feature / Plugin | plugin, text-extraction, java, tika, mcp, ai-integration |
| [#150](internal/150-distributed-mcp-server-ecosystem.md) | - | Distributed MCP Server Ecosystem for Plugins | 📋 Open | High | Architecture / Feature | mcp, plugin-system, ai-integration, architecture, distributed |
| [#153](internal/153-deliver-docker-documentation-environment-for-knife.md) | - | Deliver Docker Documentation Environment for KNIFE Project | 📋 Open | High | Feature | feature, documentation, docker, customer-delivery, knife-project |
| [#185](internal/185-plugin-security-hardening.md) | - | Plugin Security Hardening | 📋 Open | Medium | Enhancement / Security | enhancement, plugin-system, security, grpc, tls, signing, audit |

## Internal Issues — Done

| Internal | Public | Title | Status | Priority | Type | Labels |
| -------- | ------ | ----- | ------ | -------- | ---- | ------ |
| [#001](internal/done/001-cross-platform-os-detection.md) | PUB-001 | Cross-Platform Intelligent OS Detection System | ✅ Implemented | High | Feature | enhancement, cross-platform, powershell |
| [#002](internal/done/002-docker-management-command.md) | - | Docker Management Command | ✅ Implemented | High | Feature | enhancement, docker, cross-platform |
| [#003](internal/done/003-podman-management-command.md) | - | Podman Management Command | ✅ Implemented | High | Feature | enhancement, podman, cross-platform |
| [#004](internal/done/004-mcp-server-ai-integration.md) | PUB-002 | MCP Server for AI Assistant Integration | ✅ Implemented | High | Feature | enhancement, mcp, ai-integration |
| [#007](internal/done/007-plugin-system-grpc.md) | - | Plugin System with gRPC Architecture | ✅ Implemented | High | Feature | enhancement, plugin-system, grpc, ai-integration, cross-platform, extensibility, agile, kanban, project-management |
| [#008](internal/done/008-virtual-development-disk.md) | - | Virtual Development Disk Management — Superseded by `portunix container` + `portunix virt` + install profiles | ❌ Closed (Superseded) | High | Feature | enhancement, virtual-disk, cross-platform |
| [#010](internal/done/010-update.md) | PUB-003 | Self-Update Command | ✅ Implemented | High | Feature | enhancement, self-update, cross-platform |
| [#012](internal/done/012-powershell-linux-installation.md) | - | PowerShell Installation Support for Linux | ✅ Implemented | High | Bug Fix | enhancement, powershell, linux |
| [#013](internal/done/013-ptx-database-helper-implementation.md) | - | PTX-Database Helper Implementation (Phase 1: PostgreSQL + SQLite) | ✅ Implemented | High | Feature | helper-binary, database, mcp, installation, backup |
| [#014](internal/done/014-wizard-framework.md) | PUB-004 | Wizard Framework for Interactive CLI (ptx-wizard helper) | ✅ Implemented | High | Enhancement | enhancement, cli, wizard, framework, user-experience, core |
| [#015](internal/done/015-vps-edge-bastion-infrastructure.md) | PUB-005 | VPS Edge/Bastion Infrastructure | ✅ Implemented | High | Feature | infrastructure, edge-computing |
| [#016](internal/done/016-protoc-plugin-development-dependency.md) | - | Protocol Buffers Compiler (protoc) | ✅ Implemented | Critical | Bug Fix | critical, plugin-system, build |
| [#017](internal/done/017-qemu-kvm-windows-virtualization.md) | - | QEMU/KVM Windows 11 Virtualization with Snapshots — Superseded by `portunix virt` (#049, #089, #090–092, #061) | ❌ Closed (Superseded) | High | Feature | virtualization, qemu, kvm, windows, snapshot |
| [#019](internal/done/019-docker-windows-install-issues.md) | - | Docker Installation Issues on Windows | ✅ Implemented | High | Bug Fix | bug, docker, windows |
| [#020](internal/done/020-qemu-windows-clipboard-integration.md) | - | QEMU Windows VM Clipboard Integration — Delivered via `portunix virt` (SPICE default in #017/#049/#055, guest tools package, host packages in #093) | ✅ Implemented | Medium | Enhancement | enhancement, qemu, windows, clipboard, spice |
| [#021](internal/done/021-github-actions-local-testing.md) | - | GitHub Actions Local Testing Support with Act | ✅ Implemented | Medium | Feature | feature, github-actions, act, ci-cd, testing |
| [#022](internal/done/022-google-chrome-installation.md) | - | Google Chrome Installation Implementation | ✅ Implemented | Medium | Feature | enhancement, package-management, cross-platform |
| [#024](internal/done/024-plugin-registration-system.md) | - | Plugin Registration and Discovery System | ✅ Implemented | High | Enhancement | plugin-system, grpc, cli, mcp, discovery, dependency-management |
| [#025](internal/done/025-github-integration-core.md) | - | GitHub Integration for Portunix Core | ✅ Implemented | High | Enhancement | github, git, api-integration, core-enhancement |
| [#026](internal/done/026-github-cli-installation.md) | - | GitHub CLI (gh) Installation Support | ✅ Implemented | Medium | Feature | enhancement, package-management, github, cli, cross-platform |
| [#027](internal/done/027-container-lifecycle-cleanup-guarantees.md) | - | Container Lifecycle Management with Cleanup Guarantees | ✅ Implemented | High | Enhancement | docker, lifecycle-management, cleanup, resource-management, testing |
| [#028](internal/done/028-universal-container-parameters-support.md) | - | Universal Container Parameters Support | ✅ Implemented | High | Enhancement | docker, podman, volume-mounting, cli, container-runtime |
| [#029](internal/done/029-universal-container-command.md) | - | Universal Container Command Implementation | ✅ Implemented | High | Enhancement | container, docker, podman, configuration, universal-interface |
| [#030](internal/done/030-container-tls-certificate-verification-failure.md) | - | Container TLS Certificate Verification Failure | ✅ Implemented | High | Bug Fix | container, docker, podman, tls, certificates, networking, go-installation |
| [#031](internal/done/031-universal-container-exec-command.md) | - | Universal Container Exec Command | ✅ Implemented | High | Enhancement | container, docker, podman, universal-interface, cli, execution |
| [#033](internal/done/033-mcp-plugin-development-guide-ai-agents.md) | - | MCP Server Plugin Development Guide for AI Agents | ✅ Implemented | High | Enhancement | mcp, ai-integration, plugin-system, documentation, claude-code |
| [#034](internal/done/034-mcp-server-installation-wizard.md) | - | MCP Server Wizard — Advanced Features Completion | ✅ Implemented | Medium | Enhancement | mcp, ai-integration, wizard, user-experience, cross-platform |
| [#036](internal/done/036-default-stdio-mode-for-mcp.md) | - | Default stdio Mode for MCP When No Parameters Provided | ✅ Implemented | High | Enhancement | enhancement, mcp, ai-integration, cli, breaking-change |
| [#037](internal/done/037-revert-default-stdio-mcp-implement-mcp-serve.md) | - | Revert Default stdio Mode and Implement MCP Serve Command | ✅ Implemented | High | Enhancement | enhancement, mcp, ai-integration, cli, breaking-change, command-restructure |
| [#038](internal/done/038-container-run-shorthand-flag-parsing-failure.md) | - | Container Run Command Shorthand Flag Parsing Failure | ✅ Implemented | High | Bug Fix | bug, container, cli, flag-parsing, critical |
| [#039](internal/done/039-container-runtime-capability-detection.md) | - | Container Runtime Capability Detection | ✅ Implemented | High | Enhancement | enhancement, container, docker, podman, testing, cli |
| [#040](internal/done/040-migrate-module-name-from-portunix-cz-to-portunix-ai.md) | - | Migrate Go Module Name from portunix.cz to portunix.ai | ✅ Implemented | Medium | Refactoring | refactoring, branding, module-management, breaking-change, internal |
| [#041](internal/done/041-nodejs-npm-installation-support.md) | - | Node.js/npm Installation Support | ✅ Implemented | High | Feature | enhancement, package-management, nodejs, npm, prerequisites |
| [#042](internal/done/042-improve-container-help-clarity.md) | - | Improve Container Command Help Clarity and Recommendations | ✅ Implemented | Medium | Enhancement | enhancement, container, help, user-experience, best-practices |
| [#043](internal/done/043-container-rm-command-alias.md) | - | Add Container RM Command Alias for Better Docker/Podman Compatibility | ✅ Implemented | Low | Enhancement | enhancement, container, usability, docker-compatibility, command-alias |
| [#044](internal/done/044-container-cp-command-missing.md) | - | Container CP Command Missing from Portunix Container System | ✅ Implemented | High | Bug/Enhancement | bug, enhancement, container, core-functionality, testing-blocker |
| [#045](internal/done/045-nodejs-installation-critical-fixes.md) | - | Node.js Installation Critical Fixes | ✅ Implemented | Critical | Bug Fix | critical, bug-fix, nodejs, container, installation |
| [#046](internal/done/046-nodejs-installation-fedora-package-manager-detection.md) | - | Node.js Installation Fails on Fedora Due to Incorrect Package Manager Detection | ✅ Implemented | High | Bug Fix | bug, nodejs, fedora, package-manager |
| [#047](internal/done/047-nodejs-archlinux-package-manager-detection.md) | - | Node.js Installation Fails on Arch Linux Due to Incorrect Package Manager Detection | ✅ Implemented | High | Bug Fix | bug, nodejs, arch-linux, package-manager, container, cross-platform |
| [#048](internal/done/048-system-info-enhanced-container-detection.md) | - | System Info Enhanced Container Detection | ✅ Implemented | Medium | Enhancement | enhancement, container, system-info, user-experience, docker, podman |
| [#049](internal/done/049-qemu-full-support-implementation.md) | - | Full QEMU/KVM Support Implementation in Portunix | ✅ Implemented | Critical | Feature | enhancement, virtualization, qemu, kvm, testing, infrastructure, critical |
| [#050](internal/done/050-multi-level-help-system.md) | - | Multi-Level Help System | ✅ Implemented | Medium | Enhancement | enhancement, help, ux, ai-integration, cli |
| [#052](internal/done/052-logging-system-implementation.md) | - | Logging System Implementation | ✅ Implemented | Critical | Enhancement | enhancement, logging, architecture, mcp, critical |
| [#053](internal/done/053-fix-module-path-naming-inconsistencies.md) | - | Fix Module Path Naming Inconsistencies | ✅ Implemented | High | Bug Fix | refactoring, module-management, consistency, architecture |
| [#054](internal/done/054-guid-generation-module.md) | - | GUID Generation Module for Portunix Core | ✅ Implemented | Medium | Enhancement | enhancement, core, utilities, cli |
| [#055](internal/done/055-vm-management-requirements-enterprise-architect.md) | - | VM Management Requirements for Enterprise Architect | ✅ Implemented | Critical | Feature | virtualization, vm-management, windows, critical, qemu, enterprise |
| [#056](internal/done/056-ansible-infrastructure-as-code-integration.md) | - | Ansible Infrastructure as Code Integration | ✅ Implemented | High | Feature | enhancement, infrastructure-as-code, ansible, helper-binary, multi-environment |
| [#057](internal/done/057-virtualbox-detection-windows-false-negative.md) | - | VirtualBox Detection False Negative on Windows | ✅ Implemented | High | Bug Fix | bug, virtualization, windows, virtualbox, detection |
| [#058](internal/done/058-virt-list-vm-info-access-denied.md) | - | VirtualBox VM Information Access Denied | ✅ Implemented | High | Bug Fix | bug, virtualization, windows, virtualbox, permissions |
| [#059](internal/done/059-playbook-help-command-not-working.md) | - | Playbook Help Command Not Working | ✅ Implemented | High | Bug Fix | bug, playbook, help, cli, ansible, user-experience |
| [#060](internal/done/060-backend-version-display-enhancement.md) | - | Backend Version Display Enhancement | ✅ Implemented | Medium | Enhancement | enhancement, system-info, virtualization, docker, podman, user-experience |
| [#061](internal/done/061-virt-snapshot-list-empty-names.md) | - | Virtual Machine Snapshot List Shows Empty Names | ✅ Implemented | High | Bug Fix | bug, virtualization, snapshot-management, virtualbox, qemu, data-parsing |
| [#062](internal/done/062-ansible-installation-issues.md) | - | Ansible Installation Issues - Platform Detection and Pip Support | ✅ Implemented | High | Bug Fix | critical, bug, installation, platform-detection, pip-support, ansible, prerequisite-resolution |
| [#063](internal/done/063-ansible-galaxy-collections-support.md) | - | Ansible Galaxy Collections Installation Support | ✅ Implemented | High | Enhancement | enhancement, ansible, galaxy, collections, automation, package-management, infrastructure-as-code |
| [#064](internal/done/064-vscode-installation-filename-issue.md) | - | Visual Studio Code Installation Filename Issue | ✅ Implemented | High | Bug Fix | critical, bug, installation, download, filename-resolution, vscode, windows, exe-installer |
| [#065](internal/done/065-terraform-installation-support.md) | - | Terraform Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, terraform, hashicorp, infrastructure-as-code, multi-platform, devops |
| [#066](internal/done/066-double-commander-installation-support.md) | - | Double Commander Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, double-commander, file-manager, sourceforge, cross-platform, gui-application |
| [#069](internal/done/069-container-command-help-display-incorrect-usage.md) | - | Container Command Help Display Shows Incorrect Usage | ✅ Implemented | Medium | Bug Fix | bug, container-management, help-system, user-experience, helper-integration |
| [#070](internal/done/070-ansible-pipx-installation-support.md) | - | Ansible pipx Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, ansible, pipx, cross-platform, installation |
| [#071](internal/done/071-container-exec-command-implementation.md) | - | Container Exec Command Implementation | ✅ Implemented | High | Bug Fix / Enhancement | bug, enhancement, container-management, core-functionality, exec, helper-binary |
| [#072](internal/done/072-cache-architecture-pip-pattern.md) | - | Cache Architecture Redesign Based on pip Pattern | ✅ Implemented | High | Enhancement | enhancement, cache-system, performance, architecture, cross-platform, pip-pattern |
| [#074](internal/done/074-post-release-documentation-automation.md) | - | Post-Release Documentation Automation and Static Site Generation | ✅ Implemented | High | Feature | enhancement, documentation, automation, release-process, github-pages, static-site |
| [#075](internal/done/075-implement-hugo-installation-support.md) | - | Implement Hugo Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, hugo, documentation, static-site-generator |
| [#076](internal/done/076-container-run-help-command-not-working.md) | - | Container Run Help Command Not Working | ✅ Implemented | High | Bug Fix | bug, container-management, help-system, user-experience, cli |
| [#077](internal/done/077-container-run-in-container-help-flag-parsing.md) | - | Container Run-in-Container Help Flag Parsing | ✅ Implemented | High | Bug Fix | bug, container-management, help-system, flag-parsing, cli |
| [#078](internal/done/078-github-cli-installation.md) | - | GitHub CLI Installation Support | ✅ Implemented | Medium | Feature | enhancement, package-management, github-cli, developer-tools, cross-platform |
| [#079](internal/done/079-custom-installation-methods-cli-parameter.md) | - | Enhanced Package Installation with Custom URLs and Methods (Phase 1) | ✅ Implemented | High | Enhancement | enhancement, package-management, installation, custom-methods, advanced-cli |
| [#080](internal/done/080-package-metadata-url-tracking-implementation.md) | - | Package Metadata URL Tracking Implementation | ✅ Implemented | Medium | Enhancement | enhancement, package-management, metadata, documentation, maintenance |
| [#081](internal/done/081-ai-prompts-package-discovery-implementation.md) | - | AI Prompts for Package Discovery Implementation | ✅ Implemented | Medium | Enhancement | enhancement, package-management, ai-integration, metadata, maintenance |
| [#082](internal/done/082-package-registry-architecture-implementation.md) | - | Package Registry Architecture Implementation | ✅ Implemented | Critical | Architecture | architecture, package-management, ai-integration, critical, migration |
| [#083](internal/done/083-hugo-registry-installation-fix.md) | - | Hugo Registry Installation Fix | ✅ Implemented | High | Bug Fix | bug, package-management, registry, hugo |
| [#084](internal/done/084-container-list-command-implementation.md) | - | Container List Command Implementation | ✅ Implemented | High | Feature | container, docker, podman, cli |
| [#085](internal/done/085-hugo-installation-permission-fix.md) | - | Hugo Installation Permission Fix | ✅ Implemented | High | Bug Fix | bug, installation, permissions, hugo, linux, architecture |
| [#086](internal/done/086-package-registry-automatic-discovery.md) | - | Package Registry Automatic Discovery System | ✅ Implemented | Critical | Architecture | critical, architecture, package-registry, discovery, testing-blocker, scalability |
| [#087](internal/done/087-assets-embedding-architecture-critical.md) | - | Assets Embedding Architecture - Critical Binary Distribution Fix | ✅ Implemented | Critical | Architecture | critical, architecture, assets-embedding, binary-distribution, container-compatibility |
| [#088](internal/done/088-virtualbox-kvm-conflict-detection.md) | - | VirtualBox/KVM Conflict Detection and Resolution | ✅ Implemented | High | Enhancement | enhancement, virtualization, virtualbox, kvm, user-experience, virt-check, conflict-resolution |
| [#089](internal/done/089-qemu-kvm-adapter-implementation.md) | - | QEMU/KVM Adapter Implementation for virt check | ✅ Implemented | High | Bug Fix | bug, virtualization, qemu, kvm, ptx-virt, detection, adapter |
| [#090](internal/done/090-libvirt-daemon-detection-and-fix.md) | - | Libvirt Daemon Detection and Auto-Fix | ✅ Implemented | High | Bug Fix | bug, virtualization, qemu, kvm, libvirt, virt-manager, daemon-management |
| [#091](internal/done/091-libvirt-dependency-failed-fix.md) | - | Libvirt Dependency Failed - Root Cause Analysis and Fix | ✅ Implemented | High | Bug Fix | bug, virtualization, libvirt, systemd, dependencies, virt-manager |
| [#092](internal/done/092-libvirt-package-installation-support.md) | - | Libvirt Package Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, libvirt, virtualization, refactoring |
| [#093](internal/done/093-spice-server-client-installation.md) | - | Spice Server and Client Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, virtualization, spice, qemu, kvm, clipboard |
| [#094](internal/done/094-container-rm-subcommand-not-recognized.md) | - | Container 'rm' Subcommand Not Recognized | ✅ Implemented | Medium | Bug Fix | bug, container, cli, command-parsing |
| [#095](internal/done/095-container-exec-returns-helper-version.md) | - | Container exec Returns Helper Version Instead of Executing Command | ✅ Implemented | High | Bug Fix | bug, container, cli, ptx-container, critical |
| [#096](internal/done/096-container-start-stop-help-flag-bug.md) | - | Container Start/Stop Commands Misinterpret --help Flag as Container Name | ✅ Implemented | Medium | Bug Fix | bug, container, help-system, user-experience, cli |
| [#097](internal/done/097-ptx-python-helper-implementation.md) | - | PTX-Python Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, python, development-tools, build-automation, code-quality |
| [#099](internal/done/099-system-info-performance-optimization.md) | - | System Info Performance Optimization | ✅ Implemented | High | Enhancement | enhancement, performance, system-info, optimization, user-experience, critical-path |
| [#100](internal/done/100-ptx-installer-helper-implementation.md) | - | PTX-Installer Helper Implementation | ✅ Implemented | High | Feature | enhancement, architecture, performance, helper-binary, package-management |
| [#101](internal/done/101-ptx-aiops-helper-implementation.md) | - | PTX-AIOps Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, ai-integration, container, gpu-support |
| [#102](internal/done/102-compose-command-implementation.md) | - | Compose Command Implementation | ✅ Implemented | High | Enhancement | enhancement, container, docker-compose, podman-compose, universal-interface |
| [#103](internal/done/103-ptx-make-helper-implementation.md) | - | PTX-Make Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, build-automation, cross-platform, makefile |
| [#104](internal/done/104-ptx-make-ls-command.md) | - | PTX-Make LS Command Implementation | ✅ Implemented | Medium | Enhancement | enhancement, helper-binary, ptx-make, cross-platform, file-operations |
| [#106](internal/done/106-install-command-help-flag-not-working.md) | - | Install Command --help Flag Not Working | ✅ Implemented | High | Bug Fix | bug, cli, help-system, install, user-experience, documentation |
| [#107](internal/done/107-ptx-pft-product-feedback-tool-helper.md) | - | PTX-PFT Product Feedback Tool Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, product-feedback, fider, synchronization |
| [#108](internal/done/108-ptx-pft-email-notifications.md) | - | PTX-PFT E-mail Notifications for User Actions | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-pft, email, notifications |
| [#109](internal/done/109-ptx-pft-clearflask-provider.md) | - | PTX-PFT ClearFlask Provider Implementation | ✅ Implemented | Medium | Enhancement | enhancement, helper-binary, product-feedback, clearflask, provider |
| [#110](internal/done/110-ptx-pft-eververse-provider.md) | - | PTX-PFT Eververse Provider Implementation | ✅ Implemented | Medium | Enhancement | enhancement, helper-binary, product-feedback, eververse, provider, high-complexity |
| [#112](internal/done/112-ptx-pft-category-management.md) | - | PTX-PFT Category Management for UC and Requirements | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-pft, categorization, organization |
| [#113](internal/done/113-mcp-help-missing-subcommands-v180.md) | - | MCP Help Missing Subcommands in v1.8.0 Release | ✅ Implemented | High | Bug Fix | bug, mcp, release, help-system, regression |
| [#114](internal/done/114-mcp-configure-default-stdio-mode.md) | - | MCP Configure Should Default to stdio Mode | ✅ Implemented | Medium | Enhancement | mcp, configuration, ux |
| [#115](internal/done/115-automated-release-notes-generation.md) | - | Automated Release Notes Generation System | ✅ Implemented | High | Enhancement | enhancement, release-process, automation, ai-integration |
| [#117](internal/done/117-ptx-pft-list-qfd-compatibility.md) | - | PTX-PFT List QFD Compatibility | ✅ Implemented | Medium | Enhancement | enhancement, helper-binary, ptx-pft, qfd, compatibility |
| [#118](internal/done/118-system-info-pprof-profiling.md) | - | System Info pprof Profiling | ✅ Implemented | Medium | Enhancement | enhancement, performance, profiling, system-info |
| [#119](internal/done/119-ptx-ansible-standalone-help-and-template-examples.md) | - | PTX-Ansible Standalone Help and Template Examples System | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-ansible, templates, user-experience, documentation |
| [#120](internal/done/120-windows-native-system-info-module.md) | - | Windows Native System Info Module | ✅ Implemented | High | Enhancement | enhancement, performance, windows, system-info, native-api |
| [#121](internal/done/121-virt-list-libvirt-detection-fix.md) | - | Fix libvirt Detection in portunix virt list | ✅ Implemented | Medium | Bug Fix | bug, ptx-virt, libvirt, detection |
| [#122](internal/done/122-consolidate-docker-podman-installation.md) | - | Consolidate Docker/Podman Installation into ptx-installer | ✅ Implemented | High | Refactoring | refactoring, architecture, ptx-installer, docker, podman |
| [#123](internal/done/123-consolidate-installation-systems.md) | - | Consolidate Installation Systems - Remove Duplicate Assets | ✅ Implemented | Medium | Refactoring | refactoring, architecture, ptx-installer, assets |
| [#124](internal/done/124-download-progress-indicator.md) | - | Download Progress Indicator | ✅ Implemented | Medium | Enhancement | enhancement, user-experience, ptx-installer, download |
| [#126](internal/done/126-gh-installation-package-manager-detection.md) | - | GitHub CLI Installation Package Manager Detection Bug | ✅ Implemented | High | Bug Fix | bug, package-management, github-cli, arch-linux, detection |
| [#127](internal/done/127-migrate-openssh-to-ptx-installer.md) | - | Migrate Win32-OpenSSH Installation to ptx-installer | ✅ Implemented | Medium | Enhancement | enhancement, package-management, openssh, cross-platform, refactoring |
| [#128](internal/done/128-docusaurus-container-performance-optimization.md) | - | Docusaurus Container Performance Optimization | ✅ Implemented | High | Enhancement | enhancement, container, docker, performance, docusaurus, playbook |
| [#129](internal/done/129-docusaurus-quickstart-script.md) | - | Docusaurus QuickStart Script for GitHub Release | ✅ Implemented | Medium | Enhancement | enhancement, documentation, user-experience, quickstart, docusaurus, release-assets |
| [#130](internal/done/130-docker-version-detection-bug.md) | - | Docker Version Detection Bug in System Info — Duplicate of #157 | ❌ Closed (Duplicate) | Low | Bug Fix | bug, system-info, docker, windows, duplicate |
| [#131](internal/done/131-openssh-reinstall-hostkeys-bug.md) | - | OpenSSH Installation Refactoring - Embedded Script Support | ✅ Implemented | High | Bug Fix | bug, openssh, windows, installation, refactoring |
| [#133](internal/done/133-plugin-run-command-argument-forwarding.md) | - | Plugin Run Command Argument Forwarding | ✅ Implemented | Medium | Enhancement | enhancement, plugin-system, cli, ai-integration, user-experience |
| [#134](internal/done/134-pft-config-cross-platform-path-support.md) | - | PFT Config Cross-Platform Path Support | ✅ Implemented | High | Enhancement | bug, enhancement, helper-binary, ptx-pft, cross-platform, configuration |
| [#135](internal/done/135-pft-assign-show-path-parameter-bug.md) | - | PFT Assign/Show Commands Ignore --path Parameter | ✅ Implemented | High | Bug Fix | bug, helper-binary, ptx-pft, cli, path-parameter, regression |
| [#136](internal/done/136-ptx-credential-helper-implementation.md) | - | PTX-Credential Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, security, credential-storage, encryption, cross-platform |
| [#137](internal/done/137-graalvm-installation-support.md) | - | GraalVM Installation Support | ✅ Implemented | High | Feature | enhancement, package-management, graalvm, java, native-image, cross-platform |
| [#138](internal/done/138-ptx-python-project-local-venv-support.md) | - | PTX-Python Project-Local Virtual Environment Support | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-python, python, virtual-environment, developer-experience |
| [#139](internal/done/139-tea-gitea-cli-installation-support.md) | - | Tea (Gitea CLI) Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, gitea, cli, developer-tools, cross-platform |
| [#140](internal/done/140-version-management-strategy-implementation.md) | - | Version Management Strategy Implementation (ADR-036) | ✅ Implemented | High | Enhancement | enhancement, versioning, workflow, contributor-experience, adr-implementation |
| [#142](internal/done/142-elasticsearch-container-installation.md) | - | Elasticsearch Container Installation | ✅ Implemented | Medium | Feature | installation, container, elasticsearch, fulltext |
| [#144](internal/done/144-ptx-trace-session-delete-button.md) | - | PTX-TRACE Session Delete Button | ✅ Implemented | Medium | Enhancement | enhancement, ptx-trace, ui, dashboard, session-management |
| [#145](internal/done/145-ninja-installation-support.md) | - | Ninja Build System Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, ninja, build-system, cross-platform, ptx-installer |
| [#146](internal/done/146-rust-installation-support.md) | - | Rust Programming Language Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, rust, programming-language, cross-platform, ptx-installer, rustup |
| [#147](internal/done/147-clang-installation-support.md) | - | Clang/LLVM Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, clang, llvm, compiler, c-cpp, cross-platform, ptx-installer |
| [#148](internal/done/148-minio-installation-support.md) | - | MinIO Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, minio, object-storage, s3-compatible, cross-platform, ptx-installer, go-install |
| [#149](internal/done/149-vox-ptx-install-integration.md) | - | Vox Plugin - ptx-install Integration for Model Installation | ✅ Implemented | Medium | Feature | ptx-installer, integration, models, multi-file-download, plugin-support |
| [#151](internal/done/151-playwright-browsers-package.md) | - | Playwright Browsers Package Definition | ✅ Implemented | Medium | Feature | ptx-installer, package-management, playwright, browser-automation, cross-platform |
| [#152](internal/done/152-plugin-create-missing-go-mod.md) | - | Plugin Create Missing go.mod | ✅ Implemented | Medium | Bug Fix | bug, plugin-system, cli, developer-experience |
| [#154](internal/done/154-additional-documentation-engines-vitepress-mkdocs.md) | - | Additional Documentation Engines — VitePress, MkDocs | ✅ Implemented | Medium | Feature | feature, documentation, docker, enhancement |
| [#155](internal/done/155-plugin-prerequisites-validation.md) | - | Plugin Prerequisites Validation | ✅ Implemented | High | Enhancement | enhancement, plugin-system, validation, prerequisites, user-experience |
| [#156](internal/done/156-plugin-health-not-enabled-for-helper-plugins.md) | - | Plugin Health Reports "Not Enabled" for All Helper Plugins | ✅ Implemented | Low | Bug Fix | bug, plugin-system, cli, ux |
| [#157](internal/done/157-system-info-docker-version-text-truncated.md) | - | System Info Docker Version Text Truncated | ✅ Implemented | Medium | Bug Fix | bug, system-info, docker, text-parsing |
| [#158](internal/done/158-python-wheel-plugin-installation-support.md) | - | Python Wheel Plugin Installation Support | ✅ Implemented | High | Enhancement | enhancement, plugin-system, python, wheel, venv, installation |
| [#159](internal/done/159-service-orchestration-command-group.md) | - | Service Orchestration Command Group | ✅ Implemented | High | Feature | feature, grpc, service-management, process-orchestration |
| [#160](internal/done/160-python-plugin-grpc-service-support.md) | - | Python Plugin gRPC Service Support | ✅ Implemented | High | Enhancement | enhancement, plugin-system, python, grpc, service-management, venv |
| [#161](internal/done/161-plugin-install-extra-wheels-support.md) | - | Plugin Install Extra Wheels Support | ✅ Implemented | High | Bug Fix / Enhancement | bug, plugin-system, python, wheel, installation |
| [#162](internal/done/162-plugin-interfaces-field-misplacement.md) | - | Plugin Interfaces Field Misplacement — Schema Mismatch | ✅ Implemented | High | Bug Fix | bug, plugin-system, manifest, schema-mismatch, interfaces |
| [#163](internal/done/163-helpers-missing-help-ai-help-expert.md) | - | Helpers Missing --help-ai and --help-expert Flags | ✅ Implemented | Medium | Enhancement | helpers, help-system, consistency, preflight |
| [#164](internal/done/164-windows-install-ps1-usability-overhaul.md) | - | Windows install.ps1 Usability Overhaul | ✅ Implemented | High | Bug Fix / Enhancement | bug, enhancement, installation, windows, powershell, user-experience, github-distribution |
| [#165](internal/done/165-authenticode-code-signing-certificate.md) | - | Self-Signed Update Channel Verification (originally Authenticode CA — CA scope deferred) | ✅ Implemented | Medium | Enhancement | enhancement, security, update, code-signing |
| [#166](internal/done/166-self-signed-certificate-update-verification.md) | - | Self-Signed Certificate for Update Verification (Phase 1) | ✅ Implemented | Medium | Enhancement | enhancement, security, update, code-signing, supply-chain |
| [#168](internal/done/168-gnu-make-installation-support.md) | - | GNU Make Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, build-tools, cross-platform, direct-download |
| [#169](internal/done/169-service-package-windows-build-failure.md) | - | Service Package Windows Build Failure | ✅ Implemented | High | Bug Fix | bug, build, windows, cross-platform, service-orchestration, syscall |
| [#170](internal/done/170-dev-setup-script-fixes.md) | - | Dev Setup Script Fixes (nodejs package name + hugo.json trailing comma) | ✅ Implemented | Medium | Bug Fix | bug, dev-setup, package-management, json, developer-experience |
| [#171](internal/done/171-uv-python-tooling-migration.md) | - | Migrate Python Tooling to uv | ✅ Implemented | Medium | Enhancement / Refactoring | enhancement, tooling, python, uv, developer-experience, ci-cd |
| [#172](internal/done/172-odoo-installation-support.md) | - | Odoo Installation Support | ✅ Implemented | High | Feature | enhancement, package-management, odoo, erp, python, postgresql, container, cross-platform, ptx-installer |
| [#173](internal/done/173-portunix-container-missing-subcommands.md) | - | Portunix Container: Missing `network`, `volume`, `inspect` Subcommands | ✅ Implemented | Medium | Enhancement | enhancement, container, podman, docker, test-methodology, ptx-container |
| [#174](internal/done/174-ptx-ssh-helper.md) | - | Implement `ptx-ssh` Helper — Unified SSH Client and Non-Interactive Password Auth | ✅ Implemented | High | Feature | feature, helper, security, ssh, new-helper, portunix-core, dispatcher |
| [#175](internal/done/175-plugin-platform-capability-query.md) | - | Plugin Platform-Capability Query Interface (Synapse Ask) | ✅ Implemented | High | Feature / Architecture | enhancement, plugin-system, architecture, grpc, cli, cross-team, api-contract, manifest |
| [#176](internal/done/176-docker-install-missing-directory-prompt.md) | - | `portunix install docker` Missing Installation Directory Prompt | ✅ Implemented | High | Bug Fix | bug, installation, docker, ptx-installer, user-experience, regression |
| [#177](internal/done/177-docker-desktop-admin-elevation.md) | - | `portunix install docker` Fails Without Admin Elevation on Windows | ✅ Implemented | High | Bug Fix | bug, installation, docker, ptx-installer, windows, user-experience, uac |
| [#178](internal/done/178-docker-auto-elevation-uac.md) | - | Auto-elevate Docker Desktop Install via UAC on Windows | ✅ Implemented | Medium | Enhancement | enhancement, installation, docker, ptx-installer, windows, user-experience, uac |
| [#179](internal/done/179-docker-install-ux-improvements.md) | - | Docker Install UX Improvements on Windows (progress, WSL prereqs, `install wsl`) | ✅ Implemented | High | Enhancement | enhancement, installation, docker, ptx-installer, windows, user-experience, wsl, hyper-v |
| [#180](internal/done/180-ed25519-release-signing-pipeline.md) | - | Ed25519 Release Signing Pipeline (Phase 2 of #166) — Superseded by #165 + ADR-042 | ❌ Closed (Superseded) | Medium | Enhancement | enhancement, security, update, code-signing, supply-chain, release-pipeline |
| [#167](internal/done/167-proxmox-vm-ct-management.md) | - | Proxmox VM/CT Management Commands | ✅ Implemented | High | Feature | feature, proxmox, virtualization, infrastructure, deployment |
| [#182](internal/done/182-windows-python-wheel-venv-exec.md) | - | Python Wheel Plugins Fail on Windows — Missing `.exe` Suffix in venv Path | ✅ Implemented | High | Bug Fix | bug, plugin-system, python, wheel, venv, windows, cross-platform |
| [#183](internal/done/183-ptx-specpm-helper-implementation.md) | - | PTX-SpecPM Helper Implementation — `portunix specpm` (Phase 1) | ✅ Implemented | High | Feature | enhancement, helper-binary, new-helper, spec-kit-pm, project-management, ai-integration, dispatcher |
| [#184](internal/done/184-ptx-specpm-phase2-upgrade.md) | - | PTX-SpecPM Phase 2 — `upgrade` Command + Air-Gapped Workflow | ✅ Implemented | Medium | Feature | enhancement, helper-binary, ptx-specpm, spec-kit-pm, air-gapped, integration-test |
| [#181](internal/done/181-container-ls-subcommand-alias.md) | - | Container `ls` Subcommand Not Recognized (alias for `list`) | ✅ Implemented | Medium | Bug Fix | bug, container, docker, podman, dispatcher, cli, alias |
| [#186](internal/done/186-ptx-installer-legacy-code-cleanup.md) | - | PTX-Installer Legacy Code Cleanup (follow-up to #100) | ✅ Implemented | Medium | Refactor / Architecture | refactor, architecture, technical-debt, helper-binary, package-management, follow-up-100 |
| [#187](internal/done/187-ptx-installer-path-append.md) | - | ptx-installer — Honor `environment.PATH_APPEND` After Install | ✅ Implemented | High | Bug Fix / Enhancement | bug, enhancement, ptx-installer, installation, path, windows, linux, user-experience |
| [#188](internal/done/188-plugin-list-description-truncation.md) | - | `portunix plugin list` — Description Truncation with `...` and `--verbose` Full Output | ✅ Implemented | Medium | Enhancement | enhancement, plugin-system, cli, user-experience, output-formatting |
| [#141](internal/done/141-ptx-trace-helper-implementation.md) | - | PTX-TRACE Universal Tracing Helper (First: Data Transformation) | ✅ Implemented | High | Feature / Helper Binary | helper-binary, data-transformation, tracing, etl, debugging, ai-integration, mcp |
| [#032](internal/done/032-universal-container-management-commands.md) | - | Universal Container Management Commands | ✅ Implemented | High | Enhancement | container, docker, podman, universal-interface, cli, management |

## Directory Structure

```text
docs/issues/
├── README.md           # This file - main tracking tables (Active + Done)
├── internal/           # Active internal issues (Open / In Progress / On Hold / Ready)
│   ├── NNN-*.md
│   └── done/           # Archived completed issues (Implemented / Closed)
│       └── NNN-*.md
└── public/
    └── mapping.json    # Mapping between internal and public issue numbers
```

## Usage

### Creating New Issues

1. **Internal Issue (all types):**
   - Create file: `internal/{next-number}-{short-title}.md`
   - Add row to the **Active** table above
   - Set Public column to `-` initially

2. **Publishing to GitHub (features/enhancements only):**
   - Assign next PUB- number in mapping.json
   - Update Public column in this README
   - Create GitHub issue with PUB- number
   - Never publish: bugs, security issues, internal tasks

### Archiving Completed Issues

When an issue reaches terminal status (✅ Implemented or ❌ Closed):

1. Move the file: `git mv internal/NNN-name.md internal/done/NNN-name.md`
2. Move its row from the **Active** table to the **Done** table in this README
3. Update the link path from `internal/…` to `internal/done/…`

### Issue Types

- **Feature**: New functionality (can be public)
- **Enhancement**: Improvement to existing features (can be public)
- **Bug Fix**: Fixing broken functionality (internal only)
- **Security**: Security-related issues (internal only)
- **Plugin**: Plugin-specific features (selective public)

### Status Legend

- 📋 Open - Issue is open and needs work
- 🔄 In Progress - Issue is being actively worked on
- ✅ Implemented - Issue has been completed and implemented (file lives in `internal/done/`)
- ❌ Closed - Issue has been closed without implementation (file lives in `internal/done/`)
- ⏸️ On Hold - Issue is temporarily paused

### Priority Legend

- **Critical** - Must be fixed immediately
- **High** - Important feature or significant bug
- **Medium** - Nice to have feature or minor bug
- **Low** - Enhancement or cosmetic issue

## Publishing Guidelines

✅ **Can be published to GitHub:**

- New features
- Enhancements
- Feature requests
- Roadmap items
- Success stories

❌ **Keep internal only:**

- Bug reports and fixes
- Security vulnerabilities
- Performance issues
- Critical errors
- Internal refactoring
- Technical debt
