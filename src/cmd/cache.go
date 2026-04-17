package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"portunix.ai/app/cache"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage Portunix cache",
	Long: `Manage Portunix download and metadata cache.

Cache stores downloaded packages, HTTP responses, build artifacts,
and metadata to improve performance and reduce bandwidth usage.

Cache categories:
  downloads  - Downloaded packages, installers, and archives
  http       - HTTP API responses and registry data
  builds     - Build artifacts and temporary files
  metadata   - Package definitions and dependency trees

Environment variables:
  PORTUNIX_CACHE_DIR       Override cache directory
  PORTUNIX_CACHE_DISABLED  Set to "true" to disable caching`,
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show cache information and statistics",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := cache.NewManager()
		info, err := mgr.GetInfo()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		status := "enabled"
		if !info.Enabled {
			status = "disabled"
		}

		fmt.Println("Cache Information")
		fmt.Println("=================")
		fmt.Printf("Status:     %s\n", status)
		fmt.Printf("Location:   %s\n", info.BaseDir)
		fmt.Printf("Total size: %s / %s\n", cache.FormatSize(info.TotalSize), cache.FormatSize(info.MaxSize))
		fmt.Printf("Files:      %d\n", info.TotalFiles)
		fmt.Println()

		fmt.Println("Categories:")
		for _, cat := range cache.AllCategories() {
			catInfo, ok := info.Categories[cat]
			if !ok {
				continue
			}
			fmt.Printf("  %-12s %s / %s (%d files)\n",
				string(cat),
				cache.FormatSize(catInfo.Size),
				cache.FormatSize(catInfo.MaxSize),
				catInfo.Files,
			)
		}
	},
}

var cacheListCmd = &cobra.Command{
	Use:   "list [category]",
	Short: "List cached items",
	Long: `List all cached items, optionally filtered by category.

Available categories: downloads, http, builds, metadata`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := cache.NewManager()

		if len(args) == 1 {
			cat := cache.Category(args[0])
			listCategory(mgr, cat)
			return
		}

		allEntries, err := mgr.ListAllEntries()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(allEntries) == 0 {
			fmt.Println("Cache is empty")
			return
		}

		for _, cat := range cache.AllCategories() {
			entries, ok := allEntries[cat]
			if !ok || len(entries) == 0 {
				continue
			}
			fmt.Printf("[%s]\n", cat)
			for _, entry := range entries {
				printEntry(entry)
			}
			fmt.Println()
		}
	},
}

var cacheCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove expired and invalid cache entries",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := cache.NewManager()
		result, err := mgr.Clean()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if result.RemovedFiles == 0 {
			fmt.Println("Cache is clean, no expired entries found")
			return
		}

		fmt.Printf("Cleaned %d entries, freed %s\n",
			result.RemovedFiles,
			cache.FormatSize(result.FreedBytes),
		)
	},
}

var cachePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Clear all cache contents",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := cache.NewManager()
		result, err := mgr.Purge()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if result.RemovedFiles == 0 {
			fmt.Println("Cache was already empty")
			return
		}

		fmt.Printf("Purged %d entries, freed %s\n",
			result.RemovedFiles,
			cache.FormatSize(result.FreedBytes),
		)
	},
}

var cacheRemoveCmd = &cobra.Command{
	Use:   "remove <pattern>",
	Short: "Remove cache entries matching a pattern",
	Long: `Remove cache entries where the source URL or filename contains the given pattern.

Examples:
  portunix cache remove nodejs
  portunix cache remove python-3.13`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		mgr := cache.NewManager()
		result, err := mgr.RemoveBySource(pattern)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if result.RemovedFiles == 0 {
			fmt.Printf("No cache entries matching '%s'\n", pattern)
			return
		}

		fmt.Printf("Removed %d entries matching '%s', freed %s\n",
			result.RemovedFiles,
			pattern,
			cache.FormatSize(result.FreedBytes),
		)
	},
}

func listCategory(mgr *cache.Manager, cat cache.Category) {
	entries, err := mgr.ListEntries(cat)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Printf("No cached entries in '%s'\n", cat)
		return
	}

	fmt.Printf("[%s] %d entries\n", cat, len(entries))
	for _, entry := range entries {
		printEntry(entry)
	}
}

func printEntry(entry *cache.EntryMeta) {
	ttl := time.Until(entry.ExpiresAt)
	ttlStr := cache.FormatDuration(ttl)

	source := entry.Source
	if len(source) > 60 {
		source = source[:57] + "..."
	}

	expired := ""
	if entry.IsExpired() {
		expired = " [EXPIRED]"
	}

	fmt.Printf("  %-40s %8s  TTL: %s%s\n",
		truncate(entry.Filename, 40),
		cache.FormatSize(entry.Size),
		ttlStr,
		expired,
	)
	if source != "" && source != entry.Filename {
		fmt.Printf("    Source: %s\n", source)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s + strings.Repeat(" ", maxLen-len(s))
	}
	return s[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
	cacheCmd.AddCommand(cachePurgeCmd)
	cacheCmd.AddCommand(cacheRemoveCmd)
}
