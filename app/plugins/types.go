package plugins

import (
	"context"
	"time"
)

// Plugin interface defines the contract for all plugins
type Plugin interface {
	// Initialize plugin with configuration
	Initialize(ctx context.Context, config PluginConfig) error
	
	// Start plugin service
	Start(ctx context.Context) error
	
	// Stop plugin service
	Stop(ctx context.Context) error
	
	// Execute a command
	Execute(ctx context.Context, request ExecuteRequest) (*ExecuteResponse, error)
	
	// Get plugin information
	GetInfo() PluginInfo
	
	// Health check
	Health(ctx context.Context) PluginHealth
	
	// Check if plugin is running
	IsRunning() bool
}

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	BinaryPath      string            `json:"binary_path"`
	Port            int               `json:"port"`
	HealthInterval  time.Duration     `json:"health_check_interval"`
	Environment     map[string]string `json:"environment"`
	Permissions     PluginPermissions `json:"permissions"`
	WorkingDir      string            `json:"working_directory"`
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name                string              `json:"name"`
	Version             string              `json:"version"`
	Description         string              `json:"description"`
	Author              string              `json:"author"`
	License             string              `json:"license"`
	SupportedOS         []string            `json:"supported_os"`
	Commands            []PluginCommand     `json:"commands"`
	Capabilities        PluginCapabilities  `json:"capabilities"`
	RequiredPermissions PluginPermissions   `json:"required_permissions"`
	Status              PluginStatus        `json:"status"`
	LastSeen            time.Time           `json:"last_seen"`
}

// PluginCommand represents a command provided by a plugin
type PluginCommand struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Subcommands  []string          `json:"subcommands"`
	Parameters   []PluginParameter `json:"parameters"`
	Examples     []string          `json:"examples"`
}

// PluginParameter represents a parameter for a plugin command
type PluginParameter struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Description  string `json:"description"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"default_value"`
}

// PluginCapabilities defines what a plugin can do
type PluginCapabilities struct {
	FilesystemAccess bool     `json:"filesystem_access"`
	NetworkAccess    bool     `json:"network_access"`
	DatabaseAccess   bool     `json:"database_access"`
	ContainerAccess  bool     `json:"container_access"`
	SystemCommands   bool     `json:"system_commands"`
	MCPTools         []string `json:"mcp_tools"`
}

// PluginPermissions defines what permissions a plugin needs
type PluginPermissions struct {
	Filesystem []string `json:"filesystem"` // read, write, execute
	Network    []string `json:"network"`    // inbound, outbound
	Database   []string `json:"database"`   // local, remote
	System     []string `json:"system"`     // commands, services
	Level      string   `json:"level"`      // limited, standard, full
}

// PluginStatus represents the current status of a plugin
type PluginStatus int

const (
	PluginStatusUnknown PluginStatus = iota
	PluginStatusStopped
	PluginStatusStarting
	PluginStatusRunning
	PluginStatusStopping
	PluginStatusFailed
)

func (s PluginStatus) String() string {
	switch s {
	case PluginStatusStopped:
		return "stopped"
	case PluginStatusStarting:
		return "starting"
	case PluginStatusRunning:
		return "running"
	case PluginStatusStopping:
		return "stopping"
	case PluginStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// PluginHealth represents the health status of a plugin
type PluginHealth struct {
	Healthy        bool              `json:"healthy"`
	Status         string            `json:"status"`
	Message        string            `json:"message"`
	UptimeSeconds  int64             `json:"uptime_seconds"`
	Metrics        map[string]string `json:"metrics"`
	LastCheckTime  time.Time         `json:"last_check_time"`
}

// ExecuteRequest represents a request to execute a command
type ExecuteRequest struct {
	Command          string            `json:"command"`
	Args             []string          `json:"args"`
	Options          map[string]string `json:"options"`
	Environment      map[string]string `json:"environment"`
	WorkingDirectory string            `json:"working_directory"`
}

// ExecuteResponse represents the response from executing a command
type ExecuteResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Output   string            `json:"output"`
	Error    string            `json:"error"`
	ExitCode int               `json:"exit_code"`
	Metadata map[string]string `json:"metadata"`
}

// PluginManifest represents the plugin.yaml manifest file
type PluginManifest struct {
	Name         string               `yaml:"name"`
	Version      string               `yaml:"version"`
	Description  string               `yaml:"description"`
	Author       string               `yaml:"author"`
	License      string               `yaml:"license"`
	Plugin       PluginBinaryConfig   `yaml:"plugin"`
	Dependencies PluginDependencies   `yaml:"dependencies"`
	AIIntegration AIIntegrationConfig  `yaml:"ai_integration"`
	Permissions  PluginPermissions    `yaml:"permissions"`
	Commands     []PluginCommand      `yaml:"commands"`
}

// PluginBinaryConfig holds binary-specific configuration
type PluginBinaryConfig struct {
	Type                string        `yaml:"type"`
	Binary              string        `yaml:"binary"`
	Port                int           `yaml:"port"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
}

// PluginDependencies holds dependency information
type PluginDependencies struct {
	PortunixMinVersion string   `yaml:"portunix_min_version"`
	OSSupport          []string `yaml:"os_support"`
}

// AIIntegrationConfig holds AI integration configuration
type AIIntegrationConfig struct {
	MCPTools []MCPTool `yaml:"mcp_tools"`
}

// MCPTool represents an MCP tool provided by the plugin
type MCPTool struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}