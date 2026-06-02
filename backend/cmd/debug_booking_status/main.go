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
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	fmt.Println("Connected to database successfully!")
	fmt.Println("\nChecking payment_status and booking_status values:\n")

	// Query to check the actual status values
	query := `
		SELECT 
			bb.id::text as booking_id,
			b.booking_reference,
			b.payment_status,
			bb.status as booking_status
		FROM bus_bookings bb
		INNER JOIN bookings b ON bb.booking_id = b.id
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error querying: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var bookingID, bookingRef string
		var paymentStatus, bookingStatus sql.NullString

		err := rows.Scan(&bookingID, &bookingRef, &paymentStatus, &bookingStatus)
		if err != nil {
			log.Fatalf("Error scanning: %v", err)
		}

		count++
		fmt.Printf("Booking %d:\n", count)
		fmt.Printf("  ID: %s\n", bookingID)
		fmt.Printf("  Reference: %s\n", bookingRef)
		fmt.Printf("  Payment Status: '%s' (Valid: %v)\n", paymentStatus.String, paymentStatus.Valid)
		fmt.Printf("  Booking Status: '%s' (Valid: %v)\n", bookingStatus.String, bookingStatus.Valid)
		fmt.Println()
	}

	if count == 0 {
		fmt.Println("No bookings found!")
	} else {
		fmt.Printf("Total checked: %d bookings\n", count)
	}
}
