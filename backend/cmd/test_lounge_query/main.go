package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"sts-backend/internal/config"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	godotenv.Load()

	cfg := config.LoadConfig()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	query := `
		SELECT 
			lb.id::text,
			COALESCE(lb.bus_booking_id::text, '') as bus_booking_id,
			COALESCE(bbs.passenger_name, '') as passenger_name,
			COALESCE(bbs.passenger_phone, '') as passenger_phone,
			COALESCE(bk.booking_reference, '') as booking_reference,
			COALESCE(lb.lounge_name, '') as lounge_name,
			COALESCE(lb.scheduled_arrival::text, lb.created_at::text) as scheduled_arrival,
			COALESCE(lb.pricing_type::text, 'Hourly') as pricing_type,
			COALESCE(lb.number_of_guests, 0) as number_of_guests,
			COALESCE(lb.selected_amenities::text, '[]') as amenities,
			COALESCE(lbp.product_names, '') as product_names,
			COALESCE(lb.booking_type::text, 'Regular') as booking_type,
			COALESCE(lb.total_amount, 0) as total_amount,
			COALESCE(lb.payment_status::text, 'Pending') as payment_status,
			COALESCE(lb.status::text, 'Pending') as status,
			COALESCE(lb.created_at::text, NOW()::text) as created_at
		FROM lounge_bookings lb
		LEFT JOIN LATERAL (
			SELECT passenger_name, passenger_phone 
			FROM bus_booking_seats 
			WHERE bus_booking_id = lb.bus_booking_id::uuid 
			LIMIT 1
		) bbs ON true
		LEFT JOIN bookings bk ON lb.bus_booking_id::uuid = bk.id
		LEFT JOIN LATERAL (
			SELECT STRING_AGG(product_name, ', ') as product_names
			FROM lounge_booking_pre_orders
			WHERE lounge_booking_id = lb.id
		) lbp ON true
		ORDER BY lb.created_at DESC
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error querying lounge bookings:", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, busBookingID, passengerName, passengerPhone, bookingRef string
		var loungeName, scheduledArrival, pricingType, amenitiesJSON, productNames, bookingType string
		var numberOfGuests int
		var totalAmount string
		var paymentStatus, status, createdAt string

		err := rows.Scan(
			&id,
			&busBookingID,
			&passengerName,
			&passengerPhone,
			&bookingRef,
			&loungeName,
			&scheduledArrival,
			&pricingType,
			&numberOfGuests,
			&amenitiesJSON,
			&productNames,
			&bookingType,
			&totalAmount,
			&paymentStatus,
			&status,
			&createdAt,
		)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}
		count++

		result := map[string]interface{}{
			"id":                id,
			"bus_booking_id":    busBookingID,
			"passenger_name":    passengerName,
			"passenger_phone":   passengerPhone,
			"booking_reference": bookingRef,
			"lounge_name":       loungeName,
			"scheduled_arrival": scheduledArrival,
			"pricing_type":      pricingType,
			"number_of_guests":  numberOfGuests,
			"product_names":     productNames,
			"booking_type":      bookingType,
			"total_amount":      totalAmount,
			"payment_status":    paymentStatus,
			"status":            status,
			"created_at":        createdAt,
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("Record %d:\n%s\n\n", count, string(jsonData))
	}

	fmt.Printf("\nTotal records found: %d\n", count)

	if err := rows.Err(); err != nil {
		log.Fatal("Error iterating rows:", err)
	}
}
