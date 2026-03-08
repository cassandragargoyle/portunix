# Issues Documentation & Tracking

This directory contains detailed documentation for all issues, feature requests, and development planning.

## Dual Numbering System

We use a dual numbering system to separate internal development tracking from public GitHub issues:

- **Internal**: All issues (bugs, security, features) tracked in `internal/` with sequential numbering (#001, #002, etc.)
- **Public**: Selected features and enhancements published to GitHub with PUB- prefix (PUB-001, PUB-002, etc.)

## Issues List

| Internal | Public | Title | Status | Priority | Type | Labels |
| -------- | ------ | ----- | ------ | -------- | ---- | ------ |
| [#001](internal/001-cross-platform-os-detection.md) | PUB-001 | Cross-Platform Intelligent OS Detection System | ✅ Implemented | High | Feature | enhancement, cross-platform, powershell |
| [#002](internal/002-docker-management-command.md) | - | Docker Management Command | ✅ Implemented | High | Feature | enhancement, docker, cross-platform |
| [#003](internal/003-podman-management-command.md) | - | Podman Management Command | ✅ Implemented | High | Feature | enhancement, podman, cross-platform |
| [#004](internal/004-mcp-server-ai-integration.md) | PUB-002 | MCP Server for AI Assistant Integration | ✅ Implemented | High | Feature | enhancement, mcp, ai-integration |
| [#007](internal/007-plugin-system-grpc.md) | - | Plugin System with gRPC Architecture | 📋 Open | High | Feature | enhancement, plugin-system, grpc |
| [#008](internal/008-virtual-development-disk.md) | - | Virtual Development Disk Management | 📋 Open | High | Feature | enhancement, virtual-disk, cross-platform |
| [#009](internal/009-configurable-datastore.md) | - | Configurable Datastore System | 📋 Open | High | Feature | enhancement, datastore, enterprise |
| [#010](internal/010-update.md) | PUB-003 | Self-Update Command | ✅ Implemented | High | Feature | enhancement, self-update, cross-platform |
| [#012](internal/012-powershell-linux-installation.md) | - | PowerShell Installation Support for Linux | ✅ Implemented | High | Bug Fix | enhancement, powershell, linux |
| [#013](internal/013-database-management-plugin.md) | - | Database Management Plugin | 📋 Open | High | Plugin | plugin, database, mcp |
| [#014](internal/014-wizard-framework.md) | PUB-004 | Wizard Framework for Interactive CLI | 📋 Open | High | Enhancement | enhancement, cli, wizard, ux |
| [#015](internal/015-vps-edge-bastion-infrastructure.md) | PUB-005 | VPS Edge/Bastion Infrastructure | ✅ Implemented | High | Feature | infrastructure, edge-computing |
| [#016](internal/016-protoc-plugin-development-dependency.md) | - | Protocol Buffers Compiler (protoc) | ✅ Implemented | Critical | Bug Fix | critical, plugin-system, build |
| [#017](internal/017-qemu-kvm-windows-virtualization.md) | - | QEMU/KVM Windows 11 Virtualization with Snapshots | 📋 Open | High | Feature | virtualization, qemu, kvm, windows, snapshot |
| [#019](internal/019-docker-windows-install-issues.md) | - | Docker Installation Issues on Windows | 🔄 In Progress | High | Bug Fix | bug, docker, windows |
| [#020](internal/020-qemu-windows-clipboard-integration.md) | - | QEMU Windows VM Clipboard Integration | 📋 Open | Medium | Enhancement | enhancement, qemu, windows, clipboard, spice |
| [#021](internal/021-github-actions-local-testing.md) | - | GitHub Actions Local Testing Support with Act | ✅ Implemented | Medium | Feature | feature, github-actions, act, ci-cd, testing |
| [#022](internal/022-google-chrome-installation.md) | - | Google Chrome Installation Implementation | ✅ Implemented | Medium | Feature | enhancement, package-management, cross-platform |
| [#023](internal/023-arch-linux-distribution-support.md) | - | Arch Linux Distribution Support Integration | 📋 Open | Medium | Feature | enhancement, package-management, linux, arch-linux, aur |
| [#024](internal/024-plugin-registration-system.md) | - | Plugin Registration and Discovery System | 📋 Ready for Implementation | High | Enhancement | plugin-system, grpc, cli, mcp, discovery, dependency-management |
| [#025](internal/025-github-integration-core.md) | - | GitHub Integration for Portunix Core | 📋 Open | High | Enhancement | github, git, api-integration, core-enhancement |
| [#026](internal/026-github-cli-installation.md) | - | GitHub CLI (gh) Installation Support | ✅ Implemented | Medium | Feature | enhancement, package-management, github, cli, cross-platform |
| [#027](internal/027-container-lifecycle-cleanup-guarantees.md) | - | Container Lifecycle Management with Cleanup Guarantees | 📋 Open | High | Enhancement | docker, lifecycle-management, cleanup, resource-management, testing |
| [#028](internal/028-universal-container-parameters-support.md) | - | Universal Container Parameters Support | ✅ Implemented | High | Enhancement | docker, podman, volume-mounting, cli, container-runtime |
| [#029](internal/029-universal-container-command.md) | - | Universal Container Command Implementation | ✅ Implemented | High | Enhancement | container, docker, podman, configuration, universal-interface |
| [#030](internal/030-container-tls-certificate-verification-failure.md) | - | Container TLS Certificate Verification Failure | ✅ Implemented | High | Bug Fix | container, docker, podman, tls, certificates, networking, go-installation |
| [#031](internal/031-universal-container-exec-command.md) | - | Universal Container Exec Command | ✅ Implemented | High | Enhancement | container, docker, podman, universal-interface, cli, execution |
| [#032](internal/032-universal-container-management-commands.md) | - | Universal Container Management Commands | 📋 Open | High | Enhancement | container, docker, podman, universal-interface, cli, management |
| [#033](internal/033-mcp-plugin-development-guide-ai-agents.md) | - | MCP Server Plugin Development Guide for AI Agents | 📋 Open | High | Enhancement | mcp, ai-integration, plugin-system, documentation, claude-code |
| [#034](internal/034-mcp-server-installation-wizard.md) | - | MCP Server Command Restructuring + Interactive Wizard | 📋 Open | High | Enhancement | mcp, ai-integration, wizard, command-restructure, user-experience, safety |
| [#057](internal/057-virtualbox-detection-windows-false-negative.md) | - | VirtualBox Detection False Negative on Windows | ✅ Implemented | High | Bug Fix | bug, virtualization, windows, virtualbox, detection |
| [#035](internal/035-ai-assistant-installation-support.md) | - | AI Assistant Installation Support | 📋 Open | High | Enhancement | package-management, ai-integration, installation, cross-platform, mcp |
| [#036](internal/036-default-stdio-mode-for-mcp.md) | - | Default stdio Mode for MCP When No Parameters Provided | ✅ Implemented | High | Enhancement | enhancement, mcp, ai-integration, cli, breaking-change |
| [#037](internal/037-revert-default-stdio-mcp-implement-mcp-serve.md) | - | Revert Default stdio Mode and Implement MCP Serve Command | ✅ Implemented | High | Enhancement | enhancement, mcp, ai-integration, cli, breaking-change, command-restructure |
| [#038](internal/038-container-run-shorthand-flag-parsing-failure.md) | - | Container Run Command Shorthand Flag Parsing Failure | ✅ Implemented | High | Bug Fix | bug, container, cli, flag-parsing, critical |
| [#039](internal/039-container-runtime-capability-detection.md) | - | Container Runtime Capability Detection | ✅ Implemented | High | Enhancement | enhancement, container, docker, podman, testing, cli |
| [#040](internal/040-migrate-module-name-from-portunix-cz-to-portunix-ai.md) | - | Migrate Go Module Name from portunix.cz to portunix.ai | ✅ Implemented | Medium | Refactoring | refactoring, branding, module-management, breaking-change, internal |
| [#041](internal/041-nodejs-npm-installation-support.md) | - | Node.js/npm Installation Support | ✅ Implemented | High | Feature | enhancement, package-management, nodejs, npm, prerequisites |
| [#042](internal/042-improve-container-help-clarity.md) | - | Improve Container Command Help Clarity and Recommendations | ✅ Implemented | Medium | Enhancement | enhancement, container, help, user-experience, best-practices |
| [#043](internal/043-container-rm-command-alias.md) | - | Add Container RM Command Alias for Better Docker/Podman Compatibility | ✅ Implemented | Low | Enhancement | enhancement, container, usability, docker-compatibility, command-alias |
| [#044](internal/044-container-cp-command-missing.md) | - | Container CP Command Missing from Portunix Container System | ✅ Implemented | High | Bug/Enhancement | bug, enhancement, container, core-functionality, testing-blocker |
| [#045](internal/045-nodejs-installation-critical-fixes.md) | - | Node.js Installation Critical Fixes | ✅ Implemented | Critical | Bug Fix | critical, bug-fix, nodejs, container, installation |
| [#046](internal/046-nodejs-installation-fedora-package-manager-detection.md) | - | Node.js Installation Fails on Fedora Due to Incorrect Package Manager Detection | ✅ Implemented | High | Bug Fix | bug, nodejs, fedora, package-manager |
| [#047](internal/047-nodejs-archlinux-package-manager-detection.md) | - | Node.js Installation Fails on Arch Linux Due to Incorrect Package Manager Detection | ✅ Implemented | High | Bug Fix | bug, nodejs, arch-linux, package-manager, container, cross-platform |
| [#048](internal/048-system-info-enhanced-container-detection.md) | - | System Info Enhanced Container Detection | ✅ Implemented | Medium | Enhancement | enhancement, container, system-info, user-experience, docker, podman |
| [#049](internal/049-qemu-full-support-implementation.md) | - | Full QEMU/KVM Support Implementation in Portunix | ✅ Implemented | Critical | Feature | enhancement, virtualization, qemu, kvm, testing, infrastructure, critical |
| [#050](internal/050-multi-level-help-system.md) | - | Multi-Level Help System | ✅ Implemented | Medium | Enhancement | enhancement, help, ux, ai-integration, cli |
| [#051](internal/051-git-dispatcher-python-distribution-architecture.md) | - | Git-like Dispatcher with Python Distribution Architecture | 🔄 In Progress (Phase 2 Complete) | High | Architecture | architecture, dispatcher, helper-binaries, version-1.6 |
| [#052](internal/052-logging-system-implementation.md) | - | Logging System Implementation | ✅ Implemented | Critical | Enhancement | enhancement, logging, architecture, mcp, critical |
| [#053](internal/053-fix-module-path-naming-inconsistencies.md) | - | Fix Module Path Naming Inconsistencies | ✅ Implemented | High | Bug Fix | refactoring, module-management, consistency, architecture |
| [#054](internal/054-guid-generation-module.md) | - | GUID Generation Module for Portunix Core | ✅ Implemented | Medium | Enhancement | enhancement, core, utilities, cli |
| [#055](internal/055-vm-management-requirements-enterprise-architect.md) | - | VM Management Requirements for Enterprise Architect | ✅ Implemented | Critical | Feature | virtualization, vm-management, windows, critical, qemu, enterprise |
| [#056](internal/056-ansible-infrastructure-as-code-integration.md) | - | Ansible Infrastructure as Code Integration | ✅ Implemented | High | Feature | enhancement, infrastructure-as-code, ansible, helper-binary, multi-environment |
| [#058](internal/058-virt-list-vm-info-access-denied.md) | - | VirtualBox VM Information Access Denied | 📋 Open | High | Bug Fix | bug, virtualization, windows, virtualbox, permissions |
| [#059](internal/059-playbook-help-command-not-working.md) | - | Playbook Help Command Not Working | ✅ Implemented | High | Bug Fix | bug, playbook, help, cli, ansible, user-experience |
| [#060](internal/060-backend-version-display-enhancement.md) | - | Backend Version Display Enhancement | ✅ Implemented | Medium | Enhancement | enhancement, system-info, virtualization, docker, podman, user-experience |
| [#061](internal/061-virt-snapshot-list-empty-names.md) | - | Virtual Machine Snapshot List Shows Empty Names | ✅ Implemented | High | Bug Fix | bug, virtualization, snapshot-management, virtualbox, qemu, data-parsing |
| [#062](internal/062-ansible-installation-issues.md) | - | Ansible Installation Issues - Platform Detection and Pip Support | 📋 Open | High | Bug Fix | critical, bug, installation, platform-detection, pip-support, ansible, prerequisite-resolution |
| [#063](internal/063-ansible-galaxy-collections-support.md) | - | Ansible Galaxy Collections Installation Support | ✅ Implemented | High | Enhancement | enhancement, ansible, galaxy, collections, automation, package-management, infrastructure-as-code |
| [#064](internal/064-vscode-installation-filename-issue.md) | - | Visual Studio Code Installation Filename Issue | ✅ Implemented | High | Bug Fix | critical, bug, installation, download, filename-resolution, vscode, windows, exe-installer |
| [#065](internal/065-terraform-installation-support.md) | - | Terraform Installation Support | 📋 Open | High | Enhancement | enhancement, package-management, terraform, hashicorp, infrastructure-as-code, multi-platform, devops |
| [#066](internal/066-double-commander-installation-support.md) | - | Double Commander Installation Support | 📋 Open | Medium | Enhancement | enhancement, package-management, double-commander, file-manager, sourceforge, cross-platform, gui-application |
| [#067](internal/067-disk-image-files-helper.md) | - | Disk Image Files Helper for Multiple Formats | 📋 Open | High | Enhancement | enhancement, virtualization, disk-management, cross-platform, vdi, vmdk, vhd, qcow2, image-processing |
| [#068](internal/068-main-binary-ptx-virt-helper-integration.md) | - | Main Binary ptx-virt Helper Integration | 📋 Open | High | Enhancement | enhancement, virtualization, dispatcher, helper-binary, integration, consistency |
| [#069](internal/069-container-command-help-display-incorrect-usage.md) | - | Container Command Help Display Shows Incorrect Usage | ✅ Implemented | Medium | Bug Fix | bug, container-management, help-system, user-experience, helper-integration |
| [#078](internal/078-github-cli-installation.md) | - | GitHub CLI Installation Support | ✅ Implemented | Medium | Feature | enhancement, package-management, github-cli, developer-tools, cross-platform |
| [#079](internal/079-custom-installation-methods-cli-parameter.md) | - | Enhanced Package Installation with Custom URLs and Methods | 📋 Open | High | Enhancement | enhancement, package-management, installation, custom-methods, advanced-cli |
| [#070](internal/070-ansible-pipx-installation-support.md) | - | Ansible pipx Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, ansible, pipx, cross-platform, installation |
| [#071](internal/071-container-exec-command-implementation.md) | - | Container Exec Command Implementation | ✅ Implemented | High | Bug Fix / Enhancement | bug, enhancement, container-management, core-functionality, exec, helper-binary |
| [#072](internal/072-cache-architecture-pip-pattern.md) | - | Cache Architecture Redesign Based on pip Pattern | 📋 Open | High | Enhancement | enhancement, cache-system, performance, architecture, cross-platform, pip-pattern |
| [#073](internal/073-ptx-prompting-helper-implementation.md) | - | PTX-Prompting Helper Implementation | 📋 Open | High | Feature | enhancement, helper-system, ai-integration, template-system |
| [#074](internal/074-post-release-documentation-automation.md) | - | Post-Release Documentation Automation and Static Site Generation | ✅ Implemented | High | Feature | enhancement, documentation, automation, release-process, github-pages, static-site |
| [#075](internal/075-implement-hugo-installation-support.md) | - | Implement Hugo Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, hugo, documentation, static-site-generator |
| [#076](internal/076-container-run-help-command-not-working.md) | - | Container Run Help Command Not Working | ✅ Implemented | High | Bug Fix | bug, container-management, help-system, user-experience, cli |
| [#077](internal/077-container-run-in-container-help-flag-parsing.md) | - | Container Run-in-Container Help Flag Parsing | ✅ Implemented | High | Bug Fix | bug, container-management, help-system, flag-parsing, cli |
| [#080](internal/080-package-metadata-url-tracking-implementation.md) | - | Package Metadata URL Tracking Implementation | 📋 Open | Medium | Enhancement | enhancement, package-management, metadata, documentation, maintenance |
| [#081](internal/081-ai-prompts-package-discovery-implementation.md) | - | AI Prompts for Package Discovery Implementation | 📋 Open | Medium | Enhancement | enhancement, package-management, ai-integration, metadata, maintenance |
| [#082](internal/082-package-registry-architecture-implementation.md) | - | Package Registry Architecture Implementation | ✅ Implemented | Critical | Architecture | architecture, package-management, ai-integration, critical, migration |
| [#083](internal/083-hugo-registry-installation-fix.md) | - | Hugo Registry Installation Fix | ✅ Implemented | High | Bug Fix | bug, package-management, registry, hugo |
| [#084](internal/084-container-list-command-implementation.md) | - | Container List Command Implementation | ✅ Implemented | High | Feature | container, docker, podman, cli |
| [#085](internal/085-hugo-installation-permission-fix.md) | - | Hugo Installation Permission Fix | ✅ Implemented | High | Bug Fix | bug, installation, permissions, hugo, linux, architecture |
| [#086](internal/086-package-registry-automatic-discovery.md) | - | Package Registry Automatic Discovery System | ✅ Implemented | Critical | Architecture | critical, architecture, package-registry, discovery, testing-blocker, scalability |
| [#087](internal/087-assets-embedding-architecture-critical.md) | - | Assets Embedding Architecture - Critical Binary Distribution Fix | ✅ Implemented | Critical | Architecture | critical, architecture, assets-embedding, binary-distribution, container-compatibility |
| [#088](internal/088-virtualbox-kvm-conflict-detection.md) | - | VirtualBox/KVM Conflict Detection and Resolution | ✅ Implemented | High | Enhancement | enhancement, virtualization, virtualbox, kvm, user-experience, virt-check, conflict-resolution |
| [#089](internal/089-qemu-kvm-adapter-implementation.md) | - | QEMU/KVM Adapter Implementation for virt check | ✅ Implemented | High | Bug Fix | bug, virtualization, qemu, kvm, ptx-virt, detection, adapter |
| [#090](internal/090-libvirt-daemon-detection-and-fix.md) | - | Libvirt Daemon Detection and Auto-Fix | ✅ Implemented | High | Bug Fix | bug, virtualization, qemu, kvm, libvirt, virt-manager, daemon-management |
| [#091](internal/091-libvirt-dependency-failed-fix.md) | - | Libvirt Dependency Failed - Root Cause Analysis and Fix | ✅ Implemented | High | Bug Fix | bug, virtualization, libvirt, systemd, dependencies, virt-manager |
| [#092](internal/092-libvirt-package-installation-support.md) | - | Libvirt Package Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, libvirt, virtualization, refactoring |
| [#093](internal/093-spice-server-client-installation.md) | - | Spice Server and Client Installation Support | 📋 Open | High | Enhancement | enhancement, package-management, virtualization, spice, qemu, kvm, clipboard |
| [#094](internal/094-container-rm-subcommand-not-recognized.md) | - | Container 'rm' Subcommand Not Recognized | ✅ Implemented | Medium | Bug Fix | bug, container, cli, command-parsing |
| [#095](internal/095-container-exec-returns-helper-version.md) | - | Container exec Returns Helper Version Instead of Executing Command | ✅ Implemented | High | Bug Fix | bug, container, cli, ptx-container, critical |
| [#096](internal/096-container-start-stop-help-flag-bug.md) | - | Container Start/Stop Commands Misinterpret --help Flag as Container Name | ✅ Implemented | Medium | Bug Fix | bug, container, help-system, user-experience, cli |
| [#097](internal/097-ptx-python-helper-implementation.md) | - | PTX-Python Helper Implementation | 🔄 In Progress (Phase 2 Complete) | High | Feature | enhancement, helper-binary, python, development-tools, build-automation, code-quality |
| [#098](internal/098-ptx-vocalio-helper-implementation.md) | - | PTX-Vocalio Helper Implementation | 📋 Open | High | Feature | enhancement, helper-binary, speech-recognition, text-to-speech, ai-integration, accessibility |
| [#099](internal/099-system-info-performance-optimization.md) | - | System Info Performance Optimization | 📋 Open | High | Enhancement | enhancement, performance, system-info, optimization, user-experience, critical-path |
| [#100](internal/100-ptx-installer-helper-implementation.md) | - | PTX-Installer Helper Implementation | 🔄 In Progress (Phase 4 Complete) | High | Feature | enhancement, architecture, performance, helper-binary, package-management |
| [#101](internal/101-ptx-aiops-helper-implementation.md) | - | PTX-AIOps Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, ai-integration, container, gpu-support |
| [#102](internal/102-compose-command-implementation.md) | - | Compose Command Implementation | ✅ Implemented | High | Enhancement | enhancement, container, docker-compose, podman-compose, universal-interface |
| [#103](internal/103-ptx-make-helper-implementation.md) | - | PTX-Make Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, build-automation, cross-platform, makefile |
| [#104](internal/104-ptx-make-ls-command.md) | - | PTX-Make LS Command Implementation | 📋 Open | Medium | Enhancement | enhancement, helper-binary, ptx-make, cross-platform, file-operations |
| [#105](internal/105-ptx-make-gobuild-cross-platform-compilation.md) | - | PTX-Make GoBuild Cross-Platform Compilation | 📋 Open | High | Enhancement | enhancement, helper-binary, ptx-make, cross-platform, go-compilation |
| [#106](internal/106-install-command-help-flag-not-working.md) | - | Install Command --help Flag Not Working | ✅ Implemented | High | Bug Fix | bug, cli, help-system, install, user-experience, documentation |
| [#107](internal/107-ptx-pft-product-feedback-tool-helper.md) | - | PTX-PFT Product Feedback Tool Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, product-feedback, fider, synchronization |
| [#108](internal/108-ptx-pft-email-notifications.md) | - | PTX-PFT E-mail Notifications for User Actions | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-pft, email, notifications |
| [#109](internal/109-ptx-pft-clearflask-provider.md) | - | PTX-PFT ClearFlask Provider Implementation | ✅ Implemented | Medium | Enhancement | enhancement, helper-binary, product-feedback, clearflask, provider |
| [#110](internal/110-ptx-pft-eververse-provider.md) | - | PTX-PFT Eververse Provider Implementation | ✅ Implemented | Medium | Enhancement | enhancement, helper-binary, product-feedback, eververse, provider, high-complexity |
| [#111](internal/111-ptx-pft-mcp-integration.md) | - | PTX-PFT MCP Integration | 📋 Open | High | Enhancement | enhancement, helper-binary, ptx-pft, ptx-mcp, ai-integration, mcp |
| [#112](internal/112-ptx-pft-category-management.md) | - | PTX-PFT Category Management for UC and Requirements | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-pft, categorization, organization |
| [#113](internal/113-mcp-help-missing-subcommands-v180.md) | - | MCP Help Missing Subcommands in v1.8.0 Release | ✅ Implemented | High | Bug Fix | bug, mcp, release, help-system, regression |
| [#114](internal/114-mcp-configure-default-stdio-mode.md) | - | MCP Configure Should Default to stdio Mode | ✅ Implemented | Medium | Enhancement | mcp, configuration, ux |
| [#115](internal/115-automated-release-notes-generation.md) | - | Automated Release Notes Generation System | ✅ Implemented | High | Enhancement | enhancement, release-process, automation, ai-integration |
| [#116](internal/116-ptx-pft-iso16355-qfd-project-structure.md) | - | PTX-PFT ISO 16355 QFD Project Structure | 📋 Open | High | Enhancement | enhancement, helper-binary, ptx-pft, iso-16355, qfd, requirements-management |
| [#117](internal/117-ptx-pft-list-qfd-compatibility.md) | - | PTX-PFT List QFD Compatibility | 📋 Open | Medium | Enhancement | enhancement, helper-binary, ptx-pft, qfd, compatibility |
| [#118](internal/118-system-info-pprof-profiling.md) | - | System Info pprof Profiling | 📋 Open | Medium | Enhancement | enhancement, performance, profiling, system-info |
| [#119](internal/119-ptx-ansible-standalone-help-and-template-examples.md) | - | PTX-Ansible Standalone Help and Template Examples System | 🔄 In Progress | High | Enhancement | enhancement, helper-binary, ptx-ansible, templates, user-experience, documentation |
| [#120](internal/120-windows-native-system-info-module.md) | - | Windows Native System Info Module | 📋 Open | High | Enhancement | enhancement, performance, windows, system-info, native-api |
| [#121](internal/121-virt-list-libvirt-detection-fix.md) | - | Fix libvirt Detection in portunix virt list | ✅ Implemented | Medium | Bug Fix | bug, ptx-virt, libvirt, detection |
| [#122](internal/122-consolidate-docker-podman-installation.md) | - | Consolidate Docker/Podman Installation into ptx-installer | ✅ Implemented | High | Refactoring | refactoring, architecture, ptx-installer, docker, podman |
| [#123](internal/123-consolidate-installation-systems.md) | - | Consolidate Installation Systems - Remove Duplicate Assets | ✅ Implemented | Medium | Refactoring | refactoring, architecture, ptx-installer, assets |
| [#124](internal/124-download-progress-indicator.md) | - | Download Progress Indicator | ✅ Implemented | Medium | Enhancement | enhancement, user-experience, ptx-installer, download |
| [#125](internal/125-cross-platform-binary-distribution.md) | - | Cross-Platform Binary Distribution | 📋 Open | High | Enhancement | enhancement, architecture, distribution, cross-platform, container, vm |
| [#126](internal/126-gh-installation-package-manager-detection.md) | - | GitHub CLI Installation Package Manager Detection Bug | 🔄 In Progress | High | Bug Fix | bug, package-management, github-cli, arch-linux, detection |
| [#127](internal/127-migrate-openssh-to-ptx-installer.md) | - | Migrate Win32-OpenSSH Installation to ptx-installer | 🔄 In Progress | Medium | Enhancement | enhancement, package-management, openssh, cross-platform, refactoring |
| [#128](internal/128-docusaurus-container-performance-optimization.md) | - | Docusaurus Container Performance Optimization | ✅ Implemented | High | Enhancement | enhancement, container, docker, performance, docusaurus, playbook |
| [#129](internal/129-docusaurus-quickstart-script.md) | - | Docusaurus QuickStart Script for GitHub Release | ✅ Implemented | Medium | Enhancement | enhancement, documentation, user-experience, quickstart, docusaurus, release-assets |
| [#131](internal/131-openssh-reinstall-hostkeys-bug.md) | - | OpenSSH Installation Refactoring - Embedded Script Support | ✅ Implemented | High | Bug Fix | bug, openssh, windows, installation, refactoring |
| [#132](internal/132-text-extractor-plugin-integration.md) | - | Text Extractor Plugin Integration | 📋 Open | High | Feature / Plugin | plugin, text-extraction, java, tika, mcp, ai-integration |
| [#133](internal/133-plugin-run-command-argument-forwarding.md) | - | Plugin Run Command Argument Forwarding | ✅ Implemented | Medium | Enhancement | enhancement, plugin-system, cli, ai-integration, user-experience |
| [#134](internal/134-pft-config-cross-platform-path-support.md) | - | PFT Config Cross-Platform Path Support | 📋 Open | High | Enhancement | bug, enhancement, helper-binary, ptx-pft, cross-platform, configuration |
| [#135](internal/135-pft-assign-show-path-parameter-bug.md) | - | PFT Assign/Show Commands Ignore --path Parameter | ✅ Implemented | High | Bug Fix | bug, helper-binary, ptx-pft, cli, path-parameter, regression |
| [#136](internal/136-ptx-credential-helper-implementation.md) | - | PTX-Credential Helper Implementation | ✅ Implemented | High | Feature | enhancement, helper-binary, security, credential-storage, encryption, cross-platform |
| [#137](internal/137-graalvm-installation-support.md) | - | GraalVM Installation Support | ✅ Implemented | High | Feature | enhancement, package-management, graalvm, java, native-image, cross-platform |
| [#138](internal/138-ptx-python-project-local-venv-support.md) | - | PTX-Python Project-Local Virtual Environment Support | ✅ Implemented | High | Enhancement | enhancement, helper-binary, ptx-python, python, virtual-environment, developer-experience |
| [#139](internal/139-tea-gitea-cli-installation-support.md) | - | Tea (Gitea CLI) Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, gitea, cli, developer-tools, cross-platform |
| [#140](internal/140-version-management-strategy-implementation.md) | - | Version Management Strategy Implementation (ADR-036) | 📋 Open | High | Enhancement | enhancement, versioning, workflow, contributor-experience, adr-implementation |
| [#141](internal/141-ptx-trace-helper-implementation.md) | - | PTX-TRACE Universal Tracing Helper (First: Data Transformation) | 📋 Open | High | Feature / Helper Binary | helper-binary, data-transformation, tracing, etl, debugging, ai-integration, mcp |
| [#142](internal/142-elasticsearch-container-installation.md) | - | Elasticsearch Container Installation | ✅ Implemented | Medium | Feature | installation, container, elasticsearch, fulltext |
| [#144](internal/144-ptx-trace-session-delete-button.md) | - | PTX-TRACE Session Delete Button | ✅ Implemented | Medium | Enhancement | enhancement, ptx-trace, ui, dashboard, session-management |
| [#145](internal/145-ninja-installation-support.md) | - | Ninja Build System Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, ninja, build-system, cross-platform, ptx-installer |
| [#146](internal/146-rust-installation-support.md) | - | Rust Programming Language Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, rust, programming-language, cross-platform, ptx-installer, rustup |
| [#147](internal/147-clang-installation-support.md) | - | Clang/LLVM Installation Support | ✅ Implemented | Medium | Enhancement | enhancement, package-management, clang, llvm, compiler, c-cpp, cross-platform, ptx-installer |
| [#148](internal/148-minio-installation-support.md) | - | MinIO Installation Support | ✅ Implemented | High | Enhancement | enhancement, package-management, minio, object-storage, s3-compatible, cross-platform, ptx-installer, go-install |
| [#149](internal/149-vox-ptx-install-integration.md) | - | Vox Plugin - ptx-install Integration for Model Installation | ✅ Implemented | Medium | Feature | ptx-installer, integration, models, multi-file-download, plugin-support |
| [#150](internal/150-distributed-mcp-server-ecosystem.md) | - | Distributed MCP Server Ecosystem for Plugins | 📋 Open | High | Architecture / Feature | mcp, plugin-system, ai-integration, architecture, distributed |
| [#151](internal/151-playwright-browsers-package.md) | - | Playwright Browsers Package Definition | ✅ Implemented | Medium | Feature | ptx-installer, package-management, playwright, browser-automation, cross-platform |
| [#152](internal/152-plugin-create-missing-go-mod.md) | - | Plugin Create Missing go.mod | ✅ Implemented | Medium | Bug Fix | bug, plugin-system, cli, developer-experience |
| [#153](internal/153-deliver-docker-documentation-environment-for-knife.md) | - | Deliver Docker Documentation Environment for KNIFE Project | 📋 Open | High | Feature | feature, documentation, docker, customer-delivery, knife-project |
| [#154](internal/154-additional-documentation-engines-vitepress-mkdocs.md) | - | Additional Documentation Engines — VitePress, MkDocs | 📋 Open | Medium | Feature | feature, documentation, docker, enhancement |

## Directory Structure

```text
docs/issues/
├── README.md           # This file - main tracking table
├── internal/           # All internal issues (not published to GitHub)
│   ├── 001-*.md
│   ├── 002-*.md
│   └── ...
└── public/            
    └── mapping.json   # Mapping between internal and public issue numbers
```

## Usage

### Creating New Issues

1. **Internal Issue (all types):**
   - Create file: `internal/{next-number}-{short-title}.md`
   - Update this README with issue entry
   - Set Public column to `-` initially

2. **Publishing to GitHub (features/enhancements only):**
   - Assign next PUB- number in mapping.json
   - Update Public column in this README
   - Create GitHub issue with PUB- number
   - Never publish: bugs, security issues, internal tasks

### Issue Types

- **Feature**: New functionality (can be public)
- **Enhancement**: Improvement to existing features (can be public)  
- **Bug Fix**: Fixing broken functionality (internal only)
- **Security**: Security-related issues (internal only)
- **Plugin**: Plugin-specific features (selective public)

### Status Legend

- 📋 Open - Issue is open and needs work
- 🔄 In Progress - Issue is being actively worked on  
- ✅ Implemented - Issue has been completed and implemented
- ❌ Closed - Issue has been closed without implementation
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