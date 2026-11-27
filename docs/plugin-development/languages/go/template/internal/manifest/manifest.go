// Package manifest reads plugin.json, the single source of the plugin
// information and commands reported to Portunix over gRPC
package manifest

import (
	"encoding/json"
	"fmt"
	"os"

	pb "go-plugin-template/proto"
)

// Manifest is the part of plugin.json the plugin reports to Portunix
type Manifest struct {
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Description  string    `json:"description"`
	Author       string    `json:"author"`
	License      string    `json:"license"`
	Commands     []Command `json:"commands"`
	Dependencies struct {
		OSSupport []string `json:"os_support"`
	} `json:"dependencies"`
	AIIntegration struct {
		MCPTools []struct {
			Name string `json:"name"`
		} `json:"mcp_tools"`
	} `json:"ai_integration"`
	Permissions Permissions `json:"permissions"`
}

// Command is a command declared in plugin.json
type Command struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Subcommands []string    `json:"subcommands"`
	Parameters  []Parameter `json:"parameters"`
	Examples    []string    `json:"examples"`
}

// Parameter is a command parameter declared in plugin.json
type Parameter struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Description  string `json:"description"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"default_value"`
}

// Permissions are the permissions declared in plugin.json
type Permissions struct {
	Filesystem []string `json:"filesystem"`
	Network    []string `json:"network"`
	Database   []string `json:"database"`
	System     []string `json:"system"`
	Level      string   `json:"level"`
}

// Load reads and parses plugin.json
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path of the plugin's own manifest
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	if m.Name == "" {
		return nil, fmt.Errorf("manifest %s has no plugin name", path)
	}
	return &m, nil
}

// PluginInfo maps the manifest onto the PluginInfo message
func (m *Manifest) PluginInfo() *pb.PluginInfo {
	level := m.Permissions.Level
	if level == "" {
		level = "limited"
	}
	return &pb.PluginInfo{
		Name:         m.Name,
		Version:      m.Version,
		Description:  m.Description,
		Author:       m.Author,
		License:      m.License,
		SupportedOs:  m.Dependencies.OSSupport,
		Commands:     m.PluginCommands(),
		Capabilities: m.Capabilities(),
		RequiredPermissions: &pb.PluginPermissions{
			Filesystem: m.Permissions.Filesystem,
			Network:    m.Permissions.Network,
			Database:   m.Permissions.Database,
			System:     m.Permissions.System,
			Level:      level,
		},
	}
}

// PluginCommands maps the manifest commands onto PluginCommand messages
func (m *Manifest) PluginCommands() []*pb.PluginCommand {
	commands := make([]*pb.PluginCommand, 0, len(m.Commands))
	for _, c := range m.Commands {
		params := make([]*pb.PluginParameter, 0, len(c.Parameters))
		for _, p := range c.Parameters {
			paramType := p.Type
			if paramType == "" {
				paramType = "string"
			}
			params = append(params, &pb.PluginParameter{
				Name:         p.Name,
				Type:         paramType,
				Description:  p.Description,
				Required:     p.Required,
				DefaultValue: p.DefaultValue,
			})
		}
		commands = append(commands, &pb.PluginCommand{
			Name:        c.Name,
			Description: c.Description,
			Subcommands: c.Subcommands,
			Parameters:  params,
			Examples:    c.Examples,
		})
	}
	return commands
}

// Capabilities derives the plugin capabilities from the declared permissions
func (m *Manifest) Capabilities() *pb.PluginCapabilities {
	tools := make([]string, 0, len(m.AIIntegration.MCPTools))
	for _, t := range m.AIIntegration.MCPTools {
		tools = append(tools, t.Name)
	}
	return &pb.PluginCapabilities{
		FilesystemAccess: len(m.Permissions.Filesystem) > 0,
		NetworkAccess:    len(m.Permissions.Network) > 0,
		DatabaseAccess:   len(m.Permissions.Database) > 0,
		ContainerAccess:  false,
		SystemCommands:   len(m.Permissions.System) > 0,
		McpTools:         tools,
	}
}
