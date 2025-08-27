package llm

import (
	"fmt"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/llm/embed"
	"github.com/matthieukhl/latentia/internal/llm/generate"
	"github.com/matthieukhl/latentia/internal/types"
)

// NewEmbedder creates an embedder based on configuration
func NewEmbedder(cfg *config.LLMConfig) (types.Embedder, error) {
	switch cfg.Embedder.Provider {
	case "openai":
		return embed.NewOpenAIEmbedder(cfg.Embedder.Model, cfg.Embedder.APIKeyEnv, cfg.Embedder.APIKey)
	case "mock":
		return embed.NewMockEmbedder(cfg.Embedder.Model, 1536), nil
	default:
		return nil, fmt.Errorf("unsupported embedder provider: %s", cfg.Embedder.Provider)
	}
}

// NewGenerator creates a generator based on configuration
func NewGenerator(cfg *config.LLMConfig) (types.Generator, error) {
	switch cfg.Generator.Provider {
	case "openai":
		return generate.NewOpenAIGenerator(cfg.Generator.Model, cfg.Generator.APIKeyEnv, cfg.Generator.APIKey)
	case "anthropic":
		return generate.NewAnthropicGenerator(cfg.Generator.Model, cfg.Generator.APIKeyEnv, cfg.Generator.APIKey)
	case "mock":
		return generate.NewMockGenerator(cfg.Generator.Model), nil
	default:
		return nil, fmt.Errorf("unsupported generator provider: %s", cfg.Generator.Provider)
	}
}