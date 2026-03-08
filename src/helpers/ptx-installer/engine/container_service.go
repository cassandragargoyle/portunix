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

	// Build container name
	containerName := container.Name
	if containerName == "" {
		containerName = fmt.Sprintf("portunix-%s", options.PackageName)
	}

	// Build image reference with tag
	imageRef := container.Image
	if container.Tag != "" {
		imageRef = fmt.Sprintf("%s:%s", container.Image, container.Tag)
	}

	fmt.Printf("🐳 Installing %s as container service\n", options.PackageName)
	fmt.Printf("   Image: %s\n", imageRef)
	fmt.Printf("   Container name: %s\n", containerName)

	if options.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - Would perform the following:")
		fmt.Printf("   1. Check container runtime availability\n")
		fmt.Printf("   2. Pull image: %s\n", imageRef)
		fmt.Printf("   3. Create and start container: %s\n", containerName)
		if len(container.Ports) > 0 {
			fmt.Printf("   4. Expose ports: %v\n", container.Ports)
		}
		if container.HealthCheck != nil {
			fmt.Printf("   5. Wait for health check: %s\n", container.HealthCheck.Endpoint)
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

	// Check if container already exists
	existingContainer := checkContainerExists(portunixPath, containerName)
	if existingContainer {
		fmt.Printf("⚠️  Container '%s' already exists\n", containerName)
		fmt.Printf("   Use 'portunix container rm %s' to remove it first, or choose a different name\n", containerName)
		return fmt.Errorf("container already exists: %s", containerName)
	}

	// Build portunix container run command
	args := []string{"container", "run", "-d", "--name", containerName}

	// Add port mappings
	for _, port := range container.Ports {
		args = append(args, "-p", port)
	}

	// Add volume mappings
	for _, volume := range container.Volumes {
		args = append(args, "-v", volume)
	}

	// Add environment variables
	for key, value := range container.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add image reference
	args = append(args, imageRef)

	// Add container command if specified
	if len(container.Command) > 0 {
		args = append(args, container.Command...)
	}

	fmt.Printf("🚀 Starting container...\n")
	fmt.Printf("   Command: portunix %s\n", strings.Join(args, " "))

	cmd := exec.Command(portunixPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✅ Container '%s' started\n", containerName)

	// Print access information
	fmt.Println("\n📋 Container Information:")
	fmt.Printf("   Name: %s\n", containerName)
	if len(container.Ports) > 0 {
		fmt.Printf("   Ports: %v\n", container.Ports)
		// Extract first port for convenience message
		if len(container.Ports) > 0 {
			port := strings.Split(container.Ports[0], ":")[0]
			fmt.Printf("   Access: http://localhost:%s\n", port)
		}
	}
	fmt.Println("\n💡 Useful commands:")
	fmt.Printf("   View logs:    portunix container logs %s\n", containerName)
	fmt.Printf("   Stop:         portunix container stop %s\n", containerName)
	fmt.Printf("   Start:        portunix container start %s\n", containerName)
	fmt.Printf("   Remove:       portunix container rm %s\n", containerName)

	// Wait for health check if specified
	if container.HealthCheck != nil && container.HealthCheck.Endpoint != "" {
		fmt.Printf("\n⏳ Waiting for service to be ready...\n")
		if err := waitForHealthCheck(container.HealthCheck); err != nil {
			fmt.Printf("⚠️  Health check failed: %v\n", err)
			fmt.Println("   Container is running but service may not be ready yet")
			fmt.Printf("   Check logs: portunix container logs %s\n", containerName)
			return fmt.Errorf("container started but health check failed: %w", err)
		}
		fmt.Println("✅ Service is ready!")
	}

	return nil
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

// checkContainerExists checks if a container with given name already exists
func checkContainerExists(portunixPath, containerName string) bool {
	cmd := exec.Command(portunixPath, "container", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), containerName)
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
