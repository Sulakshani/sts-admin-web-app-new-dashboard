package models

import "database/sql"

type Bus struct {
	ID                        string         `json:"id"`
	BusOwnerID                string         `json:"bus_owner_id,omitempty"`   // For Create/Update
	PermitID                  string         `json:"permit_id,omitempty"`      // For Create/Update
	SeatLayoutID              sql.NullString `json:"seat_layout_id,omitempty"` // For Create/Update
	BusNumber                 string         `json:"bus_number"`
	CompanyName               string         `json:"company_name"`
	IdentifyOrIncorporationNo string         `json:"identify_or_incorporation_no"`
	BusinessEmail             string         `json:"business_email"`
	BusinessPhone             string         `json:"business_phone"`
	PermitNumber              string         `json:"permit_number"`
	LicensePlate              string         `json:"license_plate"`
	TotalSeats                int            `json:"total_seats"`
	BusType                   string         `json:"bus_type"`
	CustomRouteName           string         `json:"custom_route_name"`
	FarePerSeat               float64        `json:"fare_per_seat"`
	Status                    string         `json:"status"`
	VerificationStatus        string         `json:"verification_status"`       // Route permit status
	OwnerVerificationStatus   string         `json:"owner_verification_status"` // Bus owner verification status
	VerificationDocuments     []string       `json:"verification_documents"`    // Array of document URLs/paths
	CreatedAt                 string         `json:"created_at,omitempty"`
}
