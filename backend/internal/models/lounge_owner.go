package models

type LoungeOwner struct {
	ID                 string `json:"id"`
	UserID             string `json:"user_id"`
	ManagerFullName    string `json:"manager_full_name"`
	Email              string `json:"email"`
	ContactNumber      string `json:"contact_number"`
	NIC                string `json:"nic"`
	BusinessName       string `json:"business_name"`
	BusinessLicense    string `json:"business_license"`
	VerificationStatus string `json:"verification_status"`
	VerificationNotes  string `json:"verification_notes"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
