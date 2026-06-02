package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get the latest booking
	var bookingID string
	err = db.QueryRow("SELECT id::text FROM bus_bookings ORDER BY created_at DESC LIMIT 1").Scan(&bookingID)
	if err != nil {
		log.Fatalf("Failed to get latest booking: %v", err)
	}

	fmt.Printf("Latest booking ID: %s\n\n", bookingID)

	// Check bus_booking_seats
	fmt.Println("Bus Booking Seats:")
	rows, err := db.Query(`
		SELECT 
			id::text,
			bus_booking_id::text,
			trip_seat_id::text,
			passenger_name,
			is_primary_passenger
		FROM bus_booking_seats
		WHERE bus_booking_id = $1
	`, bookingID)
	if err != nil {
		log.Fatalf("Failed to query bus_booking_seats: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, busBookingID, tripSeatID sql.NullString
		var passengerName string
		var isPrimary bool
		if err := rows.Scan(&id, &busBookingID, &tripSeatID, &passengerName, &isPrimary); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		count++
		fmt.Printf("  %d. ID=%s, TripSeatID=%s, Passenger=%s, Primary=%v\n",
			count, id.String, tripSeatID.String, passengerName, isPrimary)
	}

	if count == 0 {
		fmt.Println("  (No seat records found)")
	}

	// Check trip_seats
	fmt.Println("\nTrip Seats:")
	rows2, err := db.Query(`
		SELECT 
			ts.id::text,
			ts.seat_number,
			ts.is_booked
		FROM trip_seats ts
		INNER JOIN bus_booking_seats bbs ON bbs.trip_seat_id = ts.id
		WHERE bbs.bus_booking_id = $1
	`, bookingID)
	if err != nil {
		log.Printf("Failed to query trip_seats: %v", err)
	} else {
		defer rows2.Close()

		count2 := 0
		for rows2.Next() {
			var id sql.NullString
			var seatNumber string
			var isBooked bool
			if err := rows2.Scan(&id, &seatNumber, &isBooked); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			count2++
			fmt.Printf("  %d. ID=%s, SeatNumber=%s, Booked=%v\n",
				count2, id.String, seatNumber, isBooked)
		}

		if count2 == 0 {
			fmt.Println("  (No trip_seat records found)")
		}
	}

	// Test the aggregation query
	fmt.Println("\nAggregation Query Result:")
	var seatNumbers sql.NullString
	err = db.QueryRow(`
		SELECT COALESCE(
			(
				SELECT STRING_AGG(ts2.seat_number, ', ' ORDER BY ts2.seat_number)
				FROM bus_booking_seats bbs2
				LEFT JOIN trip_seats ts2 ON bbs2.trip_seat_id = ts2.id
				WHERE bbs2.bus_booking_id = $1 AND ts2.seat_number IS NOT NULL
			),
			''
		) as seat_numbers
	`, bookingID).Scan(&seatNumbers)
	if err != nil {
		log.Printf("Failed to run aggregation query: %v", err)
	} else {
		fmt.Printf("  Seat Numbers: '%s'\n", seatNumbers.String)
	}
}
