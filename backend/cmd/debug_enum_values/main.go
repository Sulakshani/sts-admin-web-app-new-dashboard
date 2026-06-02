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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Query to get enum values for verification_status
	query := `
		SELECT 
			t.typname as enum_name,
			e.enumlabel as enum_value
		FROM pg_type t 
		JOIN pg_enum e ON t.oid = e.enumtypid  
		WHERE t.typname = 'verification_status'
		ORDER BY e.enumsortorder;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query enum values: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nverification_status ENUM values:")
	fmt.Println("====================================")
	for rows.Next() {
		var enumName, enumValue string
		if err := rows.Scan(&enumName, &enumValue); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("  - %s\n", enumValue)
	}
	fmt.Println("====================================")
}
