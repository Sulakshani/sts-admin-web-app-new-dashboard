package models

type BusOwner struct {
	ID                        string      `json:"id"`
	UserID                    string      `json:"user_id"`
	CompanyName               string      `json:"company_name"`
	LicenseNumber             string      `json:"license_number"`
	ContactPerson             string      `json:"contact_person"`
	Address                   string      `json:"address"`
	City                      string      `json:"city"`
	State                     string      `json:"state"`
	Country                   string      `json:"country"`
	PostalCode                string      `json:"postal_code"`
	VerificationStatus        string      `json:"verification_status"`
	VerificationDocuments     interface{} `json:"verification_documents"`
	BusinessEmail             string      `json:"business_email"`
	BusinessPhone             string      `json:"business_phone"`
	TaxID                     string      `json:"tax_id"`
	BankAccountDetails        interface{} `json:"bank_account_details"`
	TotalBuses                int         `json:"total_buses"`
	ProfileCompleted          bool        `json:"profile_completed"`
	IdentityOrIncorporationNo string      `json:"identity_or_incorporation_no"`
	CreatedAt                 string      `json:"created_at"`
	UpdatedAt                 string      `json:"updated_at"`
}
