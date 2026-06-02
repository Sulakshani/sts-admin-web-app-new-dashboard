package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sts-backend/internal/models"
)

type StaffRepository struct {
	db *sql.DB
}

func NewStaffRepository(db *sql.DB) *StaffRepository {
	return &StaffRepository{db: db}
}

// GetDrivers retrieves all drivers
func (r *StaffRepository) GetDrivers() ([]models.Driver, error) {
	query := `
		SELECT 
			bs.id::text,
			COALESCE(bs.emergency_contact_name, ''),
			COALESCE(bs.emergency_contact, ''),
			COALESCE(bs.license_number, ''),
			COALESCE(bs.license_expiry_date::text, ''),
			COALESCE(bs.experience_years, 0),
			COALESCE(bs.verification_status::text, 'Pending'),
			COALESCE(bs.verification_notes, ''),
			COALESCE(bse.employment_status::text, 'Active'),
			COALESCE(bse.hire_date::text, '')
		FROM bus_staff bs
		LEFT JOIN bus_staff_employment bse ON bs.id = bse.staff_id
		WHERE bs.staff_type = 'driver'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying drivers: %v", err)
	}
	defer rows.Close()

	drivers := []models.Driver{}
	for rows.Next() {
		var d models.Driver
		err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.ContactNumber,
			&d.LicenseNumber,
			&d.LicenseExpiryDate,
			&d.ExperienceYears,
			&d.VerificationStatus,
			&d.VerificationNotes,
			&d.Status,
			&d.HireDate,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning driver: %v", err)
		}
		drivers = append(drivers, d)
	}

	return drivers, nil
}

// GetPendingDrivers retrieves drivers with pending verification
func (r *StaffRepository) GetPendingDrivers() ([]models.Driver, error) {
	query := `
		SELECT 
			bs.id::text,
			COALESCE(bs.emergency_contact_name, ''),
			COALESCE(bs.emergency_contact, ''),
			COALESCE(bs.license_number, ''),
			COALESCE(bs.license_expiry_date::text, ''),
			COALESCE(bs.experience_years, 0),
			COALESCE(bs.verification_status::text, 'Pending'),
			COALESCE(bs.verification_notes, ''),
			COALESCE(bse.employment_status::text, 'Active'),
			COALESCE(bse.hire_date::text, ''),
			COALESCE(bs.created_at::text, '')
		FROM bus_staff bs
		LEFT JOIN bus_staff_employment bse ON bs.id = bse.staff_id
		WHERE bs.staff_type = 'driver' AND LOWER(bs.verification_status::text) = 'pending'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying pending drivers: %v", err)
	}
	defer rows.Close()

	drivers := []models.Driver{}
	for rows.Next() {
		var d models.Driver
		err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.ContactNumber,
			&d.LicenseNumber,
			&d.LicenseExpiryDate,
			&d.ExperienceYears,
			&d.VerificationStatus,
			&d.VerificationNotes,
			&d.Status,
			&d.HireDate,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning pending driver: %v", err)
		}
		drivers = append(drivers, d)
	}

	return drivers, nil
}

// GetDriverByID retrieves a single driver by ID
func (r *StaffRepository) GetDriverByID(id string) (*models.Driver, error) {
	query := `
		SELECT 
			bs.id::text,
			COALESCE(bs.emergency_contact_name, ''),
			COALESCE(bs.emergency_contact, ''),
			COALESCE(bs.license_number, ''),
			COALESCE(bs.license_expiry_date::text, ''),
			COALESCE(bs.experience_years, 0),
			COALESCE(bs.verification_status::text, 'Pending'),
			COALESCE(bs.verification_notes, ''),
			COALESCE(bse.employment_status::text, 'Active'),
			COALESCE(bse.hire_date::text, '')
		FROM bus_staff bs
		LEFT JOIN bus_staff_employment bse ON bs.id = bse.staff_id
		WHERE bs.id = $1 AND bs.staff_type = 'driver'
	`

	var d models.Driver
	err := r.db.QueryRow(query, id).Scan(
		&d.ID,
		&d.Name,
		&d.ContactNumber,
		&d.LicenseNumber,
		&d.LicenseExpiryDate,
		&d.ExperienceYears,
		&d.VerificationStatus,
		&d.VerificationNotes,
		&d.Status,
		&d.HireDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying driver by ID: %v", err)
	}
	return &d, nil
}

// UpdateDriverVerification updates the verification status of a driver
func (r *StaffRepository) UpdateDriverVerification(id string, status string, documents string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Map 'verified' to 'approved' to match database enum values
	isVerified := false
	if strings.ToLower(status) == "verified" {
		status = "approved"
		isVerified = true
	} else {
		status = strings.ToLower(status)
	}

	// Update bus_staff table with both verification_status and is_verified
	_, err = tx.Exec(`
		UPDATE bus_staff 
		SET verification_status = $1, verification_notes = $2, is_verified = $3 
		WHERE id = $4 AND staff_type = 'driver'
	`, status, documents, isVerified, id)
	if err != nil {
		return fmt.Errorf("error updating driver verification: %v", err)
	}

	return tx.Commit()
}

// CreateDriver creates a new driver
func (r *StaffRepository) CreateDriver(d *models.Driver) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Find an available user_id (user without existing bus_staff record)
	var userID string
	err = tx.QueryRow(`
		SELECT u.id FROM users u
		LEFT JOIN bus_staff bs ON u.id = bs.user_id
		WHERE bs.id IS NULL
		LIMIT 1
	`).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no available users found. Please create a user account first")
		}
		return fmt.Errorf("error finding available user: %v", err)
	}

	// Insert into bus_staff
	var staffID string
	err = tx.QueryRow(`
		INSERT INTO bus_staff (
			user_id,
			staff_type,
			emergency_contact_name,
			emergency_contact,
			license_number,
			license_expiry_date,
			experience_years,
			verification_status,
			verification_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, userID, "driver", d.Name, d.ContactNumber, d.LicenseNumber, d.LicenseExpiryDate, d.ExperienceYears, d.VerificationStatus, d.VerificationNotes).Scan(&staffID)

	if err != nil {
		return fmt.Errorf("error inserting driver into bus_staff: %v", err)
	}

	d.ID = staffID

	// Find an available bus_owner_id
	var busOwnerID string
	err = tx.QueryRow(`
		SELECT id FROM bus_owners
		ORDER BY id
		LIMIT 1
	`).Scan(&busOwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no bus owners found. Please create a bus first")
		}
		return fmt.Errorf("error finding bus owner: %v", err)
	}

	// Insert into bus_staff_employment
	_, err = tx.Exec(`
		INSERT INTO bus_staff_employment (
			staff_id,
			bus_owner_id,
			employment_status,
			hire_date
		) VALUES ($1, $2, $3, $4)
	`, staffID, busOwnerID, strings.ToLower(d.Status), d.HireDate)

	if err != nil {
		return fmt.Errorf("error inserting driver employment: %v", err)
	}

	return tx.Commit()
}

// UpdateDriver updates an existing driver
func (r *StaffRepository) UpdateDriver(d *models.Driver) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Update bus_staff
	_, err = tx.Exec(`
		UPDATE bus_staff SET
			emergency_contact_name = $1,
			emergency_contact = $2,
			license_number = $3,
			license_expiry_date = $4,
			experience_years = $5,
			verification_status = $6,
			verification_notes = $7
		WHERE id = $8 AND staff_type = 'driver'
	`, d.Name, d.ContactNumber, d.LicenseNumber, d.LicenseExpiryDate, d.ExperienceYears, strings.ToLower(d.VerificationStatus), d.VerificationNotes, d.ID)

	if err != nil {
		return fmt.Errorf("error updating driver in bus_staff: %v", err)
	}

	// Update bus_staff_employment
	// Check if employment record exists
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM bus_staff_employment WHERE staff_id = $1)`, d.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking employment record: %v", err)
	}

	if exists {
		_, err = tx.Exec(`
			UPDATE bus_staff_employment SET
				employment_status = $1,
				hire_date = $2
			WHERE staff_id = $3
		`, strings.ToLower(d.Status), d.HireDate, d.ID)
	} else {
		_, err = tx.Exec(`
			INSERT INTO bus_staff_employment (staff_id, employment_status, hire_date)
			VALUES ($1, $2, $3)
		`, d.ID, strings.ToLower(d.Status), d.HireDate)
	}

	if err != nil {
		return fmt.Errorf("error updating driver employment: %v", err)
	}

	return tx.Commit()
}

// GetConductors retrieves all conductors
func (r *StaffRepository) GetConductors() ([]models.Conductor, error) {
	query := `
		SELECT 
			bs.id::text,
			COALESCE(bs.emergency_contact_name, ''),
			COALESCE(bs.emergency_contact, ''),
			COALESCE(bs.license_number, ''),
			COALESCE(bs.license_expiry_date::text, ''),
			COALESCE(bs.experience_years, 0),
			COALESCE(bs.verification_status::text, 'Pending'),
			COALESCE(bs.verification_notes, ''),
			COALESCE(bse.employment_status::text, 'Active'),
			COALESCE(bse.hire_date::text, '')
		FROM bus_staff bs
		LEFT JOIN bus_staff_employment bse ON bs.id = bse.staff_id
		WHERE bs.staff_type = 'conductor'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying conductors: %v", err)
	}
	defer rows.Close()

	conductors := []models.Conductor{}
	for rows.Next() {
		var c models.Conductor
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.ContactNumber,
			&c.LicenseNumber,
			&c.LicenseExpiryDate,
			&c.ExperienceYears,
			&c.VerificationStatus,
			&c.VerificationNotes,
			&c.Status,
			&c.HireDate,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning conductor: %v", err)
		}
		conductors = append(conductors, c)
	}

	return conductors, nil
}

// GetPendingConductors retrieves conductors with pending verification
func (r *StaffRepository) GetPendingConductors() ([]models.Conductor, error) {
	query := `
		SELECT 
			bs.id::text,
			COALESCE(bs.emergency_contact_name, ''),
			COALESCE(bs.emergency_contact, ''),
			COALESCE(bs.license_number, ''),
			COALESCE(bs.license_expiry_date::text, ''),
			COALESCE(bs.experience_years, 0),
			COALESCE(bs.verification_status::text, 'Pending'),
			COALESCE(bs.verification_notes, ''),
			COALESCE(bse.employment_status::text, 'Active'),
			COALESCE(bse.hire_date::text, ''),
			COALESCE(bs.created_at::text, '')
		FROM bus_staff bs
		LEFT JOIN bus_staff_employment bse ON bs.id = bse.staff_id
		WHERE bs.staff_type = 'conductor' AND LOWER(bs.verification_status::text) = 'pending'
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying pending conductors: %v", err)
	}
	defer rows.Close()

	conductors := []models.Conductor{}
	for rows.Next() {
		var c models.Conductor
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.ContactNumber,
			&c.LicenseNumber,
			&c.LicenseExpiryDate,
			&c.ExperienceYears,
			&c.VerificationStatus,
			&c.VerificationNotes,
			&c.Status,
			&c.HireDate,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning pending conductor: %v", err)
		}
		conductors = append(conductors, c)
	}

	return conductors, nil
}

// GetConductorByID retrieves a single conductor by ID
func (r *StaffRepository) GetConductorByID(id string) (*models.Conductor, error) {
	query := `
		SELECT 
			bs.id::text,
			COALESCE(bs.emergency_contact_name, ''),
			COALESCE(bs.emergency_contact, ''),
			COALESCE(bs.license_number, ''),
			COALESCE(bs.license_expiry_date::text, ''),
			COALESCE(bs.experience_years, 0),
			COALESCE(bs.verification_status::text, 'Pending'),
			COALESCE(bs.verification_notes, ''),
			COALESCE(bse.employment_status::text, 'Active'),
			COALESCE(bse.hire_date::text, '')
		FROM bus_staff bs
		LEFT JOIN bus_staff_employment bse ON bs.id = bse.staff_id
		WHERE bs.id = $1 AND bs.staff_type = 'conductor'
	`

	var c models.Conductor
	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.Name,
		&c.ContactNumber,
		&c.LicenseNumber,
		&c.LicenseExpiryDate,
		&c.ExperienceYears,
		&c.VerificationStatus,
		&c.VerificationNotes,
		&c.Status,
		&c.HireDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying conductor by ID: %v", err)
	}
	return &c, nil
}

// UpdateConductorVerification updates the verification status of a conductor
func (r *StaffRepository) UpdateConductorVerification(id string, status string, documents string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Map 'verified' to 'approved' to match database enum values
	isVerified := false
	if strings.ToLower(status) == "verified" {
		status = "approved"
		isVerified = true
	} else {
		status = strings.ToLower(status)
	}

	// Update bus_staff table with both verification_status and is_verified
	_, err = tx.Exec(`
		UPDATE bus_staff 
		SET verification_status = $1, verification_notes = $2, is_verified = $3 
		WHERE id = $4 AND staff_type = 'conductor'
	`, status, documents, isVerified, id)
	if err != nil {
		return fmt.Errorf("error updating conductor verification: %v", err)
	}

	return tx.Commit()
}

// CreateConductor creates a new conductor
func (r *StaffRepository) CreateConductor(c *models.Conductor) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Find an available user_id (user without existing bus_staff record)
	var userID string
	err = tx.QueryRow(`
		SELECT u.id FROM users u
		LEFT JOIN bus_staff bs ON u.id = bs.user_id
		WHERE bs.id IS NULL
		LIMIT 1
	`).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no available users found. Please create a user account first")
		}
		return fmt.Errorf("error finding available user: %v", err)
	}

	// Insert into bus_staff
	var staffID string
	err = tx.QueryRow(`
		INSERT INTO bus_staff (
			user_id,
			staff_type,
			emergency_contact_name,
			emergency_contact,
			license_number,
			license_expiry_date,
			experience_years,
			verification_status,
			verification_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, userID, "conductor", c.Name, c.ContactNumber, c.LicenseNumber, c.LicenseExpiryDate, c.ExperienceYears, c.VerificationStatus, c.VerificationNotes).Scan(&staffID)

	if err != nil {
		return fmt.Errorf("error inserting conductor into bus_staff: %v", err)
	}

	c.ID = staffID

	// Find an available bus_owner_id
	var busOwnerID string
	err = tx.QueryRow(`
		SELECT id FROM bus_owners
		ORDER BY id
		LIMIT 1
	`).Scan(&busOwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no bus owners found. Please create a bus first")
		}
		return fmt.Errorf("error finding bus owner: %v", err)
	}

	// Insert into bus_staff_employment
	_, err = tx.Exec(`
		INSERT INTO bus_staff_employment (
			staff_id,
			bus_owner_id,
			employment_status,
			hire_date
		) VALUES ($1, $2, $3, $4)
	`, staffID, busOwnerID, strings.ToLower(c.Status), c.HireDate)

	if err != nil {
		return fmt.Errorf("error inserting conductor employment: %v", err)
	}

	return tx.Commit()
}

// UpdateConductor updates an existing conductor
func (r *StaffRepository) UpdateConductor(c *models.Conductor) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Update bus_staff
	_, err = tx.Exec(`
		UPDATE bus_staff SET
			emergency_contact_name = $1,
			emergency_contact = $2,
			license_number = $3,
			license_expiry_date = $4,
			experience_years = $5,
			verification_status = $6,
			verification_notes = $7
		WHERE id = $8 AND staff_type = 'conductor'
	`, c.Name, c.ContactNumber, c.LicenseNumber, c.LicenseExpiryDate, c.ExperienceYears, strings.ToLower(c.VerificationStatus), c.VerificationNotes, c.ID)

	if err != nil {
		return fmt.Errorf("error updating conductor in bus_staff: %v", err)
	}

	// Update bus_staff_employment
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM bus_staff_employment WHERE staff_id = $1)`, c.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking employment record: %v", err)
	}

	if exists {
		_, err = tx.Exec(`
			UPDATE bus_staff_employment SET
				employment_status = $1,
				hire_date = $2
			WHERE staff_id = $3
		`, strings.ToLower(c.Status), c.HireDate, c.ID)
	} else {
		_, err = tx.Exec(`
			INSERT INTO bus_staff_employment (staff_id, employment_status, hire_date)
			VALUES ($1, $2, $3)
		`, c.ID, strings.ToLower(c.Status), c.HireDate)
	}

	if err != nil {
		return fmt.Errorf("error updating conductor employment: %v", err)
	}

	return tx.Commit()
}
