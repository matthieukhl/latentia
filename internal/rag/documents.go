package rag

import (
	"context"
	"fmt"
	"encoding/json"

	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/types"
)

type DocumentStore struct {
	db       *database.DB
	embedder types.Embedder
}

type Document struct {
	ID       int64  `json:"id" db:"id"`
	Title    string `json:"title" db:"title"`
	Content  string `json:"content" db:"content"`
	Category string `json:"category" db:"category"`
	URL      string `json:"url" db:"url"`
}

type DocumentChunk struct {
	ID        int64     `json:"id" db:"id"`
	DocID     int64     `json:"doc_id" db:"doc_id"`
	ChunkID   int       `json:"chunk_id" db:"chunk_id"`
	Text      string    `json:"text" db:"text"`
	Embedding []float32 `json:"embedding" db:"embedding"`
	Metadata  string    `json:"metadata" db:"metadata"`
}

type SearchResult struct {
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	Document   string  `json:"document"`
	Category   string  `json:"category"`
	URL        string  `json:"url"`
}

func NewDocumentStore(db *database.DB, embedder types.Embedder) *DocumentStore {
	return &DocumentStore{
		db:       db,
		embedder: embedder,
	}
}

// SeedTiDBOptimizationDocs adds curated TiDB optimization documentation
func (ds *DocumentStore) SeedTiDBOptimizationDocs() error {
	docs := []Document{
		{
			Title:    "TiDB Query Performance Optimization",
			Category: "performance",
			URL:      "https://docs.pingcap.com/tidb/stable/sql-tuning-overview",
			Content: `TiDB query optimization focuses on several key areas:

1. Index Usage: Ensure queries use appropriate indexes. Use EXPLAIN to check execution plans.
   - Create composite indexes for multi-column WHERE clauses
   - Consider covering indexes to avoid table lookups
   - Use prefix indexes for string columns when appropriate

2. JOIN Optimization:
   - Place tables with smaller result sets first in JOIN order
   - Use appropriate JOIN types (INNER, LEFT, etc.)
   - Consider using EXISTS instead of IN for subqueries
   - Avoid Cartesian products by ensuring proper JOIN conditions

3. WHERE Clause Optimization:
   - Push WHERE conditions as early as possible
   - Use indexed columns in WHERE clauses
   - Avoid functions in WHERE clauses that prevent index usage
   - Use LIMIT to reduce result sets when possible`,
		},
		{
			Title:    "TiDB Index Best Practices",
			Category: "indexes",
			URL:      "https://docs.pingcap.com/tidb/stable/best-practices-for-indexing",
			Content: `TiDB indexing best practices:

1. Primary Key Design:
   - Use AUTO_INCREMENT or UUID for primary keys
   - Avoid hotspot issues with sequential inserts
   - Consider SHARD_ROW_ID_BITS for high-write tables

2. Secondary Index Strategy:
   - Create indexes on frequently queried columns
   - Use composite indexes for multi-column queries
   - Order index columns by selectivity (most selective first)
   - Monitor index usage with EXPLAIN ANALYZE

3. Index Types:
   - B-tree indexes for range and equality queries
   - Hash indexes for exact matches (in memory tables)
   - Partial indexes with WHERE conditions to reduce size
   - Expression indexes for computed columns`,
		},
		{
			Title:    "TiDB JOIN Optimization Techniques",
			Category: "joins",
			URL:      "https://docs.pingcap.com/tidb/stable/join-reorder",
			Content: `TiDB JOIN optimization techniques:

1. JOIN Reordering:
   - TiDB automatically reorders JOINs based on statistics
   - Use STRAIGHT_JOIN to force specific join order when needed
   - Ensure tables with smaller cardinality are joined first

2. JOIN Types:
   - Hash Join: Good for large datasets, one side fits in memory
   - Index Nested Loop Join: Efficient when outer table is small
   - Merge Join: Optimal when both tables are sorted on join keys

3. Optimization Tips:
   - Use EXISTS instead of IN for subqueries
   - Convert complex subqueries to JOINs when possible
   - Use appropriate indexes on join columns
   - Consider denormalization for frequently joined data`,
		},
		{
			Title:    "TiDB Aggregation and GROUP BY Optimization",
			Category: "aggregation",
			URL:      "https://docs.pingcap.com/tidb/stable/aggregation-optimization",
			Content: `TiDB aggregation optimization:

1. GROUP BY Optimization:
   - Use indexes on GROUP BY columns
   - Order GROUP BY columns to match index order
   - Use covering indexes to avoid additional lookups
   - Consider pre-aggregating data in materialized views

2. Aggregate Functions:
   - COUNT(*) is optimized and should be preferred over COUNT(column)
   - Use approximate functions like APPROX_COUNT_DISTINCT for large datasets
   - Push aggregation down to storage layer when possible
   - Use window functions for running totals and rankings

3. HAVING vs WHERE:
   - Use WHERE to filter before aggregation
   - Use HAVING only for post-aggregation filtering
   - Combine conditions efficiently to reduce data processing`,
		},
		{
			Title:    "TiDB EXPLAIN ANALYZE and Query Plans",
			Category: "analysis",
			URL:      "https://docs.pingcap.com/tidb/stable/explain-analyze",
			Content: `Understanding TiDB EXPLAIN ANALYZE:

1. Reading Execution Plans:
   - execution_info shows actual runtime statistics
   - rows shows estimated vs actual row counts
   - time shows execution time for each operator
   - memory shows memory usage

2. Key Operators:
   - TableFullScan: Full table scan (may indicate missing index)
   - IndexRangeScan: Index-based range scan (generally good)
   - HashJoin/IndexJoin: Different join algorithms
   - Sort: Explicit sorting operations
   - Projection: Column selection and transformation

3. Optimization Indicators:
   - High row count differences indicate stale statistics
   - Long execution times in specific operators show bottlenecks
   - Memory usage helps identify memory-intensive operations
   - Multiple table scans suggest missing indexes`,
		},
		{
			Title:    "TiDB Query Hints and Optimizer Control",
			Category: "hints",
			URL:      "https://docs.pingcap.com/tidb/stable/optimizer-hints",
			Content: `TiDB optimizer hints for query control:

1. Index Hints:
   - USE INDEX(table_name, index_name): Force index usage
   - IGNORE INDEX(table_name, index_name): Prevent index usage
   - FORCE INDEX(table_name, index_name): Strongly prefer index

2. Join Hints:
   - HASH_JOIN(table_names): Force hash join
   - MERGE_JOIN(table_names): Force sort merge join  
   - INL_JOIN(table_names): Force index nested loop join
   - STRAIGHT_JOIN(): Disable join reordering

3. Other Optimizer Hints:
   - MAX_EXECUTION_TIME(N): Set query timeout
   - MEMORY_QUOTA(N MB): Control memory usage
   - USE_TOJA(boolean): Control subquery optimization
   - TIDB_SMJ(table_names): Force sort merge join`,
		},
	}
	
	for _, doc := range docs {
		err := ds.addDocument(doc)
		if err != nil {
			return fmt.Errorf("failed to add document %s: %w", doc.Title, err)
		}
	}
	
	return nil
}

func (ds *DocumentStore) addDocument(doc Document) error {
	// First, check if document exists
	var docID int64
	err := ds.db.QueryRow(`
		SELECT id FROM app_documents WHERE title = ?
	`, doc.Title).Scan(&docID)
	
	if err != nil {
		// Document doesn't exist, insert it
		result, err := ds.db.Exec(`
			INSERT INTO app_documents (title, content, category, url, created_at)
			VALUES (?, ?, ?, ?, NOW())
		`, doc.Title, doc.Content, doc.Category, doc.URL)
		
		if err != nil {
			return err
		}
		
		docID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		// Document exists, update it and clear old embeddings
		_, err = ds.db.Exec(`
			UPDATE app_documents 
			SET content = ?, category = ?, url = ?
			WHERE id = ?
		`, doc.Content, doc.Category, doc.URL, docID)
		if err != nil {
			return err
		}
		
		// Delete existing embeddings for this document
		_, err = ds.db.Exec(`DELETE FROM app_embeddings WHERE doc_id = ?`, docID)
		if err != nil {
			return err
		}
	}
	
	// Chunk the content and create embeddings
	chunks := chunkText(doc.Content, 400, 50) // 400 chars with 50 char overlap
	if len(chunks) == 0 {
		return nil
	}
	
	ctx := context.Background()
	embeddings, err := ds.embedder.Embed(ctx, chunks)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}
	
	// Store chunks and embeddings
	for i, chunk := range chunks {
		if i >= len(embeddings) {
			break
		}
		
		metadata := fmt.Sprintf(`{"doc_title": "%s", "category": "%s", "chunk": %d}`, 
			doc.Title, doc.Category, i)
		
		// Convert embedding to JSON string for TiDB VECTOR type
		embeddingJSON, err := json.Marshal(embeddings[i])
		if err != nil {
			return fmt.Errorf("failed to marshal embedding %d: %w", i, err)
		}
		
		_, err = ds.db.Exec(`
			INSERT INTO app_embeddings (doc_id, chunk_id, text, embedding, metadata)
			VALUES (?, ?, ?, CAST(? AS VECTOR(1536)), ?)
		`, docID, i, chunk, string(embeddingJSON), metadata)
		
		if err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", i, err)
		}
	}
	
	return nil
}

// Search performs vector similarity search for relevant documentation
func (ds *DocumentStore) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	// Generate embedding for the query
	embeddings, err := ds.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}
	
	queryEmbedding := embeddings[0]
	
	// Convert query embedding to JSON string for TiDB VECTOR comparison
	queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}
	
	// Search for similar embeddings using TiDB vector search
	searchSQL := `
		SELECT 
			e.text,
			d.title as document,
			d.category,
			d.url,
			VEC_COSINE_DISTANCE(e.embedding, CAST(? AS VECTOR(1536))) as distance
		FROM app_embeddings e
		JOIN app_documents d ON e.doc_id = d.id
		WHERE VEC_COSINE_DISTANCE(e.embedding, CAST(? AS VECTOR(1536))) < 0.5
		ORDER BY distance ASC
		LIMIT ?`
	
	queryVector := string(queryEmbeddingJSON)
	rows, err := ds.db.Query(searchSQL, queryVector, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to execute vector search: %w", err)
	}
	defer rows.Close()
	
	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var distance float64
		
		err := rows.Scan(&result.Text, &result.Document, &result.Category, &result.URL, &distance)
		if err != nil {
			return nil, err
		}
		
		// Convert distance to similarity score (1 - distance)
		result.Score = 1.0 - distance
		results = append(results, result)
	}
	
	return results, nil
}

// chunkText splits text into overlapping chunks
func chunkText(text string, chunkSize, overlap int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}
	
	var chunks []string
	start := 0
	
	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		
		chunk := text[start:end]
		chunks = append(chunks, chunk)
		
		if end >= len(text) {
			break
		}
		
		start += chunkSize - overlap
	}
	
	return chunks
}