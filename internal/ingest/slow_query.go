package ingest

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/models"
)

type SlowQueryIngester struct {
	db *database.DB
}

func NewSlowQueryIngester(db *database.DB) *SlowQueryIngester {
	return &SlowQueryIngester{db: db}
}

// RecordGeneratedSlowQuery records a slow query that we generated ourselves
func (s *SlowQueryIngester) RecordGeneratedSlowQuery(query string, startTime time.Time, queryTime float64, database string, user string) error {
	digest := generateSQLDigest(query)
	
	_, err := s.db.Exec(`
		INSERT INTO app_slow_queries (
			digest, sample_sql, started_at, query_time, db, 
			index_names, is_internal, user, host, tables, source
		) VALUES (?, ?, ?, ?, ?, '', FALSE, ?, '', '[]', ?)
	`, digest, query, startTime, queryTime, database, user, models.SourceGenerated)
	
	return err
}

// IngestFromInformationSchema reads slow queries from INFORMATION_SCHEMA.SLOW_QUERY
func (s *SlowQueryIngester) IngestFromInformationSchema(minQueryTime float64, limit int) error {
	// First, check if we can access INFORMATION_SCHEMA.SLOW_QUERY
	canAccess, err := s.canAccessInformationSchema()
	if err != nil {
		return fmt.Errorf("failed to check INFORMATION_SCHEMA access: %w", err)
	}
	
	if !canAccess {
		return fmt.Errorf("INFORMATION_SCHEMA.SLOW_QUERY is not accessible (common in managed TiDB)")
	}
	
	// Fetch slow queries from INFORMATION_SCHEMA
	queries, err := s.fetchFromInformationSchema(minQueryTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch from INFORMATION_SCHEMA: %w", err)
	}
	
	// Insert new queries (avoid duplicates based on digest + start_time)
	inserted := 0
	for _, query := range queries {
		exists, err := s.slowQueryExists(query.Digest, query.StartTime)
		if err != nil {
			return fmt.Errorf("failed to check if query exists: %w", err)
		}
		
		if !exists {
			err = s.insertInformationSchemaQuery(query)
			if err != nil {
				return fmt.Errorf("failed to insert query: %w", err)
			}
			inserted++
		}
	}
	
	return nil
}

// canAccessInformationSchema checks if we can read from INFORMATION_SCHEMA.SLOW_QUERY
func (s *SlowQueryIngester) canAccessInformationSchema() (bool, error) {
	_, err := s.db.Exec("SELECT 1 FROM INFORMATION_SCHEMA.SLOW_QUERY LIMIT 1")
	if err != nil {
		if strings.Contains(err.Error(), "command denied") || 
		   strings.Contains(err.Error(), "Unknown table") ||
		   strings.Contains(err.Error(), "doesn't exist") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// fetchFromInformationSchema retrieves slow queries from INFORMATION_SCHEMA
func (s *SlowQueryIngester) fetchFromInformationSchema(minQueryTime float64, limit int) ([]models.InformationSchemaSlowQuery, error) {
	query := `
		SELECT 
			Start_time,
			Query_time,
			Digest,
			Query,
			COALESCE(DB, '') as DB,
			COALESCE(Index_names, '') as Index_names,
			Is_internal,
			COALESCE(User, '') as User,
			COALESCE(Host, '') as Host
		FROM INFORMATION_SCHEMA.SLOW_QUERY 
		WHERE Query_time >= ? 
		ORDER BY Start_time DESC 
		LIMIT ?`
	
	rows, err := s.db.Query(query, minQueryTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var queries []models.InformationSchemaSlowQuery
	for rows.Next() {
		var q models.InformationSchemaSlowQuery
		
		err := rows.Scan(
			&q.StartTime,
			&q.QueryTime,
			&q.Digest,
			&q.Query,
			&q.DB,
			&q.IndexNames,
			&q.IsInternal,
			&q.User,
			&q.Host,
		)
		if err != nil {
			return nil, err
		}
		
		queries = append(queries, q)
	}
	
	return queries, nil
}

// slowQueryExists checks if a slow query with the same digest and start time already exists
func (s *SlowQueryIngester) slowQueryExists(digest, startTime string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM app_slow_queries WHERE digest = ? AND started_at = ?",
		digest, startTime,
	).Scan(&count)
	
	return count > 0, err
}

// insertInformationSchemaQuery inserts a query from INFORMATION_SCHEMA into our app table
func (s *SlowQueryIngester) insertInformationSchemaQuery(q models.InformationSchemaSlowQuery) error {
	// Parse start time
	startTime, err := time.Parse("2006-01-02 15:04:05", q.StartTime)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}
	
	// Extract table names from query (simplified)
	tables := extractTableNames(q.Query)
	tablesJSON := fmt.Sprintf(`["%s"]`, strings.Join(tables, `","`))
	
	_, err = s.db.Exec(`
		INSERT INTO app_slow_queries (
			digest, sample_sql, started_at, query_time, db, 
			index_names, is_internal, user, host, tables, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, q.Digest, q.Query, startTime, q.QueryTime, q.DB, q.IndexNames, 
		q.IsInternal, q.User, q.Host, tablesJSON, models.SourceInformationSchema)
	
	return err
}

// GetSlowQueries retrieves slow queries from our app table for processing
func (s *SlowQueryIngester) GetSlowQueries(status string, limit int) ([]models.SlowQuery, error) {
	query := `
		SELECT 
			id, digest, sample_sql, started_at, query_time, 
			COALESCE(db, '') as db,
			COALESCE(index_names, '') as index_names,
			is_internal, 
			COALESCE(user, '') as user, 
			COALESCE(host, '') as host,
			COALESCE(tables, '[]') as tables,
			source, status,
			last_analyzed_at, best_rewrite_id
		FROM app_slow_queries 
		WHERE status = ? 
		ORDER BY query_time DESC, started_at DESC 
		LIMIT ?`
	
	rows, err := s.db.Query(query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var queries []models.SlowQuery
	for rows.Next() {
		var q models.SlowQuery
		
		err := rows.Scan(
			&q.ID, &q.Digest, &q.SampleSQL, &q.StartedAt, &q.QueryTime,
			&q.DB, &q.IndexNames, &q.IsInternal, &q.User, &q.Host,
			&q.Tables, &q.Source, &q.Status,
			&q.LastAnalyzedAt, &q.BestRewriteID,
		)
		if err != nil {
			return nil, err
		}
		
		queries = append(queries, q)
	}
	
	return queries, nil
}

// generateSQLDigest creates a simple digest/fingerprint for a SQL query
func generateSQLDigest(query string) string {
	// Normalize the query by removing extra whitespace and converting to lowercase
	normalized := strings.ToLower(strings.TrimSpace(query))
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}
	
	// Create MD5 hash
	hash := md5.Sum([]byte(normalized))
	return fmt.Sprintf("%x", hash)
}

// extractTableNames extracts table names from a SQL query (simplified implementation)
func extractTableNames(query string) []string {
	query = strings.ToLower(query)
	tables := []string{}
	
	// Simple regex-like extraction for common patterns
	if strings.Contains(query, " from ") {
		words := strings.Fields(query)
		for i, word := range words {
			if word == "from" && i+1 < len(words) {
				table := strings.TrimRight(words[i+1], ",")
				table = strings.TrimSpace(table)
				if table != "" && !contains(tables, table) {
					tables = append(tables, table)
				}
			}
		}
	}
	
	// Add JOIN tables
	if strings.Contains(query, " join ") {
		words := strings.Fields(query)
		for i, word := range words {
			if word == "join" && i+1 < len(words) {
				table := strings.TrimRight(words[i+1], ",")
				table = strings.TrimSpace(table)
				if table != "" && !contains(tables, table) {
					tables = append(tables, table)
				}
			}
		}
	}
	
	return tables
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}