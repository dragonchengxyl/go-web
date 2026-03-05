package achievement

import "errors"

var (
	ErrNotFound      = errors.New("achievement not found")
	ErrAlreadyUnlocked = errors.New("achievement already unlocked")
	ErrInsufficientPoints = errors.New("insufficient points")
)
