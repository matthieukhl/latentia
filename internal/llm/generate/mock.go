package generate

import (
	"context"
	"strings"
	"time"

	"github.com/matthieukhl/latentia/internal/types"
)

type MockGenerator struct {
	model string
}

func NewMockGenerator(model string) *MockGenerator {
	return &MockGenerator{model: model}
}

func (g *MockGenerator) Complete(ctx context.Context, prompt string, opts map[string]any) (string, error) {
	// Simulate API delay
	time.Sleep(500 * time.Millisecond)
	
	// Generate contextual response based on the prompt content
	prompt = strings.ToLower(prompt)
	
	if strings.Contains(prompt, "join") {
		return g.generateJoinOptimization(prompt), nil
	}
	
	if strings.Contains(prompt, "select *") {
		return g.generateSelectOptimization(prompt), nil
	}
	
	if strings.Contains(prompt, "group by") || strings.Contains(prompt, "order by") {
		return g.generateAggregationOptimization(prompt), nil
	}
	
	if strings.Contains(prompt, "sleep") {
		return g.generateSleepOptimization(prompt), nil
	}
	
	// Default optimization response
	return g.generateGenericOptimization(prompt), nil
}

func (g *MockGenerator) Model() string {
	return g.model + "-mock"
}

func (g *MockGenerator) generateJoinOptimization(prompt string) string {
	return `PROPOSED_SQL:
SELECT c.email, o.total, p.name 
FROM customers c 
INNER JOIN orders o ON c.id = o.customer_id 
INNER JOIN order_items oi ON o.id = oi.order_id
INNER JOIN products p ON oi.product_id = p.id
WHERE c.city = 'New York'
LIMIT 100;

RATIONALE:
• Added LIMIT to reduce result set size
• Used INNER JOIN instead of implicit comma joins for better readability
• Specific equality condition on indexed column (city)

EXPECTED_PLAN_CHANGE:
• Index usage on customers.city for faster filtering
• Proper join execution plan with smaller intermediate results
• Reduced memory usage with LIMIT clause

CAVEATS:
• Result set is now limited to 100 rows - verify this meets business requirements
• INNER JOIN semantics may exclude customers without orders`
}

func (g *MockGenerator) generateSelectOptimization(prompt string) string {
	return `PROPOSED_SQL:
SELECT c.id, c.email, c.first_name, c.last_name
FROM customers c 
WHERE c.email LIKE 'john%@%'
AND c.created_at > '2024-01-01'
LIMIT 50;

RATIONALE:
• Replaced SELECT * with specific columns to reduce data transfer
• More specific LIKE pattern to use index prefix when available
• Added date filter to reduce scan range
• Added LIMIT to control result size

EXPECTED_PLAN_CHANGE:
• Reduced I/O with column projection
• Better index utilization with specific patterns
• Faster query execution with smaller result sets

CAVEATS:
• Specific column selection requires maintenance if schema changes
• LIKE patterns with leading wildcards still require full scan`
}

func (g *MockGenerator) generateAggregationOptimization(prompt string) string {
	return `PROPOSED_SQL:
SELECT c.city, c.country, 
       COUNT(*) as customer_count,
       ROUND(AVG(o.total), 2) as avg_order_value
FROM customers c
INNER JOIN orders o ON c.id = o.customer_id
WHERE o.status IN ('paid', 'shipped', 'delivered')
GROUP BY c.city, c.country
HAVING COUNT(*) >= 5
ORDER BY customer_count DESC, avg_order_value DESC
LIMIT 20;

RATIONALE:
• Added WHERE filter to reduce data processed before aggregation
• Used HAVING to filter groups efficiently
• Added LIMIT to control result size
• Optimized ORDER BY to use aggregated columns

EXPECTED_PLAN_CHANGE:
• Pre-aggregation filtering reduces processing overhead
• More efficient GROUP BY execution with filtered data
• Better memory usage with HAVING clause

CAVEATS:
• HAVING clause changes result semantics by filtering small groups
• ORDER BY on aggregated columns may require additional sorting`
}

func (g *MockGenerator) generateSleepOptimization(prompt string) string {
	return `PROPOSED_SQL:
SELECT c.id, c.email 
FROM customers c
WHERE c.id = 1
LIMIT 1;

RATIONALE:
• Removed SLEEP() function which serves no business purpose
• Added specific WHERE condition on primary key for instant lookup
• Minimal column selection for fastest execution
• LIMIT 1 ensures single row result

EXPECTED_PLAN_CHANGE:
• Primary key lookup instead of function execution
• Eliminates artificial delay from SLEEP()
• Index-based point query with minimal cost

CAVEATS:
• Removes artificial delay - verify this was only for testing
• Specific ID filter may need adjustment based on requirements`
}

func (g *MockGenerator) generateGenericOptimization(prompt string) string {
	return `PROPOSED_SQL:
-- Analysis shows this query can be optimized with proper indexing

RATIONALE:
• Consider adding indexes on frequently queried columns
• Use EXPLAIN ANALYZE to identify bottlenecks
• Filter early with WHERE conditions on indexed columns
• Consider query rewriting for better performance

EXPECTED_PLAN_CHANGE:
• Improved execution plan with proper index usage
• Reduced table scan operations
• Better join algorithms selection

CAVEATS:
• Specific optimizations depend on actual query patterns
• Index recommendations need analysis of query frequency
• Performance gains vary with data size and distribution`
}

// Compile-time interface check
var _ types.Generator = (*MockGenerator)(nil)