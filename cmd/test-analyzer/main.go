package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/matthieukhl/latentia/internal/analyze"
	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/llm"
	"github.com/matthieukhl/latentia/internal/rag"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup schema
	if err := db.SetupTestSchema(); err != nil {
		log.Fatalf("Failed to setup schema: %v", err)
	}

	// Initialize LLM providers
	embedder, err := llm.NewEmbedder(&cfg.LLM)
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	generator, err := llm.NewGenerator(&cfg.LLM)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Initialize document store and seed it
	docStore := rag.NewDocumentStore(db, embedder)
	fmt.Println("Seeding TiDB optimization documentation...")
	if err := docStore.SeedTiDBOptimizationDocs(); err != nil {
		log.Fatalf("Failed to seed documentation: %v", err)
	}

	// Initialize optimization engine
	engine := analyze.NewOptimizationEngine(db, docStore, generator)

	// Test SQL queries with various patterns
	testQueries := []struct {
		name string
		sql  string
	}{
		{
			name: "SELECT * Anti-pattern",
			sql:  "SELECT * FROM customers c JOIN orders o ON c.id = o.customer_id WHERE c.city = 'Paris'",
		},
		{
			name: "Leading Wildcard LIKE",
			sql:  "SELECT id, name FROM products WHERE name LIKE '%widget%' ORDER BY created_at",
		},
		{
			name: "Cartesian Product Risk",
			sql:  "SELECT c.name, p.name FROM customers c, products p WHERE c.city = 'London'",
		},
		{
			name: "Missing LIMIT on Large Result",
			sql:  "SELECT c.*, o.* FROM customers c JOIN orders o ON c.id = o.customer_id ORDER BY o.created_at",
		},
		{
			name: "Complex Multi-table Join",
			sql:  "SELECT c.email, o.total, p.name, oi.quantity FROM customers c JOIN orders o ON c.id = o.customer_id JOIN order_items oi ON o.id = oi.order_id JOIN products p ON oi.product_id = p.id WHERE c.city = 'New York' AND o.status = 'shipped' ORDER BY o.created_at",
		},
	}

	ctx := context.Background()

	for i, test := range testQueries {
		fmt.Printf("\n=== Testing Query %d: %s ===\n", i+1, test.name)
		fmt.Printf("SQL: %s\n", test.sql)

		// First, insert a slow query record
		slowQueryResult, err := db.Exec(`
			INSERT INTO app_slow_queries (digest, sample_sql, started_at, query_time, db, source, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, fmt.Sprintf("test_%d", i), test.sql, time.Now(), 2.5, "latentia", "generated", "pending")
		
		if err != nil {
			log.Printf("Failed to insert slow query: %v", err)
			continue
		}

		slowQueryID, err := slowQueryResult.LastInsertId()
		if err != nil {
			log.Printf("Failed to get slow query ID: %v", err)
			continue
		}

		// Run optimization
		fmt.Println("\nAnalyzing patterns...")
		result, err := engine.OptimizeQuery(ctx, slowQueryID, test.sql)
		if err != nil {
			log.Printf("Failed to optimize query: %v", err)
			continue
		}

		// Display results
		fmt.Printf("\nOptimization Results:\n")
		fmt.Printf("Query Type: %s\n", result.Pattern.Type)
		fmt.Printf("Complexity: %s\n", result.Pattern.Complexity)
		fmt.Printf("Tables: %v\n", result.Pattern.Tables)
		fmt.Printf("Anti-patterns: %v\n", result.Pattern.AntiPatterns)
		fmt.Printf("Optimization Ops: %v\n", result.Pattern.OptimizationOps)
		fmt.Printf("Confidence Score: %.2f\n", result.ConfidenceScore)
		fmt.Printf("\nOptimized SQL:\n%s\n", result.OptimizedSQL)
		fmt.Printf("\nRationale:\n%s\n", result.Rationale)
		fmt.Printf("\nExpected Improvement:\n%s\n", result.ExpectedImprovement)
		if result.Caveats != "" {
			fmt.Printf("\nCaveats:\n%s\n", result.Caveats)
		}

		// Small delay to respect API rate limits
		time.Sleep(2 * time.Second)
	}

	// Show pending optimizations
	fmt.Printf("\n=== Pending Optimizations Summary ===\n")
	pending, err := engine.ListPendingOptimizations(ctx, 10)
	if err != nil {
		log.Printf("Failed to list pending optimizations: %v", err)
		return
	}

	for i, opt := range pending {
		fmt.Printf("%d. ID: %d, Confidence: %.2f, Type: %s\n", 
			i+1, opt.ID, opt.ConfidenceScore, opt.Pattern.Type)
	}

	fmt.Println("\nSQL Analysis Engine test completed successfully!")
}