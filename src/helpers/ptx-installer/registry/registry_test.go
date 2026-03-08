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
		name        string
		platformType string
		wantErr     bool
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
