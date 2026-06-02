package config

// Assignment represents the team assignment for each escalation level
type Assignment struct {
	TeamName    string
	RoleName    string
	PhoneNumber string
}

// CategoryAssignment maps complaint categories to escalation levels and their assignments
var CategoryAssignment = map[string]map[int]Assignment{
	"Service Issue": {
		1: {"Customer Service", "Support Agent", "94715342627"},
		2: {"Customer Service", "Team Lead", "94715342627"},
		3: {"Customer Service", "Manager", "94715342627"},
	},
	"Operations & Scheduling": {
		1: {"Operations Team", "Operations Officer", "94715342627"},
		2: {"Operations Team", "Operations Manager", "94715342627"},
		3: {"Operations Team", "Operations Director", "94715342627"},
	},
	"Vehicle & Facility": {
		1: {"Maintenance Team", "Maintenance Technician", "94715342627"},
		2: {"Maintenance Team", "Maintenance Supervisor", "94715342627"},
		3: {"Maintenance Team", "Maintenance Manager", "94715342627"},
	},
	"Safety & Security": {
		1: {"Safety & Security", "Security Officer", "94715342627"},
		2: {"Safety & Security", "Security Manager", "94715342627"},
		3: {"Safety & Security", "Security Director", "94715342627"},
	},
	"Other": {
		1: {"General Support", "Support Agent", "94715342627"},
		2: {"General Support", "Support Manager", "94715342627"},
		3: {"General Support", "General Manager", "94715342627"},
	},
}

// GetAssignmentForCategory returns the assignment for a specific category and level
func GetAssignmentForCategory(category string, level int) *Assignment {
	categoryAssignments, exists := CategoryAssignment[category]
	if !exists {
		// Default to "Other" if category not found
		categoryAssignments = CategoryAssignment["Other"]
	}

	assignment, exists := categoryAssignments[level]
	if !exists {
		// Return the highest level if requested level doesn't exist
		maxLevel := 0
		for lvl := range categoryAssignments {
			if lvl > maxLevel {
				maxLevel = lvl
			}
		}
		if maxLevel > 0 {
			assignment = categoryAssignments[maxLevel]
		} else {
			return nil
		}
	}

	return &assignment
}

// GetMaxLevelForCategory returns the maximum escalation level for a category
func GetMaxLevelForCategory(category string) int {
	categoryAssignments, exists := CategoryAssignment[category]
	if !exists {
		categoryAssignments = CategoryAssignment["Other"]
	}

	maxLevel := 0
	for level := range categoryAssignments {
		if level > maxLevel {
			maxLevel = level
		}
	}
	return maxLevel
}
