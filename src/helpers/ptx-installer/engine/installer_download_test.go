/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

func TestInstallDownload_SingleFile(t *testing.T) {
	// Create test HTTP server serving a fake model file
	fileContent := "fake-onnx-model-data"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fileContent))
	}))
	defer server.Close()

	// Create temp directories
	targetDir := t.TempDir()
	cacheDir := t.TempDir()

	// Create installer with empty registry
	reg := &registry.PackageRegistry{}
	installer := &Installer{
		registry: reg,
		cacheDir: cacheDir,
	}

	variant := &registry.VariantSpec{
		Version:   "1.0.0",
		URL:       server.URL + "/model.onnx",
		ExtractTo: targetDir,
	}

	platform := &registry.PlatformSpec{Type: "download"}
	options := &InstallOptions{PackageName: "test-model"}

	err := installer.installDownload(platform, variant, options)
	if err != nil {
		t.Fatalf("installDownload failed: %v", err)
	}

	// Verify file was downloaded
	content, err := os.ReadFile(filepath.Join(targetDir, "model.onnx"))
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != fileContent {
		t.Errorf("file content mismatch: got %q, want %q", string(content), fileContent)
	}
}

func TestInstallDownload_WithAdditionalFiles(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content-of-" + r.URL.Path))
	}))
	defer server.Close()

	targetDir := t.TempDir()
	cacheDir := t.TempDir()

	reg := &registry.PackageRegistry{}
	installer := &Installer{
		registry: reg,
		cacheDir: cacheDir,
	}

	variant := &registry.VariantSpec{
		Version:   "1.0.0",
		URL:       server.URL + "/model.onnx",
		ExtractTo: targetDir,
		AdditionalFiles: []registry.AdditionalFile{
			{URL: server.URL + "/model.onnx.json"},
		},
	}

	platform := &registry.PlatformSpec{Type: "download"}
	options := &InstallOptions{PackageName: "test-model"}

	err := installer.installDownload(platform, variant, options)
	if err != nil {
		t.Fatalf("installDownload failed: %v", err)
	}

	// Verify both files were downloaded
	for _, filename := range []string{"model.onnx", "model.onnx.json"} {
		path := filepath.Join(targetDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", filename)
		}
	}
}

func TestInstallDownload_AdditionalFilesCustomFilename(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer server.Close()

	targetDir := t.TempDir()
	cacheDir := t.TempDir()

	reg := &registry.PackageRegistry{}
	installer := &Installer{
		registry: reg,
		cacheDir: cacheDir,
	}

	variant := &registry.VariantSpec{
		Version:   "1.0.0",
		URL:       server.URL + "/file",
		ExtractTo: targetDir,
		AdditionalFiles: []registry.AdditionalFile{
			{URL: server.URL + "/config", Filename: "custom-name.json"},
		},
	}

	platform := &registry.PlatformSpec{Type: "download"}
	options := &InstallOptions{PackageName: "test"}

	err := installer.installDownload(platform, variant, options)
	if err != nil {
		t.Fatalf("installDownload failed: %v", err)
	}

	// Verify custom filename was used
	path := filepath.Join(targetDir, "custom-name.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file custom-name.json to exist")
	}
}

func TestInstallDownload_NoURLError(t *testing.T) {
	cacheDir := t.TempDir()

	reg := &registry.PackageRegistry{}
	installer := &Installer{
		registry: reg,
		cacheDir: cacheDir,
	}

	variant := &registry.VariantSpec{
		Version: "1.0.0",
		// No URL and no AdditionalFiles
	}

	platform := &registry.PlatformSpec{Type: "download"}
	options := &InstallOptions{PackageName: "test"}

	err := installer.installDownload(platform, variant, options)
	if err == nil {
		t.Error("expected error for variant with no URL and no additional files")
	}
}
