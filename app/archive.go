package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Unzip(arguments []string) error {
	argsMap, _ := ProcessArgumentsUnzip(arguments)
	if argsMap["destinationpath"] == "" {
		argsMap["destinationpath"] = "."
	}
	err := unzip(argsMap["path"], argsMap["destinationpath"])
	return err
}

func UnzipFile(src string, dest string) error {
	return unzip(src, dest)
}

func ProcessArgumentsUnzip(arguments []string) (map[string]string, []string) {
	//TODO: use list
	enabledArguments := []string{"path", "destinationpath"}
	argsMap, other := ProcessArguments(arguments, enabledArguments)
	if len(arguments) > 0 && argsMap["path"] == "" && argsMap["destinationpath"] == "" && !isParameter(arguments[0]) {
		argsMap["path"] = arguments[0]
		if len(arguments) > 1 && argsMap["destinationpath"] == "" && !isParameter(arguments[1]) {
			argsMap["destinationpath"] = arguments[1]
		}
	}
	return argsMap, other
}

func isParameter(str string) bool {
	if strings.HasPrefix(str, "-") || strings.HasPrefix(str, "--") || strings.Contains(str, "=") {
		return true
	}
	return false
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

	// Create the destination directory
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err := os.MkdirAll(dest, os.ModePerm); err != nil {
			return err
		}
	}

	// Open the ZIP file
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Iterate through the files in the archive
	for _, f := range r.File {
		filePath := filepath.Join(dest, f.Name)
		fmt.Println("Extracting:", filePath)

		// Protect against ZipSlip vulnerability
		/* TODO: Something is not working here
		if !filepath.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}*/

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
		} else {
			// Create file
			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
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

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}

			rc.Close()
		}
	}

	return nil
}
