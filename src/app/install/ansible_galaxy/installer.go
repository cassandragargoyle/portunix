package ansible_galaxy

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// AnsibleGalaxyInstaller handles Ansible Galaxy collection installation
type AnsibleGalaxyInstaller struct {
	Debug         bool
	DryRun        bool
	Timeout       time.Duration
	AnsiblePath   string
	GalaxyPath    string
	CollectionDir string
}

// NewAnsibleGalaxyInstaller creates a new Ansible Galaxy installer
func NewAnsibleGalaxyInstaller() *AnsibleGalaxyInstaller {
	installer := &AnsibleGalaxyInstaller{
		Debug:   false,
		DryRun:  false,
		Timeout: 10 * time.Minute,
	}

	// Detect ansible and ansible-galaxy paths
	installer.detectPaths()

	return installer
}

// CollectionInfo represents information about an Ansible collection
type CollectionInfo struct {
	Name        string
	Namespace   string
	Collection  string
	Version     string
	Description string
	Installed   bool
	Available   bool
}

// IsSupported checks if Ansible Galaxy is available
func (ag *AnsibleGalaxyInstaller) IsSupported() bool {
	// Check if ansible-galaxy is available
	if ag.GalaxyPath == "" {
		ag.detectPaths()
	}

	if ag.GalaxyPath == "" {
		return false
	}

	// Test ansible-galaxy version command
	cmd := exec.Command(ag.GalaxyPath, "--version")
	return cmd.Run() == nil
}

// detectPaths finds ansible and ansible-galaxy executables
func (ag *AnsibleGalaxyInstaller) detectPaths() {
	// Try to find ansible-galaxy
	galaxyCommands := []string{"ansible-galaxy", "python3 -m ansible.galaxy", "python -m ansible.galaxy"}

	for _, cmd := range galaxyCommands {
		parts := strings.Fields(cmd)
		var execCmd *exec.Cmd
		if len(parts) == 1 {
			execCmd = exec.Command(parts[0], "--version")
		} else {
			execCmd = exec.Command(parts[0], append(parts[1:], "--version")...)
		}

		if execCmd.Run() == nil {
			ag.GalaxyPath = cmd
			break
		}
	}

	// Try to find ansible
	ansibleCommands := []string{"ansible", "python3 -m ansible", "python -m ansible"}

	for _, cmd := range ansibleCommands {
		parts := strings.Fields(cmd)
		var execCmd *exec.Cmd
		if len(parts) == 1 {
			execCmd = exec.Command(parts[0], "--version")
		} else {
			execCmd = exec.Command(parts[0], append(parts[1:], "--version")...)
		}

		if execCmd.Run() == nil {
			ag.AnsiblePath = cmd
			break
		}
	}
}

// ValidateCollectionName validates the format of a collection name
func (ag *AnsibleGalaxyInstaller) ValidateCollectionName(collection string) error {
	// Collection name must be in format namespace.collection
	parts := strings.Split(collection, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid collection format, expected 'namespace.collection', got '%s'", collection)
	}

	// Validate namespace and collection name parts
	nameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

	if !nameRegex.MatchString(parts[0]) {
		return fmt.Errorf("invalid namespace '%s': must start with letter and contain only letters, numbers, and underscores", parts[0])
	}

	if !nameRegex.MatchString(parts[1]) {
		return fmt.Errorf("invalid collection name '%s': must start with letter and contain only letters, numbers, and underscores", parts[1])
	}

	return nil
}

// Install installs a collection using ansible-galaxy
func (ag *AnsibleGalaxyInstaller) Install(collections []string) error {
	if len(collections) == 0 {
		return fmt.Errorf("no collections specified for installation")
	}

	for _, collection := range collections {
		if err := ag.installSingleCollection(collection); err != nil {
			return fmt.Errorf("failed to install collection '%s': %w", collection, err)
		}
	}

	return nil
}

// installSingleCollection installs a single collection
func (ag *AnsibleGalaxyInstaller) installSingleCollection(collection string) error {
	// Validate collection name
	if err := ag.ValidateCollectionName(collection); err != nil {
		return err
	}

	if ag.Debug {
		fmt.Printf("📦 Installing Ansible collection: %s\n", collection)
	}

	if ag.DryRun {
		fmt.Printf("🔄 [DRY-RUN] Would install Ansible collection: %s\n", collection)
		return nil
	}

	// Parse collection name and version
	collectionName, version := ag.parseCollectionWithVersion(collection)

	// Build ansible-galaxy install command
	parts := strings.Fields(ag.GalaxyPath)
	args := append(parts[1:], "collection", "install", collectionName)

	// Add version constraint if specified
	if version != "" && version != "latest" {
		args = append(args, ":"+version)
	}

	// Add upgrade flag to ensure latest version
	args = append(args, "--upgrade")

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0], args...)
	} else {
		cmd = exec.Command(parts[0], args...)
	}

	if ag.Debug {
		fmt.Printf("🔧 Executing: %s %s\n", cmd.Path, strings.Join(args, " "))
	}

	// Set up output handling
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ansible-galaxy install: %w", err)
	}

	// Read output in real-time
	go ag.readOutput(stdout, "STDOUT")
	go ag.readOutput(stderr, "STDERR")

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ansible-galaxy install failed: %w", err)
	}

	fmt.Printf("✅ Successfully installed collection: %s\n", collection)
	return nil
}

// parseCollectionWithVersion parses collection name with optional version
func (ag *AnsibleGalaxyInstaller) parseCollectionWithVersion(collection string) (string, string) {
	if strings.Contains(collection, ":") {
		parts := strings.Split(collection, ":")
		return parts[0], parts[1]
	}
	return collection, ""
}

// readOutput reads and displays command output
func (ag *AnsibleGalaxyInstaller) readOutput(pipe io.ReadCloser, prefix string) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		line := scanner.Text()
		if ag.Debug {
			fmt.Printf("[%s] %s\n", prefix, line)
		} else {
			// Show important lines even in non-debug mode
			if ag.shouldShowLine(line) {
				fmt.Printf("   %s\n", line)
			}
		}
	}
}

// shouldShowLine determines if a line should be shown in non-debug mode
func (ag *AnsibleGalaxyInstaller) shouldShowLine(line string) bool {
	lowerLine := strings.ToLower(line)
	return strings.Contains(lowerLine, "installing") ||
		strings.Contains(lowerLine, "installed") ||
		strings.Contains(lowerLine, "downloading") ||
		strings.Contains(lowerLine, "error") ||
		strings.Contains(lowerLine, "warning") ||
		strings.Contains(lowerLine, "failed")
}

// IsInstalled checks if a collection is installed
func (ag *AnsibleGalaxyInstaller) IsInstalled(collection string) bool {
	collections, err := ag.ListInstalled()
	if err != nil {
		return false
	}

	for _, installed := range collections {
		if installed.Name == collection {
			return true
		}
	}

	return false
}

// ListInstalled returns list of installed collections
func (ag *AnsibleGalaxyInstaller) ListInstalled() ([]CollectionInfo, error) {
	parts := strings.Fields(ag.GalaxyPath)
	args := append(parts[1:], "collection", "list")

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0], args...)
	} else {
		cmd = exec.Command(parts[0], args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return ag.parseCollectionList(string(output)), nil
}

// parseCollectionList parses the output of ansible-galaxy collection list
func (ag *AnsibleGalaxyInstaller) parseCollectionList(output string) []CollectionInfo {
	var collections []CollectionInfo
	lines := strings.Split(output, "\n")

	// Parse collection list output
	// Expected format: namespace.collection    version
	collectionRegex := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_]*\.[a-zA-Z][a-zA-Z0-9_]*)\s+([0-9]+\.[0-9]+\.[0-9]+.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := collectionRegex.FindStringSubmatch(line); len(matches) == 3 {
			name := matches[1]
			version := strings.TrimSpace(matches[2])

			parts := strings.Split(name, ".")
			namespace := parts[0]
			collection := parts[1]

			collections = append(collections, CollectionInfo{
				Name:       name,
				Namespace:  namespace,
				Collection: collection,
				Version:    version,
				Installed:  true,
				Available:  true,
			})
		}
	}

	return collections
}

// GetCollectionInfo retrieves detailed information about a collection
func (ag *AnsibleGalaxyInstaller) GetCollectionInfo(collection string) (*CollectionInfo, error) {
	// Validate collection name
	if err := ag.ValidateCollectionName(collection); err != nil {
		return nil, err
	}

	// Check if installed
	installed := ag.IsInstalled(collection)

	parts := strings.Split(collection, ".")
	info := &CollectionInfo{
		Name:       collection,
		Namespace:  parts[0],
		Collection: parts[1],
		Installed:  installed,
		Available:  true, // Assume available unless proven otherwise
	}

	// If installed, get version
	if installed {
		collections, err := ag.ListInstalled()
		if err == nil {
			for _, installedCollection := range collections {
				if installedCollection.Name == collection {
					info.Version = installedCollection.Version
					break
				}
			}
		}
	}

	return info, nil
}

// Upgrade upgrades a collection to the latest version
func (ag *AnsibleGalaxyInstaller) Upgrade(collection string) error {
	// Upgrade is the same as install with --upgrade flag
	return ag.installSingleCollection(collection)
}

// Remove removes a collection (manual implementation)
func (ag *AnsibleGalaxyInstaller) Remove(collection string) error {
	if ag.DryRun {
		fmt.Printf("🔄 [DRY-RUN] Would remove collection: %s\n", collection)
		return nil
	}

	// ansible-galaxy doesn't have a native remove command
	// We need to manually remove the collection directory
	fmt.Printf("⚠️  Warning: Collection removal requires manual directory cleanup\n")
	fmt.Printf("   Collection '%s' may still be present in your ansible collections path\n", collection)
	fmt.Printf("   To completely remove, manually delete the collection directory\n")

	// TODO: Implement manual collection removal via filesystem operations
	// This would require finding the collections path and removing the specific collection directory

	return fmt.Errorf("collection removal not yet implemented - manual cleanup required")
}