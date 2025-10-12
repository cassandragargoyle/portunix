# PTX-Prompting Interactive Mode

## Overview
PTX-Prompting helper supports an intelligent interactive mode that makes it easy to fill in template placeholders without having to remember all parameter names or provide them via command line.

## Features

### 1. Automatic Interactive Mode Activation
When you run the `build` command without providing values for required placeholders, the tool automatically switches to interactive mode:

```bash
# Template with 3 placeholders: {file_path}, {programming_language}, {focus_area}
ptx-prompting build template.md

# Automatically prompts for all 3 missing values
```

### 2. Partial Value Support
You can provide some values via command line and be prompted only for missing ones:

```bash
# Provide 2 values, get prompted for the 3rd
ptx-prompting build template.md --var file_path=main.go --var programming_language=Go

# Will only prompt for: focus_area
```

### 3. Explicit Interactive Mode
Force interactive mode even when values could be optional:

```bash
ptx-prompting build template.md --interactive
# or
ptx-prompting build template.md -i
```

### 4. Smart Defaults
The interactive mode provides intelligent defaults for common placeholders:

- `target_language`: Defaults to "English"
- `source_language`: Defaults to "Czech"
- `programming_language`: Defaults to "Go"
- `audience`: Defaults to "developers"

When a default is available, you can press Enter to accept it.

### 5. Contextual Help
Each placeholder shows helpful descriptions based on its name:

- Placeholders ending with `_file`: "Path to a file"
- Placeholders ending with `_path`: "Path to a file or directory"
- Placeholders ending with `_language`: "Language name"
- Common placeholders have specific help text (e.g., `source_file`, `context_description`)

### 6. Friendly Display Names
Placeholder names are automatically converted to readable format:
- `source_file` → "Source File"
- `programming_language` → "Programming Language"
- `focus_area_1` → "Focus Area 1"

## Interactive Mode Flow

1. **Template Analysis**: The tool analyzes your template to find all placeholders
2. **Value Resolution**: Checks which values are already provided (via CLI or defaults)
3. **Interactive Prompts**: For each missing value:
   - Shows a friendly name
   - Displays contextual help (if available)
   - Shows default value (if available)
   - Waits for user input
4. **Confirmation**: Shows what value was set after each input
5. **Build Completion**: Generates the final prompt with all values filled

## Example Session

```bash
$ ptx-prompting build code-review.md --var file_path=main.go

📝 Interactive mode - Please provide values for template placeholders:
────────────────────────────────────────────────────────────

[1/2] 📌 Programming Language
   ℹ️  Programming language being used (e.g., Go, Python, JavaScript)
   Default: Go
   Enter value (or press Enter for default):
   ✓ Using default: Go

[2/2] 📌 Focus Area
   ℹ️  First area to focus on during review
   Enter value: security
   ✓ Set to: security

────────────────────────────────────────────────────────────

# Code Review Request

Please review the code in main.go.

Language: Go
Focus: security
```

## Advanced Usage

### Combining with Other Flags

Interactive mode works seamlessly with other build options:

```bash
# Interactive mode + copy to clipboard
ptx-prompting build template.md --copy

# Interactive mode + save to file
ptx-prompting build template.md -o output.txt

# Interactive mode + verbose output
ptx-prompting build template.md -v
```

### Preview Before Interactive Mode

Check what placeholders exist before entering interactive mode:

```bash
ptx-prompting build template.md --preview
```

### Allow Incomplete Builds

Skip interactive mode and allow missing placeholders:

```bash
ptx-prompting build template.md --allow-incomplete
```

## Tips and Best Practices

1. **Use Preview First**: Run with `--preview` to see all required placeholders before starting
2. **Partial Automation**: Provide common values via CLI, use interactive for variable ones
3. **Default Values File**: Create a defaults file for frequently used values (coming soon)
4. **Template Organization**: Name your placeholders descriptively for better help text

## Placeholder Naming Conventions

For best interactive experience, follow these naming patterns:

- **Files**: Use `_file` suffix (e.g., `source_file`, `config_file`)
- **Paths**: Use `_path` suffix (e.g., `project_path`, `output_path`)
- **Languages**: Use `_language` suffix (e.g., `programming_language`, `target_language`)
- **Descriptions**: Use `_description` suffix (e.g., `context_description`, `task_description`)
- **Numbered items**: Use `_1`, `_2`, etc. (e.g., `focus_area_1`, `focus_area_2`)

## Troubleshooting

### Issue: Not entering interactive mode automatically
**Solution**: Check if you have `--allow-incomplete` flag set, which disables auto-interactive mode

### Issue: Want to skip interactive mode
**Solution**: Use `--allow-incomplete` flag or provide all values via `--var` flags

### Issue: Need to provide many values repeatedly
**Solution**: Consider creating a defaults file or using command history with `--var` flags

## Integration with Portunix

When used through the main Portunix dispatcher:

```bash
# Via Portunix main command
portunix prompt build template.md

# Same interactive features apply
portunix prompt build template.md --var file_path=main.go
```

## Future Enhancements

- [ ] Support for default values from JSON/YAML files
- [ ] Template-specific help text definitions
- [ ] Value validation during interactive input
- [ ] Multi-select options for certain placeholders
- [ ] History of previously used values