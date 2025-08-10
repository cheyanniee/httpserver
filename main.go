package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"httpserver/handlers"
	"httpserver/models"

	_ "github.com/lib/pq"
)

// Config holds DB configuration
var config = models.Config{
	DBUser:     "postgres",
	DBPassword: "12345",
	DBName:     "postgres",
	DBHost:     "localhost",
	DBPort:     5432,
	ServerPort: ":3333",
}

func main() {
	// Setup DB
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName,
	)

	var err error
	models.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	if err = models.DB.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	log.Println("Database connection established")

	// Create accounts table if not exists
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS accounts (
        account_id VARCHAR(255) PRIMARY KEY,
        balance NUMERIC NOT NULL
    );`

	_, err = models.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create accounts table: %v", err)
	}

	// Handler functions
	http.HandleFunc("/accounts", handlers.CreateAccountHandler)
	http.HandleFunc("/accounts/", handlers.GetAccountHandler)
	http.HandleFunc("/transactions", handlers.TransactionHandler)

	log.Printf("Server running on %s\n", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, nil))
}
