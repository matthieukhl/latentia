package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/llm"
	"github.com/spf13/cobra"
)

var testLLMCmd = &cobra.Command{
	Use:   "test-llm",
	Short: "Test LLM provider connections",
	Long: `Test connections to configured LLM providers (embedder and generator).
This helps verify API keys and connectivity before running the full optimization pipeline.`,
	RunE: testLLMProviders,
}

func init() {
	rootCmd.AddCommand(testLLMCmd)
}

func testLLMProviders(cmd *cobra.Command, args []string) error {
	fmt.Println("🧪 Testing LLM provider connections...")
	
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Test embedder
	fmt.Printf("🔤 Testing embedder (%s/%s)...\n", cfg.LLM.Embedder.Provider, cfg.LLM.Embedder.Model)
	embedder, err := llm.NewEmbedder(&cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create embedder: %w", err)
	}
	
	testTexts := []string{
		"SELECT * FROM orders WHERE created_at > '2024-01-01'",
		"Optimize this SQL query for better performance",
	}
	
	embeddings, err := embedder.Embed(ctx, testTexts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}
	
	fmt.Printf("   ✅ Generated %d embeddings, dimension: %d\n", len(embeddings), embedder.Dim())
	
	// Test generator
	fmt.Printf("🤖 Testing generator (%s/%s)...\n", cfg.LLM.Generator.Provider, cfg.LLM.Generator.Model)
	generator, err := llm.NewGenerator(&cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}
	
	testPrompt := `Please analyze this SQL query and suggest one optimization:

SELECT * FROM orders o 
JOIN customers c ON o.customer_id = c.id 
WHERE o.created_at > '2024-01-01'
ORDER BY o.created_at DESC;

Respond with just the optimization suggestion in 1-2 sentences.`
	
	response, err := generator.Complete(ctx, testPrompt, map[string]any{
		"max_tokens": 200,
		"system":     "You are a concise SQL optimization expert.",
	})
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}
	
	fmt.Printf("   ✅ Generated response: %s\n", response)
	
	fmt.Println("\n🎉 All LLM providers are working correctly!")
	return nil
}