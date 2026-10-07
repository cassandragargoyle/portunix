# Issue #137: GraalVM Installation Support

## Summary
Add GraalVM JDK installation support to Portunix to enable high-performance polyglot runtime environments with native image compilation capabilities.

## Problem Description

Attempting to install GraalVM via Portunix fails with package not found error:

**Current Error:**
```
$ portunix install graalvm --variant=21
Embedded package discovery complete: 39 packages loaded, 0 errors
Package registry loaded from embedded assets

🔧 Installing package: graalvm

❌ Installation failed: package not found: package 'graalvm' not found
Error: exit status 1
```

**Use Cases:**
- High-performance Java applications with native image compilation
- Polyglot development (Java, JavaScript, Python, Ruby, R, LLVM languages)
- Microservices with fast startup and low memory footprint
- Cloud-native applications optimized for container environments

## Expected Behavior

Portunix should provide GraalVM installation through its unified install command:

```bash
# Install GraalVM with specific version
portunix install graalvm --variant=21
portunix install graalvm --variant=17

# Install with native-image component
portunix install graalvm --variant=21 --components=native-image
```

## Current Available Java Packages

Portunix currently supports standard OpenJDK:
- ✅ `java` - Eclipse Adoptium OpenJDK (8, 11, 17, 21)
- ❌ `graalvm` - **MISSING**

## Impact

**Blocked Use Cases:**
1. **Native Image Compilation** - Cannot build native executables from Java applications
2. **Polyglot Development** - Missing GraalVM's multi-language support
3. **Performance Optimization** - Cannot leverage GraalVM's advanced JIT compiler
4. **Cloud-Native Java** - Missing key tool for building efficient container images

## Proposed Solution

### 1. Package Definition

Create `src/helpers/ptx-installer/assets/packages/graalvm.json`:

```json
{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "graalvm",
    "displayName": "GraalVM",
    "description": "High-performance JDK with native image compilation and polyglot support",
    "category": "development/languages",
    "homepage": "https://www.graalvm.org/",
    "license": "GPL-2.0 with Classpath Exception",
    "maintainer": "Oracle"
  },
  "spec": {
    "hasVariants": true,
    "platforms": {
      "windows": {
        "type": "zip",
        "variants": {
          "17": {
            "version": "17.0.x",
            "urls": {
              "x64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-17.x.x/graalvm-community-jdk-17.x.x_windows-x64_bin.zip"
            },
            "extractTo": "${ProgramFiles}/GraalVM/graalvm-jdk-17"
          },
          "21": {
            "version": "21.0.x",
            "urls": {
              "x64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-21.x.x/graalvm-community-jdk-21.x.x_windows-x64_bin.zip"
            },
            "extractTo": "${ProgramFiles}/GraalVM/graalvm-jdk-21"
          },
          "23": {
            "version": "23.0.x",
            "urls": {
              "x64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-23.x.x/graalvm-community-jdk-23.x.x_windows-x64_bin.zip"
            },
            "extractTo": "${ProgramFiles}/GraalVM/graalvm-jdk-23"
          }
        },
        "verification": {
          "command": "java -version",
          "expectedOutput": "GraalVM"
        },
        "environment": {
          "GRAALVM_HOME": "${extract_to}",
          "JAVA_HOME": "${extract_to}",
          "PATH_APPEND": "${extract_to}/bin"
        }
      },
      "linux": {
        "type": "tar.gz",
        "variants": {
          "17": {
            "version": "17.0.x",
            "urls": {
              "x64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-17.x.x/graalvm-community-jdk-17.x.x_linux-x64_bin.tar.gz",
              "arm64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-17.x.x/graalvm-community-jdk-17.x.x_linux-aarch64_bin.tar.gz"
            },
            "extractTo": "/opt/graalvm/graalvm-jdk-17",
            "postInstall": [
              "sudo update-alternatives --install /usr/bin/java java /opt/graalvm/graalvm-jdk-17/bin/java 2",
              "sudo update-alternatives --install /usr/bin/javac javac /opt/graalvm/graalvm-jdk-17/bin/javac 2",
              "sudo update-alternatives --install /usr/bin/native-image native-image /opt/graalvm/graalvm-jdk-17/bin/native-image 2"
            ]
          },
          "21": {
            "version": "21.0.x",
            "urls": {
              "x64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-21.x.x/graalvm-community-jdk-21.x.x_linux-x64_bin.tar.gz",
              "arm64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-21.x.x/graalvm-community-jdk-21.x.x_linux-aarch64_bin.tar.gz"
            },
            "extractTo": "/opt/graalvm/graalvm-jdk-21",
            "postInstall": [
              "sudo update-alternatives --install /usr/bin/java java /opt/graalvm/graalvm-jdk-21/bin/java 2",
              "sudo update-alternatives --install /usr/bin/javac javac /opt/graalvm/graalvm-jdk-21/bin/javac 2",
              "sudo update-alternatives --install /usr/bin/native-image native-image /opt/graalvm/graalvm-jdk-21/bin/native-image 2"
            ]
          },
          "23": {
            "version": "23.0.x",
            "urls": {
              "x64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-23.x.x/graalvm-community-jdk-23.x.x_linux-x64_bin.tar.gz",
              "arm64": "https://github.com/graalvm/graalvm-ce-builds/releases/download/jdk-23.x.x/graalvm-community-jdk-23.x.x_linux-aarch64_bin.tar.gz"
            },
            "extractTo": "/opt/graalvm/graalvm-jdk-23",
            "postInstall": [
              "sudo update-alternatives --install /usr/bin/java java /opt/graalvm/graalvm-jdk-23/bin/java 2",
              "sudo update-alternatives --install /usr/bin/javac javac /opt/graalvm/graalvm-jdk-23/bin/javac 2",
              "sudo update-alternatives --install /usr/bin/native-image native-image /opt/graalvm/graalvm-jdk-23/bin/native-image 2"
            ]
          }
        },
        "verification": {
          "command": "java -version",
          "expectedOutput": "GraalVM"
        },
        "environment": {
          "GRAALVM_HOME": "${extract_to}",
          "JAVA_HOME": "${extract_to}",
          "PATH_APPEND": "${extract_to}/bin"
        }
      }
    },
    "sources": {
      "graalvm-ce": {
        "type": "github",
        "url": "https://github.com/graalvm/graalvm-ce-builds",
        "apiEndpoint": "https://api.github.com/repos/graalvm/graalvm-ce-builds/releases",
        "pattern": "jdk-{version}"
      }
    },
    "aiPrompts": {
      "versionDiscovery": "Check GitHub API for latest releases in graalvm/graalvm-ce-builds repository. Parse release tags to extract version numbers (format: jdk-X.Y.Z). Focus on LTS releases (17, 21) and latest (23). Verify download assets include x64 for Windows, x64 and aarch64 for Linux. Use GitHub API: https://api.github.com/repos/graalvm/graalvm-ce-builds/releases",
      "urlResolution": "For Windows: graalvm-community-jdk-{version}_windows-x64_bin.zip. For Linux: graalvm-community-jdk-{version}_linux-{arch}_bin.tar.gz. Replace {version} with full version and {arch} with x64/aarch64.",
      "updateGuidance": "Research available GraalVM Community Edition releases. Focus on LTS releases (17, 21) and latest stable (23). For each version, verify both x64 and ARM64 binaries are available. Check that native-image is included in distribution."
    },
    "dependencies": [],
    "templates": ["tar-archive"]
  }
}
```

### 2. Version Support

| Variant | Version | Support Level |
|---------|---------|---------------|
| 17 | 17.0.x LTS | Long-term support |
| 21 | 21.0.x LTS | Long-term support (recommended) |
| 23 | 23.0.x | Latest stable |

### 3. Architecture Support

| Platform | x64 | ARM64 |
|----------|-----|-------|
| Windows | ✅ | ❌ |
| Linux | ✅ | ✅ |
| macOS | ✅ | ✅ |

### 4. Native Image Component

GraalVM Community Edition 21+ includes native-image by default. For older versions or additional components, consider future enhancement for component management:

```bash
# Future enhancement
portunix install graalvm --variant=21 --components=native-image,js
```

## Acceptance Criteria

- [ ] `portunix install graalvm` works on Windows/Linux
- [ ] `portunix install graalvm --variant=21` installs GraalVM 21 LTS
- [ ] `portunix install graalvm --variant=17` installs GraalVM 17 LTS
- [ ] `portunix install graalvm --variant=23` installs GraalVM 23
- [ ] Default variant is `21` (current LTS)
- [ ] Installation verified with `java -version` showing "GraalVM"
- [ ] `native-image --version` works after installation (GraalVM 21+)
- [ ] GRAALVM_HOME and JAVA_HOME environment variables set correctly
- [ ] Cross-platform compatibility (Windows x64, Linux x64/ARM64)
- [ ] Package registered in registry index

## Files to Create/Modify

1. **Create** `src/helpers/ptx-installer/assets/packages/graalvm.json` - Package definition
2. **Modify** `src/helpers/ptx-installer/assets/registry/index.json` - Add "graalvm" to packages list
3. **Modify** `docs/FEATURES_OVERVIEW.md` - Document GraalVM support

## Technical Considerations

### GraalVM vs OpenJDK Coexistence

Both `java` (OpenJDK) and `graalvm` packages use update-alternatives. The priority values should be:
- OpenJDK: priority 1
- GraalVM: priority 2 (higher = preferred)

Users can switch between them using:
```bash
sudo update-alternatives --config java
```

### Native Image Prerequisites

On Linux, native-image may require additional dependencies:
- `build-essential` (gcc, make)
- `zlib1g-dev`

Consider adding these as optional dependencies or documenting in help output.

## Priority

**HIGH** - GraalVM is increasingly popular for cloud-native Java development and native compilation

## Test Cases

After implementation, this should work:

```bash
# Test GraalVM installation with default variant (21)
portunix install graalvm
java -version
# Expected: OpenJDK Runtime Environment GraalVM CE 21.0.x

# Test native-image
native-image --version
# Expected: GraalVM version info

# Test specific variant
portunix install graalvm --variant=17
java -version
# Expected: OpenJDK Runtime Environment GraalVM CE 17.0.x

# Container testing
portunix container run ubuntu:22.04
./portunix install graalvm --variant=21
java -version  # Should show GraalVM
native-image --version  # Should work
```

## References

- GraalVM Downloads: https://www.graalvm.org/downloads/
- GraalVM GitHub Releases: https://github.com/graalvm/graalvm-ce-builds/releases
- Native Image Documentation: https://www.graalvm.org/latest/reference-manual/native-image/

---

**Reporter:** User
**Date:** 2025-01-20
**Category:** Feature Request
**Priority:** High
**Labels:** enhancement, package-management, graalvm, java, native-image, cross-platform
**Status:** ✅ Implemented
