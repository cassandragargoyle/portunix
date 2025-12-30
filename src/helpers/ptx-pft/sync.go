package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ParseMarkdownFile parses a feedback markdown file
func ParseMarkdownFile(filePath string) (*FeedbackItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	item := &FeedbackItem{
		ID:       strings.TrimSuffix(filepath.Base(filePath), ".md"),
		FilePath: filePath,
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var descriptionLines []string
	var inDescription bool

	for scanner.Scan() {
		line := scanner.Text()

		// Parse main title (# heading)
		if strings.HasPrefix(line, "# ") && item.Title == "" {
			item.Title = strings.TrimPrefix(line, "# ")
			continue
		}

		// Parse section headers
		if strings.HasPrefix(line, "## ") {
			currentSection = strings.TrimPrefix(line, "## ")
			inDescription = currentSection == "Description"
			continue
		}

		// Parse section content
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			if inDescription {
				descriptionLines = append(descriptionLines, "")
			}
			continue
		}

		switch currentSection {
		case "Summary":
			if item.Summary == "" {
				item.Summary = trimmedLine
			}
		case "Priority":
			if item.Priority == "" {
				item.Priority = trimmedLine
			}
		case "Status":
			if item.Status == "" {
				item.Status = trimmedLine
			}
		case "Categories":
			// Parse comma-separated categories
			if len(item.Categories) == 0 && trimmedLine != "" {
				cats := strings.Split(trimmedLine, ",")
				for _, cat := range cats {
					cat = strings.TrimSpace(cat)
					if cat != "" {
						item.Categories = append(item.Categories, cat)
					}
				}
			}
		case "Description":
			descriptionLines = append(descriptionLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Build description for Fider
	var sb strings.Builder
	if item.Summary != "" {
		sb.WriteString(item.Summary)
		sb.WriteString("\n\n")
	}
	if len(descriptionLines) > 0 {
		sb.WriteString(strings.Join(descriptionLines, "\n"))
	}
	item.Description = strings.TrimSpace(sb.String())

	return item, nil
}

// ScanFeedbackDirectory scans a directory for feedback markdown files
// It recursively scans subdirectories (e.g., needs/, verbatims/) for QFD structure compatibility
func ScanFeedbackDirectory(dir string, feedbackType string) ([]*FeedbackItem, error) {
	var items []*FeedbackItem

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively scan subdirectories (QFD structure: needs/, verbatims/, etc.)
			subItems, err := ScanFeedbackDirectory(entryPath, feedbackType)
			if err != nil {
				fmt.Printf("Warning: failed to scan subdirectory %s: %v\n", entry.Name(), err)
				continue
			}
			items = append(items, subItems...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		item, err := ParseMarkdownFile(entryPath)
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		item.Type = feedbackType
		items = append(items, item)
	}

	return items, nil
}

// PushToFider pushes feedback items to Fider
func PushToFider(client *FiderClient, items []*FeedbackItem, dryRun bool) error {
	if len(items) == 0 {
		fmt.Println("No items to push")
		return nil
	}

	for _, item := range items {
		title := item.Title
		if title == "" {
			title = item.ID
		}

		// Clean title - remove ID prefix if present (e.g., "UC001: User Login" -> "User Login")
		re := regexp.MustCompile(`^[A-Z]+\d+:\s*`)
		cleanTitle := re.ReplaceAllString(title, "")
		if cleanTitle == "" {
			cleanTitle = title
		}

		if dryRun {
			fmt.Printf("  [DRY-RUN] Would create: %s\n", cleanTitle)
			fmt.Printf("            Description: %s...\n", truncate(item.Description, 50))
			continue
		}

		post, err := client.CreatePost(cleanTitle, item.Description)
		if err != nil {
			fmt.Printf("  ✗ Failed to create '%s': %v\n", cleanTitle, err)
			continue
		}

		fmt.Printf("  ✓ Created #%d: %s\n", post.Number, cleanTitle)
	}

	return nil
}

// truncate shortens a string to maxLen characters
func truncate(s string, maxLen int) string {
	// Remove newlines for display
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	Pushed   int
	Pulled   int
	Errors   int
	Skipped  int
}

// GenerateMarkdownFromPost creates markdown content from a Fider post
func GenerateMarkdownFromPost(post *FiderPost, feedbackType string) string {
	var sb strings.Builder

	// Generate ID prefix based on type
	prefix := "FB"
	if feedbackType == "voc" {
		prefix = "UC"
	} else if feedbackType == "vos" {
		prefix = "REQ"
	}

	// Title
	sb.WriteString(fmt.Sprintf("# %s%03d: %s\n\n", prefix, post.Number, post.Title))

	// Summary (first line of description)
	description := post.Description
	summary := description
	if idx := strings.Index(description, "\n"); idx > 0 {
		summary = description[:idx]
	}
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	sb.WriteString("## Summary\n")
	sb.WriteString(summary + "\n\n")

	// Priority (based on votes)
	sb.WriteString("## Priority\n")
	if post.VotesCount >= 10 {
		sb.WriteString("High\n\n")
	} else if post.VotesCount >= 5 {
		sb.WriteString("Medium\n\n")
	} else {
		sb.WriteString("Low\n\n")
	}

	// Status
	sb.WriteString("## Status\n")
	status := post.Status
	if status == "" {
		status = "Open"
	}
	sb.WriteString(status + "\n\n")

	// Description
	sb.WriteString("## Description\n")
	sb.WriteString(description + "\n\n")

	// Metadata
	sb.WriteString("## Metadata\n")
	sb.WriteString(fmt.Sprintf("- Fider ID: %d\n", post.Number))
	sb.WriteString(fmt.Sprintf("- Author: %s\n", post.User.Name))
	sb.WriteString(fmt.Sprintf("- Votes: %d\n", post.VotesCount))
	sb.WriteString(fmt.Sprintf("- Created: %s\n", post.CreatedAt.Format("2006-01-02")))

	return sb.String()
}

// FindNextAvailableNumber finds the next available number for UC/REQ files in a directory
func FindNextAvailableNumber(targetDir string, prefix string) int {
	maxNum := 0
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return 1
	}

	// Pattern: UC001, UC002, REQ001, etc.
	pattern := regexp.MustCompile(fmt.Sprintf(`^%s(\d+)-`, prefix))

	for _, entry := range entries {
		if matches := pattern.FindStringSubmatch(entry.Name()); len(matches) > 1 {
			if num, err := fmt.Sscanf(matches[1], "%d", new(int)); err == nil && num > 0 {
				var n int
				fmt.Sscanf(matches[1], "%d", &n)
				if n > maxNum {
					maxNum = n
				}
			}
		}
	}

	return maxNum + 1
}

// CreateSlugFromTitle creates a URL-friendly slug from a title, preserving Czech diacritics
func CreateSlugFromTitle(title string) string {
	slug := strings.ToLower(title)
	var result strings.Builder
	lastWasDash := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') ||
		   (r >= 'á' && r <= 'ž') || r == 'ě' || r == 'š' || r == 'č' || r == 'ř' ||
		   r == 'ž' || r == 'ý' || r == 'á' || r == 'í' || r == 'é' || r == 'ú' || r == 'ů' {
			result.WriteRune(r)
			lastWasDash = false
		} else if !lastWasDash {
			result.WriteRune('-')
			lastWasDash = true
		}
	}
	slug = strings.Trim(result.String(), "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	return slug
}

// GenerateFilenameWithNumber creates a filename with a specific number
func GenerateFilenameWithNumber(post *FiderPost, prefix string, num int) string {
	slug := CreateSlugFromTitle(post.Title)
	return fmt.Sprintf("%s%03d-%s.md", prefix, num, slug)
}

// GenerateFilename creates a filename from a Fider post (legacy, finds next available number)
func GenerateFilename(post *FiderPost, feedbackType string, targetDir string) string {
	prefix := "FB"
	if feedbackType == "voc" {
		prefix = "UC"
	} else if feedbackType == "vos" {
		prefix = "REQ"
	}

	slug := CreateSlugFromTitle(post.Title)
	nextNum := FindNextAvailableNumber(targetDir, prefix)

	return fmt.Sprintf("%s%03d-%s.md", prefix, nextNum, slug)
}

// HasFiderID checks if a feedback item has been synced with Fider (has Fider ID in metadata)
func HasFiderID(item *FeedbackItem) bool {
	if item.Metadata == nil {
		return false
	}
	_, ok := item.Metadata["fider_id"]
	return ok
}

// ExtractFiderID extracts Fider ID from file content if present
func ExtractFiderID(filePath string) (int, bool) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- Fider ID: ") {
			var id int
			if _, err := fmt.Sscanf(line, "- Fider ID: %d", &id); err == nil {
				return id, true
			}
		}
	}
	return 0, false
}

// UpdateFileWithFiderID updates a markdown file to add Fider ID in metadata
func UpdateFileWithFiderID(filePath string, fiderID int, authorName string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Check if file already has metadata section
	contentStr := string(content)
	if strings.Contains(contentStr, "## Metadata") {
		// Already has metadata, check if Fider ID exists
		if strings.Contains(contentStr, "- Fider ID:") {
			return nil // Already has Fider ID
		}
		// Add Fider ID to existing metadata
		contentStr = strings.Replace(contentStr, "## Metadata\n",
			fmt.Sprintf("## Metadata\n- Fider ID: %d\n- Author: %s\n- Synced: %s\n",
				fiderID, authorName, time.Now().Format("2006-01-02")), 1)
	} else {
		// Add metadata section at the end
		contentStr += fmt.Sprintf("\n## Metadata\n- Fider ID: %d\n- Author: %s\n- Synced: %s\n",
			fiderID, authorName, time.Now().Format("2006-01-02"))
	}

	return os.WriteFile(filePath, []byte(contentStr), 0644)
}

// PushNewToFider pushes only new (unsynced) local files to Fider
func PushNewToFider(client *FiderClient, items []*FeedbackItem, dryRun bool, authorName string) (int, int, error) {
	pushed := 0
	skipped := 0

	// Fetch existing posts from Fider to prevent duplicates
	existingPosts, err := client.ListPosts()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch existing posts: %w", err)
	}

	// Create map of existing post slugs for quick lookup
	existingSlugs := make(map[string]int)
	for _, post := range existingPosts {
		slug := CreateSlugFromTitle(post.Title)
		existingSlugs[slug] = post.Number
	}

	for _, item := range items {
		// Check if already synced (has Fider ID in metadata)
		if fiderID, hasFiderID := ExtractFiderID(item.FilePath); hasFiderID {
			if dryRun {
				fmt.Printf("  [SKIP] Already synced (Fider #%d): %s\n", fiderID, item.Title)
			}
			skipped++
			continue
		}

		// Check if similar post already exists in Fider (by slug match)
		title := item.Title
		if title == "" {
			title = item.ID
		}
		re := regexp.MustCompile(`^[A-Z]+\d+:\s*`)
		cleanTitle := re.ReplaceAllString(title, "")
		if cleanTitle == "" {
			cleanTitle = title
		}
		localSlug := CreateSlugFromTitle(cleanTitle)
		if fiderNum, exists := existingSlugs[localSlug]; exists {
			if dryRun {
				fmt.Printf("  [SKIP] Already in Fider (#%d): %s\n", fiderNum, cleanTitle)
			} else {
				// Update local file with Fider ID since it matches existing post
				if err := UpdateFileWithFiderID(item.FilePath, fiderNum, authorName); err != nil {
					fmt.Printf("  ⚠ Matched Fider #%d but failed to update local file: %v\n", fiderNum, err)
				} else {
					fmt.Printf("  ↔ Linked to existing Fider #%d: %s\n", fiderNum, cleanTitle)
				}
			}
			skipped++
			continue
		}

		// title and cleanTitle already set above for slug check

		if dryRun {
			fmt.Printf("  [NEW] Would push: %s\n", cleanTitle)
			pushed++
			continue
		}

		post, err := client.CreatePost(cleanTitle, item.Description)
		if err != nil {
			fmt.Printf("  ✗ Failed to push '%s': %v\n", cleanTitle, err)
			continue
		}

		// Update local file with Fider ID
		if err := UpdateFileWithFiderID(item.FilePath, post.Number, authorName); err != nil {
			fmt.Printf("  ⚠ Created #%d but failed to update local file: %v\n", post.Number, err)
		} else {
			fmt.Printf("  ✓ Pushed #%d: %s (local file updated)\n", post.Number, cleanTitle)
		}
		pushed++
	}

	return pushed, skipped, nil
}

// FindFileBySlug searches directory for a file whose name contains the given slug
func FindFileBySlug(dir string, slug string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Extract slug from filename (e.g., "UC001-user-login.md" -> "user-login")
		name := strings.TrimSuffix(entry.Name(), ".md")
		parts := strings.SplitN(name, "-", 2)
		if len(parts) == 2 {
			fileSlug := parts[1]
			// Compare slugs (case-insensitive, partial match)
			if strings.Contains(strings.ToLower(fileSlug), strings.ToLower(slug)) ||
			   strings.Contains(strings.ToLower(slug), strings.ToLower(fileSlug)) {
				return entry.Name(), true
			}
		}
	}
	return "", false
}

// FindFileWithFiderID searches directory for a file containing given Fider ID in metadata
func FindFileWithFiderID(dir string, fiderID int) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		if id, found := ExtractFiderID(filePath); found && id == fiderID {
			return entry.Name(), true
		}
	}
	return "", false
}

// PullFromFider pulls posts from Fider and saves them as markdown files
func PullFromFider(client *FiderClient, targetDir string, feedbackType string, dryRun bool) (int, int, error) {
	posts, err := client.ListPosts()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("   No posts found in Fider")
		return 0, 0, nil
	}

	// Ensure target directory exists
	if !dryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return 0, 0, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Get prefix for this feedback type
	prefix := "FB"
	if feedbackType == "voc" {
		prefix = "UC"
	} else if feedbackType == "vos" {
		prefix = "REQ"
	}

	// Find starting number for new files
	nextNum := FindNextAvailableNumber(targetDir, prefix)

	created := 0
	skipped := 0

	for _, post := range posts {
		// First check if any local file already has this Fider ID
		if existingFile, found := FindFileWithFiderID(targetDir, post.Number); found {
			if dryRun {
				fmt.Printf("  [DRY-RUN] Would skip (synced): %s (Fider #%d)\n", existingFile, post.Number)
			}
			skipped++
			continue
		}

		// Check if a file with similar title already exists (by slug match)
		postSlug := CreateSlugFromTitle(post.Title)
		if existingFile, found := FindFileBySlug(targetDir, postSlug); found {
			if dryRun {
				fmt.Printf("  [DRY-RUN] Would skip (exists): %s (matches '%s')\n", existingFile, post.Title)
			}
			skipped++
			continue
		}

		filename := GenerateFilenameWithNumber(&post, prefix, nextNum)
		nextNum++ // Increment for next file
		filePath := filepath.Join(targetDir, filename)

		// Check if file with this name already exists
		if _, err := os.Stat(filePath); err == nil {
			if dryRun {
				fmt.Printf("  [DRY-RUN] Would skip (exists): %s\n", filename)
			}
			skipped++
			continue
		}

		content := GenerateMarkdownFromPost(&post, feedbackType)

		if dryRun {
			fmt.Printf("  [DRY-RUN] Would create: %s\n", filename)
			fmt.Printf("            Title: %s\n", post.Title)
			created++
			continue
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			fmt.Printf("  ✗ Failed to write %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("  ✓ Created: %s\n", filename)
		created++
	}

	return created, skipped, nil
}

// ConflictDetector handles sync conflict detection
type ConflictDetector struct {
	Resolution ConflictResolution
}

// NewConflictDetector creates a new conflict detector with specified resolution strategy
func NewConflictDetector(resolution ConflictResolution) *ConflictDetector {
	return &ConflictDetector{Resolution: resolution}
}

// DetectConflict compares local and remote items to detect conflicts
func (cd *ConflictDetector) DetectConflict(local *FeedbackItem, remote *FeedbackItem) *SyncConflict {
	if local == nil || remote == nil {
		return nil
	}

	// Compare title and description for changes
	titleDiff := local.Title != remote.Title
	descDiff := local.Description != remote.Description
	statusDiff := local.Status != remote.Status

	if !titleDiff && !descDiff && !statusDiff {
		return nil // No conflict
	}

	reason := buildConflictReason(titleDiff, descDiff, statusDiff)

	return &SyncConflict{
		ItemID:     local.ID,
		LocalItem:  *local,
		RemoteItem: *remote,
		Reason:     reason,
	}
}

// buildConflictReason generates a human-readable conflict reason
func buildConflictReason(titleDiff, descDiff, statusDiff bool) string {
	var parts []string
	if titleDiff {
		parts = append(parts, "title")
	}
	if descDiff {
		parts = append(parts, "description")
	}
	if statusDiff {
		parts = append(parts, "status")
	}
	return fmt.Sprintf("modified fields: %s", strings.Join(parts, ", "))
}

// ResolveConflict resolves a conflict based on the configured strategy
func (cd *ConflictDetector) ResolveConflict(conflict *SyncConflict) (*FeedbackItem, string, error) {
	if conflict == nil {
		return nil, "", fmt.Errorf("no conflict to resolve")
	}

	switch cd.Resolution {
	case ConflictLocal:
		conflict.Resolution = "kept local version"
		return &conflict.LocalItem, "local", nil

	case ConflictRemote:
		conflict.Resolution = "accepted remote version"
		return &conflict.RemoteItem, "remote", nil

	case ConflictTimestamp:
		// Compare timestamps if available
		localTime := parseTimestamp(conflict.LocalItem.UpdatedAt)
		remoteTime := parseTimestamp(conflict.RemoteItem.UpdatedAt)

		if localTime.After(remoteTime) {
			conflict.Resolution = "kept local (newer)"
			return &conflict.LocalItem, "local", nil
		}
		conflict.Resolution = "accepted remote (newer)"
		return &conflict.RemoteItem, "remote", nil

	case ConflictManual:
		return nil, "", fmt.Errorf("manual resolution required")

	default:
		return nil, "", fmt.Errorf("unknown resolution strategy: %s", cd.Resolution)
	}
}

// parseTimestamp parses a timestamp string to time.Time
func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t
		}
	}

	return time.Time{}
}

// DetectAllConflicts compares local items with remote items and returns all conflicts
func DetectAllConflicts(localItems []*FeedbackItem, remoteItems []FeedbackItem) []SyncConflict {
	var conflicts []SyncConflict
	detector := NewConflictDetector(ConflictTimestamp)

	// Build map of remote items by ID and external ID for quick lookup
	remoteByID := make(map[string]*FeedbackItem)
	for i := range remoteItems {
		if remoteItems[i].ExternalID != "" {
			remoteByID[remoteItems[i].ExternalID] = &remoteItems[i]
		}
		if remoteItems[i].ID != "" {
			remoteByID[remoteItems[i].ID] = &remoteItems[i]
		}
	}

	for _, local := range localItems {
		// Try to find matching remote item
		var remote *FeedbackItem
		if local.ExternalID != "" {
			remote = remoteByID[local.ExternalID]
		}
		if remote == nil && local.ID != "" {
			remote = remoteByID[local.ID]
		}

		if remote == nil {
			continue // No remote match, no conflict
		}

		if conflict := detector.DetectConflict(local, remote); conflict != nil {
			conflicts = append(conflicts, *conflict)
		}
	}

	return conflicts
}

// PrintConflicts displays conflicts in a human-readable format
func PrintConflicts(conflicts []SyncConflict) {
	if len(conflicts) == 0 {
		fmt.Println("✓ No conflicts detected")
		return
	}

	fmt.Printf("⚠️  Found %d conflict(s):\n\n", len(conflicts))
	for i, c := range conflicts {
		fmt.Printf("%d. Item: %s\n", i+1, c.ItemID)
		fmt.Printf("   Reason: %s\n", c.Reason)
		fmt.Printf("   Local title:  %s\n", truncate(c.LocalItem.Title, 50))
		fmt.Printf("   Remote title: %s\n", truncate(c.RemoteItem.Title, 50))
		if c.Resolution != "" {
			fmt.Printf("   Resolution: %s\n", c.Resolution)
		}
		fmt.Println()
	}
}

// UpdateFileCategories updates the Categories section in a markdown file
func UpdateFileCategories(filePath string, categories []string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	categoriesLine := strings.Join(categories, ", ")

	// Check if file has Categories section
	if strings.Contains(contentStr, "## Categories") {
		// Replace existing Categories section content
		re := regexp.MustCompile(`(## Categories\n)[^\n#]*`)
		if len(categories) == 0 {
			// Remove entire Categories section if empty
			contentStr = regexp.MustCompile(`## Categories\n[^\n#]*\n?`).ReplaceAllString(contentStr, "")
		} else {
			contentStr = re.ReplaceAllString(contentStr, "${1}"+categoriesLine+"\n")
		}
	} else if len(categories) > 0 {
		// Add Categories section after Summary (or after title if no Summary)
		if strings.Contains(contentStr, "## Summary") {
			// Find the end of Summary section and insert after
			re := regexp.MustCompile(`(## Summary\n[^\n#]*\n)`)
			contentStr = re.ReplaceAllString(contentStr, "${1}\n## Categories\n"+categoriesLine+"\n")
		} else if strings.Contains(contentStr, "## Priority") {
			// Insert before Priority
			contentStr = strings.Replace(contentStr, "## Priority", "## Categories\n"+categoriesLine+"\n\n## Priority", 1)
		} else {
			// Find first ## section and insert before it
			re := regexp.MustCompile(`(\n## )`)
			if re.MatchString(contentStr) {
				contentStr = re.ReplaceAllStringFunc(contentStr, func(s string) string {
					return "\n## Categories\n" + categoriesLine + "\n" + s
				})
			} else {
				// No sections found, append at end
				contentStr += "\n## Categories\n" + categoriesLine + "\n"
			}
		}
	}

	return os.WriteFile(filePath, []byte(contentStr), 0644)
}

// AddCategoryToFile adds a category to a file's Categories section
func AddCategoryToFile(filePath string, categoryID string) error {
	item, err := ParseMarkdownFile(filePath)
	if err != nil {
		return err
	}

	// Check if category already exists
	for _, cat := range item.Categories {
		if cat == categoryID {
			return nil // Already has this category
		}
	}

	// Add category
	item.Categories = append(item.Categories, categoryID)
	return UpdateFileCategories(filePath, item.Categories)
}

// RemoveCategoryFromFile removes a category from a file's Categories section
func RemoveCategoryFromFile(filePath string, categoryID string) error {
	item, err := ParseMarkdownFile(filePath)
	if err != nil {
		return err
	}

	// Remove category
	newCategories := make([]string, 0, len(item.Categories))
	found := false
	for _, cat := range item.Categories {
		if cat == categoryID {
			found = true
		} else {
			newCategories = append(newCategories, cat)
		}
	}

	if !found {
		return nil // Category not found, nothing to remove
	}

	return UpdateFileCategories(filePath, newCategories)
}

// ClearCategoriesFromFile removes all categories from a file
func ClearCategoriesFromFile(filePath string) error {
	return UpdateFileCategories(filePath, nil)
}
