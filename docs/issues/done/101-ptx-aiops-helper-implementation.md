# Issue #101: PTX-AIOps Helper Implementation

**Status**: 🔄 In Progress (Phase 2.5 Complete)
**Priority**: High
**Type**: Feature / Architecture
**Created**: 2025-11-29
**Updated**: 2025-11-30
**Branch**: feature/issue-101-ptx-aiops-helper
**Related Issues**:

- [#100](100-ptx-installer-helper-implementation.md) - PTX-Installer Helper (template)
- [#035](../035-ai-assistant-installation-support.md) - AI Assistant Installation Support

## Summary

Implement `ptx-aiops` helper binary for AI/ML operations tooling with focus on containerized AI infrastructure. Primary features include Ollama installation in containers, Open WebUI deployment, and NVIDIA GPU toolkit support for container workloads.

## Problem Statement

### Current Limitations

1. **No Local LLM Infrastructure**: Users cannot easily deploy local AI/LLM infrastructure
2. **Complex GPU Setup**: NVIDIA Container Toolkit requires manual configuration
3. **Fragmented Tools**: Ollama, Open WebUI, and GPU support configured separately
4. **Container Integration Gap**: No unified way to deploy AI tools in containers

### User Needs

- One-command local LLM deployment (Ollama + Open WebUI)
- Automatic GPU detection and toolkit installation
- Container-native AI infrastructure
- Consistent cross-platform experience

## Proposed Solution

### PTX-AIOps Helper Architecture

```text
User Command: portunix aiops ollama install
                    ↓
         ┌──────────────────┐
         │ Main Dispatcher  │  (lightweight, fast startup)
         │   (portunix)     │
         └────────┬─────────┘
                  │ delegates to
                  ↓
         ┌──────────────────┐
         │   ptx-aiops      │  (AI operations subsystem)
         │   Helper Binary  │
         └────────┬─────────┘
                  │
    ┌─────────────┼─────────────┐
    ↓             ↓             ↓
[Ollama]    [Open WebUI]   [GPU Toolkit]
[Container] [Container]    [Detection/Install]
```

### Core Components

#### 1. GPU Detection & Monitoring (ptx-aiops)

- Detect NVIDIA GPU presence (`nvidia-smi`)
- Check NVIDIA driver version and CUDA availability
- **Real-time GPU monitoring** with auto-refresh (`--watch` mode)
- GPU utilization, memory usage, temperature, power consumption
- Process list showing GPU-consuming applications
- Verify GPU + toolkit readiness for containers (`gpu check`)

#### 1a. NVIDIA Container Toolkit (ptx-installer)

- Install via `portunix install nvidia-container-toolkit`
- APT/DNF repository configuration
- Post-install runtime configuration (`nvidia-ctk runtime configure`)
- Validate GPU availability in containers

#### 2. Ollama Container Management

- Deploy Ollama in container with GPU support
- Model management (pull, list, remove)
- API endpoint configuration
- Health monitoring

#### 3. Open WebUI Container Deployment

- Deploy Open WebUI container
- Configure connection to Ollama
- Persistent storage management
- Authentication setup (optional)

#### 4. Integration with PTX-Container

- Use shared container code from `ptx-container`
- Consistent container runtime detection (Docker/Podman)
- Unified volume mounting
- Network configuration

## Command Structure

```bash
# GPU Operations (ptx-aiops)
portunix aiops gpu status              # Show GPU status and driver info
portunix aiops gpu status --watch      # Real-time monitoring with auto-refresh (default: 1s)
portunix aiops gpu status --watch --interval 2  # Custom refresh interval (seconds)
portunix aiops gpu usage               # Show current GPU utilization summary
portunix aiops gpu processes           # List processes using GPU
portunix aiops gpu check               # Verify GPU and toolkit readiness for containers

# GPU Toolkit Installation (via ptx-installer)
portunix install nvidia-container-toolkit           # Install NVIDIA Container Toolkit
portunix install nvidia-container-toolkit --dry-run # Preview installation

# Ollama Container Operations
portunix aiops ollama container create          # Create Ollama container (with GPU if available)
portunix aiops ollama container create --cpu    # Force CPU-only mode
portunix aiops ollama container status          # Show Ollama container status
portunix aiops ollama container start           # Start stopped Ollama container
portunix aiops ollama container stop            # Stop running Ollama container
portunix aiops ollama container remove          # Remove Ollama container

# Model Operations (require --container or use default "portunix-ollama")
portunix aiops model list                          # List models in default container
portunix aiops model list --container my-ollama    # List models in specific container
portunix aiops model list --available              # List available models from Ollama registry
portunix aiops model install <name>                # Install model to default container
portunix aiops model install <name> --container my-ollama  # Install to specific container
portunix aiops model install <name> --size <variant>       # Install specific variant (7b, 13b, 70b)
portunix aiops model info <name>                   # Show model details (size, parameters, quantization)
portunix aiops model remove <name>                 # Remove model from default container
portunix aiops model remove <name> --container my-ollama   # Remove from specific container
portunix aiops model run <name>                    # Interactive chat (default container)
portunix aiops model run <name> --container my-ollama      # Chat with specific container
portunix aiops model run <name> --prompt "..."     # Single prompt execution

# Open WebUI Container Operations
portunix aiops webui container create           # Create Open WebUI container
portunix aiops webui container status           # Show WebUI container status
portunix aiops webui container start            # Start stopped WebUI container
portunix aiops webui container stop             # Stop running WebUI container
portunix aiops webui container remove           # Remove WebUI container
portunix aiops webui open                       # Open WebUI in browser

# Stack Operations (All Containers)
portunix aiops stack create            # Create full stack (Ollama + WebUI containers)
portunix aiops stack create --models llama3.2,mistral  # Create with pre-installed models
portunix aiops stack status            # Show all containers status
portunix aiops stack start             # Start all stopped containers
portunix aiops stack stop              # Stop all running containers
portunix aiops stack remove            # Remove all containers
```

## Model Registry

### Supported Model Providers

#### Ollama Models (Primary - Phase 2)

Models pulled from Ollama library via containerized Ollama instance.

| Model | Variants | Size Range | Use Case |
| ----- | -------- | ---------- | -------- |
| llama3.2 | 1b, 3b | 1.3GB - 2.0GB | Lightweight, fast inference |
| llama3.1 | 8b, 70b, 405b | 4.7GB - 231GB | General purpose, high quality |
| mistral | 7b | 4.1GB | Fast, balanced performance |
| mixtral | 8x7b, 8x22b | 26GB - 80GB | Mixture of experts |
| codellama | 7b, 13b, 34b, 70b | 3.8GB - 40GB | Code generation |
| phi3 | mini, medium | 2.2GB - 7.9GB | Microsoft's efficient model |
| gemma2 | 2b, 9b, 27b | 1.6GB - 16GB | Google's open model |
| qwen2.5 | 0.5b - 72b | 0.4GB - 47GB | Alibaba's multilingual |
| deepseek-coder-v2 | 16b, 236b | 8.9GB - 133GB | Advanced coding |

#### Future Providers (Planned)

- **Hugging Face**: Direct model downloads
- **LocalAI**: OpenAI-compatible API
- **vLLM**: High-throughput serving

### Expected Output Examples

**GPU Status:**

```text
$ portunix aiops gpu status

NVIDIA GPU Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GPU 0: NVIDIA GeForce RTX 4080
  Driver Version:    560.35.03
  CUDA Version:      12.6

  Utilization:       45%  [████████████░░░░░░░░░░░░░░░░░]
  Memory:            8.2 GB / 16.0 GB (51%)
                     [██████████████░░░░░░░░░░░░░░░]
  Temperature:       62°C
  Power:             185W / 320W (58%)
  Fan Speed:         45%

Container Toolkit:   ✓ Installed (v1.14.3)
Container Runtime:   Docker (GPU access verified)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**GPU Status (No Toolkit):**

```text
$ portunix aiops gpu status

NVIDIA GPU Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GPU 0: NVIDIA GeForce RTX 4080
  Driver Version:    560.35.03
  CUDA Version:      12.6

  Utilization:       12%  [███░░░░░░░░░░░░░░░░░░░░░░░░░░]
  Memory:            1.2 GB / 16.0 GB (8%)
  Temperature:       42°C

Container Toolkit:   ✗ Not installed
                     Run: portunix install nvidia-container-toolkit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**GPU Status (Watch Mode):**

```text
$ portunix aiops gpu status --watch

NVIDIA GPU Monitor (Refreshing every 1s, press Ctrl+C to exit)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GPU 0: NVIDIA GeForce RTX 4080                              2025-11-29 15:42:33
┌──────────────┬────────────────────────────────────────────────────────┬───────┐
│ UTILIZATION  │ ████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │  38%  │
│ MEMORY       │ ██████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │  51%  │
│ TEMPERATURE  │ ████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │  62°C │
│ POWER        │ ███████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░ │ 185W  │
│ FAN          │ ██████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │  45%  │
└──────────────┴────────────────────────────────────────────────────────┴───────┘

Memory: 8.2 GB / 16.0 GB    Power Limit: 320W    CUDA: 12.6
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**GPU Processes:**

```text
$ portunix aiops gpu processes

GPU Processes:
┌───────┬────────────────────────────────┬────────────┬─────────────┐
│ PID   │ PROCESS                        │ GPU MEMORY │ GPU USAGE   │
├───────┼────────────────────────────────┼────────────┼─────────────┤
│ 15234 │ ollama_llama_server            │ 4.2 GB     │ 35%         │
│ 12456 │ /usr/bin/gnome-shell           │ 512 MB     │ 2%          │
│ 18901 │ firefox                        │ 256 MB     │ 1%          │
└───────┴────────────────────────────────┴────────────┴─────────────┘

Total GPU Memory Used: 4.97 GB / 16.0 GB (31%)
```

**GPU Usage (Summary):**

```text
$ portunix aiops gpu usage

GPU Utilization Summary:
┌─────────────────────────────────────────────────────────────────────────────┐
│ GPU 0: NVIDIA GeForce RTX 4080                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ Compute:  ████████████░░░░░░░░  38%   │ Memory:   ██████████░░░░░░░░░░  51% │
│ Encode:   ░░░░░░░░░░░░░░░░░░░░   0%   │ Decode:   ░░░░░░░░░░░░░░░░░░░░   0% │
├─────────────────────────────────────────────────────────────────────────────┤
│ Temp: 62°C  │  Power: 185W/320W  │  Fan: 45%  │  Memory: 8.2/16.0 GB       │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Model List (Installed):**

```text
$ portunix aiops model list
Using container: portunix-ollama (default)

Installed Models:
┌─────────────┬────────────┬─────────────────┬──────────────┬─────────────────────┐
│ NAME        │ SIZE       │ PARAMETERS      │ QUANTIZATION │ MODIFIED            │
├─────────────┼────────────┼─────────────────┼──────────────┼─────────────────────┤
│ llama3.2:3b │ 2.0 GB     │ 3B              │ Q4_K_M       │ 2025-11-28 14:32:15 │
│ mistral:7b  │ 4.1 GB     │ 7B              │ Q4_0         │ 2025-11-27 09:15:42 │
│ codellama   │ 3.8 GB     │ 7B              │ Q4_K_M       │ 2025-11-25 16:20:33 │
└─────────────┴────────────┴─────────────────┴──────────────┴─────────────────────┘

Total: 3 models (9.9 GB)
```

**Model List (Available):**

```text
$ portunix aiops model list --available

Available Ollama Models:
┌──────────────────────┬─────────────────────────────────────────┬───────────────────┐
│ NAME                 │ DESCRIPTION                             │ SIZES             │
├──────────────────────┼─────────────────────────────────────────┼───────────────────┤
│ llama3.2             │ Meta's latest lightweight model         │ 1b, 3b            │
│ llama3.1             │ General purpose, high quality           │ 8b, 70b, 405b     │
│ mistral              │ Fast and efficient 7B model             │ 7b                │
│ mixtral              │ Mixture of experts architecture         │ 8x7b, 8x22b       │
│ codellama            │ Specialized for code generation         │ 7b, 13b, 34b, 70b │
│ phi3                 │ Microsoft's efficient model             │ mini, medium      │
│ gemma2               │ Google's open model                     │ 2b, 9b, 27b       │
│ qwen2.5              │ Alibaba's multilingual model            │ 0.5b - 72b        │
│ deepseek-coder-v2    │ Advanced coding capabilities            │ 16b, 236b         │
└──────────────────────┴─────────────────────────────────────────┴───────────────────┘

Use 'portunix aiops model install <name>' to download a model.
```

**Model Install (Progress):**

```text
$ portunix aiops model install llama3.2:3b
Using container: portunix-ollama (default)

Installing model: llama3.2:3b
Pulling manifest... done
Downloading model layers:
  [████████████████████████████████████████] 100% 2.0 GB / 2.0 GB (12.5 MB/s)

Verifying integrity... done

✓ Model llama3.2:3b installed successfully
  Container: portunix-ollama
  Size: 2.0 GB
  Parameters: 3B
  Quantization: Q4_K_M
```

**Model Install (Custom Container):**

```text
$ portunix aiops model install mistral:7b --container my-dev-ollama
Using container: my-dev-ollama

Installing model: mistral:7b
...
```

**Model Info:**

```text
$ portunix aiops model info llama3.2

Model: llama3.2
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Display Name:    Llama 3.2
Provider:        Meta (Ollama Library)
License:         Llama 3.2 Community License

Available Variants:
  • llama3.2:1b  - 1.3 GB (1B parameters, Q4_K_M)
  • llama3.2:3b  - 2.0 GB (3B parameters, Q4_K_M)

Description:
  Meta's latest lightweight language model optimized for
  fast inference on consumer hardware. Excellent for
  general-purpose text generation, summarization, and
  simple coding tasks.

System Requirements:
  Minimum RAM:   4 GB (1b) / 6 GB (3b)
  Recommended:   8 GB RAM or 4 GB VRAM
  Disk Space:    1.3 - 2.0 GB

Use Cases:
  ✓ General text generation
  ✓ Summarization
  ✓ Simple Q&A
  ✓ Basic coding assistance

Install: portunix aiops model install llama3.2:3b
```

**Model Run (Single Prompt):**

```text
$ portunix aiops model run llama3.2:3b --prompt "What is the capital of France?"

Using model: llama3.2:3b
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The capital of France is Paris. It is the largest city in France and serves
as the country's political, economic, and cultural center. Paris is known for
its iconic landmarks such as the Eiffel Tower, the Louvre Museum, and
Notre-Dame Cathedral.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Tokens: 67 | Time: 1.2s | Speed: 55.8 tokens/s
```

**Model Run (Interactive):**

```text
$ portunix aiops model run llama3.2:3b

Using model: llama3.2:3b
Type 'exit' or Ctrl+C to quit.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

You: Hello, how are you?

AI: Hello! I'm doing well, thank you for asking. I'm an AI language model,
so I don't have feelings in the traditional sense, but I'm functioning
properly and ready to help you. How can I assist you today?

You: exit

Goodbye!
```

## Implementation Phases

### Phase 1: Foundation and GPU Detection

**Goal**: Create helper skeleton and GPU detection capabilities

#### Tasks:

1. **Helper Binary Infrastructure**
   - [x] Create `src/helpers/ptx-aiops/` directory structure
   - [x] Implement Cobra CLI structure
   - [x] Add `--version`, `--help` commands
   - [x] Set up build pipeline in Makefile/GoReleaser

2. **GPU Detection Module**
   - [x] Implement NVIDIA GPU detection (`nvidia-smi` parsing)
   - [x] Detect NVIDIA driver version
   - [x] Check CUDA availability
   - [x] Implement GPU capabilities enumeration

3. **GPU Monitoring & Utilization**
   - [x] Implement `gpu status` - basic GPU info display
   - [x] Implement `gpu status --watch` - real-time monitoring with auto-refresh
   - [x] Parse `nvidia-smi` output for utilization metrics
   - [x] Display GPU utilization, memory usage, temperature, power
   - [x] Implement `--interval` flag for custom refresh rate (default: 1s)
   - [x] Implement `gpu usage` - compact utilization summary
   - [x] Implement `gpu processes` - list GPU-consuming processes
   - [ ] Handle terminal resize in watch mode
   - [x] Graceful exit on Ctrl+C

4. **NVIDIA Container Toolkit Detection**
   - [x] Detect if toolkit is installed (`nvidia-ctk --version`)
   - [x] Implement `gpu check` - verify GPU + toolkit readiness
   - [x] Validate container GPU access
   - [x] Test GPU passthrough in containers
   - [x] Suggest `portunix install nvidia-container-toolkit` if not installed

5. **PTX-Installer Integration** (separate task, add to package registry)
   - [ ] Add `nvidia-container-toolkit` package to ptx-installer registry
   - [ ] Implement APT repository setup (Debian/Ubuntu)
   - [ ] Implement DNF repository setup (Fedora/RHEL)
   - [ ] Implement post-install runtime configuration
   - [ ] Test installation via `portunix install nvidia-container-toolkit`

#### Success Criteria:

- [x] `ptx-aiops gpu status` shows GPU information
- [x] `ptx-aiops gpu status --watch` provides real-time monitoring
- [x] `ptx-aiops gpu usage` shows compact utilization summary
- [x] `ptx-aiops gpu processes` lists GPU-consuming processes
- [x] `ptx-aiops gpu check` validates GPU + toolkit readiness
- [ ] `portunix install nvidia-container-toolkit` installs toolkit correctly
- [x] GPU detection works on systems without GPU (graceful degradation)
- [x] Watch mode refreshes correctly with configurable interval

#### Deliverable:

Working GPU detection, monitoring, and toolkit installation via ptx-installer

---

### Phase 2: Ollama Container Integration

**Goal**: Create and manage Ollama containers

#### Tasks:

1. **Ollama Container Creation**
   - [x] Implement `ollama container create` command
   - [x] Create Ollama container configuration
   - [x] Implement GPU passthrough (when available)
   - [x] Configure persistent model storage
   - [x] Set up API port mapping (default: 11434)

2. **Container Lifecycle Management (via PTX-Container)**
   - [x] Use shared container runtime detection
   - [x] Implement `ollama container create` - create new container
   - [x] Implement `ollama container start` - start stopped container
   - [x] Implement `ollama container stop` - stop running container
   - [x] Implement `ollama container status` - show container status
   - [x] Implement `ollama container remove` - remove container
   - [x] Volume mounting for model persistence
   - [x] Network configuration

3. **Model Management**
   - [x] Implement `ollama pull` command passthrough
   - [x] List installed models
   - [x] Remove models
   - [x] Show model information

4. **Health Monitoring**
   - [x] Implement health check endpoint
   - [x] Status reporting
   - [ ] Resource usage monitoring

#### Success Criteria:

- [x] `portunix aiops ollama container create` creates container
- [x] `portunix aiops ollama container start/stop` works correctly
- [x] `portunix aiops ollama container status` shows container state
- [x] Models persist across container restarts
- [x] GPU acceleration works when available
- [x] CPU fallback works correctly (`--cpu` flag)

#### Deliverable:

Fully functional Ollama container lifecycle management

---

### Phase 2.5: Model Management (Priority Implementation)

**Goal**: Implement comprehensive model listing and installation for Ollama containers
**Status**: ✅ Complete

#### Tasks:

1. **Container Selection**
   - [x] Implement `--container` flag for all model commands
   - [x] Default container name: `portunix-ollama`
   - [x] Validate container exists and is running
   - [x] Clear error message if container not found
   - [x] List available Ollama containers on error

2. **Model Listing**
   - [x] Implement `model list` - show installed models in container
   - [x] Implement `model list --container <name>` - target specific container
   - [x] Parse Ollama API response (`GET /api/tags`) with fallback to container exec
   - [x] Display model name, size, modification date, quantization
   - [x] Implement `model list --available` - embedded registry with GPU rating
   - [ ] Cache available models list (reduce API calls) - not needed (embedded)
   - [ ] Format output (table, JSON with `--format=json`) - future enhancement

3. **Model Installation**
   - [x] Implement `model install <name>` - pull model to container
   - [x] Support `--container` flag for target container
   - [x] Execute via container exec (`ollama pull`)
   - [x] Show download progress (streaming response)
   - [ ] Support variant selection (`--size 7b`, `--size 13b`) - future enhancement
   - [x] Validate container exists and is running
   - [x] Handle large downloads gracefully
   - [x] Verify installation success

4. **Model Information**
   - [x] Implement `model info <name>` - show detailed model info
   - [x] Display: parameters, quantization, license, description
   - [x] Show disk space requirements
   - [x] Show memory requirements (RAM/VRAM)
   - [x] Display model family and capabilities
   - [x] Expanded registry with 11 models (llama3.2, llama3.1, mistral, mixtral, codellama, phi3, gemma2, qwen2.5, deepseek-coder-v2, llava, starcoder2)

5. **Model Removal**
   - [x] Implement `model remove <name>` - delete model from container
   - [x] Support `--container` flag for target container
   - [x] Execute via container exec (`ollama rm`)
   - [x] Confirm before deletion (interactive mode)
   - [x] Support `--force` flag for non-interactive

6. **Model Execution**
   - [x] Implement `model run <name>` - interactive chat
   - [x] Support `--container` flag for target container
   - [x] Use container exec for Ollama (`ollama run`)
   - [x] Support `--prompt "..."` for single execution
   - [x] Stream response output
   - [ ] Handle context/conversation history - future enhancement

#### Technical Implementation:

**Ollama API Integration:**

```go
// Model listing via container exec or API
type OllamaModel struct {
    Name       string    `json:"name"`
    Size       int64     `json:"size"`
    Digest     string    `json:"digest"`
    ModifiedAt time.Time `json:"modified_at"`
    Details    struct {
        Format            string   `json:"format"`
        Family            string   `json:"family"`
        Families          []string `json:"families"`
        ParameterSize     string   `json:"parameter_size"`
        QuantizationLevel string   `json:"quantization_level"`
    } `json:"details"`
}

// List installed models
func ListInstalledModels() ([]OllamaModel, error) {
    // Option 1: Direct API call (if port exposed)
    resp, err := http.Get("http://localhost:11434/api/tags")

    // Option 2: Container exec
    // portunix container exec portunix-ollama ollama list
}

// Pull model
func PullModel(name string) error {
    // POST /api/pull with streaming progress
    payload := map[string]string{"name": name}
    // Handle streaming response for progress
}
```

**Available Models Registry (Embedded):**

```go
// Embedded registry of popular Ollama models
var OllamaModelRegistry = map[string]ModelInfo{
    "llama3.2": {
        DisplayName: "Llama 3.2",
        Variants:    []string{"1b", "3b"},
        Description: "Meta's latest lightweight model",
        License:     "Llama 3.2 Community License",
        UseCase:     "General purpose, fast inference",
    },
    "mistral": {
        DisplayName: "Mistral 7B",
        Variants:    []string{"7b"},
        Description: "Fast and efficient 7B model",
        License:     "Apache 2.0",
        UseCase:     "Balanced performance",
    },
    // ... more models
}
```

#### Success Criteria:

- [x] `portunix aiops model list` shows models in default container (portunix-ollama)
- [x] `portunix aiops model list --container my-ollama` targets specific container
- [x] `portunix aiops model list --available` shows downloadable models with GPU rating
- [x] `portunix aiops model install llama3.2` downloads to default container
- [x] `portunix aiops model install llama3.2 --container my-ollama` targets specific container
- [x] Download progress visible during installation
- [x] `portunix aiops model info llama3.2` shows model details
- [x] `portunix aiops model remove llama3.2` removes model from default container
- [x] `portunix aiops model remove llama3.2 --force` removes without confirmation
- [x] `portunix aiops model run llama3.2 --prompt "Hello"` works with default container
- [x] Clear error if container doesn't exist or is not running

#### Deliverable:

Complete model management for Ollama container ✅

---

### Phase 3: Open WebUI Container Integration

**Goal**: Create Open WebUI container with Ollama integration

#### Tasks:

1. **WebUI Container Creation**
   - [ ] Implement `webui container create` command
   - [ ] Create Open WebUI container configuration
   - [ ] Configure Ollama API connection
   - [ ] Set up persistent data volume
   - [ ] Port mapping (default: 3000 or 8080)

2. **Container Lifecycle Management**
   - [ ] Implement `webui container create` - create new container
   - [ ] Implement `webui container start` - start stopped container
   - [ ] Implement `webui container stop` - stop running container
   - [ ] Implement `webui container status` - show container status
   - [ ] Implement `webui container remove` - remove container

3. **Integration with Ollama Container**
   - [ ] Automatic Ollama endpoint discovery
   - [ ] Container networking (same network)
   - [ ] Dependency checking (Ollama container must be running)

4. **User Experience**
   - [ ] Implement `webui open` command (open browser)
   - [ ] Status display with URL
   - [ ] First-run setup guidance

#### Success Criteria:

- [ ] `portunix aiops webui container create` creates WebUI container
- [ ] `portunix aiops webui container start/stop` works correctly
- [ ] WebUI connects to Ollama container automatically
- [ ] `portunix aiops webui open` opens browser to correct URL

#### Deliverable:

Working Open WebUI container with Ollama integration

---

### Phase 4: Stack Management and Testing

**Goal**: Unified stack container management and comprehensive testing

#### Tasks:

1. **Stack Container Commands**
   - [ ] Implement `stack create` - create all containers (Ollama + WebUI)
   - [ ] Implement `stack status` - show all containers status
   - [ ] Implement `stack start` - start all stopped containers
   - [ ] Implement `stack stop` - stop all running containers
   - [ ] Implement `stack remove` - remove all containers
   - [ ] Dependency ordering (Ollama container before WebUI)
   - [ ] Support `--models` flag for pre-installing models during create

2. **Testing**
   - [ ] Unit tests for GPU detection
   - [ ] Integration tests for container lifecycle operations
   - [ ] E2E tests for full stack creation
   - [ ] Test on systems with/without GPU

3. **Documentation**
   - [ ] User documentation
   - [ ] Troubleshooting guide
   - [ ] GPU setup requirements

#### Success Criteria:

- [ ] `portunix aiops stack create` creates all containers
- [ ] `portunix aiops stack start/stop` manages all containers
- [ ] All tests pass
- [ ] Documentation complete

#### Deliverable:

Production-ready PTX-AIOps helper

---

## Technical Specifications

### Container Images

| Component | Image | Default Port |
| --------- | ----- | ------------ |
| Ollama | `ollama/ollama:latest` | 11434 |
| Open WebUI | `ghcr.io/open-webui/open-webui:main` | 3000 |

### Volume Mounts

```bash
Ollama:
  ~/.portunix/aiops/ollama/models → /root/.ollama

Open WebUI:
  ~/.portunix/aiops/webui/data → /app/backend/data
```

### GPU Passthrough Configuration

```bash
# Docker
docker run --gpus all ollama/ollama

# Podman
podman run --device nvidia.com/gpu=all ollama/ollama
```

### NVIDIA Container Toolkit Installation

```bash
# Ubuntu/Debian
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
    sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
    sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
sudo apt-get update && sudo apt-get install -y nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker

# Fedora/RHEL
curl -s -L https://nvidia.github.io/libnvidia-container/stable/rpm/nvidia-container-toolkit.repo | \
    sudo tee /etc/yum.repos.d/nvidia-container-toolkit.repo
sudo dnf install -y nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

## Integration Points

### PTX-Container Integration

```go
// Use shared container package
import "portunix.ai/portunix/src/helpers/ptx-container/engine"

// Detect runtime
runtime := engine.DetectContainerRuntime()

// Run container with GPU
config := engine.ContainerConfig{
    Image: "ollama/ollama:latest",
    Name:  "portunix-ollama",
    GPUEnabled: true,
    Volumes: []string{
        fmt.Sprintf("%s/.portunix/aiops/ollama/models:/root/.ollama", homeDir),
    },
    Ports: []string{"11434:11434"},
}
engine.RunContainer(runtime, config)
```

### PTX-Installer Integration

Add NVIDIA Container Toolkit to ptx-installer package registry.
This enables installation via standard `portunix install nvidia-container-toolkit` command:

```json
{
  "name": "nvidia-container-toolkit",
  "displayName": "NVIDIA Container Toolkit",
  "description": "Toolkit for running GPU-accelerated containers",
  "category": "infrastructure/container",
  "platforms": {
    "linux": {
      "debian": {
        "type": "apt-repo",
        "repo": "https://nvidia.github.io/libnvidia-container/stable/deb",
        "keyUrl": "https://nvidia.github.io/libnvidia-container/gpgkey",
        "package": "nvidia-container-toolkit"
      },
      "fedora": {
        "type": "dnf-repo",
        "repo": "https://nvidia.github.io/libnvidia-container/stable/rpm/nvidia-container-toolkit.repo",
        "package": "nvidia-container-toolkit"
      }
    }
  },
  "postInstall": [
    "nvidia-ctk runtime configure --runtime=${CONTAINER_RUNTIME}",
    "systemctl restart ${CONTAINER_RUNTIME}"
  ]
}
```

### Shared Platform Utilities

```go
// Use shared platform package (ADR-026)
import "portunix.ai/portunix/src/pkg/platform"

// Check GPU availability
func IsNvidiaGPUAvailable() bool {
    if platform.IsWindows() {
        // Windows GPU detection
    } else if platform.IsLinux() {
        // Linux nvidia-smi check
    }
}
```

## Dependencies

### Technical Dependencies

- Go 1.21+ for implementation
- Cobra CLI framework
- Shared `ptx-container` code
- Shared `pkg/platform` utilities
- PTX-Installer for toolkit installation

### External Dependencies

- NVIDIA Driver (user-provided)
- Docker or Podman runtime
- Network access for container images

## Risk Assessment

### Technical Risks

**Risk**: NVIDIA driver version incompatibility
**Impact**: Medium
**Probability**: Medium
**Mitigation**: Document minimum driver versions, provide version checking

**Risk**: Container runtime differences (Docker vs Podman GPU support)
**Impact**: Medium
**Probability**: Low
**Mitigation**: Unified abstraction layer via ptx-container

**Risk**: GPU not available, breaking user experience
**Impact**: Low
**Probability**: High
**Mitigation**: Graceful CPU fallback, clear messaging

## Acceptance Criteria

### Functional Requirements

- [ ] GPU detection works correctly (with and without GPU)
- [ ] NVIDIA Container Toolkit installation automated
- [ ] Ollama container deploys with GPU support
- [ ] Open WebUI connects to Ollama automatically
- [ ] Stack deployment works with single command
- [ ] All containers persist data correctly

### Performance Requirements

- [ ] GPU passthrough achieves native performance
- [ ] Container startup time < 30 seconds
- [ ] Model loading times comparable to native Ollama

### Quality Requirements

- [ ] Graceful degradation without GPU
- [ ] Clear error messages
- [ ] Help text for all commands
- [ ] Cross-platform consistency (Docker/Podman)

## Labels

- enhancement
- helper-binary
- ai-integration
- container
- gpu-support
- high-priority

## Notes

### Future Enhancements

- Additional AI tools (LocalAI, vLLM, text-generation-webui)
- Multi-GPU support
- Distributed inference
- Model quantization helpers
- Fine-tuning support
- ComfyUI integration for image generation

### Design Principles

1. **Container-Native**: All AI tools run in containers
2. **GPU-Aware**: Automatic detection and configuration
3. **Graceful Degradation**: Works without GPU (CPU mode)
4. **Integration**: Reuses existing ptx-container code
5. **User-Friendly**: Single commands for complex setups

---

**Created**: 2025-11-29
**Status**: Open - awaiting implementation assignment
