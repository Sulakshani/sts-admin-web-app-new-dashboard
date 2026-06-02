package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sts-backend/internal/models"
)

func GetAllBusOwners() ([]models.BusOwner, error) {
	query := `
		SELECT 
			id,
			COALESCE(user_id::text, '') as user_id,
			COALESCE(company_name, '') as company_name,
			COALESCE(license_number, '') as license_number,
			COALESCE(contact_person, '') as contact_person,
			COALESCE(address, '') as address,
			COALESCE(city, '') as city,
			COALESCE(state, '') as state,
			COALESCE(country, 'Sri Lanka') as country,
			COALESCE(postal_code, '') as postal_code,
			verification_status::text,
			verification_documents,
			COALESCE(business_email, '') as business_email,
			COALESCE(business_phone, '') as business_phone,
			COALESCE(tax_id, '') as tax_id,
			bank_account_details,
			COALESCE(total_buses, 0) as total_buses,
			COALESCE(profile_completed, false) as profile_completed,
			COALESCE(identity_or_incorporation_no, '') as identity_or_incorporation_no,
			created_at,
			updated_at
		FROM bus_owners
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying bus owners: %w", err)
	}
	defer rows.Close()

	var owners []models.BusOwner
	for rows.Next() {
		var owner models.BusOwner
		var verificationDocs, bankAccountDetails []byte

		err := rows.Scan(
			&owner.ID,
			&owner.UserID,
			&owner.CompanyName,
			&owner.LicenseNumber,
			&owner.ContactPerson,
			&owner.Address,
			&owner.City,
			&owner.State,
			&owner.Country,
			&owner.PostalCode,
			&owner.VerificationStatus,
			&verificationDocs,
			&owner.BusinessEmail,
			&owner.BusinessPhone,
			&owner.TaxID,
			&bankAccountDetails,
			&owner.TotalBuses,
			&owner.ProfileCompleted,
			&owner.IdentityOrIncorporationNo,
			&owner.CreatedAt,
			&owner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning bus owner: %w", err)
		}

		// Parse JSONB fields
		if verificationDocs != nil {
			json.Unmarshal(verificationDocs, &owner.VerificationDocuments)
		}
		if bankAccountDetails != nil {
			json.Unmarshal(bankAccountDetails, &owner.BankAccountDetails)
		}

		owners = append(owners, owner)
	}

	if owners == nil {
		owners = []models.BusOwner{}
	}

	return owners, nil
}

func GetBusOwnerByID(id string) (*models.BusOwner, error) {
	query := `
		SELECT 
			id,
			COALESCE(user_id::text, '') as user_id,
			COALESCE(company_name, '') as company_name,
			COALESCE(license_number, '') as license_number,
			COALESCE(contact_person, '') as contact_person,
			COALESCE(address, '') as address,
			COALESCE(city, '') as city,
			COALESCE(state, '') as state,
			COALESCE(country, 'Sri Lanka') as country,
			COALESCE(postal_code, '') as postal_code,
			verification_status::text,
			verification_documents,
			COALESCE(business_email, '') as business_email,
			COALESCE(business_phone, '') as business_phone,
			COALESCE(tax_id, '') as tax_id,
			bank_account_details,
			COALESCE(total_buses, 0) as total_buses,
			COALESCE(profile_completed, false) as profile_completed,
			COALESCE(identity_or_incorporation_no, '') as identity_or_incorporation_no,
			created_at,
			updated_at
		FROM bus_owners
		WHERE id = $1
	`

	var owner models.BusOwner
	var verificationDocs, bankAccountDetails []byte

	err := DB.QueryRow(query, id).Scan(
		&owner.ID,
		&owner.UserID,
		&owner.CompanyName,
		&owner.LicenseNumber,
		&owner.ContactPerson,
		&owner.Address,
		&owner.City,
		&owner.State,
		&owner.Country,
		&owner.PostalCode,
		&owner.VerificationStatus,
		&verificationDocs,
		&owner.BusinessEmail,
		&owner.BusinessPhone,
		&owner.TaxID,
		&bankAccountDetails,
		&owner.TotalBuses,
		&owner.ProfileCompleted,
		&owner.IdentityOrIncorporationNo,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting bus owner: %w", err)
	}

	// Parse JSONB fields
	if verificationDocs != nil {
		json.Unmarshal(verificationDocs, &owner.VerificationDocuments)
	}
	if bankAccountDetails != nil {
		json.Unmarshal(bankAccountDetails, &owner.BankAccountDetails)
	}

	return &owner, nil
}

func GetPendingBusOwners() ([]models.BusOwner, error) {
	query := `
		SELECT 
			id,
			COALESCE(user_id::text, '') as user_id,
			COALESCE(company_name, '') as company_name,
			COALESCE(license_number, '') as license_number,
			COALESCE(contact_person, '') as contact_person,
			COALESCE(address, '') as address,
			COALESCE(city, '') as city,
			COALESCE(state, '') as state,
			COALESCE(country, 'Sri Lanka') as country,
			COALESCE(postal_code, '') as postal_code,
			verification_status::text,
			verification_documents,
			COALESCE(business_email, '') as business_email,
			COALESCE(business_phone, '') as business_phone,
			COALESCE(tax_id, '') as tax_id,
			bank_account_details,
			COALESCE(total_buses, 0) as total_buses,
			COALESCE(profile_completed, false) as profile_completed,
			COALESCE(identity_or_incorporation_no, '') as identity_or_incorporation_no,
			created_at,
			updated_at
		FROM bus_owners
		WHERE verification_status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying pending bus owners: %w", err)
	}
	defer rows.Close()

	var owners []models.BusOwner
	for rows.Next() {
		var owner models.BusOwner
		var verificationDocs, bankAccountDetails []byte

		err := rows.Scan(
			&owner.ID,
			&owner.UserID,
			&owner.CompanyName,
			&owner.LicenseNumber,
			&owner.ContactPerson,
			&owner.Address,
			&owner.City,
			&owner.State,
			&owner.Country,
			&owner.PostalCode,
			&owner.VerificationStatus,
			&verificationDocs,
			&owner.BusinessEmail,
			&owner.BusinessPhone,
			&owner.TaxID,
			&bankAccountDetails,
			&owner.TotalBuses,
			&owner.ProfileCompleted,
			&owner.IdentityOrIncorporationNo,
			&owner.CreatedAt,
			&owner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning bus owner: %w", err)
		}

		// Parse JSONB fields
		if verificationDocs != nil {
			json.Unmarshal(verificationDocs, &owner.VerificationDocuments)
		}
		if bankAccountDetails != nil {
			json.Unmarshal(bankAccountDetails, &owner.BankAccountDetails)
		}

		owners = append(owners, owner)
	}

	if owners == nil {
		owners = []models.BusOwner{}
	}

	return owners, nil
}

func CreateBusOwner(owner *models.BusOwner) error {
	query := `
		INSERT INTO bus_owners (
			user_id,
			company_name,
			business_email,
			business_phone,
			identity_or_incorporation_no,
			verification_status,
			verification_documents,
			country,
			profile_completed,
			total_buses
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	verificationDocsJSON, _ := json.Marshal(owner.VerificationDocuments)

	err := DB.QueryRow(
		query,
		owner.UserID,
		owner.CompanyName,
		owner.BusinessEmail,
		owner.BusinessPhone,
		owner.IdentityOrIncorporationNo,
		owner.VerificationStatus,
		verificationDocsJSON,
		owner.Country,
		owner.ProfileCompleted,
		owner.TotalBuses,
	).Scan(&owner.ID, &owner.CreatedAt, &owner.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating bus owner: %w", err)
	}

	return nil
}

func UpdateBusOwner(owner *models.BusOwner) error {
	query := `
		UPDATE bus_owners
		SET company_name = $1,
			business_email = $2,
			business_phone = $3,
			identity_or_incorporation_no = $4,
			updated_at = now()
		WHERE id = $5
		RETURNING id, COALESCE(user_id::text, '') as user_id, company_name, business_email, business_phone, 
			identity_or_incorporation_no, verification_status, 
			COALESCE(license_number, '') as license_number,
			COALESCE(contact_person, '') as contact_person,
			COALESCE(address, '') as address,
			COALESCE(city, '') as city,
			COALESCE(state, '') as state,
			COALESCE(country, 'Sri Lanka') as country,
			COALESCE(postal_code, '') as postal_code,
			COALESCE(tax_id, '') as tax_id,
			COALESCE(total_buses, 0) as total_buses,
			COALESCE(profile_completed, false) as profile_completed,
			created_at, updated_at
	`

	err := DB.QueryRow(
		query,
		owner.CompanyName,
		owner.BusinessEmail,
		owner.BusinessPhone,
		owner.IdentityOrIncorporationNo,
		owner.ID,
	).Scan(
		&owner.ID,
		&owner.UserID,
		&owner.CompanyName,
		&owner.BusinessEmail,
		&owner.BusinessPhone,
		&owner.IdentityOrIncorporationNo,
		&owner.VerificationStatus,
		&owner.LicenseNumber,
		&owner.ContactPerson,
		&owner.Address,
		&owner.City,
		&owner.State,
		&owner.Country,
		&owner.PostalCode,
		&owner.TaxID,
		&owner.TotalBuses,
		&owner.ProfileCompleted,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return fmt.Errorf("bus owner not found")
	}

	if err != nil {
		return fmt.Errorf("error updating bus owner: %w", err)
	}

	return nil
}

func DeleteBusOwner(id string) error {
	query := `DELETE FROM bus_owners WHERE id = $1`

	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting bus owner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bus owner not found")
	}

	return nil
}

func VerifyBusOwner(id string, status string, documents interface{}) error {
	query := `
		UPDATE bus_owners
		SET verification_status = $1,
			verification_documents = $2,
			updated_at = now()
		WHERE id = $3
		RETURNING id
	`

	var verificationDocsJSON []byte
	if documents != nil {
		verificationDocsJSON, _ = json.Marshal(documents)
	} else {
		verificationDocsJSON = []byte("null")
	}

	var returnedID string
	err := DB.QueryRow(query, status, verificationDocsJSON, id).Scan(&returnedID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("bus owner not found")
	}
	if err != nil {
		return fmt.Errorf("error verifying bus owner: %w", err)
	}

	return nil
}
