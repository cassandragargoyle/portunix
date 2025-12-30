# TODO - Portunix Development Tasks

## Active Issues

### Issue #098: PTX-Vocalio Helper Implementation - Phase 1
**Status**:  Phase 1 Testing Complete (Bug Fixed)
**Priority**: High
**Branch**: `feature/issue-098-ptx-vocalio-helper`
**Last Updated**: 2025-10-24

#### Current Status Summary

**Phase 1: Foundation & Interactive Wizard** -  IMPLEMENTED & TESTED

**Acceptance Testing Results**:
-  TC001: Wizard Configuration Creation - PASS
-  TC002: Python Version Verification (ADR-029) - PASS (critical bug fixed)
-  TC003: Create Command Available - PASS

**Critical Bug Fixed** (2025-10-24):
- **File**: `src/helpers/ptx-vocalio/install.go:146-154`
- **Issue**: Python version handling in `setupVirtualEnvironment()` - user choice was ignored
- **Fix**: Code now checks `pythonCmd` prefix and conditionally adds `--python` flag
- **Verification**: Re-tested with Python 3.13 system, config requiring 3.11 - works correctly

**Implemented Features**:
1.  Helper Binary Infrastructure
   - `ptx-vocalio` binary builds and executes
   - Build pipeline integrated in Makefile
   - Dispatcher integration working

2.  Main Binary Dispatcher
   - `portunix vocalio prepare` - interactive wizard
   - `portunix vocalio install <name>` - install tools and models
   - `portunix vocalio create <name>` - create standalone executable
   - `portunix vocalio config` - configuration management
   - Helper availability check working

3.  Interactive Configuration Wizard
   - All 6 questions implemented and tested:
     1. Vocalio name
     2. Target OS (Linux, Windows, Raspberry Pi, macOS)
     3. Language selection
     4. STT tool selection (OpenAI Whisper, Vosk, etc.)
     5. TTS tool selection (Coqui TTS, Piper, espeak-ng, etc.)
     6. Target system and output type
   - Configuration saved to `~/.portunix/vocalio/<name>.vocalio.yaml`
   - OS-specific recommendations working

4.  Configuration Management
   - `config list` - lists all configurations
   - `config show <name>` - displays formatted config
   - `config validate <file>` - validates YAML structure
   - `config use <name>` - sets default config
   - Multiple configurations supported

5.  Python Version Compatibility (ADR-029)
   - Package definitions include `pythonRequirements`
   - Version resolution algorithm implemented
   - Compatible Python version calculation works
   - Installation-time verification implemented
   - User prompted with 3 options on version mismatch
   - User choice correctly applied (bug fixed)

6.  Install Command
   - YAML configuration parser working
   - Python venv creation via `ptx-python` integration
   - Package installation from embedded definitions
   - Python version verification working
   - Dependencies installation tested

7.  Create Command
   - Command structure implemented
   - Prerequisites verification working
   - Full E2E testing deferred to Phase 2

8.  PTX-Python Integration
   - Venv creation working
   - Package installation working
   - Python version handling working

**Acceptance Protocol**: `docs/testing/internal/acceptance-098-phase1.md`

**Ready for Merge**:  YES
- All critical bugs resolved
- ADR-029 compliance achieved
- Core functionality verified on Linux (Ubuntu 6.14.0-34, Python 3.13.3)

#### Next Steps

**Before Merge to Main**:
- [ ] Review acceptance protocol
- [ ] Verify all commits squashed/cleaned if needed
- [ ] Merge to main branch
- [ ] Close Phase 1 in issue tracker

**Phase 2 Tasks** (Future):
- [ ] Option 1 testing (Install Python via portunix) - full E2E
- [ ] Container-based testing for multiple Python versions (3.8-3.13)
- [ ] Complete E2E testing of create command with PyInstaller
- [ ] Test model downloads and caching
- [ ] Test actual STT/TTS operations (speech recognition, synthesis)
- [ ] Cross-platform testing (Windows, macOS, Raspberry Pi)

**Phase 3 Tasks** (Future):
- [ ] MCP server voice interface integration
- [ ] Real-time streaming support
- [ ] API endpoints for voice operations
- [ ] Advanced model management (update, cache cleanup)

---

### Issue #100: PTX-Installer Helper Implementation
**Status**: ✅ Phase 4 Complete (Phase 5 In Progress)
**Priority**: High
**Branch**: `feature/issue-100-ptx-installer-helper`
**Last Updated**: 2025-10-29

#### Current Status Summary

**Completed Phases**:
- ✅ Phase 1: Helper Binary Skeleton - COMPLETE
- ✅ Phase 2: Core Installation Migration - COMPLETE
- ✅ Phase 2.5: Shared Platform Utilities (ADR-026) - COMPLETE
- ✅ Phase 3: Package Management Commands - COMPLETE
- ✅ Phase 4: Testing and Optimization - COMPLETE

**Current Phase**: Phase 5 - Cleanup and Release

#### Phase 4 Results (2025-10-29)

**Test Coverage**: 31 test cases across 5 categories
**Success Rate**: 93.5% (29/31 full pass, 2 partial requiring Phase 5)

**Test Results**:
- ✅ Performance Testing: Helper binary excellent (20ms commands), Main binary requires Phase 5 cleanup
- ✅ Functional Testing: 15/15 passed (100%)
- ✅ Cross-Platform Testing: 3/3 passed (Linux/x64)
- ✅ Regression Testing: 5/5 passed (100%)
- ✅ Integration Testing: 5/5 passed (100%)

**Helper Binary Performance** (Excellent):
- Command execution: 20ms average ✅
- Binary size: 13MB ✅
- Memory usage: 94MB ✅
- Embedded assets: 34 packages ✅

**Main Binary Performance** (Requires Phase 5):
- System info: 1.42s (target: <50ms) ⚠️
- Binary size: 24MB (target: <20MB) ⚠️
- Memory: 125MB (target: <20MB) ⚠️
- Root cause: Installation subsystem still in main binary

**Real Installation Verified**:
```bash
./ptx-installer install act
✅ Downloaded 6.90 MB
✅ Extracted to /usr/local/bin
✅ Verified: act version 0.2.68
```

**Implemented Features**:
1. ✅ Helper Binary Infrastructure
   - `ptx-installer` binary with embedded assets (34 packages)
   - Dispatcher integration via `portunix install` commands
   - Standalone execution capability

2. ✅ Package Management Commands
   - `package list` - List all packages with filtering
   - `package list --category=<name>` - Filter by category
   - `package list --platform=<name>` - Filter by platform
   - `package list --format=json` - JSON output
   - `package search <query>` - Search across name, description, category
   - `package info <name>` - Comprehensive package details

3. ✅ Installation Commands
   - `install <package>` - Install packages
   - `install <package> --variant=<name>` - Variant selection
   - `install <package> --dry-run` - Preview installation
   - tar.gz extraction working
   - APT package detection working

4. ✅ Shared Platform Utilities (ADR-026)
   - `src/pkg/platform/` package created
   - OS and architecture detection
   - Permission and privilege utilities
   - 12 unit tests (all passing)
   - Code duplication eliminated (~100 lines)

5. ✅ Container Deployment
   - Binary portability verified (host → container)
   - Embedded assets functional in isolation
   - Real installations tested in Ubuntu 22.04 container
   - Podman integration working

**Architecture Decision Records**:
- ADR-025: PTX-Installer Helper Architecture
- ADR-026: Shared Platform Utilities Package

**Acceptance Protocols**:
- Phase 2: `test/integration/acceptance_issue_100_phase2.md`
- Phase 4: `test/integration/acceptance_issue_100_phase4.md`

#### Phase 5 Tasks (In Progress)

**Goal**: Remove legacy code and finalize release v1.7.0

**Critical Tasks**:
1. [ ] Remove installation subsystem from main binary
   - [ ] Remove `src/app/install/` directory
   - [ ] Clean up obsolete imports
   - [ ] Update dispatcher to only route to helper
   - [ ] Remove installation-related dependencies

2. [ ] Performance Validation
   - [ ] Re-benchmark `portunix system info` (target: <50ms)
   - [ ] Measure main binary size (target: <20MB)
   - [ ] Measure memory footprint (target: <20MB RSS)
   - [ ] Verify 24× performance improvement

3. [ ] Build System Updates
   - [ ] Update Makefile for helper building
   - [ ] Update `build-with-version.sh` script
   - [ ] Test build process on all platforms

4. [ ] Documentation
   - [ ] Update user documentation
   - [ ] Update developer documentation
   - [ ] Create migration guide for contributors

5. [ ] Release Preparation
   - [ ] Version bump to v1.7.0
   - [ ] Create release notes
   - [ ] Prepare changelog
   - [ ] Tag release candidate

**Expected Improvements After Phase 5**:
- System info: 1.42s → <0.05s (24× faster)
- Binary size: 24MB → <20MB (17% smaller)
- Memory: 125MB → <20MB (84% reduction)

**Next Steps**:
1. Start Phase 5: Code cleanup
2. Remove installation subsystem from main binary
3. Re-run performance benchmarks
4. Finalize v1.7.0 release

---

### Issue #107: PTX-PFT Product Feedback Tool Helper
**Status**: 🚧 Phase 6 In Progress
**Priority**: Medium
**Branch**: `feature/107-ptx-pft-helper`
**Last Updated**: 2025-12-24

#### Current Status Summary

**Completed Phases**:
- ✅ Phase 1: Helper Binary Skeleton - COMPLETE
- ✅ Phase 2: Project Initialization & Markdown Structure - COMPLETE
- ✅ Phase 3: Sync/Push/Pull Commands - COMPLETE
- ✅ Phase 4: Fider Container Integration - COMPLETE
- ✅ Phase 5: User/Customer Registry - COMPLETE
- 🚧 Phase 6: Advanced Features - IN PROGRESS

**Architecture Decision Record**: `docs/adr/029-ptx-pft-product-feedback-tool-helper.md`

#### Implemented Features

1. ✅ **Helper Binary Infrastructure**
   - `ptx-pft` binary builds and executes
   - Dispatcher integration via `portunix pft` commands
   - Standalone execution capability

2. ✅ **Project Initialization**
   - `pft init` - Initialize new PFT project
   - Directory structure: voc/, vos/, vob/, voe/ (ISO 16355 categories)
   - Default role definitions per category

3. ✅ **Sync Commands**
   - `pft sync` - Bidirectional sync with Fider
   - `pft push` - Push new local items to Fider
   - `pft pull` - Pull new Fider posts to local
   - Duplicate file prevention (FindFileBySlug)

4. ✅ **Fider Container Integration**
   - `pft fider start` - Start Fider + Postgres + Mailhog containers
   - `pft fider stop` - Stop containers
   - `pft fider status` - Container health check
   - EMAIL_SMTP_* configuration fix
   - Mailhog for email capture at localhost:8025

5. ✅ **User/Customer Registry**
   - `pft user add` - Add new user
   - `pft user show <id>` - Show user details
   - `pft user list [--category]` - List users
   - `pft user role <id> <category> <role> [--proxy]` - Assign role
   - `pft user link <id> --fider <fider-id>` - Link external IDs
   - `pft user remove <id>` - Remove user
   - `pft role list <category>` - List available roles
   - `pft role init` - Initialize default role files

6. ✅ **ISO 16355 Role System**
   - VoC (Voice of Customer): customer, customer-admin, customer-support, proxy-customer
   - VoS (Voice of Stakeholder): architect, ceo, cio, dev-lead, developer, facilitator, product-manager, support, support-lead, tech-consultant, tester
   - VoB (Voice of Business): ceo, dev-lead, marketing, sales, support
   - VoE (Voice of Engineer): architect, developer, devops, qa, senior-developer, support, tester
   - Proxy attribute for representatives

7. ✅ **Fider User Synchronization** (NEW)
   - `pft user sync` - Sync users from Fider to local registry
   - `pft user sync --voc` - Sync only VoC Fider users
   - `pft user sync --vos` - Sync only VoS Fider users
   - `pft user sync --dry-run` - Preview sync without changes
   - Auto-link by email matching
   - Default role assignment (VoC: customer, VoS: developer)
   - Fider API: ListUsers(), GetUser() endpoints

#### Test Environment
- Location: `/tmp/pft-demo-test`
- Test users:
  - `admin@local.test` - VoC (customer, proxy), VoS (facilitator)
  - `splichal@tovek.cz` - VoS (dev-lead), VoC (proxy-customer, proxy)
- Fider: http://localhost:3000 (when running)
- Mailhog: http://localhost:8025 (when running)

#### Recent Commits
- `2cfb25f` - feat(pft): implement Phase 5 - User/Customer Registry
- `99a67a6` - docs(adr-029): add user registry with roles, proxy attribute and 4 categories
- `ca3ed92` - fix(pft): prevent duplicate files during sync

#### Next Steps (Phase 6 - Remaining)

**Remaining Features**:
- [ ] Role-based filtering in sync commands
- [ ] User assignment to feedback items
- [ ] Export reports by role/category
- [ ] Advanced conflict resolution

**Integration Testing**:
- [ ] Full E2E workflow test with Fider running
- [ ] Multi-user scenario testing
- [ ] Cross-category role assignment testing

---

### Issue #110: PTX-PFT Eververse Provider Implementation
**Status**: 🚧 Testing Blocked (Docker Hub Rate Limit)
**Priority**: Medium
**Branch**: `feature/110-eververse-provider`
**Last Updated**: 2025-12-24

#### Current Status Summary

**Implementation**: ✅ COMPLETE
**Testing**: ⏸️ BLOCKED - Docker Hub rate limit reached

#### Implemented Features

1. ✅ **EververseProvider** (`eververse_provider.go`)
   - Implements FeedbackProvider interface
   - Supabase REST API client with JWT authentication
   - Connect, List, Get, Create, Update, Delete, Close methods
   - Maps Eververse features/feedback to FeedbackItem struct

2. ✅ **Eververse Deploy** (`eververse_deploy.go`)
   - Full Supabase self-hosted stack (12 containers)
   - Resource validation (~6GB RAM requirement)
   - Automatic Eververse image build from GitHub source
   - `pull_policy: never` for local image handling
   - Kong API Gateway configuration generation

3. ✅ **Package Definition** (`assets/packages/eververse.json`)
   - Complete 12-service Supabase stack configuration
   - `pullPolicy: never` for eververse service
   - Resource requirements documentation
   - Known issues and comparison with Fider/ClearFlask

4. ✅ **Config Updates** (`config.go`)
   - SupabaseConfig struct added
   - Supabase options in GetProviderConfig()

5. ✅ **Main CLI Updates** (`main.go`)
   - Switch cases for eververse in deploy, status, destroy commands

6. ✅ **Deploy Structs** (`deploy.go`)
   - PullPolicy field added to ServiceSpec and ComposeService
   - generateComposeYAML() copies pullPolicy

#### Commits
- `fa58532` - feat(110): implement Eververse provider for ptx-pft

#### Testing Blockers

**Docker Hub Rate Limit**:
```
Error: toomanyrequests: You have reached your unauthenticated pull rate limit
```

**Solution Required**:
```bash
# Login to Docker Hub to increase limit (100 → 200 pulls/6h)
podman login docker.io
```

Or wait 6 hours for rate limit reset.

#### Test Configuration
- Location: `/tmp/test-eververse/.pft-config.json`
- Compose: `/home/zdenek/.portunix/pft/eververse/docker-compose.yaml`
- Eververse image: `portunix/eververse:latest` (built locally)

#### Verified Working
- ✅ Eververse image build from GitHub
- ✅ `pull_policy: never` correctly generated in compose
- ✅ Eververse service skipped during pull (confirmed: "eververse Skipped")
- ⏸️ Full stack startup - waiting for Supabase images

#### Next Steps (After Rate Limit Reset)
1. [ ] `podman login docker.io` (increase rate limit)
2. [ ] Run deploy again: `portunix pft deploy`
3. [ ] Wait 60-120s for full stack startup
4. [ ] Test provider API operations (list, create, get)
5. [ ] Update acceptance protocol
6. [ ] Merge to main

---

## Other Active Tasks

(Add other TODO items here as needed)

---

**Last Updated**: 2025-12-25
**Maintained by**: Development Team
