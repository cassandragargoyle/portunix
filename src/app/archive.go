package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnzipFile(src string, dest string) error {
	return unzip(src, dest)
}

// Unzip extracts a ZIP file to the specified destination directory.
func unzip(src string, dest string) error {

	// Determine the destination directory
	if dest == "" || dest == "." || dest == ".\\" {
		absPath, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		dest = filepath.Dir(absPath)
	}

	// Open the ZIP file first so a missing archive leaves no directory behind
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create the destination directory
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err := os.MkdirAll(dest, 0755); err != nil { // #nosec G301 -- extracted content must be readable by all users
			return err
		}
	}

	// Iterate through the files in the archive
	for _, f := range r.File {
		// Protect against ZipSlip vulnerability
		if !filepath.IsLocal(f.Name) {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}

		filePath := filepath.Join(dest, f.Name) // #nosec G305 -- entry name validated with filepath.IsLocal above
		fmt.Println("Extracting:", filePath)

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(filePath, 0755); err != nil { // #nosec G301 -- extracted content must be readable by all users
				return err
			}
		} else {
			// Create file
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil { // #nosec G301 -- extracted content must be readable by all users
				return err
			}

			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			rc, err := f.Open()
			if err != nil {
				return err
			}

			err = copyLimited(outFile, rc, maxExtractedFileSize)
			_ = rc.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// maxExtractedFileSize caps a single extracted file to stop decompression bombs
const maxExtractedFileSize int64 = 8 << 30

// copyLimited copies src to dst and fails once more than limit bytes were written
func copyLimited(dst io.Writer, src io.Reader, limit int64) error {
	written, err := io.CopyN(dst, src, limit+1)
	if err != nil && err != io.EOF {
		return err
	}
	if written > limit {
		return fmt.Errorf("extracted file exceeds the %d byte limit", limit)
	}
	return nil
}
