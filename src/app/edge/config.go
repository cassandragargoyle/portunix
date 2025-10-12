package edge

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// Config represents the edge infrastructure configuration
type Config struct {
	Edge EdgeConfig `yaml:"edge"`
}

// EdgeConfig contains all edge infrastructure settings
type EdgeConfig struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Server      ServerConfig     `yaml:"server"`
	Domains     []DomainConfig   `yaml:"domains"`
	VPN         VPNConfig        `yaml:"vpn"`
	Security    SecurityConfig   `yaml:"security"`
	Containers  ContainerConfig  `yaml:"containers"`
	Monitoring  MonitoringConfig `yaml:"monitoring,omitempty"`
	Backup      BackupConfig     `yaml:"backup,omitempty"`
}

// ServerConfig contains VPS server settings
type ServerConfig struct {
	PublicIP   string `yaml:"public_ip"`
	SSHPort    int    `yaml:"ssh_port"`
	AdminEmail string `yaml:"admin_email"`
}

// DomainConfig contains domain and routing settings
type DomainConfig struct {
	Name     string         `yaml:"name"`
	Type     string         `yaml:"type"` // primary, additional
	Upstream UpstreamConfig `yaml:"upstream"`
	Static   *StaticConfig  `yaml:"static,omitempty"`
	TLS      TLSConfig      `yaml:"tls"`
	Paths    []PathConfig   `yaml:"paths,omitempty"`
}

// UpstreamConfig contains upstream service settings
type UpstreamConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// StaticConfig contains static file serving settings
type StaticConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// TLSConfig contains TLS/SSL settings
type TLSConfig struct {
	Email    string `yaml:"email"`
	Provider string `yaml:"provider"` // letsencrypt, custom
	KeyType  string `yaml:"key_type,omitempty"`
}

// PathConfig contains path-based routing settings
type PathConfig struct {
	Path         string `yaml:"path"`
	UpstreamPort int    `yaml:"upstream_port"`
	HealthCheck  string `yaml:"health_check,omitempty"`
}

// VPNConfig contains VPN settings
type VPNConfig struct {
	Type     string      `yaml:"type"` // wireguard, tailscale
	Network  string      `yaml:"network"`
	ServerIP string      `yaml:"server_ip"`
	Port     int         `yaml:"port"`
	Clients  []VPNClient `yaml:"clients"`
}

// VPNClient represents a VPN client configuration
type VPNClient struct {
	Name                string `yaml:"name"`
	IP                  string `yaml:"ip"`
	PublicKey           string `yaml:"public_key"`
	AllowedIPs          string `yaml:"allowed_ips"`
	PersistentKeepalive int    `yaml:"persistent_keepalive,omitempty"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	Firewall FirewallConfig `yaml:"firewall"`
	Fail2ban Fail2banConfig `yaml:"fail2ban"`
	SSL      SSLConfig      `yaml:"ssl,omitempty"`
}

// FirewallConfig contains firewall settings
type FirewallConfig struct {
	Enabled    bool              `yaml:"enabled"`
	SSHPort    int               `yaml:"ssh_port"`
	AllowedIPs []AllowedIPConfig `yaml:"allowed_ips,omitempty"`
}

// AllowedIPConfig represents an allowed IP configuration
type AllowedIPConfig struct {
	IP          string `yaml:"ip"`
	Description string `yaml:"description"`
}

// Fail2banConfig contains fail2ban settings
type Fail2banConfig struct {
	Enabled  bool `yaml:"enabled"`
	BanTime  int  `yaml:"ban_time"` // seconds
	MaxRetry int  `yaml:"max_retry"`
	FindTime int  `yaml:"find_time"` // seconds
}

// SSLConfig contains SSL/TLS settings
type SSLConfig struct {
	Provider string `yaml:"provider"`
	Email    string `yaml:"email"`
	KeyType  string `yaml:"key_type"`
}

// ContainerConfig contains container runtime settings
type ContainerConfig struct {
	Runtime    string                   `yaml:"runtime"` // podman, docker
	Network    string                   `yaml:"network"`
	AutoUpdate bool                     `yaml:"auto_update"`
	Services   map[string]ServiceConfig `yaml:"services,omitempty"`
}

// ServiceConfig contains individual service settings
type ServiceConfig struct {
	Image        string            `yaml:"image"`
	Ports        []string          `yaml:"ports,omitempty"`
	Volumes      []string          `yaml:"volumes,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty"`
	Capabilities []string          `yaml:"capabilities,omitempty"`
	HealthCheck  bool              `yaml:"health_check,omitempty"`
	NetworkMode  string            `yaml:"network_mode,omitempty"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	UptimeChecks bool   `yaml:"uptime_checks"`
	LogRetention string `yaml:"log_retention"`
	AlertEmail   string `yaml:"alert_email"`
}

// BackupConfig contains backup settings
type BackupConfig struct {
	Enabled      bool                      `yaml:"enabled"`
	Schedule     string                    `yaml:"schedule"` // cron format
	Retention    string                    `yaml:"retention"`
	Destinations []BackupDestinationConfig `yaml:"destinations"`
}

// BackupDestinationConfig represents a backup destination
type BackupDestinationConfig struct {
	Type      string            `yaml:"type"` // local, s3, ftp
	Path      string            `yaml:"path,omitempty"`
	Bucket    string            `yaml:"bucket,omitempty"`
	Region    string            `yaml:"region,omitempty"`
	AccessKey string            `yaml:"access_key,omitempty"`
	Options   map[string]string `yaml:"options,omitempty"`
}

// LoadConfig loads edge configuration from YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	setDefaults(&config)

	return &config, nil
}

// SaveConfig saves edge configuration to YAML file
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	if config.Edge.Server.SSHPort == 0 {
		config.Edge.Server.SSHPort = 22
	}

	if config.Edge.VPN.Port == 0 {
		config.Edge.VPN.Port = 51820
	}

	if config.Edge.VPN.Type == "" {
		config.Edge.VPN.Type = "wireguard"
	}

	if config.Edge.VPN.Network == "" {
		config.Edge.VPN.Network = "10.10.10.0/24"
	}

	if config.Edge.VPN.ServerIP == "" {
		config.Edge.VPN.ServerIP = "10.10.10.1"
	}

	if config.Edge.Containers.Runtime == "" {
		config.Edge.Containers.Runtime = "podman"
	}

	if config.Edge.Containers.Network == "" {
		config.Edge.Containers.Network = "edge-network"
	}

	if config.Edge.Security.Firewall.SSHPort == 0 {
		config.Edge.Security.Firewall.SSHPort = config.Edge.Server.SSHPort
	}

	if config.Edge.Security.Fail2ban.BanTime == 0 {
		config.Edge.Security.Fail2ban.BanTime = 600
	}

	if config.Edge.Security.Fail2ban.MaxRetry == 0 {
		config.Edge.Security.Fail2ban.MaxRetry = 3
	}

	if config.Edge.Security.Fail2ban.FindTime == 0 {
		config.Edge.Security.Fail2ban.FindTime = 600
	}

	// Set default TLS settings for domains
	for i := range config.Edge.Domains {
		domain := &config.Edge.Domains[i]
		if domain.TLS.Provider == "" {
			domain.TLS.Provider = "letsencrypt"
		}
		if domain.TLS.Email == "" {
			domain.TLS.Email = config.Edge.Server.AdminEmail
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Edge.Server.PublicIP == "" {
		return fmt.Errorf("server public_ip is required")
	}

	if c.Edge.Server.AdminEmail == "" {
		return fmt.Errorf("server admin_email is required")
	}

	if len(c.Edge.Domains) == 0 {
		return fmt.Errorf("at least one domain is required")
	}

	for _, domain := range c.Edge.Domains {
		if domain.Name == "" {
			return fmt.Errorf("domain name is required")
		}
		if domain.Upstream.Host == "" {
			return fmt.Errorf("domain upstream host is required for %s", domain.Name)
		}
		if domain.Upstream.Port == 0 {
			return fmt.Errorf("domain upstream port is required for %s", domain.Name)
		}
	}

	return nil
}
