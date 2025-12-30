package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RoleAssignment represents a role in a specific category
type RoleAssignment struct {
	Role  string `json:"role"`
	Proxy bool   `json:"proxy"`
}

// UserRoles contains role assignments for each category
type UserRoles struct {
	VoC *RoleAssignment `json:"voc,omitempty"`
	VoS *RoleAssignment `json:"vos,omitempty"`
	VoB *RoleAssignment `json:"vob,omitempty"`
	VoE *RoleAssignment `json:"voe,omitempty"`
}

// ExternalIDs contains external system identifiers
type ExternalIDs struct {
	Fider int `json:"fider,omitempty"`
}

// User represents a user in the registry
type User struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Organization string       `json:"organization,omitempty"`
	ExternalIDs  *ExternalIDs `json:"external_ids,omitempty"`
	Roles        UserRoles    `json:"roles"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// UserRegistry contains all users
type UserRegistry struct {
	Users []User `json:"users"`
}

// LoadUserRegistry loads users from JSON file
func LoadUserRegistry(projectDir string) (*UserRegistry, error) {
	filePath := filepath.Join(projectDir, "users.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &UserRegistry{Users: []User{}}, nil
		}
		return nil, fmt.Errorf("failed to read users.json: %w", err)
	}

	var registry UserRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse users.json: %w", err)
	}

	return &registry, nil
}

// SaveUserRegistry saves users to JSON file
func SaveUserRegistry(projectDir string, registry *UserRegistry) error {
	filePath := filepath.Join(projectDir, "users.json")

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write users.json: %w", err)
	}

	return nil
}

// FindUser finds a user by ID
func (r *UserRegistry) FindUser(id string) *User {
	for i := range r.Users {
		if r.Users[i].ID == id {
			return &r.Users[i]
		}
	}
	return nil
}

// FindUserByFiderID finds a user by Fider ID
func (r *UserRegistry) FindUserByFiderID(fiderID int) *User {
	for i := range r.Users {
		if r.Users[i].ExternalIDs != nil && r.Users[i].ExternalIDs.Fider == fiderID {
			return &r.Users[i]
		}
	}
	return nil
}

// FindUserByEmail finds a user by email (case-insensitive)
func (r *UserRegistry) FindUserByEmail(email string) *User {
	emailLower := strings.ToLower(email)
	for i := range r.Users {
		if strings.ToLower(r.Users[i].ID) == emailLower {
			return &r.Users[i]
		}
	}
	return nil
}

// FindUserByName finds a user by name (case-insensitive)
func (r *UserRegistry) FindUserByName(name string) *User {
	nameLower := strings.ToLower(name)
	for i := range r.Users {
		if strings.ToLower(r.Users[i].Name) == nameLower {
			return &r.Users[i]
		}
	}
	return nil
}

// GetRoleForArea returns the role for a specific area
func (u *User) GetRoleForArea(area string) string {
	switch strings.ToLower(area) {
	case "voc":
		if u.Roles.VoC != nil {
			return u.Roles.VoC.Role
		}
	case "vos":
		if u.Roles.VoS != nil {
			return u.Roles.VoS.Role
		}
	case "vob":
		if u.Roles.VoB != nil {
			return u.Roles.VoB.Role
		}
	case "voe":
		if u.Roles.VoE != nil {
			return u.Roles.VoE.Role
		}
	}
	return ""
}

// AddUser adds a new user to the registry
func (r *UserRegistry) AddUser(user User) error {
	if r.FindUser(user.ID) != nil {
		return fmt.Errorf("user '%s' already exists", user.ID)
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	r.Users = append(r.Users, user)
	return nil
}

// RemoveUser removes a user from the registry
func (r *UserRegistry) RemoveUser(id string) error {
	for i, user := range r.Users {
		if user.ID == id {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("user '%s' not found", id)
}

// UpdateUser updates an existing user
func (r *UserRegistry) UpdateUser(id string, updateFn func(*User)) error {
	user := r.FindUser(id)
	if user == nil {
		return fmt.Errorf("user '%s' not found", id)
	}

	updateFn(user)
	user.UpdatedAt = time.Now()
	return nil
}

// ListUsersByCategory returns users that have a role in the specified category
func (r *UserRegistry) ListUsersByCategory(category string) []User {
	var result []User
	for _, user := range r.Users {
		var hasRole bool
		switch category {
		case "voc":
			hasRole = user.Roles.VoC != nil
		case "vos":
			hasRole = user.Roles.VoS != nil
		case "vob":
			hasRole = user.Roles.VoB != nil
		case "voe":
			hasRole = user.Roles.VoE != nil
		default:
			continue
		}
		if hasRole {
			result = append(result, user)
		}
	}
	return result
}

// SetRole sets a role for a user in a specific category
func (u *User) SetRole(category, role string, proxy bool) error {
	assignment := &RoleAssignment{Role: role, Proxy: proxy}

	switch category {
	case "voc":
		u.Roles.VoC = assignment
	case "vos":
		u.Roles.VoS = assignment
	case "vob":
		u.Roles.VoB = assignment
	case "voe":
		u.Roles.VoE = assignment
	default:
		return fmt.Errorf("unknown category: %s", category)
	}

	u.UpdatedAt = time.Now()
	return nil
}

// RemoveRole removes a role from a specific category
func (u *User) RemoveRole(category string) error {
	switch category {
	case "voc":
		u.Roles.VoC = nil
	case "vos":
		u.Roles.VoS = nil
	case "vob":
		u.Roles.VoB = nil
	case "voe":
		u.Roles.VoE = nil
	default:
		return fmt.Errorf("unknown category: %s", category)
	}

	u.UpdatedAt = time.Now()
	return nil
}

// LinkFider links a Fider ID to the user
func (u *User) LinkFider(fiderID int) {
	if u.ExternalIDs == nil {
		u.ExternalIDs = &ExternalIDs{}
	}
	u.ExternalIDs.Fider = fiderID
	u.UpdatedAt = time.Now()
}

// PrintUser prints user details
func PrintUser(user *User) {
	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Name: %s\n", user.Name)
	if user.Organization != "" {
		fmt.Printf("Organization: %s\n", user.Organization)
	}
	if user.ExternalIDs != nil && user.ExternalIDs.Fider > 0 {
		fmt.Printf("Fider ID: %d\n", user.ExternalIDs.Fider)
	}
	fmt.Println("Roles:")
	if user.Roles.VoC != nil {
		proxyStr := ""
		if user.Roles.VoC.Proxy {
			proxyStr = " (proxy)"
		}
		fmt.Printf("  VoC: %s%s\n", user.Roles.VoC.Role, proxyStr)
	}
	if user.Roles.VoS != nil {
		proxyStr := ""
		if user.Roles.VoS.Proxy {
			proxyStr = " (proxy)"
		}
		fmt.Printf("  VoS: %s%s\n", user.Roles.VoS.Role, proxyStr)
	}
	if user.Roles.VoB != nil {
		proxyStr := ""
		if user.Roles.VoB.Proxy {
			proxyStr = " (proxy)"
		}
		fmt.Printf("  VoB: %s%s\n", user.Roles.VoB.Role, proxyStr)
	}
	if user.Roles.VoE != nil {
		proxyStr := ""
		if user.Roles.VoE.Proxy {
			proxyStr = " (proxy)"
		}
		fmt.Printf("  VoE: %s%s\n", user.Roles.VoE.Role, proxyStr)
	}
}

// PrintUserList prints a list of users in table format
func PrintUserList(users []User, category string) {
	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	fmt.Printf("%-30s %-20s %-15s %-10s\n", "ID", "Name", "Role", "Proxy")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, user := range users {
		var role string
		var proxy bool

		switch category {
		case "voc":
			if user.Roles.VoC != nil {
				role = user.Roles.VoC.Role
				proxy = user.Roles.VoC.Proxy
			}
		case "vos":
			if user.Roles.VoS != nil {
				role = user.Roles.VoS.Role
				proxy = user.Roles.VoS.Proxy
			}
		case "vob":
			if user.Roles.VoB != nil {
				role = user.Roles.VoB.Role
				proxy = user.Roles.VoB.Proxy
			}
		case "voe":
			if user.Roles.VoE != nil {
				role = user.Roles.VoE.Role
				proxy = user.Roles.VoE.Proxy
			}
		default:
			// Show all roles
			roles := []string{}
			if user.Roles.VoC != nil {
				roles = append(roles, "VoC:"+user.Roles.VoC.Role)
			}
			if user.Roles.VoS != nil {
				roles = append(roles, "VoS:"+user.Roles.VoS.Role)
			}
			if user.Roles.VoB != nil {
				roles = append(roles, "VoB:"+user.Roles.VoB.Role)
			}
			if user.Roles.VoE != nil {
				roles = append(roles, "VoE:"+user.Roles.VoE.Role)
			}
			for i, r := range roles {
				if i == 0 {
					fmt.Printf("%-30s %-20s %s\n", user.ID, user.Name, r)
				} else {
					fmt.Printf("%-30s %-20s %s\n", "", "", r)
				}
			}
			continue
		}

		proxyStr := "no"
		if proxy {
			proxyStr = "yes"
		}
		fmt.Printf("%-30s %-20s %-15s %-10s\n", user.ID, user.Name, role, proxyStr)
	}
}
