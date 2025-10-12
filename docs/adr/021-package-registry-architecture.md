# ADR-021: Package Registry Architecture

**Status:** Proposed
**Date:** 2025-09-27
**Architect:** Claude (AI Assistant)

## Context

The current `assets/install-packages.json` file has become unmanageably large and complex:

- **Size:** 107KB, 2637 lines
- **Packages:** 33 software packages with complex nested structures
- **Complexity:** Deep nesting for platforms, architectures, and variants
- **Maintenance:** Manual updates of versions and URLs are error-prone
- **AI Integration:** 10 packages already have AI prompts for version discovery
- **Scalability:** Adding new packages becomes increasingly difficult

### Current Structure Problems

1. **Monolithic File:** Single JSON file containing all package definitions
2. **Data Duplication:** Repeated patterns across similar packages
3. **Manual Maintenance:** Version updates require manual file editing
4. **Complex Nesting:** Deep hierarchical structures difficult to navigate
5. **No Validation:** Schema validation happens only at runtime
6. **Limited Extensibility:** Hard to add new package types or metadata

### Example Current Structure Complexity
```json
{
  "packages": {
    "java": {
      "platforms": {
        "windows": {
          "variants": {
            "21": {
              "version": "21.0.8_9",
              "urls": {
                "x64": "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.8%2B9/OpenJDK21U-jdk_x64_windows_hotspot_21.0.8_9.msi",
                "x86": "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.8%2B9/OpenJDK21U-jdk_x86_windows_hotspot_21.0.8_9.msi"
              }
            }
          }
        }
      }
    }
  }
}
```

## Decision

Implement a **Distributed Package Registry Architecture** with the following components:

### 1. Package Registry Structure

```
assets/
├── packages/                    # Individual package definitions
│   ├── java.json               # Complete Java package definition
│   ├── python.json             # Complete Python package definition
│   ├── nodejs.json             # Complete Node.js package definition
│   ├── vscode.json             # Complete VS Code package definition
│   ├── chrome.json             # Complete Chrome package definition
│   └── ...                     # One file per package
├── templates/                   # Package type templates (optional)
│   ├── msi-installer.json      # Template for MSI packages
│   ├── tar-archive.json        # Template for tar.gz packages
│   └── github-release.json     # Template for GitHub releases
└── registry/
    ├── index.json              # Package registry index
    └── categories.json         # Package categories
```

### 2. Package Definition Format

#### Complete Package Definition (`java.json`)
```json
{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "java",
    "displayName": "Java (OpenJDK)",
    "description": "Java Development Kit from Eclipse Adoptium",
    "category": "development/languages",
    "homepage": "https://adoptium.net/",
    "license": "GPL-2.0-with-classpath-exception",
    "maintainer": "Eclipse Adoptium"
  },
  "spec": {
    "hasVariants": true,
    "defaultVariant": "21",
    "aiPrompts": {
      "packageResearch": "Research Eclipse Adoptium Temurin OpenJDK releases. Focus on LTS versions (8, 11, 17, 21). Check GitHub releases at adoptium/temurin*-binaries repositories.",
      "versionDiscovery": "Check GitHub API for latest releases in adoptium/temurin{8,11,17,21}-binaries repositories. Parse release tags to extract version numbers."
    },
    "metadataUrls": {
      "documentation": "https://adoptium.net/temurin/releases/",
      "releases": "https://api.github.com/repos/adoptium/temurin21-binaries/releases/latest"
    },
    "verification": {
      "command": "java -version",
      "expectedExitCode": 0
    },
    "platforms": {
      "windows": {
        "installer": {
          "type": "msi",
          "args": ["ADDLOCAL=ALL", "/quiet"]
        },
        "environment": {
          "JAVA_HOME": "${install_path}",
          "PATH_APPEND": "${install_path}/bin"
        },
        "installPath": "${ProgramFiles}/Eclipse Adoptium/jdk-${version}-hotspot"
      },
      "linux": {
        "installer": {
          "type": "tar.gz",
          "extractTo": "/opt/java/jdk-${version}"
        },
        "postInstall": [
          "sudo update-alternatives --install /usr/bin/java java ${install_path}/bin/java 1",
          "sudo update-alternatives --install /usr/bin/javac javac ${install_path}/bin/javac 1"
        ]
      }
    },
    "variants": {
      "21": {
        "version": "21.0.8_9",
        "stability": "stable",
        "lts": true,
        "sources": {
          "windows": {
            "x64": {
              "url": "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.8%2B9/OpenJDK21U-jdk_x64_windows_hotspot_21.0.8_9.msi",
              "checksum": "sha256:..."
            },
            "x86": {
              "url": "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.8%2B9/OpenJDK21U-jdk_x86_windows_hotspot_21.0.8_9.msi",
              "checksum": "sha256:..."
            }
          },
          "linux": {
            "x64": {
              "url": "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.8%2B9/OpenJDK21U-jdk_x64_linux_hotspot_21.0.8_9.tar.gz",
              "checksum": "sha256:..."
            },
            "arm64": {
              "url": "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.8%2B9/OpenJDK21U-jdk_aarch64_linux_hotspot_21.0.8_9.tar.gz",
              "checksum": "sha256:..."
            }
          }
        }
      },
      "17": {
        "version": "17.0.16_8",
        "stability": "stable",
        "lts": true,
        "sources": {
          "windows": {
            "x64": {
              "url": "https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.16%2B8/OpenJDK17U-jdk_x64_windows_hotspot_17.0.16_8.msi",
              "checksum": "sha256:..."
            }
          },
          "linux": {
            "x64": {
              "url": "https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.16%2B8/OpenJDK17U-jdk_x64_linux_hotspot_17.0.16_8.tar.gz",
              "checksum": "sha256:..."
            }
          }
        }
      }
    }
  }
}
```

#### Simple Package Example (`nodejs.json`)
```json
{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "nodejs",
    "displayName": "Node.js",
    "description": "JavaScript runtime built on Chrome's V8 JavaScript engine",
    "category": "development/languages"
  },
  "spec": {
    "hasVariants": false,
    "platforms": {
      "windows": {
        "installer": {
          "type": "msi",
          "args": ["/quiet"]
        }
      },
      "linux": {
        "installer": {
          "type": "tar.gz",
          "extractTo": "/opt/nodejs"
        }
      }
    },
    "sources": {
      "windows": {
        "x64": {
          "url": "https://nodejs.org/dist/v20.18.0/node-v20.18.0-x64.msi",
          "checksum": "sha256:..."
        }
      },
      "linux": {
        "x64": {
          "url": "https://nodejs.org/dist/v20.18.0/node-v20.18.0-linux-x64.tar.xz",
          "checksum": "sha256:..."
        },
        "arm64": {
          "url": "https://nodejs.org/dist/v20.18.0/node-v20.18.0-linux-arm64.tar.xz",
          "checksum": "sha256:..."
        }
      }
    }
  }
}
```

### 3. Registry Management System

#### Registry Index (`registry/index.json`)
```json
{
  "apiVersion": "v1",
  "kind": "Registry",
  "metadata": {
    "version": "2.0",
    "lastUpdated": "2025-09-27T10:00:00Z",
    "totalPackages": 33
  },
  "packages": [
    {
      "name": "java",
      "category": "development/languages",
      "hasVariants": true,
      "variants": ["8", "11", "17", "21"],
      "platforms": ["windows", "linux"],
      "status": "stable"
    },
    {
      "name": "python",
      "category": "development/languages",
      "hasVariants": true,
      "variants": ["3.11", "3.12", "3.13"],
      "platforms": ["windows", "linux"],
      "status": "stable"
    },
    {
      "name": "nodejs",
      "category": "development/languages",
      "hasVariants": false,
      "platforms": ["windows", "linux"],
      "status": "stable"
    }
  ],
  "categories": [
    {
      "id": "development/languages",
      "name": "Programming Languages",
      "description": "Programming language runtimes and SDKs"
    },
    {
      "id": "development/tools",
      "name": "Development Tools",
      "description": "IDEs, editors, and development utilities"
    },
    {
      "id": "system/containers",
      "name": "Container Tools",
      "description": "Docker, Podman, and container management"
    }
  ]
}
```

### 4. Template System

#### MSI Installer Template (`templates/msi-installer.json`)
```json
{
  "apiVersion": "v1",
  "kind": "Template",
  "metadata": {
    "name": "msi-installer",
    "description": "Template for Windows MSI installer packages"
  },
  "spec": {
    "installer": {
      "type": "msi",
      "defaultArgs": ["/quiet"]
    },
    "validation": {
      "fileExtensions": [".msi"],
      "requiredFields": [
        "url",
        "checksum",
        "installPath"
      ]
    },
    "environment": {
      "pathHandling": "append",
      "variableExpansion": true
    },
    "verification": {
      "methods": ["command", "registry", "file"]
    }
  }
}
```

### 5. Dynamic Loading System

The package registry will be loaded dynamically at runtime:

```go
type PackageRegistry struct {
    packages    map[string]*Package
    templates   map[string]*Template  // optional
    categories  map[string]*Category
    index       *RegistryIndex
}

type Package struct {
    Metadata    PackageMetadata
    Spec        PackageSpec
}

func LoadPackageRegistry(registryPath string) (*PackageRegistry, error) {
    // Load registry index
    index, err := loadRegistryIndex(filepath.Join(registryPath, "registry/index.json"))
    if err != nil {
        return nil, err
    }

    // Load all packages from individual JSON files
    packages := make(map[string]*Package)
    packageFiles, err := filepath.Glob(filepath.Join(registryPath, "packages", "*.json"))
    if err != nil {
        return nil, err
    }

    for _, packageFile := range packageFiles {
        packageName := strings.TrimSuffix(filepath.Base(packageFile), ".json")
        package, err := loadPackageFromFile(packageFile)
        if err != nil {
            return nil, fmt.Errorf("failed to load package %s: %w", packageName, err)
        }
        packages[packageName] = package
    }

    return &PackageRegistry{
        packages: packages,
        index:    index,
    }, nil
}
```

### 6. Migration Strategy

#### Phase 1: Create Registry Structure
1. Create new directory structure in `assets/`
2. Split current `install-packages.json` into individual package files
3. Implement registry loader in Go
4. Maintain backward compatibility with existing JSON

#### Phase 2: Template System
1. Create package type templates
2. Implement template-based package generation
3. Migrate packages to use templates
4. Add schema validation

#### Phase 3: AI Integration Enhancement
1. Expand AI prompts to all packages
2. Implement automatic version discovery
3. Add metadata URL tracking
4. Create update automation

#### Phase 4: Advanced Features
1. Package dependency management
2. Automatic checksum verification
3. Package signing and verification
4. Remote package registries

### 7. Benefits of New Architecture

#### Developer Experience
- **Modular Structure:** Easy to find and edit specific packages
- **Template Reuse:** Common patterns shared across packages
- **Schema Validation:** Catch errors early with proper validation
- **Git Workflow:** Better diff and merge experience with separate files

#### Maintenance
- **Automated Updates:** AI-driven version discovery and updates
- **Reduced Duplication:** Templates eliminate repeated code
- **Clear Structure:** Organized by logical separation of concerns
- **Version Control:** Individual file tracking for better change history

#### Extensibility
- **Plugin Architecture:** Easy to add new package types
- **Category System:** Organized package discovery
- **Dependency Management:** Cross-package relationships
- **Remote Registries:** Support for external package sources

#### Performance
- **Lazy Loading:** Load only needed packages
- **Caching:** Better caching strategies for package metadata
- **Parallel Processing:** Concurrent package loading and validation
- **Memory Efficiency:** Reduced memory footprint

## Consequences

### Positive
1. **Maintainability:** Much easier to maintain individual package files
2. **Scalability:** Can handle hundreds of packages without complexity
3. **Collaboration:** Multiple developers can work on different packages simultaneously
4. **Automation:** AI can automatically update package versions and metadata
5. **Validation:** Schema validation prevents configuration errors
6. **Performance:** Faster loading and processing of package definitions

### Negative
1. **Migration Effort:** Significant work to migrate existing packages
2. **Complexity:** More complex initial setup and understanding
3. **File Count:** Many more files to manage in the repository
4. **Backward Compatibility:** Need to maintain compatibility during transition

### Mitigation Strategies
1. **Gradual Migration:** Implement new system alongside existing one
2. **Tooling:** Create tools to help with package creation and validation
3. **Documentation:** Comprehensive documentation and examples
4. **Testing:** Extensive testing to ensure compatibility

## Implementation Timeline

### Phase 1 (Weeks 1-2): Foundation
- Create directory structure
- Implement basic registry loader
- Migrate 5 simple packages (nodejs, python, go, vscode, chrome)

### Phase 2 (Weeks 3-4): Complex Packages
- Migrate complex packages (java with variants)
- Implement template system
- Add schema validation

### Phase 3 (Weeks 5-6): AI Integration
- Enhance AI prompts for all packages
- Implement automatic version discovery
- Add metadata URL tracking

### Phase 4 (Weeks 7-8): Advanced Features
- Package dependency management
- Automatic checksum verification
- Remote registry support
- Complete migration and deprecate old system

## Related ADRs
- ADR-013: Software Manifests System
- ADR-019: Package Metadata URL Tracking
- ADR-020: AI Prompts for Package Discovery

---

## Product Owner Decision

**Status: APPROVED WITH RECOMMENDATIONS**
**Date:** 2025-09-27
**Product Owner:** Claude (AI Assistant)

### Business Value Assessment ⭐⭐⭐⭐⭐

This proposal represents a **strategically important step** for Portunix development. The architecture has the potential to significantly improve user experience and project maintainability.

### Key Positives from Product Perspective:

#### 1. **Scalability** 🚀
- Current 33 packages in single file is at maintainability limit
- Distributed architecture enables growth to hundreds of packages
- **Business Impact**: Enables faster addition of new technologies

#### 2. **Developer Experience** 👨‍💻
- Modularity simplifies community contributions
- Better Git workflow for multiple concurrent developers
- **Business Impact**: Faster development, more contributors

#### 3. **Automation** 🤖
- AI-driven version updates reduce manual work
- **Business Impact**: Lower operational costs, more current packages

#### 4. **Reliability** ✅
- Schema validation prevents configuration errors
- Checksum verification increases security
- **Business Impact**: Fewer support issues, higher user trust

### Acceptance Criteria:

#### Must Have (MVP):
1. **Backward Compatibility**: Existing `portunix install` commands must work without changes
2. **Zero-Downtime Migration**: Gradual migration without functionality interruption
3. **Performance**: Package loading must not be slower than current state
4. **Error Handling**: Clear error messages for registry issues

#### Should Have:
1. **Schema Validation**: Automatic validation during package loading
2. **AI Prompts**: Extended AI prompts for all key packages
3. **Template System**: Functional templates for MSI and tar.gz packages

#### Could Have:
1. **Remote Registry**: Support for external registries (for enterprise)
2. **Dependency Management**: Automatic dependency resolution

### Implementation Recommendations:

#### 1. **User Story Priority**:
```
Epic: Package Registry Refactoring
├── US1: Migrate top 10 most used packages (HIGH)
├── US2: Implement basic template system (HIGH)
├── US3: Add AI prompts for auto-updates (MEDIUM)
├── US4: Schema validation system (MEDIUM)
└── US5: Remote registry support (LOW)
```

#### 2. **Risk Management**:
- **Risk**: Migration could break existing installations
- **Mitigation**: Dual-mode support during transition period
- **Acceptance Test**: All existing `install-packages.json` tests must pass

#### 3. **Timeline Adjustment**:
- **Architect Proposes**: 8 weeks
- **PO Recommends**: 6 weeks with focus on core functionality
- **Reasoning**: Faster value delivery, iterative improvements

### MVP Scope Definition:

**Release 1.6.0 Target**:
- [x] Basic registry structure (`assets/packages/`)
- [x] Migration of 10 most-used packages
- [x] Backward compatibility with `install-packages.json`
- [x] Basic template system for MSI and tar.gz
- [x] Schema validation for new packages

**Release 1.7.0 Target**:
- [x] AI prompts for version auto-discovery
- [x] Complete migration of all packages
- [x] Deprecation warning for old system

### Business Metrics for Success:

1. **Package Addition Speed**: Time to add new package < 30 minutes
2. **Update Automation**: 80% of versions updated automatically by AI
3. **Error Reduction**: 50% fewer support issues with package installation
4. **Developer Productivity**: 3x faster onboarding of new packages

### Final Decision:

**✅ APPROVED** with condition of adhering to MVP scope and focus on user value first.

**Next Steps**:
1. Create user stories for each component
2. Start implementation with Phase 1 (Foundation)
3. Weekly review with Product Owner during implementation

**Priority**: **HIGH** - critical for project scalability

---

**Decision Status:** Approved by Product Owner