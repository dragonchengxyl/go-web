package music

import "errors"

var (
	ErrNotFound      = errors.New("music not found")
	ErrSlugExists    = errors.New("album slug already exists")
	ErrInvalidFormat = errors.New("invalid audio format")
)
