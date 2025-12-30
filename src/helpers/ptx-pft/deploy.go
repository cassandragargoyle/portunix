package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	fiderComposeFile = "docker-compose.yaml"
	fiderEnvFile     = ".env"
	fiderProjectName = "portunix-fider"
)

// Package JSON structures
type PackageDefinition struct {
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   PackageMetadata `json:"metadata"`
	Spec       PackageSpec     `json:"spec"`
}

type PackageMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type PackageSpec struct {
	Type      string         `json:"type"`
	Container ContainerSpec  `json:"container"`
	Defaults  map[string]string `json:"defaults"`
}

type ContainerSpec struct {
	Provider    string                    `json:"provider"`
	ProjectName string                    `json:"projectName"`
	Services    map[string]ServiceSpec    `json:"services"`
	Volumes     map[string]interface{}    `json:"volumes"`
	Networks    map[string]NetworkSpec    `json:"networks"`
}

type ServiceSpec struct {
	Image         string            `json:"image"`
	PullPolicy    string            `json:"pullPolicy"`
	ContainerName string            `json:"containerName"`
	Ports         []string          `json:"ports"`
	Environment   map[string]string `json:"environment"`
	Volumes       []string          `json:"volumes"`
	Healthcheck   *HealthcheckSpec  `json:"healthcheck"`
	DependsOn     map[string]DependsOnSpec `json:"dependsOn"`
	Restart       string            `json:"restart"`
}

type HealthcheckSpec struct {
	Test     []string `json:"test"`
	Interval string   `json:"interval"`
	Timeout  string   `json:"timeout"`
	Retries  int      `json:"retries"`
}

type DependsOnSpec struct {
	Condition string `json:"condition"`
}

type NetworkSpec struct {
	Name string `json:"name"`
}

// Docker Compose YAML structures
type ComposeFile struct {
	Version  string                     `yaml:"version,omitempty"`
	Services map[string]ComposeService  `yaml:"services"`
	Volumes  map[string]interface{}     `yaml:"volumes,omitempty"`
	Networks map[string]ComposeNetwork  `yaml:"networks,omitempty"`
}

type ComposeService struct {
	Image         string                 `yaml:"image"`
	PullPolicy    string                 `yaml:"pull_policy,omitempty"`
	ContainerName string                 `yaml:"container_name,omitempty"`
	Ports         []string               `yaml:"ports,omitempty"`
	Environment   map[string]string      `yaml:"environment,omitempty"`
	Volumes       []string               `yaml:"volumes,omitempty"`
	Healthcheck   *ComposeHealthcheck    `yaml:"healthcheck,omitempty"`
	DependsOn     interface{}            `yaml:"depends_on,omitempty"`
	Restart       string                 `yaml:"restart,omitempty"`
	Networks      []string               `yaml:"networks,omitempty"`
}

type ComposeHealthcheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

type ComposeNetwork struct {
	Name string `yaml:"name,omitempty"`
}

// DeployResult contains deployment information
type DeployResult struct {
	Success     bool
	URL         string
	ComposeFile string
	EnvFile     string
	Message     string
}

// ComposePreflightResult contains compose readiness check result
type ComposePreflightResult struct {
	Ready           bool
	ErrorMessage    string
	FixInstructions string
}

// CheckComposePreflight checks if compose is ready to use by calling portunix container compose-preflight
func CheckComposePreflight() (*ComposePreflightResult, error) {
	portunixPath, err := findPortunix()
	if err != nil {
		return nil, fmt.Errorf("failed to find portunix: %w", err)
	}

	cmd := exec.Command(portunixPath, "container", "compose-preflight", "--json")
	// Use Output() to get only stdout (JSON is written to stdout)
	output, _ := cmd.Output()

	// Parse JSON output
	var result struct {
		Ready       bool   `json:"ready"`
		Runtime     string `json:"runtime"`
		Version     string `json:"version"`
		DaemonReady bool   `json:"daemon_running"`
		Error       string `json:"error"`
		Fix         string `json:"fix"`
	}

	if jsonErr := json.Unmarshal(output, &result); jsonErr != nil {
		// If JSON parsing fails, return generic error
		return &ComposePreflightResult{
			Ready:           false,
			ErrorMessage:    "Could not determine compose status",
			FixInstructions: "Run 'portunix container compose-preflight' for details",
		}, nil
	}

	return &ComposePreflightResult{
		Ready:           result.Ready,
		ErrorMessage:    result.Error,
		FixInstructions: result.Fix,
	}, nil
}

// loadPackageDefinition loads a package definition from JSON file
func loadPackageDefinition(name string) (*PackageDefinition, error) {
	// Find package file - check multiple locations
	var packagePath string
	var data []byte
	var err error

	// Get executable directory
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	// Possible locations for package definitions
	locations := []string{
		filepath.Join(execDir, "assets", "packages", name+".json"),
		filepath.Join(execDir, "..", "assets", "packages", name+".json"),
		filepath.Join("assets", "packages", name+".json"),
	}

	for _, loc := range locations {
		if data, err = os.ReadFile(loc); err == nil {
			packagePath = loc
			break
		}
	}

	if packagePath == "" {
		return nil, fmt.Errorf("package definition not found: %s (searched: %v)", name, locations)
	}

	var pkg PackageDefinition
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package definition %s: %w", packagePath, err)
	}

	return &pkg, nil
}

// generateComposeYAML generates docker-compose.yaml from package definition
func generateComposeYAML(pkg *PackageDefinition) ([]byte, error) {
	compose := ComposeFile{
		Services: make(map[string]ComposeService),
		Volumes:  pkg.Spec.Container.Volumes,
		Networks: make(map[string]ComposeNetwork),
	}

	// Convert services
	for name, svc := range pkg.Spec.Container.Services {
		composeSvc := ComposeService{
			Image:         svc.Image,
			PullPolicy:    svc.PullPolicy,
			ContainerName: svc.ContainerName,
			Ports:         svc.Ports,
			Environment:   svc.Environment,
			Volumes:       svc.Volumes,
			Restart:       svc.Restart,
		}

		if svc.Healthcheck != nil {
			composeSvc.Healthcheck = &ComposeHealthcheck{
				Test:     svc.Healthcheck.Test,
				Interval: svc.Healthcheck.Interval,
				Timeout:  svc.Healthcheck.Timeout,
				Retries:  svc.Healthcheck.Retries,
			}
		}

		if len(svc.DependsOn) > 0 {
			dependsOn := make(map[string]map[string]string)
			for depName, depSpec := range svc.DependsOn {
				dependsOn[depName] = map[string]string{
					"condition": depSpec.Condition,
				}
			}
			composeSvc.DependsOn = dependsOn
		}

		compose.Services[name] = composeSvc
	}

	// Convert networks
	for name, net := range pkg.Spec.Container.Networks {
		compose.Networks[name] = ComposeNetwork{Name: net.Name}
	}

	return yaml.Marshal(&compose)
}

// generateSecret generates a random secret string
func generateSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("portunix-secret-%d", os.Getpid())
	}
	return hex.EncodeToString(bytes)
}

// getDeployDir returns the directory where compose files are stored
func getDeployDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "fider")
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create deploy directory: %w", err)
	}

	return deployDir, nil
}

// writeComposeFile generates and writes docker-compose.yaml from package JSON
func writeComposeFile(deployDir string) (string, error) {
	pkg, err := loadPackageDefinition("fider")
	if err != nil {
		return "", err
	}

	yamlData, err := generateComposeYAML(pkg)
	if err != nil {
		return "", fmt.Errorf("failed to generate compose YAML: %w", err)
	}

	composePath := filepath.Join(deployDir, fiderComposeFile)
	if err := os.WriteFile(composePath, yamlData, 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	return composePath, nil
}

// writeEnvFile writes environment variables for docker-compose
func writeEnvFile(deployDir string, config *Config) (string, error) {
	envPath := filepath.Join(deployDir, fiderEnvFile)

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
	dbPassword := existingEnv["FIDER_DB_PASSWORD"]
	if dbPassword == "" {
		dbPassword = generateSecret(16)
	}

	jwtSecret := existingEnv["FIDER_JWT_SECRET"]
	if jwtSecret == "" {
		jwtSecret = generateSecret(32)
	}

	// Determine base URL
	baseURL := config.GetEndpoint()
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	env := fmt.Sprintf(`# Fider environment configuration
# Generated by portunix pft deploy

FIDER_DB_PASSWORD=%s
FIDER_JWT_SECRET=%s
FIDER_BASE_URL=%s
FIDER_PORT=3000
FIDER_EMAIL_NOREPLY=noreply@localhost
`, dbPassword, jwtSecret, baseURL)

	if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
		return "", fmt.Errorf("failed to write env file: %w", err)
	}

	return envPath, nil
}

// findPortunix returns the path to portunix binary
func findPortunix() (string, error) {
	// Get executable directory (ptx-pft is next to portunix)
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Check for portunix in same directory
	portunixPath := filepath.Join(execDir, "portunix")
	if _, err := os.Stat(portunixPath); err == nil {
		return portunixPath, nil
	}

	// Check for portunix.exe on Windows
	portunixPath = filepath.Join(execDir, "portunix.exe")
	if _, err := os.Stat(portunixPath); err == nil {
		return portunixPath, nil
	}

	// Try PATH
	if path, err := exec.LookPath("portunix"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("portunix binary not found")
}

// runContainerCompose executes portunix container compose command
func runContainerCompose(deployDir string, args ...string) error {
	portunixPath, err := findPortunix()
	if err != nil {
		return err
	}

	// Build full command: portunix container compose -f <file> --env-file <env> -p <project> <args...>
	fullArgs := []string{
		"container", "compose",
		"-f", filepath.Join(deployDir, fiderComposeFile),
		"--env-file", filepath.Join(deployDir, fiderEnvFile),
		"-p", fiderProjectName,
	}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command(portunixPath, fullArgs...)
	cmd.Dir = deployDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// Deploy deploys Fider.io using Docker Compose
func Deploy(config *Config) (*DeployResult, error) {
	result := &DeployResult{}

	// Get deploy directory
	deployDir, err := getDeployDir()
	if err != nil {
		return nil, err
	}

	// Write compose file (generated from JSON)
	composePath, err := writeComposeFile(deployDir)
	if err != nil {
		return nil, err
	}
	result.ComposeFile = composePath

	// Write env file
	envPath, err := writeEnvFile(deployDir, config)
	if err != nil {
		return nil, err
	}
	result.EnvFile = envPath

	fmt.Println("Starting Fider deployment...")
	fmt.Printf("  Compose file: %s\n", composePath)
	fmt.Printf("  Environment: %s\n", envPath)
	fmt.Println()

	// Pull images
	fmt.Println("Pulling container images...")
	if err := runContainerCompose(deployDir, "pull"); err != nil {
		return nil, fmt.Errorf("failed to pull images: %w", err)
	}

	// Start services
	fmt.Println()
	fmt.Println("Starting services...")
	if err := runContainerCompose(deployDir, "up", "-d"); err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	// Determine URL
	baseURL := config.GetEndpoint()
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	result.Success = true
	result.URL = baseURL
	result.Message = fmt.Sprintf("Fider deployed successfully!\n\nAccess Fider at: %s\n\nNote: First startup may take 30-60 seconds for database initialization.", baseURL)

	return result, nil
}

// DeployInstance deploys a named Fider instance on specified port
func DeployInstance(instanceName string, port int, config *Config) (*DeployResult, error) {
	result := &DeployResult{}

	// Get instance-specific deploy directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "fider-"+instanceName)
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create deploy directory: %w", err)
	}

	projectName := fmt.Sprintf("portunix-fider-%s", instanceName)

	// Generate compose file with custom port
	composeContent := generateInstanceComposeYAML(instanceName, port)
	composePath := filepath.Join(deployDir, fiderComposeFile)
	if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write compose file: %w", err)
	}
	result.ComposeFile = composePath

	// Write env file with instance-specific settings
	envPath, err := writeInstanceEnvFile(deployDir, instanceName, port, config)
	if err != nil {
		return nil, err
	}
	result.EnvFile = envPath

	// Pull images
	if err := runInstanceContainerCompose(deployDir, projectName, "pull"); err != nil {
		return nil, fmt.Errorf("failed to pull images: %w", err)
	}

	// Start services
	if err := runInstanceContainerCompose(deployDir, projectName, "up", "-d"); err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	result.Success = true
	result.URL = baseURL
	result.Message = fmt.Sprintf("Fider (%s) deployed on port %d", instanceName, port)

	return result, nil
}

// generateInstanceComposeYAML generates docker-compose.yaml for a specific instance
func generateInstanceComposeYAML(instanceName string, port int) string {
	mailhogPort := port + 100 // Mailhog web UI on port+100 (e.g., 3200 for VoC, 3201 for VoS)
	return fmt.Sprintf(`services:
  db:
    image: postgres:15-alpine
    container_name: fider-%s-db
    environment:
      POSTGRES_DB: fider
      POSTGRES_USER: fider
      POSTGRES_PASSWORD: ${FIDER_DB_PASSWORD}
    volumes:
      - fider-%s-db:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U fider"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  mailhog:
    image: mailhog/mailhog:latest
    container_name: fider-%s-mail
    ports:
      - "%d:8025"
    restart: unless-stopped

  fider:
    image: getfider/fider:stable
    container_name: fider-%s-app
    ports:
      - "%d:3000"
    environment:
      BASE_URL: ${FIDER_BASE_URL}
      DATABASE_URL: postgres://fider:${FIDER_DB_PASSWORD}@db:5432/fider?sslmode=disable
      JWT_SECRET: ${FIDER_JWT_SECRET}
      EMAIL_NOREPLY: noreply@fider.local
      EMAIL_SMTP_HOST: mailhog
      EMAIL_SMTP_PORT: 1025
    depends_on:
      db:
        condition: service_healthy
      mailhog:
        condition: service_started
    restart: unless-stopped

volumes:
  fider-%s-db:
`, instanceName, instanceName, instanceName, mailhogPort, instanceName, port, instanceName)
}

// writeInstanceEnvFile writes environment file for a specific instance
func writeInstanceEnvFile(deployDir, instanceName string, port int, config *Config) (string, error) {
	envPath := filepath.Join(deployDir, fiderEnvFile)

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
	dbPassword := existingEnv["FIDER_DB_PASSWORD"]
	if dbPassword == "" {
		dbPassword = generateSecret(16)
	}

	jwtSecret := existingEnv["FIDER_JWT_SECRET"]
	if jwtSecret == "" {
		jwtSecret = generateSecret(32)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	env := fmt.Sprintf(`# Fider %s environment configuration
# Generated by portunix pft deploy

FIDER_DB_PASSWORD=%s
FIDER_JWT_SECRET=%s
FIDER_BASE_URL=%s
`, instanceName, dbPassword, jwtSecret, baseURL)

	if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
		return "", fmt.Errorf("failed to write env file: %w", err)
	}

	return envPath, nil
}

// runInstanceContainerCompose executes portunix container compose for a specific instance
func runInstanceContainerCompose(deployDir, projectName string, args ...string) error {
	portunixPath, err := findPortunix()
	if err != nil {
		return err
	}

	fullArgs := []string{
		"container", "compose",
		"-f", filepath.Join(deployDir, fiderComposeFile),
		"--env-file", filepath.Join(deployDir, fiderEnvFile),
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

// GetStatus returns the status of Fider deployment
func GetStatus() (string, error) {
	deployDir, err := getDeployDir()
	if err != nil {
		return "", err
	}

	// Check if compose file exists
	composePath := filepath.Join(deployDir, fiderComposeFile)
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
		"-p", fiderProjectName,
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

	// Check if all containers are running
	for _, state := range strings.Split(states, "\n") {
		if state != "running" {
			return "partial", nil
		}
	}

	return "running", nil
}

// GetContainerInfo returns detailed container information
func GetContainerInfo() (string, error) {
	deployDir, err := getDeployDir()
	if err != nil {
		return "", err
	}

	composePath := filepath.Join(deployDir, fiderComposeFile)
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return "Fider is not deployed.", nil
	}

	portunixPath, err := findPortunix()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(portunixPath,
		"container", "compose",
		"-f", composePath,
		"-p", fiderProjectName,
		"ps",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil
	}

	return string(output), nil
}

// DeployEmailOnly deploys only Mailhog for email-only mode
func DeployEmailOnly(config *Config) (*DeployResult, error) {
	result := &DeployResult{}

	// Get deploy directory for email-only mode
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "email-only")
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create deploy directory: %w", err)
	}

	projectName := "portunix-email"

	// Generate compose file with only Mailhog
	composeContent := generateEmailOnlyComposeYAML()
	composePath := filepath.Join(deployDir, fiderComposeFile)
	if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write compose file: %w", err)
	}
	result.ComposeFile = composePath

	// Write minimal env file
	envPath := filepath.Join(deployDir, fiderEnvFile)
	envContent := "# Email-only mode - no additional configuration needed\n"
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		return nil, fmt.Errorf("failed to write env file: %w", err)
	}
	result.EnvFile = envPath

	fmt.Println("Starting email-only deployment (Mailhog only)...")
	fmt.Printf("  Compose file: %s\n", composePath)
	fmt.Println()

	// Pull images
	fmt.Println("Pulling Mailhog image...")
	if err := runEmailOnlyContainerCompose(deployDir, projectName, "pull"); err != nil {
		return nil, fmt.Errorf("failed to pull images: %w", err)
	}

	// Start services
	fmt.Println()
	fmt.Println("Starting Mailhog...")
	if err := runEmailOnlyContainerCompose(deployDir, projectName, "up", "-d"); err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	result.Success = true
	result.URL = "http://localhost:3200"
	result.Message = fmt.Sprintf(`Email-only mode deployed successfully!

Mailhog Web UI: http://localhost:3200
SMTP Server:    localhost:1025

Note: In email-only mode, sync/pull/push commands are disabled.
      Use 'pft notify' to send emails and 'pft votes' to check responses.`)

	return result, nil
}

// generateEmailOnlyComposeYAML generates docker-compose.yaml with only Mailhog
func generateEmailOnlyComposeYAML() string {
	return `services:
  mailhog:
    image: mailhog/mailhog:latest
    container_name: pft-mailhog
    ports:
      - "3200:8025"
      - "1025:1025"
    restart: unless-stopped
`
}

// runEmailOnlyContainerCompose executes portunix container compose for email-only mode
func runEmailOnlyContainerCompose(deployDir, projectName string, args ...string) error {
	portunixPath, err := findPortunix()
	if err != nil {
		return err
	}

	fullArgs := []string{
		"container", "compose",
		"-f", filepath.Join(deployDir, fiderComposeFile),
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

// GetEmailOnlyStatus returns the status of email-only deployment
func GetEmailOnlyStatus() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "email-only")
	composePath := filepath.Join(deployDir, fiderComposeFile)

	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return "not_deployed", nil
	}

	portunixPath, err := findPortunix()
	if err != nil {
		return "unknown", err
	}

	cmd := exec.Command(portunixPath,
		"container", "compose",
		"-f", composePath,
		"-p", "portunix-email",
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

	if strings.Contains(states, "running") {
		return "running", nil
	}

	return "stopped", nil
}

// DestroyEmailOnly removes the email-only deployment
func DestroyEmailOnly(removeVolumes bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	deployDir := filepath.Join(homeDir, ".portunix", "pft", "email-only")
	composePath := filepath.Join(deployDir, fiderComposeFile)

	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("email-only mode is not deployed")
	}

	fmt.Println("Stopping and removing Mailhog container...")

	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	if err := runEmailOnlyContainerCompose(deployDir, "portunix-email", args...); err != nil {
		return fmt.Errorf("failed to destroy deployment: %w", err)
	}

	fmt.Println("Email-only deployment removed successfully.")

	if removeVolumes {
		if err := os.RemoveAll(deployDir); err != nil {
			fmt.Printf("Warning: could not remove deploy directory: %v\n", err)
		}
	}

	return nil
}

// Destroy removes the Fider deployment
func Destroy(removeVolumes bool) error {
	deployDir, err := getDeployDir()
	if err != nil {
		return err
	}

	composePath := filepath.Join(deployDir, fiderComposeFile)
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("Fider is not deployed")
	}

	fmt.Println("Stopping and removing Fider containers...")

	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
		fmt.Println("  (including volumes)")
	}

	if err := runContainerCompose(deployDir, args...); err != nil {
		return fmt.Errorf("failed to destroy deployment: %w", err)
	}

	fmt.Println("Fider deployment removed successfully.")

	// Optionally remove deploy directory
	if removeVolumes {
		if err := os.RemoveAll(deployDir); err != nil {
			fmt.Printf("Warning: could not remove deploy directory: %v\n", err)
		}
	}

	return nil
}
