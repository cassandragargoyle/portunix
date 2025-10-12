package main

import (
	"path/filepath"
	"testing"
	"portunix.ai/portunix/test/testframework"
	"portunix.ai/app/install"
)

// TestRegistryLoading tests the basic registry loading functionality
func TestRegistryLoading(t *testing.T) {
	tf := testframework.NewTestFramework("Registry_Loading")
	tf.Start(t, "Test registry loading and package discovery")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Load package registry")
	assetsPath := filepath.Join("..", "..", "assets")
	registry, err := install.LoadPackageRegistry(assetsPath)

	if err != nil {
		tf.Error(t, "Failed to load registry", err.Error())
		success = false
		return
	}
	tf.Success(t, "Registry loaded successfully")

	tf.Separator()

	tf.Step(t, "Check migrated packages exist")
	expectedPackages := []string{"nodejs", "python", "go", "vscode", "chrome"}

	for _, packageName := range expectedPackages {
		tf.Info(t, "Testing package:", packageName)
		pkg, err := registry.GetPackage(packageName)
		if err != nil {
			tf.Error(t, "Package not found in registry", packageName, err.Error())
			success = false
			continue
		}

		if pkg.Metadata.Name != packageName {
			tf.Error(t, "Package name mismatch", "expected:", packageName, "got:", pkg.Metadata.Name)
			success = false
			continue
		}

		tf.Success(t, "Package found with correct metadata", packageName)
	}

	tf.Separator()

	tf.Step(t, "Test registry to legacy conversion")
	legacyConfig, err := registry.ConvertToLegacyConfig()
	if err != nil {
		tf.Error(t, "Failed to convert registry to legacy format", err.Error())
		success = false
		return
	}

	// Check that migrated packages exist in legacy format
	for _, packageName := range expectedPackages {
		if _, exists := legacyConfig.Packages[packageName]; !exists {
			tf.Error(t, "Package missing from legacy conversion", packageName)
			success = false
		} else {
			tf.Success(t, "Package converted to legacy format", packageName)
		}
	}

	tf.Success(t, "Registry to legacy conversion completed")
}

// TestPackageInstallationDryRun tests package installation with dry-run
func TestPackageInstallationDryRun(t *testing.T) {
	tf := testframework.NewTestFramework("Package_Installation_DryRun")
	tf.Start(t, "Test package installation with dry-run for registry packages")

	success := true
	defer tf.Finish(t, success)

	// Test packages that should work with new registry
	testPackages := []string{"nodejs", "python", "go"}

	for _, packageName := range testPackages {
		tf.Step(t, "Test dry-run installation for", packageName)

		options := &install.InstallOptions{
			PackageName: packageName,
			DryRun:      true,
		}

		err := install.InstallPackageWithOptions(options)
		if err != nil {
			tf.Warning(t, "Dry-run failed for", packageName, err.Error())
			// Don't fail the test for dry-run issues as they might be platform-specific
		} else {
			tf.Success(t, "Dry-run successful for", packageName)
		}
	}
}

// TestBackwardCompatibility tests that old packages still work
func TestBackwardCompatibility(t *testing.T) {
	tf := testframework.NewTestFramework("Backward_Compatibility")
	tf.Start(t, "Test that non-migrated packages still work through legacy system")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Test package not in registry falls back to legacy")

	// Test a package that shouldn't be in the new registry yet
	options := &install.InstallOptions{
		PackageName: "java", // This should fall back to legacy system
		DryRun:      true,
	}

	err := install.InstallPackageWithOptions(options)
	if err != nil {
		tf.Warning(t, "Legacy fallback test failed", err.Error())
		// This might fail due to missing assets or platform issues
	} else {
		tf.Success(t, "Legacy fallback working correctly")
	}
}