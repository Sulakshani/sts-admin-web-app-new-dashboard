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

	// Check lounge_status enum values
	rows, err := db.Query(`
		SELECT e.enumlabel 
		FROM pg_type t 
		JOIN pg_enum e ON t.oid = e.enumtypid  
		WHERE t.typname = 'lounge_status'
		ORDER BY e.enumsortorder
	`)
	if err != nil {
		fmt.Println("Error querying lounge_status enum:", err)
		return
	}
	defer rows.Close()

	fmt.Println("Enum values for lounge_status:")
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  - '%s'\n", val)
	}
}
