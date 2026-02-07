package repository

import (
	"context"
	"errors"
	"fmt"
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/models"
	"strings"
	"time"
)

type RegistrationRepository struct {
	db *database.Database
}

func NewRegistrationRepository(db *database.Database) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// Create adds a new registration.
// If type is 'permanent', it automatically deactivates previous active permanent registrations.
func (r *RegistrationRepository) Create(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error) {
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
	tx, err := r.db.DB().Begin()
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

	return r.GetByID(ctx, id)
}

func (r *RegistrationRepository) GetByID(ctx context.Context, id int64) (*models.RegistrationOutput, error) {
	var reg models.RegistrationOutput
	row := r.db.DB().QueryRow(`
		SELECT id, citizen_id, registration_type, region, district, settlement,
		       street, house_number, apartment_number, registration_date, 
		       COALESCE(deregistration_date, ''), COALESCE(basis_document, ''), is_active,
		       created_at, updated_at
		FROM registrations WHERE id = ?
	`, id)

	err := row.Scan(
		&reg.ID, &reg.CitizenID, &reg.RegistrationType, &reg.Region, &reg.District, &reg.Settlement,
		&reg.Street, &reg.HouseNumber, &reg.ApartmentNumber, &reg.RegistrationDate,
		&reg.DeregistrationDate, &reg.BasisDocument, &reg.IsActive,
		&reg.CreatedAt, &reg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *RegistrationRepository) GetByCitizenID(ctx context.Context, citizenID int64) ([]models.RegistrationOutput, error) {
	rows, err := r.db.DB().Query(`
		SELECT id, citizen_id, registration_type, region, district, settlement,
		       street, house_number, apartment_number, registration_date, 
		       COALESCE(deregistration_date, ''), COALESCE(basis_document, ''), is_active
		FROM registrations WHERE citizen_id = ? ORDER BY registration_date DESC
	`, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.RegistrationOutput
	for rows.Next() {
		var reg models.RegistrationOutput
		err := rows.Scan(
			&reg.ID, &reg.CitizenID, &reg.RegistrationType, &reg.Region, &reg.District, &reg.Settlement,
			&reg.Street, &reg.HouseNumber, &reg.ApartmentNumber, &reg.RegistrationDate,
			&reg.DeregistrationDate, &reg.BasisDocument, &reg.IsActive,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, reg)
	}
	return results, nil
}

func (r *RegistrationRepository) Deregister(ctx context.Context, id int64, deregistrationDate string) error {
	if deregistrationDate == "" {
		deregistrationDate = time.Now().Format("2006-01-02")
	}

	_, err := r.db.DB().Exec(`
		UPDATE registrations 
		SET is_active = 0, deregistration_date = ? 
		WHERE id = ?
	`, deregistrationDate, id)

	return err
}
func (r *RegistrationRepository) ListAll(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error) {
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

	var items []models.RegistrationListItem
	var total int
	var totalPages int

	if search != "" {
		// When searching, we must decrypt encrypted fields to verify matches.
		// For simplicity/accuracy, we fetch all relevant (by status) items and filter in Go.
		sqlQuery := `SELECT 
			r.id, r.citizen_id, r.registration_type, r.region, r.district, r.settlement,
			r.street, r.house_number, r.apartment_number, r.registration_date, 
			COALESCE(r.deregistration_date, ''), COALESCE(r.basis_document, ''), r.is_active,
			r.created_at, r.updated_at,
			c.last_name, c.first_name, c.middle_name, c.birth_date, c.phone, c.tax_number` +
			queryBase + " ORDER BY r.registration_date DESC"

		rows, err := r.db.DB().Query(sqlQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var matchedItems []models.RegistrationListItem
		searchLower := strings.ToLower(search)
		for rows.Next() {
			var item models.RegistrationListItem
			var lName, fName, mName string
			err := rows.Scan(
				&item.ID, &item.CitizenID, &item.RegistrationType, &item.Region, &item.District, &item.Settlement,
				&item.Street, &item.HouseNumber, &item.ApartmentNumber, &item.RegistrationDate,
				&item.DeregistrationDate, &item.BasisDocument, &item.IsActive,
				&item.CreatedAt, &item.UpdatedAt,
				&lName, &fName, &mName, &item.CitizenBirth, &item.CitizenPhone, &item.CitizenTax,
			)
			if err != nil {
				return nil, err
			}
			item.CitizenName = fmt.Sprintf("%s %s %s", lName, fName, mName)

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
		err := r.db.DB().QueryRow("SELECT COUNT(*) "+queryBase, args...).Scan(&total)
		if err != nil {
			return nil, err
		}

		sqlQuery := `SELECT 
			r.id, r.citizen_id, r.registration_type, r.region, r.district, r.settlement,
			r.street, r.house_number, r.apartment_number, r.registration_date, 
			COALESCE(r.deregistration_date, ''), COALESCE(r.basis_document, ''), r.is_active,
			r.created_at, r.updated_at,
			c.last_name, c.first_name, c.middle_name, c.birth_date, c.phone, c.tax_number` +
			queryBase + " ORDER BY r.registration_date DESC LIMIT ? OFFSET ?"

		args = append(args, limit, offset)

		rows, err := r.db.DB().Query(sqlQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var item models.RegistrationListItem
			var lName, fName, mName string
			err := rows.Scan(
				&item.ID, &item.CitizenID, &item.RegistrationType, &item.Region, &item.District, &item.Settlement,
				&item.Street, &item.HouseNumber, &item.ApartmentNumber, &item.RegistrationDate,
				&item.DeregistrationDate, &item.BasisDocument, &item.IsActive,
				&item.CreatedAt, &item.UpdatedAt,
				&lName, &fName, &mName, &item.CitizenBirth, &item.CitizenPhone, &item.CitizenTax,
			)
			if err != nil {
				return nil, err
			}
			item.CitizenName = fmt.Sprintf("%s %s %s", lName, fName, mName)
			items = append(items, item)
		}
		totalPages = (total + limit - 1) / limit
	}

	return &models.RegistrationListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
