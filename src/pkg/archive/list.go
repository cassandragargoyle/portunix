/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TopLevelEntry is a file or directory at the root of an archive
type TopLevelEntry struct {
	Name  string
	IsDir bool
}

// ListTopLevelEntries returns the entries an archive creates directly in the
// extraction directory, sorted by name (e.g. "go" for the Go tarball, or
// "hugo", "LICENSE", "README.md" for the Hugo tarball)
func ListTopLevelEntries(archivePath string) ([]TopLevelEntry, error) {
	lower := strings.ToLower(archivePath)
	var names []string
	var err error

	switch {
	case strings.HasSuffix(lower, ".zip"):
		names, err = listZipNames(archivePath)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		names, err = listTarNames(archivePath, true)
	case strings.HasSuffix(lower, ".tar"):
		names, err = listTarNames(archivePath, false)
	default:
		return nil, fmt.Errorf("unsupported archive type: %s", filepath.Ext(archivePath))
	}
	if err != nil {
		return nil, err
	}

	return topLevelEntries(names), nil
}

// topLevelEntries reduces archive member paths to their first path component
func topLevelEntries(names []string) []TopLevelEntry {
	entries := make(map[string]bool)
	for _, name := range names {
		name = strings.TrimPrefix(filepath.ToSlash(name), "./")
		name = strings.TrimPrefix(name, "/")
		if name == "" || name == "." {
			continue
		}
		first, _, hasRest := strings.Cut(name, "/")
		if first == "" || first == "." || first == ".." {
			continue
		}
		// A member below the first component, or an explicit "dir/" entry, makes it a directory
		entries[first] = entries[first] || hasRest
	}

	result := make([]TopLevelEntry, 0, len(entries))
	for name, isDir := range entries {
		result = append(result, TopLevelEntry{Name: name, IsDir: isDir})
	}
	sort.Slice(result, func(a, b int) bool { return result[a].Name < result[b].Name })
	return result
}

func listZipNames(zipFile string) ([]string, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names, nil
}

func listTarNames(tarFile string, gzipped bool) ([]string, error) {
	file, err := os.Open(tarFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if gzipped {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	tr := tar.NewReader(reader)
	var names []string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}
		name := header.Name
		if header.Typeflag == tar.TypeDir && !strings.HasSuffix(name, "/") {
			name += "/"
		}
		names = append(names, name)
	}
	return names, nil
}
