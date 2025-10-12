# PDF Generation Workflow

## Purpose
This document describes the standardized workflow for generating PDF documents from Markdown documentation using Claude Code and pandoc.

## Prerequisites

### Required Tools
- **pandoc**: Version 3.1.11.1 or higher
- **LaTeX packages**: Required for PDF generation
- **XeLaTeX engine**: For better Unicode and font support

### Installation Commands
```bash
# Install pandoc and LaTeX packages
sudo apt install -y pandoc texlive-xetex texlive-lang-czechslovak texlive-latex-base texlive-fonts-recommended texlive-latex-extra

# Verify installation
pandoc --version
```

## Workflow Process

### Step 1: Markdown Preparation
Before generating PDF, ensure proper Markdown formatting for lists and sections:

**Critical formatting rules:**
- Add empty line after section headers that precede lists
- Add empty line after bold text that introduces lists
- Proper list structure with `-` bullets

**Example of correct formatting:**
```markdown
**Important notes:**

- First item
- Second item
- Third item
```

**Incorrect formatting (will merge into paragraph):**
```markdown
**Important notes:**
- First item
- Second item
- Third item
```

### Step 2: PDF Generation Command
Use the following standardized pandoc command:

```bash
pandoc FILENAME.md -o FILENAME.pdf \
  --pdf-engine=xelatex \
  -V geometry:margin=2cm \
  -V documentclass=article \
  -V lang=cs \
  -V fontsize=11pt \
  -V mainfont="DejaVu Sans" \
  -V monofont="DejaVu Sans Mono" \
  --highlight-style=tango \
  --toc \
  --toc-depth=2
```

### Step 3: Quality Check
After PDF generation:
1. Check file size and verify creation: `ls -lh FILENAME.pdf`
2. Verify PDF structure: `file FILENAME.pdf`
3. Review formatting, especially:
   - Lists display as proper bullet points
   - Code blocks are highlighted
   - Table of contents is generated
   - Czech characters display correctly

## Common Issues and Solutions

### Issue: Lists Display as Paragraphs
**Symptom**: Bullet points merge into single paragraph
**Cause**: Missing empty line before list
**Solution**: Add empty line after header/bold text that precedes list

### Issue: LaTeX Package Errors
**Symptom**: `! LaTeX Error: File 'xcolor.sty' not found`
**Cause**: Missing LaTeX packages
**Solution**: Install additional packages:
```bash
sudo apt install -y texlive-xetex texlive-lang-czechslovak
```

### Issue: Font Problems
**Symptom**: Characters not displaying correctly
**Cause**: Missing fonts or incorrect engine
**Solution**: Use XeLaTeX engine with DejaVu fonts as shown in command above

### Issue: PDF Not Generated
**Symptom**: Command runs without errors but no PDF created
**Cause**: Various LaTeX compilation issues
**Solution**: Check current directory and file permissions

## File Organization

### For Translated Documents
- **Source**: `docs/contributing/tutorials/FILENAME.md`
- **Translation**: `.translated/cs/docs/contributing/tutorials/FILENAME.md`
- **PDF Output**: `.translated/cs/docs/contributing/tutorials/FILENAME.pdf`

### For Original Documents
- **Source**: `docs/contributing/FILENAME.md`  
- **PDF Output**: `docs/contributing/FILENAME.pdf`

## Markdown Formatting Best Practices

### Section Headers and Lists
```markdown
## Section Title

### Subsection with List

**List introduction:**

- Item 1
- Item 2
- Item 3
```

### Code Examples
````markdown
**Example usage:**
```xml
<!-- XML code here -->
```
````

### Tables
```markdown
| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data 1   | Data 2   | Data 3   |
```

### Important Notes
```markdown
**⚠️ IMPORTANT:**

- Critical point 1
- Critical point 2
```

## Pandoc Command Parameters Explained

| Parameter | Purpose | Value |
|-----------|---------|-------|
| `--pdf-engine=xelatex` | PDF generation engine | XeLaTeX for Unicode support |
| `-V geometry:margin=2cm` | Page margins | 2cm all sides |
| `-V documentclass=article` | Document class | Standard article format |
| `-V lang=cs` | Language setting | Czech language support |
| `-V fontsize=11pt` | Font size | 11 point readable size |
| `-V mainfont="DejaVu Sans"` | Main font | Unicode-compatible font |
| `-V monofont="DejaVu Sans Mono"` | Monospace font | For code blocks |
| `--highlight-style=tango` | Code highlighting | Tango color scheme |
| `--toc` | Table of contents | Generate TOC |
| `--toc-depth=2` | TOC depth | Include h1 and h2 headings |

## Automation Considerations

For frequently updated documents, consider creating a script:

```bash
#!/bin/bash
# pdf-generate.sh

SOURCE_FILE="$1"
OUTPUT_FILE="${SOURCE_FILE%.md}.pdf"

echo "Generating PDF from $SOURCE_FILE..."

pandoc "$SOURCE_FILE" -o "$OUTPUT_FILE" \
  --pdf-engine=xelatex \
  -V geometry:margin=2cm \
  -V documentclass=article \
  -V lang=cs \
  -V fontsize=11pt \
  -V mainfont="DejaVu Sans" \
  -V monofont="DejaVu Sans Mono" \
  --highlight-style=tango \
  --toc \
  --toc-depth=2

if [ $? -eq 0 ]; then
    echo "PDF generated successfully: $OUTPUT_FILE"
    ls -lh "$OUTPUT_FILE"
else
    echo "Error generating PDF"
fi
```

## Quality Assurance Checklist

Before distributing PDF:
- [ ] All lists display as bullet points (not merged paragraphs)
- [ ] Code blocks are syntax highlighted
- [ ] Table of contents is complete and accurate
- [ ] Czech characters display correctly
- [ ] File size is reasonable (typically 50-200KB for documentation)
- [ ] PDF opens correctly in standard viewers
- [ ] All internal links work (if any)
- [ ] Formatting is consistent throughout

## Version Control

- **Include in Git**: Source Markdown files
- **Exclude from Git**: Generated PDF files (add to .gitignore)
- **Regenerate**: PDFs should be regenerated when source changes
- **Distribution**: PDFs can be generated on-demand or as part of CI/CD

## Maintenance

### Regular Updates
- Review workflow quarterly
- Update pandoc and LaTeX packages as needed
- Test PDF generation with new document types
- Refine formatting rules based on common issues

### Documentation Updates
When source documentation changes significantly:
1. Update source Markdown first
2. Regenerate PDF using standard command
3. Review formatting and fix any issues
4. No need to maintain PDF version history

---

**Note**: This workflow is designed for internal team documentation. For external distribution, additional formatting and branding considerations may apply.

*Created: 2025-09-07*
*Last updated: 2025-09-07*