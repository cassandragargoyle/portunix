package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	eververseComposeFile = "docker-compose.yaml"
	eververseEnvFile     = ".env"
	eververseProjectName = "portunix-eververse"
	eververseImageName   = "portunix/eververse:latest"
	eververseGitRepo     = "https://github.com/haydenbleasel/eververse.git"

	// Minimum RAM requirement in bytes (6GB)
	eververseMinRAMBytes = 6 * 1024 * 1024 * 1024
)

// getEververseDeployDir returns the directory where Eververse compose files are stored
func getEververseDeployDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "eververse")
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create deploy directory: %w", err)
	}

	return deployDir, nil
}

// checkEververseResourceRequirements checks if system has enough resources for Eververse
func checkEververseResourceRequirements() (bool, string) {
	var memInfo struct {
		Total uint64
	}

	// Try to read memory info on Linux
	if runtime.GOOS == "linux" {
		file, err := os.Open("/proc/meminfo")
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
							memInfo.Total = val * 1024 // Convert from KB to bytes
						}
					}
					break
				}
			}
		}
	}

	// Check if we have enough RAM
	if memInfo.Total > 0 {
		totalGB := float64(memInfo.Total) / (1024 * 1024 * 1024)
		minGB := float64(eververseMinRAMBytes) / (1024 * 1024 * 1024)

		if memInfo.Total < eververseMinRAMBytes {
			return false, fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════════════════╗
║  ⚠️  WARNING: INSUFFICIENT MEMORY                                            ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  Eververse requires approximately %.0fGB RAM for the full Supabase stack.    ║
║  Your system has only %.1fGB available.                                      ║
║                                                                              ║
║  Running with insufficient memory may cause:                                 ║
║    - Services failing to start                                               ║
║    - OOM (Out of Memory) kills                                               ║
║    - System instability                                                      ║
║                                                                              ║
║  Recommendations:                                                            ║
║    1. Use Fider instead (requires ~200MB) - ./ptx-pft deploy fider           ║
║    2. Use ClearFlask (~2GB) - ./ptx-pft deploy clearflask                    ║
║    3. Connect to external Eververse instance                                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
`, minGB, totalGB)
		}

		// Warning if RAM is low but sufficient
		if totalGB < 8 {
			return true, fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════════════════╗
║  ⚠️  LOW MEMORY WARNING                                                      ║
╠══════════════════════════════════════════════════════════════════════════════╣
║  Your system has %.1fGB RAM. Eververse with Supabase stack works best with   ║
║  8GB or more. Performance may be degraded.                                   ║
╚══════════════════════════════════════════════════════════════════════════════╝
`, totalGB)
		}
	}

	return true, ""
}

// checkEververseImageExists checks if the Eververse Docker image exists locally
func checkEververseImageExists() bool {
	portunixPath, err := findPortunix()
	if err != nil {
		return false
	}

	cmd := exec.Command(portunixPath, "container", "image", "exists", eververseImageName)
	err = cmd.Run()
	return err == nil
}

// getEververseBuildDir returns the directory where Eververse source is cloned
func getEververseBuildDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	buildDir := filepath.Join(homeDir, ".portunix", "pft", "eververse-build")
	return buildDir, nil
}

// cloneEververseRepo clones the Eververse repository from GitHub
func cloneEververseRepo(buildDir string) error {
	// Check if already cloned
	gitDir := filepath.Join(buildDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		fmt.Println("  Updating existing Eververse repository...")
		cmd := exec.Command("git", "-C", buildDir, "pull", "--ff-only")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Clone fresh
	fmt.Println("  Cloning Eververse from GitHub...")
	if err := os.MkdirAll(filepath.Dir(buildDir), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", "--depth", "1", eververseGitRepo, buildDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// generateEververseDockerfile returns the Dockerfile content for building Eververse
func generateEververseDockerfile() string {
	return `# Eververse Dockerfile
# Generated by portunix pft deploy

FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat git
WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy package files
COPY package.json pnpm-lock.yaml* ./
COPY pnpm-workspace.yaml* ./

# Copy all package.json files for workspace
COPY apps/app/package.json ./apps/app/
COPY packages/*/package.json ./packages/

RUN pnpm install --frozen-lockfile || pnpm install

# Build stage
FROM base AS builder
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Set build-time environment variables
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Build the app
RUN pnpm build --filter=app || (echo "Build failed, trying alternative..." && cd apps/app && pnpm build)

# Production stage
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy built application
COPY --from=builder /app/apps/app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/apps/app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
`
}

// buildEververseImage builds the Eververse Docker image from source
func buildEververseImage(buildDir string) error {
	// Write Dockerfile
	dockerfilePath := filepath.Join(buildDir, "Dockerfile.portunix")
	dockerfile := generateEververseDockerfile()
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Println("  Building Eververse Docker image (this may take 5-10 minutes)...")

	portunixPath, err := findPortunix()
	if err != nil {
		return err
	}

	cmd := exec.Command(portunixPath, "container", "build",
		"-t", eververseImageName,
		"-f", dockerfilePath,
		buildDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = buildDir

	return cmd.Run()
}

// ensureEververseImage ensures the Eververse image exists, building if necessary
func ensureEververseImage() error {
	// Check if image already exists
	if checkEververseImageExists() {
		fmt.Println("Eververse Docker image found locally.")
		return nil
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  📦 BUILDING EVERVERSE IMAGE                                                 ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Eververse Docker image not found. Building from source...                  ║")
	fmt.Println("║  This is a one-time process and may take 5-10 minutes.                      ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  Requirements:                                                               ║")
	fmt.Println("║    - Git (for cloning repository)                                            ║")
	fmt.Println("║    - ~2GB disk space for build                                               ║")
	fmt.Println("║    - Internet connection                                                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	buildDir, err := getEververseBuildDir()
	if err != nil {
		return err
	}

	// Clone or update repository
	if err := cloneEververseRepo(buildDir); err != nil {
		return fmt.Errorf("failed to clone Eververse repository: %w", err)
	}

	// Build the image
	if err := buildEververseImage(buildDir); err != nil {
		return fmt.Errorf("failed to build Eververse image: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Eververse Docker image built successfully!")
	fmt.Println()

	return nil
}

// writeEververseComposeFile generates and writes docker-compose.yaml for Eververse
func writeEververseComposeFile(deployDir string) (string, error) {
	// Try to load from package definition first
	pkg, err := loadPackageDefinition("eververse")
	if err == nil {
		yamlData, err := generateComposeYAML(pkg)
		if err != nil {
			return "", fmt.Errorf("failed to generate compose YAML: %w", err)
		}

		composePath := filepath.Join(deployDir, eververseComposeFile)
		if err := os.WriteFile(composePath, yamlData, 0644); err != nil {
			return "", fmt.Errorf("failed to write compose file: %w", err)
		}

		return composePath, nil
	}

	// Fallback to embedded compose content if package not found
	composeContent := generateEververseComposeYAML()
	composePath := filepath.Join(deployDir, eververseComposeFile)
	if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	return composePath, nil
}

// generateEververseComposeYAML generates the full docker-compose.yaml for Eververse + Supabase
func generateEververseComposeYAML() string {
	return `# Eververse with Supabase Stack
# WARNING: This stack requires ~6GB RAM minimum
# Generated by portunix pft deploy

services:
  # ==================== DATABASE ====================
  db:
    image: supabase/postgres:15.1.0.147
    container_name: eververse-db
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
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
      GOTRUE_JWT_SECRET: ${JWT_SECRET}
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
      PGRST_JWT_SECRET: ${JWT_SECRET}
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
      JWT_SECRET: ${JWT_SECRET}
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
      PGRST_JWT_SECRET: ${JWT_SECRET}
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
      JWT_SECRET: ${JWT_SECRET}
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
      LOGFLARE_API_KEY: ${LOGFLARE_API_KEY}
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
    image: portunix/eververse:latest
    pull_policy: never
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
`
}

// writeEververseEnvFile writes environment variables for Eververse docker-compose
func writeEververseEnvFile(deployDir string, config *Config) (string, error) {
	envPath := filepath.Join(deployDir, eververseEnvFile)

	// Check if env file already exists (reuse secrets)
	existingEnv := make(map[string]string)
	if data, err := os.ReadFile(envPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				existingEnv[parts[0]] = parts[1]
			}
		}
	}

	// Generate or reuse secrets
	postgresPassword := existingEnv["POSTGRES_PASSWORD"]
	if postgresPassword == "" {
		postgresPassword = generateSecret(24)
	}

	jwtSecret := existingEnv["JWT_SECRET"]
	if jwtSecret == "" {
		jwtSecret = generateSecret(64)
	}

	anonKey := existingEnv["ANON_KEY"]
	if anonKey == "" {
		// Generate a JWT-like anon key (simplified)
		anonKey = generateSecret(64)
	}

	serviceKey := existingEnv["SERVICE_KEY"]
	if serviceKey == "" {
		serviceKey = generateSecret(64)
	}

	logflareKey := existingEnv["LOGFLARE_API_KEY"]
	if logflareKey == "" {
		logflareKey = generateSecret(32)
	}

	// Determine URLs
	siteURL := "http://localhost:3000"
	apiURL := "http://localhost:8000"

	if config.GetEndpoint() != "" {
		siteURL = config.GetEndpoint()
	}

	env := fmt.Sprintf(`# Eververse with Supabase environment configuration
# Generated by portunix pft deploy
# WARNING: Keep this file secure - it contains sensitive secrets

# ==================== DATABASE ====================
POSTGRES_PASSWORD=%s

# ==================== JWT & AUTH ====================
JWT_SECRET=%s
ANON_KEY=%s
SERVICE_KEY=%s

# ==================== URLS ====================
SITE_URL=%s
API_EXTERNAL_URL=%s

# ==================== LOGGING ====================
LOGFLARE_API_KEY=%s

# ==================== STRIPE (optional, dummy for local) ====================
STRIPE_SECRET_KEY=sk_test_dummy
STRIPE_WEBHOOK_SECRET=whsec_dummy
`, postgresPassword, jwtSecret, anonKey, serviceKey, siteURL, apiURL, logflareKey)

	if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
		return "", fmt.Errorf("failed to write env file: %w", err)
	}

	return envPath, nil
}

// createKongConfig creates the Kong API Gateway configuration
func createKongConfig(deployDir string) error {
	kongDir := filepath.Join(deployDir, "volumes", "kong")
	if err := os.MkdirAll(kongDir, 0755); err != nil {
		return err
	}

	kongConfig := `_format_version: "2.1"

services:
  - name: auth-v1
    url: http://auth:9999/
    routes:
      - name: auth-v1-routes
        paths:
          - /auth/v1/
        strip_path: true

  - name: rest-v1
    url: http://rest:3000/
    routes:
      - name: rest-v1-routes
        paths:
          - /rest/v1/
        strip_path: true

  - name: realtime-v1
    url: http://realtime:4000/
    routes:
      - name: realtime-v1-routes
        paths:
          - /realtime/v1/
        strip_path: true

  - name: storage-v1
    url: http://storage:5000/
    routes:
      - name: storage-v1-routes
        paths:
          - /storage/v1/
        strip_path: true

  - name: functions-v1
    url: http://functions:8000/
    routes:
      - name: functions-v1-routes
        paths:
          - /functions/v1/
        strip_path: true

  - name: meta
    url: http://meta:8080/
    routes:
      - name: meta-routes
        paths:
          - /pg/
        strip_path: true

plugins:
  - name: cors
    config:
      origins:
        - "*"
      methods:
        - GET
        - POST
        - PUT
        - PATCH
        - DELETE
        - OPTIONS
      headers:
        - Accept
        - Accept-Version
        - Authorization
        - Content-Length
        - Content-Type
        - apikey
        - x-client-info
      exposed_headers:
        - X-Total-Count
      credentials: true
      max_age: 3600
`

	kongPath := filepath.Join(kongDir, "kong.yml")
	return os.WriteFile(kongPath, []byte(kongConfig), 0644)
}

// createFunctionsDir creates the functions directory for edge functions
func createFunctionsDir(deployDir string) error {
	functionsDir := filepath.Join(deployDir, "volumes", "functions")
	return os.MkdirAll(functionsDir, 0755)
}

// runEververseContainerCompose executes portunix container compose command for Eververse
func runEververseContainerCompose(deployDir string, args ...string) error {
	portunixPath, err := findPortunix()
	if err != nil {
		return err
	}

	fullArgs := []string{
		"container", "compose",
		"-f", filepath.Join(deployDir, eververseComposeFile),
		"--env-file", filepath.Join(deployDir, eververseEnvFile),
		"-p", eververseProjectName,
	}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command(portunixPath, fullArgs...)
	cmd.Dir = deployDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// DeployEververse deploys Eververse using Docker Compose with Supabase
func DeployEververse(config *Config) (*DeployResult, error) {
	result := &DeployResult{}

	// Check resource requirements
	sufficient, warning := checkEververseResourceRequirements()
	if warning != "" {
		fmt.Println(warning)
	}
	if !sufficient {
		fmt.Print("Do you want to continue anyway? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return nil, fmt.Errorf("deployment cancelled due to insufficient resources")
		}
	}

	// Get deploy directory
	deployDir, err := getEververseDeployDir()
	if err != nil {
		return nil, err
	}

	// Create Kong configuration
	fmt.Println("Creating Kong API Gateway configuration...")
	if err := createKongConfig(deployDir); err != nil {
		return nil, fmt.Errorf("failed to create Kong config: %w", err)
	}

	// Create functions directory
	if err := createFunctionsDir(deployDir); err != nil {
		return nil, fmt.Errorf("failed to create functions directory: %w", err)
	}

	// Write compose file
	composePath, err := writeEververseComposeFile(deployDir)
	if err != nil {
		return nil, err
	}
	result.ComposeFile = composePath

	// Write env file
	envPath, err := writeEververseEnvFile(deployDir, config)
	if err != nil {
		return nil, err
	}
	result.EnvFile = envPath

	fmt.Println()
	fmt.Println("Starting Eververse deployment with Supabase stack...")
	fmt.Printf("  Compose file: %s\n", composePath)
	fmt.Printf("  Environment: %s\n", envPath)
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  ⚠️  WARNING: HIGH COMPLEXITY DEPLOYMENT                                     ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Eververse requires the full Supabase stack (~12 containers).               ║")
	fmt.Println("║  This deployment:                                                           ║")
	fmt.Println("║    - Requires ~6GB RAM minimum                                              ║")
	fmt.Println("║    - Takes 60-120 seconds to fully start                                    ║")
	fmt.Println("║    - Uses ~10GB disk space for images and data                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Ensure Eververse image exists (build from source if needed)
	if err := ensureEververseImage(); err != nil {
		return nil, fmt.Errorf("failed to prepare Eververse image: %w", err)
	}

	// Pull images
	fmt.Println("Pulling container images (this may take several minutes)...")
	if err := runEververseContainerCompose(deployDir, "pull"); err != nil {
		return nil, fmt.Errorf("failed to pull images: %w", err)
	}

	// Start services
	fmt.Println()
	fmt.Println("Starting services...")
	if err := runEververseContainerCompose(deployDir, "up", "-d"); err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	// Determine URL
	baseURL := config.GetEndpoint()
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	result.Success = true
	result.URL = baseURL
	result.Message = fmt.Sprintf(`Eververse deployed successfully!

Access Eververse at:     %s
Supabase Studio at:      http://localhost:3001
Supabase API at:         http://localhost:8000

Note: Full startup may take 60-120 seconds for all services to initialize.
      Database initialization happens on first start.

Services (12 containers):
  - PostgreSQL:          Database
  - Kong:                API Gateway
  - GoTrue:              Authentication
  - PostgREST:           REST API
  - Realtime:            WebSocket server
  - Storage:             File storage
  - imgproxy:            Image processing
  - postgres-meta:       Postgres management
  - Edge Functions:      Serverless functions
  - Logflare:            Analytics
  - Studio:              Dashboard UI
  - Eververse:           Main application

Resource usage: ~6GB RAM minimum`, baseURL)

	return result, nil
}

// GetEververseStatus returns the status of Eververse deployment
func GetEververseStatus() (string, error) {
	deployDir, err := getEververseDeployDir()
	if err != nil {
		return "", err
	}

	// Check if compose file exists
	composePath := filepath.Join(deployDir, eververseComposeFile)
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return "not_deployed", nil
	}

	// Check container status using portunix container compose
	portunixPath, err := findPortunix()
	if err != nil {
		return "unknown", err
	}

	cmd := exec.Command(portunixPath,
		"container", "compose",
		"-f", composePath,
		"-p", eververseProjectName,
		"ps", "--format", "{{.State}}",
	)
	output, err := cmd.Output()
	if err != nil {
		return "error", nil
	}

	states := strings.TrimSpace(string(output))
	if states == "" {
		return "stopped", nil
	}

	// Count running vs total
	lines := strings.Split(states, "\n")
	running := 0
	for _, state := range lines {
		if state == "running" {
			running++
		}
	}

	if running == 0 {
		return "stopped", nil
	}
	if running < len(lines) {
		return "partial", nil
	}
	return "running", nil
}

// GetEververseContainerInfo returns detailed container information for Eververse
func GetEververseContainerInfo() (string, error) {
	deployDir, err := getEververseDeployDir()
	if err != nil {
		return "", err
	}

	composePath := filepath.Join(deployDir, eververseComposeFile)
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return "Eververse is not deployed.", nil
	}

	portunixPath, err := findPortunix()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(portunixPath,
		"container", "compose",
		"-f", composePath,
		"-p", eververseProjectName,
		"ps",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil
	}

	return string(output), nil
}

// DestroyEververse removes the Eververse deployment
func DestroyEververse(removeVolumes bool) error {
	deployDir, err := getEververseDeployDir()
	if err != nil {
		return err
	}

	composePath := filepath.Join(deployDir, eververseComposeFile)
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("Eververse is not deployed")
	}

	fmt.Println("Stopping and removing Eververse containers...")
	fmt.Println("(This includes all 12 Supabase stack containers)")

	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
		fmt.Println("  (including volumes - this will delete all data)")
	}

	if err := runEververseContainerCompose(deployDir, args...); err != nil {
		return fmt.Errorf("failed to destroy deployment: %w", err)
	}

	fmt.Println("Eververse deployment removed successfully.")

	// Optionally remove deploy directory
	if removeVolumes {
		if err := os.RemoveAll(deployDir); err != nil {
			fmt.Printf("Warning: could not remove deploy directory: %v\n", err)
		}
	}

	return nil
}

// DeployEververseInstance deploys a named Eververse instance on specified port
func DeployEververseInstance(instanceName string, port int, config *Config) (*DeployResult, error) {
	result := &DeployResult{}

	// Check resource requirements
	sufficient, warning := checkEververseResourceRequirements()
	if warning != "" {
		fmt.Println(warning)
	}
	if !sufficient {
		fmt.Print("Do you want to continue anyway? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return nil, fmt.Errorf("deployment cancelled due to insufficient resources")
		}
	}

	// Get instance-specific deploy directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "eververse-"+instanceName)
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create deploy directory: %w", err)
	}

	projectName := fmt.Sprintf("portunix-eververse-%s", instanceName)

	// Create Kong configuration
	if err := createKongConfig(deployDir); err != nil {
		return nil, fmt.Errorf("failed to create Kong config: %w", err)
	}

	// Create functions directory
	if err := createFunctionsDir(deployDir); err != nil {
		return nil, fmt.Errorf("failed to create functions directory: %w", err)
	}

	// Generate compose file with custom ports
	composeContent := generateEververseInstanceComposeYAML(instanceName, port)
	composePath := filepath.Join(deployDir, eververseComposeFile)
	if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write compose file: %w", err)
	}
	result.ComposeFile = composePath

	// Write env file with instance-specific settings
	envPath, err := writeEververseInstanceEnvFile(deployDir, instanceName, port, config)
	if err != nil {
		return nil, err
	}
	result.EnvFile = envPath

	fmt.Printf("Starting Eververse (%s) deployment on port %d...\n", instanceName, port)
	fmt.Println("Note: Eververse requires ~6GB RAM and may take 60-120 seconds to start.")
	fmt.Println()

	// Ensure Eververse image exists (build from source if needed)
	if err := ensureEververseImage(); err != nil {
		return nil, fmt.Errorf("failed to prepare Eververse image: %w", err)
	}

	// Pull images
	fmt.Println("Pulling container images...")
	if err := runEververseInstanceContainerCompose(deployDir, projectName, "pull"); err != nil {
		return nil, fmt.Errorf("failed to pull images: %w", err)
	}

	// Start services
	fmt.Println()
	fmt.Println("Starting services...")
	if err := runEververseInstanceContainerCompose(deployDir, projectName, "up", "-d"); err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	studioPort := port + 1
	apiPort := port + 10

	result.Success = true
	result.URL = baseURL
	result.Message = fmt.Sprintf("Eververse (%s) deployed on port %d\nSupabase Studio: http://localhost:%d\nSupabase API: http://localhost:%d",
		instanceName, port, studioPort, apiPort)

	return result, nil
}

// generateEververseInstanceComposeYAML generates docker-compose.yaml for a specific instance
func generateEververseInstanceComposeYAML(instanceName string, port int) string {
	studioPort := port + 1
	apiPort := port + 10
	dbPort := port + 20

	return fmt.Sprintf(`# Eververse instance: %s
# Ports: App=%d, Studio=%d, API=%d, DB=%d

services:
  db:
    image: supabase/postgres:15.1.0.147
    container_name: eververse-%s-db
    restart: unless-stopped
    ports:
      - "%d:5432"
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: postgres
    volumes:
      - eververse-%s-db:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  kong:
    image: kong:2.8.1
    container_name: eververse-%s-kong
    restart: unless-stopped
    ports:
      - "%d:8000"
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

  auth:
    image: supabase/gotrue:v2.99.0
    container_name: eververse-%s-auth
    restart: unless-stopped
    environment:
      GOTRUE_API_HOST: 0.0.0.0
      GOTRUE_API_PORT: 9999
      API_EXTERNAL_URL: http://localhost:%d
      GOTRUE_DB_DRIVER: postgres
      GOTRUE_DB_DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres?search_path=auth
      GOTRUE_SITE_URL: http://localhost:%d
      GOTRUE_JWT_SECRET: ${JWT_SECRET}
      GOTRUE_JWT_EXP: 3600
      GOTRUE_DISABLE_SIGNUP: "false"
    depends_on:
      db:
        condition: service_healthy

  rest:
    image: postgrest/postgrest:v11.2.0
    container_name: eververse-%s-rest
    restart: unless-stopped
    environment:
      PGRST_DB_URI: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
      PGRST_DB_SCHEMAS: public,storage,graphql_public
      PGRST_DB_ANON_ROLE: anon
      PGRST_JWT_SECRET: ${JWT_SECRET}
    depends_on:
      db:
        condition: service_healthy

  realtime:
    image: supabase/realtime:v2.25.35
    container_name: eververse-%s-realtime
    restart: unless-stopped
    environment:
      PORT: 4000
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${POSTGRES_PASSWORD}
      DB_NAME: postgres
      DB_SSL: "false"
      JWT_SECRET: ${JWT_SECRET}
      REPLICATION_MODE: RLS
      SECURE_CHANNELS: "true"
    depends_on:
      db:
        condition: service_healthy

  storage:
    image: supabase/storage-api:v0.43.11
    container_name: eververse-%s-storage
    restart: unless-stopped
    environment:
      ANON_KEY: ${ANON_KEY}
      SERVICE_KEY: ${SERVICE_KEY}
      POSTGREST_URL: http://rest:3000
      PGRST_JWT_SECRET: ${JWT_SECRET}
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
      STORAGE_BACKEND: file
      FILE_STORAGE_BACKEND_PATH: /var/lib/storage
      TENANT_ID: stub
      REGION: stub
      GLOBAL_S3_BUCKET: stub
    volumes:
      - eververse-%s-storage:/var/lib/storage
    depends_on:
      db:
        condition: service_healthy
      rest:
        condition: service_started

  meta:
    image: supabase/postgres-meta:v0.68.0
    container_name: eververse-%s-meta
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

  studio:
    image: supabase/studio:20231123-64a766a
    container_name: eververse-%s-studio
    restart: unless-stopped
    ports:
      - "%d:3000"
    environment:
      STUDIO_PG_META_URL: http://meta:8080
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      DEFAULT_ORGANIZATION_NAME: Eververse
      DEFAULT_PROJECT_NAME: Eververse %s
      SUPABASE_URL: http://kong:8000
      SUPABASE_PUBLIC_URL: http://localhost:%d
      SUPABASE_ANON_KEY: ${ANON_KEY}
      SUPABASE_SERVICE_KEY: ${SERVICE_KEY}
    depends_on:
      - kong
      - meta

  eververse:
    image: portunix/eververse:latest
    pull_policy: never
    container_name: eververse-%s-app
    restart: unless-stopped
    ports:
      - "%d:3000"
    environment:
      NODE_ENV: production
      NEXT_PUBLIC_SUPABASE_URL: http://kong:8000
      NEXT_PUBLIC_SUPABASE_ANON_KEY: ${ANON_KEY}
      SUPABASE_SERVICE_ROLE_KEY: ${SERVICE_KEY}
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/postgres
      NEXT_PUBLIC_DISABLE_ANALYTICS: "true"
    depends_on:
      - kong
      - auth
      - rest
      - storage

volumes:
  eververse-%s-db:
  eververse-%s-storage:

networks:
  default:
    name: eververse-%s-network
`, instanceName, port, studioPort, apiPort, dbPort,
		instanceName, dbPort, instanceName,
		instanceName, apiPort,
		instanceName, apiPort, port,
		instanceName,
		instanceName,
		instanceName, instanceName,
		instanceName,
		instanceName, studioPort, instanceName, apiPort,
		instanceName, port,
		instanceName, instanceName, instanceName)
}

// writeEververseInstanceEnvFile writes environment file for a specific instance
func writeEververseInstanceEnvFile(deployDir, instanceName string, port int, config *Config) (string, error) {
	envPath := filepath.Join(deployDir, eververseEnvFile)

	// Check if env file already exists (reuse secrets)
	existingEnv := make(map[string]string)
	if data, err := os.ReadFile(envPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				existingEnv[parts[0]] = parts[1]
			}
		}
	}

	// Generate or reuse secrets
	postgresPassword := existingEnv["POSTGRES_PASSWORD"]
	if postgresPassword == "" {
		postgresPassword = generateSecret(24)
	}

	jwtSecret := existingEnv["JWT_SECRET"]
	if jwtSecret == "" {
		jwtSecret = generateSecret(64)
	}

	anonKey := existingEnv["ANON_KEY"]
	if anonKey == "" {
		anonKey = generateSecret(64)
	}

	serviceKey := existingEnv["SERVICE_KEY"]
	if serviceKey == "" {
		serviceKey = generateSecret(64)
	}

	siteURL := fmt.Sprintf("http://localhost:%d", port)
	apiURL := fmt.Sprintf("http://localhost:%d", port+10)

	env := fmt.Sprintf(`# Eververse %s environment configuration
# Generated by portunix pft deploy

POSTGRES_PASSWORD=%s
JWT_SECRET=%s
ANON_KEY=%s
SERVICE_KEY=%s
SITE_URL=%s
API_EXTERNAL_URL=%s
`, instanceName, postgresPassword, jwtSecret, anonKey, serviceKey, siteURL, apiURL)

	if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
		return "", fmt.Errorf("failed to write env file: %w", err)
	}

	return envPath, nil
}

// runEververseInstanceContainerCompose executes portunix container compose for a specific instance
func runEververseInstanceContainerCompose(deployDir, projectName string, args ...string) error {
	portunixPath, err := findPortunix()
	if err != nil {
		return err
	}

	fullArgs := []string{
		"container", "compose",
		"-f", filepath.Join(deployDir, eververseComposeFile),
		"--env-file", filepath.Join(deployDir, eververseEnvFile),
		"-p", projectName,
	}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command(portunixPath, fullArgs...)
	cmd.Dir = deployDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
