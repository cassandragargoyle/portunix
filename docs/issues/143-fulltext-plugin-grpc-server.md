# Issue #143: Fulltext Plugin gRPC Server Mode

**Type**: Feature / Plugin Enhancement
**Priority**: Medium
**Status**: ✅ Implemented
**Created**: 2026-01-27
**Labels**: plugin, fulltext, grpc, integration
**Related**: Issue #141 (PTX-TRACE Helper Implementation)
**Repository**: portunix-plugins

> **IMPLEMENTATION NOTE**: This feature was implemented in **portunix-plugins** repository.
> See: `portunix-plugins/docs/issues/internal/017-fulltext-plugin-grpc-server.md`
>
> **Implemented**: 2026-01-27
> - gRPC server mode (`portunix fulltext serve`)
> - Proto definition with all RPCs
> - Python integration tests (8/8 passing)
> - Version: fulltext-plugin 1.1.0-SNAPSHOT

## Summary

Extend the fulltext plugin to run as a gRPC server, enabling PTX-TRACE and other Portunix components to communicate with it programmatically for document indexing and search operations.

## Problem Statement

The fulltext plugin currently operates only as a CLI tool. For seamless integration with PTX-TRACE export functionality, we need:

- Real-time indexing of trace events
- Programmatic search API
- Efficient bulk operations
- Cross-process communication

## Requirements

### Functional Requirements

1. **gRPC Server Mode**

```bash
   portunix fulltext serve                    # Start gRPC server
   portunix fulltext serve --port 50051       # Custom port
   portunix fulltext serve --host 0.0.0.0     # Bind to all interfaces
```

2. **gRPC Service Definition**

```protobuf
   service FulltextService {
     // Document indexing
     rpc IndexDocument(IndexRequest) returns (IndexResponse);
     rpc IndexBatch(stream IndexRequest) returns (BatchIndexResponse);
     rpc DeleteDocument(DeleteRequest) returns (DeleteResponse);

     // Search
     rpc Search(SearchRequest) returns (SearchResponse);
     rpc SearchStream(SearchRequest) returns (stream SearchResult);

     // Index management
     rpc GetStats(StatsRequest) returns (StatsResponse);
     rpc EnsureIndex(EnsureIndexRequest) returns (EnsureIndexResponse);
     rpc DeleteIndex(DeleteIndexRequest) returns (DeleteIndexResponse);

     // Health
     rpc HealthCheck(HealthRequest) returns (HealthResponse);
   }
```

3. **PTX-TRACE Integration**

   - Export trace events to fulltext index
   - Search trace events by content
   - Filter by session, operation, time range
   - Highlight matching terms

4. **Configuration**

```json
   {
     "grpc": {
       "enabled": true,
       "port": 50051,
       "host": "localhost",
       "max_connections": 100,
       "timeout_seconds": 30
     }
   }
```

### Non-Functional Requirements

- Server starts within 5 seconds
- Supports concurrent connections
- Graceful shutdown
- Health check endpoint for monitoring
- TLS support (optional)

## Technical Design

### Architecture

```text
┌─────────────────────────────────────────────────────────┐
│                    PTX-TRACE                            │
│  ┌─────────────────────────────────────────────────┐    │
│  │  export fulltext command                        │    │
│  └──────────────────────┬──────────────────────────┘    │
└─────────────────────────┼───────────────────────────────┘
                          │ gRPC
                          ▼
┌─────────────────────────────────────────────────────────┐
│               Fulltext Plugin (gRPC Server)             │
│  ┌─────────────────────────────────────────────────┐    │
│  │  FulltextServiceImpl                            │    │
│  └──────────────────────┬──────────────────────────┘    │
│                         │                               │
│  ┌──────────────────────▼──────────────────────────┐    │
│  │  SearchEngine (Lucene / Elasticsearch)          │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### Proto Definition

```protobuf
syntax = "proto3";

package portunix.fulltext.v1;

option java_package = "ai.portunix.fulltext.grpc";

message IndexRequest {
  string id = 1;
  string content = 2;
  map<string, string> metadata = 3;
  string content_type = 4;
}

message IndexResponse {
  bool success = 1;
  string message = 2;
}

message SearchRequest {
  string query = 1;
  int32 limit = 2;
  int32 offset = 3;
  bool highlight = 4;
  map<string, string> filters = 5;
}

message SearchResult {
  string id = 1;
  float score = 2;
  string highlight = 3;
  map<string, string> metadata = 4;
}

message SearchResponse {
  repeated SearchResult results = 1;
  int64 total_hits = 2;
  int64 took_ms = 3;
}

// ... additional messages
```

### Java Implementation

```java
public class FulltextGrpcServer {
    private final Server server;
    private final SearchEngine engine;

    public void start() throws IOException {
        server = ServerBuilder.forPort(port)
            .addService(new FulltextServiceImpl(engine))
            .build()
            .start();
    }

    public void blockUntilShutdown() throws InterruptedException {
        server.awaitTermination();
    }
}
```

## Implementation Tasks

### Plugin Repository (portunix-plugins) - ✅ DONE

- [x] Add gRPC dependencies to pom.xml
- [x] Create proto definition file
- [x] Generate Java gRPC stubs
- [x] Implement FulltextServiceImpl
- [x] Implement FulltextGrpcServer
- [x] Add `serve` command to CLI
- [x] Add configuration support for gRPC
- [x] Implement health check service
- [x] Add graceful shutdown handling
- [ ] Write unit tests (skipped - have integration tests)
- [x] Write integration tests (Python)

### PTX-TRACE Integration (portunix repo) - ✅ DONE

- [x] Add gRPC client to ptx-trace
- [x] Implement `export fulltext` command
- [x] Handle connection errors gracefully
- [x] Add retry logic for transient failures

## Acceptance Criteria

- [x] `portunix fulltext serve` starts gRPC server
- [x] Server responds to health checks
- [x] IndexDocument RPC works correctly
- [x] Search RPC returns expected results
- [x] PTX-TRACE can export events to fulltext index
- [x] Server handles concurrent requests
- [x] Graceful shutdown works properly
- [x] Configuration via fulltext.json works

## Dependencies

- Issue #141: PTX-TRACE Helper Implementation (Phase 4b)
- Issue #142: Elasticsearch container installation
- gRPC Java library
- Protocol Buffers compiler

## Notes

- Plugin is in separate repository: portunix-plugins
- Follow existing plugin patterns (text-extractor, agile)
- Consider backward compatibility with CLI-only mode
- gRPC server mode should be optional (not required for basic CLI usage)
