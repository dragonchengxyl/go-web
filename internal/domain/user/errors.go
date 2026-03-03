package user

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailExists   = errors.New("email already exists")
	ErrUsernameExists = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountSuspended = errors.New("account suspended")
	ErrAccountBanned = errors.New("account banned")
)
