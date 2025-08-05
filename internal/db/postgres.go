package db

import (
	"database/sql"
	"fmt"
	"log"

	"go-rbac-api/internal/config"
	sqlc "go-rbac-api/internal/db/sqlc"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
	*sqlc.Queries
}

func NewDB(cfg *config.Config) (*DB, error) {
	connStr := cfg.GetDBConnString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	queries := sqlc.New(db)

	return &DB{
		DB:      db,
		Queries: queries,
	}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
