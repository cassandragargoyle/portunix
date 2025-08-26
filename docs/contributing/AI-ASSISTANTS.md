# Using AI Assistants

## Purpose
This document defines guidelines and best practices for using AI assistants within the CassandraGargoyle team development workflow.

## Approved AI Tools

### Primary AI Assistant
- **Claude Code**: Preferred AI assistant for the CassandraGargoyle team
  - Pre-configured with project-specific context and scripts
  - Optimized workflows and templates available
  - Integrated with project documentation and standards
  - Direct access to project files and structure

### Alternative AI Tools
Other AI assistants are not strictly prohibited but come with additional responsibilities:
- **Claude (Anthropic)**: General development and documentation tasks
- **GitHub Copilot**: Code completion and suggestions  
- **ChatGPT**: Documentation and problem-solving support
- **Other AI tools**: Require team lead approval before use

**Note**: When using alternative AI tools, users must:
- Prepare their own context and background materials
- Ensure the AI understands project-specific standards and conventions
- Manually provide project structure and documentation references
- Take full responsibility for adapting outputs to team standards

## Best Practices

### Code Development
- Use AI assistants for code scaffolding and boilerplate generation
- Always review and understand AI-generated code before committing
- Test all AI-generated code thoroughly
- Document any complex AI-assisted implementations

### Documentation
- AI can help with documentation structure and content
- Always fact-check AI-generated technical information
- Maintain consistent voice and style across team documentation
- Use AI for translation and clarity improvements

### Problem Solving
- Use AI for debugging assistance and error analysis
- Leverage AI for architecture discussions and design patterns
- Ask AI to explain complex code or concepts to team members

## Security Guidelines

### What NOT to Share with AI
- ❌ API keys, tokens, and credentials
- ❌ Production database connection strings
- ❌ Customer data or personal information
- ❌ Proprietary algorithms or business logic
- ❌ Internal server configurations

### Safe to Share
- ✅ Public code snippets and examples
- ✅ General programming questions
- ✅ Documentation templates and structures
- ✅ Error messages (without sensitive context)
- ✅ Configuration examples (sanitized)

## Code Review Considerations

### AI-Generated Code Review
- Mark commits that contain significant AI-generated content
- Pay extra attention to logic flow and edge cases
- Verify that AI code follows team coding standards
- Ensure proper error handling and validation

### Review Checklist for AI Code
- [ ] Code follows team style guidelines
- [ ] Proper error handling implemented
- [ ] Security best practices followed
- [ ] Performance considerations addressed
- [ ] Comments and documentation included
- [ ] Tests written and passing

## Team Integration

### Communication
- Mention when AI assistance was used for significant contributions
- Share useful AI prompts and techniques with the team
- Report any AI tool limitations or issues encountered

### Learning and Development
- Use AI to learn new technologies and frameworks
- Share AI-generated learning resources with team members
- Document successful AI workflows for team reuse

## Quality Standards

### Code Quality
- AI-generated code must meet the same quality standards as human-written code
- Refactor AI suggestions to match team conventions
- Add appropriate comments explaining complex AI-generated logic

### Testing Requirements
- All AI-generated code must include comprehensive tests
- Test edge cases that AI might have missed
- Validate AI assumptions through proper testing

## Workflow Integration

### TODO Management
- Use AI assistants to help organize and prioritize TODOs
- Ask AI to suggest TODO numbering and categorization
- Leverage AI for TODO completion estimates

### Project Planning
- Use AI for project breakdown and task estimation
- Get AI assistance with technical decision-making
- Leverage AI for risk assessment and mitigation planning

## Troubleshooting

### Common Issues
- AI suggestions don't follow team standards → Always adapt to team conventions
- AI-generated code has bugs → Thorough testing and review required
- AI doesn't understand project context → Provide more specific context in prompts

### Getting Help
- Consult with team lead for AI tool approval
- Share successful AI prompts in team channels
- Report persistent AI issues to project maintainers

## Examples

### Good AI Prompt
```
"Create a function that validates email addresses according to RFC 5322, 
following our team's TypeScript coding standards with proper error handling 
and JSDoc comments."
```

### Poor AI Prompt
```
"Make email validation function"
```

## Updates and Maintenance

This document should be reviewed and updated:
- When new AI tools are approved for team use
- After significant changes in AI capabilities
- Following security incidents or concerns
- Quarterly during team retrospectives

---

**Note**: These guidelines apply to all CassandraGargoyle team members and must be followed when using AI assistance in any project work.

*Created: 2025-08-23*
*Last updated: 2025-08-23*