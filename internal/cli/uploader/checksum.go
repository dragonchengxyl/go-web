package uploader

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// CalculateChecksum calculates SHA256 checksum of a file
func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// VerifyChecksum verifies file checksum matches expected value
func VerifyChecksum(filePath string, expectedChecksum string) (bool, error) {
	actualChecksum, err := CalculateChecksum(filePath)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}
