package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PtxbookFile represents the top-level structure of a .ptxbook file
type PtxbookFile struct {
	APIVersion string           `yaml:"apiVersion" json:"apiVersion"`
	Kind       string           `yaml:"kind" json:"kind"`
	Metadata   PtxbookMetadata  `yaml:"metadata" json:"metadata"`
	Spec       PtxbookSpec      `yaml:"spec" json:"spec"`
}

// PtxbookMetadata represents the metadata section
type PtxbookMetadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// PtxbookSpec represents the spec section
type PtxbookSpec struct {
	Variables    map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	Requirements *PtxbookRequirements   `yaml:"requirements,omitempty" json:"requirements,omitempty"`
	Portunix     *PtxbookPortunix       `yaml:"portunix,omitempty" json:"portunix,omitempty"`
	Ansible      *PtxbookAnsible        `yaml:"ansible,omitempty" json:"ansible,omitempty"`
	Scripts      map[string]string      `yaml:"scripts,omitempty" json:"scripts,omitempty"`       // Custom scripts (init, build, serve, etc.)
	ScriptsExt   map[string]ScriptConfig `yaml:"scripts_ext,omitempty" json:"scripts_ext,omitempty"` // Extended scripts with conditions (Issue #128)
	// Phase 3: Advanced features
	Rollback     *PtxbookRollback       `yaml:"rollback,omitempty" json:"rollback,omitempty"`     // Rollback configuration
	Environment  map[string]interface{} `yaml:"environment,omitempty" json:"environment,omitempty"` // Environment configuration (target, runtime, image)
	// Phase 4: Enterprise features
	Secrets      map[string]interface{} `yaml:"secrets,omitempty" json:"secrets,omitempty"`     // Secret references
}

// ScriptConfig represents a script with optional condition (Issue #128 Phase 3)
type ScriptConfig struct {
	Command     string `yaml:"command" json:"command"`                           // The command to execute
	Condition   string `yaml:"condition,omitempty" json:"condition,omitempty"`   // Shell condition to evaluate (e.g., "! -d ./site")
	Description string `yaml:"description,omitempty" json:"description,omitempty"` // Optional description for --list-scripts
}

// PtxbookRequirements represents the requirements section
type PtxbookRequirements struct {
	Ansible *AnsibleRequirements `yaml:"ansible,omitempty" json:"ansible,omitempty"`
}

// AnsibleRequirements represents Ansible version requirements
type AnsibleRequirements struct {
	MinVersion string `yaml:"min_version,omitempty" json:"min_version,omitempty"`
}

// PtxbookPortunix represents the Portunix package management section
type PtxbookPortunix struct {
	Packages []PtxbookPackage `yaml:"packages,omitempty" json:"packages,omitempty"`
}

// PtxbookPackage represents a Portunix package installation
type PtxbookPackage struct {
	Name    string                 `yaml:"name" json:"name"`
	Variant string                 `yaml:"variant,omitempty" json:"variant,omitempty"`
	When    string                 `yaml:"when,omitempty" json:"when,omitempty"`     // Phase 3: Conditional execution
	Vars    map[string]interface{} `yaml:"vars,omitempty" json:"vars,omitempty"`     // Phase 3: Package-specific variables
}

// PtxbookAnsible represents the Ansible playbooks section
type PtxbookAnsible struct {
	Playbooks []AnsiblePlaybook `yaml:"playbooks,omitempty" json:"playbooks,omitempty"`
}

// AnsiblePlaybook represents an Ansible playbook reference
type AnsiblePlaybook struct {
	Path string                 `yaml:"path" json:"path"`
	When string                 `yaml:"when,omitempty" json:"when,omitempty"`     // Phase 3: Conditional execution
	Vars map[string]interface{} `yaml:"vars,omitempty" json:"vars,omitempty"`     // Phase 3: Playbook-specific variables
}

// ParsePtxbookFile parses a .ptxbook file and returns the structured data
func ParsePtxbookFile(filePath string) (*PtxbookFile, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("playbook file not found: %s", filePath)
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".ptxbook") {
		return nil, fmt.Errorf("invalid file extension: expected .ptxbook, got %s", filepath.Ext(filePath))
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook file: %v", err)
	}

	// Parse YAML content
	var ptxbook PtxbookFile
	if err := yaml.Unmarshal(content, &ptxbook); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Validate the parsed content
	if err := ValidatePtxbookFile(&ptxbook); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	return &ptxbook, nil
}

// ValidatePtxbookFile validates the structure and content of a parsed .ptxbook file
func ValidatePtxbookFile(ptxbook *PtxbookFile) error {
	// Validate API version
	if ptxbook.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if ptxbook.APIVersion != "portunix.ai/v1" {
		return fmt.Errorf("unsupported apiVersion: %s (expected: portunix.ai/v1)", ptxbook.APIVersion)
	}

	// Validate kind
	if ptxbook.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if ptxbook.Kind != "Playbook" {
		return fmt.Errorf("unsupported kind: %s (expected: Playbook)", ptxbook.Kind)
	}

	// Validate metadata
	if ptxbook.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	// Validate that at least Portunix, Ansible, or Scripts section exists
	hasPortunix := ptxbook.Spec.Portunix != nil && len(ptxbook.Spec.Portunix.Packages) > 0
	hasAnsible := ptxbook.Spec.Ansible != nil && len(ptxbook.Spec.Ansible.Playbooks) > 0
	hasScripts := len(ptxbook.Spec.Scripts) > 0
	if !hasPortunix && !hasAnsible && !hasScripts {
		return fmt.Errorf("spec must contain at least one of: 'portunix', 'ansible', or 'scripts' section")
	}

	// Validate Portunix packages if present
	if ptxbook.Spec.Portunix != nil {
		for i, pkg := range ptxbook.Spec.Portunix.Packages {
			if pkg.Name == "" {
				return fmt.Errorf("spec.portunix.packages[%d].name is required", i)
			}
		}
	}

	// Validate Ansible playbooks if present
	if ptxbook.Spec.Ansible != nil {
		for i, playbook := range ptxbook.Spec.Ansible.Playbooks {
			if playbook.Path == "" {
				return fmt.Errorf("spec.ansible.playbooks[%d].path is required", i)
			}
		}
	}

	return nil
}

// IsPtxbookOnlyFile checks if a .ptxbook file contains only Portunix commands (no Ansible)
func IsPtxbookOnlyFile(ptxbook *PtxbookFile) bool {
	return ptxbook.Spec.Ansible == nil || len(ptxbook.Spec.Ansible.Playbooks) == 0
}

// RequiresAnsible checks if a .ptxbook file requires Ansible to be installed
func RequiresAnsible(ptxbook *PtxbookFile) bool {
	return !IsPtxbookOnlyFile(ptxbook)
}

// GetMinAnsibleVersion returns the minimum required Ansible version, if specified
func GetMinAnsibleVersion(ptxbook *PtxbookFile) string {
	if ptxbook.Spec.Requirements != nil &&
	   ptxbook.Spec.Requirements.Ansible != nil &&
	   ptxbook.Spec.Requirements.Ansible.MinVersion != "" {
		return ptxbook.Spec.Requirements.Ansible.MinVersion
	}
	return "2.15.0" // Default minimum version
}

// Phase 3: Advanced Features Structures

// PtxbookRollback represents rollback configuration for error handling
type PtxbookRollback struct {
	Enabled         bool                   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	OnFailure       []RollbackAction       `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
	PreserveLogs    bool                   `yaml:"preserve_logs,omitempty" json:"preserve_logs,omitempty"`
	Timeout         string                 `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	RetryCount      int                    `yaml:"retry_count,omitempty" json:"retry_count,omitempty"`
	CustomVariables map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// RollbackAction represents a single rollback action
type RollbackAction struct {
	Type        string                 `yaml:"type" json:"type"` // "command", "package_remove", "file_restore"
	Command     string                 `yaml:"command,omitempty" json:"command,omitempty"`
	Package     string                 `yaml:"package,omitempty" json:"package,omitempty"`
	Path        string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	When        string                 `yaml:"when,omitempty" json:"when,omitempty"`
	Vars        map[string]interface{} `yaml:"vars,omitempty" json:"vars,omitempty"`
}