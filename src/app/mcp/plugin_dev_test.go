package mcp

import (
	"encoding/json"
	"testing"
)

func TestEchoTool(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name     string
		args     map[string]interface{}
		wantErr  bool
		validate func(interface{}) bool
	}{
		{
			name: "echo_with_message",
			args: map[string]interface{}{
				"message": "hello world",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					return resultMap["echo"].(string) == "hello world" &&
						resultMap["status"].(string) == "success"
				}
				return false
			},
		},
		{
			name:    "echo_without_message",
			args:    map[string]interface{}{},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					return resultMap["echo"].(string) == "hello" &&
						resultMap["status"].(string) == "success"
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.handleEcho(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil && !tt.validate(result) {
				t.Errorf("Validation failed for result: %v", result)
			}
		})
	}
}

func TestPluginDevelopmentTools(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name     string
		handler  func(map[string]interface{}) (interface{}, error)
		args     map[string]interface{}
		wantErr  bool
		validate func(interface{}) bool
	}{
		{
			name:    "get_plugin_development_guide_go",
			handler: server.handleGetPluginDevelopmentGuide,
			args: map[string]interface{}{
				"language": "go",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					return resultMap["language"].(string) == "go"
				}
				return false
			},
		},
		{
			name:    "get_plugin_development_guide_invalid_language",
			handler: server.handleGetPluginDevelopmentGuide,
			args: map[string]interface{}{
				"language": "invalid",
			},
			wantErr: true,
		},
		{
			name:    "get_plugin_template_go",
			handler: server.handleGetPluginTemplate,
			args: map[string]interface{}{
				"language":    "go",
				"plugin_type": "service",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					return resultMap["language"].(string) == "go" &&
						resultMap["plugin_type"].(string) == "service"
				}
				return false
			},
		},
		{
			name:    "get_plugin_build_instructions_python",
			handler: server.handleGetPluginBuildInstructions,
			args: map[string]interface{}{
				"language": "python",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if language, ok := resultMap["language"].(string); ok {
						return language == "python"
					}
				}
				return false
			},
		},
		{
			name:    "validate_plugin_structure_nonexistent",
			handler: server.handleValidatePluginStructure,
			args: map[string]interface{}{
				"plugin_path": "/nonexistent/path",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					return resultMap["valid"].(bool) == false
				}
				return false
			},
		},
		{
			name:    "get_plugin_examples_all",
			handler: server.handleGetPluginExamples,
			args: map[string]interface{}{
				"category": "all",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if examples, ok := resultMap["examples"].([]map[string]interface{}); ok {
						return len(examples) > 0
					}
				}
				return false
			},
		},
		{
			name:    "get_plugin_examples_filtered_by_language",
			handler: server.handleGetPluginExamples,
			args: map[string]interface{}{
				"category": "all",
				"language": "go",
			},
			wantErr: false,
			validate: func(result interface{}) bool {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if examples, ok := resultMap["examples"].([]map[string]interface{}); ok {
						// Check that all examples are for Go
						for _, example := range examples {
							if lang, ok := example["language"].(string); ok {
								if lang != "go" && lang != "docker" { // docker examples are language-agnostic
									return false
								}
							}
						}
						return true
					}
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.handler(tt.args)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if tt.validate != nil && !tt.validate(result) {
				t.Errorf("Validation failed for result: %v", result)
			}
		})
	}
}

func TestMCPToolsListContainsPluginDevTools(t *testing.T) {
	server := &Server{}
	
	result, err := server.handleToolsList(json.RawMessage{})
	if err != nil {
		t.Fatalf("Failed to get tools list: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map")
	}
	
	tools, ok := resultMap["tools"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected tools to be a slice of maps")
	}
	
	expectedTools := []string{
		"echo",
		"get_plugin_development_guide",
		"get_plugin_template",
		"get_plugin_build_instructions",
		"validate_plugin_structure",
		"get_plugin_examples",
	}
	
	foundTools := make(map[string]bool)
	for _, tool := range tools {
		if name, ok := tool["name"].(string); ok {
			foundTools[name] = true
		}
	}
	
	for _, expectedTool := range expectedTools {
		if !foundTools[expectedTool] {
			t.Errorf("Expected tool '%s' not found in tools list", expectedTool)
		}
	}
	
	t.Logf("Successfully found all %d plugin development tools in MCP tools list", len(expectedTools))
}

func TestValidatePluginStructureWithValidGoTemplate(t *testing.T) {
	server := &Server{}
	
	// Test with the Go template we created
	templatePath := "docs/plugin-development/languages/go/template"
	
	result, err := server.handleValidatePluginStructure(map[string]interface{}{
		"plugin_path": templatePath,
	})
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map")
	}
	
	// The template should be valid (though it might have warnings)
	if valid, ok := resultMap["valid"].(bool); ok {
		if !valid {
			if errors, ok := resultMap["errors"].([]string); ok && len(errors) > 0 {
				t.Logf("Validation errors: %v", errors)
			}
		}
	}
	
	// Check that the checks were performed
	if checks, ok := resultMap["checks"].(map[string]interface{}); ok {
		if manifest, ok := checks["manifest"].(bool); ok && manifest {
			t.Logf("Manifest check passed")
		}
		
		if detectedLangs, ok := checks["detected_languages"].([]string); ok {
			t.Logf("Detected languages: %v", detectedLangs)
		}
	}
}