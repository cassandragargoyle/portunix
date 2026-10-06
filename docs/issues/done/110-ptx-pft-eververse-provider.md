# Issue #110: PTX-PFT Eververse Provider Implementation

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Labels**: enhancement, helper-binary, product-feedback, eververse, provider, high-complexity
**Depends on**: #107 (PTX-PFT base implementation)
**Branch**: `feature/110-eververse-provider` (merged to main 2025-12-24)
**Testing**: ⏸️ Blocked by Docker Hub rate limit (implementation verified, full stack test pending)

---

## Summary

Implement Eververse provider for ptx-pft helper, enabling integration with Eververse as an alternative product management platform. Eververse is an open-source alternative to Productboard and Cycle with AI-powered features for problem exploration, solution ideation, feature prioritization, and roadmap planning.

**WARNING**: This is a high-complexity implementation. Eververse requires Supabase backend which consists of 12+ containers. Total infrastructure requirements are significantly higher than ClearFlask or Fider.

## Eververse Overview

**Repository**: https://github.com/haydenbleasel/eververse
**License**: Open Source
**Technology Stack**:
- Frontend: Next.js, React, TypeScript (98.8%)
- Backend: Supabase (PostgreSQL + Auth + Realtime + Storage)
- Build: Turborepo monorepo
- UI: shadcn/ui, TipTap editor, Excalidraw canvas

**Features**:
- AI-powered problem exploration and solution ideation
- Feature prioritization and roadmap planning
- Feedback collection and user engagement
- Canvas-based brainstorming (Excalidraw)
- Rich text editing (TipTap/Novel)

## Architecture Consideration

### Deployment Complexity Comparison

| Component | Fider | ClearFlask | Eververse |
|-----------|-------|------------|-----------|
| Backend | Go (single binary) | Java (Tomcat) | Next.js + Supabase |
| Database | PostgreSQL | MySQL + ES + DynamoDB | PostgreSQL (via Supabase) |
| Containers | 2 | 4-5 | **12-15** |
| Storage | - | MinIO/S3 | Supabase Storage |
| Memory | ~200MB | ~2GB | **~4-6GB** |
| Startup | ~5s | ~30s+ | **~60s+** |
| Complexity | Low | Medium | **High** |

### Supabase Self-Hosted Stack

Eververse requires full Supabase stack which includes:

| # | Service | Purpose | RAM |
|---|---------|---------|-----|
| 1 | PostgreSQL | Core database | 256MB |
| 2 | Kong | API Gateway | 128MB |
| 3 | GoTrue (Auth) | Authentication | 128MB |
| 4 | PostgREST | REST API for Postgres | 64MB |
| 5 | Realtime | WebSocket server | 128MB |
| 6 | Storage API | File storage | 64MB |
| 7 | imgproxy | Image processing | 128MB |
| 8 | postgres-meta | Postgres management API | 64MB |
| 9 | Edge Runtime | Deno-based functions | 256MB |
| 10 | Logflare | Log management | 128MB |
| 11 | Supavisor | Connection pooler | 64MB |
| 12 | Vector | Log pipeline | 128MB |
| 13 | Studio | Dashboard UI | 256MB |
| 14 | Eververse App | Next.js frontend | 512MB |
| 15 | Eververse API | Next.js API routes | 256MB |

**Total estimated RAM: 4-6GB minimum**

### Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| High resource consumption | High | Document requirements, provide lightweight alternative |
| Vercel-specific features may fail | Medium | Test edge functions compatibility |
| Complex debugging | High | Comprehensive health checks, logging |
| Supabase version compatibility | Medium | Pin specific versions |
| Long startup time | Medium | Implement proper health checks with timeouts |
| Next.js in container issues | Medium | Use production build, optimize |

## Provider Abstraction (Existing)

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

Eververse provider will implement this interface, reusing existing sync engine.

## Requirements

### Functional Requirements

1. **Eververse Provider**
   - Implement `FeedbackProvider` interface for Eververse/Supabase API
   - Support all CRUD operations (features, feedback, roadmap items)
   - Handle Eververse-specific features (AI suggestions, canvas, priorities)

2. **Configuration**
   - Provider selection in `.pft-config.json`
   - Supabase connection parameters
   - Project/workspace selection

3. **Data Mapping**
   - Map Eververse data structures to `FeedbackItem`
   - Handle AI-generated content
   - Preserve Eververse-specific metadata (priorities, tags, roadmap links)

4. **Container Deployment**
   - Full Docker Compose template with Supabase + Eververse
   - `pft deploy --provider eververse`
   - Health checks for all 15 services
   - Proper startup ordering with dependencies

### Technical Requirements

1. **Supabase Client**
   - Use official Supabase Go client or REST API
   - Handle real-time subscriptions (optional)
   - Manage authentication tokens

2. **Next.js Container**
   - Production build of Eververse
   - Environment variable configuration
   - Proper Node.js runtime in container

## Command Changes

```bash
# Configuration with provider selection
portunix pft configure --provider eververse
portunix pft configure --provider clearflask
portunix pft configure --provider fider

# Provider-specific deployment
portunix pft deploy --provider eververse  # WARNING: High resource usage
portunix pft deploy --provider clearflask
portunix pft deploy --provider fider

# Resource check before deployment
portunix pft deploy --provider eververse --check-resources
# Output: "Eververse requires ~6GB RAM. Available: 8GB. Proceed? [y/N]"

# All other commands work unchanged
portunix pft sync
portunix pft list
```

## Configuration File Format (Extended)

```json
{
  "name": "Project Name",
  "voc": {
    "path": "./voc",
    "provider": "eververse",
    "supabase_url": "http://localhost:8000",
    "supabase_anon_key": "${PFT_SUPABASE_ANON_KEY}",
    "supabase_service_key": "${PFT_SUPABASE_SERVICE_KEY}",
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

## Implementation Phases

### Phase 1: Supabase Integration Research ✅
- [x] Study Supabase self-hosted Docker setup
- [x] Test Supabase Go client library
- [x] Verify Supabase API compatibility
- [x] Document minimum resource requirements

### Phase 2: Eververse Build & Containerization ✅
- [x] Clone and analyze Eververse repository
- [x] Create production build process
- [x] Build Eververse Docker image
- [x] Configure environment variables mapping
- [x] Test Next.js in container environment

### Phase 3: Docker Compose Infrastructure ✅
- [x] Create `assets/packages/eververse.json` package definition
- [x] Create full Docker Compose with all 12 services
- [x] Implement startup ordering and dependencies
- [x] Add health checks for each service
- [x] Add pullPolicy support for local images
- [ ] Test complete stack startup (⏸️ blocked by Docker Hub rate limit)

### Phase 4: Provider Implementation ✅
- [x] Implement `FeedbackProvider` interface
- [x] Create Supabase REST API client with JWT auth
- [x] Data mapping (Eververse → FeedbackItem)
- [x] Handle Eververse-specific features

### Phase 5: Testing & Documentation
- [ ] Integration tests with local Eververse (⏸️ blocked by Docker Hub rate limit)
- [x] Resource usage documentation
- [x] Troubleshooting guide
- [x] Provider comparison documentation

## Docker Compose Template (Draft)

```yaml
version: '3.8'

# Supabase Stack
services:
  # ==================== DATABASE ====================
  db:
    image: supabase/postgres:15.1.0.147
    container_name: eververse-db
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-your-super-secret-password}
      POSTGRES_DB: postgres
    volumes:
      - eververse-db:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ==================== API GATEWAY ====================
  kong:
    image: kong:2.8.1
    container_name: eververse-kong
    restart: unless-stopped
    ports:
      - "8000:8000"
      - "8443:8443"
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /var/lib/kong/kong.yml
      KONG_DNS_ORDER: LAST,A,CNAME
      KONG_PLUGINS: request-transformer,cors,key-auth,acl
    volumes:
      - ./volumes/kong/kong.yml:/var/lib/kong/kong.yml:ro
    depends_on:
      db:
        condition: service_healthy

  # ==================== AUTH ====================
  auth:
    image: supabase/gotrue:v2.99.0
    container_name: eververse-auth
    restart: unless-stopped
    environment:
      GOTRUE_API_HOST: 0.0.0.0
      GOTRUE_API_PORT: 9999
      API_EXTERNAL_URL: ${API_EXTERNAL_URL:-http://localhost:8000}
      GOTRUE_DB_DRIVER: postgres
      GOTRUE_DB_DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres?search_path=auth
      GOTRUE_SITE_URL: ${SITE_URL:-http://localhost:3000}
      GOTRUE_JWT_SECRET: ${JWT_SECRET:-your-super-secret-jwt-token}
      GOTRUE_JWT_EXP: 3600
      GOTRUE_DISABLE_SIGNUP: "false"
    depends_on:
      db:
        condition: service_healthy

  # ==================== REST API ====================
  rest:
    image: postgrest/postgrest:v11.2.0
    container_name: eververse-rest
    restart: unless-stopped
    environment:
      PGRST_DB_URI: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
      PGRST_DB_SCHEMAS: public,storage,graphql_public
      PGRST_DB_ANON_ROLE: anon
      PGRST_JWT_SECRET: ${JWT_SECRET:-your-super-secret-jwt-token}
    depends_on:
      db:
        condition: service_healthy

  # ==================== REALTIME ====================
  realtime:
    image: supabase/realtime:v2.25.35
    container_name: eververse-realtime
    restart: unless-stopped
    environment:
      PORT: 4000
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${POSTGRES_PASSWORD}
      DB_NAME: postgres
      DB_SSL: "false"
      JWT_SECRET: ${JWT_SECRET:-your-super-secret-jwt-token}
      REPLICATION_MODE: RLS
      SECURE_CHANNELS: "true"
    depends_on:
      db:
        condition: service_healthy

  # ==================== STORAGE ====================
  storage:
    image: supabase/storage-api:v0.43.11
    container_name: eververse-storage
    restart: unless-stopped
    environment:
      ANON_KEY: ${ANON_KEY}
      SERVICE_KEY: ${SERVICE_KEY}
      POSTGREST_URL: http://rest:3000
      PGRST_JWT_SECRET: ${JWT_SECRET:-your-super-secret-jwt-token}
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
      STORAGE_BACKEND: file
      FILE_STORAGE_BACKEND_PATH: /var/lib/storage
      TENANT_ID: stub
      REGION: stub
      GLOBAL_S3_BUCKET: stub
    volumes:
      - eververse-storage:/var/lib/storage
    depends_on:
      db:
        condition: service_healthy
      rest:
        condition: service_started

  # ==================== IMAGE PROXY ====================
  imgproxy:
    image: darthsim/imgproxy:v3.18
    container_name: eververse-imgproxy
    restart: unless-stopped
    environment:
      IMGPROXY_BIND: ":5001"
      IMGPROXY_LOCAL_FILESYSTEM_ROOT: /
      IMGPROXY_USE_ETAG: "true"

  # ==================== POSTGRES META ====================
  meta:
    image: supabase/postgres-meta:v0.68.0
    container_name: eververse-meta
    restart: unless-stopped
    environment:
      PG_META_PORT: 8080
      PG_META_DB_HOST: db
      PG_META_DB_PORT: 5432
      PG_META_DB_NAME: postgres
      PG_META_DB_USER: postgres
      PG_META_DB_PASSWORD: ${POSTGRES_PASSWORD}
    depends_on:
      db:
        condition: service_healthy

  # ==================== EDGE FUNCTIONS ====================
  functions:
    image: supabase/edge-runtime:v1.22.4
    container_name: eververse-functions
    restart: unless-stopped
    environment:
      JWT_SECRET: ${JWT_SECRET:-your-super-secret-jwt-token}
      SUPABASE_URL: http://kong:8000
      SUPABASE_ANON_KEY: ${ANON_KEY}
      SUPABASE_SERVICE_ROLE_KEY: ${SERVICE_KEY}
      SUPABASE_DB_URL: postgresql://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
    volumes:
      - ./volumes/functions:/home/deno/functions:Z
    depends_on:
      - kong

  # ==================== ANALYTICS ====================
  analytics:
    image: supabase/logflare:1.4.0
    container_name: eververse-analytics
    restart: unless-stopped
    environment:
      LOGFLARE_NODE_HOST: 127.0.0.1
      DB_USERNAME: postgres
      DB_DATABASE: postgres
      DB_HOSTNAME: db
      DB_PORT: 5432
      DB_PASSWORD: ${POSTGRES_PASSWORD}
      DB_SCHEMA: _analytics
      LOGFLARE_API_KEY: ${LOGFLARE_API_KEY:-your-super-secret-api-key}
      LOGFLARE_SINGLE_TENANT: "true"
      LOGFLARE_SUPABASE_MODE: "true"
      POSTGRES_BACKEND_URL: postgresql://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
    depends_on:
      db:
        condition: service_healthy

  # ==================== STUDIO (Dashboard) ====================
  studio:
    image: supabase/studio:20231123-64a766a
    container_name: eververse-studio
    restart: unless-stopped
    ports:
      - "3001:3000"
    environment:
      STUDIO_PG_META_URL: http://meta:8080
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      DEFAULT_ORGANIZATION_NAME: Eververse
      DEFAULT_PROJECT_NAME: Eververse Local
      SUPABASE_URL: http://kong:8000
      SUPABASE_PUBLIC_URL: ${API_EXTERNAL_URL:-http://localhost:8000}
      SUPABASE_ANON_KEY: ${ANON_KEY}
      SUPABASE_SERVICE_KEY: ${SERVICE_KEY}
    depends_on:
      - kong
      - meta

  # ==================== EVERVERSE APP ====================
  eververse:
    build:
      context: ./eververse
      dockerfile: Dockerfile
    container_name: eververse-app
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      NODE_ENV: production
      NEXT_PUBLIC_SUPABASE_URL: http://kong:8000
      NEXT_PUBLIC_SUPABASE_ANON_KEY: ${ANON_KEY}
      SUPABASE_SERVICE_ROLE_KEY: ${SERVICE_KEY}
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
      # Disable external services for local deployment
      NEXT_PUBLIC_DISABLE_ANALYTICS: "true"
      STRIPE_SECRET_KEY: ${STRIPE_SECRET_KEY:-sk_test_dummy}
      STRIPE_WEBHOOK_SECRET: ${STRIPE_WEBHOOK_SECRET:-whsec_dummy}
    depends_on:
      - kong
      - auth
      - rest
      - storage
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s

volumes:
  eververse-db:
  eververse-storage:

networks:
  default:
    name: eververse-network
```

## Eververse Dockerfile (Draft)

```dockerfile
FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install pnpm
RUN npm install -g pnpm

# Copy package files
COPY package.json pnpm-lock.yaml* ./
COPY apps/app/package.json ./apps/app/
COPY packages/ ./packages/

RUN pnpm install --frozen-lockfile

# Build stage
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED 1

RUN npm install -g pnpm
RUN pnpm build --filter=app

# Production stage
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/apps/app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/apps/app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT 3000

CMD ["node", "server.js"]
```

## Comparison: Fider vs ClearFlask vs Eververse

| Feature | Fider | ClearFlask | Eververse |
|---------|-------|------------|-----------|
| **Simplicity** | Simple | Complex | Very Complex |
| **Deployment** | Easy (2) | Moderate (4) | Hard (15) |
| **Memory** | ~200MB | ~2GB | **~6GB** |
| **Technology** | Go | Java | Next.js + Supabase |
| **AI Features** | No | No | **Yes** |
| **Roadmap** | Basic | Advanced | **AI-powered** |
| **Canvas/Whiteboard** | No | No | **Yes (Excalidraw)** |
| **Real-time** | No | No | **Yes** |
| **Self-hosted** | Easy | Moderate | **Difficult** |
| **Best for** | Simple feedback | Enterprise feedback | **AI-first product management** |

## Success Criteria

- [ ] Supabase stack starts successfully with all services healthy (⏸️ blocked by Docker Hub rate limit)
- [x] Eververse frontend builds and runs in container (verified: `portunix/eververse:latest`)
- [ ] Eververse connects to local Supabase instance (⏸️ pending stack startup)
- [x] Provider implements `FeedbackProvider` interface (`eververse_provider.go`)
- [x] Existing pft commands work with Eververse provider (`main.go` switch cases)
- [ ] Sync engine works correctly with Eververse (⏸️ pending integration test)
- [x] Documentation explains resource requirements and limitations
- [ ] Health checks pass for all 12 services within 120 seconds (⏸️ pending)

## Implementation Commits

| Commit | Description |
|--------|-------------|
| `fa58532` | feat(110): implement Eververse provider for ptx-pft |
| `8e54ff3` | feat(110): add pullPolicy support for container services |

## Known Challenges & Mitigations

### Challenge 1: Vercel-Specific Code
Eververse may use Vercel-specific features (Edge Runtime, Analytics).
**Mitigation**: Identify and mock/disable Vercel-specific code paths.

### Challenge 2: Stripe Integration
Payment features require Stripe.
**Mitigation**: Use test mode keys or mock Stripe endpoints.

### Challenge 3: Build Complexity
Turborepo monorepo with multiple apps.
**Mitigation**: Focus on `apps/app` only, create simplified build.

### Challenge 4: Environment Variables
Extensive environment configuration needed.
**Mitigation**: Provide `.env.example` with all required variables.

## Dependencies

- **#107**: PTX-PFT base implementation (must be complete)
- **ptx-container**: For container deployment
- **Node.js 20+**: For building Eververse
- **pnpm**: For package management

## Resource Requirements

| Environment | Minimum RAM | Recommended RAM | Disk Space |
|-------------|-------------|-----------------|------------|
| Development | 6GB | 8GB | 10GB |
| Production | 8GB | 16GB | 20GB |

## Alternative Approaches (If Full Self-Host Fails)

1. **Hybrid**: Use Supabase Cloud (free tier) + local Eververse container
2. **External only**: Connect to existing Eververse instance via API
3. **Simplified**: Fork Eververse, remove Supabase dependency, use SQLite

---

**Created**: 2025-12-24
**Last Updated**: 2025-12-24
