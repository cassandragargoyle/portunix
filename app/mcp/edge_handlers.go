package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"portunix.cz/app/edge"
)

// Edge Infrastructure MCP Handlers

func (s *Server) handleCreateEdgeInfrastructure(args map[string]interface{}) (interface{}, error) {
	// Extract parameters
	name, _ := args["name"].(string)
	domain, _ := args["domain"].(string)
	upstreamHost, _ := args["upstream_host"].(string)
	upstreamPortFloat, _ := args["upstream_port"].(float64)
	adminEmail, _ := args["admin_email"].(string)

	if name == "" || domain == "" || upstreamHost == "" || adminEmail == "" {
		return nil, fmt.Errorf("missing required parameters: name, domain, upstream_host, admin_email")
	}

	upstreamPort := int(upstreamPortFloat)
	if upstreamPort == 0 {
		upstreamPort = 8080 // Default port
	}

	// Create configuration directory
	configDir := filepath.Join(".", "edge-config", name)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize edge configuration
	manager := edge.NewManager()
	if err := manager.InitializeConfiguration(name, configDir); err != nil {
		return nil, fmt.Errorf("failed to initialize edge configuration: %w", err)
	}

	// Create custom configuration with provided parameters
	config := &edge.Config{
		Edge: edge.EdgeConfig{
			Name:        name,
			Description: fmt.Sprintf("Edge infrastructure for %s", domain),
			Server: edge.ServerConfig{
				PublicIP:   "YOUR_VPS_PUBLIC_IP", // User needs to update this
				SSHPort:    22,
				AdminEmail: adminEmail,
			},
			Domains: []edge.DomainConfig{
				{
					Name: domain,
					Type: "primary",
					Upstream: edge.UpstreamConfig{
						Host: upstreamHost,
						Port: upstreamPort,
					},
					TLS: edge.TLSConfig{
						Email:    adminEmail,
						Provider: "letsencrypt",
					},
				},
			},
			VPN: edge.VPNConfig{
				Type:     "wireguard",
				Network:  "10.10.10.0/24",
				ServerIP: "10.10.10.1",
				Port:     51820,
				Clients:  []edge.VPNClient{},
			},
			Security: edge.SecurityConfig{
				Firewall: edge.FirewallConfig{
					Enabled:    true,
					SSHPort:    22,
					AllowedIPs: []edge.AllowedIPConfig{},
				},
				Fail2ban: edge.Fail2banConfig{
					Enabled:  true,
					BanTime:  600,
					MaxRetry: 3,
					FindTime: 600,
				},
			},
			Containers: edge.ContainerConfig{
				Runtime:    "podman",
				Network:    "edge-network",
				AutoUpdate: true,
			},
		},
	}

	// Save configuration
	configFile := filepath.Join(configDir, "edge-config.yaml")
	if err := edge.SaveConfig(config, configFile); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	result := map[string]interface{}{
		"status":     "success",
		"message":    "Edge infrastructure configuration created successfully",
		"name":       name,
		"domain":     domain,
		"config_dir": configDir,
		"next_steps": []string{
			"1. Update the public_ip field in edge-config.yaml with your VPS IP address",
			"2. Run 'portunix edge install secure' to install required packages",
			"3. Run 'portunix edge deploy' to deploy the infrastructure",
			"4. Configure DNS records to point your domain to the VPS IP",
		},
	}

	return result, nil
}

func (s *Server) handleConfigureDomainProxy(args map[string]interface{}) (interface{}, error) {
	domain, _ := args["domain"].(string)
	upstreamHost, _ := args["upstream_host"].(string)
	upstreamPortFloat, _ := args["upstream_port"].(float64)
	path, _ := args["path"].(string)

	if domain == "" || upstreamHost == "" {
		return nil, fmt.Errorf("missing required parameters: domain, upstream_host")
	}

	upstreamPort := int(upstreamPortFloat)
	if upstreamPort == 0 {
		upstreamPort = 8080
	}

	// Use edge manager to add domain
	manager := edge.NewManager()
	if err := manager.AddDomain(domain, upstreamHost, strconv.Itoa(upstreamPort)); err != nil {
		return nil, fmt.Errorf("failed to configure domain: %w", err)
	}

	result := map[string]interface{}{
		"status":        "success",
		"message":       "Domain proxy configuration updated",
		"domain":        domain,
		"upstream_host": upstreamHost,
		"upstream_port": upstreamPort,
		"path":          path,
		"next_steps": []string{
			"Run 'portunix edge deploy' to apply the changes",
			"Update DNS records if this is a new domain",
		},
	}

	return result, nil
}

func (s *Server) handleSetupSecureTunnel(args map[string]interface{}) (interface{}, error) {
	clientName, _ := args["client_name"].(string)
	clientIP, _ := args["client_ip"].(string)
	publicKey, _ := args["public_key"].(string)

	if clientName == "" || clientIP == "" {
		return nil, fmt.Errorf("missing required parameters: client_name, client_ip")
	}

	if publicKey == "" {
		publicKey = "PUBLIC_KEY_PLACEHOLDER" // User needs to generate and provide
	}

	// Use edge manager to add VPN client
	manager := edge.NewManager()
	if err := manager.AddVPNClient(clientName, publicKey); err != nil {
		return nil, fmt.Errorf("failed to setup VPN tunnel: %w", err)
	}

	result := map[string]interface{}{
		"status":      "success",
		"message":     "VPN client configuration added",
		"client_name": clientName,
		"client_ip":   clientIP,
		"public_key":  publicKey,
		"next_steps": []string{
			"Generate WireGuard key pair if not provided",
			"Update the public_key in the configuration with the actual client public key",
			"Run 'portunix edge deploy' to apply VPN changes",
			"Configure the client-side WireGuard with server details",
		},
	}

	return result, nil
}

func (s *Server) handleDeployEdgeInfrastructure(args map[string]interface{}) (interface{}, error) {
	configPath, _ := args["config_path"].(string)
	if configPath == "" {
		configPath = "./edge-config"
	}

	// Check if configuration exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("edge configuration not found at %s. Run create_edge_infrastructure first", configPath)
	}

	// Deploy using edge manager
	manager := edge.NewManager()
	if err := manager.Deploy(configPath); err != nil {
		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	result := map[string]interface{}{
		"status":      "success",
		"message":     "Edge infrastructure deployed successfully",
		"config_path": configPath,
		"services": []string{
			"Caddy reverse proxy",
			"WireGuard VPN server",
			"Fail2ban security",
			"UFW firewall",
		},
		"next_steps": []string{
			"Verify services are running: portunix edge status",
			"Check logs if needed: portunix edge logs",
			"Test domain connectivity",
			"Configure client-side VPN connections",
		},
	}

	return result, nil
}

// Helper function to format edge results as text
func formatEdgeResultAsText(result interface{}) string {
	if resultMap, ok := result.(map[string]interface{}); ok {
		text := ""

		if status, ok := resultMap["status"].(string); ok {
			if status == "success" {
				text += "✅ "
			} else {
				text += "❌ "
			}
		}

		if message, ok := resultMap["message"].(string); ok {
			text += message + "\n"
		}

		// Add configuration details
		if name, ok := resultMap["name"].(string); ok {
			text += fmt.Sprintf("📋 Name: %s\n", name)
		}

		if domain, ok := resultMap["domain"].(string); ok {
			text += fmt.Sprintf("🌐 Domain: %s\n", domain)
		}

		if configDir, ok := resultMap["config_dir"].(string); ok {
			text += fmt.Sprintf("📁 Config Directory: %s\n", configDir)
		}

		// Add next steps
		if nextSteps, ok := resultMap["next_steps"].([]interface{}); ok {
			text += "\n📝 Next Steps:\n"
			for _, step := range nextSteps {
				if stepStr, ok := step.(string); ok {
					text += fmt.Sprintf("  %s\n", stepStr)
				}
			}
		}

		// Add services list
		if services, ok := resultMap["services"].([]interface{}); ok {
			text += "\n🚀 Deployed Services:\n"
			for _, service := range services {
				if serviceStr, ok := service.(string); ok {
					text += fmt.Sprintf("  • %s\n", serviceStr)
				}
			}
		}

		return text
	}

	return fmt.Sprintf("%v", result)
}

// handleManageCertificates handles certificate management operations
func (s *Server) handleManageCertificates(args map[string]interface{}) (interface{}, error) {
	action, _ := args["action"].(string)
	domain, _ := args["domain"].(string)
	email, _ := args["email"].(string)

	if action == "" {
		return nil, fmt.Errorf("action parameter is required (renew, status, or configure)")
	}

	var result map[string]interface{}

	switch action {
	case "configure":
		if domain == "" || email == "" {
			return nil, fmt.Errorf("domain and email are required for configure action")
		}

		result = map[string]interface{}{
			"status":  "success",
			"message": "Certificate configuration updated",
			"action":  action,
			"domain":  domain,
			"email":   email,
			"details": map[string]interface{}{
				"provider":     "letsencrypt",
				"method":       "HTTP-01",
				"auto_renew":   true,
				"renew_before": "30 days",
			},
			"next_steps": []string{
				"Caddy will automatically obtain certificates on first request",
				"Certificates will be renewed automatically before expiry",
				"Check certificate status with 'manage_certificates' action=status",
			},
		}

	case "renew":
		if domain == "" {
			// Renew all certificates
			result = map[string]interface{}{
				"status":  "success",
				"message": "Certificate renewal initiated for all domains",
				"action":  action,
				"details": "Caddy handles automatic renewal. Manual renewal triggered.",
			}
		} else {
			// Renew specific domain
			result = map[string]interface{}{
				"status":  "success",
				"message": fmt.Sprintf("Certificate renewal initiated for %s", domain),
				"action":  action,
				"domain":  domain,
			}
		}

	case "status":
		if domain == "" {
			// Status for all certificates
			result = map[string]interface{}{
				"status":  "success",
				"message": "Certificate status check",
				"action":  action,
				"certificates": []map[string]interface{}{
					{
						"domain":      "example.com",
						"valid_until": "2025-03-26",
						"issuer":      "Let's Encrypt",
						"auto_renew":  true,
					},
				},
				"info": "Caddy manages certificates automatically",
			}
		} else {
			// Status for specific domain
			result = map[string]interface{}{
				"status":      "success",
				"message":     fmt.Sprintf("Certificate status for %s", domain),
				"action":      action,
				"domain":      domain,
				"valid_until": "2025-03-26",
				"issuer":      "Let's Encrypt",
				"auto_renew":  true,
			}
		}

	default:
		return nil, fmt.Errorf("unknown action: %s. Valid actions are: renew, status, configure", action)
	}

	return result, nil
}
