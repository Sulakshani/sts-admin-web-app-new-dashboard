package main

import (
	"database/sql"
	"fmt"
	"log"
	"sts-backend/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if the bookings table exists
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'bookings'
		)
	`).Scan(&tableExists)

	if err != nil {
		fmt.Println("Error checking table existence:", err)
		return
	}

	if !tableExists {
		fmt.Println("Table bookings does not exist!")
		return
	}

	fmt.Println("Table bookings exists")

	// Get all columns from bookings table
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'bookings'
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Println("Error querying columns:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nColumns in bookings table:")
	for rows.Next() {
		var colName, dataType, nullable string
		if err := rows.Scan(&colName, &dataType, &nullable); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  - %s (%s, nullable: %s)\n", colName, dataType, nullable)
	}
}
