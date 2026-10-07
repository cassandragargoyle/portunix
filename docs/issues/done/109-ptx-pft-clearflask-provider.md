# Issue #109: PTX-PFT ClearFlask Provider Implementation

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Labels**: enhancement, helper-binary, product-feedback, clearflask, provider
**Depends on**: #107 (PTX-PFT base implementation)

---

## Summary

Implement ClearFlask provider for ptx-pft helper, enabling integration with ClearFlask as an alternative to Fider.io. ClearFlask is an open-source feedback management tool (alternative to Canny and UserVoice) with more advanced features including roadmap visualization, multi-project support, and customizable workflows.

## ClearFlask Overview

**Repository**: https://github.com/clearflask/clearflask
**License**: Apache 2.0
**Features**:
- Feedback collection and management dashboard
- Roadmap visualization and planning
- User voting and engagement
- Multi-project support with customizable workflows

## Architecture Consideration

### Provider Abstraction (Existing)

PTX-PFT already implements `FeedbackProvider` interface (Issue #107):

```go
type FeedbackProvider interface {
    Name() string
    Connect(config ProviderConfig) error
    List() ([]FeedbackItem, error)
    Get(id string) (*FeedbackItem, error)
    Create(item FeedbackItem) error
    Update(item FeedbackItem) error
    Delete(id string) error
    Close() error
}
```

ClearFlask provider will implement this interface, reusing existing sync engine and all pft commands.

### Deployment Complexity

ClearFlask has significantly more complex deployment requirements than Fider:

| Component | Fider | ClearFlask |
|-----------|-------|------------|
| Backend | Go (single binary) | Java (Tomcat) |
| Database | PostgreSQL only | MySQL + DynamoDB + Elasticsearch |
| Storage | - | MinIO/S3 |
| Memory | ~100MB | ~512MB+ |
| Startup | ~5s | ~30s+ |

**Decision needed**: Should ClearFlask deployment be:
1. **Full local** - All components in Docker Compose (heavier, ~2GB RAM)
2. **Hybrid** - Local MySQL + Elasticsearch, optional S3
3. **External only** - Only connect to existing ClearFlask instances

## Requirements

### Functional Requirements

1. **ClearFlask Provider**
   - Implement `FeedbackProvider` interface for ClearFlask API
   - Support all CRUD operations (posts, comments, votes)
   - Handle ClearFlask-specific features (roadmap, tags, categories)

2. **Configuration**
   - Provider selection in `.pft-config.json`
   - API endpoint and authentication
   - Project/board selection (ClearFlask supports multiple)

3. **Status Mapping**
   - Map ClearFlask statuses to pft internal statuses
   - Handle ClearFlask custom workflows

4. **Optional: Container Deployment**
   - Docker Compose template for local ClearFlask instance
   - `pft deploy --provider clearflask`
   - Health checks and dependency management

### Technical Requirements

1. **API Client**
   - REST API client for ClearFlask
   - Authentication (API key or OAuth)
   - Pagination handling
   - Rate limiting

2. **Data Mapping**
   - Map ClearFlask post structure to `FeedbackItem`
   - Handle multi-project scenarios
   - Preserve ClearFlask-specific metadata

## Command Changes

```bash
# Configuration with provider selection
portunix pft configure --provider clearflask
portunix pft configure --provider fider  # existing

# Provider-specific deployment (optional)
portunix pft deploy --provider clearflask
portunix pft deploy --provider fider  # existing

# All other commands work unchanged (sync, list, show, etc.)
portunix pft sync  # uses configured provider
```

## Configuration File Format (Extended)

```json
{
  "name": "Project Name",
  "voc": {
    "path": "./voc",
    "provider": "clearflask",
    "endpoint": "https://clearflask.example.com",
    "api_token": "${PFT_VOC_TOKEN}",
    "project_id": "my-project",
    "visibility": "public"
  },
  "vos": {
    "path": "./vos",
    "provider": "fider",
    "endpoint": "http://localhost:3000",
    "api_token": "${PFT_VOS_TOKEN}",
    "visibility": "internal"
  }
}
```

Note: VoC and VoS can use different providers.

## Implementation Phases

### Phase 1: ClearFlask API Client
- [ ] Research ClearFlask API documentation
- [ ] Implement REST API client
- [ ] Authentication handling
- [ ] Basic CRUD operations

### Phase 2: Provider Implementation
- [ ] Implement `FeedbackProvider` interface
- [ ] Data mapping (ClearFlask → FeedbackItem)
- [ ] Status mapping configuration
- [ ] Multi-project support

### Phase 3: Container Deployment
- [ ] Create `assets/packages/clearflask.json` package definition
- [ ] Docker Compose template with all dependencies (MySQL, Elasticsearch, MinIO, ClearFlask)
- [ ] Health check implementation for all services
- [ ] `pft deploy --provider clearflask` command
- [ ] Resource requirements documentation (~2GB RAM minimum)

### Phase 4: Testing & Documentation
- [ ] Unit tests for API client
- [ ] Integration tests with mock ClearFlask
- [ ] Provider comparison documentation
- [ ] Migration guide (Fider ↔ ClearFlask)

## ClearFlask Docker Compose Template (Draft)

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${CLEARFLASK_DB_PASSWORD}
      MYSQL_DATABASE: clearflask
    volumes:
      - clearflask-mysql:/var/lib/mysql
    restart: unless-stopped

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.0
    environment:
      - discovery.type=single-node
      - ES_JAVA_OPTS=-Xms512m -Xmx512m
    volumes:
      - clearflask-es:/usr/share/elasticsearch/data
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${CLEARFLASK_S3_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${CLEARFLASK_S3_SECRET_KEY}
    volumes:
      - clearflask-minio:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    restart: unless-stopped

  clearflask:
    image: clearflask/clearflask:latest
    ports:
      - "3100:8080"
    environment:
      DATABASE_URL: jdbc:mysql://mysql:3306/clearflask
      ELASTICSEARCH_HOST: elasticsearch
      S3_ENDPOINT: http://minio:9000
      # ... additional config
    depends_on:
      - mysql
      - elasticsearch
      - minio
    restart: unless-stopped

volumes:
  clearflask-mysql:
  clearflask-es:
  clearflask-minio:
```

## Comparison: Fider vs ClearFlask

| Feature | Fider | ClearFlask |
|---------|-------|------------|
| **Simplicity** | Simple, lightweight | Complex, feature-rich |
| **Deployment** | Easy (2 containers) | Complex (4+ containers) |
| **Memory** | ~200MB | ~2GB+ |
| **Multi-project** | No | Yes |
| **Roadmap** | Basic | Advanced |
| **Voting** | Simple | Advanced (priority, etc.) |
| **SSO** | OAuth | OAuth, SAML |
| **Self-hosted** | Easy | Moderate |
| **Best for** | Single product, simple feedback | Multi-product, enterprise |

## Success Criteria

- [ ] ClearFlask provider passes all `FeedbackProvider` interface tests
- [ ] Existing pft commands work with ClearFlask provider
- [ ] Sync engine works correctly with ClearFlask
- [ ] VoC/VoS can use mixed providers (Fider + ClearFlask)
- [ ] Documentation explains when to choose which provider

## Dependencies

- **#107**: PTX-PFT base implementation (must be complete)
- **ptx-container**: For optional local deployment

## Notes

- ClearFlask is more suitable for enterprise use cases with multiple products
- Fider remains the recommended choice for simple, single-product feedback
- Provider selection should be guided by use case, not enforced
- Consider AWS-hosted ClearFlask as alternative to self-hosting

## Decisions

1. **Deployment scope**: ✅ Full local deployment (all components in Docker Compose)
2. **API stability**: ❓ To be evaluated during Phase 1 implementation
3. **Migration**: ❌ No migration tools needed (Fider ↔ ClearFlask)

---

## Implementation Notes

**Implemented**: 2025-12-24

### What was implemented:
- Phase 1: ClearFlask REST API client (`clearflask_api.go`)
- Phase 2: FeedbackProvider implementation (`clearflask_provider.go`)
- Phase 3: Container deployment (`clearflask_deploy.go`, `clearflask.json`)
- Kernel compatibility warning for cgroups v2 issue

### Known Issues:
- **Local deployment fails on Linux kernels 6.12+** due to ClearFlask using JDK 11 which has cgroups v2 bug (NullPointerException at CgroupV2Subsystem.getInstance)
- This is an upstream issue requiring ClearFlask to upgrade to JDK 17+
- Workaround: Use external ClearFlask instance or use Fider instead
- Windows/macOS deployment should work (WSL2/Hyper-V use older kernels)

### Not implemented (Phase 4):
- Unit tests for API client
- Integration tests with mock ClearFlask
- Migration guide (Fider ↔ ClearFlask)

---

**Created**: 2025-12-23
**Last Updated**: 2025-12-24
