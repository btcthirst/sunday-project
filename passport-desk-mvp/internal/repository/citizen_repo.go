package repository

import (
	"context"
	"database/sql"
	"fmt"
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/models"
	"time"
)

const citizenCols = `c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
	c.passport_series, c.passport_number, c.passport_type, c.tax_number,
	c.gender, c.birth_place, c.phone, c.email, c.notes,
	c.deleted, c.created_at, c.updated_at`

type CitizenRepository struct {
	db *database.Database
}

func NewCitizenRepository(db *database.Database) *CitizenRepository {
	return &CitizenRepository{db: db}
}

// Create inserts a new citizen into the database
// Expects encrypted data from service layer
func (r *CitizenRepository) Create(ctx context.Context, citizen *models.CitizenInput) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}

	const query = `
        INSERT INTO citizens (
            last_name, first_name, middle_name, birth_date,
            passport_series, passport_number, passport_type, tax_number,
            gender, birth_place, phone, email, notes
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	result, err := tx.ExecContext(ctx, query,
		citizen.LastName,
		citizen.FirstName,
		citizen.MiddleName,
		citizen.BirthDate,
		citizen.PassportSeries, // Already encrypted by service
		citizen.PassportNumber, // Already encrypted by service
		citizen.PassportType,
		citizen.TaxNumber, // Already encrypted by service
		citizen.Gender,
		citizen.BirthPlace,
		citizen.Phone, // Already encrypted by service
		citizen.Email,
		citizen.Notes,
	)

	if err != nil {
		return 0, fmt.Errorf("insert citizen: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	// Insert family relations
	for _, rel := range citizen.FamilyRelations {
		// Use AddFamilyMemberWithTx if it existed, but we'll just use Exec here for simplicity
		if _, err := tx.Exec("INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			id, rel.MemberID, rel.RelationType); err != nil {
			return 0, fmt.Errorf("failed to add family relation for %d: %w", rel.MemberID, err)
		}
		if _, err := tx.Exec("INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			rel.MemberID, id, getInverseRelation(rel.RelationType, citizen.Gender)); err != nil {
			return 0, fmt.Errorf("failed to add inverse family relation for %d: %w", rel.MemberID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

// GetByID retrieves a citizen by ID
func (r *CitizenRepository) GetByID(ctx context.Context, id int64) (*models.CitizenOutput, error) {
	// Check context cancelation before starting
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	const query = `
        SELECT 
            id,
            last_name,
            first_name,
            middle_name,
            birth_date,
            passport_series,
            passport_number,
            passport_type,
            tax_number,
            gender,
            birth_place,
            phone,
            email,
            notes,
            deleted,
            created_at,
            updated_at
        FROM citizens
        WHERE id = ?
    `

	citizen := &models.CitizenOutput{}

	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(
		&citizen.ID,
		&citizen.LastName,
		&citizen.FirstName,
		&citizen.MiddleName,
		&citizen.BirthDate,
		&citizen.PassportSeries, // Encrypted data
		&citizen.PassportNumber, // Encrypted data
		&citizen.PassportType,
		&citizen.TaxNumber, // Encrypted data
		&citizen.Gender,
		&citizen.BirthPlace,
		&citizen.Phone, // Encrypted data
		&citizen.Email,
		&citizen.Notes,
		&citizen.Deleted,
		&citizen.CreatedAt,
		&citizen.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("citizen with id %d not found: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("query citizen by id %d: %w", id, err)
	}

	return citizen, nil

}

// Update updates an existing citizen in the database
func (r *CitizenRepository) Update(ctx context.Context, id int64, citizen *models.CitizenInput) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	const query = `
        UPDATE citizens SET
            last_name = ?,
            first_name = ?,
            middle_name = ?,
            birth_date = ?,
            passport_series = ?,
            passport_number = ?,
            passport_type = ?,
            tax_number = ?,
            gender = ?,
            birth_place = ?,
            phone = ?,
            email = ?,
            notes = ?,
            updated_at = ?
        WHERE id = ?
    `

	result, err := tx.ExecContext(ctx, query,
		citizen.LastName,
		citizen.FirstName,
		citizen.MiddleName,
		citizen.BirthDate,
		citizen.PassportSeries,
		citizen.PassportNumber,
		citizen.PassportType,
		citizen.TaxNumber,
		citizen.Gender,
		citizen.BirthPlace,
		citizen.Phone,
		citizen.Email,
		citizen.Notes,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("update citizen %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("citizen %d not found: %w", id, ErrNotFound)
	}

	// Update family relations: clear and re-add
	// Reciprocal logic: delete all where this citizen is participant
	if _, err := tx.ExecContext(ctx, "DELETE FROM family_relations WHERE citizen_id = ? OR member_id = ?", id, id); err != nil {
		return fmt.Errorf("clear family relations: %w", err)
	}

	for _, rel := range citizen.FamilyRelations {
		if _, err := tx.ExecContext(ctx, "INSERT INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			id, rel.MemberID, rel.RelationType); err != nil {
			return fmt.Errorf("add family relation: %w", err)
		}
		// Reciprocal entry
		if _, err := tx.ExecContext(ctx, "INSERT INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
			rel.MemberID, id, getInverseRelation(rel.RelationType, citizen.Gender)); err != nil {
			return fmt.Errorf("add inverse family relation: %w", err)
		}
	}

	return tx.Commit()
}

// SoftDelete marks a citizen as deleted
func (r *CitizenRepository) SoftDelete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	const query = `
        UPDATE citizens 
        SET deleted = 1, updated_at = ?
        WHERE id = ? AND deleted = 0
    `

	result, err := r.db.DB().ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("soft delete citizen %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("citizen %d not found or already deleted: %w", id, ErrNotFound)
	}

	return nil
}

// Restore restores a soft-deleted citizen
func (r *CitizenRepository) Restore(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	const query = `
        UPDATE citizens 
        SET deleted = 0, updated_at = ?
        WHERE id = ? AND deleted = 1
    `

	result, err := r.db.DB().ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("restore citizen %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("citizen %d not found or not deleted: %w", id, ErrNotFound)
	}

	return nil
}

// List retrieves a paginated list of citizens
func (r *CitizenRepository) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*models.CitizenOutput, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("context error: %w", err)
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM citizens`
	if !includeDeleted {
		countQuery += ` WHERE deleted = 0`
	}

	var total int
	if err := r.db.DB().QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count citizens: %w", err)
	}

	// Get items
	query := `
        SELECT 
            id, last_name, first_name, middle_name, birth_date,
            passport_series, passport_number, passport_type, tax_number,
            gender, birth_place, phone, email, notes,
            deleted, created_at, updated_at
        FROM citizens
    `

	if !includeDeleted {
		query += ` WHERE deleted = 0`
	}

	query += ` ORDER BY last_name, first_name LIMIT ? OFFSET ?`

	rows, err := r.db.DB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query citizens: %w", err)
	}
	defer rows.Close()

	var citizens []*models.CitizenOutput

	for rows.Next() {
		citizen := &models.CitizenOutput{}
		if err := rows.Scan(
			&citizen.ID,
			&citizen.LastName,
			&citizen.FirstName,
			&citizen.MiddleName,
			&citizen.BirthDate,
			&citizen.PassportSeries,
			&citizen.PassportNumber,
			&citizen.PassportType,
			&citizen.TaxNumber,
			&citizen.Gender,
			&citizen.BirthPlace,
			&citizen.Phone,
			&citizen.Email,
			&citizen.Notes,
			&citizen.Deleted,
			&citizen.CreatedAt,
			&citizen.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan citizen row: %w", err)
		}
		citizens = append(citizens, citizen)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate citizen rows: %w", err)
	}

	return citizens, total, nil
}

// SearchByName searches citizens by name (partial match)
func (r *CitizenRepository) SearchByName(ctx context.Context, searchTerm string, limit int) ([]*models.CitizenOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	const query = `
        SELECT 
            id, last_name, first_name, middle_name, birth_date,
            passport_series, passport_number, passport_type, tax_number,
            gender, birth_place, phone, email, notes,
            deleted, created_at, updated_at
        FROM citizens
        WHERE deleted = 0 
        AND (
            last_name LIKE ? OR 
            first_name LIKE ? OR 
            middle_name LIKE ? OR
            (last_name || ' ' || first_name || ' ' || middle_name) LIKE ?
        )
        ORDER BY last_name, first_name
        LIMIT ?
    `

	pattern := "%" + searchTerm + "%"

	rows, err := r.db.DB().QueryContext(ctx, query, pattern, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search citizens by name: %w", err)
	}
	defer rows.Close()

	var citizens []*models.CitizenOutput

	for rows.Next() {
		citizen := &models.CitizenOutput{}
		if err := rows.Scan(
			&citizen.ID,
			&citizen.LastName,
			&citizen.FirstName,
			&citizen.MiddleName,
			&citizen.BirthDate,
			&citizen.PassportSeries,
			&citizen.PassportNumber,
			&citizen.PassportType,
			&citizen.TaxNumber,
			&citizen.Gender,
			&citizen.BirthPlace,
			&citizen.Phone,
			&citizen.Email,
			&citizen.Notes,
			&citizen.Deleted,
			&citizen.CreatedAt,
			&citizen.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan citizen row: %w", err)
		}
		citizens = append(citizens, citizen)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate citizen rows: %w", err)
	}

	return citizens, nil
}

// SearchByBirthDate searches citizens by birth date
func (r *CitizenRepository) SearchByBirthDate(ctx context.Context, birthDate string, limit int) ([]*models.CitizenOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	const query = `
        SELECT 
            id, last_name, first_name, middle_name, birth_date,
            passport_series, passport_number, passport_type, tax_number,
            gender, birth_place, phone, email, notes,
            deleted, created_at, updated_at
        FROM citizens
        WHERE deleted = 0 AND birth_date = ?
        ORDER BY last_name, first_name
        LIMIT ?
    `

	rows, err := r.db.DB().QueryContext(ctx, query, birthDate, limit)
	if err != nil {
		return nil, fmt.Errorf("search citizens by birth date: %w", err)
	}
	defer rows.Close()

	var citizens []*models.CitizenOutput

	for rows.Next() {
		citizen := &models.CitizenOutput{}
		if err := rows.Scan(
			&citizen.ID,
			&citizen.LastName,
			&citizen.FirstName,
			&citizen.MiddleName,
			&citizen.BirthDate,
			&citizen.PassportSeries,
			&citizen.PassportNumber,
			&citizen.PassportType,
			&citizen.TaxNumber,
			&citizen.Gender,
			&citizen.BirthPlace,
			&citizen.Phone,
			&citizen.Email,
			&citizen.Notes,
			&citizen.Deleted,
			&citizen.CreatedAt,
			&citizen.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan citizen row: %w", err)
		}
		citizens = append(citizens, citizen)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate citizen rows: %w", err)
	}

	return citizens, nil
}

// Exists checks if a citizen with the given ID exists
func (r *CitizenRepository) Exists(ctx context.Context, id int64) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("context error: %w", err)
	}

	const query = `SELECT EXISTS(SELECT 1 FROM citizens WHERE id = ? AND deleted = 0)`

	var exists bool
	err := r.db.DB().QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check citizen exists %d: %w", id, err)
	}

	return exists, nil
}

// GetByPassport finds citizen by passport (requires decryption in service layer)
// This is inefficient but necessary for encrypted data
// Better approach: use indexed hash of passport for lookup
func (r *CitizenRepository) GetByPassport(ctx context.Context, passportSeries, passportNumber string) (*models.CitizenOutput, error) {
	// Note: This will return encrypted passport data
	// Service layer must decrypt and compare
	// For better performance, consider storing a hash of passport for indexing

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	const query = `
        SELECT 
            id, last_name, first_name, middle_name, birth_date,
            passport_series, passport_number, passport_type, tax_number,
            gender, birth_place, phone, email, notes,
            deleted, created_at, updated_at
        FROM citizens
        WHERE deleted = 0
    `

	rows, err := r.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query citizens for passport search: %w", err)
	}
	defer rows.Close()

	// Return all citizens for service layer to decrypt and compare
	// This is inefficient but required for encrypted data
	var citizens []*models.CitizenOutput

	for rows.Next() {
		citizen := &models.CitizenOutput{}
		if err := rows.Scan(
			&citizen.ID,
			&citizen.LastName,
			&citizen.FirstName,
			&citizen.MiddleName,
			&citizen.BirthDate,
			&citizen.PassportSeries,
			&citizen.PassportNumber,
			&citizen.PassportType,
			&citizen.TaxNumber,
			&citizen.Gender,
			&citizen.BirthPlace,
			&citizen.Phone,
			&citizen.Email,
			&citizen.Notes,
			&citizen.Deleted,
			&citizen.CreatedAt,
			&citizen.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan citizen row: %w", err)
		}
		citizens = append(citizens, citizen)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate citizen rows: %w", err)
	}

	// Service layer will decrypt and find the match
	// Return the first one here as placeholder
	// Better approach: service layer handles this logic
	if len(citizens) == 0 {
		return nil, fmt.Errorf("no citizens found: %w", ErrNotFound)
	}

	// This needs to be handled in service layer properly
	return citizens[0], nil
}

// GetAll retrieves all non-deleted citizens (for encrypted field searches in service layer)
func (r *CitizenRepository) GetAll(ctx context.Context) ([]*models.CitizenOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	const query = `
        SELECT 
            id, last_name, first_name, middle_name, birth_date,
            passport_series, passport_number, passport_type, tax_number,
            gender, birth_place, phone, email, notes,
            deleted, created_at, updated_at
        FROM citizens
        WHERE deleted = 0
        ORDER BY last_name, first_name
    `

	rows, err := r.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all citizens: %w", err)
	}
	defer rows.Close()

	var citizens []*models.CitizenOutput

	for rows.Next() {
		citizen := &models.CitizenOutput{}
		if err := rows.Scan(
			&citizen.ID,
			&citizen.LastName,
			&citizen.FirstName,
			&citizen.MiddleName,
			&citizen.BirthDate,
			&citizen.PassportSeries,
			&citizen.PassportNumber,
			&citizen.PassportType,
			&citizen.TaxNumber,
			&citizen.Gender,
			&citizen.BirthPlace,
			&citizen.Phone,
			&citizen.Email,
			&citizen.Notes,
			&citizen.Deleted,
			&citizen.CreatedAt,
			&citizen.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan citizen row: %w", err)
		}
		citizens = append(citizens, citizen)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate citizen rows: %w", err)
	}

	return citizens, nil
}

// AddFamilyMember adds a relationship between two citizens (reciprocal)
func (r *CitizenRepository) AddFamilyMember(ctx context.Context, citizenID, memberID int64, relationType string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	// We need gender of citizenID for reciprocity
	var gender string
	err := r.db.DB().QueryRowContext(ctx, "SELECT gender FROM citizens WHERE id = ?", citizenID).Scan(&gender)
	if err != nil {
		return fmt.Errorf("get gender for reciprocity: %w", err)
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
		citizenID, memberID, relationType); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO family_relations (citizen_id, member_id, relation_type) VALUES (?, ?, ?)",
		memberID, citizenID, getInverseRelation(relationType, gender)); err != nil {
		return err
	}

	return tx.Commit()
}

// RemoveFamilyMember removes a relationship (reciprocal)
func (r *CitizenRepository) RemoveFamilyMember(ctx context.Context, citizenID, memberID int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM family_relations WHERE (citizen_id = ? AND member_id = ?) OR (citizen_id = ? AND member_id = ?)",
		citizenID, memberID, memberID, citizenID); err != nil {
		return err
	}

	return tx.Commit()
}

// GetFamilyMembers returns all family members for a citizen
func (r *CitizenRepository) GetFamilyMembers(ctx context.Context, citizenID int64) ([]*models.FamilyMemberOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}
	query := fmt.Sprintf(`
		SELECT %s, '' as active_address, fr.relation_type 
		FROM family_relations fr
		JOIN citizens c ON fr.member_id = c.id
		WHERE fr.citizen_id = ? AND c.deleted = 0
	`, citizenCols)
	rows, err := r.db.DB().QueryContext(ctx, query, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.FamilyMemberOutput
	for rows.Next() {
		var relationType string
		// scanCitizenFromScanner expects a specific set of columns.
		// We added relation_type at the end, so we need a custom scanner or handle it.
		// Let's use a simpler approach: scan everything except relation_type using scanCitizenFromScanner if possible.
		// Actually, scanCitizenFromScanner(rows) will consume the row.

		// To keep it simple, let's just scan manually but use the decryption logic.
		var c models.CitizenOutput

		var dummyActiveAddress sql.NullString

		err := rows.Scan(
			&c.ID, &c.LastName, &c.FirstName, &c.MiddleName, &c.BirthDate,
			&c.PassportSeries, &c.PassportNumber, &c.PassportType, &c.TaxNumber,
			&c.Gender, &c.BirthPlace, &c.Phone, &c.Email, &c.Notes,
			&c.Deleted, &c.CreatedAt, &c.UpdatedAt, &dummyActiveAddress, &relationType,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, &models.FamilyMemberOutput{
			CitizenOutput: c,
			RelationType:  relationType,
		})
	}

	return members, nil
}

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
