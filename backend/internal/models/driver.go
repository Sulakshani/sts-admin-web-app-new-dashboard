package models

type Driver struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`           // mapped from emergency_contact_name
	ContactNumber      string `json:"contact_number"` // mapped from emergency_contact
	LicenseNumber      string `json:"license_number"`
	LicenseExpiryDate  string `json:"license_expiry_date"`
	ExperienceYears    int    `json:"experience_years"`
	VerificationStatus string `json:"verification_status"`
	VerificationNotes  string `json:"verification_notes"`
	Status             string `json:"status"` // from employment_status
	HireDate           string `json:"hire_date"`
	CreatedAt          string `json:"created_at,omitempty"`
}
