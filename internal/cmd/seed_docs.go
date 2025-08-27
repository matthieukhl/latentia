package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/llm"
	"github.com/matthieukhl/latentia/internal/rag"
	"github.com/spf13/cobra"
)

var seedDocsCmd = &cobra.Command{
	Use:   "seed-docs",
	Short: "Seed TiDB optimization documentation for RAG",
	Long: `Populate the documentation store with curated TiDB optimization 
content and generate vector embeddings for semantic search.

This creates the knowledge base that the AI uses to provide context-aware
optimization suggestions.`,
	RunE: seedDocumentation,
}

func init() {
	rootCmd.AddCommand(seedDocsCmd)
}

func seedDocumentation(cmd *cobra.Command, args []string) error {
	fmt.Println("📚 Seeding TiDB optimization documentation...")
	
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	fmt.Println("🔤 Initializing embedder...")
	embedder, err := llm.NewEmbedder(&cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create embedder: %w", err)
	}
	
	fmt.Printf("   Using %s/%s (dimension: %d)\n", 
		cfg.LLM.Embedder.Provider, cfg.LLM.Embedder.Model, embedder.Dim())
	
	docStore := rag.NewDocumentStore(db, embedder)
	
	fmt.Println("📝 Adding TiDB optimization documentation...")
	err = docStore.SeedTiDBOptimizationDocs()
	if err != nil {
		return fmt.Errorf("failed to seed documentation: %w", err)
	}
	
	fmt.Println("🔍 Testing vector search...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	testQuery := "How to optimize slow JOIN queries?"
	results, err := docStore.Search(ctx, testQuery, 3)
	if err != nil {
		return fmt.Errorf("failed to test search: %w", err)
	}
	
	fmt.Printf("\n🎯 Test search results for: \"%s\"\n", testQuery)
	for i, result := range results {
		fmt.Printf("   %d. [%.3f] %s - %s\n", 
			i+1, result.Score, result.Document, result.Category)
		fmt.Printf("      %s\n", truncateText(result.Text, 100))
	}
	
	if len(results) == 0 {
		fmt.Println("   ⚠️  No results found - check embeddings and vector search setup")
	}
	
	fmt.Println("\n✅ Documentation seeding complete!")
	fmt.Println("💡 Ready to run optimization engine with RAG context")
	
	return nil
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}