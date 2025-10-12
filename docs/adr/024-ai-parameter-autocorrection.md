# ADR-024: AI Parameter Autocorrection System

**Status:** Proposed
**Date:** 2025-09-30
**Architect:** Claude (AI Assistant)

## Context

AI assistants (Claude Code, GitHub Copilot, ChatGPT, etc.) frequently generate commands with parameter variations that are syntactically incorrect but semantically clear. These variations arise from:

1. **Natural Language Understanding**: AI interprets intent but may use non-standard parameter names
2. **Cross-Platform Confusion**: Mixing Docker/Podman/Kubernetes parameter conventions
3. **Spelling Variations**: Common typos or alternative spellings of parameters
4. **Verbose vs Short Forms**: Inconsistency between `--detached` vs `--detach`

### Real-World Example

```bash
# AI-generated command (incorrect)
./portunix container run ubuntu:22.04 --name hugo-test --detached --command "sleep infinity"

# Correct command
./portunix container run ubuntu:22.04 --name hugo-test --detach --command "sleep infinity"

# Issue: AI used --detached instead of --detach
```

### Current Behavior

```
❌ Error: unknown flag: --detached
See 'portunix container run --help'
```

User must:
1. Read error message
2. Check help documentation
3. Identify correct parameter
4. Manually fix command
5. Re-execute

This creates friction, especially when AI assistants are recommending commands.

### Frequency Analysis

Common AI parameter mistakes observed:
- `--detached` → `--detach` (Docker-style confusion)
- `--interactive` → `-it` (verbose vs shorthand)
- `--volumne` → `--volume` (typo)
- `--enviroment` → `--environment` (typo)
- `--deamon` → `--daemon` (typo)
- `--priviledged` → `--privileged` (typo)

## Decision

Implement **AI Parameter Autocorrection System** that operates in the **command parsing phase**, before any command execution:

1. **Pre-processes** command arguments during initial parsing
2. **Applies** fixed substitution rules based on known AI patterns
3. **Corrects** parameters before command dispatch/execution
4. **Logs** corrections transparently for user awareness
5. **Maintains** backward compatibility with correct syntax
6. **Avoids** documenting corrections in help text (silent feature)

### Design Philosophy

**"Be liberal in what you accept, conservative in what you output"** - Postel's Law

The system should:
- Operate in parser phase, before command execution
- Use fixed, deterministic substitution rules
- Apply corrections eagerly, not reactively
- Never wait for command errors to correct
- Log corrections for debugging but not intrude
- Never break existing correct commands

### Critical Design Constraint

**Pre-execution correction**: The system MUST correct parameters during argument parsing, not after command execution or error detection. This ensures:
- No wasted command execution attempts
- Immediate correction without retry logic
- Clean separation from command implementation
- Predictable, deterministic behavior

## Architecture

### 1. Parameter Correction Engine (Pre-Execution)

```go
package parser

import (
    "strings"
)

// ParameterCorrector handles AI-generated parameter variations
// IMPORTANT: This operates BEFORE command execution, during parsing phase
type ParameterCorrector struct {
    substitutions map[string]string  // Direct substitution map: incorrect -> correct
    enabled       bool
    logger        *Logger
}

// SubstitutionRule defines a fixed parameter substitution
type SubstitutionRule struct {
    From   string   // Incorrect parameter (normalized, without dashes)
    To     string   // Correct parameter (with dashes)
    Source string   // Why: "typo", "verbose", "cross-platform", "ai-pattern"
}

// NewParameterCorrector creates correction engine with fixed substitution rules
func NewParameterCorrector(logger *Logger) *ParameterCorrector {
    pc := &ParameterCorrector{
        substitutions: make(map[string]string),
        enabled:       true,
        logger:        logger,
    }

    pc.loadFixedSubstitutions()
    return pc
}

// CorrectParameters processes command arguments BEFORE execution
// This is called during initial parsing, not after command failure
func (pc *ParameterCorrector) CorrectParameters(args []string) []string {
    if !pc.enabled {
        return args
    }

    correctedArgs := make([]string, 0, len(args))
    correctionsMade := 0

    for i, arg := range args {
        // Non-flag arguments pass through unchanged
        if !strings.HasPrefix(arg, "-") {
            correctedArgs = append(correctedArgs, arg)
            continue
        }

        // Extract flag name (strip leading dashes)
        flagName := strings.TrimLeft(arg, "-")

        // Check fixed substitution map
        if correctFlag, found := pc.substitutions[strings.ToLower(flagName)]; found {
            pc.logCorrection(arg, correctFlag, i)
            correctedArgs = append(correctedArgs, correctFlag)
            correctionsMade++
        } else {
            // No substitution needed, use original
            correctedArgs = append(correctedArgs, arg)
        }
    }

    if correctionsMade > 0 {
        pc.logger.Debug("AI parameter autocorrection: %d parameter(s) corrected before execution", correctionsMade)
    }

    return correctedArgs
}

// logCorrection logs parameter correction for debugging
func (pc *ParameterCorrector) logCorrection(original, corrected string, position int) {
    pc.logger.Debug(
        "Pre-execution correction at position %d: %s → %s",
        position,
        original,
        corrected,
    )
}
```

### 2. Fixed Substitution Rules (Loaded at Startup)

```go
// loadFixedSubstitutions initializes fixed parameter substitution map
// These substitutions are applied BEFORE command execution
func (pc *ParameterCorrector) loadFixedSubstitutions() {
    // Map structure: normalized-incorrect-flag -> correct-flag-with-dashes
    substitutions := map[string]string{
        // Container detach variations
        "detached":   "--detach",
        "daemon":     "--detach",
        "daemonize":  "--detach",

        // Interactive variations
        "interactive": "-it",
        "tty":         "-it",

        // Volume mounting typos and variations
        "volumne":     "--volume",
        "volumn":      "--volume",
        "mount-vol":   "--volume",
        "vol":         "--volume",

        // Environment variable typos
        "enviroment":   "--environment",
        "environement": "--environment",
        "env-var":      "--environment",
        "envvar":       "--environment",

        // Privilege typos
        "priviledged": "--privileged",
        "privilaged":  "--privileged",
        "priviliged":  "--privileged",

        // Network variations
        "network-mode": "--network",
        "net":          "--network",

        // Image variations
        "from-image": "--image",
        "base-image": "--image",
        "img":        "--image",

        // Port mapping variations
        "publish-port": "--port",
        "expose-port":  "--port",
        "publish":      "--port",

        // Command variations
        "cmd":     "--command",
        "execute": "--command",
        "exec":    "--command",

        // Name variations
        "container-name": "--name",
        "hostname":       "--name",
    }

    // Load into substitutions map
    for from, to := range substitutions {
        pc.substitutions[strings.ToLower(from)] = to
        pc.logger.Debug("Loaded substitution rule: --%s → %s", from, to)
    }

    pc.logger.Info("Loaded %d fixed parameter substitution rules", len(pc.substitutions))
}

// AddCustomSubstitution allows runtime addition of custom substitutions
func (pc *ParameterCorrector) AddCustomSubstitution(from, to string) {
    from = strings.ToLower(strings.TrimLeft(from, "-"))
    pc.substitutions[from] = to
    pc.logger.Debug("Added custom substitution: %s → %s", from, to)
}
```

### 3. Integration with Command Parser (Main Entry Point)

```go
// parser/parser.go

type Parser struct {
    corrector *ParameterCorrector
    logger    *Logger
}

func NewParser(logger *Logger) *Parser {
    return &Parser{
        corrector: NewParameterCorrector(logger),
        logger:    logger,
    }
}

// Parse is the main entry point - corrects parameters FIRST, then parses
func (p *Parser) Parse(args []string) (*Command, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("no command provided")
    }

    // STEP 1: Apply fixed parameter substitutions BEFORE parsing
    // This happens regardless of command type or context
    correctedArgs := p.corrector.CorrectParameters(args)

    // STEP 2: Continue with standard command parsing using corrected args
    return p.parseCommand(correctedArgs)
}

// parseCommand handles actual command structure parsing
func (p *Parser) parseCommand(args []string) (*Command, error) {
    // Standard command parsing logic
    // At this point, all AI parameter mistakes are already fixed
    // ...
}
```

### Key Integration Points

```go
// main.go - Application entry point

func main() {
    logger := logging.NewLogger()
    parser := parser.NewParser(logger)

    // Parse arguments - corrections happen automatically
    cmd, err := parser.Parse(os.Args[1:])
    if err != nil {
        log.Fatal(err)
    }

    // Execute command with corrected parameters
    if err := cmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

### 4. Configuration Support (Optional)

```go
// config/autocorrect.go

type AutocorrectConfig struct {
    Enabled              bool                          `yaml:"enabled"`
    LogLevel             string                        `yaml:"log_level"`
    CustomSubstitutions  map[string]string            `yaml:"custom_substitutions"`
}

// Example configuration file: ~/.portunix/autocorrect.yaml
/*
enabled: true
log_level: debug

# Additional custom substitutions beyond built-in rules
custom_substitutions:
  "dockerized": "--detach"
  "bg": "--detach"
  "background": "--detach"
  "mnt": "--volume"
  "bind": "--volume"
*/

// LoadAutocorrectConfig loads configuration from file
func LoadAutocorrectConfig(path string) (*AutocorrectConfig, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config AutocorrectConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}

// ApplyConfig applies configuration to corrector
func (pc *ParameterCorrector) ApplyConfig(config *AutocorrectConfig) {
    pc.enabled = config.Enabled

    // Add custom substitutions
    for from, to := range config.CustomSubstitutions {
        pc.AddCustomSubstitution(from, to)
    }
}
```

### 5. User Experience

#### Transparent Correction (Default - Silent Mode)

```bash
# User runs AI-generated command with incorrect parameter
$ ./portunix container run ubuntu:22.04 --name hugo-test --detached --command "sleep infinity"

# Parameter is corrected BEFORE execution, command succeeds immediately
# No error, no retry, just works
✅ Container hugo-test created successfully

# Internally, before command execution:
# args: ["container", "run", "ubuntu:22.04", "--name", "hugo-test", "--detached", "--command", "sleep infinity"]
# corrected: ["container", "run", "ubuntu:22.04", "--name", "hugo-test", "--detach", "--command", "sleep infinity"]
```

#### Verbose Mode (Debug Logging)

```bash
$ PORTUNIX_DEBUG=1 ./portunix container run ubuntu:22.04 --detached

🔧 Pre-execution parameter corrections:
   Position 3: --detached → --detach (fixed substitution)
   Total corrections: 1

✅ Container started successfully
```

#### Disabled Autocorrection

```bash
$ PORTUNIX_AUTOCORRECT=false ./portunix container run ubuntu:22.04 --detached

❌ Error: unknown flag: --detached
See 'portunix container run --help'
```

## Consequences

### Positive

1. **Pre-Execution Correction**: Parameters corrected before any execution attempt, no wasted cycles
2. **Improved AI Integration**: AI assistants can recommend commands without perfect parameter knowledge
3. **Zero Retries**: Commands succeed on first attempt, no error handling needed
4. **Better UX**: Users don't need to manually fix obvious mistakes
5. **Learning Curve Reduction**: New users can use intuitive parameter names
6. **Cross-Platform Familiarity**: Users familiar with Docker can use Docker conventions
7. **Error Reduction**: Fewer failed commands due to parameter typos
8. **Silent Operation**: Works transparently without cluttering help text
9. **Simple Implementation**: Fixed substitution map, O(1) lookup performance

### Negative

1. **Hidden Behavior**: Users may not realize corrections are happening
2. **Global Scope**: Substitutions apply to all commands (not context-aware)
3. **Maintenance Overhead**: Substitution rules need updating as commands evolve
4. **Potential False Positives**: Some substitutions might apply incorrectly in edge cases
5. **Help Text Misalignment**: Help shows canonical parameters, not accepted variations

### Risks and Mitigation

#### Risk 1: False Positive Corrections
- **Risk**: Substitution applies when it shouldn't (e.g., `--exec` in different context)
- **Mitigation**: Conservative substitution rules, only obvious/unambiguous cases
- **Detection**: User feedback, integration testing

#### Risk 2: Help Text Misalignment
- **Risk**: Help shows `--detach`, AI uses `--detached`, both work but not documented
- **Mitigation**: Silent feature by design, debug logging available
- **Detection**: User confusion reports, AI prompt updates

#### Risk 3: Maintenance Burden
- **Risk**: Substitution map needs updates as commands evolve
- **Mitigation**: Centralized substitution map, easy to update
- **Detection**: User reports, AI pattern analysis

## Implementation Guidelines

### Testing Strategy

```go
// parser/corrector_test.go

func TestFixedSubstitutions(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected []string
    }{
        {
            name:     "detached to detach",
            input:    []string{"container", "run", "ubuntu:22.04", "--detached"},
            expected: []string{"container", "run", "ubuntu:22.04", "--detach"},
        },
        {
            name:     "volumne typo correction",
            input:    []string{"container", "run", "--volumne", "/data:/data"},
            expected: []string{"container", "run", "--volume", "/data:/data"},
        },
        {
            name:     "multiple corrections",
            input:    []string{"container", "run", "--detached", "--volumne", "/data"},
            expected: []string{"container", "run", "--detach", "--volume", "/data"},
        },
        {
            name:     "no correction needed",
            input:    []string{"container", "run", "--detach"},
            expected: []string{"container", "run", "--detach"},
        },
        {
            name:     "mixed correct and incorrect",
            input:    []string{"container", "run", "--name", "test", "--detached"},
            expected: []string{"container", "run", "--name", "test", "--detach"},
        },
        {
            name:     "global substitution applies everywhere",
            input:    []string{"install", "nodejs", "--detached"},
            expected: []string{"install", "nodejs", "--detach"},
            // Note: This is a design tradeoff - substitutions are global
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            corrector := NewParameterCorrector(NewTestLogger())
            result := corrector.CorrectParameters(tt.input)

            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}

func BenchmarkParameterCorrection(b *testing.B) {
    corrector := NewParameterCorrector(NewTestLogger())
    args := []string{"container", "run", "ubuntu:22.04", "--detached", "--volumne", "/data"}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = corrector.CorrectParameters(args)
    }
}
```

### Documentation Strategy

**Important**: Do NOT document autocorrections in user-facing help text.

Instead:
1. **Debug Logs**: Document corrections in debug output only
2. **Developer Docs**: Explain system in architecture documentation
3. **AI Prompts**: Include correction hints in AI assistant prompts
4. **Troubleshooting**: Mention in troubleshooting guides

## Related ADRs

- ADR-011: Multi-Level Help System
- ADR-015: Logging System Architecture
- ADR-020: AI Prompts for Package Discovery
- ADR-021: Package Registry Architecture

## Success Metrics

1. **Correction Rate**: Track how often autocorrections are applied
2. **User Reports**: Monitor user feedback on unexpected behavior
3. **AI Success**: Measure AI-generated command success rate improvement
4. **Performance**: Autocorrection adds <1ms latency per command
5. **Coverage**: Support top 20 AI-generated parameter variations

---

## Product Owner Decision

**Status: [PENDING REVIEW]**
**Date:** 2025-09-30
**Product Owner:** [To be assigned]

### Business Value Assessment

This feature enhances the **AI-first user experience** by allowing AI assistants to recommend Portunix commands with natural parameter variations.

### Key Business Benefits:

#### 1. **AI Integration Excellence** 🤖
- Positions Portunix as AI-assistant-friendly tool
- **Business Impact**: Better integration with Claude Code, GitHub Copilot, ChatGPT

#### 2. **User Experience Improvement** ✨
- Reduces command failures from parameter typos
- **Business Impact**: Lower support burden, higher user satisfaction

#### 3. **Cross-Platform Appeal** 🌍
- Users familiar with Docker can use Docker conventions
- **Business Impact**: Easier migration from other container tools

#### 4. **Learning Curve Reduction** 📚
- Intuitive parameter names work automatically
- **Business Impact**: Faster user onboarding

### Technical Considerations:

#### Implementation Complexity: Low-Medium
- Self-contained parser enhancement
- Minimal impact on existing codebase
- Clear separation of concerns

#### Maintenance Burden: Low
- Correction rules as configuration
- Easy to add/remove corrections
- No impact on core command logic

### Strategic Recommendation:

**[TO BE COMPLETED BY PRODUCT OWNER]**

Considerations for decision:
1. **AI Strategy**: Align with AI assistant integration roadmap
2. **User Feedback**: Validate need through user research
3. **Priority**: Balance against other UX improvements
4. **Risk Tolerance**: Acceptable level of autocorrection ambiguity

---

**Decision Status:** Pending Product Owner Review