// internal/database/db.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB initializes the database connection pool.
func InitDB() error {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	fmt.Printf("Host : %s\n", dbHost)

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	// Open database connection pool
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings (optional but recommended)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection with a ping (with retry logic)
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
 			log.Println("Successfully connected to thedatabase!")
			// Create table if it doesn't exist after successful connection
			return createMistakesTable()
		}
		log.Printf("Failed to ping database (attempt %d/%d): %v. Retrying in 5 seconds...", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)

}

// returns the database connection pool instance.
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("Database connection pool is not initialized. Call InitDB first.")
	}
	return db
}

// Creates the 'mistakes' table if it doesn't exist.
func createMistakesTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS mistakes (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		category VARCHAR(100)
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create mistakes table: %w", err)
	}
	log.Println("Mistakes table checked/created successfully.")
	return nil
}

// CloseDB closes the database connection pool.
func CloseDB() {
	if db != nil {
		db.Close()
		log.Println("Database connection pool closed.")
	}
}