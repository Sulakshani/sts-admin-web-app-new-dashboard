package models

import (
	"encoding/json"
	"time"
)

type BusBooking struct {
	BookingID         string    `json:"booking_id"`
	ScheduledTripID   string    `json:"scheduled_trip_id"`
	BusID             string    `json:"bus_id"`
	PassengerName     string    `json:"passenger_name"`
	PassengerPhone    string    `json:"passenger_phone"`
	BookingReference  string    `json:"booking_reference"`
	Route             string    `json:"route"`
	DepartureDateTime time.Time `json:"departure_datetime"`
	BusType           string    `json:"bus_type"`
	SeatNumber        string    `json:"seat_number"`
	TotalFare         float64   `json:"total_fare"`
	PaymentStatus     string    `json:"payment_status"`
	BookingStatus     string    `json:"booking_status"`
	CreatedAt         time.Time `json:"created_at"`
	BusNumber         string    `json:"bus_number"`
	LicensePlate      string    `json:"license_plate"`
	NumberOfSeats     int       `json:"number_of_seats"`
}

// UnmarshalJSON handles multiple datetime formats from frontend
func (b *BusBooking) UnmarshalJSON(data []byte) error {
	type Alias BusBooking
	aux := &struct {
		DepartureDateTime string `json:"departure_datetime"`
		CreatedAt         string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try multiple datetime formats
	formats := []string{
		time.RFC3339,                 // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05",        // Without timezone
		"2006-01-02T15:04",           // datetime-local format (without seconds)
		"2006-01-02 15:04:05",        // Space separator
		"2006-01-02 15:04",           // Space separator without seconds
	}

	// Parse DepartureDateTime
	if aux.DepartureDateTime != "" {
		var err error
		for _, format := range formats {
			b.DepartureDateTime, err = time.Parse(format, aux.DepartureDateTime)
			if err == nil {
				break
			}
		}
	}

	// Parse CreatedAt
	if aux.CreatedAt != "" {
		var err error
		for _, format := range formats {
			b.CreatedAt, err = time.Parse(format, aux.CreatedAt)
			if err == nil {
				break
			}
		}
	} else {
		b.CreatedAt = time.Now()
	}

	return nil
}
