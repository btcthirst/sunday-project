package repository

import (
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/models"
)

type OperatorRepository struct {
	db *database.Database
}

func NewOperatorRepository(db *database.Database) *OperatorRepository {
	return &OperatorRepository{db: db}
}

// CreateOperator creates a new operator
func (d *OperatorRepository) CreateOperator(op *models.Operator) error {
	result, err := d.db.DB().Exec(
		`INSERT INTO operators (username, password_hash, full_name) VALUES (?, ?, ?)`,
		op.Username, op.PasswordHash, op.FullName,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	op.ID = id
	return nil
}

// GetOperatorByUsername retrieves an operator by username
func (d *OperatorRepository) GetOperatorByUsername(username string) (*models.Operator, error) {
	op := &models.Operator{}
	err := d.db.DB().QueryRow(
		`SELECT id, username, password_hash, full_name, created_at, updated_at FROM operators WHERE username = ?`,
		username,
	).Scan(&op.ID, &op.Username, &op.PasswordHash, &op.FullName, &op.CreatedAt, &op.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return op, nil
}

// OperatorExists checks if any operators exist in the database
func (d *OperatorRepository) OperatorExists() (bool, error) {
	var count int
	err := d.db.DB().QueryRow(`SELECT COUNT(*) FROM operators`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
