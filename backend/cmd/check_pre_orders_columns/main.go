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

	// Check if the lounge_booking_pre_orders table exists
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'lounge_booking_pre_orders'
		)
	`).Scan(&tableExists)

	if err != nil {
		fmt.Println("Error checking table existence:", err)
		return
	}

	if !tableExists {
		fmt.Println("Table lounge_booking_pre_orders does not exist!")
		return
	}

	fmt.Println("Table lounge_booking_pre_orders exists")

	// Get all columns from lounge_booking_pre_orders table
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'lounge_booking_pre_orders'
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Println("Error querying columns:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nColumns in lounge_booking_pre_orders table:")
	for rows.Next() {
		var colName, dataType, nullable string
		var colDefault sql.NullString
		if err := rows.Scan(&colName, &dataType, &nullable, &colDefault); err != nil {
			log.Fatal(err)
		}
		defaultVal := "NULL"
		if colDefault.Valid {
			defaultVal = colDefault.String
		}
		fmt.Printf("  - %s (%s, nullable: %s, default: %s)\n", colName, dataType, nullable, defaultVal)
	}
}
