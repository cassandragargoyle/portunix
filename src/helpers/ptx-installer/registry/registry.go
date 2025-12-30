package registry

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EmbeddedAssetsFS holds the embedded assets filesystem (set from main package)
var EmbeddedAssetsFS embed.FS

// SetEmbeddedAssets sets the embedded assets filesystem from main package
func SetEmbeddedAssets(assetsFS embed.FS) {
	EmbeddedAssetsFS = assetsFS
}

// PackageRegistry represents the new registry system
type PackageRegistry struct {
	packages   map[string]*Package
	categories map[string]*Category
	index      *RegistryIndex
	assetsPath string
}

// Package represents a package definition in the new format
type Package struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   Metadata     `json:"metadata"`
	Spec       PackageSpec  `json:"spec"`
}

// Metadata contains package metadata
type Metadata struct {
	Name          string    `json:"name"`
	DisplayName   string    `json:"displayName"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Homepage      string    `json:"homepage,omitempty"`
	Documentation string    `json:"documentation,omitempty"`
	License       string    `json:"license,omitempty"`
	Maintainer    string    `json:"maintainer,omitempty"`
	Created       time.Time `json:"created,omitempty"`
	Updated       time.Time `json:"updated,omitempty"`
}

// PackageSpec contains the package specification
type PackageSpec struct {
	HasVariants  bool                      `json:"hasVariants"`
	Platforms    map[string]PlatformSpec   `json:"platforms"`
	Sources      map[string]SourceSpec     `json:"sources,omitempty"`
	Verification *VerificationSpec         `json:"verification,omitempty"`
	AIPrompts    *AIPrompts                `json:"aiPrompts,omitempty"`
	Dependencies []string                  `json:"dependencies,omitempty"`
	Templates    []string                  `json:"templates,omitempty"`
}

// PlatformSpec represents platform-specific configuration
type PlatformSpec struct {
	Type         string                   `json:"type"`
	Variants     map[string]VariantSpec   `json:"variants"`
	InstallArgs  []string                 `json:"installArgs,omitempty"`
	Environment  map[string]string        `json:"environment,omitempty"`
	Verification *VerificationSpec        `json:"verification,omitempty"`
}

// VariantSpec represents a specific variant of a package
type VariantSpec struct {
	Version       string                `json:"version"`
	Type          string                `json:"type,omitempty"`
	URL           string                `json:"url,omitempty"`
	URLs          map[string]string     `json:"urls,omitempty"`
	Packages      []string              `json:"packages,omitempty"`
	InstallScript string                `json:"installScript,omitempty"`
	InstallPath   string                `json:"installPath,omitempty"`
	ExtractTo     string                `json:"extractTo,omitempty"`
	Extract       bool                  `json:"extract,omitempty"`
	Binary        string                `json:"binary,omitempty"`
	RequiresSudo  bool                  `json:"requiresSudo,omitempty"`
	PostInstall   []string              `json:"postInstall,omitempty"`
	InstallArgs   []string              `json:"installArgs,omitempty"`
	Distributions interface{}           `json:"distributions,omitempty"`
	Checksum      map[string]string     `json:"checksum,omitempty"`
}

// SourceSpec represents source information for a package
type SourceSpec struct {
	Type        string `json:"type"` // github, gitlab, direct, etc.
	URL         string `json:"url"`
	APIEndpoint string `json:"apiEndpoint,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
}

// VerificationSpec represents verification configuration
type VerificationSpec struct {
	Command          string `json:"command"`
	ExpectedExitCode int    `json:"expectedExitCode"`
	ChecksumType     string `json:"checksumType,omitempty"`
	ChecksumURL      string `json:"checksumUrl,omitempty"`
}

// AIPrompts contains AI-related prompts for automated maintenance
type AIPrompts struct {
	VersionDiscovery string `json:"versionDiscovery,omitempty"`
	UrlResolution    string `json:"urlResolution,omitempty"`
	ChangeDetection  string `json:"changeDetection,omitempty"`
	UpdateGuidance   string `json:"updateGuidance,omitempty"`
}

// RegistryIndex represents the registry index
type RegistryIndex struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   IndexMetadata    `json:"metadata"`
	Spec       RegistryIndexSpec `json:"spec"`
}

// IndexMetadata contains registry index metadata
type IndexMetadata struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

// RegistryIndexSpec contains the registry index specification
type RegistryIndexSpec struct {
	Packages               []string `json:"packages"`
	Categories             []string `json:"categories"`
	SupportedPlatforms     []string `json:"supportedPlatforms"`
	SupportedArchitectures []string `json:"supportedArchitectures"`
}

// Category represents a package category
type Category struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon,omitempty"`
}

// CategoryIndex represents the categories index
type CategoryIndex struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   IndexMetadata       `json:"metadata"`
	Categories map[string]Category `json:"categories"`
}

// LoadPackageRegistry loads the package registry from the assets directory
// Priority 1: Try embedded assets (production/container mode)
// Priority 2: Fallback to external assets (development mode)
func LoadPackageRegistry(assetsPath string) (*PackageRegistry, error) {
	// Priority 1: Try embedded assets first (production/container mode)
	if registry, err := loadFromEmbedded(); err == nil {
		fmt.Printf("Package registry loaded from embedded assets\n")
		return registry, nil
	} else {
		fmt.Printf("Embedded assets not available, trying external assets: %v\n", err)
	}

	// Priority 2: Fallback to external assets (development mode)
	registry := &PackageRegistry{
		packages:   make(map[string]*Package),
		categories: make(map[string]*Category),
		assetsPath: assetsPath,
	}

	// Load individual packages first (automatic directory scanning)
	packagesDir := filepath.Join(assetsPath, "packages")
	if err := registry.loadPackages(packagesDir); err != nil {
		return nil, fmt.Errorf("failed to load packages from external assets: %w", err)
	}

	// Load categories (optional, graceful degradation if missing)
	categoriesPath := filepath.Join(assetsPath, "registry", "categories.json")
	if err := registry.loadCategories(categoriesPath); err != nil {
		// Log warning but continue - categories are optional
		fmt.Printf("Warning: Failed to load categories from external assets: %v\n", err)
	}

	// Generate index automatically from discovered packages
	registry.generateIndex()

	fmt.Printf("Package registry loaded from external assets\n")
	return registry, nil
}

// loadIndex loads the registry index
func (r *PackageRegistry) loadIndex(indexPath string) error {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Index doesn't exist yet, create empty one
		r.index = &RegistryIndex{
			APIVersion: "v1",
			Kind:       "PackageIndex",
			Metadata: IndexMetadata{
				Name:        "portunix-registry",
				Version:     "1.0.0",
				Description: "Portunix Package Registry Index",
				Created:     time.Now(),
			},
			Spec: RegistryIndexSpec{
				Packages:               []string{},
				Categories:             []string{},
				SupportedPlatforms:     []string{"windows", "linux", "darwin"},
				SupportedArchitectures: []string{"amd64", "arm64", "386"},
			},
		}
		return nil
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse registry index: %w", err)
	}

	r.index = &index
	return nil
}

// loadCategories loads the categories
func (r *PackageRegistry) loadCategories(categoriesPath string) error {
	if _, err := os.Stat(categoriesPath); os.IsNotExist(err) {
		// Categories file doesn't exist, use empty categories
		return nil
	}

	data, err := os.ReadFile(categoriesPath)
	if err != nil {
		return err
	}

	var categoryIndex CategoryIndex
	if err := json.Unmarshal(data, &categoryIndex); err != nil {
		return fmt.Errorf("failed to parse categories: %w", err)
	}

	for name, category := range categoryIndex.Categories {
		r.categories[name] = &category
	}

	return nil
}

// loadPackages loads all packages from the packages directory
func (r *PackageRegistry) loadPackages(packagesDir string) error {
	if _, err := os.Stat(packagesDir); os.IsNotExist(err) {
		// Packages directory doesn't exist yet
		return nil
	}

	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return err
	}

	loadedCount := 0
	errorCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		packagePath := filepath.Join(packagesDir, entry.Name())
		if err := r.loadPackage(packagePath); err != nil {
			// Log error but continue with other packages (graceful degradation)
			fmt.Printf("Warning: Failed to load package %s: %v\n", entry.Name(), err)
			errorCount++
			continue
		}
		loadedCount++
	}

	fmt.Printf("Package discovery complete: %d packages loaded, %d errors\n", loadedCount, errorCount)

	// Only return error if no packages were loaded at all
	if loadedCount == 0 && errorCount > 0 {
		return fmt.Errorf("failed to load any packages: %d errors encountered", errorCount)
	}

	return nil
}

// loadPackage loads a single package
func (r *PackageRegistry) loadPackage(packagePath string) error {
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}

	var pkg Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	// Validate package
	if err := r.validatePackage(&pkg); err != nil {
		return fmt.Errorf("package validation failed: %w", err)
	}

	r.packages[pkg.Metadata.Name] = &pkg
	return nil
}

// validatePackage validates a package definition
func (r *PackageRegistry) validatePackage(pkg *Package) error {
	// Basic structure validation
	if pkg.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if pkg.APIVersion != "v1" {
		return fmt.Errorf("apiVersion must be 'v1', got '%s'", pkg.APIVersion)
	}
	if pkg.Kind != "Package" {
		return fmt.Errorf("kind must be 'Package', got '%s'", pkg.Kind)
	}

	// Metadata validation
	if err := r.validateMetadata(&pkg.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Spec validation
	if err := r.validatePackageSpec(&pkg.Spec); err != nil {
		return fmt.Errorf("spec validation failed: %w", err)
	}

	return nil
}

// validateMetadata validates package metadata
func (r *PackageRegistry) validateMetadata(metadata *Metadata) error {
	if metadata.Name == "" {
		return fmt.Errorf("name is required")
	}
	if metadata.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}
	if metadata.Description == "" {
		return fmt.Errorf("description is required")
	}
	if metadata.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate package name format (lowercase, alphanumeric, hyphens only)
	if !isValidPackageName(metadata.Name) {
		return fmt.Errorf("name must be lowercase alphanumeric with hyphens only, got '%s'", metadata.Name)
	}

	// Validate category format
	if !isValidCategory(metadata.Category) {
		return fmt.Errorf("category must be in format 'group/subgroup', got '%s'", metadata.Category)
	}

	return nil
}

// validatePackageSpec validates package specification
func (r *PackageRegistry) validatePackageSpec(spec *PackageSpec) error {
	if len(spec.Platforms) == 0 {
		return fmt.Errorf("at least one platform must be specified")
	}

	// Validate platforms
	for platformName, platform := range spec.Platforms {
		if err := r.validatePlatform(platformName, &platform); err != nil {
			return fmt.Errorf("platform %s validation failed: %w", platformName, err)
		}
	}

	// Validate AI prompts if present
	if spec.AIPrompts != nil {
		if err := r.validateAIPrompts(spec.AIPrompts); err != nil {
			return fmt.Errorf("AI prompts validation failed: %w", err)
		}
	}

	return nil
}

// validatePlatform validates a platform configuration
func (r *PackageRegistry) validatePlatform(platformName string, platform *PlatformSpec) error {
	// Validate platform name
	validPlatforms := []string{"windows", "linux", "darwin"}
	if !containsString(validPlatforms, platformName) {
		return fmt.Errorf("unsupported platform '%s', must be one of: %v", platformName, validPlatforms)
	}

	// Validate platform type
	if platform.Type == "" {
		return fmt.Errorf("type is required")
	}

	validTypes := []string{"msi", "exe", "zip", "tar.gz", "deb", "rpm", "apt", "dnf", "pacman", "snap", "repository", "powershell", "script", "winget", "npm", "brew", "redirect"}
	if !containsString(validTypes, platform.Type) {
		return fmt.Errorf("unsupported type '%s', must be one of: %v", platform.Type, validTypes)
	}

	// Validate variants
	if len(platform.Variants) == 0 {
		return fmt.Errorf("at least one variant must be specified")
	}

	for variantName, variant := range platform.Variants {
		if err := r.validateVariant(variantName, &variant); err != nil {
			return fmt.Errorf("variant %s validation failed: %w", variantName, err)
		}
	}

	return nil
}

// validateVariant validates a variant configuration
func (r *PackageRegistry) validateVariant(variantName string, variant *VariantSpec) error {
	if variant.Version == "" {
		return fmt.Errorf("version is required")
	}

	// Ensure at least one installation method is specified
	hasInstallMethod := false
	if variant.URL != "" || len(variant.URLs) > 0 || len(variant.Packages) > 0 || variant.InstallScript != "" {
		hasInstallMethod = true
	}

	if !hasInstallMethod {
		return fmt.Errorf("variant must specify at least one installation method (url, urls, packages, or installScript)")
	}

	return nil
}

// validateAIPrompts validates AI prompts configuration
func (r *PackageRegistry) validateAIPrompts(prompts *AIPrompts) error {
	// AI prompts are optional but if present should be non-empty strings
	if prompts.VersionDiscovery != "" && len(prompts.VersionDiscovery) < 10 {
		return fmt.Errorf("versionDiscovery prompt is too short")
	}
	if prompts.UrlResolution != "" && len(prompts.UrlResolution) < 10 {
		return fmt.Errorf("urlResolution prompt is too short")
	}
	if prompts.UpdateGuidance != "" && len(prompts.UpdateGuidance) < 10 {
		return fmt.Errorf("updateGuidance prompt is too short")
	}

	return nil
}

// isValidPackageName checks if package name follows naming conventions
func isValidPackageName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return true
}

// isValidCategory checks if category follows the group/subgroup format
func isValidCategory(category string) bool {
	parts := strings.Split(category, "/")
	return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0
}

// containsString checks if a slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetPackage returns a package by name
func (r *PackageRegistry) GetPackage(name string) (*Package, error) {
	pkg, exists := r.packages[name]
	if !exists {
		return nil, fmt.Errorf("package '%s' not found", name)
	}
	return pkg, nil
}

// GetAllPackages returns all packages
func (r *PackageRegistry) GetAllPackages() map[string]*Package {
	return r.packages
}

// SearchPackages searches for packages matching the query
// Searches in package name, display name, description, and category
// Returns a slice of matching packages
func (r *PackageRegistry) SearchPackages(query string) []*Package {
	query = strings.ToLower(query)
	matches := make([]*Package, 0)

	for _, pkg := range r.packages {
		// Search in package name
		if strings.Contains(strings.ToLower(pkg.Metadata.Name), query) {
			matches = append(matches, pkg)
			continue
		}

		// Search in display name
		if strings.Contains(strings.ToLower(pkg.Metadata.DisplayName), query) {
			matches = append(matches, pkg)
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(pkg.Metadata.Description), query) {
			matches = append(matches, pkg)
			continue
		}

		// Search in category
		if strings.Contains(strings.ToLower(pkg.Metadata.Category), query) {
			matches = append(matches, pkg)
			continue
		}
	}

	return matches
}

// ResolveDependencies resolves package dependencies in correct installation order
func (r *PackageRegistry) ResolveDependencies(packageName string) ([]string, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]string, 0)

	err := r.resolveDependenciesRecursive(packageName, visited, visiting, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// resolveDependenciesRecursive performs recursive dependency resolution with cycle detection
func (r *PackageRegistry) resolveDependenciesRecursive(packageName string, visited, visiting map[string]bool, result *[]string) error {
	if visiting[packageName] {
		return fmt.Errorf("circular dependency detected involving package: %s", packageName)
	}

	if visited[packageName] {
		return nil
	}

	pkg, exists := r.packages[packageName]
	if !exists {
		return fmt.Errorf("package not found: %s", packageName)
	}

	visiting[packageName] = true

	// Process dependencies first
	for _, depName := range pkg.Spec.Dependencies {
		err := r.resolveDependenciesRecursive(depName, visited, visiting, result)
		if err != nil {
			return err
		}
	}

	visiting[packageName] = false
	visited[packageName] = true
	*result = append(*result, packageName)

	return nil
}

// GetPackageDependencies returns direct dependencies of a package
func (r *PackageRegistry) GetPackageDependencies(packageName string) ([]string, error) {
	pkg, exists := r.packages[packageName]
	if !exists {
		return nil, fmt.Errorf("package not found: %s", packageName)
	}

	return pkg.Spec.Dependencies, nil
}

// GetDependentPackages returns packages that depend on the given package
func (r *PackageRegistry) GetDependentPackages(packageName string) []string {
	dependents := make([]string, 0)

	for name, pkg := range r.packages {
		for _, dep := range pkg.Spec.Dependencies {
			if dep == packageName {
				dependents = append(dependents, name)
				break
			}
		}
	}

	return dependents
}

// GetCategory returns a category by name
func (r *PackageRegistry) GetCategory(name string) (*Category, error) {
	category, exists := r.categories[name]
	if !exists {
		return nil, fmt.Errorf("category '%s' not found", name)
	}
	return category, nil
}

// GetPackagesByCategory returns packages in a specific category
func (r *PackageRegistry) GetPackagesByCategory(categoryName string) ([]*Package, error) {
	var packages []*Package
	for _, pkg := range r.packages {
		if pkg.Metadata.Category == categoryName {
			packages = append(packages, pkg)
		}
	}
	return packages, nil
}

// NOTE: Legacy config conversion functions removed from ptx-installer
// The helper uses only the new registry system (ADR-021)
// Legacy config support remains in main binary for backward compatibility

// SavePackage saves a package to the registry
func (r *PackageRegistry) SavePackage(pkg *Package) error {
	// Validate package
	if err := r.validatePackage(pkg); err != nil {
		return fmt.Errorf("package validation failed: %w", err)
	}

	// Update package in memory
	r.packages[pkg.Metadata.Name] = pkg

	// Save to file
	packagePath := filepath.Join(r.assetsPath, "packages", pkg.Metadata.Name+".json")
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package: %w", err)
	}

	if err := os.WriteFile(packagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package file: %w", err)
	}

	// Update index
	r.updateIndex()

	return nil
}

// generateIndex automatically generates registry index from discovered packages
func (r *PackageRegistry) generateIndex() {
	r.index = &RegistryIndex{
		APIVersion: "v1",
		Kind:       "PackageIndex",
		Metadata: IndexMetadata{
			Name:        "portunix-registry",
			Version:     "1.0.0",
			Description: "Portunix Package Registry Index (Auto-Generated)",
			Created:     time.Now(),
		},
		Spec: RegistryIndexSpec{
			Packages:               []string{},
			Categories:             []string{},
			SupportedPlatforms:     []string{"windows", "linux", "darwin"},
			SupportedArchitectures: []string{"amd64", "arm64", "386"},
		},
	}

	// Build package list from discovered packages
	for name := range r.packages {
		r.index.Spec.Packages = append(r.index.Spec.Packages, name)
	}

	// Build categories list from discovered packages
	categoriesSet := make(map[string]bool)
	for _, pkg := range r.packages {
		if pkg.Metadata.Category != "" {
			categoriesSet[pkg.Metadata.Category] = true
		}
	}

	for category := range categoriesSet {
		r.index.Spec.Categories = append(r.index.Spec.Categories, category)
	}
}

// updateIndex updates the registry index with current packages
func (r *PackageRegistry) updateIndex() {
	if r.index == nil {
		r.generateIndex()
		return
	}

	// Clear and rebuild package list
	r.index.Spec.Packages = []string{}
	for name := range r.packages {
		r.index.Spec.Packages = append(r.index.Spec.Packages, name)
	}

	// Update categories
	categoriesSet := make(map[string]bool)
	for _, pkg := range r.packages {
		if pkg.Metadata.Category != "" {
			categoriesSet[pkg.Metadata.Category] = true
		}
	}

	r.index.Spec.Categories = []string{}
	for category := range categoriesSet {
		r.index.Spec.Categories = append(r.index.Spec.Categories, category)
	}
}

// SaveIndex saves the registry index
func (r *PackageRegistry) SaveIndex() error {
	if r.index == nil {
		return fmt.Errorf("no index to save")
	}

	indexPath := filepath.Join(r.assetsPath, "registry", "index.json")
	data, err := json.MarshalIndent(r.index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// loadFromEmbedded loads package registry from embedded assets
func loadFromEmbedded() (*PackageRegistry, error) {
	registry := &PackageRegistry{
		packages:   make(map[string]*Package),
		categories: make(map[string]*Category),
		assetsPath: "embedded",
	}

	// Load packages from embedded assets/packages/ directory
	if err := registry.loadPackagesFromEmbedded(); err != nil {
		return nil, fmt.Errorf("failed to load packages from embedded assets: %w", err)
	}

	// Load categories from embedded assets (optional)
	if err := registry.loadCategoriesFromEmbedded(); err != nil {
		// Log warning but continue - categories are optional
		fmt.Printf("Warning: Failed to load categories from embedded assets: %v\n", err)
	}

	// Generate index automatically from discovered packages
	registry.generateIndex()

	return registry, nil
}

// loadPackagesFromEmbedded loads all packages from embedded assets/packages/ directory
func (r *PackageRegistry) loadPackagesFromEmbedded() error {
	entries, err := EmbeddedAssetsFS.ReadDir("assets/packages")
	if err != nil {
		return fmt.Errorf("failed to read embedded packages directory: %w", err)
	}

	loadedCount := 0
	errorCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		if err := r.loadPackageFromEmbedded(entry.Name()); err != nil {
			// Log error but continue with other packages (graceful degradation)
			fmt.Printf("Warning: Failed to load embedded package %s: %v\n", entry.Name(), err)
			errorCount++
			continue
		}
		loadedCount++
	}

	fmt.Printf("Embedded package discovery complete: %d packages loaded, %d errors\n", loadedCount, errorCount)

	// Only return error if no packages were loaded at all
	if loadedCount == 0 && errorCount > 0 {
		return fmt.Errorf("failed to load any embedded packages: %d errors encountered", errorCount)
	}

	return nil
}

// loadPackageFromEmbedded loads a single package from embedded assets
func (r *PackageRegistry) loadPackageFromEmbedded(filename string) error {
	// Use forward slashes for embedded FS (cross-platform compatibility)
	packagePath := "assets/packages/" + filename
	data, err := EmbeddedAssetsFS.ReadFile(packagePath)
	if err != nil {
		return fmt.Errorf("failed to read embedded package file %s: %w", packagePath, err)
	}

	var pkg Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to parse embedded package: %w", err)
	}

	// Validate package
	if err := r.validatePackage(&pkg); err != nil {
		return fmt.Errorf("embedded package validation failed: %w", err)
	}

	r.packages[pkg.Metadata.Name] = &pkg
	return nil
}

// loadCategoriesFromEmbedded loads categories from embedded assets
func (r *PackageRegistry) loadCategoriesFromEmbedded() error {
	categoriesPath := "assets/registry/categories.json"

	// Check if categories file exists in embedded assets
	if _, err := fs.Stat(EmbeddedAssetsFS, categoriesPath); err != nil {
		// Categories file doesn't exist in embedded assets, use empty categories
		return nil
	}

	data, err := EmbeddedAssetsFS.ReadFile(categoriesPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded categories file: %w", err)
	}

	var categoryIndex CategoryIndex
	if err := json.Unmarshal(data, &categoryIndex); err != nil {
		return fmt.Errorf("failed to parse embedded categories: %w", err)
	}

	for name, category := range categoryIndex.Categories {
		r.categories[name] = &category
	}

	return nil
}