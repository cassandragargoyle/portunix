# Exposing MCP Tools from Plugins

The Model Context Protocol (MCP) allows plugins to expose their functionality to AI agents like Claude Code. This guide shows how to implement MCP tools in your plugins to enable AI-assisted development workflows.

## Overview

MCP tools are functions that AI agents can call to perform specific tasks. When you expose tools from your plugin, AI agents can:

- Discover available tools through the `ListTools` RPC
- Call tools with structured parameters through the `CallTool` RPC
- Receive structured responses for further processing

## Basic MCP Implementation

### 1. Define Tool Schema

First, define your tools using JSON Schema to describe parameters and their types:

```json
{
  "name": "generate_code",
  "description": "Generate code based on specifications",
  "schema": {
    "type": "object",
    "properties": {
      "language": {
        "type": "string",
        "enum": ["go", "python", "javascript", "java"],
        "description": "Programming language for the generated code"
      },
      "specification": {
        "type": "string",
        "description": "Detailed specification of what to generate"
      },
      "style": {
        "type": "string",
        "enum": ["functional", "object-oriented", "minimal"],
        "default": "functional",
        "description": "Code style preference"
      }
    },
    "required": ["language", "specification"]
  }
}
```

### 2. Implement ListTools RPC

The `ListTools` RPC should return all available tools:

```go
// Go implementation
func (h *PluginHandler) ListTools(ctx context.Context, req *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
    tools := []*pb.MCPTool{
        {
            Name:        "generate_code",
            Description: "Generate code based on specifications",
            Schema: `{
                "type": "object",
                "properties": {
                    "language": {"type": "string", "enum": ["go", "python", "javascript", "java"]},
                    "specification": {"type": "string"},
                    "style": {"type": "string", "enum": ["functional", "object-oriented", "minimal"], "default": "functional"}
                },
                "required": ["language", "specification"]
            }`,
        },
        {
            Name:        "validate_code",
            Description: "Validate code syntax and style",
            Schema: `{
                "type": "object",
                "properties": {
                    "code": {"type": "string"},
                    "language": {"type": "string"},
                    "strict": {"type": "boolean", "default": false}
                },
                "required": ["code", "language"]
            }`,
        },
    }
    
    return &pb.ListToolsResponse{Tools: tools}, nil
}
```

```python
# Python implementation
def ListTools(self, request, context):
    """List available MCP tools."""
    tools = [
        plugin_pb2.MCPTool(
            name="generate_code",
            description="Generate code based on specifications",
            schema=json.dumps({
                "type": "object",
                "properties": {
                    "language": {"type": "string", "enum": ["go", "python", "javascript", "java"]},
                    "specification": {"type": "string"},
                    "style": {"type": "string", "enum": ["functional", "object-oriented", "minimal"], "default": "functional"}
                },
                "required": ["language", "specification"]
            })
        ),
        plugin_pb2.MCPTool(
            name="validate_code",
            description="Validate code syntax and style",
            schema=json.dumps({
                "type": "object",
                "properties": {
                    "code": {"type": "string"},
                    "language": {"type": "string"},
                    "strict": {"type": "boolean", "default": False}
                },
                "required": ["code", "language"]
            })
        )
    ]
    
    return plugin_pb2.ListToolsResponse(tools=tools)
```

### 3. Implement CallTool RPC

The `CallTool` RPC handles actual tool execution:

```go
// Go implementation
func (h *PluginHandler) CallTool(ctx context.Context, req *pb.CallToolRequest) (*pb.CallToolResponse, error) {
    switch req.ToolName {
    case "generate_code":
        return h.handleGenerateCode(ctx, req.Arguments)
    case "validate_code":
        return h.handleValidateCode(ctx, req.Arguments)
    default:
        return nil, fmt.Errorf("unknown tool: %s", req.ToolName)
    }
}

func (h *PluginHandler) handleGenerateCode(ctx context.Context, args string) (*pb.CallToolResponse, error) {
    var params struct {
        Language      string `json:"language"`
        Specification string `json:"specification"`
        Style         string `json:"style"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_INVALID_ARGS,
            ErrorMessage: fmt.Sprintf("Invalid arguments: %v", err),
        }, nil
    }
    
    // Generate code using your service
    code, err := h.service.GenerateCode(ctx, params.Language, params.Specification, params.Style)
    if err != nil {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_ERROR,
            ErrorMessage: err.Error(),
        }, nil
    }
    
    return &pb.CallToolResponse{
        Result: code,
        Status: pb.CallToolResponse_SUCCESS,
    }, nil
}
```

```python
# Python implementation
def CallTool(self, request, context):
    """Call an MCP tool."""
    try:
        if request.tool_name == "generate_code":
            return self._handle_generate_code(request.arguments)
        elif request.tool_name == "validate_code":
            return self._handle_validate_code(request.arguments)
        else:
            return plugin_pb2.CallToolResponse(
                status=plugin_pb2.CallToolResponse.NOT_FOUND,
                error_message=f"Tool not found: {request.tool_name}"
            )
    except Exception as e:
        return plugin_pb2.CallToolResponse(
            status=plugin_pb2.CallToolResponse.ERROR,
            error_message=str(e)
        )

def _handle_generate_code(self, arguments: str) -> plugin_pb2.CallToolResponse:
    """Handle generate_code tool."""
    try:
        params = json.loads(arguments)
    except json.JSONDecodeError as e:
        return plugin_pb2.CallToolResponse(
            status=plugin_pb2.CallToolResponse.INVALID_ARGS,
            error_message=f"Invalid JSON arguments: {e}"
        )
    
    language = params.get("language")
    specification = params.get("specification")
    style = params.get("style", "functional")
    
    # Generate code using your service
    code = self.service.generate_code(language, specification, style)
    
    return plugin_pb2.CallToolResponse(
        result=code,
        status=plugin_pb2.CallToolResponse.SUCCESS
    )
```

## Advanced MCP Patterns

### 1. Multi-Step Workflows

Create tools that can be chained together for complex workflows:

```json
{
  "name": "analyze_project",
  "description": "Analyze project structure and suggest improvements",
  "schema": {
    "type": "object",
    "properties": {
      "project_path": {"type": "string"},
      "analysis_depth": {"type": "string", "enum": ["quick", "thorough"], "default": "quick"},
      "focus_areas": {
        "type": "array",
        "items": {"type": "string", "enum": ["security", "performance", "maintainability", "testing"]},
        "default": ["maintainability"]
      }
    },
    "required": ["project_path"]
  }
}
```

### 2. File Operation Tools

Tools for reading, writing, and manipulating files:

```json
{
  "name": "read_file_with_context",
  "description": "Read file content with surrounding context for better understanding",
  "schema": {
    "type": "object",
    "properties": {
      "file_path": {"type": "string"},
      "line_range": {
        "type": "object",
        "properties": {
          "start": {"type": "integer"},
          "end": {"type": "integer"}
        }
      },
      "include_imports": {"type": "boolean", "default": true},
      "include_comments": {"type": "boolean", "default": true}
    },
    "required": ["file_path"]
  }
}
```

### 3. Interactive Tools

Tools that can prompt for additional information:

```json
{
  "name": "interactive_refactor",
  "description": "Interactively refactor code with AI guidance",
  "schema": {
    "type": "object",
    "properties": {
      "code": {"type": "string"},
      "refactor_type": {"type": "string", "enum": ["extract_method", "rename_variable", "optimize_performance"]},
      "confirmation_required": {"type": "boolean", "default": true}
    },
    "required": ["code", "refactor_type"]
  }
}
```

## Tool Design Best Practices

### 1. Clear and Descriptive Names

- Use verb-noun pattern: `generate_dockerfile`, `validate_config`, `analyze_dependencies`
- Be specific: `format_go_code` instead of `format_code`
- Avoid abbreviations: `validate_configuration` instead of `validate_config`

### 2. Comprehensive Descriptions

```json
{
  "name": "optimize_database_query",
  "description": "Analyze and optimize database queries for better performance. Supports SQL and NoSQL queries. Returns optimized query with performance metrics and suggestions.",
  "schema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "The database query to optimize (SQL or NoSQL syntax)"
      },
      "database_type": {
        "type": "string",
        "enum": ["mysql", "postgresql", "mongodb", "elasticsearch"],
        "description": "Type of database system for query optimization"
      },
      "performance_target": {
        "type": "string",
        "enum": ["speed", "memory", "balanced"],
        "default": "balanced",
        "description": "Optimization target: prioritize query speed, memory usage, or balanced approach"
      }
    },
    "required": ["query", "database_type"]
  }
}
```

### 3. Robust Error Handling

```go
func (h *PluginHandler) handleOptimizeQuery(ctx context.Context, args string) (*pb.CallToolResponse, error) {
    var params struct {
        Query            string `json:"query"`
        DatabaseType     string `json:"database_type"`
        PerformanceTarget string `json:"performance_target"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_INVALID_ARGS,
            ErrorMessage: "Invalid JSON format in arguments",
        }, nil
    }
    
    // Validate required parameters
    if params.Query == "" {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_INVALID_ARGS,
            ErrorMessage: "Query parameter is required and cannot be empty",
        }, nil
    }
    
    // Validate enum values
    validDatabases := map[string]bool{
        "mysql": true, "postgresql": true, "mongodb": true, "elasticsearch": true,
    }
    if !validDatabases[params.DatabaseType] {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_INVALID_ARGS,
            ErrorMessage: fmt.Sprintf("Unsupported database type: %s", params.DatabaseType),
        }, nil
    }
    
    // Set default values
    if params.PerformanceTarget == "" {
        params.PerformanceTarget = "balanced"
    }
    
    // Perform optimization
    result, err := h.service.OptimizeQuery(ctx, params.Query, params.DatabaseType, params.PerformanceTarget)
    if err != nil {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_ERROR,
            ErrorMessage: fmt.Sprintf("Query optimization failed: %v", err),
        }, nil
    }
    
    return &pb.CallToolResponse{
        Result: result,
        Status: pb.CallToolResponse_SUCCESS,
    }, nil
}
```

### 4. Structured Return Values

Return structured data that AI agents can easily process:

```go
type OptimizationResult struct {
    OriginalQuery     string                 `json:"original_query"`
    OptimizedQuery    string                 `json:"optimized_query"`
    PerformanceGain   float64               `json:"performance_gain_percent"`
    Suggestions       []string              `json:"suggestions"`
    Warnings          []string              `json:"warnings"`
    ExecutionPlan     map[string]interface{} `json:"execution_plan"`
    EstimatedCost     EstimatedCost         `json:"estimated_cost"`
}

type EstimatedCost struct {
    CPU    float64 `json:"cpu_units"`
    Memory float64 `json:"memory_mb"`
    IO     float64 `json:"io_operations"`
}

func (s *Service) OptimizeQuery(ctx context.Context, query, dbType, target string) (string, error) {
    result := OptimizationResult{
        OriginalQuery:   query,
        OptimizedQuery:  optimizedQuery,
        PerformanceGain: 35.7,
        Suggestions: []string{
            "Consider adding index on 'user_id' column",
            "Use LIMIT clause to reduce result set size",
        },
        Warnings: []string{
            "Query contains potentially expensive LIKE operation",
        },
        ExecutionPlan: map[string]interface{}{
            "type": "index_scan",
            "table": "users",
            "estimated_rows": 1500,
        },
        EstimatedCost: EstimatedCost{
            CPU:    0.25,
            Memory: 12.5,
            IO:     45,
        },
    }
    
    resultJSON, err := json.Marshal(result)
    if err != nil {
        return "", fmt.Errorf("failed to marshal result: %w", err)
    }
    
    return string(resultJSON), nil
}
```

## Testing MCP Tools

### Unit Testing

```go
func TestMCPToolExecution(t *testing.T) {
    handler := setupTestHandler()
    
    tests := []struct {
        name           string
        toolName       string
        arguments      string
        expectedStatus pb.CallToolResponse_Status
        expectedResult string
    }{
        {
            name:     "valid generate code request",
            toolName: "generate_code",
            arguments: `{
                "language": "go",
                "specification": "HTTP server with health check endpoint",
                "style": "functional"
            }`,
            expectedStatus: pb.CallToolResponse_SUCCESS,
        },
        {
            name:     "invalid arguments",
            toolName: "generate_code",
            arguments: `{"invalid": "json"}`,
            expectedStatus: pb.CallToolResponse_INVALID_ARGS,
        },
        {
            name:     "unknown tool",
            toolName: "unknown_tool",
            arguments: `{}`,
            expectedStatus: pb.CallToolResponse_NOT_FOUND,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &pb.CallToolRequest{
                ToolName:  tt.toolName,
                Arguments: tt.arguments,
            }
            
            resp, err := handler.CallTool(context.Background(), req)
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expectedStatus, resp.Status)
            
            if tt.expectedResult != "" {
                assert.Contains(t, resp.Result, tt.expectedResult)
            }
        })
    }
}
```

### Integration Testing with Claude Code

```python
import asyncio
import json
from mcp_client import MCPClient

async def test_mcp_integration():
    """Test MCP tools integration with Claude Code."""
    client = MCPClient("localhost:50051")
    
    # Test tool discovery
    tools = await client.list_tools()
    assert len(tools) > 0
    
    tool_names = [tool.name for tool in tools]
    assert "generate_code" in tool_names
    
    # Test tool execution
    result = await client.call_tool(
        "generate_code",
        {
            "language": "python",
            "specification": "Function to calculate fibonacci numbers",
            "style": "functional"
        }
    )
    
    assert result.status == "SUCCESS"
    assert "def fibonacci" in result.result
    assert "python" in result.result.lower()

if __name__ == "__main__":
    asyncio.run(test_mcp_integration())
```

## Plugin Manifest Configuration

Configure MCP tools in your plugin manifest:

```yaml
name: code-generator-plugin
version: 1.0.0
description: AI-powered code generation and analysis tools
type: mcp

mcp:
  tools:
    - name: generate_code
      description: Generate code based on specifications
      category: generation
      complexity: medium
      
    - name: analyze_project
      description: Analyze project structure and suggest improvements
      category: analysis
      complexity: high
      
    - name: validate_code
      description: Validate code syntax and style
      category: validation
      complexity: low

  categories:
    - generation: Code generation and scaffolding
    - analysis: Code analysis and metrics
    - validation: Code validation and linting
    - refactoring: Code refactoring and optimization

  ai_integration:
    claude_code:
      priority: high
      auto_discover: true
      suggestions_enabled: true
```

## Real-World Examples

### 1. Docker Tool
```json
{
  "name": "generate_dockerfile",
  "description": "Generate optimized Dockerfile for any project type",
  "schema": {
    "type": "object",
    "properties": {
      "project_path": {"type": "string"},
      "base_image": {"type": "string", "default": "auto"},
      "optimization_level": {"type": "string", "enum": ["size", "speed", "security"], "default": "balanced"},
      "include_healthcheck": {"type": "boolean", "default": true},
      "multi_stage": {"type": "boolean", "default": true}
    },
    "required": ["project_path"]
  }
}
```

### 2. Database Migration Tool
```json
{
  "name": "generate_migration",
  "description": "Generate database migration scripts from schema changes",
  "schema": {
    "type": "object",
    "properties": {
      "old_schema": {"type": "string"},
      "new_schema": {"type": "string"},
      "database_type": {"type": "string", "enum": ["postgresql", "mysql", "sqlite"]},
      "migration_type": {"type": "string", "enum": ["safe", "fast", "minimal"], "default": "safe"},
      "include_rollback": {"type": "boolean", "default": true}
    },
    "required": ["old_schema", "new_schema", "database_type"]
  }
}
```

### 3. API Documentation Tool
```json
{
  "name": "generate_api_docs",
  "description": "Generate comprehensive API documentation from code",
  "schema": {
    "type": "object",
    "properties": {
      "source_path": {"type": "string"},
      "output_format": {"type": "string", "enum": ["openapi", "markdown", "html"], "default": "openapi"},
      "include_examples": {"type": "boolean", "default": true},
      "include_schemas": {"type": "boolean", "default": true},
      "authentication_info": {"type": "object"}
    },
    "required": ["source_path"]
  }
}
```

## Monitoring and Analytics

Track MCP tool usage to improve AI assistance:

```go
type MCPMetrics struct {
    ToolCalls     prometheus.CounterVec
    ToolDuration  prometheus.HistogramVec
    ToolErrors    prometheus.CounterVec
    AIAgentUsage  prometheus.CounterVec
}

func (h *PluginHandler) CallTool(ctx context.Context, req *pb.CallToolRequest) (*pb.CallToolResponse, error) {
    timer := prometheus.NewTimer(h.metrics.ToolDuration.WithLabelValues(req.ToolName))
    defer timer.ObserveDuration()
    
    h.metrics.ToolCalls.WithLabelValues(req.ToolName, "claude_code").Inc()
    
    response, err := h.executeTool(ctx, req)
    
    if err != nil {
        h.metrics.ToolErrors.WithLabelValues(req.ToolName, "execution_error").Inc()
    } else if response.Status != pb.CallToolResponse_SUCCESS {
        h.metrics.ToolErrors.WithLabelValues(req.ToolName, response.Status.String()).Inc()
    }
    
    return response, err
}
```

## Next Steps

1. **Start Simple**: Begin with basic tools that perform single, well-defined tasks
2. **Iterate Based on Usage**: Monitor which tools are used most and improve them
3. **Add Complexity Gradually**: Build more sophisticated tools as you understand AI agent patterns
4. **Test Extensively**: Ensure tools work reliably with different AI agents
5. **Document Thoroughly**: Provide clear examples and use cases

## Resources

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [JSON Schema Documentation](https://json-schema.org/)
- [Claude Code MCP Integration Guide](../../../ai-assistants/claude-code/plugin-development.md)
- [Plugin Testing Guide](../../testing/mcp-testing.md)