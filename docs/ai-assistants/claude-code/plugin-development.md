# Claude Code Plugin Development Guide

This guide shows how to use Claude Code to develop Portunix plugins efficiently, leveraging AI assistance throughout the development lifecycle.

## Overview

Claude Code can assist with every aspect of plugin development:

- **Project Setup**: Initialize plugin structure and configuration
- **Code Generation**: Generate boilerplate code and implement functionality
- **Testing**: Create comprehensive test suites
- **Documentation**: Generate documentation and examples
- **Debugging**: Troubleshoot issues and optimize performance
- **Deployment**: Package and deploy plugins

## Getting Started with Claude Code

### 1. Configure Claude Code with Portunix

First, ensure Claude Code can access Portunix's MCP server:

```bash
# Check if Portunix MCP server is running
portunix mcp status

# Configure Claude Code MCP connection
claude-code config set mcp.servers.portunix.url "http://localhost:8000"
claude-code config set mcp.servers.portunix.enabled true

# Verify connection
claude-code mcp test portunix
```

### 2. AI-Assisted Plugin Creation

Use Claude Code to create a new plugin:

**Human:** "Create a new Go plugin for code analysis that can detect code smells and suggest improvements"

**Claude Code Response:**
```bash
# Claude Code will execute these commands
portunix plugin create code-analyzer --language=go
cd code-analyzer

# Then generate the plugin structure with appropriate functionality
```

Claude Code will:
1. Create the plugin directory structure
2. Generate plugin.yaml with appropriate configuration
3. Implement basic gRPC handlers
4. Add code analysis business logic
5. Create comprehensive tests
6. Generate documentation

## Development Workflows

### Workflow 1: Feature-Driven Development

**Human:** "Add a new MCP tool to analyze Go code complexity"

**Claude Code will:**

1. **Understand Requirements**
   - Analyze the existing plugin structure
   - Understand the current MCP tools
   - Plan the new complexity analysis tool

2. **Implement the Tool**
   - Add tool definition to `ListTools`
   - Implement `complexity_analysis` handler in `CallTool`
   - Create business logic for complexity calculation
   - Add appropriate error handling

3. **Create Tests**
   - Unit tests for complexity calculation
   - Integration tests for MCP tool
   - Test cases with various Go code samples

4. **Update Documentation**
   - Add tool description to README
   - Create usage examples
   - Update API documentation

**Example Generated Code:**

```go
// Added to ListTools
{
    Name:        "analyze_complexity",
    Description: "Analyze Go code complexity and suggest improvements",
    Schema: `{
        "type": "object",
        "properties": {
            "file_path": {"type": "string", "description": "Path to Go file to analyze"},
            "threshold": {"type": "integer", "default": 10, "description": "Complexity threshold for warnings"},
            "include_suggestions": {"type": "boolean", "default": true}
        },
        "required": ["file_path"]
    }`,
}

// Added to CallTool
case "analyze_complexity":
    return h.handleComplexityAnalysis(ctx, req.Arguments)

// Implementation
func (h *PluginHandler) handleComplexityAnalysis(ctx context.Context, args string) (*pb.CallToolResponse, error) {
    var params struct {
        FilePath           string `json:"file_path"`
        Threshold          int    `json:"threshold"`
        IncludeSuggestions bool   `json:"include_suggestions"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_INVALID_ARGS,
            ErrorMessage: fmt.Sprintf("Invalid arguments: %v", err),
        }, nil
    }
    
    result, err := h.service.AnalyzeComplexity(ctx, params.FilePath, params.Threshold, params.IncludeSuggestions)
    if err != nil {
        return &pb.CallToolResponse{
            Status:       pb.CallToolResponse_ERROR,
            ErrorMessage: err.Error(),
        }, nil
    }
    
    return &pb.CallToolResponse{
        Result: result,
        Status: pb.CallToolResponse_SUCCESS,
    }, nil
}
```

### Workflow 2: Test-Driven Development

**Human:** "Let's use TDD to add a security vulnerability scanner to the plugin"

**Claude Code will:**

1. **Write Tests First**
   ```go
   func TestSecurityVulnerabilityScanner(t *testing.T) {
       tests := []struct {
           name           string
           codeContent    string
           expectedVulns  int
           expectedTypes  []string
       }{
           {
               name: "SQL injection vulnerability",
               codeContent: `
                   query := "SELECT * FROM users WHERE id = " + userInput
                   db.Query(query)
               `,
               expectedVulns: 1,
               expectedTypes: []string{"sql_injection"},
           },
           {
               name: "hardcoded credentials",
               codeContent: `
                   password := "admin123"
                   apiKey := "sk-1234567890abcdef"
               `,
               expectedVulns: 2,
               expectedTypes: []string{"hardcoded_credential", "hardcoded_api_key"},
           },
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               result, err := service.ScanSecurity(tt.codeContent)
               assert.NoError(t, err)
               assert.Len(t, result.Vulnerabilities, tt.expectedVulns)
               
               for _, expectedType := range tt.expectedTypes {
                   found := false
                   for _, vuln := range result.Vulnerabilities {
                       if vuln.Type == expectedType {
                           found = true
                           break
                       }
                   }
                   assert.True(t, found, "Expected vulnerability type: %s", expectedType)
               }
           })
       }
   }
   ```

2. **Implement to Pass Tests**
   - Create security scanning service
   - Implement vulnerability detection patterns
   - Add MCP tool interface

3. **Refactor and Optimize**
   - Improve detection accuracy
   - Add more vulnerability patterns
   - Optimize performance

### Workflow 3: Bug Fix and Optimization

**Human:** "The plugin is using too much memory when analyzing large files. Please optimize it."

**Claude Code will:**

1. **Analyze the Problem**
   - Review current implementation
   - Identify memory-intensive operations
   - Profile memory usage patterns

2. **Propose Solutions**
   - Stream processing for large files
   - Implement memory pooling
   - Add garbage collection hints
   - Optimize data structures

3. **Implement Optimizations**
   ```go
   // Before: Loading entire file into memory
   func (s *Service) AnalyzeFile(filePath string) (*AnalysisResult, error) {
       content, err := ioutil.ReadFile(filePath)
       if err != nil {
           return nil, err
       }
       return s.analyzeContent(string(content))
   }
   
   // After: Streaming analysis
   func (s *Service) AnalyzeFile(filePath string) (*AnalysisResult, error) {
       file, err := os.Open(filePath)
       if err != nil {
           return nil, err
       }
       defer file.Close()
       
       scanner := bufio.NewScanner(file)
       scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max line
       
       result := &AnalysisResult{}
       lineNum := 0
       
       for scanner.Scan() {
           lineNum++
           line := scanner.Text()
           s.analyzeLine(line, lineNum, result)
           
           // Process in batches to control memory usage
           if lineNum%1000 == 0 {
               runtime.GC() // Hint for garbage collection
           }
       }
       
       return result, scanner.Err()
   }
   ```

4. **Add Performance Tests**
   ```go
   func BenchmarkAnalyzeLargeFile(b *testing.B) {
       // Create large test file
       tempFile := createLargeTestFile(b, 10*1024*1024) // 10MB
       defer os.Remove(tempFile)
       
       service := NewService()
       
       b.ResetTimer()
       b.ReportAllocs()
       
       for i := 0; i < b.N; i++ {
           _, err := service.AnalyzeFile(tempFile)
           if err != nil {
               b.Fatal(err)
           }
       }
   }
   ```

## Advanced AI-Assisted Patterns

### 1. Code Generation with Context

**Human:** "Generate a plugin that integrates with popular CI/CD systems"

Claude Code will analyze:
- Existing plugin patterns in the codebase
- CI/CD system APIs (GitHub Actions, GitLab CI, Jenkins)
- Authentication and configuration patterns
- Error handling conventions

Then generate:
- Plugin structure following established patterns
- CI/CD system integrations
- Configuration management
- Comprehensive test suite
- Documentation with examples

### 2. API-First Development

**Human:** "Design the MCP tools first for a database management plugin"

Claude Code will:

1. **Design Tool Schemas**
   ```json
   {
     "tools": [
       {
         "name": "execute_query",
         "description": "Execute SQL queries safely with automatic escaping",
         "schema": {
           "type": "object",
           "properties": {
             "query": {"type": "string"},
             "parameters": {"type": "object"},
             "connection": {"type": "string"},
             "dry_run": {"type": "boolean", "default": true}
           }
         }
       },
       {
         "name": "generate_migration",
         "description": "Generate database migration from schema diff",
         "schema": {
           "type": "object",
           "properties": {
             "from_schema": {"type": "string"},
             "to_schema": {"type": "string"},
             "migration_type": {"type": "string", "enum": ["safe", "fast"]}
           }
         }
       }
     ]
   }
   ```

2. **Generate Implementation**
   - Create plugin structure
   - Implement each tool with proper validation
   - Add database connection management
   - Include security measures

3. **Create Mock Tests**
   - Test each tool independently
   - Verify schema validation
   - Test error conditions

### 3. Documentation-Driven Development

**Human:** "Create a plugin for Kubernetes management with comprehensive documentation"

Claude Code will:

1. **Start with Documentation**
   - Create README with clear use cases
   - Document all MCP tools with examples
   - Provide installation and configuration guide
   - Include troubleshooting section

2. **Generate Code from Documentation**
   - Implement features described in docs
   - Ensure examples work correctly
   - Add validation for documented parameters

3. **Keep Documentation Synchronized**
   - Update docs when code changes
   - Validate examples in CI/CD
   - Generate API reference automatically

## Testing Strategies with Claude Code

### 1. Comprehensive Test Generation

**Human:** "Generate comprehensive tests for the authentication module"

Claude Code will create:

```go
// Unit Tests
func TestAuthenticationService(t *testing.T) {
    tests := []struct {
        name           string
        credentials    Credentials
        expectedResult AuthResult
        expectedError  string
    }{
        {
            name: "valid API key authentication",
            credentials: Credentials{
                Type:   "api_key",
                APIKey: "valid-key-123",
            },
            expectedResult: AuthResult{Success: true, UserID: "user-123"},
        },
        {
            name: "invalid API key",
            credentials: Credentials{
                Type:   "api_key",
                APIKey: "invalid-key",
            },
            expectedError: "invalid API key",
        },
        {
            name: "expired token",
            credentials: Credentials{
                Type:  "bearer",
                Token: generateExpiredToken(),
            },
            expectedError: "token expired",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewAuthService()
            result, err := service.Authenticate(tt.credentials)
            
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
        })
    }
}

// Integration Tests
func TestMCPAuthenticationFlow(t *testing.T) {
    // Test complete authentication flow through MCP
    client := setupMCPTestClient(t)
    
    // Test tool authentication
    response, err := client.CallTool("authenticate", map[string]interface{}{
        "method": "api_key",
        "credentials": map[string]string{
            "api_key": "test-key-123",
        },
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "SUCCESS", response.Status)
    
    var result AuthResult
    err = json.Unmarshal([]byte(response.Result), &result)
    assert.NoError(t, err)
    assert.True(t, result.Success)
}

// Performance Tests
func BenchmarkAuthentication(b *testing.B) {
    service := NewAuthService()
    credentials := Credentials{
        Type:   "api_key",
        APIKey: "bench-key-123",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.Authenticate(credentials)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 2. Property-Based Testing

**Human:** "Add property-based tests for the data validation functions"

Claude Code will generate:

```go
import "github.com/leanovate/gopter"
import "github.com/leanovate/gopter/gen"
import "github.com/leanovate/gopter/prop"

func TestDataValidationProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    // Property: Valid email should always pass validation
    properties.Property("valid emails pass validation", prop.ForAll(
        func(localPart, domain string) bool {
            if localPart == "" || domain == "" {
                return true // Skip invalid inputs
            }
            
            email := localPart + "@" + domain + ".com"
            return ValidateEmail(email) == nil
        },
        gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 64 }),
        gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 63 }),
    ))
    
    // Property: Invalid characters should always fail validation
    properties.Property("invalid characters fail validation", prop.ForAll(
        func(email string) bool {
            invalidChars := []string{"<", ">", "(", ")", "[", "]", "\\", ",", ";", ":", "\""}
            hasInvalidChar := false
            
            for _, char := range invalidChars {
                if strings.Contains(email, char) {
                    hasInvalidChar = true
                    break
                }
            }
            
            if hasInvalidChar {
                return ValidateEmail(email) != nil
            }
            return true
        },
        gen.AnyString(),
    ))
    
    properties.TestingRun(t)
}
```

## Debugging and Troubleshooting

### 1. AI-Assisted Debugging

**Human:** "The plugin is crashing when processing large JSON files. Help me debug this."

Claude Code will:

1. **Analyze Stack Traces**
   - Parse error logs and stack traces
   - Identify the root cause
   - Suggest potential fixes

2. **Add Debugging Code**
   ```go
   func (s *Service) ProcessJSON(data []byte) (*Result, error) {
       // Add memory monitoring
       var m1, m2 runtime.MemStats
       runtime.ReadMemStats(&m1)
       defer func() {
           runtime.ReadMemStats(&m2)
           s.logger.Debug("Memory usage", 
               "before", m1.Alloc, 
               "after", m2.Alloc, 
               "delta", m2.Alloc-m1.Alloc)
       }()
       
       // Add size validation
       if len(data) > s.config.MaxJSONSize {
           return nil, fmt.Errorf("JSON too large: %d bytes (max: %d)", 
               len(data), s.config.MaxJSONSize)
       }
       
       // Use streaming parser for large files
       if len(data) > s.config.StreamingThreshold {
           return s.processJSONStreaming(data)
       }
       
       return s.processJSONInMemory(data)
   }
   ```

3. **Create Reproduction Tests**
   ```go
   func TestLargeJSONProcessing(t *testing.T) {
       service := NewService()
       
       // Generate large JSON
       largeJSON := generateLargeJSON(10 * 1024 * 1024) // 10MB
       
       result, err := service.ProcessJSON(largeJSON)
       assert.NoError(t, err)
       assert.NotNil(t, result)
   }
   ```

### 2. Performance Profiling

**Human:** "Profile the plugin performance and optimize the bottlenecks"

Claude Code will:

1. **Add Profiling Code**
   ```go
   import _ "net/http/pprof"
   
   func main() {
       // Start profiling server in development
       if os.Getenv("ENABLE_PROFILING") == "true" {
           go func() {
               log.Println(http.ListenAndServe("localhost:6060", nil))
           }()
       }
       
       // Rest of main function...
   }
   ```

2. **Generate Profiling Scripts**
   ```bash
   #!/bin/bash
   # profile.sh
   
   echo "Starting plugin with profiling..."
   ENABLE_PROFILING=true ./plugin &
   PLUGIN_PID=$!
   
   sleep 5  # Let plugin start
   
   echo "Running CPU profile..."
   go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30 &
   
   echo "Running memory profile..."
   go tool pprof -http=:8082 http://localhost:6060/debug/pprof/heap &
   
   echo "Profiling servers started on :8081 (CPU) and :8082 (Memory)"
   echo "Press Enter to stop..."
   read
   
   kill $PLUGIN_PID
   ```

3. **Analyze and Optimize**
   - Identify CPU and memory bottlenecks
   - Suggest optimizations
   - Implement performance improvements

## Deployment and Distribution

### 1. Automated Release Process

**Human:** "Set up automated releases for the plugin"

Claude Code will create:

```yaml
# .github/workflows/release.yml
name: Release Plugin

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: [linux, windows, darwin]
        arch: [amd64, arm64]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Build
      run: |
        GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} \
        go build -o plugin-${{ matrix.os }}-${{ matrix.arch }} main.go
    
    - name: Package
      run: |
        tar -czf plugin-${{ matrix.os }}-${{ matrix.arch }}.tar.gz \
        plugin-${{ matrix.os }}-${{ matrix.arch }} plugin.yaml
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: plugin-${{ matrix.os }}-${{ matrix.arch }}
        path: plugin-${{ matrix.os }}-${{ matrix.arch }}.tar.gz

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v3
    
    - name: Create Release
      uses: softprops/action-gh-release@v1
      with:
        files: '**/*.tar.gz'
        generate_release_notes: true
```

### 2. Plugin Registry Integration

Claude Code will also:

1. **Generate Plugin Registry Entry**
   ```json
   {
     "name": "code-analyzer",
     "version": "1.0.0",
     "description": "Advanced code analysis and improvement suggestions",
     "author": "Developer Name",
     "license": "MIT",
     "repository": "https://github.com/user/code-analyzer-plugin",
     "download_url": "https://github.com/user/code-analyzer-plugin/releases/download/v1.0.0",
     "checksums": {
       "linux-amd64": "sha256:abc123...",
       "windows-amd64": "sha256:def456...",
       "darwin-amd64": "sha256:ghi789..."
     },
     "dependencies": [],
     "compatibility": {
       "portunix": ">=1.5.0"
     },
     "mcp_tools": [
       {
         "name": "analyze_complexity",
         "description": "Analyze code complexity",
         "category": "analysis"
       }
     ]
   }
   ```

2. **Create Installation Scripts**
   ```bash
   #!/bin/bash
   # install.sh
   
   set -e
   
   PLUGIN_NAME="code-analyzer"
   VERSION=${1:-"latest"}
   
   echo "Installing $PLUGIN_NAME plugin..."
   
   # Detect OS and architecture
   OS=$(uname -s | tr '[:upper:]' '[:lower:]')
   ARCH=$(uname -m)
   
   case $ARCH in
       x86_64) ARCH="amd64" ;;
       aarch64|arm64) ARCH="arm64" ;;
       *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
   esac
   
   # Download and install
   DOWNLOAD_URL="https://github.com/user/code-analyzer-plugin/releases/download/v${VERSION}/plugin-${OS}-${ARCH}.tar.gz"
   
   curl -L "$DOWNLOAD_URL" | tar -xz
   portunix plugin install ./plugin-${OS}-${ARCH}
   
   echo "Plugin installed successfully!"
   ```

## Best Practices for AI-Assisted Development

### 1. Clear Communication

**Good:** "Add a tool that validates Kubernetes YAML files, checks for security issues, and suggests resource optimizations"

**Better:** "Add an MCP tool called 'validate_k8s_yaml' that:
- Takes a YAML file path as input
- Validates Kubernetes API schema compliance
- Checks for security issues (privileged containers, missing security contexts)
- Suggests resource limit optimizations
- Returns structured results with severity levels
- Supports both single files and directory scanning"

### 2. Iterative Development

Break large features into smaller, manageable chunks:

1. **Phase 1:** Basic YAML validation
2. **Phase 2:** Security checks
3. **Phase 3:** Resource optimization suggestions
4. **Phase 4:** Directory scanning support

### 3. Testing-First Approach

Always request tests alongside implementation:

**Human:** "Implement the feature AND generate comprehensive tests including unit tests, integration tests, and error cases"

### 4. Documentation Updates

Request documentation updates with every change:

**Human:** "Update the plugin documentation to reflect the new tool, including usage examples and troubleshooting tips"

## Common Development Scenarios

### Scenario 1: Adding External API Integration

**Human:** "Add integration with GitHub API to analyze repository metrics"

Claude Code will:
1. Add GitHub API client with proper authentication
2. Implement rate limiting and error handling
3. Create MCP tools for repository analysis
4. Add configuration for API credentials
5. Generate tests with mocked API responses
6. Update documentation with setup instructions

### Scenario 2: Performance Optimization

**Human:** "The file processing is too slow. Optimize it for better performance"

Claude Code will:
1. Profile current implementation
2. Identify bottlenecks
3. Implement optimizations (concurrency, caching, streaming)
4. Add performance benchmarks
5. Create performance regression tests
6. Document performance characteristics

### Scenario 3: Security Enhancement

**Human:** "Add security scanning capabilities to detect common vulnerabilities"

Claude Code will:
1. Research common vulnerability patterns
2. Implement detection algorithms
3. Add security rule configuration
4. Create comprehensive test cases
5. Add security reporting features
6. Document security best practices

## Monitoring and Maintenance

### Health Checks and Metrics

Claude Code can generate comprehensive monitoring:

```go
// Health check implementation
func (s *Service) HealthCheck() *HealthStatus {
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }
    
    // Check external dependencies
    if err := s.checkGitHubAPI(); err != nil {
        status.Checks["github_api"] = CheckResult{
            Status: "unhealthy",
            Error:  err.Error(),
        }
        status.Status = "degraded"
    } else {
        status.Checks["github_api"] = CheckResult{Status: "healthy"}
    }
    
    // Check database connection
    if err := s.checkDatabase(); err != nil {
        status.Checks["database"] = CheckResult{
            Status: "unhealthy",
            Error:  err.Error(),
        }
        status.Status = "unhealthy"
    } else {
        status.Checks["database"] = CheckResult{Status: "healthy"}
    }
    
    return status
}

// Metrics collection
func (s *Service) collectMetrics() {
    // Tool usage metrics
    toolUsageGauge.WithLabelValues("analyze_complexity").Set(float64(s.complexityAnalysisCount))
    toolUsageGauge.WithLabelValues("validate_k8s").Set(float64(s.k8sValidationCount))
    
    // Performance metrics
    avgResponseTimeGauge.Set(s.getAverageResponseTime())
    errorRateGauge.Set(s.getErrorRate())
}
```

This comprehensive guide demonstrates how Claude Code can accelerate every aspect of plugin development while maintaining high quality standards and best practices.