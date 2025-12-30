package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Permission represents a specific permission
type Permission string

const (
	// Playbook permissions
	PermissionPlaybookRead    Permission = "playbook:read"
	PermissionPlaybookWrite   Permission = "playbook:write"
	PermissionPlaybookExecute Permission = "playbook:execute"
	PermissionPlaybookDelete  Permission = "playbook:delete"

	// Secret permissions
	PermissionSecretRead   Permission = "secret:read"
	PermissionSecretWrite  Permission = "secret:write"
	PermissionSecretDelete Permission = "secret:delete"

	// Environment permissions
	PermissionEnvLocal     Permission = "env:local"
	PermissionEnvContainer Permission = "env:container"
	PermissionEnvVirt      Permission = "env:virt"

	// System permissions
	PermissionSystemAdmin Permission = "system:admin"
	PermissionSystemAudit Permission = "system:audit"
	PermissionSystemMCP   Permission = "system:mcp"

	// CI/CD permissions
	PermissionCICDRead    Permission = "cicd:read"
	PermissionCICDWrite   Permission = "cicd:write"
	PermissionCICDExecute Permission = "cicd:execute"
)

// Role represents a collection of permissions
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	Environment []string     `json:"environment,omitempty"` // Environment restrictions
	Resources   []string     `json:"resources,omitempty"`   // Resource patterns (regex)
	CreatedAt   time.Time    `json:"created_at"`
	CreatedBy   string       `json:"created_by"`
}

// User represents a user with assigned roles
type User struct {
	Username    string            `json:"username"`
	FullName    string            `json:"full_name"`
	Email       string            `json:"email"`
	Roles       []string          `json:"roles"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
}

// AccessRequest represents a request for access to a resource
type AccessRequest struct {
	User        string            `json:"user"`
	Permission  Permission        `json:"permission"`
	Resource    string            `json:"resource,omitempty"`
	Environment string            `json:"environment"`
	Context     map[string]string `json:"context,omitempty"`
}

// AccessResult represents the result of an access check
type AccessResult struct {
	Granted      bool              `json:"granted"`
	Reason       string            `json:"reason"`
	RequiredRole string            `json:"required_role,omitempty"`
	MatchedRole  string            `json:"matched_role,omitempty"`
	Context      map[string]string `json:"context,omitempty"`
}

// RBACManager manages role-based access control
type RBACManager struct {
	config    *RBACConfig
	users     map[string]*User
	roles     map[string]*Role
	auditMgr  *AuditManager
	dataFile  string
	rolesFile string
}

// RBACConfig represents RBAC system configuration
type RBACConfig struct {
	Enabled              bool     `json:"enabled"`
	DefaultRole          string   `json:"default_role"`
	AdminUsers           []string `json:"admin_users"`
	EnvironmentIsolation bool     `json:"environment_isolation"`
	RequireAuthentication bool     `json:"require_authentication"`
	DataDir              string   `json:"data_dir"`
}

// RBACData represents the persisted RBAC data
type RBACData struct {
	Users map[string]*User `json:"users"`
	Roles map[string]*Role `json:"roles"`
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(config *RBACConfig, auditMgr *AuditManager) (*RBACManager, error) {
	if err := os.MkdirAll(config.DataDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create RBAC data directory: %w", err)
	}

	mgr := &RBACManager{
		config:    config,
		users:     make(map[string]*User),
		roles:     make(map[string]*Role),
		auditMgr:  auditMgr,
		dataFile:  filepath.Join(config.DataDir, "rbac-data.json"),
		rolesFile: filepath.Join(config.DataDir, "rbac-roles.json"),
	}

	// Load existing data
	if err := mgr.loadData(); err != nil {
		return nil, fmt.Errorf("failed to load RBAC data: %w", err)
	}

	// Initialize default roles if none exist
	if len(mgr.roles) == 0 {
		if err := mgr.initializeDefaultRoles(); err != nil {
			return nil, fmt.Errorf("failed to initialize default roles: %w", err)
		}
	}

	// Initialize admin users
	for _, adminUser := range config.AdminUsers {
		if _, exists := mgr.users[adminUser]; !exists {
			mgr.createUser(adminUser, "System Administrator", "", []string{"admin"}, true)
		}
	}

	return mgr, nil
}

// CheckAccess verifies if a user has permission to perform an action
func (rbac *RBACManager) CheckAccess(req *AccessRequest) *AccessResult {
	if !rbac.config.Enabled {
		return &AccessResult{
			Granted: true,
			Reason:  "RBAC disabled",
		}
	}

	// Get user
	user, exists := rbac.users[req.User]
	if !exists {
		rbac.auditMgr.LogRoleAccess(req.User, req.Environment, "unknown", string(req.Permission), false)
		return &AccessResult{
			Granted: false,
			Reason:  "User not found",
		}
	}

	if !user.Enabled {
		rbac.auditMgr.LogRoleAccess(req.User, req.Environment, "disabled", string(req.Permission), false)
		return &AccessResult{
			Granted: false,
			Reason:  "User account disabled",
		}
	}

	// Check each role assigned to the user
	for _, roleName := range user.Roles {
		role, exists := rbac.roles[roleName]
		if !exists {
			continue
		}

		// Check if role has the required permission
		if rbac.roleHasPermission(role, req.Permission) {
			// Check environment restrictions
			if !rbac.checkEnvironmentAccess(role, req.Environment) {
				continue
			}

			// Check resource restrictions
			if req.Resource != "" && !rbac.checkResourceAccess(role, req.Resource) {
				continue
			}

			// Access granted
			rbac.auditMgr.LogRoleAccess(req.User, req.Environment, roleName, string(req.Permission), true)
			return &AccessResult{
				Granted:     true,
				Reason:      "Permission granted",
				MatchedRole: roleName,
			}
		}
	}

	// Access denied
	rbac.auditMgr.LogRoleAccess(req.User, req.Environment, "none", string(req.Permission), false)
	return &AccessResult{
		Granted:      false,
		Reason:       "Insufficient permissions",
		RequiredRole: rbac.findRequiredRole(req.Permission),
	}
}

// CreateRole creates a new role
func (rbac *RBACManager) CreateRole(name, description, createdBy string, permissions []Permission, environments, resources []string) error {
	if _, exists := rbac.roles[name]; exists {
		return fmt.Errorf("role '%s' already exists", name)
	}

	role := &Role{
		Name:        name,
		Description: description,
		Permissions: permissions,
		Environment: environments,
		Resources:   resources,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
	}

	rbac.roles[name] = role
	return rbac.saveData()
}

// UpdateRole updates an existing role
func (rbac *RBACManager) UpdateRole(name string, permissions []Permission, environments, resources []string) error {
	role, exists := rbac.roles[name]
	if !exists {
		return fmt.Errorf("role '%s' not found", name)
	}

	role.Permissions = permissions
	role.Environment = environments
	role.Resources = resources

	return rbac.saveData()
}

// DeleteRole deletes a role
func (rbac *RBACManager) DeleteRole(name string) error {
	if _, exists := rbac.roles[name]; !exists {
		return fmt.Errorf("role '%s' not found", name)
	}

	// Check if any users have this role
	for _, user := range rbac.users {
		for _, roleName := range user.Roles {
			if roleName == name {
				return fmt.Errorf("cannot delete role '%s': assigned to user '%s'", name, user.Username)
			}
		}
	}

	delete(rbac.roles, name)
	return rbac.saveData()
}

// CreateUser creates a new user
func (rbac *RBACManager) CreateUser(username, fullName, email string, roles []string, enabled bool) error {
	return rbac.createUser(username, fullName, email, roles, enabled)
}

func (rbac *RBACManager) createUser(username, fullName, email string, roles []string, enabled bool) error {
	if _, exists := rbac.users[username]; exists {
		return fmt.Errorf("user '%s' already exists", username)
	}

	// Validate roles exist
	for _, roleName := range roles {
		if _, exists := rbac.roles[roleName]; !exists {
			return fmt.Errorf("role '%s' not found", roleName)
		}
	}

	user := &User{
		Username:  username,
		FullName:  fullName,
		Email:     email,
		Roles:     roles,
		Enabled:   enabled,
		CreatedAt: time.Now(),
	}

	rbac.users[username] = user
	return rbac.saveData()
}

// UpdateUser updates an existing user
func (rbac *RBACManager) UpdateUser(username string, roles []string, enabled bool) error {
	user, exists := rbac.users[username]
	if !exists {
		return fmt.Errorf("user '%s' not found", username)
	}

	// Validate roles exist
	for _, roleName := range roles {
		if _, exists := rbac.roles[roleName]; !exists {
			return fmt.Errorf("role '%s' not found", roleName)
		}
	}

	user.Roles = roles
	user.Enabled = enabled

	return rbac.saveData()
}

// DeleteUser deletes a user
func (rbac *RBACManager) DeleteUser(username string) error {
	if _, exists := rbac.users[username]; !exists {
		return fmt.Errorf("user '%s' not found", username)
	}

	delete(rbac.users, username)
	return rbac.saveData()
}

// GetUser retrieves a user
func (rbac *RBACManager) GetUser(username string) (*User, error) {
	user, exists := rbac.users[username]
	if !exists {
		return nil, fmt.Errorf("user '%s' not found", username)
	}
	return user, nil
}

// GetRole retrieves a role
func (rbac *RBACManager) GetRole(name string) (*Role, error) {
	role, exists := rbac.roles[name]
	if !exists {
		return nil, fmt.Errorf("role '%s' not found", name)
	}
	return role, nil
}

// ListUsers lists all users
func (rbac *RBACManager) ListUsers() []*User {
	users := make([]*User, 0, len(rbac.users))
	for _, user := range rbac.users {
		users = append(users, user)
	}
	return users
}

// ListRoles lists all roles
func (rbac *RBACManager) ListRoles() []*Role {
	roles := make([]*Role, 0, len(rbac.roles))
	for _, role := range rbac.roles {
		roles = append(roles, role)
	}
	return roles
}

// RecordLogin records a user login event
func (rbac *RBACManager) RecordLogin(username string) error {
	user, exists := rbac.users[username]
	if !exists {
		return fmt.Errorf("user '%s' not found", username)
	}

	now := time.Now()
	user.LastLoginAt = &now
	return rbac.saveData()
}

// Helper functions

func (rbac *RBACManager) roleHasPermission(role *Role, permission Permission) bool {
	for _, p := range role.Permissions {
		if p == permission {
			return true
		}
		// Check for wildcard permissions
		if strings.HasSuffix(string(p), ":*") {
			prefix := strings.TrimSuffix(string(p), "*")
			if strings.HasPrefix(string(permission), prefix) {
				return true
			}
		}
	}
	return false
}

func (rbac *RBACManager) checkEnvironmentAccess(role *Role, environment string) bool {
	if len(role.Environment) == 0 {
		return true // No environment restrictions
	}

	for _, env := range role.Environment {
		if env == "*" || env == environment {
			return true
		}
	}
	return false
}

func (rbac *RBACManager) checkResourceAccess(role *Role, resource string) bool {
	if len(role.Resources) == 0 {
		return true // No resource restrictions
	}

	for _, pattern := range role.Resources {
		if matched, _ := regexp.MatchString(pattern, resource); matched {
			return true
		}
	}
	return false
}

func (rbac *RBACManager) findRequiredRole(permission Permission) string {
	for _, role := range rbac.roles {
		if rbac.roleHasPermission(role, permission) {
			return role.Name
		}
	}
	return "admin" // Default fallback
}

func (rbac *RBACManager) initializeDefaultRoles() error {
	// Admin role - full access
	adminRole := &Role{
		Name:        "admin",
		Description: "Full system administrator",
		Permissions: []Permission{
			PermissionPlaybookRead, PermissionPlaybookWrite, PermissionPlaybookExecute, PermissionPlaybookDelete,
			PermissionSecretRead, PermissionSecretWrite, PermissionSecretDelete,
			PermissionEnvLocal, PermissionEnvContainer, PermissionEnvVirt,
			PermissionSystemAdmin, PermissionSystemAudit, PermissionSystemMCP,
			PermissionCICDRead, PermissionCICDWrite, PermissionCICDExecute,
		},
		CreatedAt: time.Now(),
		CreatedBy: "system",
	}

	// Developer role - development access
	developerRole := &Role{
		Name:        "developer",
		Description: "Standard developer access",
		Permissions: []Permission{
			PermissionPlaybookRead, PermissionPlaybookWrite, PermissionPlaybookExecute,
			PermissionSecretRead,
			PermissionEnvLocal, PermissionEnvContainer,
			PermissionSystemMCP,
			PermissionCICDRead, PermissionCICDWrite,
		},
		CreatedAt: time.Now(),
		CreatedBy: "system",
	}

	// Operator role - production execution
	operatorRole := &Role{
		Name:        "operator",
		Description: "Production operations access",
		Permissions: []Permission{
			PermissionPlaybookRead, PermissionPlaybookExecute,
			PermissionSecretRead,
			PermissionEnvVirt,
			PermissionCICDExecute,
		},
		Environment: []string{"production", "staging"},
		CreatedAt:   time.Now(),
		CreatedBy:   "system",
	}

	// Auditor role - read-only access
	auditorRole := &Role{
		Name:        "auditor",
		Description: "Audit and compliance access",
		Permissions: []Permission{
			PermissionPlaybookRead,
			PermissionSystemAudit,
			PermissionCICDRead,
		},
		CreatedAt: time.Now(),
		CreatedBy: "system",
	}

	rbac.roles["admin"] = adminRole
	rbac.roles["developer"] = developerRole
	rbac.roles["operator"] = operatorRole
	rbac.roles["auditor"] = auditorRole

	return rbac.saveData()
}

func (rbac *RBACManager) loadData() error {
	// Load users and roles from data file
	if _, err := os.Stat(rbac.dataFile); os.IsNotExist(err) {
		return nil // File doesn't exist yet, that's OK
	}

	data, err := os.ReadFile(rbac.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read RBAC data file: %w", err)
	}

	var rbacData RBACData
	if err := json.Unmarshal(data, &rbacData); err != nil {
		return fmt.Errorf("failed to unmarshal RBAC data: %w", err)
	}

	rbac.users = rbacData.Users
	rbac.roles = rbacData.Roles

	// Initialize maps if nil
	if rbac.users == nil {
		rbac.users = make(map[string]*User)
	}
	if rbac.roles == nil {
		rbac.roles = make(map[string]*Role)
	}

	return nil
}

func (rbac *RBACManager) saveData() error {
	rbacData := &RBACData{
		Users: rbac.users,
		Roles: rbac.roles,
	}

	data, err := json.MarshalIndent(rbacData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal RBAC data: %w", err)
	}

	if err := os.WriteFile(rbac.dataFile, data, 0640); err != nil {
		return fmt.Errorf("failed to write RBAC data file: %w", err)
	}

	return nil
}

// GetDefaultRBACConfig returns default RBAC configuration
func GetDefaultRBACConfig() *RBACConfig {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".portunix", "rbac")

	return &RBACConfig{
		Enabled:               true,
		DefaultRole:           "developer",
		AdminUsers:            []string{}, // Will be populated from environment
		EnvironmentIsolation:  true,
		RequireAuthentication: true,
		DataDir:               dataDir,
	}
}

// RequirePermission is a middleware function that checks if user has required permission
func RequirePermission(rbac *RBACManager, user, environment string, permission Permission) func() *AccessResult {
	return func() *AccessResult {
		req := &AccessRequest{
			User:        user,
			Permission:  permission,
			Environment: environment,
		}
		return rbac.CheckAccess(req)
	}
}