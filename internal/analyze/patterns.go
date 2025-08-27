package analyze

import (
	"regexp"
	"strings"
)

// QueryPattern represents the type and characteristics of a SQL query
type QueryPattern struct {
	Type            string   `json:"type"`
	Tables          []string `json:"tables"`
	AntiPatterns    []string `json:"anti_patterns"`
	OptimizationOps []string `json:"optimization_opportunities"`
	Complexity      string   `json:"complexity"` // simple, medium, complex
	Keywords        []string `json:"keywords"`
}

// QueryAnalyzer detects patterns and anti-patterns in SQL queries
type QueryAnalyzer struct {
	joinRegex     *regexp.Regexp
	tableRegex    *regexp.Regexp
	subqueryRegex *regexp.Regexp
}

func NewQueryAnalyzer() *QueryAnalyzer {
	return &QueryAnalyzer{
		joinRegex:     regexp.MustCompile(`(?i)\b(INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|FULL\s+JOIN|JOIN)\b`),
		tableRegex:    regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		subqueryRegex: regexp.MustCompile(`\([^)]*SELECT[^)]*\)`),
	}
}

// AnalyzeQuery examines a SQL query and identifies patterns and optimization opportunities
func (qa *QueryAnalyzer) AnalyzeQuery(sql string) QueryPattern {
	sql = strings.TrimSpace(sql)
	sqlLower := strings.ToLower(sql)
	
	pattern := QueryPattern{
		Tables:          qa.extractTables(sql),
		AntiPatterns:    []string{},
		OptimizationOps: []string{},
		Keywords:        []string{},
	}
	
	// Detect primary query type
	pattern.Type = qa.detectQueryType(sqlLower)
	
	// Detect anti-patterns
	pattern.AntiPatterns = qa.detectAntiPatterns(sqlLower)
	
	// Identify optimization opportunities
	pattern.OptimizationOps = qa.identifyOptimizations(sqlLower, pattern.AntiPatterns)
	
	// Assess complexity
	pattern.Complexity = qa.assessComplexity(sqlLower, len(pattern.Tables))
	
	// Extract relevant keywords
	pattern.Keywords = qa.extractKeywords(sqlLower)
	
	return pattern
}

// detectQueryType identifies the primary type of SQL operation
func (qa *QueryAnalyzer) detectQueryType(sql string) string {
	if strings.Contains(sql, "sleep(") {
		return "sleep-test"
	}
	
	joinCount := len(qa.joinRegex.FindAllString(sql, -1))
	if joinCount > 0 {
		if joinCount >= 3 {
			return "complex-join"
		}
		return "simple-join"
	}
	
	if strings.Contains(sql, "group by") || strings.Contains(sql, "order by") {
		return "aggregation"
	}
	
	if strings.Contains(sql, "like") && strings.Contains(sql, "%") {
		return "pattern-search"
	}
	
	if strings.Contains(sql, "select *") {
		return "full-select"
	}
	
	if strings.Contains(sql, "where") {
		return "filtered-select"
	}
	
	return "basic-select"
}

// detectAntiPatterns identifies performance anti-patterns
func (qa *QueryAnalyzer) detectAntiPatterns(sql string) []string {
	antiPatterns := []string{}
	
	// SELECT * usage
	if strings.Contains(sql, "select *") {
		antiPatterns = append(antiPatterns, "select-star")
	}
	
	// Leading wildcard LIKE patterns
	if strings.Contains(sql, "like '%") {
		antiPatterns = append(antiPatterns, "leading-wildcard-like")
	}
	
	// Missing LIMIT on potentially large result sets
	if !strings.Contains(sql, "limit") && (strings.Contains(sql, "join") || strings.Contains(sql, "order by")) {
		antiPatterns = append(antiPatterns, "missing-limit")
	}
	
	// Cartesian product risk (comma joins)
	if strings.Contains(sql, " from ") && strings.Count(sql, ",") > 0 && !strings.Contains(sql, "join") {
		antiPatterns = append(antiPatterns, "cartesian-join")
	}
	
	// Functions in WHERE clause
	if regexp.MustCompile(`(?i)WHERE[^=]*\([^)]*\)\s*[=<>]`).MatchString(sql) {
		antiPatterns = append(antiPatterns, "function-in-where")
	}
	
	// Subquery instead of JOIN
	if len(qa.subqueryRegex.FindAllString(sql, -1)) > 0 && strings.Contains(sql, "in (") {
		antiPatterns = append(antiPatterns, "subquery-instead-of-join")
	}
	
	// ORDER BY without LIMIT
	if strings.Contains(sql, "order by") && !strings.Contains(sql, "limit") {
		antiPatterns = append(antiPatterns, "order-without-limit")
	}
	
	return antiPatterns
}

// identifyOptimizations suggests optimization opportunities based on detected patterns
func (qa *QueryAnalyzer) identifyOptimizations(sql string, antiPatterns []string) []string {
	optimizations := []string{}
	
	for _, pattern := range antiPatterns {
		switch pattern {
		case "select-star":
			optimizations = append(optimizations, "specify-columns")
		case "leading-wildcard-like":
			optimizations = append(optimizations, "optimize-like-patterns")
		case "missing-limit":
			optimizations = append(optimizations, "add-limit-clause")
		case "cartesian-join":
			optimizations = append(optimizations, "explicit-join-syntax")
		case "function-in-where":
			optimizations = append(optimizations, "move-functions-to-select")
		case "subquery-instead-of-join":
			optimizations = append(optimizations, "convert-to-join")
		case "order-without-limit":
			optimizations = append(optimizations, "add-result-limiting")
		}
	}
	
	// Additional optimizations based on query structure
	if strings.Contains(sql, "group by") {
		optimizations = append(optimizations, "index-group-by-columns")
	}
	
	if strings.Contains(sql, "join") {
		optimizations = append(optimizations, "index-join-columns")
	}
	
	if strings.Contains(sql, "where") && !strings.Contains(sql, "index") {
		optimizations = append(optimizations, "index-where-columns")
	}
	
	return optimizations
}

// assessComplexity determines query complexity based on various factors
func (qa *QueryAnalyzer) assessComplexity(sql string, tableCount int) string {
	score := 0
	
	// Table count factor
	if tableCount >= 4 {
		score += 3
	} else if tableCount >= 2 {
		score += 1
	}
	
	// Join complexity
	joinCount := len(qa.joinRegex.FindAllString(sql, -1))
	score += joinCount
	
	// Subquery complexity
	subqueryCount := len(qa.subqueryRegex.FindAllString(sql, -1))
	score += subqueryCount * 2
	
	// Aggregation complexity
	if strings.Contains(sql, "group by") {
		score += 1
	}
	if strings.Contains(sql, "having") {
		score += 1
	}
	if strings.Contains(sql, "window") || strings.Contains(sql, "over(") {
		score += 2
	}
	
	// UNION complexity
	if strings.Contains(sql, "union") {
		score += 2
	}
	
	if score >= 6 {
		return "complex"
	} else if score >= 3 {
		return "medium"
	}
	return "simple"
}

// extractTables extracts table names from FROM and JOIN clauses
func (qa *QueryAnalyzer) extractTables(sql string) []string {
	tables := []string{}
	tableSet := make(map[string]bool)
	
	matches := qa.tableRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			table := strings.TrimSpace(match[1])
			// Remove alias if present
			if spaceIdx := strings.Index(table, " "); spaceIdx > 0 {
				table = table[:spaceIdx]
			}
			if table != "" && !tableSet[table] {
				tables = append(tables, table)
				tableSet[table] = true
			}
		}
	}
	
	return tables
}

// extractKeywords identifies relevant SQL keywords for optimization context
func (qa *QueryAnalyzer) extractKeywords(sql string) []string {
	keywords := []string{}
	
	keywordMap := map[string]string{
		"join":     "joins",
		"index":    "indexes", 
		"group by": "aggregation",
		"order by": "sorting",
		"limit":    "limiting",
		"where":    "filtering",
		"having":   "aggregation",
		"distinct": "deduplication",
		"union":    "set-operations",
		"exists":   "subqueries",
		"in (":     "subqueries",
		"like":     "pattern-matching",
	}
	
	for keyword, category := range keywordMap {
		if strings.Contains(sql, keyword) {
			keywords = append(keywords, category)
		}
	}
	
	return removeDuplicates(keywords)
}

// Helper function to remove duplicates from string slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}