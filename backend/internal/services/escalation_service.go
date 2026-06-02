package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"sts-backend/internal/models"

	"sts-backend/internal/config"
	"github.com/lib/pq"
)

type EscalationService struct {
	db         *sql.DB
	smsService *SMSService
}

func NewEscalationService(db *sql.DB, cfg *config.Config) *EscalationService {
	return &EscalationService{
		db:         db,
		smsService: NewSMSService(cfg),
	}
}

const (
	level1SLA = 48 * time.Hour
	level2SLA = 24 * time.Hour
)

func getEscalationSLA(sourceApp string, level int) time.Duration {
	if sourceApp == "driver" {
		// Driver workflow: 5 days for admin resolution, then 5 days for supervisor.
		if level == 1 || level == 2 {
			return 5 * 24 * time.Hour
		}
	}

	if level == 1 {
		return level1SLA
	}
	return level2SLA
}

// EnsureEscalationsForOpenComplaints initializes missing escalation rows for unresolved complaints.
func (s *EscalationService) EnsureEscalationsForOpenComplaints() error {
	query := `
		SELECT ri.id
		FROM report_issues ri
		LEFT JOIN complaint_escalations ce ON ce.complaint_id = ri.id
		WHERE ri.status NOT IN ('resolved', 'closed')
		  AND ce.complaint_id IS NULL
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query complaints without escalation: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var complaintID string
		if err := rows.Scan(&complaintID); err != nil {
			return fmt.Errorf("failed to scan complaint id: %w", err)
		}
		if err := s.InitializeEscalation(complaintID, ""); err != nil {
			log.Printf("failed to initialize escalation for complaint %s: %v", complaintID, err)
		}
	}

	return rows.Err()
}

// InitializeEscalation sets up escalation tracking for a new complaint in the database.
func (s *EscalationService) InitializeEscalation(complaintID, _ string) error {
	var exists bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM complaint_escalations WHERE complaint_id = $1)`, complaintID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check existing escalation: %w", err)
	}
	if exists {
		return nil
	}

	sourceApp, err := s.getSourceAppForComplaint(complaintID)
	if err != nil {
		return err
	}

	assignedAdminID, assignedName, err := s.findInitialAdminAssignee(sourceApp)
	if err != nil {
		return err
	}

	nextDue := time.Now().Add(getEscalationSLA(sourceApp, 1))
	_, err = s.db.Exec(`
		INSERT INTO complaint_escalations (
			complaint_id, current_level, current_team, source_app,
			assigned_to_admin_id, previous_assigned_admin_id,
			last_escalated_at, next_escalation_due, created_at, updated_at
		)
		VALUES ($1, 1, $2, $3, $4, NULL, NOW(), $5, NOW(), NOW())
	`, complaintID, sourceApp+"_admin", sourceApp, nullableString(assignedAdminID), nextDue)
	if err != nil {
		return fmt.Errorf("failed to insert escalation record: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO complaint_escalation_history (
			complaint_id, level, team_name, escalated_at, escalated_by,
			from_admin_id, to_admin_id, reason, created_at
		)
		VALUES ($1, 1, $2, NOW(), 'system', NULL, $3, $4, NOW())
	`, complaintID, sourceApp+"_admin", nullableString(assignedAdminID), "Initial app admin assignment")
	if err != nil {
		return fmt.Errorf("failed to insert escalation history: %w", err)
	}

	if assignedAdminID != "" {
		log.Printf("initialized complaint %s to %s (%s)", complaintID, assignedName, sourceApp)
	}

	return nil
}

// CheckAndEscalateComplaints finds complaints that need escalation and escalates them.
func (s *EscalationService) CheckAndEscalateComplaints() (int, error) {
	if err := s.EnsureEscalationsForOpenComplaints(); err != nil {
		return 0, err
	}

	query := `
		SELECT ce.complaint_id, ce.current_level, COALESCE(ri.issue_type, 'other')
		FROM complaint_escalations ce
		JOIN report_issues ri ON ri.id = ce.complaint_id
		WHERE ri.status NOT IN ('resolved', 'closed')
		  AND ce.next_escalation_due IS NOT NULL
		  AND ce.next_escalation_due <= NOW()
		  AND ce.current_level < 3
		ORDER BY ce.next_escalation_due ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to query due escalations: %w", err)
	}
	defer rows.Close()

	escalatedCount := 0

	for rows.Next() {
		var complaintID, issueType string
		var currentLevel int
		if err := rows.Scan(&complaintID, &currentLevel, &issueType); err != nil {
			log.Printf("failed to scan due escalation: %v", err)
			continue
		}

		if err := s.EscalateToNextLevel(complaintID, issueType, currentLevel, "auto"); err != nil {
			log.Printf("failed to escalate complaint %s: %v", complaintID, err)
			continue
		}
		escalatedCount++
	}

	return escalatedCount, nil
}

// EscalateToNextLevel escalates a complaint to the next owner level in the database.
func (s *EscalationService) EscalateToNextLevel(complaintID, _ string, currentLevel int, escalatedBy string) error {
	var sourceApp string
	var currentAssigned sql.NullString
	err := s.db.QueryRow(`
		SELECT source_app, assigned_to_admin_id::text
		FROM complaint_escalations
		WHERE complaint_id = $1
	`, complaintID).Scan(&sourceApp, &currentAssigned)
	if err != nil {
		if err == sql.ErrNoRows {
			if initErr := s.InitializeEscalation(complaintID, ""); initErr != nil {
				return initErr
			}
			return s.EscalateToNextLevel(complaintID, "", currentLevel, escalatedBy)
		}
		return fmt.Errorf("failed to get escalation state: %w", err)
	}

	if currentLevel >= 3 {
		return nil
	}

	fromAdminID := nullableToString(currentAssigned)
	var toAdminID, teamName, reason string
	var nextDue *time.Time
	newLevel := currentLevel + 1

	if newLevel == 2 {
		toAdminID, _, err = s.findSupervisorAssignee(sourceApp, fromAdminID)
		if err != nil {
			return err
		}
		teamName = sourceApp + "_supervisor"
		reason = "Auto-escalated to supervisor due to unresolved SLA"
		t := time.Now().Add(getEscalationSLA(sourceApp, 2))
		nextDue = &t
	} else {
		toAdminID, _, err = s.findSuperAdminAssignee()
		if err != nil {
			return err
		}
		teamName = "super_admin"
		reason = "Escalated to super admin after supervisor SLA"
		nextDue = nil
	}

	_, err = s.db.Exec(`
		UPDATE complaint_escalations
		SET current_level = $2,
			current_team = $3,
			previous_assigned_admin_id = assigned_to_admin_id,
			assigned_to_admin_id = $4,
			last_escalated_at = NOW(),
			next_escalation_due = $5,
			updated_at = NOW()
		WHERE complaint_id = $1
	`, complaintID, newLevel, teamName, nullableString(toAdminID), nextDue)
	if err != nil {
		return fmt.Errorf("failed to update escalation row: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO complaint_escalation_history (
			complaint_id, level, team_name, escalated_at, escalated_by,
			from_admin_id, to_admin_id, reason, created_at
		)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6, $7, NOW())
	`, complaintID, newLevel, teamName, escalatedBy, nullableString(fromAdminID), nullableString(toAdminID), reason)
	if err != nil {
		return fmt.Errorf("failed to insert escalation history: %w", err)
	}

	log.Printf("escalated complaint %s to level %d (%s)", complaintID, newLevel, teamName)
	return nil
}

// GetComplaintEscalation retrieves escalation info for a complaint from database.
func (s *EscalationService) GetComplaintEscalation(complaintID string) (*models.ComplaintEscalation, error) {
	query := `
		SELECT
			ce.current_level,
			ce.current_team,
			ce.source_app,
			ce.assigned_to_admin_id::text,
			ce.previous_assigned_admin_id::text,
			COALESCE(au.full_name, ''),
			ce.last_escalated_at,
			ce.next_escalation_due
		FROM complaint_escalations ce
		LEFT JOIN admin_users au ON ce.assigned_to_admin_id = au.id
		WHERE ce.complaint_id = $1
	`

	var escalation models.ComplaintEscalation
	var assignedID, prevAssignedID sql.NullString
	var assignedName string
	var lastEscalatedAt sql.NullTime
	var nextDue sql.NullTime

	err := s.db.QueryRow(query, complaintID).Scan(
		&escalation.CurrentLevel,
		&escalation.CurrentTeam,
		&escalation.SourceApp,
		&assignedID,
		&prevAssignedID,
		&assignedName,
		&lastEscalatedAt,
		&nextDue,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query complaint escalation: %w", err)
	}

	escalation.AssignedToAdminID = nullableStringPtr(assignedID)
	escalation.PreviousAssignedAdminID = nullableStringPtr(prevAssignedID)
	escalation.AssignedToName = assignedName
	if lastEscalatedAt.Valid {
		t := lastEscalatedAt.Time
		escalation.LastEscalatedAt = &t
	}
	if nextDue.Valid {
		t := nextDue.Time
		escalation.NextEscalationDue = &t
	}

	return &escalation, nil
}

// GetComplaintEscalations retrieves escalation info for many complaints in a single query.
func (s *EscalationService) GetComplaintEscalations(complaintIDs []string) (map[string]*models.ComplaintEscalation, error) {
	result := make(map[string]*models.ComplaintEscalation, len(complaintIDs))
	if len(complaintIDs) == 0 {
		return result, nil
	}

	rows, err := s.db.Query(`
		SELECT
			ce.complaint_id,
			ce.current_level,
			ce.current_team,
			ce.source_app,
			ce.assigned_to_admin_id::text,
			ce.previous_assigned_admin_id::text,
			COALESCE(au.full_name, ''),
			ce.last_escalated_at,
			ce.next_escalation_due
		FROM complaint_escalations ce
		LEFT JOIN admin_users au ON ce.assigned_to_admin_id = au.id
		WHERE ce.complaint_id = ANY($1)
	`, pq.Array(complaintIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query complaint escalations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var complaintID string
		var escalation models.ComplaintEscalation
		var assignedID, prevAssignedID sql.NullString
		var assignedName string
		var lastEscalatedAt sql.NullTime
		var nextDue sql.NullTime

		if err := rows.Scan(
			&complaintID,
			&escalation.CurrentLevel,
			&escalation.CurrentTeam,
			&escalation.SourceApp,
			&assignedID,
			&prevAssignedID,
			&assignedName,
			&lastEscalatedAt,
			&nextDue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan complaint escalation: %w", err)
		}

		escalation.AssignedToAdminID = nullableStringPtr(assignedID)
		escalation.PreviousAssignedAdminID = nullableStringPtr(prevAssignedID)
		escalation.AssignedToName = assignedName
		if lastEscalatedAt.Valid {
			t := lastEscalatedAt.Time
			escalation.LastEscalatedAt = &t
		}
		if nextDue.Valid {
			t := nextDue.Time
			escalation.NextEscalationDue = &t
		}

		copyEscalation := escalation
		result[complaintID] = &copyEscalation
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate complaint escalations: %w", err)
	}

	return result, nil
}

// GetEscalationHistory retrieves the escalation history for a complaint from database.
func (s *EscalationService) GetEscalationHistory(complaintID string) ([]models.EscalationHistoryEntry, error) {
	rows, err := s.db.Query(`
		SELECT level, team_name, escalated_at, escalated_by,
			from_admin_id::text, to_admin_id::text, COALESCE(reason, '')
		FROM complaint_escalation_history
		WHERE complaint_id = $1
		ORDER BY escalated_at ASC
	`, complaintID)
	if err != nil {
		return nil, fmt.Errorf("failed to query escalation history: %w", err)
	}
	defer rows.Close()

	history := []models.EscalationHistoryEntry{}
	for rows.Next() {
		entry := models.EscalationHistoryEntry{}
		var fromAdminID, toAdminID sql.NullString
		if err := rows.Scan(
			&entry.Level,
			&entry.TeamName,
			&entry.EscalatedAt,
			&entry.EscalatedBy,
			&fromAdminID,
			&toAdminID,
			&entry.Reason,
		); err != nil {
			return nil, fmt.Errorf("failed to scan escalation history: %w", err)
		}
		entry.FromAdminID = nullableStringPtr(fromAdminID)
		entry.ToAdminID = nullableStringPtr(toAdminID)
		history = append(history, entry)
	}

	return history, rows.Err()
}

// AssignComplaintToAdmin assigns a complaint to a specific admin in database.
func (s *EscalationService) AssignComplaintToAdmin(complaintID, adminID string) error {
	_, err := s.db.Exec(`
		UPDATE complaint_escalations
		SET previous_assigned_admin_id = assigned_to_admin_id,
			assigned_to_admin_id = $2,
			updated_at = NOW()
		WHERE complaint_id = $1
	`, complaintID, adminID)
	if err != nil {
		return fmt.Errorf("failed to assign complaint: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO complaint_escalation_history (
			complaint_id, level, team_name, escalated_at, escalated_by,
			from_admin_id, to_admin_id, reason, created_at
		)
		SELECT complaint_id, current_level, current_team, NOW(), 'manual',
			previous_assigned_admin_id, assigned_to_admin_id, 'Manual assignment', NOW()
		FROM complaint_escalations
		WHERE complaint_id = $1
	`, complaintID)
	if err != nil {
		return fmt.Errorf("failed to record manual assignment history: %w", err)
	}

	return nil
}

// GetEscalationStats returns escalation statistics from database.
func (s *EscalationService) GetEscalationStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	rows, err := s.db.Query(`
		SELECT current_level, COUNT(*)
		FROM complaint_escalations ce
		JOIN report_issues ri ON ri.id = ce.complaint_id
		WHERE ri.status NOT IN ('resolved', 'closed')
		GROUP BY current_level
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query level stats: %w", err)
	}
	defer rows.Close()

	levelCounts := map[int]int{}
	total := 0
	for rows.Next() {
		var level, count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("failed to scan level stats: %w", err)
		}
		levelCounts[level] = count
		total += count
	}

	var overdueCount int
	if err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM complaint_escalations ce
		JOIN report_issues ri ON ri.id = ce.complaint_id
		WHERE ri.status NOT IN ('resolved', 'closed')
		  AND ce.next_escalation_due IS NOT NULL
		  AND ce.next_escalation_due <= NOW()
	`).Scan(&overdueCount); err != nil {
		return nil, fmt.Errorf("failed to query overdue count: %w", err)
	}

	stats["by_level"] = levelCounts
	stats["overdue_count"] = overdueCount
	stats["total_tracked"] = total

	return stats, nil
}

func (s *EscalationService) getSourceAppForComplaint(complaintID string) (string, error) {
	var app string
	err := s.db.QueryRow(`
		SELECT CASE
			WHEN 'driver' = ANY(u.roles) THEN 'driver'
			WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounges'
			WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
			WHEN 'bus_owner' = ANY(u.roles) OR 'conductor' = ANY(u.roles) THEN 'bus'
			ELSE 'bus'
		END
		FROM report_issues ri
		JOIN users u ON u.id = ri.reported_by_id
		WHERE ri.id = $1
	`, complaintID).Scan(&app)
	if err != nil {
		return "", fmt.Errorf("failed to get source app for complaint: %w", err)
	}
	return app, nil
}

func (s *EscalationService) findInitialAdminAssignee(appScope string) (string, string, error) {
	adminID, name, err := s.findAdminByRoleScope("admin", appScope)
	if err == nil {
		return adminID, name, nil
	}

	adminID, name, err = s.findAdminByRoleScope("supervisor", appScope)
	if err == nil {
		return adminID, name, nil
	}

	return s.findSuperAdminAssignee()
}

func (s *EscalationService) findSupervisorAssignee(appScope, currentAdminID string) (string, string, error) {
	if currentAdminID != "" {
		var supervisorID, supervisorName sql.NullString
		err := s.db.QueryRow(`
			SELECT su.id::text, su.full_name
			FROM admin_users au
			JOIN admin_users su ON su.id = au.supervisor_id
			WHERE au.id = $1 AND su.is_active = true
		`, currentAdminID).Scan(&supervisorID, &supervisorName)
		if err == nil && supervisorID.Valid {
			return supervisorID.String, supervisorName.String, nil
		}
	}

	adminID, name, err := s.findAdminByRoleScope("supervisor", appScope)
	if err == nil {
		return adminID, name, nil
	}

	return s.findSuperAdminAssignee()
}

func (s *EscalationService) findSuperAdminAssignee() (string, string, error) {
	return s.findAdminByRoleScope("super_admin", "")
}

func (s *EscalationService) findAdminByRoleScope(role, scope string) (string, string, error) {
	query := `
		SELECT id::text, full_name
		FROM admin_users
		WHERE role = $1
		  AND is_active = true
	`
	args := []interface{}{role}
	if scope != "" {
		query += " AND app_scope = $2"
		args = append(args, scope)
	}
	query += " ORDER BY created_at ASC LIMIT 1"

	var id, name string
	if err := s.db.QueryRow(query, args...).Scan(&id, &name); err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("no active %s found", role)
		}
		return "", "", fmt.Errorf("failed to resolve %s assignee: %w", role, err)
	}

	return id, name, nil
}

func nullableToString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func nullableString(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func nullableStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func normalizeAccessToken(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func buildTeamName(appRole, appScope string) string {
	role := normalizeAccessToken(appRole)
	scope := normalizeAccessToken(appScope)
	if role == "super_admin" {
		return "super_admin"
	}
	if scope == "" {
		return role
	}
	return scope + "_" + role
}

func (s *EscalationService) getEscalationSourceApp(complaintID string) (string, error) {
	var sourceApp string
	err := s.db.QueryRow(`
		SELECT COALESCE(source_app, '')
		FROM complaint_escalations
		WHERE complaint_id = $1
	`, complaintID).Scan(&sourceApp)
	if err != nil {
		return "", err
	}
	return sourceApp, nil
}

// AssignComplaintToRole assigns complaint ownership to a role/scope team instead of a specific admin ID.
func (s *EscalationService) AssignComplaintToRole(complaintID, appRole, appScope string) error {
	role := normalizeAccessToken(appRole)
	scope := normalizeAccessToken(appScope)

	if role == "" {
		return fmt.Errorf("app_role is required")
	}

	if scope == "" && role != "super_admin" {
		sourceApp, err := s.getEscalationSourceApp(complaintID)
		if err != nil {
			return fmt.Errorf("failed to resolve complaint source app: %w", err)
		}
		scope = normalizeAccessToken(sourceApp)
	}

	teamName := buildTeamName(role, scope)

	var toAdminID string
	if role == "super_admin" {
		id, _, err := s.findSuperAdminAssignee()
		if err == nil {
			toAdminID = id
		}
	} else if scope != "" {
		id, _, err := s.findAdminByRoleScope(role, scope)
		if err == nil {
			toAdminID = id
		}
	}

	_, err := s.db.Exec(`
		UPDATE complaint_escalations
		SET previous_assigned_admin_id = assigned_to_admin_id,
			assigned_to_admin_id = $2,
			current_team = $3,
			updated_at = NOW()
		WHERE complaint_id = $1
	`, complaintID, nullableString(toAdminID), teamName)
	if err != nil {
		return fmt.Errorf("failed to assign complaint by role: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO complaint_escalation_history (
			complaint_id, level, team_name, escalated_at, escalated_by,
			from_admin_id, to_admin_id, reason, created_at
		)
		SELECT complaint_id, current_level, current_team, NOW(), 'manual',
			previous_assigned_admin_id, assigned_to_admin_id, $2, NOW()
		FROM complaint_escalations
		WHERE complaint_id = $1
	`, complaintID, fmt.Sprintf("Manual role assignment to %s", teamName))
	if err != nil {
		return fmt.Errorf("failed to record role assignment history: %w", err)
	}

	return nil
}
