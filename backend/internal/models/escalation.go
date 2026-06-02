package models

import "time"

// EscalationLevel represents a level in the escalation hierarchy
type EscalationLevel struct {
	Level            int      `json:"level"`
	TeamName         string   `json:"team_name"`
	AssignedTo       []string `json:"assigned_to"`        // Admin user IDs or emails
	EscalationDays   int      `json:"escalation_days"`    // Days before escalating to next level
	NotificationEmails []string `json:"notification_emails,omitempty"`
}

// EscalationConfig defines the escalation configuration for each category
type EscalationConfig struct {
	Category string             `json:"category"`
	Levels   []EscalationLevel  `json:"levels"`
}

// ComplaintEscalation tracks the escalation status of a complaint
type ComplaintEscalation struct {
	CurrentLevel      int       `json:"current_level"`
	CurrentTeam       string    `json:"current_team"`
	SourceApp         string    `json:"source_app"`
	AssignedToAdminID *string   `json:"assigned_to_admin_id,omitempty"`
	PreviousAssignedAdminID *string `json:"previous_assigned_admin_id,omitempty"`
	AssignedToName    string    `json:"assigned_to_name,omitempty"`
	LastEscalatedAt   *time.Time `json:"last_escalated_at,omitempty"`
	NextEscalationDue *time.Time `json:"next_escalation_due,omitempty"`
	EscalationHistory []EscalationHistoryEntry `json:"escalation_history,omitempty"`
}

// EscalationHistoryEntry tracks each escalation event
type EscalationHistoryEntry struct {
	Level        int       `json:"level"`
	TeamName     string    `json:"team_name"`
	EscalatedAt  time.Time `json:"escalated_at"`
	EscalatedBy  string    `json:"escalated_by"` // "auto" or admin user ID
	FromAdminID  *string   `json:"from_admin_id,omitempty"`
	ToAdminID    *string   `json:"to_admin_id,omitempty"`
	Reason       string    `json:"reason"`
}

// GetEscalationConfigs returns the escalation configuration for all categories
func GetEscalationConfigs() map[string]EscalationConfig {
	return map[string]EscalationConfig{
		"bus_delay": {
			Category: "bus_delay",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "Operations Support Team",
					AssignedTo:     []string{}, // To be filled with admin IDs
					EscalationDays: 5,
					NotificationEmails: []string{"operations@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Operations Manager",
					AssignedTo:     []string{},
					EscalationDays: 3,
					NotificationEmails: []string{"operations.manager@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Senior Management",
					AssignedTo:     []string{},
					EscalationDays: 0, // Final level
					NotificationEmails: []string{"management@aasl.lk"},
				},
			},
		},
		"maintenance_issue": {
			Category: "maintenance_issue",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "Maintenance Team",
					AssignedTo:     []string{},
					EscalationDays: 5,
					NotificationEmails: []string{"maintenance@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Maintenance Supervisor",
					AssignedTo:     []string{},
					EscalationDays: 3,
					NotificationEmails: []string{"maintenance.supervisor@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Technical Manager",
					AssignedTo:     []string{},
					EscalationDays: 0,
					NotificationEmails: []string{"technical.manager@aasl.lk"},
				},
			},
		},
		"flat_wheel": {
			Category: "flat_wheel",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "Maintenance Team",
					AssignedTo:     []string{},
					EscalationDays: 5,
					NotificationEmails: []string{"maintenance@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Fleet Supervisor",
					AssignedTo:     []string{},
					EscalationDays: 2,
					NotificationEmails: []string{"fleet.supervisor@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Technical Manager",
					AssignedTo:     []string{},
					EscalationDays: 0,
					NotificationEmails: []string{"technical.manager@aasl.lk"},
				},
			},
		},
		"passenger_complaint": {
			Category: "passenger_complaint",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "Customer Service Team",
					AssignedTo:     []string{},
					EscalationDays: 5,
					NotificationEmails: []string{"customerservice@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Customer Service Manager",
					AssignedTo:     []string{},
					EscalationDays: 3,
					NotificationEmails: []string{"cs.manager@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Head of Customer Experience",
					AssignedTo:     []string{},
					EscalationDays: 0,
					NotificationEmails: []string{"customer.experience@aasl.lk"},
				},
			},
		},
		"safety_concern": {
			Category: "safety_concern",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "Safety & Compliance Team",
					AssignedTo:     []string{},
					EscalationDays: 5,
					NotificationEmails: []string{"safety@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Safety Officer",
					AssignedTo:     []string{},
					EscalationDays: 2,
					NotificationEmails: []string{"safety.officer@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Head of Safety & Compliance",
					AssignedTo:     []string{},
					EscalationDays: 0,
					NotificationEmails: []string{"head.safety@aasl.lk"},
				},
			},
		},
		"equipment_malfunction": {
			Category: "equipment_malfunction",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "Technical Support Team",
					AssignedTo:     []string{},
					EscalationDays: 5,
					NotificationEmails: []string{"techsupport@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Technical Supervisor",
					AssignedTo:     []string{},
					EscalationDays: 3,
					NotificationEmails: []string{"tech.supervisor@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Technical Manager",
					AssignedTo:     []string{},
					EscalationDays: 0,
					NotificationEmails: []string{"technical.manager@aasl.lk"},
				},
			},
		},
		"other": {
			Category: "other",
			Levels: []EscalationLevel{
				{
					Level:          1,
					TeamName:       "General Support Team",
					AssignedTo:     []string{},
					EscalationDays: 5,
					NotificationEmails: []string{"support@aasl.lk"},
				},
				{
					Level:          2,
					TeamName:       "Support Manager",
					AssignedTo:     []string{},
					EscalationDays: 3,
					NotificationEmails: []string{"support.manager@aasl.lk"},
				},
				{
					Level:          3,
					TeamName:       "Senior Management",
					AssignedTo:     []string{},
					EscalationDays: 0,
					NotificationEmails: []string{"management@aasl.lk"},
				},
			},
		},
	}
}

// GetEscalationConfigForCategory returns the escalation config for a specific category
func GetEscalationConfigForCategory(category string) *EscalationConfig {
	configs := GetEscalationConfigs()
	if config, exists := configs[category]; exists {
		return &config
	}
	// Default to "other" if category not found
	if config, exists := configs["other"]; exists {
		return &config
	}
	return nil
}
