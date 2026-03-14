package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/studio/platform/configs"
)

type apiEnvelope[T any] struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data"`
	RequestID string `json:"request_id"`
	Timestamp int64  `json:"timestamp"`
}

type optionalLoader struct {
	loaded bool
	cfg    *configs.Config
	err    error
}

var globalConfigCache = map[*Options]*optionalLoader{}

func (o *Options) loadConfig(required bool) (*configs.Config, error) {
	cache, ok := globalConfigCache[o]
	if !ok {
		cache = &optionalLoader{}
		globalConfigCache[o] = cache
	}
	if cache.loaded {
		if required && cache.err != nil {
			return nil, cache.err
		}
		return cache.cfg, nil
	}

	cache.loaded = true
	if strings.TrimSpace(o.ConfigPath) == "" {
		if required {
			cache.err = fmt.Errorf("config path is required")
			return nil, cache.err
		}
		return nil, nil
	}

	if _, err := os.Stat(o.ConfigPath); err != nil {
		if required {
			cache.err = fmt.Errorf("config file %q not found: %w", o.ConfigPath, err)
			return nil, cache.err
		}
		return nil, nil
	}

	cfg, err := configs.Load(o.ConfigPath)
	if err != nil {
		if required {
			cache.err = err
			return nil, err
		}
		return nil, nil
	}

	cache.cfg = cfg
	return cfg, nil
}

func (o *Options) serverBaseURL() string {
	if serverURL := strings.TrimSpace(o.ServerURL); serverURL != "" {
		return strings.TrimRight(serverURL, "/")
	}
	if cfg, err := o.loadConfig(false); err == nil && cfg != nil && cfg.Server.Port > 0 {
		return fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	}
	return "http://localhost:8080"
}

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func doJSON[T any](ctx context.Context, client *http.Client, method, url, token string, payload any, out *apiEnvelope[T]) (int, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if out != nil && out.Message != "" {
			return resp.StatusCode, fmt.Errorf("http %d: %s", resp.StatusCode, out.Message)
		}
		return resp.StatusCode, fmt.Errorf("http %d", resp.StatusCode)
	}
	if out != nil && out.Code != 0 {
		return resp.StatusCode, fmt.Errorf("api code %d: %s", out.Code, out.Message)
	}

	return resp.StatusCode, nil
}

func doRequest(ctx context.Context, client *http.Client, method, url, token string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("build request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resp.StatusCode, body, fmt.Errorf("http %d", resp.StatusCode)
	}
	return resp.StatusCode, body, nil
}

func writeSection(w io.Writer, title string) {
	fmt.Fprintf(w, "\n== %s ==\n", title)
}
