# Terminology and Abbreviations

## Purpose
This document defines common terminology, abbreviations, and technical terms used across all CassandraGargoyle projects. It serves as a reference to ensure consistent communication and understanding within the development team.

## General Abbreviations

### Project Management
- **ADR** - Architecture Decision Record
- **API** - Application Programming Interface
- **CI/CD** - Continuous Integration/Continuous Deployment
- **CLI** - Command Line Interface
- **CRUD** - Create, Read, Update, Delete
- **DDD** - Domain-Driven Design
- **DevOps** - Development Operations (development and IT operations collaboration)
- **GUI** - Graphical User Interface
- **IDE** - Integrated Development Environment
- **MVP** - Minimum Viable Product
- **POC** - Proof of Concept
- **QA** - Quality Assurance
- **REST** - Representational State Transfer
- **SDK** - Software Development Kit
- **SLA** - Service Level Agreement
- **SOP** - Standard Operating Procedure
- **UI** - User Interface
- **UX** - User Experience

### Development
- **AAA** - Arrange, Act, Assert (testing pattern)
- **AOP** - Aspect-Oriented Programming
- **DI** - Dependency Injection
- **DRY** - Don't Repeat Yourself
- **IoC** - Inversion of Control
- **KISS** - Keep It Simple, Stupid
- **MVC** - Model-View-Controller
- **OOP** - Object-Oriented Programming
- **ORM** - Object-Relational Mapping
- **SOLID** - Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **TDD** - Test-Driven Development
- **YAGNI** - You Aren't Gonna Need It

### Infrastructure & DevOps
- **CDN** - Content Delivery Network
- **DNS** - Domain Name System
- **HTTP/HTTPS** - HyperText Transfer Protocol (Secure)
- **IaC** - Infrastructure as Code
- **JSON** - JavaScript Object Notation
- **K8s** - Kubernetes
- **LB** - Load Balancer
- **RBAC** - Role-Based Access Control
- **SSH** - Secure Shell
- **SSL/TLS** - Secure Sockets Layer/Transport Layer Security
- **VM** - Virtual Machine
- **YAML** - YAML Ain't Markup Language

## CassandraGargoyle Specific Terms

### Project Names
- **Bootstrap Scripts** - Centralized initialization scripts for development environment setup
- **Portunix** - Cross-platform development environment management tool
- **Dream** - [Project description to be added]

### Team Roles
- **SD** - Software Developer
- **TL** - Tech Lead
- **AR** - Architect
- **TS** - Tester
- **US** - User (end user perspective)

### Development Concepts
- **CLAUDE.md** - Project-specific instructions file for AI assistants
- **MCP** - Model Context Protocol (Claude integration)
- **Self-deployment** - Automatic copying and setup of Portunix in virtual environments
- **VM Environment** - Isolated virtual machine environment for testing and development

### Security & Access Control (Portunix)
- **RBAC Roles** - Predefined permission sets: `admin` (full access), `developer` (dev environments), `operator` (production), `auditor` (read-only)
- **Permissions** - Granular access controls: `playbook:execute`, `env:local`, `env:container`, `env:virt`, `secret:read`, `secret:write`
- **Environment Isolation** - RBAC-enforced separation between local, container, and virtualization environments
- **Access Request** - RBAC validation request containing user, permission, resource, and environment context
- **Audit Trail** - Comprehensive logging of all RBAC decisions and playbook executions for compliance

## Technical Terms by Category

### Programming Languages
- **Go** - Programming language used for Portunix core
- **PowerShell** - Cross-platform scripting language and shell
- **Bash** - Unix shell and command language
- **C++** - Systems programming language
- **Java** - Object-oriented programming language
- **Python** - High-level programming language

### Package Managers
- **apt** - Advanced Package Tool (Debian/Ubuntu)
- **chocolatey** - Package manager for Windows
- **homebrew** - Package manager for macOS
- **npm** - Node Package Manager
- **winget** - Windows Package Manager

### Containerization
- **Docker** - Container platform with daemon-based architecture
- **Podman** - Pod Manager, daemonless and rootless container engine
- **Container** - Lightweight, portable runtime environment
- **Image** - Template for creating containers
- **Registry** - Repository for container images
- **Pod** - Group of one or more containers with shared storage/network (Kubernetes concept)

### Operating Systems
- **Distro** - Linux distribution
- **LTS** - Long Term Support (Ubuntu/Debian releases)
- **WSL** - Windows Subsystem for Linux

### Testing
- **E2E** - End-to-End testing
- **Integration Test** - Testing component interactions
- **Unit Test** - Testing individual functions/methods
- **Mock** - Simulated object for testing
- **Fixture** - Test data or setup
- **Coverage** - Percentage of code tested
- **Suite** - Collection of related tests

### Version Control
- **PR** - Pull Request
- **MR** - Merge Request
- **Branch** - Independent line of development
- **Commit** - Saved change to repository
- **Tag** - Named reference to specific commit
- **HEAD** - Reference to current branch tip

### Build & Deployment
- **Artifact** - Built output (executable, library, etc.)
- **Pipeline** - Automated build/deployment process
- **Stage** - Phase in deployment pipeline
- **Release** - Packaged version for distribution
- **Rollback** - Reverting to previous version

## File Extensions & Formats

### Configuration
- **.yaml/.yml** - YAML configuration files
- **.json** - JSON configuration/data files
- **.toml** - TOML configuration files
- **.env** - Environment variables file
- **.config** - Generic configuration file

### Documentation
- **.md** - Markdown documentation
- **.rst** - reStructuredText documentation
- **.adoc** - AsciiDoc documentation

### Scripts & Code
- **.sh** - Shell script (Linux/macOS)
- **.ps1** - PowerShell script
- **.bat/.cmd** - Windows batch file
- **.go** - Go source file
- **.py** - Python source file
- **.js** - JavaScript source file

### Build & Package
- **Dockerfile** - Docker image definition
- **Makefile** - Build automation file
- **go.mod** - Go module definition
- **package.json** - Node.js package definition
- **requirements.txt** - Python dependencies

## Environment Variables

### Common Patterns
- **PATH** - Executable search path
- **HOME** - User home directory
- **USER** - Current username
- **PWD** - Present working directory
- **TMPDIR** - Temporary directory

### Project-Specific
- **PORTUNIX_HOME** - Portunix installation directory
- **PORTUNIX_CONFIG** - Configuration file path
- **DEBUG** - Enable debug mode
- **LOG_LEVEL** - Logging verbosity level

## Network & Security

### Protocols
- **gRPC** - Remote Procedure Call framework
- **WebSocket** - Full-duplex communication protocol
- **MQTT** - Message Queuing Telemetry Transport
- **TCP/UDP** - Transmission Control/User Datagram Protocol

### Security
- **JWT** - JSON Web Token
- **OAuth** - Open Authorization
- **RBAC** - Role-Based Access Control: Security model that restricts system access based on user roles and permissions. In Portunix playbook system, controls who can execute, read, write, or delete playbooks in different environments (local, container, virt)
- **SAML** - Security Assertion Markup Language
- **2FA/MFA** - Two-Factor/Multi-Factor Authentication
- **PKI** - Public Key Infrastructure

### Networking
- **CIDR** - Classless Inter-Domain Routing
- **NAT** - Network Address Translation
- **VPN** - Virtual Private Network
- **DHCP** - Dynamic Host Configuration Protocol

## Database & Storage

### Database Types
- **RDBMS** - Relational Database Management System
- **NoSQL** - Not Only SQL databases
- **ACID** - Atomicity, Consistency, Isolation, Durability
- **CAP** - Consistency, Availability, Partition tolerance

### Storage
- **BLOB** - Binary Large Object
- **S3** - Simple Storage Service (AWS)
- **CDN** - Content Delivery Network
- **RAID** - Redundant Array of Independent Disks

## Monitoring & Logging

### Metrics
- **SLI** - Service Level Indicator
- **SLO** - Service Level Objective
- **KPI** - Key Performance Indicator
- **APM** - Application Performance Monitoring

### Logging
- **ELK** - Elasticsearch, Logstash, Kibana
- **JSON** - Structured log format
- **Syslog** - Standard logging protocol

## Usage Guidelines

### Consistency Rules
1. **Use standard abbreviations** from this document
2. **Define project-specific terms** in project documentation
3. **Avoid creating new abbreviations** without team consensus
4. **Update this document** when introducing new terminology

### Documentation Standards
1. **First use rule**: Write out full term with abbreviation in parentheses
   - Example: "Application Programming Interface (API)"
2. **Consistent capitalization**: Follow established patterns
3. **Context sensitivity**: Consider your audience's technical level

### Communication
1. **Team discussions**: Use abbreviations freely within team
2. **External documentation**: Define terms for broader audience
3. **Code comments**: Use full terms for clarity
4. **Commit messages**: Abbreviations acceptable for brevity

## References and Resources

### Official Documentation
- [Go Documentation](https://golang.org/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [GitHub Documentation](https://docs.github.com/)

### Industry Standards
- [RFC Standards](https://www.ietf.org/rfc/)
- [ISO Standards](https://www.iso.org/standards.html)
- [NIST Guidelines](https://www.nist.gov/)

### CassandraGargoyle Resources
- [Bootstrap Scripts Documentation](../README.md)
- [Contributing Guidelines](README.md)
- [Code Style Guidelines](CODE-STYLE-*.md)

---

**Note**: This terminology document is a living document that should be updated as the project evolves and new terms are introduced. Team members are encouraged to suggest additions or clarifications.

*Created: 2025-08-23*
*Last updated: 2025-08-23*