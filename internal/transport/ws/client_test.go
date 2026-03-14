package ws

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsOriginAllowed(t *testing.T) {
	t.Cleanup(func() {
		SetAllowedOrigins(nil)
	})

	t.Run("allows exact configured origin", func(t *testing.T) {
		SetAllowedOrigins([]string{"https://app.example.com"})

		req := httptest.NewRequest("GET", "http://api.example.com/ws/chat", nil)
		req.Header.Set("Origin", "https://app.example.com")

		assert.True(t, isOriginAllowed(req))
	})

	t.Run("allows wildcard origin", func(t *testing.T) {
		SetAllowedOrigins([]string{"*"})

		req := httptest.NewRequest("GET", "http://api.example.com/ws/chat", nil)
		req.Header.Set("Origin", "https://anywhere.example.com")

		assert.True(t, isOriginAllowed(req))
	})

	t.Run("allows requests without origin header", func(t *testing.T) {
		SetAllowedOrigins([]string{"https://app.example.com"})

		req := httptest.NewRequest("GET", "http://api.example.com/ws/chat", nil)

		assert.True(t, isOriginAllowed(req))
	})

	t.Run("rejects invalid origin", func(t *testing.T) {
		SetAllowedOrigins([]string{"https://app.example.com"})

		req := httptest.NewRequest("GET", "http://api.example.com/ws/chat", nil)
		req.Header.Set("Origin", "://bad-origin")

		assert.False(t, isOriginAllowed(req))
	})

	t.Run("rejects browser origin when allowlist is empty", func(t *testing.T) {
		SetAllowedOrigins(nil)

		req := httptest.NewRequest("GET", "http://api.example.com/ws/chat", nil)
		req.Header.Set("Origin", "https://app.example.com")

		assert.False(t, isOriginAllowed(req))
	})
}
