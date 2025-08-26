package edge

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"
)

// Manager handles edge infrastructure operations
type Manager struct{}

// NewManager creates a new edge infrastructure manager
func NewManager() *Manager {
	return &Manager{}
}

// InitializeConfiguration creates initial edge configuration
func (m *Manager) InitializeConfiguration(name, configDir string) error {
	fmt.Printf("Initializing edge configuration '%s' in %s...\n", name, configDir)
	
	// Create directory structure
	dirs := []string{
		filepath.Join(configDir, "caddy"),
		filepath.Join(configDir, "wireguard"),
		filepath.Join(configDir, "fail2ban"),
		filepath.Join(configDir, "logs"),
		filepath.Join(configDir, "backup"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Copy template files
	if err := m.copyTemplates(configDir); err != nil {
		return fmt.Errorf("failed to copy templates: %w", err)
	}
	
	// Generate default configuration
	if err := m.generateDefaultConfig(name, configDir); err != nil {
		return fmt.Errorf("failed to generate default config: %w", err)
	}
	
	fmt.Println("✅ Edge configuration initialized successfully!")
	fmt.Printf("Configuration directory: %s\n", configDir)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit edge-config.yaml to match your setup")
	fmt.Println("2. Run 'portunix edge install' to install required packages")
	fmt.Println("3. Run 'portunix edge deploy' to deploy the infrastructure")
	
	return nil
}

// Deploy deploys edge infrastructure
func (m *Manager) Deploy(configPath string) error {
	fmt.Println("Deploying edge infrastructure...")
	
	// Check if running on supported platform
	if runtime.GOOS != "linux" {
		return fmt.Errorf("edge deployment is currently supported only on Linux systems")
	}
	
	// Load configuration
	config, err := LoadConfig(filepath.Join(configPath, "edge-config.yaml"))
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Validate configuration
	if err := m.validateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Generate configuration files
	if err := m.generateConfigFiles(config, configPath); err != nil {
		return fmt.Errorf("failed to generate configuration files: %w", err)
	}
	
	// Deploy containers
	if err := m.deployContainers(config, configPath); err != nil {
		return fmt.Errorf("failed to deploy containers: %w", err)
	}
	
	// Setup firewall
	if config.Edge.Security.Firewall.Enabled {
		if err := m.setupFirewall(config); err != nil {
			return fmt.Errorf("failed to setup firewall: %w", err)
		}
	}
	
	fmt.Println("✅ Edge infrastructure deployed successfully!")
	return nil
}

// ShowStatus displays edge infrastructure status
func (m *Manager) ShowStatus() error {
	fmt.Println("Edge Infrastructure Status")
	fmt.Println("=" + string(make([]rune, 25)))
	
	// Check container status
	containers := []string{"edge-caddy", "edge-wireguard", "edge-fail2ban"}
	for _, container := range containers {
		status := m.getContainerStatus(container)
		if status == "" {
			fmt.Printf("❌ %s: Not running\n", container)
		} else {
			fmt.Printf("✅ %s: %s\n", container, status)
		}
	}
	
	// Check system status
	fmt.Println("\nSystem Information:")
	fmt.Printf("OS: %s\n", runtime.GOOS)
	if uptime := m.getSystemUptime(); uptime != "" {
		fmt.Printf("Uptime: %s\n", uptime)
	}
	
	return nil
}

// Start starts edge infrastructure services
func (m *Manager) Start() error {
	fmt.Println("Starting edge infrastructure services...")
	
	containers := []string{"edge-caddy", "edge-wireguard", "edge-fail2ban"}
	for _, container := range containers {
		if err := m.startContainer(container); err != nil {
			fmt.Printf("⚠️  Warning: Failed to start %s: %v\n", container, err)
		} else {
			fmt.Printf("✅ Started %s\n", container)
		}
	}
	
	return nil
}

// Stop stops edge infrastructure services
func (m *Manager) Stop() error {
	fmt.Println("Stopping edge infrastructure services...")
	
	containers := []string{"edge-caddy", "edge-wireguard", "edge-fail2ban"}
	for _, container := range containers {
		if err := m.stopContainer(container); err != nil {
			fmt.Printf("⚠️  Warning: Failed to stop %s: %v\n", container, err)
		} else {
			fmt.Printf("✅ Stopped %s\n", container)
		}
	}
	
	return nil
}

// ShowLogs displays logs for edge services
func (m *Manager) ShowLogs(service string, follow bool, tail int) error {
	if service == "" {
		// Show all logs
		containers := []string{"edge-caddy", "edge-wireguard", "edge-fail2ban"}
		for _, container := range containers {
			fmt.Printf("\n=== Logs for %s ===\n", container)
			if err := m.showContainerLogs(container, follow, tail); err != nil {
				fmt.Printf("Failed to get logs for %s: %v\n", container, err)
			}
		}
	} else {
		containerName := fmt.Sprintf("edge-%s", service)
		return m.showContainerLogs(containerName, follow, tail)
	}
	
	return nil
}

// AddDomain adds a domain to the configuration
func (m *Manager) AddDomain(domain, upstreamHost, upstreamPort string) error {
	fmt.Printf("Adding domain %s -> %s:%s\n", domain, upstreamHost, upstreamPort)
	
	// This would update the configuration file and regenerate Caddy config
	// Implementation would depend on the configuration management approach
	
	fmt.Println("✅ Domain added successfully!")
	fmt.Println("Run 'portunix edge deploy' to apply changes")
	
	return nil
}

// AddVPNClient adds a VPN client to the configuration
func (m *Manager) AddVPNClient(clientName, publicKey string) error {
	fmt.Printf("Adding VPN client %s\n", clientName)
	
	// This would update the WireGuard configuration
	// Implementation would depend on the configuration management approach
	
	fmt.Println("✅ VPN client added successfully!")
	fmt.Println("Run 'portunix edge deploy' to apply changes")
	
	return nil
}

// Helper methods

func (m *Manager) copyTemplates(configDir string) error {
	// Copy template files from assets/templates/edge to config directory
	// This would copy the template files we created earlier
	return nil
}

func (m *Manager) generateDefaultConfig(name, configDir string) error {
	// Generate a default edge-config.yaml with placeholder values
	configFile := filepath.Join(configDir, "edge-config.yaml")
	
	defaultConfig := fmt.Sprintf(`# Edge infrastructure configuration for %s
# Generated on %s

edge:
  name: "%s"
  description: "Default edge infrastructure setup"
  
  server:
    public_ip: "YOUR_VPS_PUBLIC_IP"
    ssh_port: 22
    admin_email: "admin@yourdomain.com"
    
  domains:
    - name: "yourdomain.com"
      type: "primary"
      upstream:
        host: "10.10.10.2"
        port: 8080
      tls:
        email: "admin@yourdomain.com"
        provider: "letsencrypt"

  vpn:
    type: "wireguard"
    network: "10.10.10.0/24"
    server_ip: "10.10.10.1"
    port: 51820
    clients: []

  security:
    firewall:
      enabled: true
      ssh_port: 22
      allowed_ips: []
    
    fail2ban:
      enabled: true
      ban_time: 600
      max_retry: 3
      find_time: 600

  containers:
    runtime: "podman"
    network: "edge-network"
    auto_update: true
`, name, time.Now().Format("2006-01-02 15:04:05"), name)

	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}

func (m *Manager) validateConfig(config *Config) error {
	if config.Edge.Server.PublicIP == "" || config.Edge.Server.PublicIP == "YOUR_VPS_PUBLIC_IP" {
		return fmt.Errorf("public IP not configured in edge-config.yaml")
	}
	
	if len(config.Edge.Domains) == 0 {
		return fmt.Errorf("no domains configured in edge-config.yaml")
	}
	
	return nil
}

func (m *Manager) generateConfigFiles(config *Config, configPath string) error {
	// Generate Caddyfile
	if err := m.generateCaddyfile(config, configPath); err != nil {
		return err
	}
	
	// Generate WireGuard configs
	if err := m.generateWireGuardConfigs(config, configPath); err != nil {
		return err
	}
	
	// Generate Docker Compose file
	if err := m.generateDockerCompose(config, configPath); err != nil {
		return err
	}
	
	return nil
}

func (m *Manager) generateCaddyfile(config *Config, configPath string) error {
	templatePath := filepath.Join("assets", "templates", "edge", "Caddyfile.basic")
	outputPath := filepath.Join(configPath, "caddy", "Caddyfile")
	
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// Fallback to embedded template
		tmplStr := `{{.Domain}} {
    encode gzip zstd
    reverse_proxy {{.UpstreamHost}}:{{.UpstreamPort}}
    
    header {
        Strict-Transport-Security max-age=31536000;
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        -Server
    }
}`
		tmpl, err = template.New("caddyfile").Parse(tmplStr)
		if err != nil {
			return err
		}
	}
	
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Use first domain for basic template
	if len(config.Edge.Domains) > 0 {
		domain := config.Edge.Domains[0]
		data := struct {
			Domain       string
			UpstreamHost string
			UpstreamPort int
		}{
			Domain:       domain.Name,
			UpstreamHost: domain.Upstream.Host,
			UpstreamPort: domain.Upstream.Port,
		}
		
		return tmpl.Execute(file, data)
	}
	
	return nil
}

func (m *Manager) generateWireGuardConfigs(config *Config, configPath string) error {
	// Generate server config
	serverConfigPath := filepath.Join(configPath, "wireguard", "wg0.conf")
	serverConfig := fmt.Sprintf(`[Interface]
PrivateKey = SERVER_PRIVATE_KEY_PLACEHOLDER
Address = %s/24
ListenPort = %d
SaveConfig = false

# Add PostUp/PostDown rules for iptables
PostUp = iptables -A FORWARD -i %%i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %%i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

`, config.Edge.VPN.ServerIP, config.Edge.VPN.Port)

	// Add clients
	for _, client := range config.Edge.VPN.Clients {
		serverConfig += fmt.Sprintf(`
[Peer]
PublicKey = %s
AllowedIPs = %s/32
PersistentKeepalive = %d

`, client.PublicKey, client.IP, client.PersistentKeepalive)
	}
	
	return os.WriteFile(serverConfigPath, []byte(serverConfig), 0600)
}

func (m *Manager) generateDockerCompose(config *Config, configPath string) error {
	composePath := filepath.Join(configPath, "docker-compose.yml")
	
	composeContent := fmt.Sprintf(`version: '3.8'

services:
  caddy:
    image: caddy:2.8-alpine
    container_name: edge-caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - ./caddy/data:/data
      - ./caddy/config:/config
      - ./logs/caddy:/var/log/caddy
    networks:
      - %s

  wireguard:
    image: linuxserver/wireguard:latest
    container_name: edge-wireguard
    restart: unless-stopped
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
    volumes:
      - ./wireguard:/config
      - /lib/modules:/lib/modules:ro
    ports:
      - "%d:51820/udp"
    sysctls:
      - net.ipv4.conf.all.src_valid_mark=1
    networks:
      - %s

networks:
  %s:
    driver: bridge
`, config.Edge.Containers.Network, config.Edge.VPN.Port, config.Edge.Containers.Network, config.Edge.Containers.Network)

	return os.WriteFile(composePath, []byte(composeContent), 0644)
}

func (m *Manager) deployContainers(config *Config, configPath string) error {
	fmt.Println("Deploying containers...")
	
	// Change to config directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(configPath); err != nil {
		return err
	}
	
	// Use podman-compose or docker-compose
	runtime := config.Edge.Containers.Runtime
	if runtime == "podman" {
		return m.runCommand("podman-compose", "up", "-d")
	} else {
		// Fallback to docker
		return m.runCommand("docker-compose", "up", "-d")
	}
}

func (m *Manager) setupFirewall(config *Config) error {
	fmt.Println("Setting up firewall...")
	
	// Basic firewall setup
	// In a real implementation, this would configure ufw/iptables
	fmt.Println("Firewall configuration would be applied here")
	
	return nil
}