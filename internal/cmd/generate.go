package cmd

import (
	"fmt"
	"time"

	"github.com/matthieukhl/latentia/internal/config"
	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/ingest"
	"github.com/spf13/cobra"
)

var (
	queryType string
	duration  int
	count     int
	record    bool
)

var generateCmd = &cobra.Command{
	Use:   "generate-slow",
	Short: "Generate slow queries for testing",
	Long: `Generate intentionally slow queries that will appear in TiDB's 
INFORMATION_SCHEMA.SLOW_QUERY table. These can then be used to test 
the optimization engine.

Available query types:
- sleep: Uses SLEEP() function for guaranteed slow queries
- full-scan: Queries without proper indexes (table scans)
- complex-join: Inefficient JOIN patterns
- aggregation: Heavy GROUP BY/ORDER BY operations`,
	RunE: generateSlowQuery,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	
	generateCmd.Flags().StringVar(&queryType, "type", "sleep", "Type of slow query (sleep|full-scan|complex-join|aggregation)")
	generateCmd.Flags().IntVar(&duration, "duration", 2, "Duration in seconds for sleep-based queries")
	generateCmd.Flags().IntVar(&count, "count", 1, "Number of slow queries to generate")
	generateCmd.Flags().BoolVar(&record, "record", true, "Record slow queries to app_slow_queries table")
}

func generateSlowQuery(cmd *cobra.Command, args []string) error {
	fmt.Printf("🐌 Generating %d slow quer%s of type '%s'...\n", count, pluralize(count), queryType)
	if record {
		fmt.Println("📝 Recording slow queries to app_slow_queries table...")
	}
	
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	var ingester *ingest.SlowQueryIngester
	if record {
		ingester = ingest.NewSlowQueryIngester(db)
	}
	
	switch queryType {
	case "sleep":
		return generateSleepQueries(db, ingester, cfg)
	case "full-scan":
		return generateFullScanQueries(db, ingester, cfg)
	case "complex-join":
		return generateComplexJoinQueries(db, ingester, cfg)
	case "aggregation":
		return generateAggregationQueries(db, ingester, cfg)
	default:
		return fmt.Errorf("unknown query type: %s", queryType)
	}
}

func generateSleepQueries(db *database.DB, ingester *ingest.SlowQueryIngester, cfg *config.Config) error {
	fmt.Printf("   ⏰ Running SLEEP(%d) queries...\n", duration)
	
	for i := 0; i < count; i++ {
		query := fmt.Sprintf("SELECT SLEEP(%d), id, email FROM customers LIMIT 1", duration)
		
		start := time.Now()
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute sleep query %d: %w", i+1, err)
		}
		elapsed := time.Since(start)
		queryTime := elapsed.Seconds()
		
		// Record to app_slow_queries if enabled
		if ingester != nil && queryTime >= 0.1 { // Only record queries >= 100ms
			err = ingester.RecordGeneratedSlowQuery(query, start, queryTime, "latentia", "agent-generator")
			if err != nil {
				fmt.Printf("   ⚠️  Failed to record query %d: %v\n", i+1, err)
			} else {
				fmt.Printf("   📝 Query %d recorded to app_slow_queries\n", i+1)
			}
		}
		
		fmt.Printf("   ✅ Query %d completed in %v\n", i+1, elapsed)
		
		// Small delay between queries to avoid overwhelming the system
		if i < count-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	return nil
}

// executeAndRecord is a helper function to execute a query and optionally record it
func executeAndRecord(db *database.DB, ingester *ingest.SlowQueryIngester, query string, queryNum int) (time.Duration, error) {
	start := time.Now()
	rows, err := db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query %d: %w", queryNum, err)
	}
	
	// Count rows to force full execution
	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	rows.Close()
	
	elapsed := time.Since(start)
	queryTime := elapsed.Seconds()
	
	// Record to app_slow_queries if enabled and query was slow enough
	if ingester != nil && queryTime >= 0.01 { // Record queries >= 10ms for testing
		err = ingester.RecordGeneratedSlowQuery(query, start, queryTime, "latentia", "agent-generator")
		if err != nil {
			fmt.Printf("   ⚠️  Failed to record query %d: %v\n", queryNum, err)
		} else {
			fmt.Printf("   📝 Query %d recorded to app_slow_queries\n", queryNum)
		}
	}
	
	fmt.Printf("   ✅ Query %d: %d rows in %v\n", queryNum, rowCount, elapsed)
	return elapsed, nil
}

func generateFullScanQueries(db *database.DB, ingester *ingest.SlowQueryIngester, cfg *config.Config) error {
	fmt.Println("   🔍 Running full table scan queries...")
	
	queries := []string{
		// Search in product names (no index on name field)
		"SELECT * FROM products WHERE name LIKE '%Book%'",
		
		// Search in order notes (text field, no index)
		"SELECT * FROM orders WHERE notes LIKE '%special%'",
		
		// Search in customer emails with wildcard (defeats index)
		"SELECT * FROM customers WHERE email LIKE '%gmail%'",
		
		// Complex WHERE on multiple unindexed fields
		"SELECT * FROM products WHERE description LIKE '%professional%' AND name LIKE '%Pro%'",
		
		// Range query on unindexed total field
		"SELECT * FROM orders WHERE total BETWEEN 100 AND 300",
	}
	
	for i := 0; i < count; i++ {
		query := queries[i%len(queries)]
		
		_, err := executeAndRecord(db, ingester, query, i+1)
		if err != nil {
			return err
		}
		
		if i < count-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	
	return nil
}

func generateComplexJoinQueries(db *database.DB, ingester *ingest.SlowQueryIngester, cfg *config.Config) error {
	fmt.Println("   🔗 Running complex JOIN queries...")
	
	queries := []string{
		// Inefficient cross join pattern
		`SELECT c.email, o.total, p.name 
		 FROM customers c, orders o, products p, order_items oi
		 WHERE c.id = o.customer_id 
		 AND o.id = oi.order_id 
		 AND oi.product_id = p.id
		 AND c.city LIKE '%New%'`,
		
		// Multiple JOINs with text search
		`SELECT c.company, COUNT(*) as order_count, SUM(o.total) as total_spent
		 FROM customers c 
		 JOIN orders o ON c.id = o.customer_id
		 JOIN order_items oi ON o.id = oi.order_id
		 JOIN products p ON oi.product_id = p.id
		 WHERE p.description LIKE '%professional%'
		 GROUP BY c.company
		 ORDER BY total_spent DESC`,
		
		// Subquery with JOIN
		`SELECT * FROM orders o
		 WHERE o.customer_id IN (
		   SELECT c.id FROM customers c 
		   WHERE c.email LIKE '%@gmail%' 
		   AND c.company LIKE '%Tech%'
		 )
		 AND o.total > (
		   SELECT AVG(total) FROM orders
		 )`,
	}
	
	for i := 0; i < count; i++ {
		query := queries[i%len(queries)]
		
		_, err := executeAndRecord(db, ingester, query, i+1)
		if err != nil {
			return err
		}
		
		if i < count-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	
	return nil
}

func generateAggregationQueries(db *database.DB, ingester *ingest.SlowQueryIngester, cfg *config.Config) error {
	fmt.Println("   📊 Running heavy aggregation queries...")
	
	queries := []string{
		// Heavy GROUP BY with ORDER BY
		`SELECT c.city, c.country, COUNT(*) as customers, 
		        SUM(o.total) as total_sales,
		        AVG(o.total) as avg_order
		 FROM customers c
		 LEFT JOIN orders o ON c.id = o.customer_id
		 GROUP BY c.city, c.country
		 ORDER BY total_sales DESC, avg_order DESC`,
		
		// Complex aggregation with text operations
		`SELECT 
		   UPPER(p.category) as category,
		   COUNT(DISTINCT c.id) as unique_customers,
		   COUNT(oi.id) as items_sold,
		   SUM(oi.quantity * oi.price) as revenue,
		   CONCAT(MIN(p.name), ' to ', MAX(p.name)) as product_range
		 FROM products p
		 JOIN order_items oi ON p.id = oi.product_id
		 JOIN orders o ON oi.order_id = o.id
		 JOIN customers c ON o.customer_id = c.id
		 WHERE p.description LIKE '%professional%' OR p.description LIKE '%premium%'
		 GROUP BY p.category
		 HAVING revenue > 100
		 ORDER BY COUNT(DISTINCT c.id) DESC, revenue DESC`,
		
		// Window functions with aggregation
		`SELECT 
		   c.email,
		   o.total,
		   ROW_NUMBER() OVER (PARTITION BY c.city ORDER BY o.total DESC) as city_rank,
		   SUM(o.total) OVER (PARTITION BY c.country) as country_total,
		   RANK() OVER (ORDER BY o.total DESC) as global_rank
		 FROM customers c
		 JOIN orders o ON c.id = o.customer_id
		 WHERE o.status IN ('paid', 'shipped', 'delivered')
		 ORDER BY o.total DESC`,
	}
	
	for i := 0; i < count; i++ {
		query := queries[i%len(queries)]
		
		_, err := executeAndRecord(db, ingester, query, i+1)
		if err != nil {
			return err
		}
		
		if i < count-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	
	return nil
}

func pluralize(count int) string {
	if count == 1 {
		return "y"
	}
	return "ies"
}