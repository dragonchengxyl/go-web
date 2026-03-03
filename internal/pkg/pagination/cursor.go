package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cursor represents a pagination cursor
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// EncodeCursor encodes a cursor to base64 string
func EncodeCursor(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%d:%s", createdAt.Unix(), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes a base64 cursor string
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor structure")
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor timestamp: %w", err)
	}

	return &Cursor{
		CreatedAt: time.Unix(timestamp, 0),
		ID:        parts[1],
	}, nil
}

// PageInfo represents pagination metadata
type PageInfo struct {
	HasNextPage bool   `json:"has_next_page"`
	NextCursor  string `json:"next_cursor,omitempty"`
}
