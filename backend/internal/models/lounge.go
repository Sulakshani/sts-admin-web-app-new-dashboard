package models

type Lounge struct {
	LoungeID         string   `json:"lounge_id"`
	LoungeOwner      string   `json:"lounge_owner"`
	OwnerNIC         string   `json:"owner_nic,omitempty"`
	OwnerEmail       string   `json:"owner_email,omitempty"`
	OwnerContact     string   `json:"owner_contact,omitempty"`
	LoungeName       string   `json:"lounge_name"`
	LoungeContact    string   `json:"lounge_contact"`
	Address          string   `json:"address"`
	Capacity         int      `json:"capacity"`
	PricePerHour     float64  `json:"price_per_hour"`
	Facilities       []string `json:"facilities"`
	Marketplace      string   `json:"marketplace"`
	Verification     string   `json:"verification"`
	VerificationNote string   `json:"verification_note"`
	Operational      bool     `json:"operational"`
	CreatedAt        string   `json:"created_at,omitempty"`
}
