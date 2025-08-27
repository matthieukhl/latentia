package models

import (
	"encoding/json"
	"time"
)

type SlowQuery struct {
	ID               int64           `json:"id" db:"id"`
	Digest           string          `json:"digest" db:"digest"`
	SampleSQL        string          `json:"sample_sql" db:"sample_sql"`
	StartedAt        time.Time       `json:"started_at" db:"started_at"`
	LatencyMs        int64           `json:"latency_ms" db:"latency_ms"`
	Tables           json.RawMessage `json:"tables" db:"tables"`
	Status           string          `json:"status" db:"status"`
	LastAnalyzedAt   *time.Time      `json:"last_analyzed_at" db:"last_analyzed_at"`
	BestRewriteID    *int64          `json:"best_rewrite_id" db:"best_rewrite_id"`
}

const (
	StatusPending   = "pending"
	StatusAnalyzing = "analyzing"
	StatusCompleted = "completed"
)