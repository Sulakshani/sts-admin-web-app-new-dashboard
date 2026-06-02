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

	fmt.Println("Checking bookings table structure:\n")

	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'bookings'
		ORDER BY ordinal_position
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error querying: %v", err)
	}
	defer rows.Close()

	fmt.Println("Column Name          | Data Type      | Nullable | Default")
	fmt.Println("---------------------|----------------|----------|------------------")
	for rows.Next() {
		var colName, dataType, nullable string
		var colDefault sql.NullString
		err := rows.Scan(&colName, &dataType, &nullable, &colDefault)
		if err != nil {
			log.Fatalf("Error scanning: %v", err)
		}
		def := "NULL"
		if colDefault.Valid {
			def = colDefault.String
		}
		fmt.Printf("%-20s | %-14s | %-8s | %s\n", colName, dataType, nullable, def)
	}
}
