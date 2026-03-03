package comment

import "errors"

var (
	ErrNotFound        = errors.New("comment not found")
	ErrAlreadyLiked    = errors.New("comment already liked")
	ErrNotLiked        = errors.New("comment not liked")
	ErrCannotDelete    = errors.New("cannot delete comment with replies")
	ErrUnauthorized    = errors.New("unauthorized to modify comment")
)
