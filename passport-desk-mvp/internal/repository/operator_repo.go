package repository

import (
	"context"
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
func (d *OperatorRepository) Create(ctx context.Context, op *models.Operator) error {
	result, err := d.db.DB().ExecContext(ctx,
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
func (d *OperatorRepository) GetByUsername(ctx context.Context, username string) (*models.Operator, error) {
	op := &models.Operator{}
	err := d.db.DB().QueryRowContext(ctx,
		`SELECT id, username, password_hash, full_name, created_at, updated_at FROM operators WHERE username = ?`,
		username,
	).Scan(&op.ID, &op.Username, &op.PasswordHash, &op.FullName, &op.CreatedAt, &op.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return op, nil
}

// OperatorExists checks if any operators exist in the database
func (d *OperatorRepository) Exists(ctx context.Context) (bool, error) {
	var count int
	err := d.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM operators`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAll retrieves all operators
func (d *OperatorRepository) GetAll(ctx context.Context) ([]*models.Operator, error) {
	rows, err := d.db.DB().QueryContext(ctx, `SELECT id, username, password_hash, full_name, created_at, updated_at FROM operators`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var operators []*models.Operator
	for rows.Next() {
		op := &models.Operator{}
		if err := rows.Scan(&op.ID, &op.Username, &op.PasswordHash, &op.FullName, &op.CreatedAt, &op.UpdatedAt); err != nil {
			return nil, err
		}
		operators = append(operators, op)
	}
	return operators, nil
}

// Update updates an operator's data (e.g. password hash)
func (d *OperatorRepository) Update(ctx context.Context, op *models.Operator) error {
	_, err := d.db.DB().ExecContext(ctx,
		`UPDATE operators SET username = ?, password_hash = ?, full_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		op.Username, op.PasswordHash, op.FullName, op.ID,
	)
	return err
}
