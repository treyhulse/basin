package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Test connection string using localhost
	connStr := "host=localhost port=5433 user=postgres password=Th01017319 dbname=go_rbac_db sslmode=disable"
	fmt.Printf("Testing connection with: %s\n", connStr)

	// Try to connect
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("SUCCESS: Database connection successful!")

	// Try to query a table
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query users table: %v", err)
	}

	fmt.Printf("SUCCESS: Found %d users in the database\n", count)
}
