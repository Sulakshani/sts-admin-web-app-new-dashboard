package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sts-backend/internal/models"
)

type BusRepository struct {
	db *sql.DB
}

func NewBusRepository(db *sql.DB) *BusRepository {
	return &BusRepository{db: db}
}

// GetAllBuses retrieves all buses with their related details
func (r *BusRepository) GetAllBuses() ([]models.Bus, error) {
	query := `
		select 
			b.id::text,
			COALESCE(b.bus_owner_id::text, ''),
			COALESCE(b.permit_id::text, ''),
			COALESCE(b.seat_layout_id::text, ''),
			COALESCE(b.bus_number, ''),
			COALESCE(bo.company_name, ''),
			COALESCE(bo.identity_or_incorporation_no, ''),
			COALESCE(bo.business_email, ''),
			COALESCE(bo.business_phone, ''),
			COALESCE(rp.permit_number, ''),
			b.license_plate,
			COALESCE(bslt.total_seats, rp.approved_seating_capacity, 0) as total_seats,
			COALESCE(b.bus_type, 'Standard'),
			COALESCE(mr.route_name, 'Unknown Route') as route_name,
			COALESCE(rp.approved_fare::float8, 0.0),
			COALESCE(b.status, 'Inactive'),
			COALESCE(rp.status::text, 'Pending') as verification_status,
		COALESCE(NULLIF(bo.verification_status::text, ''), 'Pending') as owner_verification_status,
			COALESCE(rp.verification_documents::text, '{}')
		from buses b
		left join bus_owners bo on b.bus_owner_id = bo.id
		left join route_permits rp on b.permit_id = rp.id
		left join master_routes mr on rp.master_route_id = mr.id
		left join bus_seat_layout_templates bslt on b.seat_layout_id = bslt.id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying buses: %v", err)
	}
	defer rows.Close()

	buses := []models.Bus{}
	for rows.Next() {
		var bus models.Bus
		var verificationDocsStr sql.NullString
		var seatLayoutID sql.NullString

		err := rows.Scan(
			&bus.ID,
			&bus.BusOwnerID,
			&bus.PermitID,
			&seatLayoutID,
			&bus.BusNumber,
			&bus.CompanyName,
			&bus.IdentifyOrIncorporationNo,
			&bus.BusinessEmail,
			&bus.BusinessPhone,
			&bus.PermitNumber,
			&bus.LicensePlate,
			&bus.TotalSeats,
			&bus.BusType,
			&bus.CustomRouteName,
			&bus.FarePerSeat,
			&bus.Status,
			&bus.VerificationStatus,
			&bus.OwnerVerificationStatus,
			&verificationDocsStr,
		)
		if err != nil {
			fmt.Printf("Error scanning bus: %v\n", err)
			return nil, fmt.Errorf("error scanning bus row: %v", err)
		}

		bus.SeatLayoutID = seatLayoutID

		if verificationDocsStr.Valid && verificationDocsStr.String != "" {
			if err := json.Unmarshal([]byte(verificationDocsStr.String), &bus.VerificationDocuments); err != nil {
				bus.VerificationDocuments = []string{}
			}
		}

		buses = append(buses, bus)
	}

	return buses, nil
}

// GetBusByID retrieves a single bus by ID
func (r *BusRepository) GetBusByID(id string) (*models.Bus, error) {
	query := `
		SELECT 
			b.id::text,
			COALESCE(b.bus_owner_id::text, ''),
			COALESCE(b.permit_id::text, ''),
			COALESCE(b.seat_layout_id::text, ''),
			COALESCE(b.bus_number, ''),
			COALESCE(bo.company_name, ''),
			COALESCE(bo.identity_or_incorporation_no, ''),
			COALESCE(bo.business_email, ''),
			COALESCE(bo.business_phone, ''),
			COALESCE(rp.permit_number, ''),
			COALESCE(b.license_plate, ''),
			COALESCE(bslt.total_seats, rp.approved_seating_capacity, 0) as total_seats,
			COALESCE(b.bus_type, 'Standard'),
			COALESCE(mr.route_name, 'Unknown Route') as route_name,
			COALESCE(rp.approved_fare::float8, 0.0),
			COALESCE(b.status, 'Inactive'),
			COALESCE(rp.status::text, 'Pending') as verification_status,
		COALESCE(NULLIF(bo.verification_status::text, ''), 'Pending') as owner_verification_status,
			COALESCE(rp.verification_documents::text, '{}')
		FROM buses b
		LEFT JOIN bus_owners bo ON b.bus_owner_id = bo.id
		LEFT JOIN route_permits rp ON b.permit_id = rp.id
		LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
		LEFT JOIN bus_seat_layout_templates bslt ON b.seat_layout_id = bslt.id
		WHERE b.id = $1
	`

	var bus models.Bus
	var verificationDocsStr sql.NullString
	var seatLayoutID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&bus.ID,
		&bus.BusOwnerID,
		&bus.PermitID,
		&seatLayoutID,
		&bus.BusNumber,
		&bus.CompanyName,
		&bus.IdentifyOrIncorporationNo,
		&bus.BusinessEmail,
		&bus.BusinessPhone,
		&bus.PermitNumber,
		&bus.LicensePlate,
		&bus.TotalSeats,
		&bus.BusType,
		&bus.CustomRouteName,
		&bus.FarePerSeat,
		&bus.Status,
		&bus.VerificationStatus,
		&bus.OwnerVerificationStatus,
		&verificationDocsStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or custom error
		}
		return nil, fmt.Errorf("error getting bus by id: %v", err)
	}

	bus.SeatLayoutID = seatLayoutID

	if verificationDocsStr.Valid && verificationDocsStr.String != "" {
		if err := json.Unmarshal([]byte(verificationDocsStr.String), &bus.VerificationDocuments); err != nil {
			bus.VerificationDocuments = []string{}
		}
	}

	return &bus, nil
}

// CreateBus creates a new bus record along with owner and permit details
func (r *BusRepository) CreateBus(bus *models.Bus) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in CreateBus: %v\n", r)
		}
	}()

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Insert/Get Bus Owner
	var busOwnerID string
	// Check if owner exists by identity_or_incorporation_no
	err = tx.QueryRow(`SELECT id FROM bus_owners WHERE identity_or_incorporation_no = $1`, bus.IdentifyOrIncorporationNo).Scan(&busOwnerID)

	if err == sql.ErrNoRows {
		// Owner doesn't exist, create a new one without user_id
		// user_id can be assigned later when the owner registers/logs in
		err = tx.QueryRow(`
			INSERT INTO bus_owners (company_name, identity_or_incorporation_no, business_email, business_phone)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, bus.CompanyName, bus.IdentifyOrIncorporationNo, bus.BusinessEmail, bus.BusinessPhone).Scan(&busOwnerID)
		if err != nil {
			return fmt.Errorf("error creating bus owner: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking bus owner: %v", err)
	}

	// 2. Insert/Get Master Route
	var masterRouteID string
	// Check if route exists
	err = tx.QueryRow(`SELECT id FROM master_routes WHERE route_name = $1`, bus.CustomRouteName).Scan(&masterRouteID)
	if err == sql.ErrNoRows {
		// Create new route
		// We insert id because it has a not-null constraint and likely no default value.
		// 'route_id' column does not exist in this schema.
		err = tx.QueryRow(`
			INSERT INTO master_routes (id, route_name, route_number, origin_city, destination_city)
			VALUES (uuid_generate_v4(), $1, 'TEMP-' || substring(md5(random()::text) from 1 for 6), 'Unknown', 'Unknown')
			RETURNING id
		`, bus.CustomRouteName).Scan(&masterRouteID)
		if err != nil {
			return fmt.Errorf("error creating master route: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking master route: %v", err)
	}

	// 3. Insert Route Permit
	var permitID string
	// Convert verification documents to JSON string
	docsJSON, _ := json.Marshal(bus.VerificationDocuments)

	err = tx.QueryRow(`
		INSERT INTO route_permits (
			permit_number, 
			approved_seating_capacity, 
			approved_fare, 
			status, 
			verification_documents,
			master_route_id,
			bus_owner_id,
			bus_registration_number,
			issue_date,
			expiry_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW() + interval '1 year')
		RETURNING id
	`, bus.PermitNumber, bus.TotalSeats, bus.FarePerSeat, strings.ToLower(bus.VerificationStatus), docsJSON, masterRouteID, busOwnerID, bus.LicensePlate).Scan(&permitID)
	if err != nil {
		return fmt.Errorf("error creating route permit: %v", err)
	}

	// 4. Insert Bus
	// Handle seat_layout_id - if not provided, set to NULL
	var seatLayoutIDValue interface{}
	if bus.SeatLayoutID.Valid {
		seatLayoutIDValue = bus.SeatLayoutID.String
	} else {
		seatLayoutIDValue = nil
	}

	query := `
		INSERT INTO buses (
			bus_owner_id,
			permit_id,
			bus_number,
			license_plate,
			bus_type,
			seat_layout_id,
			status,
			has_wifi,
			has_ac,
			has_charging_ports,
			has_entertainment,
			has_refreshments
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	err = tx.QueryRow(query,
		busOwnerID,
		permitID,
		bus.BusNumber,
		bus.LicensePlate,
		bus.BusType,
		seatLayoutIDValue,
		bus.Status,
		// bus.BusinessPhone, // Removed contact as it doesn't exist in buses table
		false, // has_wifi
		false, // has_ac
		false, // has_charging_ports
		false, // has_entertainment
		false, // has_refreshments
	).Scan(&bus.ID)

	if err != nil {
		return fmt.Errorf("error creating bus: %v", err)
	}

	return tx.Commit()
}

// UpdateBus updates an existing bus record and its related entities
func (r *BusRepository) UpdateBus(bus *models.Bus) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Get existing foreign keys
	var busOwnerID, permitID string
	err = tx.QueryRow(`SELECT bus_owner_id, permit_id FROM buses WHERE id = $1`, bus.ID).Scan(&busOwnerID, &permitID)
	if err != nil {
		return fmt.Errorf("error finding bus: %v", err)
	}

	// 2. Update Bus Owner
	_, err = tx.Exec(`
		UPDATE bus_owners 
		SET company_name = $1, 
			identity_or_incorporation_no = $2, 
			business_email = $3, 
			business_phone = $4
		WHERE id = $5
	`, bus.CompanyName, bus.IdentifyOrIncorporationNo, bus.BusinessEmail, bus.BusinessPhone, busOwnerID)
	if err != nil {
		return fmt.Errorf("error updating bus owner: %v", err)
	}

	// 3. Handle Master Route (Find existing or Create new)
	var masterRouteID string
	err = tx.QueryRow(`SELECT id FROM master_routes WHERE route_name = $1`, bus.CustomRouteName).Scan(&masterRouteID)
	if err == sql.ErrNoRows {
		// Create new route
		err = tx.QueryRow(`
			INSERT INTO master_routes (id, route_name, route_number, origin_city, destination_city)
			VALUES (uuid_generate_v4(), $1, 'TEMP-' || substring(md5(random()::text) from 1 for 6), 'Unknown', 'Unknown')
			RETURNING id
		`, bus.CustomRouteName).Scan(&masterRouteID)
		if err != nil {
			return fmt.Errorf("error creating master route: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking master route: %v", err)
	}

	// 4. Update Route Permit
	docsJSON, _ := json.Marshal(bus.VerificationDocuments)
	_, err = tx.Exec(`
		UPDATE route_permits 
		SET permit_number = $1, 
			approved_seating_capacity = $2, 
			approved_fare = $3, 
			master_route_id = $4,
			verification_documents = $5,
			status = $6
		WHERE id = $7
	`, bus.PermitNumber, bus.TotalSeats, bus.FarePerSeat, masterRouteID, docsJSON, strings.ToLower(bus.VerificationStatus), permitID)
	if err != nil {
		return fmt.Errorf("error updating route permit: %v", err)
	}

	// 5. Update Bus
	query := `
		UPDATE buses SET
			bus_number = $1,
			license_plate = $2,
			bus_type = $3,
			seat_layout_id = $4,
			status = $5,
			updated_at = NOW()
		WHERE id = $6
	`

	_, err = tx.Exec(query,
		bus.BusNumber,
		bus.LicensePlate,
		bus.BusType,
		bus.SeatLayoutID,
		bus.Status,
		bus.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating bus: %v", err)
	}

	return tx.Commit()
}

// DeleteBus deletes a bus record
func (r *BusRepository) DeleteBus(id string) error {
	query := `DELETE FROM buses WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting bus: %v", err)
	}
	return nil
}

// GetPendingBuses retrieves all buses with pending verification status
func (r *BusRepository) GetPendingBuses() ([]models.Bus, error) {
	query := `
		SELECT 
			b.id::text,
			COALESCE(b.bus_owner_id::text, ''),
			COALESCE(b.permit_id::text, ''),
			COALESCE(b.seat_layout_id::text, ''),
			COALESCE(b.bus_number, ''),
			COALESCE(bo.company_name, ''),
			COALESCE(bo.identity_or_incorporation_no, ''),
			COALESCE(bo.business_email, ''),
			COALESCE(bo.business_phone, ''),
			COALESCE(rp.permit_number, ''),
			COALESCE(b.license_plate, ''),
			COALESCE(bslt.total_seats, rp.approved_seating_capacity, 0) as total_seats,
			COALESCE(b.bus_type, 'Standard'),
			COALESCE(mr.route_name, 'Unknown Route') as route_name,
			COALESCE(rp.approved_fare::float8, 0.0),
			COALESCE(b.status, 'Inactive'),
			COALESCE(rp.status::text, 'Pending') as verification_status,
		COALESCE(NULLIF(bo.verification_status::text, ''), 'Pending') as owner_verification_status,
			COALESCE(rp.verification_documents::text, '{}'),
			COALESCE(b.created_at::text, '')
		FROM buses b
		LEFT JOIN bus_owners bo ON b.bus_owner_id = bo.id
		LEFT JOIN route_permits rp ON b.permit_id = rp.id
		LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
		LEFT JOIN bus_seat_layout_templates bslt ON b.seat_layout_id = bslt.id
		WHERE rp.status = 'pending'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying pending buses: %v", err)
	}
	defer rows.Close()

	buses := []models.Bus{}
	for rows.Next() {
		var bus models.Bus
		var verificationDocsStr sql.NullString
		var seatLayoutID sql.NullString

		err := rows.Scan(
			&bus.ID,
			&bus.BusOwnerID,
			&bus.PermitID,
			&seatLayoutID,
			&bus.BusNumber,
			&bus.CompanyName,
			&bus.IdentifyOrIncorporationNo,
			&bus.BusinessEmail,
			&bus.BusinessPhone,
			&bus.PermitNumber,
			&bus.LicensePlate,
			&bus.TotalSeats,
			&bus.BusType,
			&bus.CustomRouteName,
			&bus.FarePerSeat,
			&bus.Status,
			&bus.VerificationStatus,
			&bus.OwnerVerificationStatus,
			&verificationDocsStr,
			&bus.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning bus row: %v", err)
		}

		bus.SeatLayoutID = seatLayoutID

		if verificationDocsStr.Valid && verificationDocsStr.String != "" {
			if err := json.Unmarshal([]byte(verificationDocsStr.String), &bus.VerificationDocuments); err != nil {
				bus.VerificationDocuments = []string{}
			}
		}

		buses = append(buses, bus)
	}

	return buses, nil
}

// UpdateBusVerification updates the verification status of a bus (only updates route_permits, not bus_owners)
func (r *BusRepository) UpdateBusVerification(id string, status string, documents string) error {
	// Use transaction to update only route_permits
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Convert status to lowercase to match database enum (pending, verified, rejected)
	status = strings.ToLower(status)
	fmt.Printf("UpdateBusVerification: bus_id=%s, status=%s, documents=%s\n", id, status, documents)

	// Get permit_id from the bus
	var permitID string
	err = tx.QueryRow(`SELECT permit_id FROM buses WHERE id = $1`, id).Scan(&permitID)
	if err != nil {
		return fmt.Errorf("error getting bus permit_id: %v", err)
	}
	fmt.Printf("Found: permit_id=%s\n", permitID)

	// Prepare documents JSON array
	var docsJSON string
	if documents != "" {
		// Convert single document string to JSON array
		docsArray := []string{documents}
		docsBytes, _ := json.Marshal(docsArray)
		docsJSON = string(docsBytes)
	} else {
		docsJSON = "[]"
	}

	// Update only route_permits status and documents
	result, err := tx.Exec(`UPDATE route_permits SET status = $1, verification_documents = $2 WHERE id = $3`, status, docsJSON, permitID)
	if err != nil {
		return fmt.Errorf("error updating route permit verification status: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated route_permits: %d rows\n", rowsAffected)

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	fmt.Printf("✓ Successfully committed route_permits update\n")
	return nil
}
