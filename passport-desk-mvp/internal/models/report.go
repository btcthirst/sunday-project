package models

type StatsOutput struct {
	TotalCitizens       int `json:"total_citizens"`
	TotalRegistrations  int `json:"total_registrations"`
	ActiveRegistrations int `json:"active_registrations"`
	NewThisMonth        int `json:"new_this_month"`
}

type FamilyCertificateOptions struct {
	CitizenIDs        []int64 `json:"citizen_ids"`
	IssuerName        string  `json:"issuer_name"`
	TargetInstitution string  `json:"target_institution"`
	Purpose           string  `json:"purpose"`
	SignatoryTitle    string  `json:"signatory_title"`
	SignatoryName     string  `json:"signatory_name"`
	City              string  `json:"city"`
}
