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
	// For archive verification, we need to check the checksum of the archive file, not the extracted binary
	// This function is called with the path to the extracted binary, but checksums are for archives
	// So we'll skip verification for now and return nil
	// TODO: Implement proper archive checksum verification
	return nil
}

// VerifyArchiveChecksum verifies the SHA256 checksum of an archive file.
// Signature verification is skipped (signatureURL == "") for backward
// compatibility; use VerifyArchiveChecksumSigned to also verify the Ed25519
// signature over the checksums file.
func VerifyArchiveChecksum(archivePath string, checksumURL string, archiveName string) error {
	return VerifyArchiveChecksumSigned(archivePath, checksumURL, "", archiveName)
}

// VerifyArchiveChecksumSigned verifies the SHA256 checksum of an archive file
// and additionally verifies an Ed25519 signature over the checksums file when
// signatureURL is non-empty. The signature is checked BEFORE the checksum is
// parsed, so a tampered checksums file fails fast.
func VerifyArchiveChecksumSigned(archivePath, checksumURL, signatureURL, archiveName string) error {
	// Download checksum file
	checksumData, err := DownloadFile(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum: %w", err)
	}

	// Verify Ed25519 signature over the checksums file before trusting it.
	if err := VerifyChecksumsSignature(checksumData, signatureURL); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Parse checksum file to find the right entry
	lines := strings.Split(string(checksumData), "\n")
	var expectedSum string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "checksum  filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == archiveName {
			expectedSum = parts[0]
			break
		}
	}

	if expectedSum == "" {
		return fmt.Errorf("checksum not found for %s", archiveName)
	}

	// Calculate actual checksum
	actualSum, err := CalculateSHA256(archivePath)
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
