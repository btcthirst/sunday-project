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

const citizenCols = `c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
	c.passport_series, c.passport_number, c.passport_type, c.tax_number,
	c.gender, c.birth_place, c.phone, c.email, c.notes,
	c.deleted, c.created_at, c.updated_at`

func getInverseRelation(rel string, gender string) string {
	isMale := gender == "M"

	switch rel {
	case "Батько", "Мати":
		if isMale {
			return "Син"
		}
		return "Дочка"
	case "Син", "Дочка":
		if isMale {
			return "Батько"
		}
		return "Мати"
	case "Чоловік":
		return "Дружина"
	case "Дружина":
		return "Чоловік"
	case "Брат", "Сестра":
		if isMale {
			return "Брат"
		}
		return "Сестра"
	case "Дідусь", "Бабуся":
		if isMale {
			return "Онук"
		}
		return "Онука"
	case "Онук", "Онука":
		if isMale {
			return "Дідусь"
		}
		return "Бабуся"
	default:
		return rel
	}
}

// NewCitizenService creates a new citizen service
func NewCitizenService(db *database.Database, crypto *security.Crypto) *CitizenService {
	return &CitizenService{db: db, crypto: crypto}
}

// FamilyRelationInput represents input for a family relationship
type FamilyRelationInput struct {
	MemberID     int64  `json:"member_id"`
	RelationType string `json:"relation_type"`
}

// CitizenInput is the input structure for creating/updating citizens
type CitizenInput struct {
	LastName        string                `json:"last_name"`
	FirstName       string                `json:"first_name"`
	MiddleName      string                `json:"middle_name"`
	BirthDate       string                `json:"birth_date"` // YYYY-MM-DD
	PassportSeries  string                `json:"passport_series"`
	PassportNumber  string                `json:"passport_number"`
	PassportType    string                `json:"passport_type"`
	TaxNumber       string                `json:"tax_number"`
	Gender          string                `json:"gender"` // M or F
	BirthPlace      string                `json:"birth_place"`
	Phone           string                `json:"phone"`
	Email           string                `json:"email"`
	Notes           string                `json:"notes"`
	FamilyRelations []FamilyRelationInput `json:"family_relations"`
}

// FamilyMemberOutput represents a family member with their details
type FamilyMemberOutput struct {
	CitizenOutput
	RelationType string `json:"relation_type"`
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

// Create adds a new citizen with encrypted fields
func (s *CitizenService) Create(input *CitizenInput) (*CitizenOutput, error) {
	tx, err := s.db.DB().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	encPassportSeries, _ := s.crypto.Encrypt(input.PassportSeries)
	encPassportNumber, _ := s.crypto.Encrypt(input.PassportNumber)
	encTaxNumber, _ := s.crypto.Encrypt(input.TaxNumber)
	encPhone, _ := s.crypto.Encrypt(input.Phone)

	result, err := tx.Exec(`
		INSERT INTO citizens (
			last_name, first_name, middle_name, birth_date,
			passport_series, passport_number, passport_type, tax_number,
			gender, birth_place, phone, email, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.LastName, input.FirstName, input.MiddleName, normalizeDate(input.BirthDate),
		encPassportSeries, encPassportNumber, input.PassportType, encTaxNumber,
		input.Gender, input.BirthPlace, encPhone, input.Email, input.Notes)

	if err != nil {
		return nil, fmt.Errorf("failed to insert citizen: %w", err)
	}

	id, _ := result.LastInsertId()

	// Insert family relations
	for _, rel := range input.FamilyRelations {
		// Use AddFamilyMemberWithTx if it existed, but we'll just use Exec here for simplicity
		if _, err := tx.Exec("INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			id, rel.MemberID, rel.RelationType); err != nil {
			return nil, fmt.Errorf("failed to add family relation for %d: %w", rel.MemberID, err)
		}
		if _, err := tx.Exec("INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			rel.MemberID, id, getInverseRelation(rel.RelationType, input.Gender)); err != nil {
			return nil, fmt.Errorf("failed to add inverse family relation for %d: %w", rel.MemberID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// GetByID retrieves a citizen by ID
func (s *CitizenService) GetByID(id int64) (*CitizenOutput, error) {
	row := s.db.DB().QueryRow(fmt.Sprintf(`
		SELECT %s,
			(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
		FROM citizens c
		LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
		WHERE c.id = ?
	`, citizenCols), id)

	return s.scanCitizen(row)
}

// Update updates citizen details
func (s *CitizenService) Update(id int64, input *CitizenInput) error {
	tx, err := s.db.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	encPassportSeries, _ := s.crypto.Encrypt(input.PassportSeries)
	encPassportNumber, _ := s.crypto.Encrypt(input.PassportNumber)
	encTaxNumber, _ := s.crypto.Encrypt(input.TaxNumber)
	encPhone, _ := s.crypto.Encrypt(input.Phone)

	_, err = tx.Exec(`
		UPDATE citizens SET
			last_name = ?, first_name = ?, middle_name = ?, birth_date = ?,
			passport_series = ?, passport_number = ?, passport_type = ?, tax_number = ?,
			gender = ?, birth_place = ?, phone = ?, email = ?, notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, input.LastName, input.FirstName, input.MiddleName, normalizeDate(input.BirthDate),
		encPassportSeries, encPassportNumber, input.PassportType, encTaxNumber,
		input.Gender, input.BirthPlace, encPhone, input.Email, input.Notes, id)

	if err != nil {
		return err
	}

	// Update family relations (simplest way: delete and re-insert)
	// We only delete relations FROM this citizen.
	if _, err := tx.Exec("DELETE FROM family_relations WHERE citizen_id = ?", id); err != nil {
		return fmt.Errorf("failed to clear old family relations: %w", err)
	}

	for _, rel := range input.FamilyRelations {
		if _, err := tx.Exec("INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			id, rel.MemberID, rel.RelationType); err != nil {
			return fmt.Errorf("failed to update family relation for %d: %w", rel.MemberID, err)
		}
		if _, err := tx.Exec("INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			rel.MemberID, id, getInverseRelation(rel.RelationType, input.Gender)); err != nil {
			return fmt.Errorf("failed to update inverse family relation for %d: %w", rel.MemberID, err)
		}
	}

	return tx.Commit()
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
	query := fmt.Sprintf(`
		SELECT %s,
			(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
		FROM citizens c
		LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
	`, citizenCols)
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
		sqlQuery = fmt.Sprintf(`
			SELECT %s,
				(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
			FROM citizens c
			LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
`, citizenCols) + `
			WHERE c.deleted = 0 AND (
				c.last_name LIKE ? OR c.first_name LIKE ? OR c.middle_name LIKE ?
				OR (c.last_name || ' ' || c.first_name || ' ' || c.middle_name) LIKE ?
			)
			ORDER BY c.last_name, c.first_name
			LIMIT 50
		`
		args = []interface{}{pattern, pattern, pattern, pattern}

	case "birth_date":
		sqlQuery = fmt.Sprintf(`
			SELECT %s,
				(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
			FROM citizens c
			LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
`, citizenCols) + `
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
	rows, err := s.db.DB().Query(fmt.Sprintf(`
		SELECT %s,
			(r.settlement || ', ' || r.street || ' ' || r.house_number) as active_address
		FROM citizens c
		LEFT JOIN registrations r ON c.id = r.citizen_id AND r.is_active = 1
		WHERE c.deleted = 0
	`, citizenCols))
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
	c.BirthDate = normalizeDate(c.BirthDate)

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

// AddFamilyMember adds a relationship between two citizens
func (s *CitizenService) AddFamilyMember(citizenID, memberID int64, relationType string) error {
	_, err := s.db.DB().Exec(
		"INSERT INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
		citizenID, memberID, relationType,
	)
	return err
}

// RemoveFamilyMember removes a relationship
func (s *CitizenService) RemoveFamilyMember(citizenID, memberID int64) error {
	_, err := s.db.DB().Exec(
		"DELETE FROM family_relations WHERE citizen_id = ? AND member_id = ?",
		citizenID, memberID,
	)
	return err
}

// GetFamilyMembers returns all family members for a citizen
func (s *CitizenService) GetFamilyMembers(citizenID int64) ([]FamilyMemberOutput, error) {
	query := fmt.Sprintf(`
		SELECT %s, '' as active_address, fr.relation_type 
		FROM family_relations fr
		JOIN citizens c ON fr.member_id = c.id
		WHERE fr.citizen_id = ? AND c.deleted = 0
	`, citizenCols)
	rows, err := s.db.DB().Query(query, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []FamilyMemberOutput
	for rows.Next() {
		var relationType string
		// scanCitizenFromScanner expects a specific set of columns.
		// We added relation_type at the end, so we need a custom scanner or handle it.
		// Let's use a simpler approach: scan everything except relation_type using scanCitizenFromScanner if possible.
		// Actually, scanCitizenFromScanner(rows) will consume the row.

		// To keep it simple, let's just scan manually but use the decryption logic.
		var c CitizenOutput
		var encPassportSeries, encPassportNumber, encTaxNumber, encPhone string
		var createdAt, updatedAt time.Time
		var dummyActiveAddress sql.NullString

		err := rows.Scan(
			&c.ID, &c.LastName, &c.FirstName, &c.MiddleName, &c.BirthDate,
			&encPassportSeries, &encPassportNumber, &c.PassportType, &encTaxNumber,
			&c.Gender, &c.BirthPlace, &encPhone, &c.Email, &c.Notes,
			&c.Deleted, &createdAt, &updatedAt, &dummyActiveAddress, &relationType,
		)
		if err != nil {
			return nil, err
		}

		c.BirthDate = normalizeDate(c.BirthDate)

		// Decrypt sensitive fields
		c.PassportSeries, _ = s.crypto.Decrypt(encPassportSeries)
		c.PassportNumber, _ = s.crypto.Decrypt(encPassportNumber)
		c.TaxNumber, _ = s.crypto.Decrypt(encTaxNumber)
		c.Phone, _ = s.crypto.Decrypt(encPhone)

		// Computed fields
		c.FullName = strings.TrimSpace(c.LastName + " " + c.FirstName + " " + c.MiddleName)
		c.PassportMasked = maskPassport(c.PassportSeries, c.PassportNumber)
		c.TaxNumberMasked = maskTaxNumber(c.TaxNumber)
		c.GenderDisplay = genderToUkrainian(c.Gender)
		c.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		c.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

		members = append(members, FamilyMemberOutput{
			CitizenOutput: c,
			RelationType:  relationType,
		})
	}

	return members, nil
}

// normalizeDate strips time from ISO date string
func normalizeDate(d string) string {
	if idx := strings.Index(d, "T"); idx != -1 {
		return d[:idx]
	}
	return d
}
