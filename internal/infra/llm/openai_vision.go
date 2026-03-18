package llm

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

// OpenAICompatibleVisionClient calls an OpenAI-compatible multimodal chat endpoint.
type OpenAICompatibleVisionClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAICompatibleVisionClient(baseURL, apiKey, model string, timeout time.Duration) *OpenAICompatibleVisionClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if strings.TrimSpace(model) == "" {
		model = "gpt-4o-mini"
	}
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return &OpenAICompatibleVisionClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *OpenAICompatibleVisionClient) Configured() bool {
	return c != nil && c.apiKey != ""
}

func (c *OpenAICompatibleVisionClient) AnalyzeImages(ctx context.Context, prompt string, imageURLs []string) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("vision provider api key is not configured")
	}
	content := []map[string]any{
		{
			"type": "text",
			"text": prompt,
		},
	}
	for _, imageURL := range imageURLs {
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]any{
				"url": imageURL,
			},
		})
	}

	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0.2,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal vision request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create vision request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call vision provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("vision provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode vision response: %w", err)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("vision provider returned empty content")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
