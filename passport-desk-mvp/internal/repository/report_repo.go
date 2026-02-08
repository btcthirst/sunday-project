package repository

import (
	"context"
	"fmt"
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/models"
	"time"
)

type ReportRepository struct {
	db *database.Database
}

func NewReportRepository(db *database.Database) *ReportRepository {
	return &ReportRepository{db: db}
}

// ExportRegisteredCitizens generates an Excel file with registered citizens
func (s *ReportRepository) ExportRegisteredCitizens(ctx context.Context, from, to string) ([]models.CitizenOutput, error) {

	// Join with citizens to get full info for report but only select what we need or map correctly.
	// The previous query was selecting some columns but trying to scan into full CitizenOutput.
	// We should update the query to select all necessary columns to fill CitizenOutput as much as possible,
	// because the return type is []models.CitizenOutput.

	query := `
		SELECT 
			c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
			c.passport_series, c.passport_number, c.passport_type, c.tax_number,
			c.gender, c.birth_place, c.phone, c.email, c.notes,
			c.deleted, c.created_at, c.updated_at,
			r.region, r.settlement, r.street, r.house_number, r.registration_date
		FROM registrations r
		JOIN citizens c ON r.citizen_id = c.id
		WHERE r.registration_date >= ? AND r.registration_date <= ?
	`
	// Handle empty dates (all time)
	if from == "" {
		from = "1900-01-01"
	}
	if to == "" {
		to = "2100-01-01"
	}

	rows, err := s.db.DB().QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	citizens := []models.CitizenOutput{}
	for rows.Next() {
		citizen := &models.CitizenOutput{}
		var region, settlement, street, houseNumber, regDate string

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
			&region, &settlement, &street, &houseNumber, &regDate,
		); err != nil {
			return nil, err
		}

		// Construct address from registration data
		citizen.ActiveAddress = fmt.Sprintf("%s, %s, буд. %s", settlement, street, houseNumber)

		citizens = append(citizens, *citizen)
	}

	return citizens, nil
}

// GetStats returns dashboard statistics
func (s *ReportRepository) GetStats(ctx context.Context) (*models.StatsOutput, error) {
	stats := &models.StatsOutput{}

	// Total Citizens
	if err := s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM citizens WHERE deleted = 0").Scan(&stats.TotalCitizens); err != nil {
		return nil, err
	}

	// Total Registrations
	if err := s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM registrations").Scan(&stats.TotalRegistrations); err != nil {
		return nil, err
	}

	// Active Registrations
	if err := s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM registrations WHERE is_active = 1").Scan(&stats.ActiveRegistrations); err != nil {
		return nil, err
	}

	// New This Month
	startOfMonth := time.Now().Format("2006-01") + "-01"
	if err := s.db.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM registrations WHERE registration_date >= ?", startOfMonth).Scan(&stats.NewThisMonth); err != nil {
		return nil, err
	}

	return stats, nil
}
