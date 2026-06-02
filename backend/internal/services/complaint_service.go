package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
	"time"
)

var (
	escalationServiceInstance *EscalationService
	db                        *sql.DB
)

type ComplaintAdminContext struct {
	AdminID  string
	Role     string
	AppScope string
}

func getExpectedDriverComplaintTeam(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "driver_supervisor"
	case "supervisor":
		return "driver_admin"
	default:
		return ""
	}
}

// InitComplaintService initializes the complaint service with escalation support
func InitComplaintService(database *sql.DB, cfg *config.Config) {
	db = database
	escalationServiceInstance = NewEscalationService(database, cfg)
}

func GetComplaintsByRole(role string) ([]models.ComplaintResponse, error) {
	complaints, err := database.GetComplaintsByRole(role)
	if err != nil {
		return nil, err
	}
	return transformComplaints(complaints), nil
}

func GetComplaintsForAdmin(ctx ComplaintAdminContext) ([]models.ComplaintResponse, error) {
	role := strings.ToLower(strings.TrimSpace(ctx.Role))
	scope := strings.ToLower(strings.TrimSpace(ctx.AppScope))

	if role == "super_admin" {
		complaints, err := database.GetAllComplaints()
		if err != nil {
			return nil, err
		}
		return transformComplaints(complaints), nil
	}

	if scope == "driver" && (role == "admin" || role == "supervisor") {
		complaints, err := database.GetComplaintsForAdmin(ctx.AdminID, ctx.Role, ctx.AppScope)
		if err != nil {
			return nil, err
		}
		return transformComplaints(complaints), nil
	}

	complaints, err := database.GetComplaintsForAdmin(ctx.AdminID, ctx.Role, ctx.AppScope)
	if err != nil {
		return nil, err
	}

	return transformComplaints(complaints), nil
}

func GetComplaintsForAdminPaginated(ctx ComplaintAdminContext, reporterRole string, page, pageSize int) (models.ComplaintListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	complaints, total, err := database.GetComplaintsForAdminPaginated(ctx.AdminID, ctx.Role, ctx.AppScope, reporterRole, pageSize, offset)
	if err != nil {
		return models.ComplaintListResponse{}, err
	}

	responses := transformComplaints(complaints)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return models.ComplaintListResponse{
		Data:       responses,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func GetAllComplaints() ([]models.ComplaintResponse, error) {
	complaints, err := database.GetAllComplaints()
	if err != nil {
		return nil, err
	}
	return transformComplaints(complaints), nil
}

func GetComplaintByID(id string) (*models.ComplaintResponse, error) {
	complaint, err := database.GetComplaintByID(id)
	if err != nil {
		return nil, err
	}
	if complaint == nil {
		return nil, nil
	}

	response := transformComplaint(*complaint)
	return &response, nil
}

func GetComplaintByIDForAdmin(id string, ctx ComplaintAdminContext) (*models.ComplaintResponse, error) {
	complaint, err := database.GetComplaintByIDForAdmin(id, ctx.AdminID, ctx.Role)
	if err != nil {
		return nil, err
	}
	if complaint == nil {
		return nil, nil
	}

	response := transformComplaint(*complaint)

	return &response, nil
}

func UpdateComplaintStatus(id string, status string, resolvedByID string, resolutionNotes *string) error {
	var resolver *string
	if status == "resolved" || status == "closed" {
		resolver = &resolvedByID
	}
	return database.UpdateComplaintStatus(id, status, resolver, resolutionNotes)
}

// Helper function to transform complaints to response format
func transformComplaints(complaints []models.Complaint) []models.ComplaintResponse {
	escalations := map[string]*models.ComplaintEscalation{}
	if escalationServiceInstance != nil && len(complaints) > 0 {
		ids := make([]string, 0, len(complaints))
		for _, complaint := range complaints {
			ids = append(ids, complaint.ID)
		}

		if loadedEscalations, err := escalationServiceInstance.GetComplaintEscalations(ids); err == nil {
			escalations = loadedEscalations
		} else {
			log.Printf("warning: failed to batch-load complaint escalations: %v", err)
		}
	}

	responses := make([]models.ComplaintResponse, 0, len(complaints))
	for _, c := range complaints {
		responses = append(responses, transformComplaintWithEscalation(c, escalations[c.ID], false))
	}
	return responses
}

func transformComplaint(c models.Complaint) models.ComplaintResponse {
	return transformComplaintWithEscalation(c, nil, true)
}

func transformComplaintWithEscalation(c models.Complaint, escalation *models.ComplaintEscalation, allowInitialization bool) models.ComplaintResponse {
	category := formatCategory(c.IssueType)

	// Only detail endpoints perform direct escalation lookup/initialization.
	// List endpoints rely on preloaded batch escalation data to avoid N+1 queries.
	if escalationServiceInstance != nil && escalation == nil && allowInitialization {
		escalation, _ = escalationServiceInstance.GetComplaintEscalation(c.ID)
		if allowInitialization && escalation == nil && c.Status != "resolved" && c.Status != "closed" {
			if err := escalationServiceInstance.InitializeEscalation(c.ID, category); err == nil {
				escalation, _ = escalationServiceInstance.GetComplaintEscalation(c.ID)
			}
		}
	}

	// Get assigned team from escalation or fallback to mapping
	assignedTeam := getAssignedTeamFromEscalation(escalation, category, c.IssueType)

	response := models.ComplaintResponse{
		ID:              c.ID,
		Role:            capitalizeRole(c.ReporterRole),
		Name:            c.ReporterName,
		Contact:         c.ReporterPhone,
		Category:        category,
		Message:         c.Description,
		Media:           getMediaDisplay(c.ImageURL),
		DateTime:        formatDateTime(c.CreatedAt),
		AssignedTeam:    assignedTeam,
		ResolvedBy:      getResolverName(c.ResolverName),
		Activity:        formatActivity(c.ResolutionNotes),
		Status:          formatStatus(c.Status),
		Priority:        c.Priority,
		Latitude:        c.Latitude,
		Longitude:       c.Longitude,
		LocationAddress: c.LocationAddress,
		Escalation:      escalation,
	}
	return response
}

func capitalizeRole(role string) string {
	switch role {
	case "driver":
		return "Driver"
	case "conductor":
		return "Conductor"
	case "passenger":
		return "Passenger"
	case "bus_owner":
		return "Bus Owner"
	case "lounge_owner":
		return "Lounge Owner"
	default:
		return role
	}
}

func formatCategory(issueType string) string {
	switch issueType {
	case "bus_delay":
		return "Bus Delay"
	case "maintenance_issue":
		return "Maintenance Issue"
	case "flat_wheel":
		return "Flat Wheel"
	case "passenger_complaint":
		return "Passenger Complaint"
	case "safety_concern":
		return "Safety Concern"
	case "equipment_malfunction":
		return "Equipment Malfunction"
	default:
		return strings.Title(strings.ReplaceAll(issueType, "_", " "))
	}
}

func getMediaDisplay(imageURL *string) string {
	if imageURL != nil && *imageURL != "" {
		return "Image"
	}
	return "Empty"
}

func formatDateTime(dateTime string) string {
	t, err := time.Parse(time.RFC3339, dateTime)
	if err != nil {
		return dateTime
	}
	return t.Format("2006-01-02 15:04")
}

// getAssignedTeamFromEscalation gets the team from escalation or uses category-based assignment
func getAssignedTeamFromEscalation(escalation *models.ComplaintEscalation, category string, issueType string) string {
	if escalation != nil && escalation.CurrentTeam != "" {
		return escalation.CurrentTeam
	}

	// Get team from escalation config (Level 1 by default)
	assignment := config.GetAssignmentForCategory(category, 1)
	if assignment != nil {
		return assignment.TeamName
	}

	// Fallback to old mapping if config not found
	switch issueType {
	case "bus_delay":
		return "Operations Team"
	case "maintenance_issue", "flat_wheel", "equipment_malfunction":
		return "Maintenance Team"
	case "passenger_complaint":
		return "Customer Service"
	case "safety_concern":
		return "Safety & Security"
	default:
		return "General Support"
	}
}

func getResolverName(resolverName *string) string {
	if resolverName != nil && *resolverName != "" {
		return *resolverName
	}
	return ""
}

func formatActivity(resolutionNotes *string) string {
	if resolutionNotes != nil && *resolutionNotes != "" {
		return fmt.Sprintf("\"%s\"", *resolutionNotes)
	}
	return ""
}

func formatStatus(status string) string {
	switch status {
	case "reported":
		return "Pending"
	case "in_progress":
		return "In progress"
	case "resolved":
		return "Resolved"
	case "acknowledged":
		return "Acknowledged"
	case "closed":
		return "Closed"
	default:
		return status
	}
}

// MapIssuetypeToCategory maps issue_type to escalation category
func MapIssuetypeToCategory(issueType string) string {
	switch issueType {
	case "bus_delay", "maintenance_issue", "flat_wheel", "passenger_complaint", "safety_concern", "equipment_malfunction":
		return issueType
	default:
		return "Other"
	}
}
