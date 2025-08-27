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

type OpenAIGenerator struct {
	apiKey string
	model  string
	client *http.Client
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAIGenerator(model string, apiKeyEnv string, directAPIKey string) (*OpenAIGenerator, error) {
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
	
	return &OpenAIGenerator{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (g *OpenAIGenerator) Complete(ctx context.Context, prompt string, opts map[string]any) (string, error) {
	maxTokens := 4000
	if val, ok := opts["max_tokens"].(int); ok && val > 0 {
		maxTokens = val
	}
	
	temperature := 0.7
	if val, ok := opts["temperature"].(float64); ok {
		temperature = val
	}
	
	system := "You are a TiDB performance expert specializing in SQL optimization."
	if val, ok := opts["system"].(string); ok && val != "" {
		system = val
	}
	
	messages := []openAIMessage{
		{
			Role:    "system",
			Content: system,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	req := openAIRequest{
		Model:       g.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
	
	if val, ok := opts["top_p"].(float64); ok {
		req.TopP = val
	}
	
	if val, ok := opts["stop"].([]string); ok {
		req.Stop = val
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.apiKey))
	
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	
	return response.Choices[0].Message.Content, nil
}

func (g *OpenAIGenerator) Model() string {
	return g.model
}

// Compile-time interface check
var _ types.Generator = (*OpenAIGenerator)(nil)