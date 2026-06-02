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

	rows, err := db.Query("SELECT column_name, data_type, udt_name FROM information_schema.columns WHERE table_name = 'bus_staff'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Columns in bus_staff table:")
	for rows.Next() {
		var name, dtype, udt string
		if err := rows.Scan(&name, &dtype, &udt); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: data_type=%s, udt_name=%s\n", name, dtype, udt)
	}

	// Check enum values
	rows2, err := db.Query(`
		SELECT e.enumlabel 
		FROM pg_type t 
		JOIN pg_enum e ON t.oid = e.enumtypid  
		WHERE t.typname = 'staff_verification_status'
		ORDER BY e.enumsortorder
	`)
	if err != nil {
		fmt.Println("No enum type 'staff_verification_status' found")
	} else {
		defer rows2.Close()
		fmt.Println("\nEnum values for staff_verification_status:")
		for rows2.Next() {
			var val string
			if err := rows2.Scan(&val); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  - '%s'\n", val)
		}
	}
}
