# Issue #128: Docusaurus Container Performance Optimization

## Summary

Optimize Docker container workflow for Docusaurus to reduce build/serve time from ~20 minutes to acceptable levels. Current implementation runs full create/install/build/serve cycle every time, which is inefficient.

## Problem Statement

Current playbook workflow for Docusaurus:

1. Creates container with `node:22` image
2. Runs `npx create-docusaurus@latest ./site classic --typescript --skip-install && cd site && npm install`
3. Runs `cd site && npm run build`
4. Runs `cd site && npm run serve -- --host 0.0.0.0`

**Issues:**

- Steps 2+3 take ~20 minutes combined
- `node_modules` on bind mount is extremely slow (thousands of small files)
- No caching of npm packages between runs
- `create-docusaurus` runs every time (should be one-time initialization)
- Using `build + serve` instead of `start` for development

## Root Causes

### 1. Bind Mount Performance

```yaml
volumes:
  - "./:/workspace"
```
When `node_modules` is on bind mount, Docker has to sync thousands of small files between host and container, causing massive slowdown.

### 2. No npm Cache

Each container run downloads all packages fresh - no persistent cache.

### 3. Wrong Dev Workflow

Using `npm run build && npm run serve` for development instead of `npm run start` (which has hot reload).

### 4. Repeated Initialization

`npx create-docusaurus` runs every time instead of being a one-time setup.

## Proposed Solutions

### Solution 1: Separate node_modules into Named Volume

```yaml
services:
  docs:
    image: node:22
    working_dir: /workspace/site
    ports:
      - "3000:3000"
    volumes:
      - ./site:/workspace/site
      - docs_node_modules:/workspace/site/node_modules
      - docs_npm_cache:/root/.npm
    command: sh -lc "npm install && npm run start -- --host 0.0.0.0"

volumes:
  docs_node_modules:
  docs_npm_cache:
```

**Benefits:**

- Source files editable on host
- `node_modules` stays in Docker volume (fast)
- npm cache persists between runs

### Solution 2: Split Init vs Run Scripts

Instead of one `init` script that does everything:

```yaml
scripts:
  # One-time project creation (run manually once)
  create: "npx create-docusaurus@latest ./site classic --typescript"

  # Regular development (fast, hot reload)
  dev: "cd site && npm install && npm run start -- --host 0.0.0.0"

  # Production build (CI/release only)
  build: "cd site && npm ci && npm run build"

  # Serve production build
  serve: "cd site && npm run serve -- --host 0.0.0.0"
```

### Solution 3: Multi-stage Dockerfile for Production

```dockerfile
# syntax=docker/dockerfile:1.6

FROM node:22 AS deps
WORKDIR /app
COPY site/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

FROM node:22 AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY site/ ./
RUN npm run build

FROM nginx:alpine AS serve
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
```

### Solution 4: Playbook Engine Enhancement

Enhance `portunix playbook` to support:

1. **Persistent volumes** for `node_modules`:

   ```yaml
   spec:
     environment:
       volumes:
         - "./site:/workspace/site"
         - "node_modules:/workspace/site/node_modules:named"
         - "npm_cache:/root/.npm:named"
   ```

2. **Script dependencies** - only run `init` if project doesn't exist:

   ```yaml
   scripts:
     init:
       command: "npx create-docusaurus@latest ./site classic"
       condition: "! -d ./site"
   ```

3. **Dev vs Prod modes**:

   ```yaml
   modes:
     dev:
       script: dev
       ports: ["3000:3000"]
     prod:
       script: build
       dockerfile: Dockerfile.prod
   ```

4. **Script selection parameter** - specify which scripts to run:

   ```bash
   # Run only init script
   portunix playbook run my-docs.ptxbook --script init

   # Run only dev script (start server)
   portunix playbook run my-docs.ptxbook --script dev

   # Run multiple scripts in order
   portunix playbook run my-docs.ptxbook --script init,build

   # Run all scripts (current default behavior)
   portunix playbook run my-docs.ptxbook --script all

   # List available scripts
   portunix playbook run my-docs.ptxbook --list-scripts
   ```

   This allows users to:
   - Run one-time initialization separately
   - Start dev server without rebuilding
   - Run production build only when needed
   - Chain specific scripts as needed

## Implementation Steps

### Phase 1: Quick Wins (Playbook Changes)

- [ ] Change default script from `build+serve` to `start` for dev
- [ ] Add `--host 0.0.0.0` to start command
- [ ] Document that `create` is one-time operation

### Phase 2: Volume Optimization

- [ ] Add support for named volumes in playbook engine
- [ ] Auto-create `node_modules` and `npm_cache` volumes
- [ ] Update Docusaurus template with optimized volumes

### Phase 3: Script Conditions

- [ ] Add conditional script execution (`condition` field)
- [ ] Skip `create` if project already exists
- [ ] Add `--force` flag to override conditions

### Phase 4: Multi-stage Build Support

- [ ] Generate Dockerfile for production builds
- [ ] Add `portunix playbook build` for production images
- [ ] Support caching layers

## Expected Results

| Metric | Before | After |
|--------|--------|-------|
| First run (create + install) | ~10 min | ~10 min (one-time) |
| Subsequent dev start | ~15 min | ~30 sec |
| Production build | ~10 min | ~2-3 min (with cache) |
| Hot reload iteration | N/A | Instant |

## Related Issues

- #119 - PTX-Ansible Standalone Help and Template Examples System
- #102 - Compose Command Implementation

## Priority

High

## Labels

enhancement, container, docker, performance, docusaurus, playbook, developer-experience

---

**Created**: 2026-01-04
**Status**: 📋 Open
