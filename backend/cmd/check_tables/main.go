package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if trip_seats table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'trip_seats'
		)
	`).Scan(&exists)
	
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("trip_seats table exists: %v\n", exists)

	if exists {
		// Get trip_seats columns
		fmt.Println("\ntrip_seats columns:")
		rows, _ := db.Query(`
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = 'trip_seats'
			ORDER BY ordinal_position
		`)
		defer rows.Close()
		for rows.Next() {
			var col, dtype, nullable string
			rows.Scan(&col, &dtype, &nullable)
			fmt.Printf("  - %s (%s) %s\n", col, dtype, nullable)
		}
	}

	// Check bus_booking_seats
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'bus_booking_seats'
		)
	`).Scan(&exists)

	fmt.Printf("\nbus_booking_seats table exists: %v\n", exists)

	if exists {
		fmt.Println("\nbus_booking_seats columns:")
		rows, _ := db.Query(`
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = 'bus_booking_seats'
			ORDER BY ordinal_position
		`)
		defer rows.Close()
		for rows.Next() {
			var col, dtype, nullable string
			rows.Scan(&col, &dtype, &nullable)
			fmt.Printf("  - %s (%s) %s\n", col, dtype, nullable)
		}
	}
}
