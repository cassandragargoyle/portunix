package edge

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	
	"strings"
	
	"golang.org/x/crypto/curve25519"
)

// WireGuardKeyPair represents a WireGuard key pair
type WireGuardKeyPair struct {
	PrivateKey string
	PublicKey  string
}

// GenerateWireGuardKeyPair generates a new WireGuard key pair
func GenerateWireGuardKeyPair() (*WireGuardKeyPair, error) {
	// Try to use wg tool if available
	if _, err := exec.LookPath("wg"); err == nil {
		return generateKeyPairWithWG()
	}
	
	// Fallback to pure Go implementation
	return generateKeyPairPureGo()
}

// generateKeyPairWithWG uses the wg tool to generate keys
func generateKeyPairWithWG() (*WireGuardKeyPair, error) {
	// Generate private key
	privateKeyCmd := exec.Command("wg", "genkey")
	privateKeyBytes, err := privateKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey := string(privateKeyBytes[:len(privateKeyBytes)-1]) // Remove newline
	
	// Generate public key from private key
	publicKeyCmd := exec.Command("wg", "pubkey")
	publicKeyCmd.Stdin = strings.NewReader(privateKey)
	publicKeyBytes, err := publicKeyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKey := string(publicKeyBytes[:len(publicKeyBytes)-1]) // Remove newline
	
	return &WireGuardKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// generateKeyPairPureGo generates keys using pure Go implementation
func generateKeyPairPureGo() (*WireGuardKeyPair, error) {
	// Generate 32 bytes of random data for private key
	var privateKeyBytes [32]byte
	if _, err := rand.Read(privateKeyBytes[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random private key: %w", err)
	}
	
	// Apply WireGuard key clamping
	privateKeyBytes[0] &= 248
	privateKeyBytes[31] &= 127
	privateKeyBytes[31] |= 64
	
	// Generate public key using curve25519
	var publicKeyBytes [32]byte
	curve25519.ScalarBaseMult(&publicKeyBytes, &privateKeyBytes)
	
	// Encode keys to base64
	privateKey := base64.StdEncoding.EncodeToString(privateKeyBytes[:])
	publicKey := base64.StdEncoding.EncodeToString(publicKeyBytes[:])
	
	return &WireGuardKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// GenerateWireGuardConfig generates WireGuard configuration with keys
func (m *Manager) GenerateWireGuardConfig(config *Config, configPath string) error {
	// Generate server key pair if not exists
	serverPrivateKeyFile := filepath.Join(configPath, "wireguard", "server_private.key")
	serverPublicKeyFile := filepath.Join(configPath, "wireguard", "server_public.key")
	
	var serverKeyPair *WireGuardKeyPair
	
	// Check if keys already exist
	if _, err := os.Stat(serverPrivateKeyFile); os.IsNotExist(err) {
		// Generate new key pair
		serverKeyPair, err = GenerateWireGuardKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate server key pair: %w", err)
		}
		
		// Save keys
		if err := os.WriteFile(serverPrivateKeyFile, []byte(serverKeyPair.PrivateKey), 0600); err != nil {
			return fmt.Errorf("failed to save server private key: %w", err)
		}
		if err := os.WriteFile(serverPublicKeyFile, []byte(serverKeyPair.PublicKey), 0644); err != nil {
			return fmt.Errorf("failed to save server public key: %w", err)
		}
		
		fmt.Println("✅ Generated new WireGuard server key pair")
	} else {
		// Load existing keys
		privateKeyBytes, err := os.ReadFile(serverPrivateKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read server private key: %w", err)
		}
		publicKeyBytes, err := os.ReadFile(serverPublicKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read server public key: %w", err)
		}
		
		serverKeyPair = &WireGuardKeyPair{
			PrivateKey: string(privateKeyBytes),
			PublicKey:  string(publicKeyBytes),
		}
		
		fmt.Println("✅ Using existing WireGuard server key pair")
	}
	
	// Generate configuration file with actual keys
	serverConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/24
ListenPort = %d
SaveConfig = false

PostUp = iptables -A FORWARD -i %%i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %%i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

`, serverKeyPair.PrivateKey, config.Edge.VPN.ServerIP, config.Edge.VPN.Port)

	// Add clients
	for _, client := range config.Edge.VPN.Clients {
		serverConfig += fmt.Sprintf(`[Peer]
# %s
PublicKey = %s
AllowedIPs = %s/32
PersistentKeepalive = %d

`, client.Name, client.PublicKey, client.IP, client.PersistentKeepalive)
	}
	
	// Save configuration
	configFile := filepath.Join(configPath, "wireguard", "wg0.conf")
	if err := os.WriteFile(configFile, []byte(serverConfig), 0600); err != nil {
		return fmt.Errorf("failed to save WireGuard configuration: %w", err)
	}
	
	fmt.Printf("📝 WireGuard configuration saved to %s\n", configFile)
	fmt.Printf("🔑 Server Public Key: %s\n", serverKeyPair.PublicKey)
	fmt.Println("\nShare this public key with clients for their configuration")
	
	return nil
}

// GenerateClientConfig generates client-side WireGuard configuration
func (m *Manager) GenerateClientConfig(clientName, serverPublicKey, serverEndpoint string) (*WireGuardKeyPair, string, error) {
	// Generate client key pair
	clientKeyPair, err := GenerateWireGuardKeyPair()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate client key pair: %w", err)
	}
	
	// Generate client configuration
	clientConfig := fmt.Sprintf(`[Interface]
# Client: %s
PrivateKey = %s
Address = 10.10.10.2/24
DNS = 8.8.8.8, 1.1.1.1

[Peer]
# VPS Server
PublicKey = %s
Endpoint = %s
AllowedIPs = 10.10.10.0/24, 192.168.0.0/16
PersistentKeepalive = 25
`, clientName, clientKeyPair.PrivateKey, serverPublicKey, serverEndpoint)
	
	return clientKeyPair, clientConfig, nil
}