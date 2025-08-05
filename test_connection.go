package main

import (
	"fmt"
	"log"
	"os"

	"go-rbac-api/internal/config"
)

func main() {
	fmt.Println("Environment Variables:")
	fmt.Printf("DB_HOST: %s\n", os.Getenv("DB_HOST"))
	fmt.Printf("DB_PORT: %s\n", os.Getenv("DB_PORT"))
	fmt.Printf("DB_USER: %s\n", os.Getenv("DB_USER"))
	fmt.Printf("DB_PASSWORD: %s\n", os.Getenv("DB_PASSWORD"))
	fmt.Printf("DB_NAME: %s\n", os.Getenv("DB_NAME"))
	fmt.Printf("DB_SSLMODE: %s\n", os.Getenv("DB_SSLMODE"))
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	fmt.Println("Database Configuration:")
	fmt.Printf("Host: %s\n", cfg.DBHost)
	fmt.Printf("Port: %d\n", cfg.DBPort)
	fmt.Printf("User: %s\n", cfg.DBUser)
	fmt.Printf("Password: %s\n", cfg.DBPassword)
	fmt.Printf("Database: %s\n", cfg.DBName)
	fmt.Printf("SSL Mode: %s\n", cfg.DBSSLMode)
	fmt.Println()
	fmt.Printf("Connection String: %s\n", cfg.GetDBConnString())
}
