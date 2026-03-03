package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64MB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// Format: base64(salt):base64(hash)
	encoded := fmt.Sprintf("%s:%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, encoded string) (bool, error) {
	var salt, hash []byte

	_, err := fmt.Sscanf(encoded, "%s:%s", &salt, &hash)
	if err != nil {
		// Try base64 decoding
		parts := splitEncodedHash(encoded)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid hash format")
		}

		salt, err = base64.RawStdEncoding.DecodeString(parts[0])
		if err != nil {
			return false, fmt.Errorf("failed to decode salt: %w", err)
		}

		hash, err = base64.RawStdEncoding.DecodeString(parts[1])
		if err != nil {
			return false, fmt.Errorf("failed to decode hash: %w", err)
		}
	}

	computedHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

func splitEncodedHash(encoded string) []string {
	result := make([]string, 0, 2)
	start := 0
	for i, c := range encoded {
		if c == ':' {
			result = append(result, encoded[start:i])
			start = i + 1
		}
	}
	if start < len(encoded) {
		result = append(result, encoded[start:])
	}
	return result
}
