package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RoleDefinition defines a role
type RoleDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleFile represents a roles.json file
type RoleFile struct {
	Type  string           `json:"type"`
	Roles []RoleDefinition `json:"roles"`
}

// DefaultVoCRoles returns default VoC roles
func DefaultVoCRoles() *RoleFile {
	return &RoleFile{
		Type: "voc",
		Roles: []RoleDefinition{
			{ID: "customer", Name: "Customer", Description: "Direct end-user providing feedback"},
			{ID: "customer-admin", Name: "Customer Admin", Description: "Administrator at customer organization"},
			{ID: "customer-support", Name: "Customer Support", Description: "Support staff at customer organization"},
			{ID: "facilitator", Name: "Facilitator", Description: "Person responsible for gathering, moderating and documenting customer feedback"},
			{ID: "proxy-customer", Name: "Proxy Customer", Description: "Representative speaking on behalf of customers"},
		},
	}
}

// DefaultVoSRoles returns default VoS roles
func DefaultVoSRoles() *RoleFile {
	return &RoleFile{
		Type: "vos",
		Roles: []RoleDefinition{
			{ID: "architect", Name: "Architect", Description: "Software/Solution architect"},
			{ID: "ceo", Name: "CEO", Description: "Chief Executive Officer / Company owner"},
			{ID: "cio", Name: "CIO", Description: "Chief Information Officer"},
			{ID: "dev-lead", Name: "Dev Lead", Description: "Development team lead"},
			{ID: "developer", Name: "Developer", Description: "Software developer"},
			{ID: "facilitator", Name: "Facilitator", Description: "Person responsible for gathering, moderating and documenting requirements"},
			{ID: "product-manager", Name: "Product Manager", Description: "Product management"},
			{ID: "support", Name: "Support", Description: "Support staff at software vendor"},
			{ID: "support-lead", Name: "Support Lead", Description: "Support team lead at software vendor"},
			{ID: "tech-consultant", Name: "Technical Consultant", Description: "Technical consultant"},
			{ID: "tester", Name: "Tester", Description: "Software tester providing stakeholder perspective"},
		},
	}
}

// DefaultVoBRoles returns default VoB roles
func DefaultVoBRoles() *RoleFile {
	return &RoleFile{
		Type: "vob",
		Roles: []RoleDefinition{
			{ID: "ceo", Name: "CEO", Description: "Chief Executive Officer / Company owner"},
			{ID: "dev-lead", Name: "Dev Lead", Description: "Development team lead"},
			{ID: "marketing", Name: "Marketing", Description: "Marketing specialist"},
			{ID: "sales", Name: "Sales", Description: "Sales representative"},
			{ID: "support", Name: "Support", Description: "Support staff at software vendor"},
		},
	}
}

// DefaultVoERoles returns default VoE roles
func DefaultVoERoles() *RoleFile {
	return &RoleFile{
		Type: "voe",
		Roles: []RoleDefinition{
			{ID: "architect", Name: "Architect", Description: "Software/Solution architect"},
			{ID: "developer", Name: "Developer", Description: "Software developer"},
			{ID: "devops", Name: "DevOps", Description: "DevOps engineer"},
			{ID: "qa", Name: "QA", Description: "Quality assurance engineer"},
			{ID: "senior-developer", Name: "Senior Developer", Description: "Senior software developer"},
			{ID: "support", Name: "Support", Description: "Support staff at software vendor"},
			{ID: "tester", Name: "Tester", Description: "Software tester"},
		},
	}
}

// LoadRoles loads roles from a category directory
func LoadRoles(projectDir, category string) (*RoleFile, error) {
	dirPath := getVoiceDir(projectDir, category)
	filePath := filepath.Join(dirPath, "roles.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults if file doesn't exist
			switch category {
			case "voc":
				return DefaultVoCRoles(), nil
			case "vos":
				return DefaultVoSRoles(), nil
			case "vob":
				return DefaultVoBRoles(), nil
			case "voe":
				return DefaultVoERoles(), nil
			default:
				return nil, fmt.Errorf("unknown category: %s", category)
			}
		}
		return nil, fmt.Errorf("failed to read roles.json: %w", err)
	}

	var roleFile RoleFile
	if err := json.Unmarshal(data, &roleFile); err != nil {
		return nil, fmt.Errorf("failed to parse roles.json: %w", err)
	}

	return &roleFile, nil
}

// SaveRoles saves roles to a category directory
func SaveRoles(projectDir, category string, roleFile *RoleFile) error {
	dirPath := getVoiceDir(projectDir, category)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dirPath, "roles.json")

	data, err := json.MarshalIndent(roleFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal roles: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write roles.json: %w", err)
	}

	return nil
}

// InitializeRoles creates default role files if they don't exist
func InitializeRoles(projectDir string) error {
	categories := []struct {
		name     string
		defaults func() *RoleFile
	}{
		{"voc", DefaultVoCRoles},
		{"vos", DefaultVoSRoles},
		{"vob", DefaultVoBRoles},
		{"voe", DefaultVoERoles},
	}

	for _, cat := range categories {
		// Use getVoiceDir to find existing QFD or basic directory
		dirPath := getVoiceDir(projectDir, cat.name)
		filePath := filepath.Join(dirPath, "roles.json")

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create %s directory: %w", cat.name, err)
			}

			if err := SaveRoles(projectDir, cat.name, cat.defaults()); err != nil {
				return fmt.Errorf("failed to save %s roles: %w", cat.name, err)
			}
			// Show relative path from projectDir
			relPath, _ := filepath.Rel(projectDir, filePath)
			fmt.Printf("  ✓ Created %s\n", relPath)
		}
	}

	return nil
}

// ValidateRole checks if a role is valid for a category
func ValidateRole(projectDir, category, roleID string) (bool, error) {
	roles, err := LoadRoles(projectDir, category)
	if err != nil {
		return false, err
	}

	for _, role := range roles.Roles {
		if role.ID == roleID {
			return true, nil
		}
	}

	return false, nil
}

// GetRoleDefinition returns a role definition by ID
func GetRoleDefinition(projectDir, category, roleID string) (*RoleDefinition, error) {
	roles, err := LoadRoles(projectDir, category)
	if err != nil {
		return nil, err
	}

	for _, role := range roles.Roles {
		if role.ID == roleID {
			return &role, nil
		}
	}

	return nil, fmt.Errorf("role '%s' not found in category '%s'", roleID, category)
}

// PrintRoles prints roles for a category
func PrintRoles(roleFile *RoleFile) {
	if len(roleFile.Roles) == 0 {
		fmt.Println("No roles defined.")
		return
	}

	// Sort roles by ID
	sorted := make([]RoleDefinition, len(roleFile.Roles))
	copy(sorted, roleFile.Roles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	fmt.Printf("%-20s %-25s %s\n", "ID", "Name", "Description")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, role := range sorted {
		fmt.Printf("%-20s %-25s %s\n", role.ID, role.Name, role.Description)
	}
}

// GetCategoryName returns human-readable category name
func GetCategoryName(category string) string {
	switch category {
	case "voc":
		return "VoC (Voice of Customer)"
	case "vos":
		return "VoS (Voice of Stakeholder)"
	case "vob":
		return "VoB (Voice of Business)"
	case "voe":
		return "VoE (Voice of Engineer)"
	default:
		return category
	}
}
