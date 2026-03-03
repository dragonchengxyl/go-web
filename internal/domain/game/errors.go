package game

import "errors"

var (
	ErrNotFound      = errors.New("game not found")
	ErrSlugExists    = errors.New("game slug already exists")
	ErrInvalidStatus = errors.New("invalid game status")
)
