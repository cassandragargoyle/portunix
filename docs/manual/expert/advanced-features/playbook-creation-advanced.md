# Advanced Portunix Playbook Creation

> **Audience**: Expert
> **Category**: Infrastructure as Code
> **Version**: v1.7.0+

## Overview

This guide covers advanced playbook creation techniques, enterprise features, and complex deployment scenarios using Portunix's Infrastructure as Code capabilities.

## Advanced Playbook Architecture

### Enterprise Playbook Structure

```yaml
# enterprise-deployment.ptxbook
name: "Enterprise Application Deployment"
description: "Production-grade deployment with full enterprise features"
version: "2.1.0"

metadata:
  author: "DevOps Team"
  created: "2025-09-24"
  updated: "2025-09-24"
  environment: "production"
  classification: "confidential"
  compliance:
    - "SOX"
    - "GDPR"
  tags:
    - "web-application"
    - "microservices"
    - "high-availability"

# Enterprise features configuration
enterprise:
  secrets:
    encryption: "aes-256-gcm"
    key_rotation: true
    store: "vault"
  audit:
    enabled: true
    level: "detailed"
    retention: "7y"
  rbac:
    enabled: true
    required_roles:
      - "deployment-admin"
      - "security-reviewer"

# Environment-specific configurations
environments:
  development:
    replicas: 1
    resources:
      cpu: "100m"
      memory: "128Mi"
    debug: true
    log_level: "debug"

  staging:
    replicas: 2
    resources:
      cpu: "200m"
      memory: "256Mi"
    debug: false
    log_level: "info"

  production:
    replicas: 5
    resources:
      cpu: "500m"
      memory: "512Mi"
    debug: false
    log_level: "warning"
    monitoring: true
    alerting: true

# Global variables
variables:
  app_name: "enterprise-web-app"
  app_version: "{{ lookup('env', 'APP_VERSION') | default('latest') }}"
  deployment_timestamp: "{{ ansible_date_time.epoch }}"

# Encrypted secrets (managed by secrets engine)
secrets:
  database_password: "{{ vault_database_password }}"
  api_key: "{{ vault_api_key }}"
  ssl_certificate: "{{ vault_ssl_cert }}"
  ssl_private_key: "{{ vault_ssl_key }}"

# Prerequisites and dependencies
prerequisites:
  - name: "kubernetes"
    version: ">=1.24"
    required: true
  - name: "helm"
    version: ">=3.8"
    required: true
  - name: "docker"
    version: ">=20.0"
    required: false

# Rollback configuration
rollback:
  enabled: true
  strategy: "blue-green"
  health_checks:
    - endpoint: "/health"
      timeout: 30
      retries: 5
  auto_rollback:
    enabled: true
    failure_threshold: 10

# Task execution with advanced features
tasks:
  - name: "Validate deployment prerequisites"
    block:
      - name: "Check Kubernetes cluster health"
        k8s_info:
          api_version: v1
          kind: Node
        register: cluster_nodes

      - name: "Verify cluster capacity"
        assert:
          that:
            - cluster_nodes.resources | length >= 3
          fail_msg: "Insufficient cluster nodes for production deployment"

  - name: "Deploy infrastructure components"
    block:
      - name: "Create namespace"
        k8s:
          name: "{{ app_name }}-{{ deployment_environment }}"
          api_version: v1
          kind: Namespace
          state: present

      - name: "Deploy database"
        helm:
          name: "{{ app_name }}-db"
          chart_ref: "postgresql"
          release_namespace: "{{ app_name }}-{{ deployment_environment }}"
          values:
            auth:
              postgresPassword: "{{ secrets.database_password }}"
            primary:
              persistence:
                size: "{{ database_storage_size }}"

  - name: "Deploy application"
    block:
      - name: "Build and push container image"
        docker_image:
          name: "{{ container_registry }}/{{ app_name }}"
          tag: "{{ app_version }}"
          build:
            path: "{{ playbook_dir }}/src"
            dockerfile: "{{ playbook_dir }}/src/Dockerfile"
          push: true
          force_tag: true

      - name: "Deploy application via Helm"
        helm:
          name: "{{ app_name }}"
          chart_ref: "{{ playbook_dir }}/charts/{{ app_name }}"
          release_namespace: "{{ app_name }}-{{ deployment_environment }}"
          values:
            image:
              repository: "{{ container_registry }}/{{ app_name }}"
              tag: "{{ app_version }}"
            replicaCount: "{{ environments[deployment_environment].replicas }}"
            resources: "{{ environments[deployment_environment].resources }}"
            secrets:
              apiKey: "{{ secrets.api_key }}"

  - name: "Post-deployment validation"
    block:
      - name: "Wait for deployment to be ready"
        k8s_info:
          api_version: apps/v1
          kind: Deployment
          name: "{{ app_name }}"
          namespace: "{{ app_name }}-{{ deployment_environment }}"
          wait: true
          wait_condition:
            type: Available
            status: "True"
          wait_timeout: 300

      - name: "Run health checks"
        uri:
          url: "https://{{ app_name }}-{{ deployment_environment }}.{{ cluster_domain }}/health"
          method: GET
          status_code: 200
        retries: 10
        delay: 30

      - name: "Verify database connectivity"
        postgresql_query:
          db: "{{ app_name }}"
          query: "SELECT 1"
        delegate_to: "{{ groups['database'][0] }}"

# Handlers for event-driven actions
handlers:
  - name: "Restart monitoring"
    systemd:
      name: "prometheus"
      state: "restarted"

  - name: "Clear application cache"
    uri:
      url: "https://{{ app_name }}-{{ deployment_environment }}.{{ cluster_domain }}/admin/cache/clear"
      method: POST
      headers:
        Authorization: "Bearer {{ secrets.admin_token }}"

  - name: "Send deployment notification"
    slack:
      token: "{{ secrets.slack_token }}"
      msg: "Deployment of {{ app_name }} v{{ app_version }} to {{ deployment_environment }} completed successfully"
      channel: "#deployments"
```

## Secrets Management Integration

### Vault Integration

```yaml
# vault-integration.ptxbook
secrets:
  # Using HashiCorp Vault
  database_config:
    path: "kv/data/production/database"
    key: "connection_string"

  # Using Azure Key Vault
  ssl_certificate:
    provider: "azure"
    vault: "production-keyvault"
    secret: "ssl-certificate"

  # Using AWS Secrets Manager
  api_credentials:
    provider: "aws"
    region: "us-east-1"
    secret_id: "production/api-credentials"

tasks:
  - name: "Retrieve secrets from vault"
    set_fact:
      vault_secrets: "{{ lookup('vault', secrets.database_config.path) }}"
    no_log: true

  - name: "Configure application with secrets"
    template:
      src: "app-config.j2"
      dest: "/etc/{{ app_name }}/config.yml"
      mode: "0600"
    vars:
      db_connection: "{{ vault_secrets.connection_string }}"
    notify: "restart application"
```

### Encrypted Variables

```yaml
# Using Portunix built-in encryption
encrypted_variables:
  # AES-256-GCM encrypted variables
  database_password: !vault |
    $ANSIBLE_VAULT;1.2;AES256;dev
    66386439653062336464626331386663373734373031653365636636323732336566383930...

  api_key: !vault |
    $ANSIBLE_VAULT;1.2;AES256;prod
    34663366343265333034346564363966373734373031653365636636323732336566383930...

tasks:
  - name: "Use encrypted variables"
    lineinfile:
      path: "/etc/app/config"
      line: "DATABASE_PASSWORD={{ database_password }}"
    no_log: true
```

## Role-Based Access Control (RBAC)

### Access Control Configuration

```yaml
# rbac-deployment.ptxbook
rbac:
  enabled: true
  access_control:
    # Define who can execute this playbook
    allowed_users:
      - "ops-team"
      - "deployment-admins"

    # Define required roles
    required_roles:
      - role: "deployment-manager"
        actions: ["deploy", "rollback"]
      - role: "security-reviewer"
        actions: ["deploy"]
        environments: ["staging", "production"]

    # Environment-specific restrictions
    environment_access:
      development:
        users: ["developers", "ops-team"]
        roles: ["developer", "deployment-manager"]
      staging:
        users: ["ops-team", "qa-team"]
        roles: ["deployment-manager", "security-reviewer"]
      production:
        users: ["ops-team"]
        roles: ["deployment-manager", "security-reviewer", "change-approver"]

    # Time-based access controls
    maintenance_windows:
      production:
        allowed_times:
          - start: "02:00"
            end: "06:00"
            timezone: "UTC"
            days: ["sunday"]
        emergency_override: true

tasks:
  - name: "Validate user permissions"
    assert:
      that:
        - ansible_user in rbac.access_control.allowed_users
      fail_msg: "User {{ ansible_user }} not authorized for deployment"

  - name: "Check maintenance window"
    assert:
      that:
        - deployment_environment != "production" or maintenance_window_active
      fail_msg: "Production deployments only allowed during maintenance windows"
    when: not emergency_deployment | default(false)
```

## CI/CD Pipeline Integration

### GitHub Actions Integration

```yaml
# .github/workflows/deploy.yml
name: Deploy with Portunix
on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Portunix
        run: |
          curl -L https://github.com/cassandragargoyle/Portunix/releases/latest/download/portunix_linux_amd64.tar.gz | tar xz
          sudo mv portunix /usr/local/bin/

      - name: Validate playbook
        run: portunix playbook validate deployment.ptxbook

      - name: Deploy to staging
        run: |
          portunix playbook run deployment.ptxbook \
            --env container \
            --var deployment_environment=staging \
            --var app_version=${{ github.sha }}
        env:
          VAULT_TOKEN: ${{ secrets.VAULT_TOKEN }}

      - name: Deploy to production
        if: startsWith(github.ref, 'refs/tags/v')
        run: |
          portunix playbook run deployment.ptxbook \
            --env production \
            --var deployment_environment=production \
            --var app_version=${{ github.ref_name }}
        env:
          VAULT_TOKEN: ${{ secrets.VAULT_TOKEN }}
          PRODUCTION_DEPLOY_KEY: ${{ secrets.PRODUCTION_DEPLOY_KEY }}
```

### GitLab CI Integration

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - deploy-staging
  - deploy-production

variables:
  PORTUNIX_VERSION: "v1.7.2"

before_script:
  - curl -L https://github.com/cassandragargoyle/Portunix/releases/download/${PORTUNIX_VERSION}/portunix_linux_amd64.tar.gz | tar xz
  - sudo mv portunix /usr/local/bin/

validate-playbook:
  stage: validate
  script:
    - portunix playbook validate infrastructure.ptxbook
    - portunix playbook validate --env container infrastructure.ptxbook

deploy-staging:
  stage: deploy-staging
  script:
    - portunix playbook run infrastructure.ptxbook --env staging
  environment:
    name: staging
    url: https://staging.example.com
  only:
    - main

deploy-production:
  stage: deploy-production
  script:
    - portunix playbook run infrastructure.ptxbook --env production
  environment:
    name: production
    url: https://production.example.com
  when: manual
  only:
    - tags
```

## Advanced Task Patterns

### Conditional Execution

```yaml
tasks:
  - name: "Detect operating system"
    setup:

  - name: "Install packages on Ubuntu"
    apt:
      name: "{{ packages }}"
      state: present
    when: ansible_distribution == "Ubuntu"

  - name: "Install packages on CentOS"
    yum:
      name: "{{ packages }}"
      state: present
    when: ansible_distribution == "CentOS"

  - name: "Configure firewall on production"
    ufw:
      rule: allow
      port: "{{ item }}"
    loop:
      - "80"
      - "443"
      - "22"
    when:
      - deployment_environment == "production"
      - configure_firewall | default(true)
```

### Error Handling and Recovery

```yaml
tasks:
  - name: "Deploy application with rollback"
    block:
      - name: "Create backup of current deployment"
        command: kubectl create backup {{ app_name }}-{{ ansible_date_time.epoch }}

      - name: "Deploy new version"
        k8s:
          definition:
            apiVersion: apps/v1
            kind: Deployment
            metadata:
              name: "{{ app_name }}"
            spec:
              template:
                spec:
                  containers:
                    - name: "{{ app_name }}"
                      image: "{{ app_name }}:{{ new_version }}"

      - name: "Wait for deployment readiness"
        k8s_info:
          api_version: apps/v1
          kind: Deployment
          name: "{{ app_name }}"
          wait: true
          wait_condition:
            type: Available
            status: "True"
          wait_timeout: 300

      - name: "Run health checks"
        uri:
          url: "http://{{ app_name }}/health"
          status_code: 200
        retries: 5
        delay: 10

    rescue:
      - name: "Rollback on failure"
        command: kubectl rollout undo deployment/{{ app_name }}

      - name: "Wait for rollback completion"
        k8s_info:
          api_version: apps/v1
          kind: Deployment
          name: "{{ app_name }}"
          wait: true
          wait_condition:
            type: Available
            status: "True"
          wait_timeout: 300

      - name: "Send failure notification"
        slack:
          token: "{{ slack_token }}"
          msg: "Deployment of {{ app_name }} failed and was rolled back"
          channel: "#alerts"

    always:
      - name: "Cleanup temporary resources"
        file:
          path: "/tmp/deployment-{{ ansible_date_time.epoch }}"
          state: absent
```

### Parallel Execution

```yaml
tasks:
  - name: "Deploy microservices in parallel"
    include_tasks: deploy-service.yml
    vars:
      service_name: "{{ item.name }}"
      service_config: "{{ item.config }}"
    loop:
      - name: "user-service"
        config: "{{ user_service_config }}"
      - name: "order-service"
        config: "{{ order_service_config }}"
      - name: "payment-service"
        config: "{{ payment_service_config }}"
    async: 300
    poll: 0
    register: service_deployments

  - name: "Wait for all services to deploy"
    async_status:
      jid: "{{ item.ansible_job_id }}"
    loop: "{{ service_deployments.results }}"
    register: service_status
    until: service_status.finished
    retries: 30
    delay: 10
```

## Performance Optimization

### Efficient Resource Management

```yaml
# Optimize for large-scale deployments
strategy:
  execution:
    parallelism: 10
    batch_size: 5
    failure_threshold: 20

  gathering:
    # Disable fact gathering for better performance
    gather_facts: false

  caching:
    # Enable task result caching
    cache_results: true
    cache_ttl: 3600

tasks:
  - name: "Gather minimal facts when needed"
    setup:
      gather_subset:
        - "!all"
        - "network"
        - "hardware"
    when: facts_needed | default(false)

  - name: "Use efficient loops"
    package:
      name: "{{ packages }}"
      state: present
    # More efficient than looping individual packages

  - name: "Delegate heavy tasks to appropriate hosts"
    command: "{{ heavy_computation_command }}"
    delegate_to: "{{ groups['compute_nodes'][0] }}"
    run_once: true
```

### Resource Monitoring

```yaml
tasks:
  - name: "Monitor resource usage during deployment"
    block:
      - name: "Check available memory"
        command: free -m
        register: memory_check

      - name: "Check available disk space"
        command: df -h /
        register: disk_check

      - name: "Verify resource requirements"
        assert:
          that:
            - memory_check.stdout | regex_search('Available:\\s+(\\d+)', '\\1') | int > 1000
            - disk_check.stdout | regex_search('(\\d+)% /', '\\1') | int < 80
          fail_msg: "Insufficient resources for deployment"
```

## Testing and Validation

### Comprehensive Testing Framework

```yaml
# test-suite.ptxbook
name: "Deployment Test Suite"
description: "Comprehensive testing for production deployments"

tests:
  pre_deployment:
    - name: "Infrastructure readiness"
      tasks:
        - name: "Check cluster health"
          k8s_info:
            api_version: v1
            kind: Node
          register: nodes

        - name: "Validate node resources"
          assert:
            that:
              - nodes.resources | selectattr('status.conditions', 'search', 'Ready.*True') | list | length >= 3
            fail_msg: "Insufficient healthy nodes"

  post_deployment:
    - name: "Application health"
      tasks:
        - name: "HTTP health check"
          uri:
            url: "https://{{ app_domain }}/health"
            status_code: 200

        - name: "Database connectivity"
          postgresql_query:
            db: "{{ app_name }}"
            query: "SELECT version()"

        - name: "Performance baseline"
          uri:
            url: "https://{{ app_domain }}/api/test"
            status_code: 200
          register: response_time

        - name: "Validate response time"
          assert:
            that:
              - response_time.elapsed < 2.0
            fail_msg: "Response time too slow: {{ response_time.elapsed }}s"

  integration:
    - name: "End-to-end workflows"
      tasks:
        - name: "User registration flow"
          uri:
            url: "https://{{ app_domain }}/api/register"
            method: POST
            body_format: json
            body:
              username: "test_user_{{ ansible_date_time.epoch }}"
              email: "test@example.com"
            status_code: 201

        - name: "Authentication flow"
          uri:
            url: "https://{{ app_domain }}/api/login"
            method: POST
            body_format: json
            body:
              username: "test_user_{{ ansible_date_time.epoch }}"
              password: "test_password"
            status_code: 200
```

## Monitoring and Observability

### Deployment Monitoring

```yaml
tasks:
  - name: "Setup deployment monitoring"
    block:
      - name: "Create monitoring dashboard"
        grafana_dashboard:
          url: "{{ grafana_url }}"
          api_key: "{{ grafana_api_key }}"
          dashboard: "{{ lookup('file', 'dashboards/deployment.json') }}"

      - name: "Configure alerting rules"
        prometheus_rule:
          rules:
            - alert: "DeploymentFailed"
              expr: "increase(deployment_failures_total[5m]) > 0"
              for: "0m"
              labels:
                severity: "critical"
              annotations:
                summary: "Deployment failed for {{ app_name }}"

      - name: "Setup log aggregation"
        elastic_index_template:
          name: "{{ app_name }}-logs"
          body:
            index_patterns: ["{{ app_name }}-*"]
            settings:
              number_of_shards: 1
              number_of_replicas: 1
```

## Compliance and Governance

### Audit Trail Configuration

```yaml
audit:
  enabled: true
  settings:
    # Detailed audit logging
    level: "detailed"

    # Audit event categories
    events:
      - "playbook_execution"
      - "task_execution"
      - "variable_access"
      - "secret_access"
      - "permission_check"

    # Retention and storage
    retention:
      period: "7y"
      archive: true
      encryption: true

    # Compliance reporting
    compliance:
      frameworks:
        - "SOX"
        - "GDPR"
        - "HIPAA"
      reporting:
        frequency: "monthly"
        format: "json"

tasks:
  - name: "Log compliance checkpoint"
    audit_event:
      event_type: "compliance_checkpoint"
      description: "Pre-deployment compliance validation"
      data:
        playbook: "{{ playbook_name }}"
        environment: "{{ deployment_environment }}"
        user: "{{ ansible_user }}"
        timestamp: "{{ ansible_date_time.iso8601 }}"
```

## Troubleshooting Advanced Deployments

### Debug Mode Configuration

```yaml
# Enable comprehensive debugging
debug:
  enabled: true
  settings:
    task_timing: true
    variable_dump: true
    connection_debug: true

tasks:
  - name: "Debug deployment state"
    debug:
      msg: |
        Deployment Debug Information:
        - Environment: {{ deployment_environment }}
        - Version: {{ app_version }}
        - User: {{ ansible_user }}
        - Timestamp: {{ ansible_date_time.iso8601 }}
        - Host: {{ inventory_hostname }}
        - Variables: {{ vars | to_nice_json }}
    when: debug.enabled
```

### Advanced Error Diagnosis

```yaml
tasks:
  - name: "Comprehensive error diagnosis"
    block:
      - name: "Collect system information"
        setup:
          gather_subset: "all"
        register: system_facts

      - name: "Check system resources"
        shell: |
          echo "=== CPU Usage ==="
          top -bn1 | grep "Cpu(s)"
          echo "=== Memory Usage ==="
          free -h
          echo "=== Disk Usage ==="
          df -h
          echo "=== Network Connections ==="
          netstat -tulpn
        register: resource_info

      - name: "Check application logs"
        command: journalctl -u {{ app_name }} --lines=100 --no-pager
        register: app_logs
        ignore_errors: true

      - name: "Generate diagnostic report"
        copy:
          content: |
            Diagnostic Report Generated: {{ ansible_date_time.iso8601 }}
            Playbook: {{ playbook_name }}
            Environment: {{ deployment_environment }}

            System Facts:
            {{ system_facts | to_nice_json }}

            Resource Information:
            {{ resource_info.stdout }}

            Application Logs:
            {{ app_logs.stdout }}
          dest: "/tmp/diagnostic-{{ ansible_date_time.epoch }}.txt"

    rescue:
      - name: "Minimal error collection"
        shell: |
          echo "Error occurred at: {{ ansible_date_time.iso8601 }}"
          echo "Failed task: {{ ansible_failed_task.name | default('unknown') }}"
          echo "Error message: {{ ansible_failed_result.msg | default('no message') }}"
        register: minimal_error_info

      - name: "Save error information"
        copy:
          content: "{{ minimal_error_info.stdout }}"
          dest: "/tmp/error-{{ ansible_date_time.epoch }}.txt"
```

## Best Practices for Enterprise Deployments

### 1. Security Hardening
- Always use encrypted secrets management
- Implement least-privilege access controls
- Enable comprehensive audit logging
- Validate all inputs and configurations

### 2. Reliability Patterns
- Implement health checks and rollback mechanisms
- Use blue-green or canary deployment strategies
- Monitor resource utilization during deployments
- Plan for disaster recovery scenarios

### 3. Performance Optimization
- Minimize fact gathering overhead
- Use efficient task patterns and loops
- Implement caching where appropriate
- Parallelize independent operations

### 4. Compliance and Governance
- Maintain detailed audit trails
- Implement approval workflows for production
- Document all security and compliance measures
- Regular security and compliance reviews

---

**Manual Level**: Expert
**Target Audience**: Advanced users, DevOps engineers, system administrators
**Last Updated**: 2025-09-24