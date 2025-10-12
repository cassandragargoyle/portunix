# Portunix Playbook Creation - AI Assistant Reference

> **Audience**: AI Assistants, automation tools, integration systems
> **Category**: Infrastructure as Code
> **Version**: v1.7.0+

## Machine-Readable Specification

### Playbook Structure Schema

```json
{
  "playbook_schema": {
    "required_fields": {
      "name": "string",
      "description": "string",
      "version": "string",
      "tasks": "array"
    },
    "optional_fields": {
      "metadata": "object",
      "variables": "object",
      "environments": "object",
      "prerequisites": "array"
    }
  }
}
```

### File Format Specification

```yaml
# .ptxbook file format
name: "string"                    # Required: Descriptive playbook name
description: "string"             # Required: What this playbook does
version: "string"                 # Required: Semantic version (e.g., "1.0")

metadata:                         # Optional: Playbook metadata
  author: "string"
  created: "YYYY-MM-DD"
  environment: "development|staging|production"
  tags: ["array", "of", "strings"]

variables:                        # Optional: Variable definitions
  variable_name: "value"

environments:                     # Optional: Environment-specific overrides
  environment_name:
    variable_overrides: "object"

prerequisites:                    # Optional: System requirements
  - name: "string"
    version: "string"

tasks:                           # Required: Ansible task array
  - name: "string"               # Required: Task description
    module_name:                 # Required: Ansible module
      parameter: "value"
```

## Command Interface

### Validation Commands
```json
{
  "validation": {
    "command": "portunix playbook validate [file]",
    "flags": {
      "--verbose": "detailed validation output",
      "--dry-run": "validate without execution"
    },
    "exit_codes": {
      "0": "validation_success",
      "1": "syntax_error",
      "2": "task_validation_error",
      "3": "dependency_error"
    }
  }
}
```

### Execution Commands
```json
{
  "execution": {
    "command": "portunix playbook run [file]",
    "flags": {
      "--env": "local|container|virt",
      "--target": "execution_target",
      "--image": "container_image",
      "--dry-run": "preview_changes",
      "--verbose": "detailed_output"
    },
    "environments": {
      "local": "direct_host_execution",
      "container": "isolated_container_execution",
      "virt": "virtual_machine_execution"
    }
  }
}
```

## Ansible Module Reference

### Package Management
```yaml
# System packages
- name: "Install package"
  package:
    name: "package_name"
    state: "present|absent|latest"

# Python packages
- name: "Install Python package"
  pip:
    name: ["package1", "package2"]
    state: "present|absent|latest"

# Node.js packages
- name: "Install npm package"
  npm:
    name: "package_name"
    global: true|false
```

### File Operations
```yaml
# Create directories
- name: "Create directory"
  file:
    path: "/path/to/directory"
    state: "directory|absent"
    mode: "0755"
    owner: "username"
    group: "groupname"

# Copy files
- name: "Copy file"
  copy:
    src: "source_file"
    dest: "/destination/path"
    mode: "0644"
    backup: true|false

# Template processing
- name: "Process template"
  template:
    src: "template.j2"
    dest: "/output/path"
    variables: "{{ variables }}"
```

### Service Management
```yaml
# Systemd services
- name: "Manage service"
  systemd:
    name: "service_name"
    state: "started|stopped|restarted|reloaded"
    enabled: true|false
    daemon_reload: true|false

# Generic service module
- name: "Manage service"
  service:
    name: "service_name"
    state: "started|stopped|restarted"
    enabled: true|false
```

### Command Execution
```yaml
# Execute commands
- name: "Run command"
  command: "command_to_run"
  args:
    chdir: "/working/directory"
    creates: "/file/that/indicates/completion"
    removes: "/file/that/indicates/need/to/run"

# Execute shell commands
- name: "Run shell command"
  shell: "complex | shell | command"
  args:
    executable: "/bin/bash"
```

## Variable System

### Variable Definition
```yaml
variables:
  # String variables
  app_name: "my-application"
  app_version: "1.0.0"

  # Numeric variables
  app_port: 3000
  worker_count: 4

  # Boolean variables
  debug_mode: true
  ssl_enabled: false

  # Array variables
  packages: ["git", "curl", "vim"]

  # Object variables
  database:
    host: "localhost"
    port: 5432
    name: "app_db"
```

### Variable Usage Patterns
```yaml
tasks:
  # Simple substitution
  - name: "Install {{ app_name }}"
    package:
      name: "{{ app_name }}"

  # Conditional usage
  - name: "Enable debug logging"
    lineinfile:
      path: "/etc/app/config"
      line: "DEBUG={{ debug_mode }}"
    when: debug_mode

  # Loop usage
  - name: "Install packages"
    package:
      name: "{{ item }}"
    loop: "{{ packages }}"
```

## Environment Management

### Environment-Specific Configuration
```json
{
  "environments": {
    "development": {
      "purpose": "local_development_testing",
      "characteristics": "debug_enabled,verbose_logging,test_data",
      "typical_variables": {
        "debug": true,
        "log_level": "debug",
        "use_test_data": true
      }
    },
    "staging": {
      "purpose": "pre_production_testing",
      "characteristics": "production_like,limited_resources,test_workflows",
      "typical_variables": {
        "debug": false,
        "log_level": "info",
        "performance_monitoring": true
      }
    },
    "production": {
      "purpose": "live_system_deployment",
      "characteristics": "high_performance,security_hardened,monitoring_enabled",
      "typical_variables": {
        "debug": false,
        "log_level": "warning",
        "security_mode": "strict"
      }
    }
  }
}
```

### Environment Override Patterns
```yaml
# Base configuration
variables:
  debug_mode: false
  log_level: "info"
  cache_ttl: 3600

# Environment-specific overrides
environments:
  development:
    debug_mode: true
    log_level: "debug"
    cache_ttl: 60

  production:
    log_level: "error"
    cache_ttl: 7200
    performance_monitoring: true
```

## Error Handling Patterns

### Task-Level Error Handling
```yaml
# Ignore errors and continue
- name: "Optional task"
  command: "might_fail_command"
  ignore_errors: true

# Register result for conditional logic
- name: "Test connectivity"
  uri:
    url: "http://example.com/health"
  register: health_check
  ignore_errors: true

- name: "Handle connection failure"
  debug:
    msg: "Service unavailable, using fallback"
  when: health_check.failed

# Conditional execution
- name: "Install if not present"
  package:
    name: "nginx"
    state: "present"
  when: ansible_facts['packages']['nginx'] is not defined
```

### Validation Patterns
```yaml
# Prerequisite checking
- name: "Verify system requirements"
  assert:
    that:
      - ansible_distribution_version is version('18.04', '>=')
      - ansible_memtotal_mb >= 2048
    fail_msg: "System does not meet minimum requirements"

# File existence validation
- name: "Check configuration file exists"
  stat:
    path: "/etc/app/config.yml"
  register: config_file

- name: "Create default configuration"
  copy:
    content: "default_config_content"
    dest: "/etc/app/config.yml"
  when: not config_file.stat.exists
```

## Common Automation Patterns

### Development Environment Setup
```json
{
  "pattern": "dev_environment_setup",
  "tasks_sequence": [
    "install_base_packages",
    "configure_development_tools",
    "setup_project_structure",
    "configure_services",
    "validate_installation"
  ],
  "typical_modules": ["package", "file", "template", "service", "command"]
}
```

### Application Deployment
```json
{
  "pattern": "application_deployment",
  "tasks_sequence": [
    "backup_current_version",
    "download_new_version",
    "stop_application_services",
    "update_application_files",
    "update_configuration",
    "start_application_services",
    "validate_deployment"
  ],
  "rollback_strategy": "restore_backup_on_validation_failure"
}
```

### Configuration Management
```json
{
  "pattern": "configuration_management",
  "tasks_sequence": [
    "backup_current_config",
    "generate_new_config_from_templates",
    "validate_configuration_syntax",
    "apply_new_configuration",
    "restart_affected_services",
    "verify_service_health"
  ],
  "safety_measures": "dry_run_validation,backup_restoration,service_health_checks"
}
```

## Integration Guidelines

### AI Assistant Workflow
```json
{
  "playbook_creation_workflow": {
    "step_1": {
      "action": "analyze_user_requirements",
      "output": "structured_requirements_object"
    },
    "step_2": {
      "action": "generate_playbook_structure",
      "template": "base_playbook_template"
    },
    "step_3": {
      "action": "populate_tasks_based_on_requirements",
      "reference": "module_reference_database"
    },
    "step_4": {
      "action": "add_error_handling_and_validation",
      "patterns": "standard_error_handling_patterns"
    },
    "step_5": {
      "action": "validate_generated_playbook",
      "command": "portunix playbook validate --dry-run"
    }
  }
}
```

### Template Generation
```yaml
# Basic template structure
name: "{{ playbook_name }}"
description: "{{ playbook_description }}"
version: "1.0"

metadata:
  author: "{{ author_name }}"
  created: "{{ current_date }}"
  environment: "{{ target_environment }}"

variables:
  # Variables based on requirements analysis
  {{ variable_definitions }}

tasks:
  # Tasks generated from requirements
  {{ generated_tasks }}
```

### Validation Integration
```json
{
  "validation_workflow": {
    "syntax_validation": "yaml_parser_check",
    "structure_validation": "schema_compliance_check",
    "task_validation": "ansible_module_parameter_check",
    "dependency_validation": "prerequisite_availability_check",
    "security_validation": "sensitive_data_exposure_check"
  }
}
```

## Best Practices for Automation

### Playbook Generation Guidelines
1. **Always start with requirements analysis**
2. **Use standard module patterns from reference**
3. **Include appropriate error handling for each task**
4. **Add validation steps for critical operations**
5. **Use environment-specific configurations when needed**
6. **Include rollback procedures for destructive operations**

### Variable Management
```json
{
  "variable_best_practices": {
    "naming": "snake_case_with_descriptive_names",
    "organization": "group_by_function_or_component",
    "defaults": "always_provide_sensible_defaults",
    "documentation": "include_inline_comments_for_complex_variables",
    "validation": "use_assert_module_for_critical_variables"
  }
}
```

### Task Organization
```json
{
  "task_organization": {
    "logical_grouping": "group_related_tasks_together",
    "naming_convention": "descriptive_action_oriented_names",
    "idempotency": "ensure_tasks_can_run_multiple_times_safely",
    "error_handling": "include_error_handling_for_each_critical_task",
    "documentation": "add_meaningful_descriptions_for_complex_tasks"
  }
}
```

## Output Formats

### JSON Schema for Generated Playbooks
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name", "description", "version", "tasks"],
  "properties": {
    "name": {"type": "string", "minLength": 1},
    "description": {"type": "string", "minLength": 1},
    "version": {"type": "string", "pattern": "^[0-9]+\\.[0-9]+"},
    "metadata": {
      "type": "object",
      "properties": {
        "author": {"type": "string"},
        "created": {"type": "string", "format": "date"},
        "environment": {"enum": ["development", "staging", "production"]}
      }
    },
    "variables": {"type": "object"},
    "environments": {"type": "object"},
    "prerequisites": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {"type": "string"},
          "version": {"type": "string"}
        }
      }
    },
    "tasks": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {"type": "string", "minLength": 1}
        }
      }
    }
  }
}
```

### Execution Response Format
```json
{
  "execution_response": {
    "success": {
      "status": "completed",
      "exit_code": 0,
      "tasks_executed": "number",
      "execution_time": "duration_seconds",
      "summary": "human_readable_summary"
    },
    "failure": {
      "status": "failed",
      "exit_code": "non_zero",
      "failed_task": "task_name",
      "error_message": "detailed_error_description",
      "suggestions": ["array_of_remediation_suggestions"]
    }
  }
}
```

## Common Use Cases

### Web Application Deployment
```yaml
name: "Web Application Deployment"
description: "Deploy web application with database and reverse proxy"
version: "1.0"

variables:
  app_name: "webapp"
  app_version: "{{ target_version | default('latest') }}"
  domain_name: "{{ target_domain }}"

tasks:
  - name: "Create application user"
    user:
      name: "{{ app_name }}"
      system: true
      shell: "/bin/false"

  - name: "Install application dependencies"
    package:
      name: ["nginx", "postgresql", "nodejs"]
      state: "present"

  - name: "Deploy application files"
    unarchive:
      src: "{{ app_name }}-{{ app_version }}.tar.gz"
      dest: "/opt/{{ app_name }}"
      owner: "{{ app_name }}"

  - name: "Configure database"
    postgresql_db:
      name: "{{ app_name }}_db"
      state: "present"

  - name: "Configure reverse proxy"
    template:
      src: "nginx.conf.j2"
      dest: "/etc/nginx/sites-available/{{ app_name }}"

  - name: "Enable and start services"
    service:
      name: "{{ item }}"
      state: "started"
      enabled: true
    loop: ["postgresql", "nginx", "{{ app_name }}"]
```

### Development Environment Setup
```yaml
name: "Development Environment Setup"
description: "Complete development environment with tools and configurations"
version: "1.0"

variables:
  developer_user: "{{ ansible_user }}"
  project_dir: "/home/{{ developer_user }}/projects"

tasks:
  - name: "Install development tools"
    package:
      name: ["git", "vim", "curl", "docker.io", "nodejs", "python3"]
      state: "present"

  - name: "Create project directory structure"
    file:
      path: "{{ project_dir }}/{{ item }}"
      state: "directory"
      owner: "{{ developer_user }}"
    loop: ["frontend", "backend", "scripts", "docs"]

  - name: "Configure Git settings"
    git_config:
      name: "{{ item.key }}"
      value: "{{ item.value }}"
      scope: "global"
    loop:
      - {key: "user.name", value: "{{ git_user_name }}"}
      - {key: "user.email", value: "{{ git_user_email }}"}
    when: git_user_name is defined

  - name: "Install VS Code extensions"
    command: "code --install-extension {{ item }}"
    loop: ["ms-python.python", "ms-vscode.vscode-typescript-next"]
    become_user: "{{ developer_user }}"
    ignore_errors: true
```

## Security Considerations

### Sensitive Data Handling
```yaml
# Use Ansible Vault for sensitive variables
variables:
  database_password: "{{ vault_database_password }}"
  api_key: "{{ vault_api_key }}"

# Avoid logging sensitive data
- name: "Configure database connection"
  template:
    src: "database.conf.j2"
    dest: "/etc/app/database.conf"
    mode: "0600"
  no_log: true
```

### Access Control
```yaml
# Principle of least privilege
- name: "Create service user"
  user:
    name: "appuser"
    system: true
    shell: "/bin/false"
    home: "/var/lib/app"
    create_home: false

# Secure file permissions
- name: "Set secure permissions"
  file:
    path: "/etc/app/secrets"
    mode: "0600"
    owner: "appuser"
    group: "appuser"
```

---

**Manual Level**: AI Assistant
**Target Audience**: AI assistants, automation tools, integration systems
**Last Updated**: 2025-09-24