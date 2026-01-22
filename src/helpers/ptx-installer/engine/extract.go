package engine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractArchive extracts an archive based on its type
func ExtractArchive(archivePath, targetDir string) error {
	// Determine archive type from extension
	ext := strings.ToLower(filepath.Ext(archivePath))

	switch ext {
	case ".zip":
		return ExtractZip(archivePath, targetDir)
	case ".gz":
		if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
			return ExtractTarGz(archivePath, targetDir)
		}
		return ExtractGzip(archivePath, targetDir)
	case ".tar":
		return ExtractTar(archivePath, targetDir)
	case ".xz":
		if strings.HasSuffix(strings.ToLower(archivePath), ".tar.xz") {
			return ExtractTarXz(archivePath, targetDir)
		}
		return fmt.Errorf("unsupported archive type: %s", ext)
	default:
		return fmt.Errorf("unsupported archive type: %s", ext)
	}
}

// ExtractZip extracts a ZIP archive
func ExtractZip(zipFile, destDir string) error {
	fmt.Printf("📦 Extracting ZIP: %s\n", filepath.Base(zipFile))

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract files
	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			os.MkdirAll(fpath, f.Mode())
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		// Extract file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}

		fmt.Printf("   ✓ %s\n", f.Name)
	}

	fmt.Printf("✅ Extraction complete\n")
	return nil
}

// ExtractTarGz extracts a tar.gz archive
func ExtractTarGz(tarFile, destDir string) error {
	fmt.Printf("📦 Extracting tar.gz: %s\n", filepath.Base(tarFile))

	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	return extractTarReader(gzr, destDir)
}

// ExtractTar extracts a plain tar archive
func ExtractTar(tarFile, destDir string) error {
	fmt.Printf("📦 Extracting tar: %s\n", filepath.Base(tarFile))

	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("failed to open tar: %w", err)
	}
	defer file.Close()

	return extractTarReader(file, destDir)
}

// extractTarReader extracts from a tar reader
func extractTarReader(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	extractedFiles := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Extract file
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			extractedFiles++
			if extractedFiles%10 == 0 {
				fmt.Printf("   ✓ Extracted %d files...\n", extractedFiles)
			}

		case tar.TypeSymlink:
			// Create symlink
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore symlink errors on Windows
				if !strings.Contains(err.Error(), "not supported") {
					return err
				}
			}
		}
	}

	fmt.Printf("✅ Extracted %d files\n", extractedFiles)
	return nil
}

// ExtractGzip extracts a single gzipped file
func ExtractGzip(gzipFile, destDir string) error {
	fmt.Printf("📦 Extracting gzip: %s\n", filepath.Base(gzipFile))

	file, err := os.Open(gzipFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Determine output filename (remove .gz extension)
	outName := strings.TrimSuffix(filepath.Base(gzipFile), ".gz")
	outPath := filepath.Join(destDir, outName)

	// Create output file
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gzr)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Extracted: %s\n", outName)
	return nil
}

// ExtractTarXz extracts a tar.xz archive (requires external xz command)
func ExtractTarXz(tarFile, destDir string) error {
	return fmt.Errorf("tar.xz extraction not yet implemented - requires xz-utils")
}

// FindExtractedRoot finds the actual root directory after extraction
// Many archives contain a single top-level directory (e.g., graalvm-community-openjdk-21.0.2+13.1/)
// This function returns the path to that directory, or the original path if multiple items exist
func FindExtractedRoot(extractDir string) (string, error) {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return extractDir, err
	}

	// If there's exactly one entry and it's a directory, return its path
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(extractDir, entries[0].Name())
		// Verify it looks like a valid root (has bin/ or lib/ directory)
		subEntries, err := os.ReadDir(subDir)
		if err == nil {
			for _, e := range subEntries {
				if e.IsDir() && (e.Name() == "bin" || e.Name() == "lib") {
					return subDir, nil
				}
			}
		}
	}

	return extractDir, nil
}

// FindBinaryInExtracted finds a binary file in the extracted directory
func FindBinaryInExtracted(extractDir, binaryName string) (string, error) {
	var foundPath string

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if filename matches
		if info.Name() == binaryName {
			// Check if file is executable
			if info.Mode()&0111 != 0 {
				foundPath = path
				return filepath.SkipAll
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("binary %s not found in extracted directory", binaryName)
	}

	return foundPath, nil
}
