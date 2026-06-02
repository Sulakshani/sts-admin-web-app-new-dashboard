package main

import (
	"fmt"
	"log"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
)

func main() {
	cfg := config.LoadConfig()
	database.Init(cfg)

	// Test Database connection
	err := database.DB.Ping()
	if err != nil {
		log.Fatalf("Database Connection Check: %v", err)
	}
	log.Println("Database Connection Check: Successfully connected!")

	// Test the booking query
	repo := database.NewBookingRepository(database.DB)
	bookings, err := repo.GetAllBookings()
	if err != nil {
		log.Fatalf("Error fetching bookings: %v", err)
	}

	fmt.Printf("Found %d bookings\n", len(bookings))
	for i, booking := range bookings {
		fmt.Printf("\nBooking %d:\n", i+1)
		fmt.Printf("  ID: %s\n", booking.BookingID)
		fmt.Printf("  Reference: %s\n", booking.BookingReference)
		fmt.Printf("  Passenger: %s\n", booking.PassengerName)
		fmt.Printf("  Phone: %s\n", booking.PassengerPhone)
		fmt.Printf("  Route: %s\n", booking.Route)
		fmt.Printf("  Bus Type: %s\n", booking.BusType)
		fmt.Printf("  Seats: %s\n", booking.SeatNumber)
		fmt.Printf("  Total Fare: %.2f\n", booking.TotalFare)
		fmt.Printf("  Status: %s\n", booking.BookingStatus)
		fmt.Printf("  Payment: %s\n", booking.PaymentStatus)
	}
}
