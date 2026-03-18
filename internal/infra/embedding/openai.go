package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatibleEmbedder calls an OpenAI-compatible embeddings endpoint.
type OpenAICompatibleEmbedder struct {
	baseURL    string
	apiKey     string
	model      string
	dimensions int
	httpClient *http.Client
}

// NewOpenAICompatibleEmbedder creates a new embedding client.
func NewOpenAICompatibleEmbedder(baseURL, apiKey, model string, dimensions int, timeout time.Duration) *OpenAICompatibleEmbedder {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if strings.TrimSpace(model) == "" {
		model = "text-embedding-3-small"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &OpenAICompatibleEmbedder{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		dimensions: dimensions,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Configured reports whether the embedder can call an upstream provider.
func (e *OpenAICompatibleEmbedder) Configured() bool {
	return e != nil && e.apiKey != ""
}

// Dims returns the configured dimensionality.
func (e *OpenAICompatibleEmbedder) Dims() int {
	if e == nil {
		return 0
	}
	return e.dimensions
}

// Embed returns a vector embedding for the input text.
func (e *OpenAICompatibleEmbedder) Embed(text string) ([]float64, error) {
	if !e.Configured() {
		return nil, fmt.Errorf("embedding provider api key is not configured")
	}

	payload := map[string]any{
		"model": e.model,
		"input": text,
	}
	if e.dimensions > 0 {
		payload["dimensions"] = e.dimensions
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embedding provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("embedding provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embedding provider returned empty vector")
	}
	if e.dimensions == 0 {
		e.dimensions = len(result.Data[0].Embedding)
	}
	return result.Data[0].Embedding, nil
}
