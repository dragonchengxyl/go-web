package asset

import "errors"

var (
	ErrNotFound       = errors.New("asset not found")
	ErrAlreadyOwned   = errors.New("asset already owned")
	ErrNoPermission   = errors.New("no permission to access this asset")
	ErrDownloadLimit  = errors.New("download limit exceeded")
)
