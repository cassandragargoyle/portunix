# Issue #142: Elasticsearch Container Installation

**Type**: Feature / Installation
**Priority**: Medium
**Status**: ✅ Implemented
**Created**: 2026-01-27
**Labels**: installation, container, elasticsearch, fulltext
**Related**: Issue #141 (PTX-TRACE Helper Implementation)

## Summary

Implement Elasticsearch installation support via container for `portunix install elasticsearch`. This enables local Elasticsearch instances for development, testing, and integration with the fulltext plugin and PTX-TRACE system.

## Problem Statement

Currently there is no easy way to install and run Elasticsearch locally using Portunix. Users need to manually set up Elasticsearch for:

- Local development and testing
- Integration with fulltext plugin
- PTX-TRACE fulltext search export
- Document indexing and search functionality

## Requirements

### Functional Requirements

1. **Container-based Installation**
   - Use official Elasticsearch Docker image
   - Support for Elasticsearch 8.x
   - Single-node development mode
   - Configurable memory limits

2. **CLI Commands**

```bash
   portunix install elasticsearch              # Install and start ES container
   portunix install elasticsearch --version 8.12.0
   portunix install elasticsearch --memory 2g  # Set heap size
   portunix install elasticsearch --port 9200  # Custom port
```

3. **Configuration**

   - Default single-node cluster
   - Security disabled for local development (optional)
   - Persistent data volume
   - Health check endpoint

4. **Integration Points**

   - Works with portunix fulltext plugin
   - Works with PTX-TRACE fulltext export
   - Container lifecycle management via `portunix container`

### Non-Functional Requirements

- Container starts within 60 seconds
- Memory footprint configurable (default 1GB)
- Data persists across container restarts
- Compatible with both Docker and Podman

## Technical Design

### Container Configuration

```yaml
image: docker.elastic.co/elasticsearch/elasticsearch:8.12.0
environment:
  - discovery.type=single-node
  - xpack.security.enabled=false
  - ES_JAVA_OPTS=-Xms512m -Xmx512m
ports:
  - 9200:9200
  - 9300:9300
volumes:
  - elasticsearch-data:/usr/share/elasticsearch/data
healthcheck:
  test: curl -s http://localhost:9200/_cluster/health
  interval: 30s
  timeout: 10s
  retries: 5
```

### Package Definition (assets/install-packages.json)

```json
{
  "elasticsearch": {
    "name": "Elasticsearch",
    "description": "Distributed search and analytics engine",
    "category": "database",
    "container": {
      "image": "docker.elastic.co/elasticsearch/elasticsearch",
      "tag": "8.12.0",
      "ports": ["9200:9200", "9300:9300"],
      "environment": {
        "discovery.type": "single-node",
        "xpack.security.enabled": "false"
      },
      "volumes": ["elasticsearch-data:/usr/share/elasticsearch/data"],
      "healthcheck": {
        "endpoint": "http://localhost:9200/_cluster/health",
        "timeout": 60
      }
    }
  }
}
```

## Implementation Tasks

- [ ] Add elasticsearch package definition to assets/install-packages.json
- [ ] Implement container-based installation in install system
- [ ] Add health check wait logic
- [ ] Add memory configuration option
- [ ] Add version selection option
- [ ] Test with Docker runtime
- [ ] Test with Podman runtime
- [ ] Add documentation

## Acceptance Criteria

- [ ] `portunix install elasticsearch` starts Elasticsearch container
- [ ] Container is accessible on port 9200
- [ ] Health check passes before returning success
- [ ] Data persists across container restarts
- [ ] Works with both Docker and Podman
- [ ] Memory limits are configurable
- [ ] Integration with fulltext plugin verified

## Dependencies

- Container runtime (Docker or Podman) must be available
- Issue #141: PTX-TRACE Helper Implementation (Phase 4b)

## Notes

- Elasticsearch 8.x requires more memory than 7.x
- Security features disabled by default for local development
- Production deployments should enable security
