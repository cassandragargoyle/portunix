# Creating Portunix Playbooks - Basic Tutorial

> **Audience**: Basic
> **Category**: Infrastructure as Code
> **Version**: v1.7.0+

## Quick Start

Portunix playbooks (`.ptxbook` files) are simple YAML-based configuration files that automate infrastructure tasks using Ansible. Think of them as recipes for setting up and managing your development environments.

## What You'll Learn
- How to create your first playbook
- Basic playbook structure
- Running and validating playbooks
- Common patterns and examples

## Prerequisites
- Portunix v1.7.0+ installed
- Basic understanding of YAML syntax
- Target environment (local, container, or VM)

## Your First Playbook

### Step 1: Create a Simple Playbook

Create a file called `hello-world.ptxbook`:

```yaml
# hello-world.ptxbook
name: "Hello World Infrastructure"
description: "My first Portunix playbook"
version: "1.0"

metadata:
  author: "Your Name"
  created: "2025-09-24"
  environment: "development"

tasks:
  - name: "Install basic development tools"
    package:
      name: "git"
      state: "present"

  - name: "Create project directory"
    file:
      path: "/tmp/my-project"
      state: "directory"
      mode: "0755"

  - name: "Display completion message"
    debug:
      msg: "Hello World playbook completed successfully!"
```

### Step 2: Validate Your Playbook

Before running, always validate your playbook:

```bash
portunix playbook validate hello-world.ptxbook
```

**Expected output:**
```
✓ Syntax validation passed
✓ Task validation passed
✓ Dependencies checked
✓ Playbook is ready to execute
```

### Step 3: Run Your Playbook

```bash
# Run locally (default)
portunix playbook run hello-world.ptxbook

# Or run in a container for safety
portunix playbook run hello-world.ptxbook --env container
```

**Expected output:**
```
🚀 Starting playbook execution: Hello World Infrastructure
📋 Environment: local
⏱️  Estimated duration: 30 seconds

✓ Task 1/3: Install basic development tools
✓ Task 2/3: Create project directory
✓ Task 3/3: Display completion message

🎉 Playbook completed successfully in 28 seconds
```

## Playbook Structure

### Required Sections

```yaml
# Basic information (required)
name: "Descriptive playbook name"
description: "What this playbook does"
version: "1.0"

# Metadata (recommended)
metadata:
  author: "Your name"
  created: "2025-09-24"
  environment: "development|staging|production"

# Tasks (required)
tasks:
  - name: "Task description"
    # Ansible module here
```

### Optional Sections

```yaml
# Variables
variables:
  app_name: "my-application"
  app_port: 3000

# Environment-specific settings
environments:
  development:
    debug: true
  production:
    debug: false

# Prerequisites
prerequisites:
  - name: "docker"
    version: ">=20.0"
```

## Common Patterns

### Installing Software

```yaml
tasks:
  - name: "Install Node.js"
    package:
      name: "nodejs"
      state: "present"

  - name: "Install Python packages"
    pip:
      name:
        - "flask"
        - "requests"
      state: "present"
```

### File Operations

```yaml
tasks:
  - name: "Create configuration directory"
    file:
      path: "/etc/myapp"
      state: "directory"
      mode: "0755"

  - name: "Copy configuration file"
    copy:
      src: "config.yml"
      dest: "/etc/myapp/config.yml"
      mode: "0644"

  - name: "Create file from template"
    template:
      src: "app.conf.j2"
      dest: "/etc/myapp/app.conf"
```

### Service Management

```yaml
tasks:
  - name: "Start web service"
    service:
      name: "nginx"
      state: "started"
      enabled: true

  - name: "Restart application"
    systemd:
      name: "myapp"
      state: "restarted"
```

### Environment Setup

```yaml
tasks:
  - name: "Set environment variables"
    lineinfile:
      path: "~/.bashrc"
      line: "export APP_ENV=development"
      create: true

  - name: "Clone repository"
    git:
      repo: "https://github.com/user/repo.git"
      dest: "/opt/myapp"
      version: "main"
```

## Using Variables

### Define Variables

```yaml
variables:
  app_name: "my-web-app"
  app_port: 8080
  app_user: "webapp"

tasks:
  - name: "Create app user"
    user:
      name: "{{ app_user }}"
      state: "present"

  - name: "Create app directory"
    file:
      path: "/opt/{{ app_name }}"
      state: "directory"
      owner: "{{ app_user }}"
```

### Environment-Specific Variables

```yaml
variables:
  # Default values
  debug_mode: false
  log_level: "info"

environments:
  development:
    debug_mode: true
    log_level: "debug"

  production:
    debug_mode: false
    log_level: "warning"
```

## Execution Environments

### Local Execution (Default)

```bash
# Runs directly on your machine
portunix playbook run my-playbook.ptxbook
```

### Container Execution

```bash
# Runs inside an isolated container
portunix playbook run my-playbook.ptxbook --env container

# Specify container image
portunix playbook run my-playbook.ptxbook --env container --image ubuntu:22.04
```

### Virtual Machine Execution

```bash
# Runs inside a VM (requires VM setup)
portunix playbook run my-playbook.ptxbook --env virt --target my-vm
```

## Best Practices

### 1. Always Validate First
```bash
# Never skip validation
portunix playbook validate my-playbook.ptxbook
```

### 2. Use Descriptive Names
```yaml
# Good
- name: "Install Node.js for React development"

# Bad
- name: "Install stuff"
```

### 3. Test in Containers First
```bash
# Safe testing
portunix playbook run my-playbook.ptxbook --env container --dry-run
```

### 4. Use Variables for Reusability
```yaml
variables:
  node_version: "18"

tasks:
  - name: "Install Node.js {{ node_version }}"
    package:
      name: "nodejs={{ node_version }}"
```

### 5. Add Error Handling
```yaml
tasks:
  - name: "Download application"
    get_url:
      url: "https://example.com/app.tar.gz"
      dest: "/tmp/app.tar.gz"
    ignore_errors: true
    register: download_result

  - name: "Handle download failure"
    debug:
      msg: "Download failed, using local copy"
    when: download_result.failed
```

## Common Examples

### Web Development Environment

```yaml
name: "Web Development Setup"
description: "Complete environment for web development"
version: "1.0"

variables:
  project_name: "my-website"
  project_dir: "/home/{{ ansible_user }}/projects/{{ project_name }}"

tasks:
  - name: "Install development tools"
    package:
      name:
        - "git"
        - "nodejs"
        - "npm"
        - "code"
      state: "present"

  - name: "Create project directory"
    file:
      path: "{{ project_dir }}"
      state: "directory"

  - name: "Initialize Git repository"
    command: "git init"
    args:
      chdir: "{{ project_dir }}"
      creates: "{{ project_dir }}/.git"

  - name: "Create package.json"
    copy:
      content: |
        {
          "name": "{{ project_name }}",
          "version": "1.0.0",
          "scripts": {
            "start": "node server.js"
          }
        }
      dest: "{{ project_dir }}/package.json"
```

### Docker Development Environment

```yaml
name: "Docker Development Environment"
description: "Setup Docker for containerized development"
version: "1.0"

tasks:
  - name: "Install Docker"
    package:
      name: "docker.io"
      state: "present"

  - name: "Start Docker service"
    service:
      name: "docker"
      state: "started"
      enabled: true

  - name: "Add user to docker group"
    user:
      name: "{{ ansible_user }}"
      groups: "docker"
      append: true

  - name: "Create docker-compose.yml"
    copy:
      content: |
        version: '3.8'
        services:
          web:
            image: nginx
            ports:
              - "80:80"
      dest: "/home/{{ ansible_user }}/docker-compose.yml"
```

## Troubleshooting

### Common Issues

**Playbook validation fails:**
```bash
# Check YAML syntax
portunix playbook validate my-playbook.ptxbook --verbose
```

**Task execution fails:**
```bash
# Run with debug output
portunix playbook run my-playbook.ptxbook --dry-run --verbose
```

**Permission issues:**
```bash
# Use container environment for testing
portunix playbook run my-playbook.ptxbook --env container
```

### Getting Help

```bash
# Playbook help
portunix playbook --help

# Specific subcommand help
portunix playbook run --help

# Check if helper is available
portunix playbook check
```

## Next Steps

### Learn More
- [Advanced Playbook Features](../expert/advanced-features/infrastructure-as-code.md) - Expert-level features
- [Secrets Management](secrets-management.md) - Working with sensitive data
- [Multi-Environment Deployment](multi-environment-deployment.md) - Development to production workflows

### Practice Projects
1. Create a playbook to set up your preferred development environment
2. Build a playbook to deploy a simple web application
3. Create environment-specific configurations for development and staging

### Advanced Topics
- Using Ansible Galaxy roles
- Custom Ansible modules
- Integration with CI/CD pipelines
- Enterprise features (RBAC, audit logging)

---

**Manual Level**: Basic
**Target Audience**: New users learning Infrastructure as Code
**Last Updated**: 2025-09-24