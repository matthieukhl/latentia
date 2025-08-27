package types

import "context"

// Embedder generates vector embeddings from text
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dim() int
	Model() string
}

// Generator produces text completions from prompts
type Generator interface {
	Complete(ctx context.Context, prompt string, opts map[string]any) (string, error)
	Model() string
}

// EmbeddingResult represents a text embedding with metadata
type EmbeddingResult struct {
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// GenerationOptions contains options for text generation
type GenerationOptions struct {
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}