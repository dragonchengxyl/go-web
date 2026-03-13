package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatMessage is an OpenAI-compatible chat completion message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAICompatibleClient streams tokens from any OpenAI-compatible provider.
type OpenAICompatibleClient struct {
	baseURL     string
	apiKey      string
	model       string
	temperature float64
	httpClient  *http.Client
}

// NewOpenAICompatibleClient creates a streaming chat client.
func NewOpenAICompatibleClient(baseURL, apiKey, model string, temperature float64, timeout time.Duration) *OpenAICompatibleClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	if strings.TrimSpace(model) == "" {
		model = "deepseek-chat"
	}
	if timeout <= 0 {
		timeout = 90 * time.Second
	}

	return &OpenAICompatibleClient{
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiKey:      strings.TrimSpace(apiKey),
		model:       model,
		temperature: temperature,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Configured reports whether the upstream provider can be called.
func (c *OpenAICompatibleClient) Configured() bool {
	return c != nil && c.apiKey != ""
}

// StreamChat streams assistant tokens from the configured provider.
func (c *OpenAICompatibleClient) StreamChat(ctx context.Context, messages []ChatMessage, onToken func(string) error) error {
	if !c.Configured() {
		return fmt.Errorf("assistant provider api key is not configured")
	}

	payload := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"temperature": c.temperature,
		"stream":      true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call llm provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("llm provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read llm stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if raw == "[DONE]" {
			return nil
		}

		var chunk struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error,omitempty"`
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
			return fmt.Errorf("decode llm stream chunk: %w", err)
		}
		if chunk.Error != nil {
			return fmt.Errorf("llm provider error: %s", chunk.Error.Message)
		}

		for _, choice := range chunk.Choices {
			token := choice.Delta.Content
			if token == "" {
				token = choice.Delta.ReasoningContent
			}
			if token == "" {
				continue
			}
			if err := onToken(token); err != nil {
				return err
			}
		}
	}
}
