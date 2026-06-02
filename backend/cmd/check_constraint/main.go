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

	// Get the check constraint definition
	query := `
		SELECT conname, pg_get_constraintdef(oid) 
		FROM pg_constraint 
		WHERE conrelid = 'bus_staff_employment'::regclass 
		AND contype = 'c'
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Check constraints on bus_staff_employment:")
	for rows.Next() {
		var name, def string
		rows.Scan(&name, &def)
		fmt.Printf("\nConstraint: %s\nDefinition: %s\n", name, def)
	}
}
