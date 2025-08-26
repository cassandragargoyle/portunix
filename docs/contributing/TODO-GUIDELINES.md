# TODO Guidelines

## Purpose
This document defines the standardized format for TODO comments and placeholders used in the CassandraGargoyle team.

## TODO Format

### Standard Format
```
TODO:NNN description
TODO:NNN XX: description
```

Where:
- **TODO:** - Fixed prefix (uppercase)
- **:** - Colon separator (no space before, one space after)
- **NNN** - Three-digit number with leading zeros (001, 002, 003, etc.)
- **XX:** - Optional team initials followed by colon and space
- **description** - Brief description of what needs to be done

### Examples
```markdown
TODO:001 JS: Complete the Creating New Projects section
TODO:002 MK: Add Wiki documentation structure
TODO:003 Refine Quick Command Reference commands
```

### Team Initials Assignment (Optional)
You can optionally add team initials to assign TODO to a specific team member:

- Format: `TODO:NNN XX: description`
- Place initials after number, followed by colon and space
- Authors typically use their own initials when noting personal tasks
- Indicates "I've marked this for myself to work on later"
- When no initials are specified, TODO is unassigned

## Numbering System

### Per-File Sequential Numbering
- Each file maintains its own TODO numbering sequence
- Start from 001 in each file
- Increment sequentially within the file: 001, 002, 003, etc.
- Use leading zeros to maintain three-digit format
- Do not reuse numbers within the same file even after completion

### Temporary Placeholder
- You can temporarily write `TODO:XXX description` as a placeholder
- Ask your AI assistant to assign proper numbering based on existing TODOs in the file
- Replace XXX with the next available number sequence

### File Scope
- TODO numbers are unique only within each individual file
- Different files can have the same TODO numbers (e.g., README.md:TODO:001 and setup.sh:TODO:001)
- Check existing TODOs within the same file before assigning new numbers
- Use `grep "TODO:" filename` to find TODOs in a specific file

## Usage Guidelines

### When to Use TODOs
- ✅ Incomplete documentation sections
- ✅ Placeholder content that needs refinement
- ✅ Features planned for future implementation
- ✅ Configuration that needs customization

### When NOT to Use TODOs
- ❌ Critical functionality that breaks the system
- ❌ Security issues (use FIXME or create immediate issues)
- ❌ Simple typos or formatting issues
- ❌ Temporary debugging code

### TODO Lifecycle
1. **Creation**: Add TODO with next available number
2. **Tracking**: Reference in issues or project boards
3. **Resolution**: Replace TODO with actual content
4. **Documentation**: Update this guidelines file if needed

## Integration with Issues

### Issue Creation
When creating issues for TODOs:
- Reference TODO numbers in issue titles
- Include TODO descriptions in issue content
- Link related TODOs together when appropriate

### Example Issue Structure
```markdown
# Issue #001: Complete TODO items in README.md

## Tasks to Complete

### TODO:001 Creating New Projects Section
- **Current**: "TODO:001 Complete documentation"
- **Required**: Add proper examples and commands
- **File**: README.md
- **Location**: Creating New Projects section

### TODO:002 Wiki Documentation
- **Current**: "TODO:002 to be added"
- **Required**: Set up Wiki structure
- **File**: README.md
- **Location**: Links section
```

Note: When referencing TODOs in issues, always specify the file name since TODO numbers are per-file scoped.

## Search and Management

### Finding TODOs
```bash
# Find all TODOs in project
grep -r "TODO:" .

# Find TODOs in specific file
grep "TODO:" filename

# Find specific TODO across all files
grep -r "TODO:001" .

# Count TODOs in specific file
grep "TODO:" filename | wc -l

# Count total TODOs in project
grep -r "TODO:" . | wc -l
```

### Documentation
- Keep this guidelines file updated
- Reference TODO format in contributing documentation
- Include examples in code review templates

---

## Version History
- **v1.0** - Initial TODO guidelines established
- Created: 2025-08-23
- Last Updated: 2025-08-23

## Related Documents
- [Contributing Guidelines](../CONTRIBUTING.md)
- [Issue Templates](../../.github/ISSUE_TEMPLATE/)
- [Code Review Guidelines](./CODE-REVIEW.md)