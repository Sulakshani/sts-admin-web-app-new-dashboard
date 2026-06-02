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

	// Query to get enum labels for verification_status type
	// Note: The error message says "enum verification_status", so the type name is likely verification_status
	query := `
		SELECT e.enumlabel
		FROM pg_enum e
		JOIN pg_type t ON e.enumtypid = t.oid
		WHERE t.typname = 'verification_status';
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Valid values for verification_status enum:")
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			log.Fatal(err)
		}
		fmt.Println(label)
	}
}
