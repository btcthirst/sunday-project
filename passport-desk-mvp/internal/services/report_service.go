package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"passport-desk-mvp/internal/database"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type ReportService struct {
	db             *database.Database
	citizenService *CitizenService      // Needed to get citizen details for certificate
	regService     *RegistrationService // Needed to get registration details
}

func NewReportService(db *database.Database, cs *CitizenService, rs *RegistrationService) *ReportService {
	return &ReportService{
		db:             db,
		citizenService: cs,
		regService:     rs,
	}
}

type StatsOutput struct {
	TotalCitizens       int `json:"total_citizens"`
	TotalRegistrations  int `json:"total_registrations"`
	ActiveRegistrations int `json:"active_registrations"`
	NewThisMonth        int `json:"new_this_month"`
}

// GenerateRegistrationCertificate generates a PDF certificate for a citizen's active registration
func (s *ReportService) GenerateRegistrationCertificate(citizenID int64) (string, error) {
	// 1. Get Citizen Data
	citizen, err := s.citizenService.GetByID(citizenID)
	if err != nil {
		return "", fmt.Errorf("failed to get citizen: %w", err)
	}

	// 2. Get Active Registration
	regs, err := s.regService.GetByCitizenID(citizenID)
	if err != nil {
		return "", fmt.Errorf("failed to get registrations: %w", err)
	}

	var activeReg *database.RegistrationOutput
	for _, r := range regs {
		if r.IsActive {
			activeReg = &r
			break
		}
	}

	if activeReg == nil {
		return "", fmt.Errorf("citizen has no active registration")
	}

	// 3. Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Use system font for Cyrillic support
	// We must copy it to a writable directory because gofpdf generates .z and .json files
	sourceFont := "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	destFont := filepath.Join(os.TempDir(), "DejaVuSans.ttf")

	if _, err := os.Stat(destFont); os.IsNotExist(err) {
		// Copy file
		input, err := os.ReadFile(sourceFont)
		if err != nil {
			return "", fmt.Errorf("failed to read font: %w", err)
		}
		if err := os.WriteFile(destFont, input, 0644); err != nil {
			return "", fmt.Errorf("failed to copy font to temp: %w", err)
		}
	}

	pdf.AddUTF8Font("DejaVu", "", destFont)
	pdf.SetFont("DejaVu", "", 16)

	pdf.AddPage()
	pdf.Cell(40, 10, "Довідка про реєстрацію / Registration Certificate")
	pdf.Ln(10)

	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Громадянин / Citizen: %s %s %s", citizen.LastName, citizen.FirstName, citizen.MiddleName))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Дата народження / Date of Birth: %s", citizen.BirthDate))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Адреса / Address: %s, %s, %s, %s", activeReg.Region, activeReg.Settlement, activeReg.Street, activeReg.HouseNumber))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Дата реєстрації / Registration Date: %s", activeReg.RegistrationDate))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", err
	}

	// Return base64 encoded PDF
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// ExportRegisteredCitizens generates an Excel file with registered citizens
func (s *ReportService) ExportRegisteredCitizens(from, to string) (string, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Create a new sheet.
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		return "", err
	}

	// Set headers
	headers := []string{"ID", "ПІБ", "Дата народження", "Адреса", "Дата реєстрації", "Тип"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, h)
	}

	// Query data (simplified for now, ideally strictly filtered by date)
	// We need a Join query here usually, or fetch all and filter.
	// For MVP, let's fetch list and filter in memory or use a new query in DB.
	// We will use existing ListCitizens to get a batch (not efficient for export)
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

	rows, err := s.db.DB().Query(query, from, to)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	rowIdx := 2
	for rows.Next() {
		var id int64
		var ln, fn, mn, bd, reg, set, str, hn, rdate, rtype string
		if err := rows.Scan(&id, &ln, &fn, &mn, &bd, &reg, &set, &str, &hn, &rdate, &rtype); err != nil {
			continue
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowIdx), id)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("%s %s %s", ln, fn, mn))
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowIdx), bd)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowIdx), fmt.Sprintf("%s, %s, %s", set, str, hn))
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", rowIdx), rdate)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", rowIdx), rtype)
		rowIdx++
	}

	f.SetActiveSheet(index)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GetStats returns dashboard statistics
func (s *ReportService) GetStats() (*StatsOutput, error) {
	stats := &StatsOutput{}

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
