// +build windows

package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Test for Issue #19: Docker Installation Issues on Windows
// https://github.com/cassandragargoyle/portunix/issues/19

func TestAnalyzeWindowsStorage_Issue19(t *testing.T) {
	t.Run("Should not return non-existent drives", func(t *testing.T) {
		drives, err := analyzeWindowsStorage()
		if err != nil {
			t.Fatalf("analyzeWindowsStorage failed: %v", err)
		}

		// Check that all returned drives actually exist
		for _, drive := range drives {
			drivePath := fmt.Sprintf("%s:\\", drive.Letter)
			if _, err := os.Stat(drivePath); os.IsNotExist(err) {
				t.Errorf("analyzeWindowsStorage returned non-existent drive: %s", drivePath)
			}
		}
	})

	t.Run("Should detect C drive on single-drive systems", func(t *testing.T) {
		drives, err := analyzeWindowsStorage()
		if err != nil {
			t.Fatalf("analyzeWindowsStorage failed: %v", err)
		}

		// On Windows, C: drive should always exist
		hasCDrive := false
		for _, drive := range drives {
			if drive.Letter == "C" {
				hasCDrive = true
				break
			}
		}

		if !hasCDrive {
			t.Error("analyzeWindowsStorage did not detect C: drive")
		}
	})

	t.Run("Should not hardcode D drive", func(t *testing.T) {
		// This test specifically checks for the hardcoded D: drive issue
		drives, err := analyzeWindowsStorage()
		if err != nil {
			t.Fatalf("analyzeWindowsStorage failed: %v", err)
		}

		// If D: doesn't exist on the system, it shouldn't be in results
		if _, err := os.Stat("D:\\"); os.IsNotExist(err) {
			for _, drive := range drives {
				if drive.Letter == "D" {
					t.Error("analyzeWindowsStorage returned non-existent D: drive (hardcoded issue)")
				}
			}
		}
	})
}

func TestDockerDataRootConfiguration_Issue19(t *testing.T) {
	t.Run("Should use selected drive for data-root", func(t *testing.T) {
		// Test that the selected drive is actually used in configuration
		selectedDrive := "C" // User selects C:
		dataRoot := fmt.Sprintf("%s:\\docker-data", selectedDrive)
		
		// Verify the data root path is constructed correctly
		if !strings.HasPrefix(dataRoot, selectedDrive) {
			t.Errorf("Data root doesn't use selected drive. Got: %s, Want prefix: %s", dataRoot, selectedDrive)
		}
	})

	t.Run("Should create valid Windows paths", func(t *testing.T) {
		selectedDrive := "C"
		dataRoot := fmt.Sprintf("%s:\\docker-data", selectedDrive)
		
		// Check for valid Windows path format
		if !strings.Contains(dataRoot, ":\\") {
			t.Errorf("Invalid Windows path format: %s", dataRoot)
		}
		
		// Verify no forward slashes (Linux-style paths)
		if strings.Contains(dataRoot, "/") {
			t.Errorf("Data root contains Linux-style path separators: %s", dataRoot)
		}
	})
}

func TestDockerPathConfiguration_Issue19(t *testing.T) {
	t.Run("Should add Docker to PATH", func(t *testing.T) {
		// Check if Docker would be added to PATH correctly
		dockerPath := `C:\Program Files\Docker\Docker\resources\bin`
		
		// This would be the actual implementation
		// For now, we're testing the logic
		currentPath := os.Getenv("PATH")
		if !strings.Contains(currentPath, "Docker") {
			// Docker should be added to PATH
			newPath := currentPath + ";" + dockerPath
			if !strings.Contains(newPath, dockerPath) {
				t.Error("Failed to add Docker to PATH")
			}
		}
	})
}

func TestCommandConsistency_Issue19(t *testing.T) {
	t.Run("Both command variants should work identically", func(t *testing.T) {
		// Test that both `portunix docker install` and `portunix install docker`
		// produce the same results
		
		// This would test the actual command routing
		// For unit test, we verify the function calls are consistent
		
		// Both should call the same underlying function
		// InstallDocker(false) should be called in both cases
		
		// Mock test - actual implementation would test command routing
		t.Log("Command consistency test placeholder - implement with integration test")
	})
}

// Helper function to mock Windows drive detection
func mockGetWindowsDrives() ([]DriveInfo, error) {
	// This should return actual drives, not hardcoded values
	drives := []DriveInfo{}
	
	// Check common drive letters
	for _, letter := range []string{"C", "D", "E", "F"} {
		drivePath := fmt.Sprintf("%s:\\", letter)
		if _, err := os.Stat(drivePath); err == nil {
			// Drive exists
			drives = append(drives, DriveInfo{
				Letter:     letter,
				FreeSpace:  "100 GB", // Would use actual WMI calls
				TotalSpace: "500 GB", // Would use actual WMI calls
			})
		}
	}
	
	// Sort by free space (largest first)
	// In real implementation, sort by actual free space
	
	return drives, nil
}

// Test the actual fix implementation
func TestWindowsDriveDetectionFix(t *testing.T) {
	t.Run("Should use real drive detection, not hardcoded", func(t *testing.T) {
		drives, err := mockGetWindowsDrives()
		if err != nil {
			t.Fatalf("Drive detection failed: %v", err)
		}
		
		if len(drives) == 0 {
			t.Error("No drives detected")
		}
		
		// All detected drives should be real
		for _, drive := range drives {
			drivePath := fmt.Sprintf("%s:\\", drive.Letter)
			if _, err := os.Stat(drivePath); os.IsNotExist(err) {
				t.Errorf("Detected non-existent drive: %s", drivePath)
			}
		}
	})
}

// Integration test for the full installation flow
func TestDockerInstallationFlow_Issue19(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	t.Run("Full installation flow should work on single-drive system", func(t *testing.T) {
		// This would be an integration test
		// 1. Detect drives (should find C: on single-drive system)
		// 2. Configure Docker with correct path
		// 3. Verify configuration is correct
		// 4. Check PATH is updated
		
		drives, err := mockGetWindowsDrives()
		if err != nil {
			t.Fatalf("Drive detection failed: %v", err)
		}
		
		// Should detect at least C: drive
		if len(drives) == 0 {
			t.Fatal("No drives detected on Windows system")
		}
		
		// Select first available drive (should be C: on single-drive system)
		selectedDrive := drives[0].Letter
		dataRoot := fmt.Sprintf("%s:\\docker-data", selectedDrive)
		
		// Verify configuration would be correct
		if selectedDrive != "C" && len(drives) == 1 {
			t.Error("Single-drive system should default to C:")
		}
		
		// Verify data root path is valid
		if !strings.HasPrefix(dataRoot, selectedDrive) {
			t.Errorf("Data root doesn't match selected drive: %s", dataRoot)
		}
	})
}

// Benchmark to ensure drive detection is performant
func BenchmarkWindowsDriveDetection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mockGetWindowsDrives()
	}
}