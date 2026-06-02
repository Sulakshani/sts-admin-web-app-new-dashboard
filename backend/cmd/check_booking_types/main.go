package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	fmt.Println("Checking distinct booking_type values:\n")

	rows, err := db.Query("SELECT DISTINCT booking_type FROM bookings LIMIT 10")
	if err != nil {
		log.Fatalf("Error querying: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bookingType string
		err := rows.Scan(&bookingType)
		if err != nil {
			log.Fatalf("Error scanning: %v", err)
		}
		fmt.Printf("- %s\n", bookingType)
	}
}
