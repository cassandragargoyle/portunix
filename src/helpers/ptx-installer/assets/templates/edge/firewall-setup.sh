#!/bin/bash
# Edge infrastructure firewall setup script
# Template variables: {{.WireguardPort}}, {{.SSHPort}}, {{.AllowedIPs}}

set -euo pipefail

echo "🔥 Setting up firewall rules for edge infrastructure..."

# Reset UFW to defaults
sudo ufw --force reset

# Set default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (be careful not to lock yourself out)
echo "Allowing SSH on port {{.SSHPort}}..."
sudo ufw allow {{.SSHPort}}/tcp comment "SSH"

# Allow HTTP and HTTPS
echo "Allowing HTTP/HTTPS..."
sudo ufw allow 80/tcp comment "HTTP"
sudo ufw allow 443/tcp comment "HTTPS"

# Allow WireGuard
echo "Allowing WireGuard on port {{.WireguardPort}}/udp..."
sudo ufw allow {{.WireguardPort}}/udp comment "WireGuard VPN"

{{if .AllowedIPs}}
# Allow specific IP addresses
{{range .AllowedIPs}}
echo "Allowing access from {{.IP}}..."
sudo ufw allow from {{.IP}} comment "{{.Description}}"
{{end}}
{{end}}

# Enable logging
sudo ufw logging on

# Enable UFW
echo "Enabling UFW firewall..."
sudo ufw --force enable

echo "✅ Firewall setup complete!"
echo "Current UFW status:"
sudo ufw status verbose

# Install and configure fail2ban
echo "📋 Setting up fail2ban..."
sudo apt-get update
sudo apt-get install -y fail2ban

# Create fail2ban configuration
sudo tee /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
# Ban hosts for 10 minutes
bantime = 600
# A host is banned if it has generated "maxretry" during "findtime"
findtime = 600
maxretry = 3
# Email settings (optional)
# destemail = admin@yourdomain.com
# sender = fail2ban@yourdomain.com

[sshd]
enabled = true
port = {{.SSHPort}}
filter = sshd
logpath = /var/log/auth.log
maxretry = 3

[caddy-auth]
enabled = true
port = http,https
filter = caddy-auth
logpath = /var/log/caddy/*.log
maxretry = 5
EOF

# Create custom Caddy filter
sudo tee /etc/fail2ban/filter.d/caddy-auth.conf << 'EOF'
[Definition]
failregex = ^.*"remote_ip":"<HOST>".*"status":(?:401|403|404).*$
ignoreregex =
EOF

# Enable and start fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

echo "✅ Fail2ban setup complete!"
echo "Fail2ban status:"
sudo fail2ban-client status

echo "🎉 Edge infrastructure security setup completed successfully!"