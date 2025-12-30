# PTX-Prompting Helper

Template-based prompt generation for AI assistants.

## Overview

The `ptx-prompting` helper provides functionality for loading prompt templates, detecting placeholders, and generating customized prompts for AI assistants like Claude, ChatGPT, etc.

## Features

- Load templates from .md, .yaml, .txt files
- Automatic placeholder detection (`{placeholder}` format)
- CLI arguments and interactive parameter input
- Output to stdout, clipboard, or file
- Multilingual template support (en/, cs/)

## Usage

### Via Portunix Dispatcher

```bash
# List available templates
portunix prompt list

# Build prompt from template
portunix prompt build templates/en/translate.md --var source_file=README.md --var target_language=English

# Create new template
portunix prompt create my-template.md --type translation --lang cs
```

### Direct Usage

```bash
# List templates
ptx-prompting list

# Build prompt
ptx-prompting build templates/en/translate.md --var source_file=README.md --copy

# Create template
ptx-prompting create review-template.md --interactive
```

## Commands

### build
Build a prompt from a template file.

**Flags:**
- `--copy` - Copy generated prompt to clipboard
- `--output`, `-o` - Write output to file
- `--var` - Set template variable (--var key=value)
- `--interactive`, `-i` - Interactive mode for missing variables
- `--allow-incomplete` - Allow build with missing variables
- `--preview` - Preview what will be built without building
- `--verbose`, `-v` - Show verbose output

### list
List available prompt templates.

**Flags:**
- `--lang` - Filter templates by language (en, cs, etc.)
- `--path` - Search templates in custom path
- `--detailed`, `-d` - Show detailed information
- `--builtin` - Show only built-in templates

### create
Create a new prompt template file.

**Flags:**
- `--lang` - Template language (default: en)
- `--description` - Template description
- `--interactive`, `-i` - Interactive template creation
- `--dir` - Target directory for template
- `--type` - Template type (translation, review, debug, custom)
- `--force` - Overwrite existing template

## Template Format

Templates use simple placeholder syntax:

```markdown
# Translation Request

Please translate the file {source_file} from {source_language} to {target_language}.
Save the result as {target_file}.

Requirements:
- Target audience: {audience}
```

## Built-in Templates

- `en/translate.md` - Translation request template
- `en/review.md` - Code review request template
- `cs/translate.md` - Czech translation template

## Architecture

```
ptx-prompting/
├── main.go                 # Entry point with helper integration
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command
│   ├── build.go           # Build command
│   ├── list.go            # List templates
│   └── create.go          # Create template
├── internal/              # Internal packages
│   ├── parser/            # Template parsing
│   ├── prompt/            # Prompt building
│   └── clipboard/         # Clipboard integration
└── templates/             # Default templates
    ├── en/                # English templates
    └── cs/                # Czech templates
```

## Integration

The helper integrates with the main Portunix dispatcher system:
- Binary name: `ptx-prompting`
- Commands handled: `prompt`
- Helper discovery: Automatic via `HelperDiscovery` system
- Version: Synchronized with main Portunix version

## Development

```bash
# Build helper
go build -o ptx-prompting .

# Test integration
./ptx-prompting --version
./ptx-prompting --list-commands
./ptx-prompting prompt list
```