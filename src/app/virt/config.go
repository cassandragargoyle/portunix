package virt

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config represents virtualization configuration
type Config struct {
	VirtualizationBackend string                 `yaml:"virtualization_backend"` // auto, qemu, virtualbox
	VirtDefaults          VirtDefaults           `yaml:"virt_defaults"`
	Templates             map[string]*VMTemplate `yaml:"templates,omitempty"`
}

// VirtDefaults contains default VM settings
type VirtDefaults struct {
	RAM   string `yaml:"ram"`
	CPUs  int    `yaml:"cpus"`
	Disk  string `yaml:"disk"`
	OSType string `yaml:"os_type"`
}

// VMTemplate represents a VM template
type VMTemplate struct {
	Name              string            `yaml:"name"`
	Description       string            `yaml:"description"`
	ISO               string            `yaml:"iso"`
	OSVariant         string            `yaml:"os_variant"`
	MinRAM            string            `yaml:"min_ram"`
	RecommendedRAM    string            `yaml:"recommended_ram"`
	MinDisk           string            `yaml:"min_disk"`
	RecommendedDisk   string            `yaml:"recommended_disk"`
	Features          []string          `yaml:"features"`
	PostInstall       []string          `yaml:"post_install"`
	Drivers           []string          `yaml:"drivers,omitempty"`
	RequiredFeatures  map[string]string `yaml:"required_features,omitempty"`
}

// LoadConfig loads virtualization configuration
func LoadConfig() (*Config, error) {
	config := &Config{
		VirtualizationBackend: "auto",
		VirtDefaults: VirtDefaults{
			RAM:   "4G",
			CPUs:  2,
			Disk:  "40G",
			OSType: "generic",
		},
	}

	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		if err := saveConfig(config, configPath); err != nil {
			return config, nil // Return defaults if can't save
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, nil // Return defaults on error
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return config, nil // Return defaults on error
	}

	return config, nil
}

// SaveConfig saves configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()
	return saveConfig(c, configPath)
}

// saveConfig saves configuration to specified path
func saveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the configuration file path
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".portunix/config.yaml"
	}
	return filepath.Join(homeDir, ".portunix", "config.yaml")
}

// GetPreferredBackend returns the preferred backend based on configuration and platform
func (c *Config) GetPreferredBackend() string {
	if c.VirtualizationBackend != "auto" {
		return c.VirtualizationBackend
	}

	// Auto-selection based on platform and availability
	// Prefer platform-native backend, but fallback to available one
	switch runtime.GOOS {
	case "linux":
		// Prefer QEMU/KVM on Linux, but check VirtualBox if QEMU unavailable
		return "qemu"
	case "windows":
		return "virtualbox"
	case "darwin":
		return "virtualbox" // TODO: Add support for macOS virtualization
	default:
		return "qemu"
	}
}

// LoadTemplates loads VM templates
func LoadTemplates() (map[string]*VMTemplate, error) {
	templatesPath := getTemplatesPath()

	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		// Create default templates
		templates := getDefaultTemplates()
		if err := saveTemplates(templates, templatesPath); err != nil {
			return templates, nil // Return defaults if can't save
		}
		return templates, nil
	}

	data, err := os.ReadFile(templatesPath)
	if err != nil {
		return getDefaultTemplates(), nil
	}

	var templatesFile struct {
		Templates map[string]*VMTemplate `yaml:"templates"`
	}

	if err := yaml.Unmarshal(data, &templatesFile); err != nil {
		return getDefaultTemplates(), nil
	}

	return templatesFile.Templates, nil
}

// getTemplatesPath returns the templates file path
func getTemplatesPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".portunix/vm-templates.yaml"
	}
	return filepath.Join(homeDir, ".portunix", "vm-templates.yaml")
}

// saveTemplates saves templates to file
func saveTemplates(templates map[string]*VMTemplate, templatesPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(templatesPath), 0755); err != nil {
		return err
	}

	templatesFile := struct {
		Templates map[string]*VMTemplate `yaml:"templates"`
	}{
		Templates: templates,
	}

	data, err := yaml.Marshal(templatesFile)
	if err != nil {
		return err
	}

	return os.WriteFile(templatesPath, data, 0644)
}

// getDefaultTemplates returns default VM templates
func getDefaultTemplates() map[string]*VMTemplate {
	return map[string]*VMTemplate{
		"ubuntu-24.04": {
			Name:            "Ubuntu 24.04 LTS",
			Description:     "Ubuntu 24.04 LTS Desktop",
			ISO:             "ubuntu-24.04-desktop-amd64.iso",
			OSVariant:       "ubuntu24.04",
			MinRAM:          "2G",
			RecommendedRAM:  "4G",
			MinDisk:         "25G",
			RecommendedDisk: "40G",
			Features:        []string{"uefi", "virtio"},
			PostInstall:     []string{"enable-ssh", "install-guest-tools"},
		},
		"ubuntu-22.04": {
			Name:            "Ubuntu 22.04 LTS",
			Description:     "Ubuntu 22.04 LTS Desktop",
			ISO:             "ubuntu-22.04-desktop-amd64.iso",
			OSVariant:       "ubuntu22.04",
			MinRAM:          "2G",
			RecommendedRAM:  "4G",
			MinDisk:         "25G",
			RecommendedDisk: "40G",
			Features:        []string{"uefi", "virtio"},
			PostInstall:     []string{"enable-ssh", "install-guest-tools"},
		},
		"ubuntu-22.04-server": {
			Name:            "Ubuntu 22.04 Server",
			Description:     "Ubuntu 22.04 LTS Server",
			ISO:             "ubuntu-22.04.3-server-amd64.iso",
			OSVariant:       "ubuntu22.04",
			MinRAM:          "1G",
			RecommendedRAM:  "2G",
			MinDisk:         "20G",
			RecommendedDisk: "30G",
			Features:        []string{"uefi", "virtio"},
			PostInstall:     []string{"enable-ssh"},
		},
		"debian-12": {
			Name:            "Debian 12",
			Description:     "Debian 12 (Bookworm)",
			ISO:             "debian-12.4.0-amd64-netinst.iso",
			OSVariant:       "debian12",
			MinRAM:          "1G",
			RecommendedRAM:  "2G",
			MinDisk:         "20G",
			RecommendedDisk: "30G",
			Features:        []string{"uefi", "virtio"},
			PostInstall:     []string{"enable-ssh", "install-guest-tools"},
		},
		"windows11": {
			Name:              "Windows 11",
			Description:       "Windows 11 with TPM and Secure Boot",
			ISO:               "Win11_24H2_English_x64v2.iso",
			OSVariant:         "win11",
			MinRAM:            "4G",
			RecommendedRAM:    "8G",
			MinDisk:           "64G",
			RecommendedDisk:   "100G",
			Features:          []string{"uefi", "tpm2.0", "secure-boot", "virtio"},
			Drivers:           []string{"virtio-win.iso"},
			RequiredFeatures: map[string]string{
				"tpm":         "2.0",
				"secure_boot": "true",
				"uefi":        "true",
			},
		},
		"windows10": {
			Name:            "Windows 10",
			Description:     "Windows 10 Professional",
			ISO:             "Win10_22H2_English_x64.iso",
			OSVariant:       "win10",
			MinRAM:          "2G",
			RecommendedRAM:  "4G",
			MinDisk:         "32G",
			RecommendedDisk: "60G",
			Features:        []string{"uefi", "virtio"},
			Drivers:         []string{"virtio-win.iso"},
		},
	}
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*VMTemplate, error) {
	templates, err := LoadTemplates()
	if err != nil {
		return nil, err
	}

	template, exists := templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}

// ListTemplates returns all available templates
func ListTemplates() (map[string]*VMTemplate, error) {
	return LoadTemplates()
}