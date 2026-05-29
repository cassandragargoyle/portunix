/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package registry

import (
	"testing"
)

func TestValidatePlatform_DownloadType(t *testing.T) {
	r := &PackageRegistry{}

	tests := []struct {
		name         string
		platformType string
		wantErr      bool
	}{
		{"download type is valid", "download", false},
		{"apt type is valid", "apt", false},
		{"container type is valid", "container", false},
		{"tar.gz type is valid", "tar.gz", false},
		{"zip type is valid", "zip", false},
		{"script type is valid", "script", false},
		{"invalid type rejected", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform := &PlatformSpec{
				Type: tt.platformType,
				Variants: map[string]VariantSpec{
					"default": {
						Version: "1.0.0",
						URL:     "https://example.com/file.bin",
					},
				},
			}
			err := r.validatePlatform("linux", platform)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlatform() type=%s, error = %v, wantErr %v", tt.platformType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateVariant_AdditionalFiles(t *testing.T) {
	r := &PackageRegistry{}

	// Variant with only additionalFiles (no url) should be valid
	variant := &VariantSpec{
		Version: "1.0.0",
		AdditionalFiles: []AdditionalFile{
			{URL: "https://example.com/model.onnx"},
			{URL: "https://example.com/model.onnx.json"},
		},
	}
	if err := r.validateVariant("default", variant); err != nil {
		t.Errorf("variant with additionalFiles should be valid, got: %v", err)
	}

	// Variant with url + additionalFiles should be valid
	variant2 := &VariantSpec{
		Version: "1.0.0",
		URL:     "https://example.com/model.onnx",
		AdditionalFiles: []AdditionalFile{
			{URL: "https://example.com/model.onnx.json"},
		},
	}
	if err := r.validateVariant("default", variant2); err != nil {
		t.Errorf("variant with url + additionalFiles should be valid, got: %v", err)
	}

	// Variant with nothing should be invalid
	variant3 := &VariantSpec{
		Version: "1.0.0",
	}
	if err := r.validateVariant("default", variant3); err == nil {
		t.Error("variant with no install method should be invalid")
	}
}

func TestValidatePackage_Bundle(t *testing.T) {
	r := &PackageRegistry{}

	validMeta := Metadata{
		Name:        "ai-assistant-basic",
		DisplayName: "AI Assistant (Basic)",
		Description: "Basic AI assistant setup",
		Category:    "development/ai-tools",
	}

	// Bundle with members and no platforms is valid
	bundle := &Package{
		APIVersion: "v1",
		Kind:       "Bundle",
		Metadata:   validMeta,
		Spec:       PackageSpec{Bundle: []string{"claude-code", "gemini-cli"}},
	}
	if err := r.validatePackage(bundle); err != nil {
		t.Errorf("valid bundle should pass validation, got: %v", err)
	}

	// Bundle with empty member list is invalid
	emptyBundle := &Package{
		APIVersion: "v1",
		Kind:       "Bundle",
		Metadata:   validMeta,
		Spec:       PackageSpec{},
	}
	if err := r.validatePackage(emptyBundle); err == nil {
		t.Error("bundle without members should be invalid")
	}

	// Regular package still requires platforms
	pkgNoPlatforms := &Package{
		APIVersion: "v1",
		Kind:       "Package",
		Metadata:   validMeta,
		Spec:       PackageSpec{},
	}
	if err := r.validatePackage(pkgNoPlatforms); err == nil {
		t.Error("Kind: Package without platforms should be invalid")
	}

	// Unknown kind is rejected
	badKind := &Package{
		APIVersion: "v1",
		Kind:       "Widget",
		Metadata:   validMeta,
		Spec:       PackageSpec{Bundle: []string{"claude-code"}},
	}
	if err := r.validatePackage(badKind); err == nil {
		t.Error("unknown kind should be rejected")
	}
}

func TestAdditionalFileStruct(t *testing.T) {
	af := AdditionalFile{
		URL:      "https://example.com/model.onnx.json",
		Filename: "model.onnx.json",
	}

	if af.URL != "https://example.com/model.onnx.json" {
		t.Errorf("unexpected URL: %s", af.URL)
	}
	if af.Filename != "model.onnx.json" {
		t.Errorf("unexpected Filename: %s", af.Filename)
	}

	// Filename can be empty (derived from URL)
	af2 := AdditionalFile{
		URL: "https://example.com/file.bin",
	}
	if af2.Filename != "" {
		t.Errorf("expected empty filename, got: %s", af2.Filename)
	}
}
