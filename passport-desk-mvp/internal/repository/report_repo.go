package repository

import (
	"context"
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

	// Query data (simplified for now, ideally strictly filtered by date)
	// We need a Join query here usually, or fetch all and filter.
	// For MVP, let's fetch list and filter in memory or use a new query in DB.
	// We will use existing List Citizens to get a batch (not efficient for export)
	// or add a dedicated method in `CitizenService` or just query DB directly here.
	// Direct DB query is better for reports.

	query := `
		SELECT c.id, c.last_name, c.first_name, c.middle_name, c.birth_date,
		       r.region, r.settlement, r.street, r.house_number, r.registration_date, r.registration_type
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
			return nil, err
		}

		citizens = append(citizens, *citizen)
	}

	return citizens, nil
}

// GetStats returns dashboard statistics
func (s *ReportRepository) GetStats() (*models.StatsOutput, error) {
	stats := &models.StatsOutput{}

	// Total Citizens
	if err := s.db.DB().QueryRow("SELECT COUNT(*) FROM citizens WHERE deleted = 0").Scan(&stats.TotalCitizens); err != nil {
		return nil, err
	}

	// Total Registrations
	if err := s.db.DB().QueryRow("SELECT COUNT(*) FROM registrations").Scan(&stats.TotalRegistrations); err != nil {
		return nil, err
	}

	// Active Registrations
	if err := s.db.DB().QueryRow("SELECT COUNT(*) FROM registrations WHERE is_active = 1").Scan(&stats.ActiveRegistrations); err != nil {
		return nil, err
	}

	// New This Month
	startOfMonth := time.Now().Format("2006-01") + "-01"
	if err := s.db.DB().QueryRow("SELECT COUNT(*) FROM registrations WHERE registration_date >= ?", startOfMonth).Scan(&stats.NewThisMonth); err != nil {
		return nil, err
	}

	return stats, nil
}
