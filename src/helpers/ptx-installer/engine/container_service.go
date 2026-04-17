/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// installContainer installs a service via container using portunix container commands
func (i *Installer) installContainer(_ *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if variant.Container == nil {
		return fmt.Errorf("container configuration is required for container installation type")
	}

	container := variant.Container

	// Validate required fields
	if container.Image == "" {
		return fmt.Errorf("container image is required")
	}

	// Apply --db-* overrides to container.Environment before anything else
	// reads it. Unset flags are no-ops — variant defaults from JSON stand.
	applyDBOverrides(container, options)

	// Build primary container name
	primaryName := container.Name
	if primaryName == "" {
		primaryName = fmt.Sprintf("portunix-%s", options.PackageName)
	}

	// Build image reference with tag
	imageRef := container.Image
	if container.Tag != "" {
		imageRef = fmt.Sprintf("%s:%s", container.Image, container.Tag)
	}

	fmt.Printf("🐳 Installing %s as container service\n", options.PackageName)
	fmt.Printf("   Image: %s\n", imageRef)
	fmt.Printf("   Container name: %s\n", primaryName)
	if container.Network != "" {
		fmt.Printf("   Network: %s\n", container.Network)
	}
	if len(container.Sidecars) > 0 {
		names := make([]string, 0, len(container.Sidecars))
		for _, s := range container.Sidecars {
			names = append(names, s.Name)
		}
		fmt.Printf("   Sidecars: %s\n", strings.Join(names, ", "))
	}

	if options.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - Would perform the following:")
		fmt.Printf("   1. Check container runtime availability\n")
		if container.Network != "" {
			fmt.Printf("   2. Ensure network exists: %s\n", container.Network)
		}
		for idx, s := range container.Sidecars {
			sImage := s.Image
			if s.Tag != "" {
				sImage = fmt.Sprintf("%s:%s", s.Image, s.Tag)
			}
			fmt.Printf("   %d. Start sidecar '%s' from %s\n", 3+idx, s.Name, sImage)
		}
		fmt.Printf("   %d. Pull image: %s\n", 3+len(container.Sidecars), imageRef)
		fmt.Printf("   %d. Create and start primary container: %s\n", 4+len(container.Sidecars), primaryName)
		if len(container.Ports) > 0 {
			fmt.Printf("      Expose ports: %v\n", container.Ports)
		}
		if container.HealthCheck != nil {
			fmt.Printf("      Wait for health check: %s\n", container.HealthCheck.Endpoint)
		}
		return nil
	}

	// Check if portunix binary is available
	portunixPath, err := findPortunixBinary()
	if err != nil {
		return fmt.Errorf("portunix binary not found: %w", err)
	}

	// Check container runtime availability
	fmt.Println("🔍 Checking container runtime...")
	checkCmd := exec.Command(portunixPath, "container", "info")
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("no container runtime available. Install Docker or Podman first: portunix install docker")
	}

	// Create shared network if requested
	if container.Network != "" {
		if err := ensureNetwork(container.Network); err != nil {
			return fmt.Errorf("failed to ensure network %q: %w", container.Network, err)
		}
	}

	// Launch sidecars (in declaration order) before the primary
	for _, sidecar := range container.Sidecars {
		if err := launchSidecar(portunixPath, &sidecar, container.Network); err != nil {
			return fmt.Errorf("sidecar %q failed: %w", sidecar.Name, err)
		}
	}

	// Check if primary container already exists
	if checkContainerExists(portunixPath, primaryName) {
		fmt.Printf("⚠️  Container '%s' already exists\n", primaryName)
		fmt.Printf("   Use 'portunix container rm %s' to remove it first, or choose a different name\n", primaryName)
		return fmt.Errorf("container already exists: %s", primaryName)
	}

	// Build portunix container run command for primary
	args := buildRunArgs(primaryName, imageRef, container, container.Network)

	fmt.Printf("🚀 Starting container...\n")
	fmt.Printf("   Command: portunix %s\n", strings.Join(args, " "))

	cmd := exec.Command(portunixPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✅ Container '%s' started\n", primaryName)

	// Print access information
	fmt.Println("\n📋 Container Information:")
	fmt.Printf("   Name: %s\n", primaryName)
	if len(container.Ports) > 0 {
		fmt.Printf("   Ports: %v\n", container.Ports)
		// Extract first port for convenience message
		if len(container.Ports) > 0 {
			port := strings.Split(container.Ports[0], ":")[0]
			fmt.Printf("   Access: http://localhost:%s\n", port)
		}
	}
	fmt.Println("\n💡 Useful commands:")
	fmt.Printf("   View logs:    portunix container logs %s\n", primaryName)
	fmt.Printf("   Stop:         portunix container stop %s\n", primaryName)
	fmt.Printf("   Start:        portunix container start %s\n", primaryName)
	fmt.Printf("   Remove:       portunix container rm %s\n", primaryName)

	// Wait for health check if specified
	if container.HealthCheck != nil && container.HealthCheck.Endpoint != "" {
		fmt.Printf("\n⏳ Waiting for service to be ready...\n")
		if err := waitForHealthCheck(container.HealthCheck); err != nil {
			fmt.Printf("⚠️  Health check failed: %v\n", err)
			fmt.Println("   Container is running but service may not be ready yet")
			fmt.Printf("   Check logs: portunix container logs %s\n", primaryName)
			return fmt.Errorf("container started but health check failed: %w", err)
		}
		fmt.Println("✅ Service is ready!")
	}

	return nil
}

// buildRunArgs assembles arguments for `portunix container run` from a ContainerSpec.
func buildRunArgs(name, imageRef string, c *registry.ContainerSpec, network string) []string {
	args := []string{"container", "run", "-d", "--name", name}
	if network != "" {
		args = append(args, "--network", network)
	}
	for _, port := range c.Ports {
		args = append(args, "-p", port)
	}
	for _, volume := range c.Volumes {
		args = append(args, "-v", volume)
	}
	for key, value := range c.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	args = append(args, imageRef)
	if len(c.Command) > 0 {
		args = append(args, c.Command...)
	}
	return args
}

// launchSidecar starts a dependency container and awaits its readiness.
// It is idempotent at the container level: if a container with the same name
// already exists it is left alone (assumed managed externally or from a prior run).
func launchSidecar(portunixPath string, s *registry.ContainerSpec, network string) error {
	if s.Name == "" {
		return fmt.Errorf("sidecar must have a name")
	}
	if s.Image == "" {
		return fmt.Errorf("sidecar %q must have an image", s.Name)
	}

	imageRef := s.Image
	if s.Tag != "" {
		imageRef = fmt.Sprintf("%s:%s", s.Image, s.Tag)
	}

	if checkContainerExists(portunixPath, s.Name) {
		fmt.Printf("ℹ️  Sidecar '%s' already exists — reusing\n", s.Name)
	} else {
		fmt.Printf("🚢 Starting sidecar '%s' (%s)\n", s.Name, imageRef)
		args := buildRunArgs(s.Name, imageRef, s, network)
		fmt.Printf("   Command: portunix %s\n", strings.Join(args, " "))
		cmd := exec.Command(portunixPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start sidecar: %w", err)
		}
	}

	// Await sidecar readiness
	if s.HealthCheck != nil && s.HealthCheck.Endpoint != "" {
		fmt.Printf("   Awaiting sidecar health...\n")
		if err := waitForHealthCheck(s.HealthCheck); err != nil {
			return fmt.Errorf("sidecar health check: %w", err)
		}
	} else {
		// Fallback pause so the sidecar has a chance to bind its ports before
		// the primary starts. Deliberately short — health checks are preferred.
		time.Sleep(3 * time.Second)
	}
	return nil
}

// ensureNetwork creates a container network via `portunix container network create`.
// The wrapper is idempotent (pre-existing networks are reported as informational,
// not errors), so we can rely on a single call without a prior inspect.
func ensureNetwork(name string) error {
	portunixPath, err := findPortunixBinary()
	if err != nil {
		return err
	}
	fmt.Printf("🔗 Ensuring network '%s'\n", name)
	cmd := exec.Command(portunixPath, "container", "network", "create", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("portunix container network create failed: %w", err)
	}
	return nil
}

// applyDBOverrides maps --db-* InstallOptions fields onto a ContainerSpec's
// Environment using PostgreSQL-style keys (HOST, PORT, USER, PASSWORD). Unset
// options leave the key untouched. If the map is nil, it is initialised only
// when at least one override is present.
func applyDBOverrides(c *registry.ContainerSpec, opts *InstallOptions) {
	if opts == nil {
		return
	}
	if opts.DBHost == "" && opts.DBPort == "" && opts.DBUser == "" && opts.DBPassword == "" {
		return
	}
	if c.Environment == nil {
		c.Environment = map[string]string{}
	}
	if opts.DBHost != "" {
		c.Environment["HOST"] = opts.DBHost
	}
	if opts.DBPort != "" {
		c.Environment["PORT"] = opts.DBPort
	}
	if opts.DBUser != "" {
		c.Environment["USER"] = opts.DBUser
	}
	if opts.DBPassword != "" {
		c.Environment["PASSWORD"] = opts.DBPassword
	}
}

// findPortunixBinary locates the portunix binary
// Prefers the binary next to the current executable to ensure version consistency
func findPortunixBinary() (string, error) {
	// Try relative to current executable first (ensures matching versions)
	execPath, err := os.Executable()
	if err == nil {
		dir := execPath[:strings.LastIndex(execPath, string(os.PathSeparator))]
		portunixPath := dir + string(os.PathSeparator) + "portunix"
		if _, err := os.Stat(portunixPath); err == nil {
			return portunixPath, nil
		}
	}

	// Fallback to PATH
	path, err := exec.LookPath("portunix")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("portunix binary not found in PATH or alongside ptx-installer")
}

// checkContainerExists returns true if a container with the given name is
// known to the runtime. Uses `portunix container inspect`, which exits
// non-zero when the container is absent — an exact signal that avoids the
// substring hazards of `container list` output parsing.
func checkContainerExists(portunixPath, containerName string) bool {
	cmd := exec.Command(portunixPath, "container", "inspect", containerName)
	// Suppress inspect output; we only care about the exit code.
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// waitForHealthCheck waits for the service health check to pass
func waitForHealthCheck(healthCheck *registry.HealthCheckSpec) error {
	timeout := healthCheck.Timeout
	if timeout <= 0 {
		timeout = 60 // default 60 seconds
	}

	interval := healthCheck.Interval
	if interval <= 0 {
		interval = 5 // default 5 seconds
	}

	retries := healthCheck.Retries
	if retries <= 0 {
		retries = timeout / interval
	}

	endpoint := healthCheck.Endpoint
	if endpoint == "" {
		return fmt.Errorf("no health check endpoint specified")
	}

	fmt.Printf("   Checking: %s (timeout: %ds)\n", endpoint, timeout)

	client := &http.Client{
		Timeout: time.Duration(interval) * time.Second,
	}

	for attempt := 0; attempt < retries; attempt++ {
		resp, err := client.Get(endpoint)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		fmt.Printf("   Attempt %d/%d - waiting...\n", attempt+1, retries)
		time.Sleep(time.Duration(interval) * time.Second)
	}

	return fmt.Errorf("health check timed out after %d seconds", timeout)
}
