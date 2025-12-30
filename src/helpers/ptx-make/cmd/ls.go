package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// lsOptions holds the parsed flags for ls command
type lsOptions struct {
	longFormat    bool // -l
	showHidden    bool // -a
	humanReadable bool // -h
	recursive     bool // -R
	sortByTime    bool // -t
	sortBySize    bool // -S
	reverseSort   bool // -r
}

// RunLs executes the ls command
func RunLs(args []string) error {
	opts, paths := parseLsArgs(args)

	// Default to current directory if no path specified
	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Try native ls first (if available and not on Windows without ls)
	if canUseNativeLS() {
		return executeNativeLS(args)
	}

	// Fall back to Go emulation
	return emulateLs(opts, paths)
}

// parseLsArgs parses arguments into options and paths
func parseLsArgs(args []string) (lsOptions, []string) {
	opts := lsOptions{}
	var paths []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			// Parse combined flags like -lah
			for _, ch := range arg[1:] {
				switch ch {
				case 'l':
					opts.longFormat = true
				case 'a':
					opts.showHidden = true
				case 'h':
					opts.humanReadable = true
				case 'R':
					opts.recursive = true
				case 't':
					opts.sortByTime = true
				case 'S':
					opts.sortBySize = true
				case 'r':
					opts.reverseSort = true
				}
			}
		} else if arg == "--help" {
			fmt.Println("Usage: portunix make ls [options] [path...]")
			fmt.Println()
			fmt.Println("List directory contents")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  -l    Long format with details")
			fmt.Println("  -a    Include hidden files (starting with .)")
			fmt.Println("  -h    Human-readable sizes (with -l)")
			fmt.Println("  -R    Recursive listing")
			fmt.Println("  -t    Sort by modification time (newest first)")
			fmt.Println("  -S    Sort by size (largest first)")
			fmt.Println("  -r    Reverse sort order")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  portunix make ls")
			fmt.Println("  portunix make ls -l /path/to/dir")
			fmt.Println("  portunix make ls -lah")
			fmt.Println("  portunix make ls *.go")
			os.Exit(0)
		} else {
			paths = append(paths, arg)
		}
	}

	return opts, paths
}

// canUseNativeLS checks if native ls command is available
func canUseNativeLS() bool {
	_, err := exec.LookPath("ls")
	return err == nil
}

// executeNativeLS runs the native ls command with given arguments
func executeNativeLS(args []string) error {
	cmd := exec.Command("ls", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// emulateLs provides Go-native ls emulation
func emulateLs(opts lsOptions, paths []string) error {
	for i, path := range paths {
		// Handle glob patterns
		if strings.Contains(path, "*") {
			matches, err := filepath.Glob(path)
			if err != nil {
				return fmt.Errorf("invalid pattern: %s", path)
			}
			if len(matches) == 0 {
				// No matches - not an error, just skip
				continue
			}
			// List matched files
			for _, match := range matches {
				if err := listPath(opts, match, len(paths) > 1 || opts.recursive); err != nil {
					return err
				}
			}
			continue
		}

		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("cannot access '%s': No such file or directory", path)
			}
			return err
		}

		// Show directory header when listing multiple paths
		showHeader := len(paths) > 1 || opts.recursive

		if info.IsDir() {
			if showHeader {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("%s:\n", path)
			}
			if err := listDirectory(opts, path); err != nil {
				return err
			}
		} else {
			// Single file
			if err := printEntry(opts, path, info); err != nil {
				return err
			}
		}
	}

	return nil
}

// listPath lists a single path (file or directory)
func listPath(opts lsOptions, path string, showHeader bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if showHeader {
			fmt.Printf("%s:\n", path)
		}
		return listDirectory(opts, path)
	}

	return printEntry(opts, path, info)
}

// listDirectory lists contents of a directory
func listDirectory(opts lsOptions, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Collect file info for sorting
	var files []fileEntry
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless -a is specified
		if !opts.showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, fileEntry{
			name:    name,
			info:    info,
			modTime: info.ModTime(),
			size:    info.Size(),
			isDir:   info.IsDir(),
		})
	}

	// Sort entries
	sortEntries(files, opts)

	// Print entries
	for _, f := range files {
		fullPath := filepath.Join(dirPath, f.name)
		if err := printEntry(opts, fullPath, f.info); err != nil {
			return err
		}
	}

	// Handle recursive listing
	if opts.recursive {
		for _, f := range files {
			if f.isDir {
				fmt.Println()
				subPath := filepath.Join(dirPath, f.name)
				fmt.Printf("%s:\n", subPath)
				if err := listDirectory(opts, subPath); err != nil {
					// Continue on error for recursive listing
					fmt.Fprintf(os.Stderr, "ls: cannot access '%s': %v\n", subPath, err)
				}
			}
		}
	}

	return nil
}

// fileEntry type for sorting
type fileEntry struct {
	name    string
	info    os.FileInfo
	modTime time.Time
	size    int64
	isDir   bool
}

// sortEntries sorts file entries based on options
func sortEntries(files []fileEntry, opts lsOptions) {
	sort.Slice(files, func(i, j int) bool {
		var less bool

		if opts.sortByTime {
			less = files[i].modTime.After(files[j].modTime)
		} else if opts.sortBySize {
			less = files[i].size > files[j].size
		} else {
			// Default: alphabetical
			less = strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
		}

		if opts.reverseSort {
			return !less
		}
		return less
	})
}

// printEntry prints a single file/directory entry
func printEntry(opts lsOptions, path string, info os.FileInfo) error {
	name := filepath.Base(path)

	if opts.longFormat {
		// Long format: permissions size date name
		perm := formatPermissions(info)
		size := formatSize(info.Size(), opts.humanReadable, info.IsDir())
		date := info.ModTime().Format("2006-01-02 15:04")

		if info.IsDir() {
			fmt.Printf("%s  %s  %s  %s/\n", perm, size, date, name)
		} else {
			fmt.Printf("%s  %s  %s  %s\n", perm, size, date, name)
		}
	} else {
		// Simple format: just name
		if info.IsDir() {
			fmt.Printf("%s/\n", name)
		} else {
			fmt.Println(name)
		}
	}

	return nil
}

// formatPermissions formats file permissions in ls -l style
func formatPermissions(info os.FileInfo) string {
	mode := info.Mode()
	var perm strings.Builder

	// File type
	if info.IsDir() {
		perm.WriteByte('d')
	} else if mode&os.ModeSymlink != 0 {
		perm.WriteByte('l')
	} else {
		perm.WriteByte('-')
	}

	// On Windows, provide simplified permissions
	if runtime.GOOS == "windows" {
		if info.IsDir() {
			perm.WriteString("rwxr-xr-x")
		} else if mode&0200 != 0 {
			perm.WriteString("rw-r--r--")
		} else {
			perm.WriteString("r--r--r--")
		}
	} else {
		// Unix permissions
		perm.WriteString(formatUnixPermissions(mode))
	}

	return perm.String()
}

// formatUnixPermissions formats Unix-style permission bits
func formatUnixPermissions(mode os.FileMode) string {
	var perm strings.Builder

	// Owner
	if mode&0400 != 0 {
		perm.WriteByte('r')
	} else {
		perm.WriteByte('-')
	}
	if mode&0200 != 0 {
		perm.WriteByte('w')
	} else {
		perm.WriteByte('-')
	}
	if mode&0100 != 0 {
		perm.WriteByte('x')
	} else {
		perm.WriteByte('-')
	}

	// Group
	if mode&0040 != 0 {
		perm.WriteByte('r')
	} else {
		perm.WriteByte('-')
	}
	if mode&0020 != 0 {
		perm.WriteByte('w')
	} else {
		perm.WriteByte('-')
	}
	if mode&0010 != 0 {
		perm.WriteByte('x')
	} else {
		perm.WriteByte('-')
	}

	// Others
	if mode&0004 != 0 {
		perm.WriteByte('r')
	} else {
		perm.WriteByte('-')
	}
	if mode&0002 != 0 {
		perm.WriteByte('w')
	} else {
		perm.WriteByte('-')
	}
	if mode&0001 != 0 {
		perm.WriteByte('x')
	} else {
		perm.WriteByte('-')
	}

	return perm.String()
}

// formatSize formats file size for display
func formatSize(size int64, humanReadable bool, isDir bool) string {
	if isDir {
		return fmt.Sprintf("%5s", "-")
	}

	if !humanReadable {
		return fmt.Sprintf("%5d", size)
	}

	// Human readable format
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%4.1fT", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%4.1fG", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%4.1fM", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%4.1fK", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%5d", size)
	}
}
