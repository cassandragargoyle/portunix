# Issue #009: Configurable Datastore System

## Summary
Implement a configurable datastore system for Portunix that allows data to be stored in different backends through plugins. The core system will provide a unified interface while specific storage implementations (MongoDB, PostgreSQL, Redis, etc.) will be handled by specialized plugins.

## Motivation
- Enable flexible data storage options for different use cases
- Allow specialized storage solutions for specific data types (documents in MongoDB, metrics in InfluxDB, etc.)
- Maintain data independence and allow users to choose their preferred storage backend
- Support enterprise requirements for specific database systems
- Enable future extensibility for new storage technologies

## Requirements

### Core Datastore Interface

#### 1. Unified Storage API
```go
type DatastoreInterface interface {
    Store(ctx context.Context, key string, value interface{}, metadata map[string]interface{}) error
    Retrieve(ctx context.Context, key string, filter map[string]interface{}) (interface{}, error)
    Query(ctx context.Context, criteria QueryCriteria) ([]interface{}, error)
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, pattern string) ([]string, error)
}

type QueryCriteria struct {
    Collection string
    Filter     map[string]interface{}
    Sort       map[string]int
    Limit      int
    Offset     int
}
```

#### 2. Datastore Configuration System
Configuration file: `~/.portunix/datastore.yaml`
```yaml
datastore:
  default_plugin: "file-plugin"
  
  # Routing rules - define what data goes where
  routes:
    - name: "user_documentation"
      pattern: "docs/*"
      plugin: "mongodb-plugin"
      config:
        database: "portunix"
        collection: "user_docs"
        indexes: ["title", "tags", "created_at"]
    
    - name: "system_logs"
      pattern: "logs/*"
      plugin: "elasticsearch-plugin"
      config:
        index: "portunix-logs"
        retention_days: 30
    
    - name: "project_metadata"
      pattern: "projects/*"
      plugin: "postgresql-plugin"
      config:
        database: "portunix_projects"
        table: "project_metadata"
    
    - name: "cache_data"
      pattern: "cache/*"
      plugin: "redis-plugin"
      config:
        db: 0
        ttl: 3600
    
    - name: "fallback"
      pattern: "*"
      plugin: "file-plugin"
      config:
        base_path: "~/.portunix/data"

  # Plugin-specific configurations
  plugins:
    mongodb-plugin:
      connection_string: "mongodb://localhost:27017"
      auth:
        username: "${MONGO_USER}"
        password: "${MONGO_PASSWORD}"
      
    postgresql-plugin:
      connection_string: "postgres://localhost:5432/portunix"
      auth:
        username: "${PG_USER}"
        password: "${PG_PASSWORD}"
      
    redis-plugin:
      host: "localhost"
      port: 6379
      auth:
        password: "${REDIS_PASSWORD}"
      
    elasticsearch-plugin:
      hosts: ["localhost:9200"]
      auth:
        username: "${ES_USER}"
        password: "${ES_PASSWORD}"
```

#### 3. Datastore Management Commands
```bash
# Datastore configuration
portunix datastore config                    # Show current configuration
portunix datastore config edit              # Edit configuration file
portunix datastore config validate          # Validate configuration

# Plugin management
portunix datastore plugins list             # List available datastore plugins
portunix datastore plugins install <name>   # Install datastore plugin
portunix datastore plugins enable <name>    # Enable datastore plugin
portunix datastore plugins test <name>      # Test plugin connection

# Data operations
portunix datastore store <key> <value>      # Store data (uses routing)
portunix datastore get <key>                # Retrieve data
portunix datastore query <pattern>          # Query data by pattern
portunix datastore migrate <from> <to>      # Migrate data between plugins
```

### Datastore Plugin Architecture

#### 1. Plugin Interface Extension
Extend existing plugin protocol with datastore-specific operations:

```protobuf
// Extended plugin service for datastore plugins
service DatastorePluginService {
    // Inherit base plugin functionality
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Health(HealthRequest) returns (HealthResponse);
    rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
    
    // Datastore-specific operations
    rpc Store(StoreRequest) returns (StoreResponse);
    rpc Retrieve(RetrieveRequest) returns (RetrieveResponse);
    rpc Query(QueryRequest) returns (QueryResponse);
    rpc Delete(DeleteRequest) returns (DeleteResponse);
    rpc List(ListRequest) returns (ListResponse);
    
    // Management operations
    rpc TestConnection(TestConnectionRequest) returns (TestConnectionResponse);
    rpc GetStats(GetStatsRequest) returns (GetStatsResponse);
    rpc Migrate(MigrateRequest) returns (MigrateResponse);
}

message StoreRequest {
    string key = 1;
    bytes value = 2;                    // Serialized data
    string content_type = 3;            // JSON, YAML, binary, etc.
    map<string, string> metadata = 4;   // Additional metadata
    map<string, string> config = 5;     // Plugin-specific config
}

message RetrieveRequest {
    string key = 1;
    map<string, string> filter = 2;
    map<string, string> config = 3;
}

message QueryRequest {
    string collection = 1;
    map<string, string> filter = 2;
    map<string, string> sort = 3;
    int32 limit = 4;
    int32 offset = 5;
    map<string, string> config = 6;
}
```

#### 2. Plugin Manifest for Datastore Plugins
```yaml
# plugin.yaml for datastore plugins
name: "mongodb-datastore"
version: "1.0.0"
description: "MongoDB datastore plugin for Portunix"
author: "CassandraGargoyle"
license: "MIT"

plugin:
  type: "grpc"
  category: "datastore"           # New category for datastore plugins
  binary: "./mongodb-datastore"
  port: 9010                      # Datastore plugins use port range 9010-9099
  health_check_interval: 30s
  
dependencies:
  portunix_min_version: "1.1.0"
  os_support: ["linux", "windows", "darwin"]
  
# Datastore-specific configuration
datastore:
  supported_operations: ["store", "retrieve", "query", "delete", "list", "aggregate"]
  supported_data_types: ["json", "bson", "binary"]
  features:
    - "indexing"
    - "aggregation"
    - "transactions"
    - "replication"
  
  # Configuration schema for this datastore
  config_schema:
    connection_string:
      type: "string"
      required: true
      description: "MongoDB connection string"
    database:
      type: "string"
      required: true
      description: "Database name"
    collection:
      type: "string"
      required: false
      default: "portunix_data"
      description: "Default collection name"
    indexes:
      type: "array"
      required: false
      description: "List of fields to index"

permissions:
  network: ["outbound"]
  database: ["full"]
  
commands:
  - name: "mongo"
    description: "MongoDB-specific datastore commands"
    subcommands:
      - "index"      # Index management
      - "aggregate"  # Aggregation operations
      - "stats"      # Database statistics
      - "backup"     # Backup operations
```

### Reference Datastore Plugins

#### 1. File Datastore Plugin (Default)
- Simple file-based storage using JSON/YAML
- Directory structure based on data keys
- Built-in to core Portunix as fallback
- No external dependencies

#### 2. MongoDB Datastore Plugin
- Document storage with indexing
- Aggregation pipeline support
- Grid FS for large files
- Replica set support

#### 3. PostgreSQL Datastore Plugin
- Relational data with JSON columns
- Full-text search capabilities
- ACID transactions
- Connection pooling

#### 4. Redis Datastore Plugin
- In-memory caching and storage
- TTL support for temporary data
- Pub/sub capabilities
- Cluster support

### Implementation Plan

#### Phase 1: Core Datastore Infrastructure
1. Design and implement core datastore interface
2. Create routing configuration system
3. Implement plugin discovery for datastore plugins
4. Create basic file-based datastore (default)

#### Phase 2: Plugin Protocol Extension
1. Extend gRPC protocol for datastore operations
2. Update plugin manager to handle datastore plugins
3. Implement configuration validation system
4. Add datastore management commands

#### Phase 3: First Datastore Plugin (MongoDB)
1. Implement MongoDB datastore plugin
2. Create plugin manifest and configuration
3. Add MongoDB-specific features (indexing, aggregation)
4. Integration testing with routing system

#### Phase 4: Additional Plugins
1. Implement PostgreSQL datastore plugin
2. Implement Redis datastore plugin
3. Create plugin templates for datastore development
4. Documentation and examples

#### Phase 5: Advanced Features
1. Data migration tools between datastores
2. Backup and restore functionality
3. Performance monitoring and metrics
4. Multi-datastore transactions (where possible)

#### Phase 6: AI Integration
1. MCP tools for datastore operations
2. AI-assisted data modeling and querying
3. Intelligent data routing suggestions
4. Automated backup and maintenance

## Use Cases

### Example 1: User Documentation Management
```bash
# Configure MongoDB for user documentation
portunix datastore config edit  # Add MongoDB route for docs/*

# Store user documentation (automatically routed to MongoDB)
echo "# My Project" | portunix datastore store "docs/my-project.md" -

# Query documentation by tags
portunix datastore query "docs/*" --filter "tags:contains:tutorial"

# MongoDB-specific aggregation
portunix mongo aggregate docs --pipeline '[{"$group": {"_id": "$tags", "count": {"$sum": 1}}}]'
```

### Example 2: Multi-Backend Setup
```yaml
# Different data types in different stores
routes:
  - pattern: "docs/*"           # User docs → MongoDB
    plugin: "mongodb-plugin"
  - pattern: "metrics/*"        # Metrics → InfluxDB
    plugin: "influxdb-plugin"
  - pattern: "cache/*"          # Cache → Redis
    plugin: "redis-plugin"
  - pattern: "projects/*"       # Project metadata → PostgreSQL
    plugin: "postgresql-plugin"
```

### Example 3: Development Workflow
```bash
# Development: Use file-based storage
portunix datastore config set default_plugin file-plugin

# Production: Switch to MongoDB for scalability
portunix datastore config set default_plugin mongodb-plugin

# Migrate data from file to MongoDB
portunix datastore migrate file-plugin mongodb-plugin --pattern "docs/*"
```

## Technical Requirements
- Maintain backward compatibility with existing data storage
- Support for data migration between different datastores
- Plugin isolation and security (each plugin runs in separate process)
- Configuration validation and error handling
- Connection pooling and performance optimization
- Cross-platform compatibility (Linux, Windows, macOS)

## Success Criteria
- [ ] Core datastore interface provides unified API for all storage operations
- [ ] Routing system correctly directs data to appropriate datastore plugins
- [ ] MongoDB plugin provides full document storage capabilities
- [ ] Configuration system is intuitive and well-documented
- [ ] Data migration tools work reliably between different datastores
- [ ] Performance is comparable to direct database access
- [ ] Plugin development is straightforward with clear documentation
- [ ] AI integration allows natural language data operations

## Benefits
- **Flexibility**: Choose appropriate storage backend for each data type
- **Scalability**: Easy migration from simple file storage to enterprise databases
- **Modularity**: Core remains lightweight, storage complexity in plugins
- **Future-proof**: Easy to add support for new storage technologies
- **Enterprise-ready**: Support for enterprise database requirements
- **Development-friendly**: Simple file storage for development, production databases for deployment

## Priority
**High** - Datastore system is fundamental for data persistence and enables enterprise adoption.

## Labels
- enhancement
- datastore
- plugin-system
- mongodb
- postgresql
- redis
- enterprise
- scalability
- data-management