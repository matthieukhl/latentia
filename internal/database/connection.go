package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/matthieukhl/latentia/internal/config"
)

type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection using the provided config
func NewConnection(cfg *config.DBConfig) (*DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return &DB{db}, nil
}

// HealthCheck performs a simple health check on the database
func (db *DB) HealthCheck() error {
	return db.Ping()
}