package analyze

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/matthieukhl/latentia/internal/rag"
)

// PromptBuilder creates context-aware optimization prompts using RAG
type PromptBuilder struct {
	docStore *rag.DocumentStore
}

func NewPromptBuilder(docStore *rag.DocumentStore) *PromptBuilder {
	return &PromptBuilder{
		docStore: docStore,
	}
}

// BuildOptimizationPrompt creates a comprehensive prompt for SQL optimization
func (pb *PromptBuilder) BuildOptimizationPrompt(sql string, pattern QueryPattern) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Build search query based on pattern analysis
	searchQuery := pb.buildSearchQuery(pattern)
	
	// Retrieve relevant documentation context
	context, err := pb.docStore.Search(ctx, searchQuery, 3)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve context: %w", err)
	}
	
	// Build the complete prompt
	prompt := pb.buildPromptTemplate(sql, pattern, context)
	
	return prompt, nil
}

// buildSearchQuery creates a search query based on the detected pattern
func (pb *PromptBuilder) buildSearchQuery(pattern QueryPattern) string {
	queryParts := []string{}
	
	// Add pattern type to search
	switch pattern.Type {
	case "complex-join", "simple-join":
		queryParts = append(queryParts, "JOIN optimization performance")
	case "aggregation":
		queryParts = append(queryParts, "GROUP BY aggregation optimization")
	case "pattern-search":
		queryParts = append(queryParts, "LIKE pattern search index optimization")
	case "full-select":
		queryParts = append(queryParts, "SELECT * column projection optimization")
	default:
		queryParts = append(queryParts, "SQL query optimization performance")
	}
	
	// Add anti-patterns to search
	for _, antiPattern := range pattern.AntiPatterns {
		switch antiPattern {
		case "leading-wildcard-like":
			queryParts = append(queryParts, "wildcard LIKE index")
		case "cartesian-join":
			queryParts = append(queryParts, "Cartesian product JOIN")
		case "missing-limit":
			queryParts = append(queryParts, "LIMIT result set")
		case "subquery-instead-of-join":
			queryParts = append(queryParts, "subquery JOIN conversion")
		}
	}
	
	// Add keywords for more context
	for _, keyword := range pattern.Keywords {
		queryParts = append(queryParts, keyword+" optimization")
	}
	
	return strings.Join(queryParts, " ")
}

// buildPromptTemplate constructs the complete optimization prompt
func (pb *PromptBuilder) buildPromptTemplate(sql string, pattern QueryPattern, context []rag.SearchResult) string {
	var prompt strings.Builder
	
	// System context
	prompt.WriteString("You are a TiDB performance expert specializing in SQL optimization. ")
	prompt.WriteString("Analyze the provided slow query and suggest concrete optimizations.\n\n")
	
	// Query information
	prompt.WriteString("SLOW QUERY ANALYSIS:\n")
	prompt.WriteString(fmt.Sprintf("Query Type: %s\n", pattern.Type))
	prompt.WriteString(fmt.Sprintf("Complexity: %s\n", pattern.Complexity))
	prompt.WriteString(fmt.Sprintf("Tables: %s\n", strings.Join(pattern.Tables, ", ")))
	
	if len(pattern.AntiPatterns) > 0 {
		prompt.WriteString(fmt.Sprintf("Anti-patterns detected: %s\n", strings.Join(pattern.AntiPatterns, ", ")))
	}
	
	if len(pattern.OptimizationOps) > 0 {
		prompt.WriteString(fmt.Sprintf("Optimization opportunities: %s\n", strings.Join(pattern.OptimizationOps, ", ")))
	}
	prompt.WriteString("\n")
	
	// Original query
	prompt.WriteString("ORIGINAL QUERY:\n```sql\n")
	prompt.WriteString(sql)
	prompt.WriteString("\n```\n\n")
	
	// Relevant documentation context
	if len(context) > 0 {
		prompt.WriteString("RELEVANT TIDB OPTIMIZATION KNOWLEDGE:\n")
		for i, result := range context {
			prompt.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, result.Document, result.Category))
			prompt.WriteString(fmt.Sprintf("   %s\n\n", result.Text))
		}
	}
	
	// Instructions
	prompt.WriteString("INSTRUCTIONS:\n")
	prompt.WriteString("Based on the query analysis and TiDB optimization knowledge above, provide a comprehensive optimization.\n")
	prompt.WriteString("Focus on the detected anti-patterns and optimization opportunities.\n\n")
	
	// Response format
	prompt.WriteString("FORMAT YOUR RESPONSE EXACTLY AS FOLLOWS:\n\n")
	prompt.WriteString("PROPOSED_SQL:\n```sql\n[Your optimized query here]\n```\n\n")
	prompt.WriteString("RATIONALE:\n")
	prompt.WriteString("• [Primary optimization applied]\n")
	prompt.WriteString("• [Secondary improvements made]\n")
	prompt.WriteString("• [Why this approach was chosen]\n\n")
	prompt.WriteString("EXPECTED_PLAN_CHANGE:\n")
	prompt.WriteString("• [Index usage improvements]\n")
	prompt.WriteString("• [Join order optimizations]\n")
	prompt.WriteString("• [Row reduction techniques]\n\n")
	prompt.WriteString("CAVEATS:\n")
	prompt.WriteString("• [Any semantic differences]\n")
	prompt.WriteString("• [Performance assumptions made]\n")
	prompt.WriteString("• [Edge cases to monitor]\n\n")
	
	// Specific guidance based on pattern
	prompt.WriteString("OPTIMIZATION FOCUS:\n")
	switch pattern.Type {
	case "complex-join", "simple-join":
		prompt.WriteString("- Optimize JOIN order and algorithms\n")
		prompt.WriteString("- Ensure proper index usage on join columns\n")
		prompt.WriteString("- Consider converting subqueries to JOINs\n")
	case "aggregation":
		prompt.WriteString("- Optimize GROUP BY and ORDER BY performance\n")
		prompt.WriteString("- Use appropriate indexes for aggregation\n")
		prompt.WriteString("- Consider pre-filtering with WHERE clauses\n")
	case "pattern-search":
		prompt.WriteString("- Optimize LIKE patterns for index usage\n")
		prompt.WriteString("- Avoid leading wildcards when possible\n")
		prompt.WriteString("- Consider full-text search alternatives\n")
	case "sleep-test":
		prompt.WriteString("- Remove artificial delays (SLEEP functions)\n")
		prompt.WriteString("- Replace with efficient query patterns\n")
		prompt.WriteString("- Ensure minimal resource usage\n")
	default:
		prompt.WriteString("- Apply general SQL optimization principles\n")
		prompt.WriteString("- Focus on index usage and query structure\n")
		prompt.WriteString("- Minimize data processing overhead\n")
	}
	
	return prompt.String()
}