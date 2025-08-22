package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerifyChecksum verifies the SHA256 checksum of a file
func VerifyChecksum(filepath string, checksumURL string) error {
	// Download checksum file
	checksumData, err := DownloadFile(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum: %w", err)
	}
	
	// Parse expected checksum
	expectedSum := strings.TrimSpace(string(checksumData))
	// Handle format: "sha256sum  filename" or just "sha256sum"
	if idx := strings.Index(expectedSum, " "); idx > 0 {
		expectedSum = expectedSum[:idx]
	}
	
	// Calculate actual checksum
	actualSum, err := CalculateSHA256(filepath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	// Compare checksums
	if actualSum != expectedSum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSum, actualSum)
	}
	
	return nil
}

// CalculateSHA256 calculates the SHA256 checksum of a file
func CalculateSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GenerateChecksum generates a SHA256 checksum file for a binary
func GenerateChecksum(binaryPath string, outputPath string) error {
	checksum, err := CalculateSHA256(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	// Write checksum to file
	content := fmt.Sprintf("%s  %s\n", checksum, binaryPath)
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write checksum file: %w", err)
	}
	
	return nil
}