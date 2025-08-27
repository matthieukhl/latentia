package embed

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

type OpenAIEmbedder struct {
	apiKey string
	model  string
	client *http.Client
}

type openAIEmbedRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type openAIEmbedResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAIEmbedder(model string, apiKeyEnv string, directAPIKey string) (*OpenAIEmbedder, error) {
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
	
	return &OpenAIEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}
	
	req := openAIEmbedRequest{
		Input: texts,
		Model: e.model,
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))
	
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response openAIEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	embeddings := make([][]float32, len(response.Data))
	for _, data := range response.Data {
		if data.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index %d", data.Index)
		}
		embeddings[data.Index] = data.Embedding
	}
	
	return embeddings, nil
}

func (e *OpenAIEmbedder) Dim() int {
	// text-embedding-3-small: 1536 dimensions
	// text-embedding-3-large: 3072 dimensions
	// text-embedding-ada-002: 1536 dimensions
	switch e.model {
	case "text-embedding-3-large":
		return 3072
	default:
		return 1536
	}
}

func (e *OpenAIEmbedder) Model() string {
	return e.model
}

// Compile-time interface check
var _ types.Embedder = (*OpenAIEmbedder)(nil)