package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"passport-desk-mvp/internal/assets"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type ReportService struct {
	repo           *repository.ReportRepository
	citizenService *CitizenService      // Needed to get citizen details for certificate
	regService     *RegistrationService // Needed to get registration details
}

func NewReportService(repo *repository.ReportRepository, cs *CitizenService, rs *RegistrationService) *ReportService {
	return &ReportService{
		repo:           repo,
		citizenService: cs,
		regService:     rs,
	}
}

func (s *ReportService) ExportRegisteredCitizens(ctx context.Context, from, to string) (string, error) {
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

	citizens, err := s.repo.ExportRegisteredCitizens(ctx, from, to)
	if err != nil {
		return "", err
	}

	for i, citizen := range citizens {
		cell, _ := excelize.CoordinatesToCellName(i+2, 1)
		f.SetCellValue("Sheet1", cell, citizen.ID)
		cell, _ = excelize.CoordinatesToCellName(i+2, 2)
		f.SetCellValue("Sheet1", cell, citizen.LastName+" "+citizen.FirstName+" "+citizen.MiddleName)
		cell, _ = excelize.CoordinatesToCellName(i+2, 3)
		f.SetCellValue("Sheet1", cell, citizen.BirthDate)
		cell, _ = excelize.CoordinatesToCellName(i+2, 4)
		f.SetCellValue("Sheet1", cell, citizen.ActiveAddress)
		cell, _ = excelize.CoordinatesToCellName(i+2, 5)
		f.SetCellValue("Sheet1", cell, citizen.Phone)
		cell, _ = excelize.CoordinatesToCellName(i+2, 6)
		f.SetCellValue("Sheet1", cell, citizen.Email)
	}

	f.SetActiveSheet(index)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *ReportService) GetStats() (*models.StatsOutput, error) {
	return s.repo.GetStats()
}

func (s *ReportService) formatInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) < 2 {
		return name
	}

	if strings.Contains(name, ".") {
		return name
	}

	lastName := parts[0]
	firstName := parts[1]

	if len(parts) >= 3 {
		middleName := parts[2]
		return fmt.Sprintf("%s.%s. %s", string([]rune(firstName)[0]), string([]rune(middleName)[0]), lastName)
	}
	return fmt.Sprintf("%s. %s", string([]rune(firstName)[0]), lastName)
}

func (s *ReportService) loadFontBytes(assetPath string) ([]byte, error) {
	data, err := assets.Fonts.ReadFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font %s: %w", assetPath, err)
	}
	return data, nil
}

// GenerateRegistrationCertificate generates a PDF certificate for a citizen's active registration
func (s *ReportService) GenerateRegistrationCertificate(ctx context.Context, citizenID int64, opts models.FamilyCertificateOptions) (string, error) {
	// 1. Get Citizen Data
	citizen, err := s.citizenService.GetByID(ctx, citizenID)
	if err != nil {
		return "", fmt.Errorf("failed to get citizen: %w", err)
	}

	// 2. Get Active Registration
	regs, err := s.regService.GetByCitizenID(ctx, citizenID)
	if err != nil {
		return "", fmt.Errorf("failed to get registrations: %w", err)
	}

	var activeReg *models.RegistrationOutput
	for _, r := range regs {
		if r.IsActive {
			activeReg = &r
			break
		}
	}

	if activeReg == nil {
		return "", fmt.Errorf("citizen has no active registration")
	}

	// 1. Load Font Bytes
	regFontBytes, err := s.loadFontBytes("fonts/DejaVuSans.ttf")
	if err != nil {
		return "", err
	}

	// 2. Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("DejaVu", "", regFontBytes)
	pdf.SetFont("DejaVu", "", 16)

	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()
	pdf.CellFormat(0, 10, "ДОВІДКА ПРО РЕЄСТРАЦІЮ МІСЦЯ ПРОЖИВАННЯ", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 10, fmt.Sprintf("Громадянин(ка): %s %s %s", citizen.LastName, citizen.FirstName, citizen.MiddleName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Дата народження: %s", s.formatDateUkr(citizen.BirthDate)), "", 1, "L", false, 0, "")

	address := fmt.Sprintf("%s, %s, буд. %s", activeReg.Settlement, activeReg.Street, activeReg.HouseNumber)
	if activeReg.ApartmentNumber != "" {
		address += ", кв. " + activeReg.ApartmentNumber
	}
	pdf.MultiCell(0, 10, fmt.Sprintf("Адреса проживання: %s", address), "", "L", false)
	pdf.CellFormat(0, 10, fmt.Sprintf("Зареєстрований(а) з: %s", s.formatDateUkr(activeReg.RegistrationDate)), "", 1, "L", false, 0, "")
	pdf.Ln(20)

	// Signature block
	pdf.SetFont("DejaVu", "", 12)
	signatory := s.formatInitials(opts.SignatoryName)
	if signatory == "" {
		signatory = "__________________"
	}
	pdf.CellFormat(0, 7, fmt.Sprintf("%s _______________ %s", opts.SignatoryTitle, signatory), "", 1, "L", false, 0, "")
	pdf.SetFont("DejaVu", "", 8)
	pdf.CellFormat(0, 5, "                               (підпис) (ініціали та прізвище)", "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", err
	}

	// Return base64 encoded PDF
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GenerateFamilyStatusCertificate generates a PDF with a list of citizens and selected columns
func (s *ReportService) GenerateFamilyStatusCertificate(ctx context.Context, opts models.FamilyCertificateOptions) (string, error) {
	if len(opts.CitizenIDs) == 0 {
		return "", fmt.Errorf("no citizens selected")
	}

	// 1. Get Citizens Data
	var citizens []*models.CitizenOutput
	for _, id := range opts.CitizenIDs {
		c, err := s.citizenService.GetByID(ctx, id)
		if err != nil {
			return "", fmt.Errorf("failed to get citizen %d: %w", id, err)
		}
		citizens = append(citizens, c)
	}

	// First citizen is the primary one
	primary := citizens[0]

	// 2. Load Font Bytes
	regFontBytes, err := s.loadFontBytes("fonts/DejaVuSans.ttf")
	if err != nil {
		return "", err
	}
	boldFontBytes, err := s.loadFontBytes("fonts/DejaVuSans-Bold.ttf")
	if err != nil {
		return "", err
	}

	// 3. Generate PDF (Portrait)
	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddUTF8FontFromBytes("DejaVu", "", regFontBytes)
	pdf.AddUTF8FontFromBytes("DejaVu", "B", boldFontBytes)

	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// Header
	pdf.SetFont("DejaVu", "B", 14) // Use Bold for header
	pdf.CellFormat(0, 10, "ДОВІДКА", "", 1, "C", false, 0, "")
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 5, "про склад сім’ї", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Date and City
	now := time.Now()
	ukrMonths := []string{"", "січня", "лютого", "березня", "квітня", "травня", "червня", "липня", "серпня", "вересня", "жовтня", "листопада", "грудня"}
	dateStr := fmt.Sprintf("“%d” %s %d р.", now.Day(), ukrMonths[now.Month()], now.Year())
	cityStr := opts.City
	if cityStr == "" {
		cityStr = "м. Київ" // Default
	}

	pdf.CellFormat(0, 10, fmt.Sprintf("%s %s", cityStr, dateStr), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Main body
	issuer := opts.IssuerName
	if issuer == "" {
		issuer = "____________________"
	}

	// Get active registration address for primary citizen
	address := "____________________"
	regs, _ := s.regService.GetByCitizenID(ctx, primary.ID)
	for _, r := range regs {
		if r.IsActive {
			address = fmt.Sprintf("%s, %s, буд. %s", r.Settlement, r.Street, r.HouseNumber)
			if r.ApartmentNumber != "" {
				address += ", кв. " + r.ApartmentNumber
			}
			break
		}
	}

	pdf.SetFont("DejaVu", "", 12)
	mainText := fmt.Sprintf("Цією довідкою %s засвідчує, що станом на %s громадянин(ка) %s, який(а) проживає за адресою: %s, має такий склад сім’ї:",
		issuer, dateStr, primary.FullName, address)

	pdf.MultiCell(0, 7, mainText, "", "L", false)
	pdf.Ln(5)

	// Family list
	pdf.SetFont("DejaVu", "", 12)
	for i, c := range citizens {
		// Try to find relation type
		relation := "головний"
		if i > 0 {
			// We need to fetch family members to get relation type.
			// But GenerateFamilyStatusCertificate only receives IDs.
			// Let's fetch primary's members and match.
			mems, _ := s.citizenService.GetFamilyMembers(ctx, primary.ID)
			for _, m := range mems {
				if m.ID == c.ID {
					relation = m.RelationType
					break
				}
			}
		}

		// Format: 1. Прізвище Ім'я По батькові, ступінь споріднення, ДД.ММ.РРРР р.н.
		// Template: прізвище, ім’я, по батькові, дата народження, ступінь споріднення
		rowText := fmt.Sprintf("%d. %s, %s, %s р.н.",
			i+1, c.FullName, relation, s.formatDateUkr(c.BirthDate))
		pdf.MultiCell(0, 7, rowText, "", "L", false)
	}
	pdf.Ln(10)

	// Footer text
	pdf.SetFont("DejaVu", "", 11)
	target := opts.TargetInstitution
	if target == "" {
		target = "____________________"
	}
	purpose := opts.Purpose
	if purpose == "" {
		purpose = "____________________"
	}

	footerText := fmt.Sprintf("Ця довідка видана на підставі статей 3, 6 Закону України “Про свободу пересування та вільний вибір місця проживання в Україні” для подання до %s з метою %s.",
		target, purpose)
	pdf.MultiCell(0, 7, footerText, "", "L", false)
	pdf.Ln(15)

	// Signature
	pdf.SetFont("DejaVu", "", 12)
	signatory := s.formatInitials(opts.SignatoryName)
	if signatory == "" {
		signatory = "__________________"
	}
	pdf.CellFormat(0, 7, fmt.Sprintf("%s _______________ %s", opts.SignatoryTitle, signatory), "", 1, "L", false, 0, "")
	pdf.SetFont("DejaVu", "", 8)
	pdf.CellFormat(0, 5, "                               (підпис) (ініціали та прізвище)", "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *ReportService) formatDateUkr(d string) string {
	// Assumes YYYY-MM-DD
	parts := strings.Split(d, "-")
	if len(parts) != 3 {
		return d
	}
	return fmt.Sprintf("%s.%s.%s", parts[2], parts[1], parts[0])
}

func (s *ReportService) getColumnLabel(col string) string {
	switch col {
	case "birth_date":
		return "Дата народж."
	case "passport":
		return "Паспорт"
	case "tax_number":
		return "ІПН"
	case "gender":
		return "Стать"
	case "address":
		return "Адреса"
	case "phone":
		return "Телефон"
	default:
		return col
	}
}

func (s *ReportService) getColumnValue(ctx context.Context, c *models.CitizenOutput, col string) string {
	switch col {
	case "birth_date":
		return c.BirthDate
	case "passport":
		if c.PassportType == "new" {
			return c.PassportNumber
		}
		return fmt.Sprintf("%s %s", c.PassportSeries, c.PassportNumber)
	case "tax_number":
		return c.TaxNumber
	case "gender":
		return c.GenderDisplay
	case "address":
		// Get active registration address
		regs, err := s.regService.GetByCitizenID(ctx, c.ID)
		if err != nil || len(regs) == 0 {
			return "-"
		}
		for _, r := range regs {
			if r.IsActive {
				return fmt.Sprintf("%s, %s, %s %s", r.Settlement, r.Street, r.HouseNumber, r.ApartmentNumber)
			}
		}
		return "-"
	case "phone":
		return c.Phone
	default:
		return ""
	}
}
