# Issue #98: PTX-Vocalio Helper Implementation

**Status**: 📋 Open
**Priority**: High
**Type**: Feature
**Created**: 2025-10-23
**Updated**: 2025-10-23
**Architecture Decision Record**: TBD

## Problem Statement

Portunix currently lacks integrated support for speech-to-text (STT) and text-to-speech (TTS) operations, which are increasingly important for:

- Voice-controlled development environments
- Accessibility features for developers with disabilities
- AI assistant voice integration (MCP server voice interface)
- Audio documentation generation from text
- Speech recognition for coding via voice commands
- Multi-language voice interface support

Currently, developers must manually install and configure various STT/TTS tools (Whisper, Coqui TTS, piper, etc.), which is complex and platform-dependent.

## Proposed Solution

Implement a dedicated `ptx-vocalio` helper binary following the established helper binary architecture pattern used by `ptx-python`, `ptx-virt`, and `ptx-ansible`.

### Key Components

#### 1. Helper Binary Architecture
- **`ptx-vocalio`** - Dedicated speech processing helper binary
- **`portunix`** - Main dispatcher routing vocalio commands
- **Integration with `ptx-python`** - Many STT/TTS tools are Python-based

#### 2. Interactive Setup Wizard
**Command**: `portunix vocalio prepare`

The helper will include an interactive configuration wizard that asks:
1. **Vocalio Name** - Name for the vocalio configuration being created (e.g., "dev-assistant", "meeting-transcriber")
2. **Target Operating System** - Which OS this vocalio will run on:
   - Linux (Ubuntu, Debian, Fedora, etc.)
   - Windows (10, 11, Server)
   - Raspberry Pi (Raspbian/Raspberry Pi OS)
   - macOS
   - Other embedded systems
3. **Language Selection** - Target language for STT/TTS (English, Czech, German, etc.)
4. **STT Tool Selection** - Choose speech-to-text engine:
   - OpenAI Whisper (local)
   - Faster-Whisper
   - Vosk
   - DeepSpeech
   - (other options to be determined)
5. **TTS Tool Selection** - Choose text-to-speech engine:
   - Coqui TTS
   - Piper TTS
   - espeak-ng
   - Festival
   - (other options to be determined)
6. **Target System and Output Type** - Combined use case and output configuration:
   - Development assistant + Real-time streaming
   - Meeting transcription + File-based (WAV, MP3, SRT)
   - Documentation generation + File-based (MP3, OGG)
   - AI assistant voice interface (MCP) + Real-time streaming + API
   - Custom configuration
   - (other combinations to be determined)

#### 3. Automatic Setup & Model Management
**Command**: `portunix vocalio install <name>`

After creating configuration with `portunix vocalio prepare`, use `install` command to:
- Read configuration from `<name>.vocalio.yaml` file
- Create Python venv via `ptx-python` (named after vocalio configuration)
- Download and install selected STT/TTS tools (optimized for target OS)
- Download pre-trained models for selected language
- Compile/optimize models for local execution and target OS
- Cache models in `.cache/vocalio/models/` (consistent with Portunix cache pattern)
- Integrate with `ptx-python` for Python-based tools
- Apply OS-specific optimizations (GPU settings, audio backends, etc.)

**Three-Step Workflow**:
1. `portunix vocalio prepare` → creates `<name>.vocalio.yaml` configuration
2. `portunix vocalio install <name>` → installs tools and downloads models based on YAML
3. `portunix vocalio create <name>` → **creates standalone executable** (exe/binary) with `--input` and `--output` parameters

**Example of final product**:
```bash
# After create, you get a standalone executable:
./meeting-assistant --input microphone --output speakers
./meeting-assistant --input audio.mp3 --output transcript.txt
./doc-generator --input document.txt --output speech.mp3
```

**Storage Locations** (consistent with Portunix architecture):
- **Configurations**: `~/.portunix/vocalio/<name>.vocalio.yaml`
- **Models cache**: `.cache/vocalio/models/` (following Portunix cache pattern)
- **Python venvs**: `~/.portunix/python/venvs/<name>` (existing ptx-python system)
- **Built executables**: `./dist/<name>` or `./dist/<name>.exe` (Windows)
- **Temporary files**: `.cache/vocalio/tmp/`

## Implementation Phases

### Phase 1: Foundation & Interactive Wizard (TODO)
**Goal**: Basic helper infrastructure with interactive setup

#### Tasks:
1. **Helper Binary Infrastructure**
   - [ ] Create `ptx-vocalio` binary project structure
   - [ ] Implement basic binary skeleton with version and help commands
   - [ ] Set up build pipeline for `ptx-vocalio` in main Makefile
   - [ ] Create integration tests for binary communication

2. **Main Binary Dispatcher**
   - [ ] Add `vocalio` command group to main `portunix` binary
   - [ ] Implement `portunix vocalio prepare` → interactive wizard (creates YAML)
   - [ ] Implement `portunix vocalio install <name>` → install tools and models from YAML
   - [ ] Implement `portunix vocalio stt` → speech-to-text operations
   - [ ] Implement `portunix vocalio tts` → text-to-speech operations
   - [ ] Add `portunix vocalio check` for helper availability

3. **Interactive Configuration Wizard** (`vocalio prepare`)
   - [ ] Implement multi-step interactive questionnaire
   - [ ] Question 1: Vocalio name (configuration identifier)
   - [ ] Question 2: Target operating system (Linux, Windows, Raspberry Pi, macOS, Other)
   - [ ] Question 3: Language selection (with auto-detection suggestion)
   - [ ] Question 4: STT tool selection (with recommendations based on OS)
   - [ ] Question 5: TTS tool selection (with recommendations based on OS)
   - [ ] Question 6: Target system and output type (combined use case + output format)
   - [ ] Store configuration in `~/.portunix/vocalio/<name>.vocalio.yaml` (wizard output format)

4. **Configuration Management**
   - [ ] Implement `config list` - list all vocalio configurations
   - [ ] Implement `config show <name>` - display configuration details
   - [ ] Implement `config edit <name>` - open YAML in editor
   - [ ] Implement `config validate <name>.vocalio.yaml` - validate YAML structure
   - [ ] Implement `config use <name>` - set default configuration
   - [ ] Support `--config <name>` parameter for all STT/TTS commands
   - [ ] YAML schema validation (structure, required fields, value types)

5. **Install Command** (`vocalio install <name>`)
   - [ ] Implement YAML configuration parser
   - [ ] Read and validate `<name>.vocalio.yaml` configuration
   - [ ] Extract target_os, language, STT/TTS engines, and dependencies
   - [ ] Create Python venv for configuration: `portunix python venv create <name>`
   - [ ] Install STT/TTS tools based on YAML specifications
   - [ ] Download language models for selected engines
   - [ ] Apply OS-specific optimizations from YAML
   - [ ] Verify installation and model availability
   - [ ] Generate installation summary report

6. **Create Command** (`vocalio create <name>`) - **Executable Generation**
   - [ ] Generate Python wrapper script from YAML configuration
   - [ ] Implement CLI argument parsing (--input, --output)
   - [ ] Input handling (microphone, audio file, text file, stdin)
   - [ ] Output handling (speakers, audio file, text file, stdout)
   - [ ] Use PyInstaller/cx_Freeze to compile to standalone executable
   - [ ] Bundle models and dependencies into executable
   - [ ] Create executable in `./dist/<name>` or `./dist/<name>.exe`
   - [ ] Test executable with sample inputs
   - [ ] Generate usage documentation for the executable

7. **PTX-Python Integration**
   - [ ] Detect existing Python installation via Portunix
   - [ ] Create dedicated venv for each vocalio configuration (named after config)
   - [ ] Coordinate with `ptx-python` for Python-based tool installation
   - [ ] Support OS-specific Python dependencies (e.g., pywin32 for Windows)

#### Success Criteria:
- [ ] `ptx-vocalio` binary builds and executes independently
- [ ] `portunix vocalio prepare` launches interactive wizard
- [ ] Interactive wizard guides user through all 6 configuration questions
- [ ] Configuration is saved as `<name>.vocalio.yaml` file with correct YAML structure
- [ ] YAML file contains all wizard choices (name, target OS, language, STT, TTS, target system + output type)
- [ ] Tool recommendations adapt based on selected target OS (e.g., lightweight models for Raspberry Pi)
- [ ] `portunix vocalio install <name>` reads YAML and installs everything automatically
- [ ] Install command creates Python venv, installs tools, downloads models
- [ ] `portunix vocalio create <name>` generates standalone executable
- [ ] Created executable accepts `--input` and `--output` parameters
- [ ] Executable works independently without Python installation
- [ ] Input sources: microphone, audio files, text files, stdin
- [ ] Output targets: speakers, audio files, text files, stdout
- [ ] Integration with `ptx-python` works correctly
- [ ] Cross-platform compatibility (Windows/Linux/Raspberry Pi/macOS)

### Phase 2: STT (Speech-to-Text) Implementation (TODO)
**Goal**: Implement speech recognition capabilities

#### Tasks:
1. **Whisper Integration** (Primary STT Tool)
   - [ ] Install OpenAI Whisper via pip
   - [ ] Download selected language model
   - [ ] Implement `stt transcribe <audio-file>` - transcribe audio file
   - [ ] Implement `stt live` - real-time transcription from microphone
   - [ ] Support multiple model sizes (tiny, base, small, medium, large)

2. **Alternative STT Tools**
   - [ ] Faster-Whisper integration (optimized Whisper)
   - [ ] Vosk integration (offline, lightweight)
   - [ ] Support for custom models

3. **Audio Input Management**
   - [ ] Microphone selection and configuration
   - [ ] Audio format conversion (MP3, WAV, OGG, FLAC)
   - [ ] Audio quality enhancement (noise reduction)

4. **Output Formats**
   - [ ] Plain text output
   - [ ] JSON format with timestamps
   - [ ] SRT subtitle format
   - [ ] VTT subtitle format

#### Success Criteria:
- [ ] Audio files are accurately transcribed to text
- [ ] Real-time microphone transcription works
- [ ] Multiple languages are supported
- [ ] Output formats are generated correctly

### Phase 3: TTS (Text-to-Speech) Implementation (TODO)
**Goal**: Implement text-to-speech synthesis

#### Tasks:
1. **Coqui TTS Integration** (Primary TTS Tool)
   - [ ] Install Coqui TTS via pip
   - [ ] Download selected voice models
   - [ ] Implement `tts speak <text>` - convert text to speech
   - [ ] Implement `tts speak-file <text-file>` - convert file to speech
   - [ ] Support voice selection and customization

2. **Alternative TTS Tools**
   - [ ] Piper TTS integration (fast, local)
   - [ ] espeak-ng integration (lightweight, multi-language)
   - [ ] Support for custom voice models

3. **Voice Configuration**
   - [ ] Voice selection (male/female, accents)
   - [ ] Speech rate adjustment
   - [ ] Pitch and volume control
   - [ ] Emotion/tone configuration

4. **Output Management**
   - [ ] WAV file output
   - [ ] MP3 file output
   - [ ] Real-time audio playback
   - [ ] Streaming audio for long texts

#### Success Criteria:
- [ ] Text is converted to natural-sounding speech
- [ ] Multiple voices and languages are supported
- [ ] Audio output quality is acceptable
- [ ] Different output formats work correctly

### Phase 4: Advanced Features & Integration (TODO)
**Goal**: Advanced workflows and ecosystem integration

#### Tasks:
1. **Model Management**
   - [ ] Implement `vocalio models list` - show installed models
   - [ ] Implement `vocalio models download <language>` - download additional models
   - [ ] Implement `vocalio models remove <model>` - remove unused models
   - [ ] Automatic model updates

2. **MCP Server Integration**
   - [ ] Voice interface for MCP server (Claude voice commands)
   - [ ] STT integration for voice-based AI queries
   - [ ] TTS integration for AI response audio
   - [ ] Real-time voice conversation with AI assistant

3. **Batch Processing**
   - [ ] Batch transcription of multiple audio files
   - [ ] Batch TTS conversion of multiple text files
   - [ ] Directory watching for automatic processing

4. **API Server Mode**
   - [ ] HTTP API for STT operations
   - [ ] HTTP API for TTS operations
   - [ ] WebSocket support for real-time streaming
   - [ ] Authentication and rate limiting

#### Success Criteria:
- [ ] Models are managed effectively
- [ ] MCP server voice interface works
- [ ] Batch processing handles large workloads
- [ ] API server mode is production-ready

## Command Structure

```bash
# Three-Step Workflow: Prepare → Install → Create
portunix vocalio prepare            # Step 1: Run interactive wizard (creates <name>.vocalio.yaml)
portunix vocalio install <name>     # Step 2: Install tools and models from <name>.vocalio.yaml
portunix vocalio create <name>      # Step 3: Create standalone executable (exe/binary)

# Install options
portunix vocalio install <name> --reinstall  # Reinstall everything
portunix vocalio install <name> --models-only  # Only download models

# Using the created executable
./dist/meeting-assistant --input microphone --output speakers
./dist/meeting-assistant --input audio.mp3 --output transcript.txt
./dist/doc-generator --input document.txt --output speech.mp3

# Configuration Management
portunix vocalio config show <name> # Show configuration from <name>.vocalio.yaml
portunix vocalio config list        # List all available vocalio configurations
portunix vocalio config edit <name> # Edit <name>.vocalio.yaml configuration manually
portunix vocalio config validate <name>.vocalio.yaml  # Validate YAML configuration file
portunix vocalio config use <name>  # Set default vocalio configuration

# Speech-to-Text (STT)
portunix vocalio stt transcribe audio.mp3                            # Use default config
portunix vocalio stt transcribe audio.mp3 --config meeting-assistant # Use specific config
portunix vocalio stt transcribe audio.mp3 --language cs --format json
portunix vocalio stt live                                            # Real-time transcription
portunix vocalio stt live --config dev-assistant --language en
portunix vocalio stt batch *.mp3 --config meeting-assistant          # Batch transcription

# Text-to-Speech (TTS)
portunix vocalio tts speak "Hello, this is a test"                   # Use default config
portunix vocalio tts speak "Dobrý den" --config dev-assistant        # Use specific config
portunix vocalio tts speak "Dobrý den" --language cs --voice female
portunix vocalio tts speak-file document.txt --output speech.mp3
portunix vocalio tts speak-file document.txt --config meeting-assistant --voice male

# Model Management
portunix vocalio models list
portunix vocalio models download whisper-medium-cs
portunix vocalio models download coqui-tts-cs-female
portunix vocalio models remove whisper-large
portunix vocalio models update      # Update all models

# Server Mode
portunix vocalio serve              # Start API server
portunix vocalio serve --port 8080 --host 0.0.0.0

# Check helper availability
portunix vocalio check
portunix vocalio version
```

## Example Workflows

### Workflow 1: Initial Setup
```bash
# Step 1: Run interactive configuration wizard
portunix vocalio prepare

# Wizard asks:
# 1. Enter vocalio name:
#    > meeting-assistant
#    (Name for this vocalio configuration)
#
# 2. Select target operating system:
#    - Linux (Ubuntu, Debian, Fedora, etc.)
#    - Windows (10, 11, Server)
#    - Raspberry Pi (Raspbian/Raspberry Pi OS)
#    - macOS
#    - Other embedded systems
#    > Linux
#
# 3. Select language:
#    - English
#    - Czech
#    - German
#    - Spanish
#    - (other languages)
#    > English
#
# 4. Select STT tool (recommendations based on Linux):
#    - OpenAI Whisper (recommended, accurate, offline)
#    - Faster-Whisper (optimized, faster, good for servers)
#    - Vosk (lightweight, fast, good for embedded)
#    - (other options)
#    > Whisper Medium
#
# 5. Select TTS tool (recommendations based on Linux):
#    - Coqui TTS (recommended, natural voices)
#    - Piper TTS (fast, good quality)
#    - espeak-ng (lightweight, many languages)
#    - (other options)
#    > Coqui TTS
#
# 6. Select target system and output type:
#    - Development assistant + Real-time streaming
#    - Meeting transcription + File-based (WAV, MP3, SRT)
#    - Documentation generation + File-based (MP3, OGG)
#    - AI assistant voice interface (MCP) + Real-time streaming + API
#    - Custom configuration
#    > Meeting transcription + File-based (WAV, MP3, SRT)

# Wizard output:
# ✅ Configuration saved: ~/.portunix/vocalio/meeting-assistant.vocalio.yaml
#
# Next steps:
#   1. Review configuration: portunix vocalio config show meeting-assistant
#   2. Install vocalio environment: portunix vocalio install meeting-assistant
#   3. Create standalone executable: portunix vocalio create meeting-assistant

# Step 2: Install the vocalio environment
portunix vocalio install meeting-assistant

# Installation process:
# 📦 Creating Python venv: meeting-assistant
# 🔧 Installing STT tool: Whisper Medium (Linux optimized)
# 🔧 Installing TTS tool: Coqui TTS
# 📥 Downloading language models: English
# ⚙️  Applying OS optimizations: Linux/ALSA audio backend
# ✅ Installation completed successfully!

# Step 3: Create standalone executable
portunix vocalio create meeting-assistant

# Create process:
# 📝 Generating Python wrapper script...
# 🔨 Compiling with PyInstaller...
# 📦 Bundling models and dependencies...
# 💾 Creating executable: ./dist/meeting-assistant
# ✅ Executable created successfully!
#
# Usage:
#   ./dist/meeting-assistant --input microphone --output speakers
#   ./dist/meeting-assistant --input meeting.mp3 --output transcript.txt
```

### Workflow 2: Transcribe Audio File
```bash
# Transcribe single audio file
portunix vocalio stt transcribe meeting-recording.mp3

# Output: Text transcription to console

# Transcribe with specific language and format
portunix vocalio stt transcribe meeting.mp3 --language cs --format srt --output subtitles.srt

# Batch transcribe multiple files
portunix vocalio stt batch *.mp3 --format json
```

### Workflow 3: Generate Speech from Text
```bash
# Simple text-to-speech
portunix vocalio tts speak "Welcome to Portunix"

# Advanced TTS with options
portunix vocalio tts speak "Vítejte v Portunixu" --language cs --voice female --rate 1.1 --output welcome.mp3

# Convert text file to audio
portunix vocalio tts speak-file documentation.md --output audio-docs.mp3
```

### Workflow 4: Real-Time Voice Transcription
```bash
# Start live transcription from microphone
portunix vocalio stt live --language en

# Output:
# 🎤 Listening... (Press Ctrl+C to stop)
# [00:01] Hello this is a test
# [00:05] Portunix is working correctly
# [00:10] Speech recognition is functioning
```

### Workflow 5: MCP Voice Interface
```bash
# Enable voice interface for MCP server
portunix vocalio serve --mcp-integration

# Now Claude can:
# - Listen to voice commands via STT
# - Respond with voice via TTS
# - Have real-time voice conversations
```

### Workflow 6: Multiple Vocalio Configurations
```bash
# Create multiple vocalio configurations for different purposes

# Configuration 1: Meeting transcription on Linux server
portunix vocalio prepare
# 1. Enter name: meeting-assistant
# 2. Target OS: Linux
# 3. Language: English
# 4. STT: Whisper Medium (recommended for Linux servers)
# 5. TTS: Coqui TTS
# 6. Target system: Meeting transcription + File-based (WAV, MP3, SRT)
# ✅ Saved: meeting-assistant.vocalio.yaml

portunix vocalio install meeting-assistant
# 📦 Installing meeting-assistant...
# ✅ Ready!

portunix vocalio create meeting-assistant
# 🔨 Creating executable...
# ✅ Created: ./dist/meeting-assistant

# Configuration 2: Development assistant on Windows workstation
portunix vocalio prepare
# 1. Enter name: dev-assistant
# 2. Target OS: Windows
# 3. Language: Czech
# 4. STT: Faster-Whisper Small (optimized for Windows)
# 5. TTS: Piper TTS (fast on Windows)
# 6. Target system: Development assistant + Real-time streaming
# ✅ Saved: dev-assistant.vocalio.yaml

portunix vocalio install dev-assistant
# 📦 Installing dev-assistant...
# ✅ Ready!

portunix vocalio create dev-assistant
# 🔨 Creating executable...
# ✅ Created: ./dist/dev-assistant.exe (Windows)

# Configuration 3: Documentation generator on Raspberry Pi
portunix vocalio prepare
# 1. Enter name: doc-generator-rpi
# 2. Target OS: Raspberry Pi
# 3. Language: English
# 4. STT: Vosk (lightweight, optimized for embedded)
# 5. TTS: espeak-ng (minimal resources, fast on Pi)
# 6. Target system: Documentation generation + File-based (MP3, OGG)
# ✅ Saved: doc-generator-rpi.vocalio.yaml

portunix vocalio install doc-generator-rpi
# 📦 Installing doc-generator-rpi...
# ✅ Ready!

portunix vocalio create doc-generator-rpi
# 🔨 Creating executable...
# ✅ Created: ./dist/doc-generator-rpi (ARM binary for Raspberry Pi)

# Now you have 3 standalone executables:
ls -lh ./dist/
# -rwxr-xr-x 1 user user 45M meeting-assistant       (Linux x86_64)
# -rwxr-xr-x 1 user user 52M dev-assistant.exe       (Windows x64)
# -rwxr-xr-x 1 user user 28M doc-generator-rpi       (ARM for Raspberry Pi)

# List all configurations
portunix vocalio config list
# Output:
# Available vocalio configurations:
#   - meeting-assistant (Linux, en, meeting-transcription, Whisper+Coqui)
#   - dev-assistant (Windows, cs, development-assistant, Faster-Whisper+Piper)
#   - doc-generator-rpi (Raspberry Pi, en, documentation, Vosk+espeak-ng)

# Use specific configuration
portunix vocalio stt transcribe meeting.mp3 --config meeting-assistant
portunix vocalio tts speak "Kód vypadá dobře" --config dev-assistant
portunix vocalio tts speak-file README.md --output readme-audio.mp3 --config doc-generator-rpi

# Set default configuration
portunix vocalio config use dev-assistant

# Now all commands use dev-assistant by default
portunix vocalio stt live  # Uses dev-assistant (Windows, Czech, Faster-Whisper)
```

## Configuration File Format

### Example: `meeting-assistant.vocalio.yaml`
```yaml
# Vocalio Configuration File
# Generated by: portunix vocalio setup
# Created: 2025-10-23

name: meeting-assistant
version: 1.0

# Target operating system
target_os: linux
target_os_details:
  distribution: ubuntu
  version: "22.04"
  architecture: x86_64

# Target system/use case and output type (combined from Question 6)
target_system: meeting-transcription
output_type: file-based
use_case_description: "Meeting transcription with file-based output (WAV, MP3, SRT subtitles)"

# Language configuration
language:
  primary: en
  fallback: cs

# Speech-to-Text (STT) configuration
stt:
  engine: whisper
  model: medium
  model_path: .cache/vocalio/models/whisper-medium-en  # Consistent with Portunix cache pattern
  options:
    sample_rate: 16000
    channels: 1
    language: en
    # Optimizations for Linux
    device: cpu
    compute_type: float32

# Text-to-Speech (TTS) configuration
tts:
  engine: coqui-tts
  voice: female
  model_path: .cache/vocalio/models/coqui-tts-en-female  # Consistent with Portunix cache pattern
  options:
    speed: 1.0
    pitch: 1.0
    volume: 0.8
    # Optimizations for Linux
    vocoder: hifigan

# Output configuration (from Question 6)
output:
  type: file-based
  formats:
    - wav
    - mp3
    - srt  # Subtitle format for meeting transcription
  directory: ~/vocalio-output/meeting-assistant
  real_time: false
  api_enabled: false

# Python environment
python:
  venv_name: meeting-assistant
  venv_path: ~/.portunix/python/venvs/meeting-assistant
  python_version: "3.11"
  dependencies:
    - openai-whisper
    - TTS
    - pydub
    - sounddevice
    # Linux-specific dependencies
    - portaudio19-dev

# Advanced settings
advanced:
  gpu_acceleration: false
  cache_models: true
  auto_update_models: false
  logging_level: info
  # OS-specific optimizations
  os_optimizations:
    linux_audio_backend: alsa
```

### Example: `dev-assistant.vocalio.yaml`
```yaml
# Vocalio Configuration File - Development Assistant
name: dev-assistant
version: 1.0

# Target operating system
target_os: windows
target_os_details:
  version: "11"
  edition: pro
  architecture: x86_64

# Target system/use case and output type (combined from Question 6)
target_system: development-assistant
output_type: real-time-streaming
use_case_description: "Development assistant with real-time voice streaming for IDE integration"

language:
  primary: cs
  fallback: en

stt:
  engine: faster-whisper
  model: small
  model_path: .cache/vocalio/models/faster-whisper-small-cs  # Consistent with Portunix cache pattern
  options:
    sample_rate: 16000
    language: cs
    # Optimizations for Windows
    device: cuda  # GPU acceleration on Windows
    compute_type: float16

tts:
  engine: piper-tts
  voice: male
  model_path: .cache/vocalio/models/piper-tts-cs-male  # Consistent with Portunix cache pattern
  options:
    speed: 1.1
    # Windows-optimized TTS
    quality: high

output:
  type: real-time-streaming
  real_time: true
  api_enabled: true
  api_port: 8080
  api_host: localhost
  streaming_buffer_size: 512

python:
  venv_name: dev-assistant
  venv_path: ~/.portunix/python/venvs/dev-assistant
  python_version: "3.11"
  dependencies:
    - faster-whisper
    - piper-tts
    - pydub
    - sounddevice
    # Windows-specific dependencies
    - pywin32

advanced:
  gpu_acceleration: true
  cache_models: true
  mcp_integration: true
  logging_level: info
  # OS-specific optimizations
  os_optimizations:
    windows_audio_backend: wasapi
```

## Benefits

### Positive Consequences
- **Accessibility**: Voice interface for developers with disabilities
- **Productivity**: Hands-free coding and documentation
- **Multi-Language Support**: Easy language switching for international teams
- **AI Integration**: Voice interface for Claude and other AI assistants
- **Offline Capable**: Local STT/TTS without cloud dependencies
- **Easy Setup**: Interactive wizard simplifies complex tool installation

### Challenges to Address
- **Model Size**: STT/TTS models can be 500MB - 2GB in size
- **Performance**: Real-time processing requires adequate CPU/GPU
- **Audio Quality**: Microphone quality affects transcription accuracy
- **Language Support**: Not all tools support all languages equally
- **Python Dependency**: Most tools require Python and pip

## Technical Considerations

### Tool Selection Criteria

#### STT Tools Comparison
| Tool | Size | Speed | Accuracy | Languages | Offline |
|------|------|-------|----------|-----------|---------|
| Whisper | Large | Medium | Excellent | 99+ | Yes |
| Faster-Whisper | Large | Fast | Excellent | 99+ | Yes |
| Vosk | Small | Very Fast | Good | 20+ | Yes |
| DeepSpeech | Medium | Fast | Good | Few | Yes |

#### TTS Tools Comparison
| Tool | Size | Speed | Quality | Languages | Offline |
|------|------|-------|---------|-----------|---------|
| Coqui TTS | Large | Medium | Excellent | 50+ | Yes |
| Piper TTS | Medium | Fast | Very Good | 40+ | Yes |
| espeak-ng | Tiny | Very Fast | Good | 100+ | Yes |
| Festival | Small | Fast | Good | Few | Yes |

### Storage Requirements
- **Models**: 500MB - 5GB depending on selections
- **Cache**: 100MB - 500MB for temporary audio files
- **Configuration**: < 1MB for config files
- **Default location**: `~/.portunix/vocalio/`

### Python Integration
```
~/.portunix/python/venvs/ptx-vocalio/
├── bin/                    # Python binaries
├── lib/                    # Installed packages
│   ├── whisper/            # STT tool
│   ├── TTS/                # Coqui TTS
│   └── piper/              # Alternative TTS
└── share/                  # Shared data
```

### Model Storage
**Location**: `.cache/vocalio/models/` (consistent with Portunix cache architecture)

```
.cache/vocalio/models/
├── stt/
│   ├── whisper-medium-en/  # English STT model
│   ├── whisper-medium-cs/  # Czech STT model
│   └── vosk-model-en/      # Alternative STT model
└── tts/
    ├── coqui-tts-en-female/  # English TTS voice
    ├── coqui-tts-cs-male/    # Czech TTS voice
    └── piper-tts-en/         # Alternative TTS voice
```

**Why `.cache/` instead of `~/.portunix/`?**
- Follows existing Portunix patterns (`.cache/python-embeddable`, `.cache/vscode`, `.cache/isos`)
- Models are large temporary files, treated as cache
- Easy cleanup without affecting configurations
- Configurations remain in `~/.portunix/vocalio/` for persistence

## Success Criteria

### Phase 1 (Foundation)
- [ ] `ptx-vocalio` helper binary successfully builds
- [ ] **STEP 1: Prepare** - `portunix vocalio prepare` launches interactive wizard
- [ ] Interactive wizard guides user through all 6 configuration questions:
  1. Vocalio name
  2. Target OS (Linux/Windows/Raspberry Pi/macOS/Other)
  3. Language selection
  4. STT tool (with OS-specific recommendations)
  5. TTS tool (with OS-specific recommendations)
  6. Target system and output type (combined use case + output format)
- [ ] Configuration is saved as `<name>.vocalio.yaml` file with correct YAML structure
- [ ] YAML includes target_os, target_system, output_type, and all tool configurations
- [ ] Tool recommendations adapt based on selected target OS (lightweight for Raspberry Pi, GPU for Windows, etc.)
- [ ] **STEP 2: Install** - `portunix vocalio install <name>` reads YAML configuration
- [ ] Install command creates Python venv via `ptx-python`
- [ ] Install command installs STT/TTS tools based on YAML specifications
- [ ] Install command downloads language models for selected engines
- [ ] Install command applies OS-specific optimizations
- [ ] **STEP 3: Create** - `portunix vocalio create <name>` generates standalone executable
- [ ] Create command generates Python wrapper script from YAML
- [ ] Create command uses PyInstaller to compile to exe/binary
- [ ] Created executable placed in `./dist/<name>` or `./dist/<name>.exe`
- [ ] Executable accepts `--input` and `--output` CLI parameters
- [ ] Executable works without Python installation on target system
- [ ] YAML configuration can be loaded and used by vocalio commands
- [ ] Integration with `ptx-python` works correctly
- [ ] Models are downloaded and cached properly with OS-specific optimizations

### Phase 2 (STT)
- [ ] Audio files are accurately transcribed
- [ ] Real-time microphone transcription works
- [ ] Multiple languages are supported
- [ ] Output formats (text, JSON, SRT) work correctly

### Phase 3 (TTS)
- [ ] Text is converted to natural-sounding speech
- [ ] Multiple voices and languages are supported
- [ ] Audio output quality is good
- [ ] File and real-time playback work

### Phase 4 (Advanced)
- [ ] Model management commands work correctly
- [ ] MCP voice interface is functional
- [ ] Batch processing handles multiple files
- [ ] API server mode is production-ready

## Dependencies

### Core Dependencies
- Python installation via Portunix package system
- `ptx-python` helper for Python environment management ([#097](done/097-ptx-python-helper-implementation.md))
- Helper binary architecture from [ADR-014](../adr/014-git-dispatcher-with-python-distribution.md)
- Portunix logging system from [#052](done/052-logging-system-implementation.md)

### Python Tool Dependencies (Installed by ptx-vocalio)
- **openai-whisper** - Speech-to-text (primary)
- **faster-whisper** - Optimized Whisper alternative
- **vosk** - Lightweight STT
- **TTS** (Coqui) - Text-to-speech (primary)
- **piper-tts** - Fast TTS alternative
- **pydub** - Audio file manipulation
- **sounddevice** - Microphone access
- **numpy** - Audio processing
- **torch** - Neural network inference (for Whisper/Coqui)

### System Dependencies
- **ffmpeg** - Audio format conversion
- **portaudio** - Microphone access (Linux)
- **CUDA** (optional) - GPU acceleration for faster processing

## Related Issues

### Core Infrastructure
- [#051](051-git-dispatcher-python-distribution-architecture.md) - Git-like Dispatcher Architecture
- [#052](done/052-logging-system-implementation.md) - Logging System Implementation
- [#097](done/097-ptx-python-helper-implementation.md) - PTX-Python Helper (critical dependency)

### AI Integration
- [#004](done/004-mcp-server-ai-integration.md) - MCP Server AI Integration
- [#073](073-ptx-prompting-helper-implementation.md) - PTX-Prompting Helper

### Helper Binary Examples
- [#056](done/056-ansible-infrastructure-as-code-integration.md) - Ansible Helper (reference architecture)
- [#068](068-main-binary-ptx-virt-helper-integration.md) - Virt Helper Integration (reference)

## Labels

`enhancement`, `helper-binary`, `speech-recognition`, `text-to-speech`, `ai-integration`, `accessibility`, `voice-interface`, `python-integration`

## Notes

### Why PTX-Vocalio?
- **Unified Interface**: Consistent with Portunix command structure
- **Simplified Setup**: Interactive wizard handles complex tool installation
- **Cross-Platform**: Abstracts platform differences for audio processing
- **Integration**: Works with `ptx-python` and MCP server
- **Offline Capable**: All processing can happen locally
- **Multi-Language**: Easy switching between languages
- **Multiple Configurations**: Support for multiple vocalio profiles (meeting, dev, docs, etc.)

### Configuration Design
- **YAML Format**: Human-readable `.vocalio.yaml` configuration files
- **Named Configurations**: Each vocalio has a unique name (e.g., `meeting-assistant`, `dev-assistant`)
- **Multiple Profiles**: Users can create multiple configurations for different purposes
- **Default Selection**: Set one configuration as default with `config use <name>`
- **Per-Command Override**: Use `--config <name>` to specify configuration for individual commands
- **Validation**: Built-in YAML schema validation for configuration integrity
- **Portability**: Configuration files can be shared across teams/machines

### Use Cases
1. **Accessibility**: Voice-controlled development for disabled developers
2. **Documentation**: Convert markdown docs to audio
3. **Meetings**: Transcribe development meetings automatically
4. **AI Voice Chat**: Voice interface for Claude via MCP
5. **Language Learning**: Practice pronunciation with TTS
6. **Code Review**: Listen to code comments and documentation

### Target Audience
- **Accessibility Users**: Developers with visual or motor disabilities
- **Multi-Language Teams**: International teams needing translation/transcription
- **AI Power Users**: Developers wanting voice-based AI interaction
- **Content Creators**: Developers creating audio documentation

## Open Questions (To Be Resolved)

1. **Model Storage**: Should models be shared globally or per-user? Per-configuration or shared?
2. **Configuration Storage**: Where to store `.vocalio.yaml` files - `~/.portunix/vocalio/` or allow custom paths?
3. **GPU Support**: How to handle CUDA/GPU acceleration configuration in YAML?
4. **Cloud Options**: Should we support cloud-based STT/TTS as alternatives?
5. **Real-Time Performance**: What are minimum hardware requirements for live transcription?
6. **Voice Training**: Should we support custom voice training for TTS?
7. **Licensing**: Are all selected tools compatible with commercial use?
8. **YAML Schema**: Should we use JSON Schema for YAML validation or custom Go validation?
9. **Configuration Sharing**: How to handle sharing configurations across team members? Git-friendly format?

---

**NOTE FOR USER**: Updated structure with:
✅ 6 wizard questions:
  1. Vocalio name
  2. Target OS (Linux/Windows/Raspberry Pi/macOS/Other)
  3. Language selection
  4. STT tool (OS-specific recommendations)
  5. TTS tool (OS-specific recommendations)
  6. Target system and output type (combined)
✅ Three-step workflow:
  - `portunix vocalio prepare` → creates YAML configuration
  - `portunix vocalio install <name>` → installs tools and models
  - `portunix vocalio create <name>` → generates standalone executable
✅ YAML configuration format (`.vocalio.yaml`) with target_os field
✅ Multiple configuration support
✅ Configuration management commands
✅ Example YAML files (Linux + Windows examples)
✅ Workflow for multiple configurations with create/build separation
✅ OS-specific optimizations in YAML and build process

Please review and add:
- Specific tool recommendations and reasons for each target system
- Detailed output type options based on use cases
- Hardware requirements for different model sizes
- Any additional use cases specific to Czech/international development
- Specific model recommendations for Czech language
- Answer Open Questions about storage, GPU support, and YAML schema
