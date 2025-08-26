# Issue #15: VPS Edge/Bastion Infrastructure Management

## Summary
Implement comprehensive VPS edge/bastion infrastructure management system that allows users to deploy minimal edge instances with reverse proxy capabilities and secure tunneling to home lab environments.

## Problem Description
Users need a solution to:
- Deploy small VPS instances as edge/bastion hosts
- Handle multiple domains with automatic TLS certificates
- Create secure tunnels (WireGuard/Tailscale/Cloudflare Tunnel) to home labs
- Manage reverse proxy configurations (HTTP/HTTPS and TCP services)
- Minimize costs while maintaining security and performance

## Proposed Solution

### Core Implementation (portunix)
1. **Edge Infrastructure Templates**
   - Pre-configured templates for common edge scenarios
   - Support for Caddy, HAProxy, WireGuard configurations
   - Podman-based container orchestration for edge services

2. **Certificate Management**
   - Let's Encrypt integration with DNS-01 challenge support
   - Automatic certificate renewal
   - Multi-domain and wildcard certificate support

3. **Network Tunneling**
   - WireGuard configuration and key management
   - Site-to-site tunnel establishment
   - Traffic routing and firewall rules

4. **Security Hardening**
   - Automated firewall setup (ufw/nftables)
   - Fail2ban integration
   - SSH hardening templates

### Architecture Components
```
Internet → VPS Edge (Caddy/HAProxy + WireGuard) → Secure Tunnel → Home Lab
```

## Technical Requirements

### Core Features
- [ ] Extend `portunix install` with edge package profiles
- [ ] Create edge infrastructure templates in `assets/templates/edge/`
- [ ] Integrate with existing `app/podman/` container management
- [ ] Add WireGuard configuration generation and management
- [ ] Implement certificate management system
- [ ] Create firewall and security hardening automation

### Package Dependencies
Add to `assets/install-packages.json`:
- Caddy (reverse proxy with automatic HTTPS)
- WireGuard (VPN tunneling)
- Podman (container runtime)
- ufw/fail2ban (security)

### MCP Integration
Extend existing `app/mcp/` with tools:
- `create_edge_infrastructure`
- `configure_domain_proxy`
- `setup_secure_tunnel`
- `manage_certificates`

## Implementation Plan

### Phase 1: Core Infrastructure
1. **Template System**
   - Create base templates for edge scenarios
   - Caddy configuration templates
   - WireGuard site-to-site templates
   - Security hardening templates

2. **Package Integration**
   - Add edge packages to install system
   - Create installation profiles (minimal-edge, full-edge)
   - Container image management

### Phase 2: Automation & Management
1. **Certificate Automation**
   - Let's Encrypt DNS-01 integration
   - Certificate renewal automation
   - Multi-provider DNS support

2. **Tunnel Management**
   - WireGuard key generation and distribution
   - Tunnel configuration and testing
   - Network routing setup

### Phase 3: Advanced Features
1. **Monitoring Integration**
   - Health checks and uptime monitoring
   - Log aggregation
   - Alert system integration

2. **Backup and Recovery**
   - Configuration backup automation
   - Disaster recovery procedures
   - Quick deployment from templates

## Related Components
- `app/install/` - Package installation system
- `app/podman/` - Container management
- `app/mcp/` - AI assistant integration
- `assets/install-packages.json` - Package definitions

## Success Criteria
- [ ] Users can deploy VPS edge infrastructure with single command
- [ ] Automatic TLS certificate management for multiple domains
- [ ] Secure tunnel establishment to home lab environments
- [ ] Template-based configuration for consistent deployments
- [ ] Integration with existing Portunix container and MCP systems

## Security Considerations
- Use rootless Podman containers where possible
- Implement fail2ban and firewall hardening by default
- Secure key management for WireGuard
- Regular security updates automation
- Principle of least privilege for all services

## Alternative Solutions Considered
1. **Tailscale Integration**: Managed VPN solution as alternative to WireGuard
2. **Cloudflare Tunnel**: Zero-trust tunnel solution for CGNAT environments
3. **Traefik**: Alternative to Caddy with more advanced routing features

## Priority: High
This feature addresses a common infrastructure need and leverages existing Portunix capabilities while providing significant value to users deploying hybrid cloud architectures.

## Labels
- enhancement
- infrastructure
- edge-computing
- containers
- networking
- security
- cross-platform