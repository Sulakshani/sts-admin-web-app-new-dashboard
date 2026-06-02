package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"sts-backend/internal/models"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// GetAllBookings retrieves all bus bookings with detailed information
func (r *BookingRepository) GetAllBookings() ([]models.BusBooking, error) {
	query := `
		SELECT DISTINCT
			bb.id::text as booking_id,
			bb.scheduled_trip_id::text,
			COALESCE(bus.id::text, '') as bus_id,
			COALESCE(
				(SELECT bbs2.passenger_name FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_name,
				'Unknown'
			) as passenger_name,
			COALESCE(
				(SELECT bbs2.passenger_phone FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_phone,
				''
			) as passenger_phone,
			b.booking_reference,
			COALESCE(b.notes, bor.custom_route_name, mr.route_name, 'Unknown Route') as route,
			st.departure_datetime,
			COALESCE(bus.bus_type, 'normal') as bus_type,
		COALESCE(
			(
				SELECT STRING_AGG(ts2.seat_number, ', ' ORDER BY ts2.seat_number)
				FROM bus_booking_seats bbs2
				LEFT JOIN trip_seats ts2 ON bbs2.trip_seat_id = ts2.id
				WHERE bbs2.bus_booking_id = bb.id AND ts2.seat_number IS NOT NULL
			),
			''
			) as seat_numbers,
			bb.total_fare,
			b.payment_status,
			bb.status as booking_status,
			bb.created_at,
		CASE 
			WHEN bus.bus_number IS NOT NULL AND bus.bus_number != '' THEN bus.bus_number
			WHEN rp.permit_number IS NOT NULL AND rp.permit_number != '' THEN rp.permit_number
			ELSE 'BUS-' || SUBSTRING(st.id::text, 1, 8)
		END as bus_number,
			COALESCE(bus.license_plate, '') as license_plate,
			bb.number_of_seats
		FROM bus_bookings bb
		INNER JOIN bookings b ON bb.booking_id = b.id
		INNER JOIN scheduled_trips st ON bb.scheduled_trip_id = st.id
		LEFT JOIN bus_owner_routes bor ON st.bus_owner_route_id = bor.id
		LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
		LEFT JOIN buses bus ON st.permit_id = bus.permit_id
		ORDER BY bb.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying bookings: %v", err)
	}
	defer rows.Close()

	bookings := []models.BusBooking{}
	for rows.Next() {
		var booking models.BusBooking
		var seatNumbers sql.NullString

		err := rows.Scan(
			&booking.BookingID,
			&booking.ScheduledTripID,
			&booking.BusID,
			&booking.PassengerName,
			&booking.PassengerPhone,
			&booking.BookingReference,
			&booking.Route,
			&booking.DepartureDateTime,
			&booking.BusType,
			&seatNumbers,
			&booking.TotalFare,
			&booking.PaymentStatus,
			&booking.BookingStatus,
			&booking.CreatedAt,
			&booking.BusNumber,
			&booking.LicensePlate,
			&booking.NumberOfSeats,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning booking row: %v", err)
		}

		if seatNumbers.Valid {
			booking.SeatNumber = seatNumbers.String
		}

		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// CreateBooking inserts a new bus booking (simplified version - creates minimal required records)
func (r *BookingRepository) CreateBooking(booking *models.BusBooking) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Generate booking reference if not provided
	if booking.BookingReference == "" {
		now := time.Now()
		booking.BookingReference = fmt.Sprintf("BL-%s-%06X", 
			now.Format("20060102"), 
			now.UnixNano()%1000000)
	}

	// Find scheduled_trip_id based on departure_datetime if provided
	// This will also get the bus_id from the scheduled trip
	if !booking.DepartureDateTime.IsZero() && booking.ScheduledTripID == "" {
		var validTripID, busID string
		tripQuery := `
			SELECT st.id::text, COALESCE(b.id::text, '') as bus_id
			FROM scheduled_trips st
			LEFT JOIN buses b ON st.permit_id = b.permit_id
			WHERE st.departure_datetime::date = $1::date
			ORDER BY ABS(EXTRACT(EPOCH FROM (st.departure_datetime - $1::timestamp)))
			LIMIT 1
		`
		err := tx.QueryRow(tripQuery, booking.DepartureDateTime).Scan(&validTripID, &busID)
		if err != nil {
			// If no exact match, get any valid scheduled trip
			err = tx.QueryRow("SELECT id::text FROM scheduled_trips LIMIT 1").Scan(&validTripID)
			if err != nil {
				return fmt.Errorf("no valid scheduled trips found: %v", err)
			}
		} else {
			booking.BusID = busID
		}
		booking.ScheduledTripID = validTripID
	} else if booking.ScheduledTripID == "" {
		// If no departure_datetime provided, get any valid scheduled trip
		var validTripID string
		err := tx.QueryRow("SELECT id::text FROM scheduled_trips LIMIT 1").Scan(&validTripID)
		if err != nil {
			return fmt.Errorf("no valid scheduled trips found: %v", err)
		}
		booking.ScheduledTripID = validTripID
	}

	// If we have scheduled_trip_id but no bus_id, get bus_id from the trip
	if booking.BusID == "" && booking.ScheduledTripID != "" {
		var busID string
		busQuery := `
			SELECT COALESCE(b.id::text, '') as bus_id
			FROM scheduled_trips st
			LEFT JOIN buses b ON st.permit_id = b.permit_id
			WHERE st.id = $1
		`
		err = tx.QueryRow(busQuery, booking.ScheduledTripID).Scan(&busID)
		if err == nil {
			booking.BusID = busID
		}
	}

	// Get a default user_id (or create system user) - using first user from database
	var userID string
	err = tx.QueryRow("SELECT id::text FROM users LIMIT 1").Scan(&userID)
	if err != nil {
		// If no users exist, use a fixed UUID for system bookings
		userID = "00000000-0000-0000-0000-000000000001"
	}

	// 1. Create parent booking record
	var parentBookingID string
	parentQuery := `
		INSERT INTO bookings (
			booking_reference,
			user_id,
			booking_type,
			subtotal,
			total_amount,
			passenger_name,
			passenger_phone,
			payment_status,
			notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	// Store custom route in notes field
	customRoute := booking.Route
	if customRoute == "" {
		customRoute = "Custom Route"
	}
	
	err = tx.QueryRow(
		parentQuery,
		booking.BookingReference,
		userID,
		"bus_only",
		booking.TotalFare,
		booking.TotalFare,
		booking.PassengerName,
		booking.PassengerPhone,
		booking.PaymentStatus,
		customRoute,
	).Scan(&parentBookingID)
	
	if err != nil {
		return fmt.Errorf("error creating parent booking: %v", err)
	}

	// 2. Create bus_booking record
	var busBookingID string
	busBookingQuery := `
		INSERT INTO bus_bookings (
			booking_id,
			scheduled_trip_id,
			number_of_seats,
			fare_per_seat,
			total_fare,
			status
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	farePerSeat := booking.TotalFare
	if booking.NumberOfSeats > 0 {
		farePerSeat = booking.TotalFare / float64(booking.NumberOfSeats)
	}
	
	err = tx.QueryRow(
		busBookingQuery,
		parentBookingID,
		booking.ScheduledTripID,
		booking.NumberOfSeats,
		farePerSeat,
		booking.TotalFare,
		booking.BookingStatus,
	).Scan(&busBookingID)
	
	if err != nil {
		return fmt.Errorf("error creating bus_booking: %v", err)
	}

	// Commit the main transaction first
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// 3. Create seat records AFTER commit (optional, non-blocking)
	// This won't fail the booking if seat creation has issues
	if booking.SeatNumber != "" {
		seats := strings.Split(booking.SeatNumber, ",")
		for i, seat := range seats {
			seat = strings.TrimSpace(seat)
			if seat == "" {
				continue
			}

			// Try to find existing trip_seat
			var tripSeatID sql.NullString
			seatQuery := `SELECT id::text FROM trip_seats WHERE scheduled_trip_id = $1 AND seat_number = $2 LIMIT 1`
			r.db.QueryRow(seatQuery, booking.ScheduledTripID, seat).Scan(&tripSeatID)

			// If not found, try to create it with all required fields
			if !tripSeatID.Valid || tripSeatID.String == "" {
				createQuery := `
					INSERT INTO trip_seats (
						scheduled_trip_id, seat_number, row_number, position, seat_price, status
					) VALUES ($1, $2, $3, $4, $5, $6) 
					RETURNING id::text
				`
				// Use seat position as row/position (e.g., "A1" -> row=1, pos=1)
				rowNum := 1
				posNum := i + 1
				seatPrice := booking.TotalFare / float64(booking.NumberOfSeats)
				r.db.QueryRow(createQuery, booking.ScheduledTripID, seat, rowNum, posNum, seatPrice, "available").Scan(&tripSeatID)
			}

			// Insert bus_booking_seat record with scheduled_trip_id (required)
			if tripSeatID.Valid && tripSeatID.String != "" {
				insertQuery := `
					INSERT INTO bus_booking_seats (
						bus_booking_id, scheduled_trip_id, trip_seat_id, passenger_name, passenger_phone, is_primary_passenger
					) VALUES ($1, $2, $3, $4, $5, $6)
				`
				isPrimary := i == 0
				r.db.Exec(insertQuery, busBookingID, booking.ScheduledTripID, tripSeatID, booking.PassengerName, booking.PassengerPhone, isPrimary)
			}
		}
	}

	booking.BookingID = busBookingID
	return nil
}

// GetBookingByID retrieves a single booking by ID
func (r *BookingRepository) GetBookingByID(id string) (*models.BusBooking, error) {
	query := `
		SELECT 
			bb.id::text as booking_id,
			bb.scheduled_trip_id::text,
			COALESCE(bus.id::text, '') as bus_id,
			COALESCE(
				(SELECT bbs2.passenger_name FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_name,
				'Unknown'
			) as passenger_name,
			COALESCE(
				(SELECT bbs2.passenger_phone FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_phone,
				''
			) as passenger_phone,
			b.booking_reference,
			COALESCE(b.notes, bor.custom_route_name, mr.route_name, 'Unknown Route') as route,
			st.departure_datetime,
			COALESCE(bus.bus_type, 'normal') as bus_type,
			(
				SELECT STRING_AGG(ts2.seat_number, ', ' ORDER BY ts2.seat_number)
				FROM bus_booking_seats bbs2
				LEFT JOIN trip_seats ts2 ON bbs2.trip_seat_id = ts2.id
				WHERE bbs2.bus_booking_id = bb.id
			) as seat_numbers,
			bb.total_fare,
			b.payment_status,
			bb.status as booking_status,
			bb.created_at,
		CASE 
			WHEN bus.bus_number IS NOT NULL AND bus.bus_number != '' THEN bus.bus_number
			WHEN rp.permit_number IS NOT NULL AND rp.permit_number != '' THEN rp.permit_number
			ELSE 'BUS-' || SUBSTRING(st.id::text, 1, 8)
		END as bus_number,
		COALESCE(bus.license_plate, '') as license_plate,
		bb.number_of_seats
	FROM bus_bookings bb
	INNER JOIN bookings b ON bb.booking_id = b.id
	INNER JOIN scheduled_trips st ON bb.scheduled_trip_id = st.id
	LEFT JOIN bus_owner_routes bor ON st.bus_owner_route_id = bor.id
	LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
		LEFT JOIN buses bus ON st.permit_id = bus.permit_id
		WHERE bb.id = $1
	`

	var booking models.BusBooking
	var seatNumbers sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&booking.BookingID,
		&booking.ScheduledTripID,
		&booking.BusID,
		&booking.PassengerName,
		&booking.PassengerPhone,
		&booking.BookingReference,
		&booking.Route,
		&booking.DepartureDateTime,
		&booking.BusType,
		&seatNumbers,
		&booking.TotalFare,
		&booking.PaymentStatus,
		&booking.BookingStatus,
		&booking.CreatedAt,
		&booking.BusNumber,
		&booking.LicensePlate,
		&booking.NumberOfSeats,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting booking by id: %v", err)
	}

	if seatNumbers.Valid {
		booking.SeatNumber = seatNumbers.String
	}

	return &booking, nil
}

// UpdateBookingStatus updates the booking status
func (r *BookingRepository) UpdateBookingStatus(id string, status string) error {
	query := `UPDATE bus_bookings SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("error updating booking status: %v", err)
	}
	return nil
}

// UpdatePaymentStatus updates the payment status
func (r *BookingRepository) UpdatePaymentStatus(bookingID string, status string) error {
	// Get the booking_id from bus_bookings to update the parent bookings table
	var parentBookingID string
	err := r.db.QueryRow(`SELECT booking_id FROM bus_bookings WHERE id = $1`, bookingID).Scan(&parentBookingID)
	if err != nil {
		return fmt.Errorf("error finding parent booking: %v", err)
	}

	query := `UPDATE bookings SET payment_status = $1, updated_at = NOW() WHERE id = $2`
	_, err = r.db.Exec(query, status, parentBookingID)
	if err != nil {
		return fmt.Errorf("error updating payment status: %v", err)
	}
	return nil
}

// UpdateBooking updates a complete booking record
func (r *BookingRepository) UpdateBooking(booking *models.BusBooking) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get the scheduled_trip_id for this booking
	var scheduledTripID string
	err = tx.QueryRow("SELECT scheduled_trip_id FROM bus_bookings WHERE id = $1", booking.BookingID).Scan(&scheduledTripID)
	if err != nil {
		return fmt.Errorf("error getting scheduled_trip_id: %v", err)
	}

	// Update bus_bookings table with all editable fields
	updateBusBookingQuery := `
		UPDATE bus_bookings 
		SET status = $1,
		    fare_per_seat = $2,
		    number_of_seats = $3,
		    total_fare = $4,
		    updated_at = NOW()
		WHERE id = $5
	`
	farePerSeat := booking.TotalFare
	if booking.NumberOfSeats > 0 {
		farePerSeat = booking.TotalFare / float64(booking.NumberOfSeats)
	}
	_, err = tx.Exec(updateBusBookingQuery, booking.BookingStatus, farePerSeat, booking.NumberOfSeats, booking.TotalFare, booking.BookingID)
	if err != nil {
		return fmt.Errorf("error updating bus_booking: %v", err)
	}

	// Update bookings table (payment_status and passenger info)
	updateBookingQuery := `
		UPDATE bookings 
		SET payment_status = $1,
		    passenger_name = $2,
		    passenger_phone = $3,
		    updated_at = NOW()
		WHERE id IN (
			SELECT booking_id FROM bus_bookings WHERE id = $4
		)
	`
	_, err = tx.Exec(updateBookingQuery, booking.PaymentStatus, booking.PassengerName, booking.PassengerPhone, booking.BookingID)
	if err != nil {
		return fmt.Errorf("error updating booking: %v", err)
	}

	// If seat numbers are provided, update seat assignments
	if booking.SeatNumber != "" {
		// Delete existing seat assignments
		_, err = tx.Exec("DELETE FROM bus_booking_seats WHERE bus_booking_id = $1", booking.BookingID)
		if err != nil {
			return fmt.Errorf("error deleting old seat assignments: %v", err)
		}

		// Parse seat numbers
		seatNumbers := strings.Split(booking.SeatNumber, ",")
		for i, seatStr := range seatNumbers {
			seatNum := strings.TrimSpace(seatStr)
			if seatNum == "" {
				continue
			}

			// Find the trip_seat_id for this seat number
			var tripSeatID string
			err = tx.QueryRow(`
				SELECT id FROM trip_seats 
				WHERE scheduled_trip_id = $1 AND seat_number = $2
				LIMIT 1
			`, scheduledTripID, seatNum).Scan(&tripSeatID)
			
			if err != nil {
				// If seat doesn't exist, create it with integer position
				err = tx.QueryRow(`
					INSERT INTO trip_seats (scheduled_trip_id, seat_number, row_number, position, seat_price, status, created_at, updated_at)
					VALUES ($1, $2, 1, $3, $4, 'available', NOW(), NOW())
					RETURNING id
				`, scheduledTripID, seatNum, i+1, farePerSeat).Scan(&tripSeatID)
				if err != nil {
					return fmt.Errorf("error creating trip_seat for %s: %v", seatNum, err)
				}
			}

			// Create bus_booking_seat record
			isPrimary := (i == 0)
			_, err = tx.Exec(`
				INSERT INTO bus_booking_seats (bus_booking_id, trip_seat_id, passenger_name, passenger_phone, is_primary_passenger, scheduled_trip_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			`, booking.BookingID, tripSeatID, booking.PassengerName, booking.PassengerPhone, isPrimary, scheduledTripID)
			if err != nil {
				return fmt.Errorf("error creating bus_booking_seat: %v", err)
			}

			// Update trip_seat status
			_, err = tx.Exec("UPDATE trip_seats SET status = 'booked', updated_at = NOW() WHERE id = $1", tripSeatID)
			if err != nil {
				return fmt.Errorf("error updating trip_seat status: %v", err)
			}
		}
	} else {
		// Just update primary passenger info if no seat number changes
		updateSeatQuery := `
			UPDATE bus_booking_seats 
			SET passenger_name = $1,
			    passenger_phone = $2,
			    updated_at = NOW()
			WHERE bus_booking_id = $3 
			  AND is_primary_passenger = true
		`
		_, err = tx.Exec(updateSeatQuery, booking.PassengerName, booking.PassengerPhone, booking.BookingID)
		if err != nil {
			return fmt.Errorf("error updating seat info: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// GetBookingsByStatus retrieves bookings filtered by status
func (r *BookingRepository) GetBookingsByStatus(status string) ([]models.BusBooking, error) {
	query := `
		SELECT DISTINCT
			bb.id::text as booking_id,
			bb.scheduled_trip_id::text,
			COALESCE(bus.id::text, '') as bus_id,
			COALESCE(
				(SELECT bbs2.passenger_name FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_name,
				'Unknown'
			) as passenger_name,
			COALESCE(
				(SELECT bbs2.passenger_phone FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_phone,
				''
			) as passenger_phone,
			b.booking_reference,
			COALESCE(b.notes, bor.custom_route_name, mr.route_name, 'Unknown Route') as route,
			st.departure_datetime,
			COALESCE(bus.bus_type, 'normal') as bus_type,
			(
				SELECT STRING_AGG(ts2.seat_number, ', ' ORDER BY ts2.seat_number)
				FROM bus_booking_seats bbs2
				LEFT JOIN trip_seats ts2 ON bbs2.trip_seat_id = ts2.id
				WHERE bbs2.bus_booking_id = bb.id
			) as seat_numbers,
			bb.total_fare,
			b.payment_status,
			bb.status as booking_status,
			bb.created_at,
		CASE 
			WHEN bus.bus_number IS NOT NULL AND bus.bus_number != '' THEN bus.bus_number
			WHEN rp.permit_number IS NOT NULL AND rp.permit_number != '' THEN rp.permit_number
			ELSE 'BUS-' || SUBSTRING(st.id::text, 1, 8)
		END as bus_number,
		COALESCE(bus.license_plate, '') as license_plate,
		bb.number_of_seats
	FROM bus_bookings bb
	INNER JOIN bookings b ON bb.booking_id = b.id
	INNER JOIN scheduled_trips st ON bb.scheduled_trip_id = st.id
	LEFT JOIN bus_owner_routes bor ON st.bus_owner_route_id = bor.id
	LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
		LEFT JOIN buses bus ON st.permit_id = bus.permit_id
		WHERE bb.status = $1
		ORDER BY bb.created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("error querying bookings by status: %v", err)
	}
	defer rows.Close()

	bookings := []models.BusBooking{}
	for rows.Next() {
		var booking models.BusBooking
		var seatNumbers sql.NullString

		err := rows.Scan(
			&booking.BookingID,
			&booking.ScheduledTripID,
			&booking.BusID,
			&booking.PassengerName,
			&booking.PassengerPhone,
			&booking.BookingReference,
			&booking.Route,
			&booking.DepartureDateTime,
			&booking.BusType,
			&seatNumbers,
			&booking.TotalFare,
			&booking.PaymentStatus,
			&booking.BookingStatus,
			&booking.CreatedAt,
			&booking.BusNumber,
			&booking.LicensePlate,
			&booking.NumberOfSeats,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning booking row: %v", err)
		}

		if seatNumbers.Valid {
			booking.SeatNumber = seatNumbers.String
		}

		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// SearchBookings searches bookings by reference number or passenger name/phone
func (r *BookingRepository) SearchBookings(searchTerm string) ([]models.BusBooking, error) {
	query := `
		SELECT DISTINCT
			bb.id::text as booking_id,
			bb.scheduled_trip_id::text,
			COALESCE(bus.id::text, '') as bus_id,
			COALESCE(
				(SELECT bbs2.passenger_name FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_name,
				'Unknown'
			) as passenger_name,
			COALESCE(
				(SELECT bbs2.passenger_phone FROM bus_booking_seats bbs2 WHERE bbs2.bus_booking_id = bb.id AND bbs2.is_primary_passenger = true LIMIT 1),
				b.passenger_phone,
				''
			) as passenger_phone,
			b.booking_reference,
			COALESCE(b.notes, bor.custom_route_name, mr.route_name, 'Unknown Route') as route,
			st.departure_datetime,
			COALESCE(bus.bus_type, 'normal') as bus_type,
			(
				SELECT STRING_AGG(ts2.seat_number, ', ' ORDER BY ts2.seat_number)
				FROM bus_booking_seats bbs2
				LEFT JOIN trip_seats ts2 ON bbs2.trip_seat_id = ts2.id
				WHERE bbs2.bus_booking_id = bb.id
			) as seat_numbers,
			bb.total_fare,
			b.payment_status,
			bb.status as booking_status,
			bb.created_at,
		CASE 
			WHEN bus.bus_number IS NOT NULL AND bus.bus_number != '' THEN bus.bus_number
			WHEN rp.permit_number IS NOT NULL AND rp.permit_number != '' THEN rp.permit_number
			ELSE 'BUS-' || SUBSTRING(st.id::text, 1, 8)
		END as bus_number,
		COALESCE(bus.license_plate, '') as license_plate,
		bb.number_of_seats
	FROM bus_bookings bb
	INNER JOIN bookings b ON bb.booking_id = b.id
	INNER JOIN scheduled_trips st ON bb.scheduled_trip_id = st.id
	LEFT JOIN bus_owner_routes bor ON st.bus_owner_route_id = bor.id
	LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN bus_owner_routes bor ON st.bus_owner_route_id = bor.id
		LEFT JOIN route_permits rp ON st.permit_id = rp.id
		LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
		LEFT JOIN buses bus ON st.permit_id = bus.permit_id
		LEFT JOIN bus_booking_seats bbs ON bb.id = bbs.bus_booking_id
		WHERE 
			b.booking_reference ILIKE $1 OR
			b.passenger_name ILIKE $1 OR
			b.passenger_phone ILIKE $1 OR
			bbs.passenger_name ILIKE $1 OR
			bbs.passenger_phone ILIKE $1
		ORDER BY bb.created_at DESC
	`

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"
	rows, err := r.db.Query(query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("error searching bookings: %v", err)
	}
	defer rows.Close()

	bookings := []models.BusBooking{}
	for rows.Next() {
		var booking models.BusBooking
		var seatNumbers sql.NullString

		err := rows.Scan(
			&booking.BookingID,
			&booking.ScheduledTripID,
			&booking.BusID,
			&booking.PassengerName,
			&booking.PassengerPhone,
			&booking.BookingReference,
			&booking.Route,
			&booking.DepartureDateTime,
			&booking.BusType,
			&seatNumbers,
			&booking.TotalFare,
			&booking.PaymentStatus,
			&booking.BookingStatus,
			&booking.CreatedAt,
			&booking.BusNumber,
			&booking.LicensePlate,
			&booking.NumberOfSeats,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning booking row: %v", err)
		}

		if seatNumbers.Valid {
			booking.SeatNumber = seatNumbers.String
		}

		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// DeleteBooking cancels/deletes a booking
func (r *BookingRepository) DeleteBooking(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Mark booking as cancelled instead of deleting
	_, err = tx.Exec(`UPDATE bus_bookings SET status = 'cancelled', cancelled_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error cancelling booking: %v", err)
	}

	// Update parent booking status
	_, err = tx.Exec(`
		UPDATE bookings 
		SET booking_status = 'cancelled', cancelled_at = NOW()
		WHERE id = (SELECT booking_id FROM bus_bookings WHERE id = $1)
	`, id)
	if err != nil {
		return fmt.Errorf("error updating parent booking status: %v", err)
	}

	return tx.Commit()
}
