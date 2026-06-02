package models

import "time"

type LoungeBooking struct {
	LoungeBookingID   string    `json:"lounge_booking_id"`
	BusBookingID      *string   `json:"bus_booking_id"`
	PassengerName     string    `json:"passenger_name"`
	PassengerPhone    string    `json:"passenger_phone"`
	BookingReference  string    `json:"booking_reference"`
	LoungeName        string    `json:"lounge_name"`
	ScheduledArrival  time.Time `json:"scheduled_arrival"`
	PricingType       string    `json:"pricing_type"` // Duration field
	NumberOfGuests    int       `json:"number_of_guests"`
	SelectedAmenities []string  `json:"selected_amenities"` // Additional features
	ProductName       string    `json:"product_name"`       // Market place
	BookingType       string    `json:"booking_type"`
	TotalAmount       float64   `json:"total_amount"`
	PaymentStatus     string    `json:"payment_status"` // Pending, Paid, Failed
	Status            string    `json:"status"`         // Confirmed, Pending, Cancelled, Completed
	CreatedAt         time.Time `json:"created_at"`
}
