package models

import "time"

type Complaint struct {
	ID                string  `json:"id"`
	ScheduledTripID   *string `json:"scheduled_trip_id"`
	ActiveTripID      *string `json:"active_trip_id"`
	ReportedByID      string  `json:"reported_by_id"`
	IssueType         string  `json:"issue_type"`
	Priority          string  `json:"priority"`
	Status            string  `json:"status"`
	Description       string  `json:"description"`
	Latitude          *float64 `json:"latitude"`
	Longitude         *float64 `json:"longitude"`
	LocationAddress   *string `json:"location_address"`
	ImageURL          *string `json:"image_url"`
	ResolvedAt        *string `json:"resolved_at"`
	ResolvedByID      *string `json:"resolved_by_id"`
	ResolutionNotes   *string `json:"resolution_notes"`
	NotifiedPassengers bool   `json:"notified_passengers"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	
	// Joined fields from users table
	ReporterName      string  `json:"reporter_name"`
	ReporterPhone     string  `json:"reporter_phone"`
	ReporterRole      string  `json:"reporter_role"`
	ResolverName      *string `json:"resolver_name"`
}

type ComplaintResponse struct {
	ID             string  `json:"id"`
	Role           string  `json:"role"`
	Name           string  `json:"name"`
	Contact        string  `json:"contact"`
	Category       string  `json:"category"`
	Message        string  `json:"message"`
	Media          string  `json:"media"`
	DateTime       string  `json:"dateTime"`
	AssignedTeam   string  `json:"assignedTeam"`
	ResolvedBy     string  `json:"resolvedBy"`
	Activity       string  `json:"activity"`
	Status         string  `json:"status"`
	Priority       string  `json:"priority"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	LocationAddress *string `json:"locationAddress"`
	
	// Escalation fields
	Escalation     *ComplaintEscalation `json:"escalation,omitempty"`
}

type ComplaintListResponse struct {
	Data       []ComplaintResponse `json:"data"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	Total      int                 `json:"total"`
	TotalPages int                 `json:"total_pages"`
}

// ComplaintWithEscalation combines complaint data with escalation status
type ComplaintWithEscalation struct {
	Complaint
	EscalationLevel      int        `json:"escalation_level"`
	EscalationTeam       string     `json:"escalation_team"`
	AssignedToAdminID    *string    `json:"assigned_to_admin_id"`
	LastEscalatedAt      *time.Time `json:"last_escalated_at"`
	NextEscalationDue    *time.Time `json:"next_escalation_due"`
	DaysUntilEscalation  *int       `json:"days_until_escalation"`
}
