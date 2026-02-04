package services

import (
	"errors"
	"fmt"
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/security"
	"strings"
	"time"
)

type RegistrationListItem struct {
	database.RegistrationOutput
	CitizenName  string `json:"citizen_name"`
	CitizenBirth string `json:"citizen_birth"`
	CitizenPhone string `json:"citizen_phone"`
	CitizenTax   string `json:"citizen_tax"`
}

type RegistrationListResult struct {
	Items      []RegistrationListItem `json:"items"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

type RegistrationService struct {
	db     *database.Database
	crypto *security.Crypto
}

func NewRegistrationService(db *database.Database, crypto *security.Crypto) *RegistrationService {
	return &RegistrationService{db: db, crypto: crypto}
}

// Create adds a new registration.
// If type is 'permanent', it automatically deactivates previous active permanent registrations.
func (s *RegistrationService) Create(input *database.RegistrationInput) (*database.RegistrationOutput, error) {
	if input.CitizenID == 0 {
		return nil, errors.New("citizen_id is required")
	}
	if input.RegistrationType != "permanent" && input.RegistrationType != "temporary" {
		return nil, errors.New("invalid registration type")
	}
	if input.RegistrationDate == "" {
		// Default to today if not provided
		input.RegistrationDate = time.Now().Format("2006-01-02")
	}

	// Transaction to handle auto-deregistration
	tx, err := s.db.DB().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// If permanent, deregister previous active permanent registrations
	if input.RegistrationType == "permanent" {
		_, err := tx.Exec(`
			UPDATE registrations 
			SET is_active = 0, deregistration_date = ? 
			WHERE citizen_id = ? AND registration_type = 'permanent' AND is_active = 1
		`, input.RegistrationDate, input.CitizenID)
		if err != nil {
			return nil, fmt.Errorf("failed to deactivate previous registrations: %w", err)
		}
	}

	// Insert new registration
	res, err := tx.Exec(`
		INSERT INTO registrations (
			citizen_id, registration_type, region, district, settlement, 
			street, house_number, apartment_number, registration_date, basis_document, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
	`, input.CitizenID, input.RegistrationType, input.Region, input.District, input.Settlement,
		input.Street, input.HouseNumber, input.ApartmentNumber, input.RegistrationDate, input.BasisDocument)

	if err != nil {
		return nil, fmt.Errorf("failed to create registration: %w", err)
	}

	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

func (s *RegistrationService) GetByID(id int64) (*database.RegistrationOutput, error) {
	var r database.RegistrationOutput
	row := s.db.DB().QueryRow(`
		SELECT id, citizen_id, registration_type, region, district, settlement,
		       street, house_number, apartment_number, registration_date, 
		       COALESCE(deregistration_date, ''), COALESCE(basis_document, ''), is_active
		FROM registrations WHERE id = ?
	`, id)

	err := row.Scan(
		&r.ID, &r.CitizenID, &r.RegistrationType, &r.Region, &r.District, &r.Settlement,
		&r.Street, &r.HouseNumber, &r.ApartmentNumber, &r.RegistrationDate,
		&r.DeregistrationDate, &r.BasisDocument, &r.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *RegistrationService) GetByCitizenID(citizenID int64) ([]database.RegistrationOutput, error) {
	rows, err := s.db.DB().Query(`
		SELECT id, citizen_id, registration_type, region, district, settlement,
		       street, house_number, apartment_number, registration_date, 
		       COALESCE(deregistration_date, ''), COALESCE(basis_document, ''), is_active
		FROM registrations WHERE citizen_id = ? ORDER BY registration_date DESC
	`, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []database.RegistrationOutput
	for rows.Next() {
		var r database.RegistrationOutput
		err := rows.Scan(
			&r.ID, &r.CitizenID, &r.RegistrationType, &r.Region, &r.District, &r.Settlement,
			&r.Street, &r.HouseNumber, &r.ApartmentNumber, &r.RegistrationDate,
			&r.DeregistrationDate, &r.BasisDocument, &r.IsActive,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (s *RegistrationService) Deregister(id int64, deregistrationDate string) error {
	if deregistrationDate == "" {
		deregistrationDate = time.Now().Format("2006-01-02")
	}

	_, err := s.db.DB().Exec(`
		UPDATE registrations 
		SET is_active = 0, deregistration_date = ? 
		WHERE id = ?
	`, deregistrationDate, id)

	return err
}
func (s *RegistrationService) ListAll(search string, isActive *bool, page, limit int) (*RegistrationListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var whereClauses []string
	var args []interface{}

	if isActive != nil {
		if *isActive {
			whereClauses = append(whereClauses, "r.is_active = 1")
		} else {
			whereClauses = append(whereClauses, "r.is_active = 0")
		}
	}

	// We'll join with citizens for name search
	queryBase := ` FROM registrations r
		JOIN citizens c ON r.citizen_id = c.id
		WHERE c.deleted = 0`

	for _, clause := range whereClauses {
		queryBase += " AND " + clause
	}

	var items []RegistrationListItem
	var total int
	var totalPages int

	if search != "" {
		// When searching, we must decrypt encrypted fields to verify matches.
		// For simplicity/accuracy, we fetch all relevant (by status) items and filter in Go.
		sqlQuery := `SELECT 
			r.id, r.citizen_id, r.registration_type, r.region, r.district, r.settlement,
			r.street, r.house_number, r.apartment_number, r.registration_date, 
			COALESCE(r.deregistration_date, ''), COALESCE(r.basis_document, ''), r.is_active,
			c.last_name, c.first_name, c.middle_name, c.birth_date, c.phone, c.tax_number` +
			queryBase + " ORDER BY r.registration_date DESC"

		rows, err := s.db.DB().Query(sqlQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var matchedItems []RegistrationListItem
		searchLower := strings.ToLower(search)
		for rows.Next() {
			var item RegistrationListItem
			var lName, fName, mName, encPhone, encTax string
			err := rows.Scan(
				&item.ID, &item.CitizenID, &item.RegistrationType, &item.Region, &item.District, &item.Settlement,
				&item.Street, &item.HouseNumber, &item.ApartmentNumber, &item.RegistrationDate,
				&item.DeregistrationDate, &item.BasisDocument, &item.IsActive,
				&lName, &fName, &mName, &item.CitizenBirth, &encPhone, &encTax,
			)
			if err != nil {
				return nil, err
			}
			item.CitizenName = fmt.Sprintf("%s %s %s", lName, fName, mName)
			item.CitizenPhone, _ = s.crypto.Decrypt(encPhone)
			item.CitizenTax, _ = s.crypto.Decrypt(encTax)

			// Match logic: Name OR Phone OR TaxNumber
			match := strings.Contains(strings.ToLower(item.CitizenName), searchLower) ||
				strings.Contains(item.CitizenPhone, search) ||
				strings.Contains(item.CitizenTax, search)

			if match {
				matchedItems = append(matchedItems, item)
			}
		}

		total = len(matchedItems)
		totalPages = (total + limit - 1) / limit

		// Apply pagination manually
		start := offset
		if start < total {
			end := start + limit
			if end > total {
				end = total
			}
			items = matchedItems[start:end]
		}
	} else {
		// No search: use SQL pagination
		err := s.db.DB().QueryRow("SELECT COUNT(*) "+queryBase, args...).Scan(&total)
		if err != nil {
			return nil, err
		}

		sqlQuery := `SELECT 
			r.id, r.citizen_id, r.registration_type, r.region, r.district, r.settlement,
			r.street, r.house_number, r.apartment_number, r.registration_date, 
			COALESCE(r.deregistration_date, ''), COALESCE(r.basis_document, ''), r.is_active,
			c.last_name, c.first_name, c.middle_name, c.birth_date, c.phone, c.tax_number` +
			queryBase + " ORDER BY r.registration_date DESC LIMIT ? OFFSET ?"

		args = append(args, limit, offset)

		rows, err := s.db.DB().Query(sqlQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var item RegistrationListItem
			var lName, fName, mName, encPhone, encTax string
			err := rows.Scan(
				&item.ID, &item.CitizenID, &item.RegistrationType, &item.Region, &item.District, &item.Settlement,
				&item.Street, &item.HouseNumber, &item.ApartmentNumber, &item.RegistrationDate,
				&item.DeregistrationDate, &item.BasisDocument, &item.IsActive,
				&lName, &fName, &mName, &item.CitizenBirth, &encPhone, &encTax,
			)
			if err != nil {
				return nil, err
			}
			item.CitizenName = fmt.Sprintf("%s %s %s", lName, fName, mName)
			item.CitizenPhone, _ = s.crypto.Decrypt(encPhone)
			item.CitizenTax, _ = s.crypto.Decrypt(encTax)
			items = append(items, item)
		}
		totalPages = (total + limit - 1) / limit
	}

	return &RegistrationListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
