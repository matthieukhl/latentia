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
	QueryTime        float64         `json:"query_time" db:"query_time"` // in seconds, matches INFORMATION_SCHEMA
	DB               string          `json:"db" db:"db"`
	IndexNames       string          `json:"index_names" db:"index_names"`
	IsInternal       bool            `json:"is_internal" db:"is_internal"`
	User             string          `json:"user" db:"user"`
	Host             string          `json:"host" db:"host"`
	Tables           json.RawMessage `json:"tables" db:"tables"`
	Source           string          `json:"source" db:"source"` // 'generated' or 'information_schema'
	Status           string          `json:"status" db:"status"`
	LastAnalyzedAt   *time.Time      `json:"last_analyzed_at" db:"last_analyzed_at"`
	BestRewriteID    *int64          `json:"best_rewrite_id" db:"best_rewrite_id"`
}

// InformationSchemaSlowQuery represents the structure from INFORMATION_SCHEMA.SLOW_QUERY
type InformationSchemaSlowQuery struct {
	StartTime   string  `db:"Start_time"`
	QueryTime   float64 `db:"Query_time"`
	Digest      string  `db:"Digest"`
	Query       string  `db:"Query"`
	DB          string  `db:"DB"`
	IndexNames  string  `db:"Index_names"`
	IsInternal  bool    `db:"Is_internal"`
	User        string  `db:"User"`
	Host        string  `db:"Host"`
}

const (
	StatusPending   = "pending"
	StatusAnalyzing = "analyzing"
	StatusCompleted = "completed"
)

const (
	SourceGenerated        = "generated"
	SourceInformationSchema = "information_schema"
)