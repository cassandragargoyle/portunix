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

// Category represents a category within an area
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Order       int    `json:"order,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CategoryRegistry holds all categories for an area
type CategoryRegistry struct {
	Version    string     `json:"version"`
	Area       string     `json:"area"`
	Categories []Category `json:"categories"`
	UpdatedAt  string     `json:"updated_at"`
}

// ValidAreaNames defines valid area identifiers
var ValidAreaNames = []string{"voc", "vos", "vob", "voe"}

// IsValidArea checks if area name is valid
func IsValidArea(area string) bool {
	for _, valid := range ValidAreaNames {
		if area == valid {
			return true
		}
	}
	return false
}

// ValidateCategoryID checks if category ID is valid (alphanumeric with hyphens, case insensitive)
func ValidateCategoryID(id string) error {
	if id == "" {
		return fmt.Errorf("category ID cannot be empty")
	}
	if len(id) > 50 {
		return fmt.Errorf("category ID too long (max 50 characters)")
	}
	// Accept both lowercase and uppercase (will be normalized to uppercase)
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`)
	if !pattern.MatchString(id) {
		return fmt.Errorf("category ID must be alphanumeric with hyphens (e.g., 'USER-AUTH' or 'user-auth')")
	}
	return nil
}

// NormalizeCategoryID converts category ID to uppercase
func NormalizeCategoryID(id string) string {
	return strings.ToUpper(id)
}

// ValidateHexColor checks if color is valid hex format
func ValidateHexColor(color string) error {
	if color == "" {
		return nil // empty is valid
	}
	pattern := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	if !pattern.MatchString(color) {
		return fmt.Errorf("color must be hex format (e.g., '#3B82F6')")
	}
	return nil
}

// GetCategoriesFilePath returns the path to categories.json for an area
func GetCategoriesFilePath(projectDir, area string) string {
	return filepath.Join(getVoiceDir(projectDir, area), "categories.json")
}

// LoadCategoryRegistry loads category registry from an area directory
func LoadCategoryRegistry(projectDir, area string) (*CategoryRegistry, error) {
	if !IsValidArea(area) {
		return nil, fmt.Errorf("invalid area: %s (valid: %s)", area, strings.Join(ValidAreaNames, ", "))
	}

	filePath := GetCategoriesFilePath(projectDir, area)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty registry if file doesn't exist
			return &CategoryRegistry{
				Version:    "1.0",
				Area:       area,
				Categories: []Category{},
				UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
			}, nil
		}
		return nil, fmt.Errorf("failed to read categories file: %w", err)
	}

	var registry CategoryRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse categories file: %w", err)
	}

	return &registry, nil
}

// SaveCategoryRegistry saves category registry to an area directory
func SaveCategoryRegistry(projectDir, area string, registry *CategoryRegistry) error {
	if !IsValidArea(area) {
		return fmt.Errorf("invalid area: %s", area)
	}

	// Ensure area directory exists
	areaDir := getVoiceDir(projectDir, area)
	if err := os.MkdirAll(areaDir, 0755); err != nil {
		return fmt.Errorf("failed to create area directory: %w", err)
	}

	registry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	filePath := GetCategoriesFilePath(projectDir, area)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write categories file: %w", err)
	}

	return nil
}

// AddCategory adds a new category to the registry
func (r *CategoryRegistry) AddCategory(cat Category) error {
	if err := ValidateCategoryID(cat.ID); err != nil {
		return err
	}
	if err := ValidateHexColor(cat.Color); err != nil {
		return err
	}
	if cat.Name == "" {
		return fmt.Errorf("category name cannot be empty")
	}

	// Normalize ID to uppercase
	cat.ID = NormalizeCategoryID(cat.ID)

	// Check for duplicate ID
	for _, existing := range r.Categories {
		if existing.ID == cat.ID {
			return fmt.Errorf("category with ID '%s' already exists", cat.ID)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	cat.CreatedAt = now
	cat.UpdatedAt = now

	// Set order if not specified
	if cat.Order == 0 {
		cat.Order = len(r.Categories) + 1
	}

	r.Categories = append(r.Categories, cat)
	return nil
}

// RemoveCategory removes a category from the registry
func (r *CategoryRegistry) RemoveCategory(id string) error {
	normalizedID := NormalizeCategoryID(id)
	for i, cat := range r.Categories {
		if cat.ID == normalizedID {
			r.Categories = append(r.Categories[:i], r.Categories[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("category '%s' not found", normalizedID)
}

// GetCategory returns a category by ID
func (r *CategoryRegistry) GetCategory(id string) (*Category, error) {
	normalizedID := NormalizeCategoryID(id)
	for i := range r.Categories {
		if r.Categories[i].ID == normalizedID {
			return &r.Categories[i], nil
		}
	}
	return nil, fmt.Errorf("category '%s' not found", normalizedID)
}

// UpdateCategory updates a category in the registry
func (r *CategoryRegistry) UpdateCategory(id string, updates Category) error {
	normalizedID := NormalizeCategoryID(id)
	for i := range r.Categories {
		if r.Categories[i].ID == normalizedID {
			if updates.Name != "" {
				r.Categories[i].Name = updates.Name
			}
			if updates.Description != "" {
				r.Categories[i].Description = updates.Description
			}
			if updates.Color != "" {
				if err := ValidateHexColor(updates.Color); err != nil {
					return err
				}
				r.Categories[i].Color = updates.Color
			}
			if updates.Order > 0 {
				r.Categories[i].Order = updates.Order
			}
			r.Categories[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("category '%s' not found", normalizedID)
}

// HasCategory checks if a category exists
func (r *CategoryRegistry) HasCategory(id string) bool {
	normalizedID := NormalizeCategoryID(id)
	for _, cat := range r.Categories {
		if cat.ID == normalizedID {
			return true
		}
	}
	return false
}

// CountItemsInCategory counts how many items have this category assigned
func CountItemsInCategory(projectDir, area, categoryID string) (int, error) {
	areaDir := filepath.Join(projectDir, area)
	items, err := ScanFeedbackDirectory(areaDir, area)
	if err != nil {
		return 0, err
	}

	normalizedID := NormalizeCategoryID(categoryID)
	count := 0
	for _, item := range items {
		for _, cat := range item.Categories {
			if NormalizeCategoryID(cat) == normalizedID {
				count++
				break
			}
		}
	}
	return count, nil
}

// GetAllCategoriesWithCounts returns all categories with item counts
func GetAllCategoriesWithCounts(projectDir, area string) ([]CategoryWithCount, error) {
	registry, err := LoadCategoryRegistry(projectDir, area)
	if err != nil {
		return nil, err
	}

	result := make([]CategoryWithCount, len(registry.Categories))
	for i, cat := range registry.Categories {
		count, err := CountItemsInCategory(projectDir, area, cat.ID)
		if err != nil {
			count = 0 // ignore errors in counting
		}
		result[i] = CategoryWithCount{
			Category: cat,
			Count:    count,
		}
	}

	return result, nil
}

// CategoryWithCount wraps Category with item count
type CategoryWithCount struct {
	Category
	Count int `json:"count"`
}
