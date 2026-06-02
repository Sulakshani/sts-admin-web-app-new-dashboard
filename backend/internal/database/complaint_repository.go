package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sts-backend/internal/models"
)

func normalizeAccessValue(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func resolveEffectiveAdminAccess(adminID, role, appScope string) (string, string) {
	effectiveRole := normalizeAccessValue(role)
	effectiveScope := normalizeAccessValue(appScope)

	if adminID == "" {
		return effectiveRole, effectiveScope
	}

	var dbRole string
	var dbScope string
	err := DB.QueryRow(
		`SELECT COALESCE(role, ''), COALESCE(app_scope, '') FROM admin_users WHERE id::text = $1`,
		adminID,
	).Scan(&dbRole, &dbScope)
	if err == nil {
		if normalized := normalizeAccessValue(dbRole); normalized != "" {
			effectiveRole = normalized
		}
		if normalized := normalizeAccessValue(dbScope); normalized != "" {
			effectiveScope = normalized
		}
	}

	return effectiveRole, effectiveScope
}

func GetComplaintsByRole(role string) ([]models.Complaint, error) {
	query := `
		SELECT 
			ri.id,
			ri.scheduled_trip_id,
			ri.active_trip_id,
			ri.reported_by_id,
			ri.issue_type,
			ri.priority,
			ri.status,
			ri.description,
			ri.latitude,
			ri.longitude,
			ri.location_address,
			ri.image_url,
			ri.resolved_at,
			ri.resolved_by_id,
			ri.resolution_notes,
			ri.notified_passengers,
			ri.created_at,
			ri.updated_at,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
			COALESCE(u.phone, '') as reporter_phone,
			COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
		FROM report_issues ri
		LEFT JOIN users u ON ri.reported_by_id = u.id
		LEFT JOIN users r ON ri.resolved_by_id = r.id
		WHERE u.roles && ARRAY[$1]::text[]
		ORDER BY ri.created_at DESC
	`

	rows, err := DB.Query(query, role)
	if err != nil {
		return nil, fmt.Errorf("error querying complaints by role: %w", err)
	}
	defer rows.Close()

	var complaints []models.Complaint
	for rows.Next() {
		var c models.Complaint
		err := rows.Scan(
			&c.ID,
			&c.ScheduledTripID,
			&c.ActiveTripID,
			&c.ReportedByID,
			&c.IssueType,
			&c.Priority,
			&c.Status,
			&c.Description,
			&c.Latitude,
			&c.Longitude,
			&c.LocationAddress,
			&c.ImageURL,
			&c.ResolvedAt,
			&c.ResolvedByID,
			&c.ResolutionNotes,
			&c.NotifiedPassengers,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.ReporterName,
			&c.ReporterPhone,
			&c.ResolverName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning complaint row: %w", err)
		}
		c.ReporterRole = role
		complaints = append(complaints, c)
	}

	return complaints, nil
}

func GetAllComplaints() ([]models.Complaint, error) {
	query := `
		SELECT 
			ri.id,
			ri.scheduled_trip_id,
			ri.active_trip_id,
			ri.reported_by_id,
			ri.issue_type,
			ri.priority,
			ri.status,
			ri.description,
			ri.latitude,
			ri.longitude,
			ri.location_address,
			ri.image_url,
			ri.resolved_at,
			ri.resolved_by_id,
			ri.resolution_notes,
			ri.notified_passengers,
			ri.created_at,
			ri.updated_at,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
			COALESCE(u.phone, '') as reporter_phone,
			CASE 
				WHEN 'driver' = ANY(u.roles) THEN 'driver'
				WHEN 'conductor' = ANY(u.roles) THEN 'conductor'
				WHEN 'bus_owner' = ANY(u.roles) THEN 'bus_owner'
				WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounge_owner'
				WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
				ELSE 'unknown'
			END as reporter_role,
			COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
		FROM report_issues ri
		LEFT JOIN users u ON ri.reported_by_id = u.id
		LEFT JOIN users r ON ri.resolved_by_id = r.id
		ORDER BY ri.created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying all complaints: %w", err)
	}
	defer rows.Close()

	var complaints []models.Complaint
	for rows.Next() {
		var c models.Complaint
		err := rows.Scan(
			&c.ID,
			&c.ScheduledTripID,
			&c.ActiveTripID,
			&c.ReportedByID,
			&c.IssueType,
			&c.Priority,
			&c.Status,
			&c.Description,
			&c.Latitude,
			&c.Longitude,
			&c.LocationAddress,
			&c.ImageURL,
			&c.ResolvedAt,
			&c.ResolvedByID,
			&c.ResolutionNotes,
			&c.NotifiedPassengers,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.ReporterName,
			&c.ReporterPhone,
			&c.ReporterRole,
			&c.ResolverName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning complaint row: %w", err)
		}
		complaints = append(complaints, c)
	}

	return complaints, nil
}

func GetComplaintsForAdmin(adminID, role, appScope string) ([]models.Complaint, error) {
	effectiveRole, effectiveScope := resolveEffectiveAdminAccess(adminID, role, appScope)

	if effectiveRole == "super_admin" {
		return GetAllComplaints()
	}

	// Driver workflow: driver admin/supervisor can only see complaints assigned to their app-role team.
	if effectiveScope == "driver" && (effectiveRole == "admin" || effectiveRole == "supervisor") {
		query := `
			SELECT 
				ri.id,
				ri.scheduled_trip_id,
				ri.active_trip_id,
				ri.reported_by_id,
				ri.issue_type,
				ri.priority,
				ri.status,
				ri.description,
				ri.latitude,
				ri.longitude,
				ri.location_address,
				ri.image_url,
				ri.resolved_at,
				ri.resolved_by_id,
				ri.resolution_notes,
				ri.notified_passengers,
				ri.created_at,
				ri.updated_at,
				COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
				COALESCE(u.phone, '') as reporter_phone,
				CASE 
					WHEN 'driver' = ANY(u.roles) THEN 'driver'
					WHEN 'conductor' = ANY(u.roles) THEN 'conductor'
					WHEN 'bus_owner' = ANY(u.roles) THEN 'bus_owner'
					WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounge_owner'
					WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
					ELSE 'unknown'
				END as reporter_role,
				COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
			FROM report_issues ri
			LEFT JOIN users u ON ri.reported_by_id = u.id
			LEFT JOIN users r ON ri.resolved_by_id = r.id
			WHERE 'driver' = ANY(u.roles)
			ORDER BY ri.created_at DESC
		`

		rows, err := DB.Query(query)
		if err != nil {
			return nil, fmt.Errorf("error querying driver complaints: %w", err)
		}
		defer rows.Close()

		var complaints []models.Complaint
		for rows.Next() {
			var c models.Complaint
			err := rows.Scan(
				&c.ID,
				&c.ScheduledTripID,
				&c.ActiveTripID,
				&c.ReportedByID,
				&c.IssueType,
				&c.Priority,
				&c.Status,
				&c.Description,
				&c.Latitude,
				&c.Longitude,
				&c.LocationAddress,
				&c.ImageURL,
				&c.ResolvedAt,
				&c.ResolvedByID,
				&c.ResolutionNotes,
				&c.NotifiedPassengers,
				&c.CreatedAt,
				&c.UpdatedAt,
				&c.ReporterName,
				&c.ReporterPhone,
				&c.ReporterRole,
				&c.ResolverName,
			)
			if err != nil {
				return nil, fmt.Errorf("error scanning complaint row: %w", err)
			}
			complaints = append(complaints, c)
		}

		return complaints, nil
	}

	// Non-driver scoped admins retain broad read visibility.
	return GetAllComplaints()
}

func GetComplaintsForAdminPaginated(adminID, role, appScope, reporterRole string, limit, offset int) ([]models.Complaint, int, error) {
	effectiveRole, effectiveScope := resolveEffectiveAdminAccess(adminID, role, appScope)
	reporterRole = normalizeAccessValue(reporterRole)
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	if effectiveScope == "driver" && (effectiveRole == "admin" || effectiveRole == "supervisor") {
		countQuery := `
			SELECT COUNT(*)
			FROM report_issues ri
			LEFT JOIN users u ON ri.reported_by_id = u.id
			WHERE 'driver' = ANY(u.roles)
		`
		var total int
		if err := DB.QueryRow(countQuery).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("error counting paginated driver complaints: %w", err)
		}

		query := `
			SELECT 
				ri.id,
				ri.scheduled_trip_id,
				ri.active_trip_id,
				ri.reported_by_id,
				ri.issue_type,
				ri.priority,
				ri.status,
				ri.description,
				ri.latitude,
				ri.longitude,
				ri.location_address,
				ri.image_url,
				ri.resolved_at,
				ri.resolved_by_id,
				ri.resolution_notes,
				ri.notified_passengers,
				ri.created_at,
				ri.updated_at,
				COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
				COALESCE(u.phone, '') as reporter_phone,
				CASE 
					WHEN 'driver' = ANY(u.roles) THEN 'driver'
					WHEN 'conductor' = ANY(u.roles) THEN 'conductor'
					WHEN 'bus_owner' = ANY(u.roles) THEN 'bus_owner'
					WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounge_owner'
					WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
					ELSE 'unknown'
				END as reporter_role,
				COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
			FROM report_issues ri
			LEFT JOIN users u ON ri.reported_by_id = u.id
			LEFT JOIN users r ON ri.resolved_by_id = r.id
			WHERE 'driver' = ANY(u.roles)
			ORDER BY ri.created_at DESC
			LIMIT $1 OFFSET $2
		`

		rows, err := DB.Query(query, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("error querying paginated driver complaints: %w", err)
		}
		defer rows.Close()

		complaints := make([]models.Complaint, 0, limit)
		for rows.Next() {
			var c models.Complaint
			err := rows.Scan(
				&c.ID,
				&c.ScheduledTripID,
				&c.ActiveTripID,
				&c.ReportedByID,
				&c.IssueType,
				&c.Priority,
				&c.Status,
				&c.Description,
				&c.Latitude,
				&c.Longitude,
				&c.LocationAddress,
				&c.ImageURL,
				&c.ResolvedAt,
				&c.ResolvedByID,
				&c.ResolutionNotes,
				&c.NotifiedPassengers,
				&c.CreatedAt,
				&c.UpdatedAt,
				&c.ReporterName,
				&c.ReporterPhone,
				&c.ReporterRole,
				&c.ResolverName,
			)
			if err != nil {
				return nil, 0, fmt.Errorf("error scanning paginated driver complaint row: %w", err)
			}
			complaints = append(complaints, c)
		}

		return complaints, total, nil
	}

	countQuery := `
		SELECT COUNT(*)
		FROM report_issues ri
		LEFT JOIN users u ON ri.reported_by_id = u.id
		WHERE ($1 = '' OR $1 = ANY(u.roles))
	`
	var total int
	if err := DB.QueryRow(countQuery, reporterRole).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting paginated complaints: %w", err)
	}

	query := `
		SELECT 
			ri.id,
			ri.scheduled_trip_id,
			ri.active_trip_id,
			ri.reported_by_id,
			ri.issue_type,
			ri.priority,
			ri.status,
			ri.description,
			ri.latitude,
			ri.longitude,
			ri.location_address,
			ri.image_url,
			ri.resolved_at,
			ri.resolved_by_id,
			ri.resolution_notes,
			ri.notified_passengers,
			ri.created_at,
			ri.updated_at,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
			COALESCE(u.phone, '') as reporter_phone,
			CASE 
				WHEN 'driver' = ANY(u.roles) THEN 'driver'
				WHEN 'conductor' = ANY(u.roles) THEN 'conductor'
				WHEN 'bus_owner' = ANY(u.roles) THEN 'bus_owner'
				WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounge_owner'
				WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
				ELSE 'unknown'
			END as reporter_role,
			COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
		FROM report_issues ri
		LEFT JOIN users u ON ri.reported_by_id = u.id
		LEFT JOIN users r ON ri.resolved_by_id = r.id
		WHERE ($1 = '' OR $1 = ANY(u.roles))
		ORDER BY ri.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := DB.Query(query, reporterRole, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying paginated complaints: %w", err)
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0, limit)
	for rows.Next() {
		var c models.Complaint
		err := rows.Scan(
			&c.ID,
			&c.ScheduledTripID,
			&c.ActiveTripID,
			&c.ReportedByID,
			&c.IssueType,
			&c.Priority,
			&c.Status,
			&c.Description,
			&c.Latitude,
			&c.Longitude,
			&c.LocationAddress,
			&c.ImageURL,
			&c.ResolvedAt,
			&c.ResolvedByID,
			&c.ResolutionNotes,
			&c.NotifiedPassengers,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.ReporterName,
			&c.ReporterPhone,
			&c.ReporterRole,
			&c.ResolverName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning paginated complaint row: %w", err)
		}
		complaints = append(complaints, c)
	}

	return complaints, total, nil
}

func GetComplaintByIDForAdmin(id, adminID, role string) (*models.Complaint, error) {
	effectiveRole, effectiveScope := resolveEffectiveAdminAccess(adminID, role, "")

	if effectiveRole == "super_admin" {
		return GetComplaintByID(id)
	}

	if effectiveScope == "driver" && (effectiveRole == "admin" || effectiveRole == "supervisor") {
		query := `
			SELECT 
				ri.id,
				ri.scheduled_trip_id,
				ri.active_trip_id,
				ri.reported_by_id,
				ri.issue_type,
				ri.priority,
				ri.status,
				ri.description,
				ri.latitude,
				ri.longitude,
				ri.location_address,
				ri.image_url,
				ri.resolved_at,
				ri.resolved_by_id,
				ri.resolution_notes,
				ri.notified_passengers,
				ri.created_at,
				ri.updated_at,
				COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
				COALESCE(u.phone, '') as reporter_phone,
				CASE 
					WHEN 'driver' = ANY(u.roles) THEN 'driver'
					WHEN 'conductor' = ANY(u.roles) THEN 'conductor'
					WHEN 'bus_owner' = ANY(u.roles) THEN 'bus_owner'
					WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounge_owner'
					WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
					ELSE 'unknown'
				END as reporter_role,
				COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
			FROM report_issues ri
			LEFT JOIN users u ON ri.reported_by_id = u.id
			LEFT JOIN users r ON ri.resolved_by_id = r.id
			WHERE ri.id = $1
			  AND 'driver' = ANY(u.roles)
		`

		var c models.Complaint
		err := DB.QueryRow(query, id).Scan(
			&c.ID,
			&c.ScheduledTripID,
			&c.ActiveTripID,
			&c.ReportedByID,
			&c.IssueType,
			&c.Priority,
			&c.Status,
			&c.Description,
			&c.Latitude,
			&c.Longitude,
			&c.LocationAddress,
			&c.ImageURL,
			&c.ResolvedAt,
			&c.ResolvedByID,
			&c.ResolutionNotes,
			&c.NotifiedPassengers,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.ReporterName,
			&c.ReporterPhone,
			&c.ReporterRole,
			&c.ResolverName,
		)

		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("error querying complaint by id for driver admin/supervisor: %w", err)
		}

		return &c, nil
	}

	return GetComplaintByID(id)
}

func GetComplaintByID(id string) (*models.Complaint, error) {
	query := `
		SELECT 
			ri.id,
			ri.scheduled_trip_id,
			ri.active_trip_id,
			ri.reported_by_id,
			ri.issue_type,
			ri.priority,
			ri.status,
			ri.description,
			ri.latitude,
			ri.longitude,
			ri.location_address,
			ri.image_url,
			ri.resolved_at,
			ri.resolved_by_id,
			ri.resolution_notes,
			ri.notified_passengers,
			ri.created_at,
			ri.updated_at,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, u.last_name, '') as reporter_name,
			COALESCE(u.phone, '') as reporter_phone,
			CASE 
				WHEN 'driver' = ANY(u.roles) THEN 'driver'
				WHEN 'conductor' = ANY(u.roles) THEN 'conductor'
				WHEN 'bus_owner' = ANY(u.roles) THEN 'bus_owner'
				WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounge_owner'
				WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
				ELSE 'unknown'
			END as reporter_role,
			COALESCE(r.first_name || ' ' || r.last_name, r.first_name, r.last_name, '') as resolver_name
		FROM report_issues ri
		LEFT JOIN users u ON ri.reported_by_id = u.id
		LEFT JOIN users r ON ri.resolved_by_id = r.id
		WHERE ri.id = $1
		  AND u.roles && ARRAY['driver', 'conductor', 'bus_owner', 'lounge_owner', 'passenger']::text[]
	`

	var c models.Complaint
	err := DB.QueryRow(query, id).Scan(
		&c.ID,
		&c.ScheduledTripID,
		&c.ActiveTripID,
		&c.ReportedByID,
		&c.IssueType,
		&c.Priority,
		&c.Status,
		&c.Description,
		&c.Latitude,
		&c.Longitude,
		&c.LocationAddress,
		&c.ImageURL,
		&c.ResolvedAt,
		&c.ResolvedByID,
		&c.ResolutionNotes,
		&c.NotifiedPassengers,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.ReporterName,
		&c.ReporterPhone,
		&c.ReporterRole,
		&c.ResolverName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying complaint by ID: %w", err)
	}

	return &c, nil
}

func UpdateComplaintStatus(id string, status string, resolvedByID *string, resolutionNotes *string) error {
	query := `
		UPDATE report_issues 
		SET status = CAST($1 AS VARCHAR), 
			resolved_by_id = CASE
				WHEN CAST($1 AS VARCHAR) IN ('resolved', 'closed') AND CAST($2 AS UUID) IS NOT NULL THEN (
					SELECT u.id
					FROM admin_users au
					JOIN users u ON lower(trim(u.email)) = lower(trim(au.email))
					WHERE au.id = CAST($2 AS UUID)
					LIMIT 1
				)
				ELSE NULL
			END,
			resolution_notes = $3,
			resolved_at = CASE WHEN CAST($1 AS VARCHAR) = 'resolved' THEN NOW() ELSE resolved_at END,
			updated_at = NOW()
		WHERE id = CAST($4 AS UUID)
	`

	_, err := DB.Exec(query, status, resolvedByID, resolutionNotes, id)
	if err != nil {
		return fmt.Errorf("error updating complaint status: %w", err)
	}

	return nil
}
