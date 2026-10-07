# Issue #132: Text Extractor Plugin Integration

## Overview
**Title**: Text Extractor Plugin Integration
**Status**: ✅ Implemented
**Priority**: High
**Type**: Feature / Plugin
**Labels**: plugin, text-extraction, java, tika, mcp, ai-integration

## Problem Description

The Text Extractor plugin has been developed in the `portunix-plugins` repository and needs to be integrated into the Portunix plugin ecosystem. The plugin provides comprehensive text extraction capabilities from various document formats using Apache Tika.

**Plugin Location**: `portunix-plugins/plugins/text-extractor`
**Built JAR**: `dist/text-extractor/text-extractor-1.0.1.jar`

## Plugin Technical Details

### Technology Stack
- **Language**: Java 21
- **Build System**: Maven
- **Core Library**: Apache Tika 3.2.3
- **Communication**: gRPC (grpc-netty-shaded 1.68.0)
- **CLI Framework**: Picocli 4.7.6
- **PDF Processing**: Apache PDFBox 3.0.5

### Supported Document Formats
- PDF (native and scanned with OCR)
- Microsoft Office (DOCX, XLSX, PPTX)
- OpenDocument (ODT, ODS, ODP)
- HTML, XML, Markdown
- RTF, LaTeX
- Archives (ZIP, TAR, 7Z)
- Images (with OCR support via Tesseract)
- E-books (EPUB, MOBI)

### Main Class
`org.cassandragargoyle.portunix.textextractor.Main`

## Requirements

### 1. Plugin Manifest Format Migration (YAML → JSON)

Migrate plugin manifest format from `plugin.yaml` to `plugin.json` for consistency with registry.json and better Go native support.

**Rationale:**

- Registry already uses JSON (`registry.json`)
- Native Go support (`encoding/json`) without external dependencies
- Stricter format - fewer parsing errors
- Existing structs already have `json` tags

**Changes required:**

- [ ] Update `manifest.go` - parse JSON instead of YAML
- [ ] Remove `gopkg.in/yaml.v3` dependency
- [ ] Update `cmd/plugin.go` - look for `plugin.json`
- [ ] Update templates and documentation

### 2. Plugin Manifest Creation

Create `plugin.json` manifest file for the text-extractor plugin:

```json
{
  "name": "text-extractor",
  "version": "1.0.1",
  "description": "Advanced text extraction plugin using Apache Tika",
  "author": "CassandraGargoyle Team",
  "license": "Apache-2.0",
  "plugin": {
    "type": "grpc",
    "binary": "text-extractor-1.0.1.jar",
    "runtime": "java",
    "runtime_version": ">=21",
    "jvm_args": ["-Xmx512m", "-Xms128m"],
    "port": 9010,
    "health_check_interval": "30s"
  },
  "dependencies": {
    "portunix_min_version": "1.10.0",
    "os_support": ["linux", "windows", "darwin"],
    "optional_tools": [
      {"name": "tesseract", "reason": "OCR for scanned documents"},
      {"name": "exiftool", "reason": "Enhanced metadata extraction"}
    ]
  },
  "ai_integration": {
    "mcp_tools": [
      {"name": "extract_document_text", "description": "Extract text from documents (PDF, Office, etc.)"},
      {"name": "analyze_document_content", "description": "Analyze document content and metadata"},
      {"name": "extract_structured_data", "description": "Extract tables and structured data from documents"}
    ]
  },
  "permissions": {
    "filesystem": ["read"],
    "network": ["outbound"],
    "level": "limited"
  },
  "commands": [
    {
      "name": "extract",
      "description": "Text extraction commands",
      "subcommands": ["text", "tables", "metadata", "batch"]
    }
  ]
}
```

### 3. Java Plugin Support in Portunix

Extend Portunix plugin system to support Java-based plugins:

- [ ] Detect Java runtime availability
- [ ] Launch JAR files with appropriate JVM arguments
- [ ] Handle Java plugin lifecycle (start/stop/health)
- [ ] Configure JVM memory settings
- [ ] Support gRPC communication with Java plugins

### 4. Plugin Distribution

- [ ] Create platform-independent distribution package
- [ ] Include JAR and plugin.json in release
- [ ] Add to portunix-plugins repository releases
- [ ] Create installation documentation

### 5. MCP Integration

Expose plugin functionality through MCP tools:

- [ ] `extract_document_text` - Basic text extraction
- [ ] `analyze_document_content` - Content analysis
- [ ] `extract_structured_data` - Table/form extraction

### 6. CLI Commands

Register plugin commands with Portunix:

```bash
portunix extract text <file>           # Extract text from document
portunix extract tables <file>         # Extract tables
portunix extract metadata <file>       # Extract metadata
portunix extract batch <directory>     # Batch processing
```

## Implementation Plan

### Phase 1: Plugin Manifest Format Migration

1. Update manifest.go - switch from YAML to JSON parsing
2. Remove gopkg.in/yaml.v3 dependency
3. Update cmd/plugin.go - look for plugin.json
4. Update templates and test files

### Phase 2: Java Plugin Support

1. Extend PluginConfig for Java plugins (runtime, runtime_version, jvm_args)
2. Implement Java plugin launcher
3. Add JVM configuration options
4. Handle Java-specific health checks

### Phase 3: Plugin Manifest and Distribution

1. Create plugin.json manifest for text-extractor
2. Package JAR with manifest
3. Test manual installation

### Phase 4: Integration Testing

1. Test plugin installation via `portunix plugin install`
2. Test gRPC communication
3. Test CLI command routing
4. Test MCP tool invocation

### Phase 5: Documentation and Release

1. Update plugin documentation
2. Add to plugin registry
3. Create release package
4. Update FEATURES_OVERVIEW.md

## Optional Dependencies Installation

The plugin works best with optional tools installed:

```bash
# Ubuntu/Debian
sudo apt install tesseract-ocr tesseract-ocr-ces libimage-exiftool-perl

# Fedora
sudo dnf install tesseract tesseract-langpack-ces perl-Image-ExifTool

# Via Portunix (future)
portunix install tesseract
portunix install exiftool
```

## Testing Strategy

### Unit Tests

- [ ] Plugin manifest parsing
- [ ] Java plugin configuration
- [ ] JVM argument construction

### Integration Tests

- [ ] Plugin installation from local path
- [ ] Plugin start/stop lifecycle
- [ ] gRPC communication
- [ ] Text extraction from sample documents
- [ ] MCP tool invocation

### E2E Tests

- [ ] Full extraction workflow via CLI
- [ ] Batch processing
- [ ] AI assistant integration

## Success Criteria

- [ ] Plugin can be installed via `portunix plugin install`
- [ ] Plugin starts and responds to health checks
- [ ] Text extraction works for PDF, DOCX, and common formats
- [ ] MCP tools are accessible to AI assistants
- [ ] CLI commands work correctly
- [ ] Plugin appears in `portunix plugin list`

## Dependencies

### Internal Dependencies

- Issue #007: Plugin System with gRPC Architecture
- Issue #024: Plugin Registration and Discovery System

### External Dependencies

- Java 21+ runtime
- Apache Tika 3.2.3
- gRPC libraries
- Optional: Tesseract OCR, ExifTool

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Java runtime not available | High | Check Java availability, provide installation guidance |
| Large JAR file size (~150MB) | Medium | Document size, consider modular packaging |
| JVM memory consumption | Medium | Configure reasonable JVM defaults, allow customization |
| gRPC version compatibility | Medium | Pin gRPC versions, test compatibility |

## Notes

- Plugin is developed by Adam with Zdenek as lead
- Uses Apache Tika for robust document parsing
- Supports 50+ document formats out of the box
- OCR requires external Tesseract installation

---

**Created**: 2026-01-11
**Author**: Zdenek
**Target Release**: v1.11.0
