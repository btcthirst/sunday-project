package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/security"
)

// CitizenService handles citizen-related operations
type CitizenService struct {
	db     *database.Database
	crypto *security.Crypto
}

// NewCitizenService creates a new citizen service
func NewCitizenService(db *database.Database, crypto *security.Crypto) *CitizenService {
	return &CitizenService{db: db, crypto: crypto}
}

// CitizenInput is the input structure for creating/updating citizens
type CitizenInput struct {
	LastName       string `json:"last_name"`
	FirstName      string `json:"first_name"`
	MiddleName     string `json:"middle_name"`
	BirthDate      string `json:"birth_date"` // YYYY-MM-DD
	PassportSeries string `json:"passport_series"`
	PassportNumber string `json:"passport_number"`
	PassportType   string `json:"passport_type"`
	TaxNumber      string `json:"tax_number"`
	Gender         string `json:"gender"` // M or F
	BirthPlace     string `json:"birth_place"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Notes          string `json:"notes"`
}

// CitizenOutput is the output structure (with decrypted fields)
type CitizenOutput struct {
	ID              int64  `json:"id"`
	LastName        string `json:"last_name"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	FullName        string `json:"full_name"` // Computed
	BirthDate       string `json:"birth_date"`
	PassportSeries  string `json:"passport_series"`
	PassportNumber  string `json:"passport_number"`
	PassportType    string `json:"passport_type"`
	PassportMasked  string `json:"passport_masked"` // For display
	TaxNumber       string `json:"tax_number"`
	TaxNumberMasked string `json:"tax_number_masked"` // For display
	Gender          string `json:"gender"`
	GenderDisplay   string `json:"gender_display"` // Чоловік/Жінка
	BirthPlace      string `json:"birth_place"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Notes           string `json:"notes"`
	Deleted         bool   `json:"deleted"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	ActiveAddress   string `json:"active_address,omitempty"`
}

// CitizenListResult contains paginated list results
type CitizenListResult struct {
	Items      []CitizenOutput `json:"items"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// Create creates a new citizen with encrypted fields
func (s *CitizenService) Create(input *CitizenInput) (*CitizenOutput, error) {
	// Encrypt sensitive fields
	encPassportSeries, err := s.crypto.Encrypt(input.PassportSeries)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt passport series: %w", err)
	}

	encPassportNumber, err := s.crypto.Encrypt(input.PassportNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt passport number: %w", err)
	}

	encTaxNumber, err := s.crypto.Encrypt(input.TaxNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tax number: %w", err)
	}

	encPhone, err := s.crypto.Encrypt(input.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt phone: %w", err)
	}

	result, err := s.db.DB().Exec(`
		INSERT INTO citizens (
			last_name, first_name, middle_name, birth_date,
			passport_series, passport_number, passport_type, tax_number,
			gender, birth_place, phone, email, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.LastName, input.FirstName, input.MiddleName, input.BirthDate,
		encPassportSeries, encPassportNumber, input.PassportType, encTaxNumber,
		input.Gender, input.BirthPlace, encPhone, input.Email, input.Notes)

	if err != nil {
		return nil, fmt.Errorf("failed to insert citizen: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

// GetByID retrieves a citizen by ID
func (s *CitizenService) GetByID(id int64) (*CitizenOutput, error) {
	row := s.db.DB().QueryRow(`
		SELECT c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
			c.passport_series, c.passport_number, c.passport_type, c.tax_number,
			c.gender, c.birth_place, c.phone, c.email, c.notes,
			c.deleted, c.created_at, c.updated_at,
			(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
		FROM citizens c
		LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
		WHERE c.id = ?
	`, id)

	return s.scanCitizen(row)
}

// Update updates a citizen
func (s *CitizenService) Update(id int64, input *CitizenInput) error {
	// Encrypt sensitive fields
	encPassportSeries, err := s.crypto.Encrypt(input.PassportSeries)
	if err != nil {
		return fmt.Errorf("failed to encrypt passport series: %w", err)
	}

	encPassportNumber, err := s.crypto.Encrypt(input.PassportNumber)
	if err != nil {
		return fmt.Errorf("failed to encrypt passport number: %w", err)
	}

	encTaxNumber, err := s.crypto.Encrypt(input.TaxNumber)
	if err != nil {
		return fmt.Errorf("failed to encrypt tax number: %w", err)
	}

	encPhone, err := s.crypto.Encrypt(input.Phone)
	if err != nil {
		return fmt.Errorf("failed to encrypt phone: %w", err)
	}

	_, err = s.db.DB().Exec(`
		UPDATE citizens SET
			last_name = ?, first_name = ?, middle_name = ?, birth_date = ?,
			passport_series = ?, passport_number = ?, passport_type = ?, tax_number = ?,
			gender = ?, birth_place = ?, phone = ?, email = ?, notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, input.LastName, input.FirstName, input.MiddleName, input.BirthDate,
		encPassportSeries, encPassportNumber, input.PassportType, encTaxNumber,
		input.Gender, input.BirthPlace, encPhone, input.Email, input.Notes, id)

	return err
}

// Delete performs soft delete
func (s *CitizenService) Delete(id int64) error {
	_, err := s.db.DB().Exec(`UPDATE citizens SET deleted = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

// Restore restores a soft-deleted citizen
func (s *CitizenService) Restore(id int64) error {
	_, err := s.db.DB().Exec(`UPDATE citizens SET deleted = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

// List returns paginated citizens
func (s *CitizenService) List(page, limit int, includeDeleted bool) (*CitizenListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM citizens`
	if !includeDeleted {
		countQuery += ` WHERE deleted = 0`
	}
	if err := s.db.DB().QueryRow(countQuery).Scan(&total); err != nil {
		return nil, err
	}

	// Get items
	query := `
		SELECT c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
			c.passport_series, c.passport_number, c.passport_type, c.tax_number,
			c.gender, c.birth_place, c.phone, c.email, c.notes,
			c.deleted, c.created_at, c.updated_at,
			(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
		FROM citizens c
		LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
	`
	if !includeDeleted {
		query += ` WHERE c.deleted = 0`
	}
	query += ` ORDER BY c.last_name, c.first_name LIMIT ? OFFSET ?`

	rows, err := s.db.DB().Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CitizenOutput
	for rows.Next() {
		citizen, err := s.scanCitizenFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *citizen)
	}

	totalPages := (total + limit - 1) / limit

	return &CitizenListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// Search searches citizens by various fields
func (s *CitizenService) Search(query string, field string) ([]CitizenOutput, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []CitizenOutput{}, nil
	}

	var sqlQuery string
	var args []interface{}

	switch field {
	case "name":
		// Search by name (partial match)
		pattern := "%" + query + "%"
		sqlQuery = `
			SELECT c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
				c.passport_series, c.passport_number, c.passport_type, c.tax_number,
				c.gender, c.birth_place, c.phone, c.email, c.notes,
				c.deleted, c.created_at, c.updated_at,
				(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
			FROM citizens c
			LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
			WHERE c.deleted = 0 AND (
				c.last_name LIKE ? OR c.first_name LIKE ? OR c.middle_name LIKE ?
				OR (c.last_name || ' ' || c.first_name || ' ' || c.middle_name) LIKE ?
			)
			ORDER BY c.last_name, c.first_name
			LIMIT 50
		`
		args = []interface{}{pattern, pattern, pattern, pattern}

	case "birth_date":
		sqlQuery = `
			SELECT c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
				c.passport_series, c.passport_number, c.passport_type, c.tax_number,
				c.gender, c.birth_place, c.phone, c.email, c.notes,
				c.deleted, c.created_at, c.updated_at,
				(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
			FROM citizens c
			LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
			WHERE c.deleted = 0 AND c.birth_date = ?
			ORDER BY c.last_name, c.first_name
			LIMIT 50
		`
		args = []interface{}{query}

	case "passport", "tax_number":
		// For encrypted fields, we need to decrypt and compare
		// This is less efficient but necessary for encrypted data
		return s.searchEncryptedField(query, field)

	default:
		return nil, fmt.Errorf("unknown search field: %s", field)
	}

	rows, err := s.db.DB().Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CitizenOutput
	for rows.Next() {
		citizen, err := s.scanCitizenFromRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *citizen)
	}

	return results, nil
}

// searchEncryptedField searches through encrypted fields
func (s *CitizenService) searchEncryptedField(query string, field string) ([]CitizenOutput, error) {
	// Get all non-deleted citizens
	rows, err := s.db.DB().Query(`
		SELECT c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
			c.passport_series, c.passport_number, c.passport_type, c.tax_number,
			c.gender, c.birth_place, c.phone, c.email, c.notes,
			c.deleted, c.created_at, c.updated_at,
			(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
		FROM citizens c
		LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
		WHERE c.deleted = 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CitizenOutput
	queryLower := strings.ToLower(query)

	for rows.Next() {
		citizen, err := s.scanCitizenFromRows(rows)
		if err != nil {
			return nil, err
		}

		var match bool
		switch field {
		case "passport":
			passport := strings.ToLower(citizen.PassportSeries + citizen.PassportNumber)
			match = strings.Contains(passport, queryLower)
		case "tax_number":
			match = strings.Contains(strings.ToLower(citizen.TaxNumber), queryLower)
		}

		if match {
			results = append(results, *citizen)
			if len(results) >= 50 {
				break
			}
		}
	}

	return results, nil
}

// Scanner interface for both Row and Rows
type scanner interface {
	Scan(dest ...interface{}) error
}

func (s *CitizenService) scanCitizen(row *sql.Row) (*CitizenOutput, error) {
	return s.scanCitizenFromScanner(row)
}

func (s *CitizenService) scanCitizenFromRows(rows *sql.Rows) (*CitizenOutput, error) {
	return s.scanCitizenFromScanner(rows)
}

func (s *CitizenService) scanCitizenFromScanner(sc scanner) (*CitizenOutput, error) {
	var c CitizenOutput
	var encPassportSeries, encPassportNumber, encTaxNumber, encPhone string
	var createdAt, updatedAt time.Time
	var activeAddress sql.NullString

	err := sc.Scan(
		&c.ID, &c.LastName, &c.FirstName, &c.MiddleName, &c.BirthDate,
		&encPassportSeries, &encPassportNumber, &c.PassportType, &encTaxNumber,
		&c.Gender, &c.BirthPlace, &encPhone, &c.Email, &c.Notes,
		&c.Deleted, &createdAt, &updatedAt, &activeAddress,
	)
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields
	c.PassportSeries, _ = s.crypto.Decrypt(encPassportSeries)
	c.PassportNumber, _ = s.crypto.Decrypt(encPassportNumber)
	c.TaxNumber, _ = s.crypto.Decrypt(encTaxNumber)
	c.Phone, _ = s.crypto.Decrypt(encPhone)

	if activeAddress.Valid {
		c.ActiveAddress = activeAddress.String
	}

	// Computed fields
	c.FullName = strings.TrimSpace(c.LastName + " " + c.FirstName + " " + c.MiddleName)
	c.PassportMasked = maskPassport(c.PassportSeries, c.PassportNumber)
	c.TaxNumberMasked = maskTaxNumber(c.TaxNumber)
	c.GenderDisplay = genderToUkrainian(c.Gender)
	c.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	c.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

	return &c, nil
}

// maskPassport masks passport number for display
func maskPassport(series, number string) string {
	if series == "" && number == "" {
		return ""
	}
	if len(number) <= 2 {
		return series + " " + number
	}
	return series + " " + number[:2] + "****"
}

// maskTaxNumber masks IPN for display
func maskTaxNumber(taxNumber string) string {
	if len(taxNumber) <= 4 {
		return taxNumber
	}
	return taxNumber[:4] + "******"
}

// genderToUkrainian converts gender code to Ukrainian
func genderToUkrainian(gender string) string {
	switch gender {
	case "M":
		return "Чоловік"
	case "F":
		return "Жінка"
	default:
		return ""
	}
}
