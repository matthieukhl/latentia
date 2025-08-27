package analyze

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/matthieukhl/latentia/internal/database"
	"github.com/matthieukhl/latentia/internal/rag"
	"github.com/matthieukhl/latentia/internal/types"
)

// OptimizationEngine provides the main SQL optimization pipeline
type OptimizationEngine struct {
	db            *database.DB
	analyzer      *QueryAnalyzer
	promptBuilder *PromptBuilder
	generator     types.Generator
}

// OptimizationResult contains the complete optimization analysis
type OptimizationResult struct {
	ID               int64         `json:"id" db:"id"`
	OriginalSQL      string        `json:"original_sql" db:"original_sql"`
	OptimizedSQL     string        `json:"optimized_sql" db:"optimized_sql"`
	Pattern          QueryPattern  `json:"pattern" db:"pattern"`
	Rationale        string        `json:"rationale" db:"rationale"`
	ExpectedImprovement string     `json:"expected_improvement" db:"expected_improvement"`
	Caveats          string        `json:"caveats" db:"caveats"`
	ConfidenceScore  float64       `json:"confidence_score" db:"confidence_score"`
	Status           string        `json:"status" db:"status"` // pending, accepted, rejected
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	ReviewedAt       *time.Time    `json:"reviewed_at" db:"reviewed_at"`
}

// LLMResponse represents the structured response from the LLM
type LLMResponse struct {
	ProposedSQL         string `json:"proposed_sql"`
	Rationale           string `json:"rationale"`
	ExpectedPlanChange  string `json:"expected_plan_change"`
	Caveats             string `json:"caveats"`
}

func NewOptimizationEngine(db *database.DB, docStore *rag.DocumentStore, generator types.Generator) *OptimizationEngine {
	return &OptimizationEngine{
		db:            db,
		analyzer:      NewQueryAnalyzer(),
		promptBuilder: NewPromptBuilder(docStore),
		generator:     generator,
	}
}

// OptimizeQuery processes a slow query through the complete optimization pipeline
func (oe *OptimizationEngine) OptimizeQuery(ctx context.Context, slowQueryID int64, sql string) (*OptimizationResult, error) {
	// Step 1: Analyze query patterns
	pattern := oe.analyzer.AnalyzeQuery(sql)
	
	// Step 2: Build context-aware prompt
	prompt, err := oe.promptBuilder.BuildOptimizationPrompt(sql, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to build optimization prompt: %w", err)
	}
	
	// Step 3: Generate optimization with LLM
	llmResponse, err := oe.generator.Complete(ctx, prompt, map[string]any{
		"max_tokens": 2000,
		"temperature": 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimization: %w", err)
	}
	
	// Step 4: Parse LLM response
	parsedResponse, err := oe.parseLLMResponse(llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	
	// Step 5: Calculate confidence score
	confidenceScore := oe.calculateConfidenceScore(pattern, parsedResponse)
	
	// Step 6: Store optimization result
	result := &OptimizationResult{
		OriginalSQL:         sql,
		OptimizedSQL:        parsedResponse.ProposedSQL,
		Pattern:             pattern,
		Rationale:           parsedResponse.Rationale,
		ExpectedImprovement: parsedResponse.ExpectedPlanChange,
		Caveats:             parsedResponse.Caveats,
		ConfidenceScore:     confidenceScore,
		Status:              "pending",
		CreatedAt:           time.Now(),
	}
	
	// Store in database
	err = oe.storeOptimizationResult(ctx, slowQueryID, result)
	if err != nil {
		return nil, fmt.Errorf("failed to store optimization result: %w", err)
	}
	
	return result, nil
}

// parseLLMResponse extracts structured information from LLM response
func (oe *OptimizationEngine) parseLLMResponse(response string) (*LLMResponse, error) {
	parsed := &LLMResponse{}
	
	// Extract SQL using regex - use (?s) for DOTALL mode
	sqlRegex := regexp.MustCompile(`(?s)PROPOSED_SQL:\s*` + "```sql" + `\s*(.*?)\s*` + "```")
	if matches := sqlRegex.FindStringSubmatch(response); len(matches) > 1 {
		parsed.ProposedSQL = strings.TrimSpace(matches[1])
	}
	
	// Extract rationale
	rationaleRegex := regexp.MustCompile(`(?s)RATIONALE:\s*(.*?)(?:\n\n|EXPECTED_PLAN_CHANGE:)`)
	if matches := rationaleRegex.FindStringSubmatch(response); len(matches) > 1 {
		parsed.Rationale = strings.TrimSpace(matches[1])
		// Clean up bullet points
		parsed.Rationale = strings.ReplaceAll(parsed.Rationale, "• ", "")
		parsed.Rationale = strings.ReplaceAll(parsed.Rationale, "\n", " ")
	}
	
	// Extract expected plan change
	planChangeRegex := regexp.MustCompile(`(?s)EXPECTED_PLAN_CHANGE:\s*(.*?)(?:\n\n|CAVEATS:)`)
	if matches := planChangeRegex.FindStringSubmatch(response); len(matches) > 1 {
		parsed.ExpectedPlanChange = strings.TrimSpace(matches[1])
		parsed.ExpectedPlanChange = strings.ReplaceAll(parsed.ExpectedPlanChange, "• ", "")
		parsed.ExpectedPlanChange = strings.ReplaceAll(parsed.ExpectedPlanChange, "\n", " ")
	}
	
	// Extract caveats
	caveatsRegex := regexp.MustCompile(`(?s)CAVEATS:\s*(.*?)(?:\n\n|$)`)
	if matches := caveatsRegex.FindStringSubmatch(response); len(matches) > 1 {
		parsed.Caveats = strings.TrimSpace(matches[1])
		parsed.Caveats = strings.ReplaceAll(parsed.Caveats, "• ", "")
		parsed.Caveats = strings.ReplaceAll(parsed.Caveats, "\n", " ")
	}
	
	// Validate that we extracted the essential parts
	if parsed.ProposedSQL == "" {
		return nil, fmt.Errorf("failed to extract optimized SQL from LLM response")
	}
	
	return parsed, nil
}

// calculateConfidenceScore assigns a confidence score based on various factors
func (oe *OptimizationEngine) calculateConfidenceScore(pattern QueryPattern, response *LLMResponse) float64 {
	score := 0.5 // Base score
	
	// Pattern-based confidence adjustments
	switch pattern.Complexity {
	case "simple":
		score += 0.3
	case "medium":
		score += 0.1
	case "complex":
		score -= 0.1
	}
	
	// Anti-pattern detection boosts confidence
	if len(pattern.AntiPatterns) > 0 {
		score += 0.2
	}
	
	// Clear optimization opportunities boost confidence
	if len(pattern.OptimizationOps) > 2 {
		score += 0.15
	}
	
	// Response quality indicators
	if len(response.Rationale) > 50 {
		score += 0.1
	}
	
	if len(response.ExpectedPlanChange) > 50 {
		score += 0.1
	}
	
	if strings.Contains(strings.ToLower(response.ProposedSQL), "index") {
		score += 0.05
	}
	
	// Clamp score between 0.1 and 1.0
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.1 {
		score = 0.1
	}
	
	return score
}

// storeOptimizationResult saves the optimization result to the database
func (oe *OptimizationEngine) storeOptimizationResult(ctx context.Context, slowQueryID int64, result *OptimizationResult) error {
	// Serialize pattern as JSON
	patternJSON, err := json.Marshal(result.Pattern)
	if err != nil {
		return fmt.Errorf("failed to serialize pattern: %w", err)
	}
	
	query := `
		INSERT INTO app_rewrites (
			slow_query_id, original_sql, optimized_sql, pattern_analysis,
			rationale, expected_improvement, caveats, confidence_score,
			status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	res, err := oe.db.Exec(query,
		slowQueryID,
		result.OriginalSQL,
		result.OptimizedSQL,
		string(patternJSON),
		result.Rationale,
		result.ExpectedImprovement,
		result.Caveats,
		result.ConfidenceScore,
		result.Status,
		result.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert optimization result: %w", err)
	}
	
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}
	
	result.ID = id
	return nil
}

// GetOptimizationByID retrieves an optimization result by ID
func (oe *OptimizationEngine) GetOptimizationByID(ctx context.Context, id int64) (*OptimizationResult, error) {
	query := `
		SELECT id, slow_query_id, original_sql, optimized_sql, pattern_analysis,
			   rationale, expected_improvement, caveats, confidence_score,
			   status, created_at, reviewed_at
		FROM app_rewrites
		WHERE id = ?
	`
	
	row := oe.db.QueryRow(query, id)
	
	var result OptimizationResult
	var patternJSON string
	var slowQueryID int64
	var reviewedAt sql.NullTime
	
	err := row.Scan(
		&result.ID,
		&slowQueryID,
		&result.OriginalSQL,
		&result.OptimizedSQL,
		&patternJSON,
		&result.Rationale,
		&result.ExpectedImprovement,
		&result.Caveats,
		&result.ConfidenceScore,
		&result.Status,
		&result.CreatedAt,
		&reviewedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("optimization result not found")
		}
		return nil, fmt.Errorf("failed to scan optimization result: %w", err)
	}
	
	// Parse pattern JSON
	err = json.Unmarshal([]byte(patternJSON), &result.Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pattern JSON: %w", err)
	}
	
	if reviewedAt.Valid {
		result.ReviewedAt = &reviewedAt.Time
	}
	
	return &result, nil
}

// ListPendingOptimizations retrieves all pending optimization results
func (oe *OptimizationEngine) ListPendingOptimizations(ctx context.Context, limit int) ([]OptimizationResult, error) {
	query := `
		SELECT id, slow_query_id, original_sql, optimized_sql, pattern_analysis,
			   rationale, expected_improvement, caveats, confidence_score,
			   status, created_at, reviewed_at
		FROM app_rewrites
		WHERE status = 'pending'
		ORDER BY confidence_score DESC, created_at DESC
		LIMIT ?
	`
	
	rows, err := oe.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending optimizations: %w", err)
	}
	defer rows.Close()
	
	var results []OptimizationResult
	for rows.Next() {
		var result OptimizationResult
		var patternJSON string
		var slowQueryID int64
		var reviewedAt sql.NullTime
		
		err := rows.Scan(
			&result.ID,
			&slowQueryID,
			&result.OriginalSQL,
			&result.OptimizedSQL,
			&patternJSON,
			&result.Rationale,
			&result.ExpectedImprovement,
			&result.Caveats,
			&result.ConfidenceScore,
			&result.Status,
			&result.CreatedAt,
			&reviewedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan optimization result: %w", err)
		}
		
		// Parse pattern JSON
		err = json.Unmarshal([]byte(patternJSON), &result.Pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pattern JSON: %w", err)
		}
		
		if reviewedAt.Valid {
			result.ReviewedAt = &reviewedAt.Time
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// AcceptOptimization marks an optimization as accepted
func (oe *OptimizationEngine) AcceptOptimization(ctx context.Context, id int64) error {
	query := `
		UPDATE app_rewrites 
		SET status = 'accepted', reviewed_at = NOW() 
		WHERE id = ? AND status = 'pending'
	`
	
	result, err := oe.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to accept optimization: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("optimization not found or already reviewed")
	}
	
	return nil
}

// RejectOptimization marks an optimization as rejected
func (oe *OptimizationEngine) RejectOptimization(ctx context.Context, id int64) error {
	query := `
		UPDATE app_rewrites 
		SET status = 'rejected', reviewed_at = NOW() 
		WHERE id = ? AND status = 'pending'
	`
	
	result, err := oe.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to reject optimization: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("optimization not found or already reviewed")
	}
	
	return nil
}