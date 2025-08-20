package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"portunix.cz/app"
	"portunix.cz/app/install"
)

// Test for InstallJavaWinRunDry
func TestInstallJavaRunDry(t *testing.T) {

	err := install.WinInstallJavaRun("", "", true)

	if err != nil {
		t.Errorf("InstallJava error %s", err)
	}

	err = install.WinInstallJavaRun("11", "", true)

	if err != nil {
		t.Errorf("InstallJava error %s", err)
	}
}

func TestProcessArgumentsInstallJava(t *testing.T) {

	arguments := []string{"11", "openjdk"}
	result := install.ProcessArgumentsInstallJava(arguments)
	expected := "11"
	paramName := "version"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}
	expected = "openjdk"
	paramName = "variant"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}

	arguments = []string{"-version", "11", "-variant", "openjdk"}
	result = install.ProcessArgumentsInstallJava(arguments)
	expected = "11"
	paramName = "version"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}
	expected = "openjdk"
	paramName = "variant"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}
}

func TestProcessArgumentsUnzip(t *testing.T) {
	arguments := []string{"tp.cli.zip"}
	result, _ := app.ProcessArgumentsUnzip(arguments)
	expected := "tp.cli.zip"
	paramName := "path"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}

	arguments = []string{"tp.cli.zip", "."}
	result, _ = app.ProcessArgumentsUnzip(arguments)
	expected = "tp.cli.zip"
	paramName = "path"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}

	expected = "."
	paramName = "destinationpath"
	if result[paramName] != expected {
		t.Errorf("Expected %s but got %s", expected, result[paramName])
	}
}

func TestUnzip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "unzip_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test file to zip
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}

	// Create a simple ZIP file for testing
	zipFile := filepath.Join(tempDir, "test.zip")
	err = createTestZip(zipFile, testFile)
	if err != nil {
		t.Fatalf("Failed to create test zip: %s", err)
	}

	// Create output directory
	outputDir := filepath.Join(tempDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %s", err)
	}

	// Test unzip
	arguments := []string{zipFile, outputDir}
	err = app.Unzip(arguments)
	if err != nil {
		t.Errorf("Unzip error: %s", err)
	}

	// Verify extracted file exists
	extractedFile := filepath.Join(outputDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Errorf("Extracted file does not exist: %s", extractedFile)
	}
}

// createTestZip creates a simple ZIP file for testing
func createTestZip(zipPath, filePath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Read the file to be zipped
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Add file to ZIP
	fileName := filepath.Base(filePath)
	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	_, err = fileWriter.Write(fileData)
	return err
}
