# Issue #149: Vox Plugin - ptx-install Integration for Model Installation

**Status**: ✅ Implemented
**Closed**: 2026-02-18
**Priority**: Medium
**Type**: Feature
**Labels**: ptx-installer, integration, models, multi-file-download, plugin-support
**Assignee**: -
**Depends on**: portunix-plugins #039

## Summary

Enable Vox plugin models and system dependencies to be installed via `portunix install` (ptx-installer). The Vox plugin (TTS/STT) requires large binary model files (~60-460 MB) and system audio libraries that should be manageable through the standard Portunix installation mechanism.

**Origin**: portunix-plugins internal issue #040 (handoff to core team)
**Depends on**: portunix-plugins #039 (Vox Plugin implementation)

## Current State

- Vox plugin handles model downloads internally via `huggingface_hub` Python package
- System audio libraries (libasound2-dev on Linux) must be installed manually
- No visibility of models in `portunix package list`
- No offline pre-provisioning possible

## Desired State

```bash
# Install system prerequisites
portunix install vox-deps

# Install TTS voice models
portunix install vox-model-tts-czech       # ~61 MB
portunix install vox-model-tts-english      # ~60 MB

# Install STT models (lower priority)
portunix install vox-model-stt-base         # ~140 MB
portunix install vox-model-stt-small        # ~460 MB

# Bundle install per language
portunix install vox-czech                  # TTS czech voice
portunix install vox-english                # TTS english voice

# List/manage
portunix package list --category plugins/models
```

## Requirements

### FR-1: System prerequisites package (`vox-deps`)

| Platform | Packages |
| -------- | -------- |
| Linux (apt) | `libasound2-dev`, `libsndfile1`, `ffmpeg` |
| Linux (dnf) | `alsa-lib-devel`, `libsndfile`, `ffmpeg` |
| Linux (pacman) | `alsa-lib`, `libsndfile`, `ffmpeg` |
| Windows | No additional packages needed |

### FR-2: TTS model packages

Each TTS voice requires **two files** downloaded from Hugging Face (`rhasspy/piper-voices` repo):

- `{voice_id}.onnx` - ONNX neural network model
- `{voice_id}.onnx.json` - Model configuration

| Package ID | Voice | HF Files | Target | Size |
| ---------- | ----- | -------- | ------ | ---- |
| `vox-model-tts-czech` | cs_CZ-jirka-medium | `cs/cs_CZ/jirka/medium/cs_CZ-jirka-medium.onnx` + `.onnx.json` | `models/vox/tts/cs_CZ-jirka-medium/` | ~61 MB |
| `vox-model-tts-english` | en_US-lessac-medium | `en/en_US/lessac/medium/en_US-lessac-medium.onnx` + `.onnx.json` | `models/vox/tts/en_US-lessac-medium/` | ~60 MB |

### FR-3: STT model packages (optional, lower priority)

STT models are auto-downloaded by `faster-whisper` library on first use. ptx-installer support would enable pre-provisioning but is not critical for MVP.

| Package ID | Model | Size | Note |
| ---------- | ----- | ---- | ---- |
| `vox-model-stt-base` | Whisper Base | ~140 MB | Multiple files (model.bin, config.json, tokenizer.json, etc.) |
| `vox-model-stt-small` | Whisper Small | ~460 MB | Better Czech accuracy |

## Architecture Decision Points

### DP-1: Where to define packages?

| Approach | Description | Effort |
| -------- | ----------- | ------ |
| **A: ptx-installer packages** (recommended) | Add JSONs to `src/helpers/ptx-installer/assets/packages/` using existing registry auto-discovery | Low |
| **B: Plugin-level JSON** | Extend `portunix-plugins/assets/install-packages.json` | Medium - requires cross-repo coordination |

**Recommendation**: Approach A — ptx-installer already has auto-discovery from `assets/packages/` directory (#086). No index updates needed. Consistent with existing packages (elasticsearch, minio, etc.).

### DP-2: Multi-file downloads

TTS models need 2 files per voice. Current package JSON schema has single `url` field.

**Options:**

| Option | Description | Impact |
| ------ | ----------- | ------ |
| **A: `additional_files` field** | Add new field to download schema for supplementary files | Schema extension needed |
| **B: Archive approach** | Bundle both files into a tar.gz/zip on a hosting server | Requires hosting, but uses existing extraction logic |
| **C: Two separate packages** | `vox-model-tts-czech-onnx` + `vox-model-tts-czech-config` with dependency | No schema change, but poor UX |

**Proposed solution** — Option A with `additional_files`:

```json
{
  "url": "https://huggingface.co/rhasspy/piper-voices/resolve/main/cs/.../cs_CZ-jirka-medium.onnx",
  "additional_files": [
    "https://huggingface.co/rhasspy/piper-voices/resolve/main/cs/.../cs_CZ-jirka-medium.onnx.json"
  ],
  "target": "models/vox/tts/cs_CZ-jirka-medium/",
  "extract": false
}
```

### DP-3: New categories

Add new categories for plugin-related packages:

```json
{
  "plugins/models": {
    "name": "Plugin Models",
    "description": "Machine learning models and data files for Portunix plugins"
  },
  "plugins/dependencies": {
    "name": "Plugin Dependencies",
    "description": "System-level dependencies required by Portunix plugins"
  }
}
```

## Example Package Definition (vox-deps)

```json
{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "vox-deps",
    "displayName": "Vox System Dependencies",
    "description": "System-level dependencies for Vox TTS/STT plugin (audio libraries)",
    "category": "plugins/dependencies",
    "license": "Various",
    "maintainer": "CassandraGargoyle"
  },
  "spec": {
    "platforms": {
      "linux": {
        "type": "apt",
        "variants": {
          "apt": {
            "version": "latest",
            "packages": ["libasound2-dev", "libsndfile1", "ffmpeg"],
            "distributions": ["ubuntu", "debian", "mint"]
          },
          "dnf": {
            "type": "dnf",
            "version": "latest",
            "packages": ["alsa-lib-devel", "libsndfile", "ffmpeg"]
          },
          "pacman": {
            "type": "pacman",
            "version": "latest",
            "packages": ["alsa-lib", "libsndfile", "ffmpeg"]
          }
        }
      },
      "windows": {
        "type": "script",
        "variants": {
          "default": {
            "description": "No system dependencies required on Windows",
            "postInstall": ["echo Vox: No additional system packages needed on Windows"]
          }
        }
      }
    }
  }
}
```

## Post-Implementation (plugins side)

Once ptx-installer supports Vox packages, the Vox plugin (`portunix-plugins/plugins/vox`) will be updated to:

1. Check `~/.portunix/models/vox/` as primary model location
2. Fall back to internal `huggingface_hub` download if not found
3. Update `vox models install` to suggest `portunix install` as alternative

## Acceptance Criteria

- [x] `portunix install vox-deps` installs audio libraries on apt/dnf/pacman systems
- [x] `portunix install vox-model-tts-czech` downloads both .onnx and .onnx.json files
- [x] `portunix install vox-model-tts-english` downloads both .onnx and .onnx.json files
- [x] Models are placed in correct target directory (`models/vox/tts/{voice_id}/`)
- [x] `portunix package list` shows vox packages under appropriate category
- [x] Windows: `vox-deps` installs cleanly (no-op with message)
- [x] Multi-file download (`additional_files`) works reliably for large files (60+ MB)

## Related

- **portunix-plugins #039**: Vox Plugin - Text-to-Speech & Speech-to-Text
- **portunix-plugins #040**: Original handoff issue (this is the core-side implementation)
- **#086**: Package Registry Automatic Discovery System
- **#087**: Assets Embedding Architecture
- **#100**: PTX-Installer Helper Implementation