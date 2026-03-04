package usecase_test

import (
	"testing"

	"github.com/studio/platform/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
)

// TestPasswordHashing tests password hashing and verification
func TestPasswordHashing(t *testing.T) {
	password := "Test123456"

	// Hash password
	hash, err := crypto.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Verify correct password
	valid, err := crypto.VerifyPassword(password, hash)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Verify incorrect password
	valid2, err2 := crypto.VerifyPassword("WrongPassword", hash)
	assert.NoError(t, err2)
	assert.False(t, valid2)
}
