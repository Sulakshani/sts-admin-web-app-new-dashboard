package database

import (
	"database/sql"
	"fmt"
	"sts-backend/internal/models"
)

func GetPendingLoungeOwners() ([]models.LoungeOwner, error) {
	query := `
		SELECT
			lo.id::text,
			COALESCE(NULLIF(to_jsonb(lo)->>'user_id', ''), '') as user_id,
			COALESCE(NULLIF(to_jsonb(lo)->>'manager_full_name', ''), NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), '') as manager_full_name,
			COALESCE(NULLIF(to_jsonb(lo)->>'email', ''), NULLIF(to_jsonb(lo)->>'manager_email', ''), NULLIF(u.email, ''), '') as email,
			COALESCE(NULLIF(to_jsonb(lo)->>'contact_number', ''), NULLIF(u.phone, ''), '') as contact_number,
			COALESCE(NULLIF(to_jsonb(lo)->>'nic', ''), NULLIF(to_jsonb(lo)->>'manager_nic_number', ''), NULLIF(u.nic, ''), '') as nic,
			COALESCE(NULLIF(to_jsonb(lo)->>'business_name', ''), '') as business_name,
			COALESCE(NULLIF(to_jsonb(lo)->>'business_license', ''), '') as business_license,
			COALESCE(NULLIF(to_jsonb(lo)->>'verification_status', ''), 'pending') as verification_status,
			COALESCE(NULLIF(to_jsonb(lo)->>'verification_notes', ''), '') as verification_notes,
			COALESCE(NULLIF(to_jsonb(lo)->>'created_at', ''), '') as created_at,
			COALESCE(NULLIF(to_jsonb(lo)->>'updated_at', ''), '') as updated_at
		FROM lounge_owners lo
		LEFT JOIN users u ON (
			CASE
				WHEN COALESCE(to_jsonb(lo)->>'user_id', '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
				THEN (to_jsonb(lo)->>'user_id')::uuid
				ELSE NULL
			END
		) = u.id
		WHERE LOWER(COALESCE(NULLIF(to_jsonb(lo)->>'verification_status', ''), 'pending')) = 'pending'
		ORDER BY lo.id DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying pending lounge owners: %w", err)
	}
	defer rows.Close()

	owners := []models.LoungeOwner{}
	for rows.Next() {
		var owner models.LoungeOwner

		err := rows.Scan(
			&owner.ID,
			&owner.UserID,
			&owner.ManagerFullName,
			&owner.Email,
			&owner.ContactNumber,
			&owner.NIC,
			&owner.BusinessName,
			&owner.BusinessLicense,
			&owner.VerificationStatus,
			&owner.VerificationNotes,
			&owner.CreatedAt,
			&owner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning lounge owner: %w", err)
		}

		owners = append(owners, owner)
	}

	return owners, nil
}

func GetLoungeOwnerByID(id string) (*models.LoungeOwner, error) {
	query := `
		SELECT
			lo.id::text,
			COALESCE(NULLIF(to_jsonb(lo)->>'user_id', ''), '') as user_id,
			COALESCE(NULLIF(to_jsonb(lo)->>'manager_full_name', ''), NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), '') as manager_full_name,
			COALESCE(NULLIF(to_jsonb(lo)->>'email', ''), NULLIF(to_jsonb(lo)->>'manager_email', ''), NULLIF(u.email, ''), '') as email,
			COALESCE(NULLIF(to_jsonb(lo)->>'contact_number', ''), NULLIF(u.phone, ''), '') as contact_number,
			COALESCE(NULLIF(to_jsonb(lo)->>'nic', ''), NULLIF(to_jsonb(lo)->>'manager_nic_number', ''), NULLIF(u.nic, ''), '') as nic,
			COALESCE(NULLIF(to_jsonb(lo)->>'business_name', ''), '') as business_name,
			COALESCE(NULLIF(to_jsonb(lo)->>'business_license', ''), '') as business_license,
			COALESCE(NULLIF(to_jsonb(lo)->>'verification_status', ''), 'pending') as verification_status,
			COALESCE(NULLIF(to_jsonb(lo)->>'verification_notes', ''), '') as verification_notes,
			COALESCE(NULLIF(to_jsonb(lo)->>'created_at', ''), '') as created_at,
			COALESCE(NULLIF(to_jsonb(lo)->>'updated_at', ''), '') as updated_at
		FROM lounge_owners lo
		LEFT JOIN users u ON (
			CASE
				WHEN COALESCE(to_jsonb(lo)->>'user_id', '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
				THEN (to_jsonb(lo)->>'user_id')::uuid
				ELSE NULL
			END
		) = u.id
		WHERE lo.id = $1
	`

	var owner models.LoungeOwner
	err := DB.QueryRow(query, id).Scan(
		&owner.ID,
		&owner.UserID,
		&owner.ManagerFullName,
		&owner.Email,
		&owner.ContactNumber,
		&owner.NIC,
		&owner.BusinessName,
		&owner.BusinessLicense,
		&owner.VerificationStatus,
		&owner.VerificationNotes,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting lounge owner: %w", err)
	}

	return &owner, nil
}

func VerifyLoungeOwner(id string, status string, notes string) error {
	query := `
		UPDATE lounge_owners
		SET verification_status = $1,
			verification_notes = $2
		WHERE id = $3
		RETURNING id
	`

	var returnedID string
	err := DB.QueryRow(query, status, notes, id).Scan(&returnedID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("lounge owner not found")
	}
	if err != nil {
		return fmt.Errorf("error verifying lounge owner: %w", err)
	}

	return nil
}
