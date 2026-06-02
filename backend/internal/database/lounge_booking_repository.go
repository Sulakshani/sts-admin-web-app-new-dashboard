package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sts-backend/internal/models"
)

type LoungeBookingRepository struct {
	db *sql.DB
}

func NewLoungeBookingRepository(db *sql.DB) *LoungeBookingRepository {
	return &LoungeBookingRepository{db: db}
}

// GetLoungeBookings retrieves all lounge bookings with joined data from related tables
func (r *LoungeBookingRepository) GetLoungeBookings() ([]models.LoungeBooking, error) {
	query := `
		SELECT 
			lb.id::text,
			COALESCE(lb.bus_booking_id::text, ''),
			COALESCE(lb.primary_guest_name, bk.passenger_name, ''),
			COALESCE(lb.primary_guest_phone, bk.passenger_phone, ''),
			COALESCE(lb.booking_reference, bk.booking_reference, ''),
			COALESCE(lb.lounge_name, ''),
			COALESCE(lb.scheduled_arrival, lb.created_at),
			COALESCE(lb.pricing_type::text, 'Hourly'),
			COALESCE(lb.number_of_guests, 0),
			COALESCE(lb.selected_amenities::text, '[]'),
			COALESCE(lbp.product_names, ''),
			COALESCE(lb.booking_type::text, 'Regular'),
			COALESCE(lb.total_amount, 0),
			COALESCE(lb.payment_status::text, 'Pending'),
			COALESCE(lb.status::text, 'Pending'),
			COALESCE(lb.created_at, NOW())
		FROM lounge_bookings lb
	LEFT JOIN bookings bk ON lb.master_booking_id = bk.id
		LEFT JOIN LATERAL (
			SELECT STRING_AGG(product_name, ', ') as product_names
			FROM lounge_booking_pre_orders
			WHERE lounge_booking_id = lb.id
		) lbp ON true
		ORDER BY lb.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying lounge bookings: %v", err)
	}
	defer rows.Close()

	var bookings []models.LoungeBooking
	for rows.Next() {
		var lb models.LoungeBooking
		var amenitiesJSON string
		var busBookingIDStr string

		err := rows.Scan(
			&lb.LoungeBookingID,
			&busBookingIDStr,
			&lb.PassengerName,
			&lb.PassengerPhone,
			&lb.BookingReference,
			&lb.LoungeName,
			&lb.ScheduledArrival,
			&lb.PricingType,
			&lb.NumberOfGuests,
			&amenitiesJSON,
			&lb.ProductName,
			&lb.BookingType,
			&lb.TotalAmount,
			&lb.PaymentStatus,
			&lb.Status,
			&lb.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning lounge booking row: %v", err)
		}

		// Handle nullable bus_booking_id
		if busBookingIDStr != "" {
			lb.BusBookingID = &busBookingIDStr
		}

		// Parse amenities JSON array
		if amenitiesJSON != "" && amenitiesJSON != "[]" {
			if err := json.Unmarshal([]byte(amenitiesJSON), &lb.SelectedAmenities); err != nil {
				lb.SelectedAmenities = []string{}
			}
		} else {
			lb.SelectedAmenities = []string{}
		}

		bookings = append(bookings, lb)
	}

	return bookings, nil
}

// GetLoungeBookingByID retrieves a single lounge booking by ID
func (r *LoungeBookingRepository) GetLoungeBookingByID(id string) (*models.LoungeBooking, error) {
	query := `
		SELECT 
			lb.id::text,
			COALESCE(lb.bus_booking_id::text, ''),
			COALESCE(lb.primary_guest_name, bk.passenger_name, ''),
			COALESCE(lb.primary_guest_phone, bk.passenger_phone, ''),
			COALESCE(lb.booking_reference, bk.booking_reference, ''),
			COALESCE(lb.lounge_name, ''),
			COALESCE(lb.scheduled_arrival, lb.created_at),
			COALESCE(lb.pricing_type::text, 'Hourly'),
			COALESCE(lb.number_of_guests, 0),
			COALESCE(lb.selected_amenities::text, '[]'),
			COALESCE(lbp.product_names, ''),
			COALESCE(lb.booking_type::text, 'Regular'),
			COALESCE(lb.total_amount, 0),
			COALESCE(lb.payment_status::text, 'Pending'),
			COALESCE(lb.status::text, 'Pending'),
			COALESCE(lb.created_at, NOW())
		FROM lounge_bookings lb
	LEFT JOIN bookings bk ON lb.master_booking_id = bk.id
		LEFT JOIN LATERAL (
			SELECT STRING_AGG(product_name, ', ') as product_names
			FROM lounge_booking_pre_orders
			WHERE lounge_booking_id = lb.id
		) lbp ON true
		WHERE lb.id = $1::uuid
	`

	var lb models.LoungeBooking
	var amenitiesJSON string
	var busBookingIDStr string

	err := r.db.QueryRow(query, id).Scan(
		&lb.LoungeBookingID,
		&busBookingIDStr,
		&lb.PassengerName,
		&lb.PassengerPhone,
		&lb.BookingReference,
		&lb.LoungeName,
		&lb.ScheduledArrival,
		&lb.PricingType,
		&lb.NumberOfGuests,
		&amenitiesJSON,
		&lb.ProductName,
		&lb.BookingType,
		&lb.TotalAmount,
		&lb.PaymentStatus,
		&lb.Status,
		&lb.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lounge booking not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error querying lounge booking: %v", err)
	}

	// Handle nullable bus_booking_id
	if busBookingIDStr != "" {
		lb.BusBookingID = &busBookingIDStr
	}

	// Parse amenities JSON array
	if amenitiesJSON != "" && amenitiesJSON != "[]" {
		if err := json.Unmarshal([]byte(amenitiesJSON), &lb.SelectedAmenities); err != nil {
			lb.SelectedAmenities = []string{}
		}
	} else {
		lb.SelectedAmenities = []string{}
	}

	return &lb, nil
}

// CreateLoungeBooking creates a new lounge booking
func (r *LoungeBookingRepository) CreateLoungeBooking(lb *models.LoungeBooking) error {
	// Convert amenities to JSON
	amenitiesJSON, err := json.Marshal(lb.SelectedAmenities)
	if err != nil {
		return fmt.Errorf("error marshaling amenities: %v", err)
	}

	query := `
		INSERT INTO lounge_bookings (
			bus_booking_id, lounge_name, scheduled_arrival, pricing_type, number_of_guests,
			selected_amenities, booking_type, total_amount, payment_status, status,
			primary_guest_name, primary_guest_phone, booking_reference
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	err = r.db.QueryRow(
		query,
		lb.BusBookingID,
		lb.LoungeName,
		lb.ScheduledArrival,
		lb.PricingType,
		lb.NumberOfGuests,
		amenitiesJSON,
		lb.BookingType,
		lb.TotalAmount,
		lb.PaymentStatus,
		lb.Status,
		lb.PassengerName,
		lb.PassengerPhone,
		lb.BookingReference,
	).Scan(&lb.LoungeBookingID)

	if err != nil {
		return fmt.Errorf("error creating lounge booking: %v", err)
	}

	// If product_name (marketplace) is provided, insert into lounge_booking_pre_orders
	// Note: product_id, product_type, unit_price, and total_price are nullable now
	if lb.ProductName != "" {
		_, err = r.db.Exec(`
			INSERT INTO lounge_booking_pre_orders (
				lounge_booking_id, 
				product_name, 
				quantity
			)
			VALUES ($1, $2, 1)
		`, lb.LoungeBookingID, lb.ProductName)
		if err != nil {
			return fmt.Errorf("error creating pre-order: %v", err)
		}
	}

	return nil
}

// UpdateLoungeBooking updates an existing lounge booking
func (r *LoungeBookingRepository) UpdateLoungeBooking(lb *models.LoungeBooking) error {
	// Convert amenities to JSON
	amenitiesJSON, err := json.Marshal(lb.SelectedAmenities)
	if err != nil {
		return fmt.Errorf("error marshaling amenities: %v", err)
	}

	query := `
		UPDATE lounge_bookings SET
			lounge_name = $1,
			scheduled_arrival = $2,
			pricing_type = $3,
			number_of_guests = $4,
			selected_amenities = $5,
			booking_type = $6,
			total_amount = $7,
			payment_status = $8,
			status = $9,
			primary_guest_name = $10,
			primary_guest_phone = $11,
			booking_reference = $12
		WHERE id = $13::uuid
	`

	result, err := r.db.Exec(
		query,
		lb.LoungeName,
		lb.ScheduledArrival,
		lb.PricingType,
		lb.NumberOfGuests,
		amenitiesJSON,
		lb.BookingType,
		lb.TotalAmount,
		lb.PaymentStatus,
		lb.Status,
		lb.PassengerName,
		lb.PassengerPhone,
		lb.BookingReference,
		lb.LoungeBookingID,
	)

	if err != nil {
		return fmt.Errorf("error updating lounge booking: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("lounge booking not found")
	}

	// Update product_name in lounge_booking_pre_orders (delete old, insert new)
	if lb.ProductName != "" {
		// First delete any existing pre-orders for this booking
		_, err = r.db.Exec(
			`DELETE FROM lounge_booking_pre_orders WHERE lounge_booking_id = $1::uuid`,
			lb.LoungeBookingID,
		)
		if err != nil {
			return fmt.Errorf("error deleting old pre-orders: %v", err)
		}

		// Then insert the new pre-order
		_, err = r.db.Exec(
			`INSERT INTO lounge_booking_pre_orders (lounge_booking_id, product_name, quantity) VALUES ($1::uuid, $2, 1)`,
			lb.LoungeBookingID,
			lb.ProductName,
		)
		if err != nil {
			return fmt.Errorf("error inserting pre-order: %v", err)
		}
	} else {
		// If no product_name, delete any existing pre-orders
		_, err = r.db.Exec(
			`DELETE FROM lounge_booking_pre_orders WHERE lounge_booking_id = $1::uuid`,
			lb.LoungeBookingID,
		)
		if err != nil {
			return fmt.Errorf("error deleting pre-orders: %v", err)
		}
	}

	return nil
}

// UpdatePaymentStatus updates the payment status of a lounge booking
func (r *LoungeBookingRepository) UpdatePaymentStatus(id string, status string) error {
	// Validate status
	validStatuses := []string{"Pending", "Paid", "Failed"}
	isValid := false
	for _, s := range validStatuses {
		if strings.EqualFold(status, s) {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid payment status: %s", status)
	}

	query := `UPDATE lounge_bookings SET payment_status = $1 WHERE id = $2::uuid`
	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("error updating payment status: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("lounge booking not found")
	}

	return nil
}

// UpdateBookingStatus updates the booking status of a lounge booking
func (r *LoungeBookingRepository) UpdateBookingStatus(id string, status string) error {
	// Validate status
	validStatuses := []string{"Confirmed", "Pending", "Cancelled", "Completed"}
	isValid := false
	for _, s := range validStatuses {
		if strings.EqualFold(status, s) {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid booking status: %s", status)
	}

	query := `UPDATE lounge_bookings SET status = $1 WHERE id = $2::uuid`
	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("error updating booking status: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("lounge booking not found")
	}

	return nil
}

// DeleteLoungeBooking deletes a lounge booking
func (r *LoungeBookingRepository) DeleteLoungeBooking(id string) error {
	// First delete related pre-orders
	_, err := r.db.Exec(`DELETE FROM lounge_booking_pre_orders WHERE lounge_booking_id = $1::uuid`, id)
	if err != nil {
		return fmt.Errorf("error deleting pre-orders: %v", err)
	}

	// Then delete the booking
	query := `DELETE FROM lounge_bookings WHERE id = $1::uuid`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting lounge booking: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("lounge booking not found")
	}

	return nil
}
