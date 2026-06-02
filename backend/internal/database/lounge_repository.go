package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sts-backend/internal/models"
)

type LoungeRepository struct {
	db *sql.DB
}

func NewLoungeRepository(db *sql.DB) *LoungeRepository {
	return &LoungeRepository{db: db}
}

func (r *LoungeRepository) GetLounges() ([]models.Lounge, error) {
	query := `
		SELECT 
			l.id::text,
			COALESCE(lo.manager_full_name, ''),
			COALESCE(lo.nic, ''),
			COALESCE(lo.email, ''),
			COALESCE(lo.contact_number, ''),
			COALESCE(l.lounge_name, ''),
			COALESCE(l.contact_phone, ''),
			COALESCE(l.address, ''),
			COALESCE(l.capacity, 0),
			COALESCE(l.price_1_hour, 0),
			COALESCE(l.amenities::text, '[]'),
			COALESCE(lmc.name, ''),
			COALESCE(l.status::text, 'Pending'),
			COALESCE(l.verification_note, ''),
			COALESCE(l.is_operational, true)
		FROM lounges l
		LEFT JOIN lounge_owners lo ON l.lounge_owner_id = lo.id
		LEFT JOIN lounge_marketplace_categories lmc ON l.marketplace_category_id = lmc.id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying lounges: %v", err)
	}
	defer rows.Close()

	lounges := []models.Lounge{}
	for rows.Next() {
		var l models.Lounge
		var amenitiesJSON string
		err := rows.Scan(
			&l.LoungeID,
			&l.LoungeOwner,
			&l.OwnerNIC,
			&l.OwnerEmail,
			&l.OwnerContact,
			&l.LoungeName,
			&l.LoungeContact,
			&l.Address,
			&l.Capacity,
			&l.PricePerHour,
			&amenitiesJSON,
			&l.Marketplace,
			&l.Verification,
			&l.VerificationNote,
			&l.Operational,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning lounge: %v", err)
		}
		// Parse JSON amenities
		if err := json.Unmarshal([]byte(amenitiesJSON), &l.Facilities); err != nil {
			l.Facilities = []string{} // Default to empty array on error
		}
		lounges = append(lounges, l)
	}

	return lounges, nil
}

func (r *LoungeRepository) CreateLounge(l models.Lounge) error {
	// Use a transaction to ensure all operations complete successfully
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Find an available user_id (user without existing lounge_owner record)
	var userID string
	err = tx.QueryRow(`
		SELECT u.id FROM users u
		LEFT JOIN lounge_owners lo ON u.id = lo.user_id
		WHERE lo.id IS NULL
		LIMIT 1
	`).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no available users found. Please create a user account first")
		}
		return fmt.Errorf("error finding available user: %v", err)
	}

	// 2. Create Owner with user_id
	var ownerID string
	err = tx.QueryRow(`
		INSERT INTO lounge_owners (user_id, manager_full_name, email, contact_number, nic, verification_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id
	`, userID, l.LoungeOwner, l.OwnerEmail, l.OwnerContact, l.OwnerNIC).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("error creating owner: %v", err)
	}

	// 3. Resolve Marketplace Category ID
	var marketplaceID sql.NullString
	if l.Marketplace != "" {
		err = tx.QueryRow("SELECT id FROM lounge_marketplace_categories WHERE name = $1", l.Marketplace).Scan(&marketplaceID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow("INSERT INTO lounge_marketplace_categories (name) VALUES ($1) RETURNING id", l.Marketplace).Scan(&marketplaceID)
		}
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error resolving marketplace category: %v", err)
		}
	}

	// 4. Create Lounge - use lowercase 'pending' for status enum
	// Convert facilities to JSON for jsonb column
	facilitiesJSON, err := json.Marshal(l.Facilities)
	if err != nil {
		return fmt.Errorf("error marshaling facilities: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO lounges (
			lounge_owner_id, lounge_name, contact_phone, address, capacity, price_1_hour, 
			is_operational, amenities, status, marketplace_category_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, ownerID, l.LoungeName, l.LoungeContact, l.Address, l.Capacity, l.PricePerHour,
		l.Operational, facilitiesJSON, strings.ToLower(l.Verification), marketplaceID)

	if err != nil {
		return fmt.Errorf("error creating lounge: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (r *LoungeRepository) UpdateLounge(l models.Lounge) error {
	// Resolve Marketplace Category ID
	var marketplaceID sql.NullString
	if l.Marketplace != "" {
		err := r.db.QueryRow("SELECT id FROM lounge_marketplace_categories WHERE name = $1", l.Marketplace).Scan(&marketplaceID)
		if err == sql.ErrNoRows {
			err = r.db.QueryRow("INSERT INTO lounge_marketplace_categories (name) VALUES ($1) RETURNING id", l.Marketplace).Scan(&marketplaceID)
		}
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error resolving marketplace category: %v", err)
		}
	}

	// Convert facilities to JSON for jsonb column
	facilitiesJSON, err := json.Marshal(l.Facilities)
	if err != nil {
		return fmt.Errorf("error marshaling facilities: %v", err)
	}

	// Update Lounge details - use lowercase for verification status
	_, err = r.db.Exec(`
		UPDATE lounges SET 
			lounge_name = $1, contact_phone = $2, address = $3, capacity = $4, 
			price_1_hour = $5, is_operational = $6, amenities = $7,
			status = $8, marketplace_category_id = $9, verification_note = $10
		WHERE id = $11
	`, l.LoungeName, l.LoungeContact, l.Address, l.Capacity, l.PricePerHour,
		l.Operational, facilitiesJSON, strings.ToLower(l.Verification), marketplaceID, l.VerificationNote, l.LoungeID)
	if err != nil {
		return fmt.Errorf("error updating lounge: %v", err)
	}

	// Update Owner details
	_, err = r.db.Exec(`
		UPDATE lounge_owners SET
			manager_full_name = $1, email = $2, contact_number = $3, nic = $4
		WHERE id = (SELECT lounge_owner_id FROM lounges WHERE id = $5)
	`, l.LoungeOwner, l.OwnerEmail, l.OwnerContact, l.OwnerNIC, l.LoungeID)

	if err != nil {
		return fmt.Errorf("error updating lounge owner: %v", err)
	}

	return nil
}

func (r *LoungeRepository) DeleteLounge(id string) error {
	_, err := r.db.Exec("DELETE FROM lounges WHERE id = $1", id)
	return err
}

// GetPendingLounges retrieves lounges with pending verification
func (r *LoungeRepository) GetPendingLounges() ([]models.Lounge, error) {
	query := `
		SELECT 
			l.id::text,
			COALESCE(lo.manager_full_name, ''),
			COALESCE(lo.nic, ''),
			COALESCE(lo.email, ''),
			COALESCE(lo.contact_number, ''),
			COALESCE(l.lounge_name, ''),
			COALESCE(l.contact_phone, ''),
			COALESCE(l.address, ''),
			COALESCE(l.capacity, 0),
			COALESCE(l.price_1_hour, 0),
			COALESCE(l.amenities::text, '[]'),
			COALESCE(lmc.name, ''),
			COALESCE(l.status::text, 'Pending'),
			COALESCE(l.verification_note, ''),
			COALESCE(l.is_operational, true),
			COALESCE(l.created_at::text, '')
		FROM lounges l
		LEFT JOIN lounge_owners lo ON l.lounge_owner_id = lo.id
		LEFT JOIN lounge_marketplace_categories lmc ON l.marketplace_category_id = lmc.id
		WHERE LOWER(l.status::text) = 'pending'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying pending lounges: %v", err)
	}
	defer rows.Close()

	lounges := []models.Lounge{}
	for rows.Next() {
		var l models.Lounge
		var amenitiesJSON string
		err := rows.Scan(
			&l.LoungeID,
			&l.LoungeOwner,
			&l.OwnerNIC,
			&l.OwnerEmail,
			&l.OwnerContact,
			&l.LoungeName,
			&l.LoungeContact,
			&l.Address,
			&l.Capacity,
			&l.PricePerHour,
			&amenitiesJSON,
			&l.Marketplace,
			&l.Verification,
			&l.VerificationNote,
			&l.Operational,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning pending lounge: %v", err)
		}
		// Parse JSON amenities
		if err := json.Unmarshal([]byte(amenitiesJSON), &l.Facilities); err != nil {
			l.Facilities = []string{} // Default to empty array on error
		}
		lounges = append(lounges, l)
	}

	return lounges, nil
}

// GetLoungeByID retrieves a single lounge by ID
func (r *LoungeRepository) GetLoungeByID(id string) (*models.Lounge, error) {
	query := `
		SELECT 
			l.id::text,
			COALESCE(lo.manager_full_name, ''),
			COALESCE(lo.nic, ''),
			COALESCE(lo.email, ''),
			COALESCE(lo.contact_number, ''),
			COALESCE(l.lounge_name, ''),
			COALESCE(l.contact_phone, ''),
			COALESCE(l.address, ''),
			COALESCE(l.capacity, 0),
			COALESCE(l.price_1_hour, 0),
			COALESCE(l.amenities::text, '[]'),
			COALESCE(lmc.name, ''),
			COALESCE(l.status::text, 'Pending'),
			COALESCE(l.verification_note, ''),
			COALESCE(l.is_operational, true)
		FROM lounges l
		LEFT JOIN lounge_owners lo ON l.lounge_owner_id = lo.id
		LEFT JOIN lounge_marketplace_categories lmc ON l.marketplace_category_id = lmc.id
		WHERE l.id = $1
	`

	var l models.Lounge
	var amenitiesJSON string
	err := r.db.QueryRow(query, id).Scan(
		&l.LoungeID,
		&l.LoungeOwner,
		&l.OwnerNIC,
		&l.OwnerEmail,
		&l.OwnerContact,
		&l.LoungeName,
		&l.LoungeContact,
		&l.Address,
		&l.Capacity,
		&l.PricePerHour,
		&amenitiesJSON,
		&l.Marketplace,
		&l.Verification,
		&l.VerificationNote,
		&l.Operational,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying lounge by ID: %v", err)
	}
	// Parse JSON amenities
	if err := json.Unmarshal([]byte(amenitiesJSON), &l.Facilities); err != nil {
		l.Facilities = []string{} // Default to empty array on error
	}
	return &l, nil
}

// UpdateLoungeVerification updates the verification status of a lounge and lounge owner
func (r *LoungeRepository) UpdateLoungeVerification(id string, status string, documents string) error {
	// Use transaction to update both tables atomically
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Map 'verified' to 'approved' to match database enum values
	if strings.ToLower(status) == "verified" {
		status = "approved"
	} else {
		status = strings.ToLower(status)
	}

	// Update lounges table
	_, err = tx.Exec(`
		UPDATE lounges 
		SET status = $1, verification_note = $2 
		WHERE id = $3
	`, status, documents, id)
	if err != nil {
		return fmt.Errorf("error updating lounge: %v", err)
	}

	// Update lounge_owners table verification_status
	_, err = tx.Exec(`
		UPDATE lounge_owners 
		SET verification_status = $1, verification_notes = $2 
		WHERE id = (SELECT lounge_owner_id FROM lounges WHERE id = $3)
	`, status, documents, id)
	if err != nil {
		return fmt.Errorf("error updating lounge owner: %v", err)
	}

	return tx.Commit()
}
