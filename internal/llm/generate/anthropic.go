package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/matthieukhl/latentia/internal/types"
)

type AnthropicGenerator struct {
	apiKey string
	model  string
	client *http.Client
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func NewAnthropicGenerator(model string, apiKeyEnv string, directAPIKey string) (*AnthropicGenerator, error) {
	var apiKey string
	
	// First try direct API key from config
	if directAPIKey != "" {
		apiKey = directAPIKey
	} else if apiKeyEnv != "" {
		// Fallback to environment variable
		apiKey = os.Getenv(apiKeyEnv)
	}
	
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found in config or environment variable %s", apiKeyEnv)
	}
	
	return &AnthropicGenerator{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (g *AnthropicGenerator) Complete(ctx context.Context, prompt string, opts map[string]any) (string, error) {
	maxTokens := 4000
	if val, ok := opts["max_tokens"].(int); ok && val > 0 {
		maxTokens = val
	}
	
	system := "You are a TiDB performance expert specializing in SQL optimization."
	if val, ok := opts["system"].(string); ok && val != "" {
		system = val
	}
	
	req := anthropicRequest{
		Model:     g.model,
		MaxTokens: maxTokens,
		System:    system,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", g.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}
	
	return response.Content[0].Text, nil
}

func (g *AnthropicGenerator) Model() string {
	return g.model
}

// Compile-time interface check
var _ types.Generator = (*AnthropicGenerator)(nil)