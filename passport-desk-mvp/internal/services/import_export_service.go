package services

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type ImportExportService struct {
	citizenService *CitizenService
}

func NewImportExportService(cs *CitizenService) *ImportExportService {
	return &ImportExportService{
		citizenService: cs,
	}
}

// ExportToExcel generates an Excel file with all citizen data
func (s *ImportExportService) ExportToExcel() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Citizens"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"ID", "Прізвище", "Ім'я", "По батькові", "Дата народження",
		"Серія паспорта", "Номер паспорта", "Тип паспорта", "ІПН",
		"Стать", "Місце народження", "Телефон", "Email", "Нотатки",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Fetch all citizens (including deleted if needed, but normally only active)
	result, err := s.citizenService.List(1, 1000000, false)
	if err != nil {
		return nil, err
	}

	for i, c := range result.Items {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), c.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), c.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), c.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), c.MiddleName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), c.BirthDate)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), c.PassportSeries)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), c.PassportNumber)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), c.PassportType)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), c.TaxNumber)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), c.Gender)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", row), c.BirthPlace)
		f.SetCellValue(sheet, fmt.Sprintf("L%d", row), c.Phone)
		f.SetCellValue(sheet, fmt.Sprintf("M%d", row), c.Email)
		f.SetCellValue(sheet, fmt.Sprintf("N%d", row), c.Notes)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportCustomToExcel generates an Excel file for specific citizens and columns
func (s *ImportExportService) ExportCustomToExcel(ids []int64, columns []string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Citizens"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	// Map of internal keys to display headers
	headerMap := map[string]string{
		"last_name":       "Прізвище",
		"first_name":      "Ім'я",
		"middle_name":     "По батькові",
		"birth_date":      "Дата народження",
		"passport":        "Паспорт (серія та номер)",
		"passport_series": "Серія паспорта",
		"passport_number": "Номер паспорта",
		"passport_type":   "Тип паспорта",
		"tax_number":      "ІПН",
		"gender":          "Стать",
		"birth_place":     "Місце народження",
		"phone":           "Телефон",
		"email":           "Email",
		"notes":           "Нотатки",
		"address":         "Адреса",
	}

	// Always include ID and Full Name as basic info if not specified?
	// User said "selected columns", so let's stick to their choice + ID.
	actualHeaders := []string{"ID", "ПІБ"}
	for _, col := range columns {
		if h, ok := headerMap[col]; ok {
			actualHeaders = append(actualHeaders, h)
		}
	}

	for i, h := range actualHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Fetch data for selected IDs
	var items []CitizenOutput
	for _, id := range ids {
		c, err := s.citizenService.GetByID(id)
		if err == nil {
			items = append(items, *c)
		}
	}

	for i, c := range items {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), c.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), c.FullName)

		for colIdx, colKey := range columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+3, row)
			var val interface{}
			switch colKey {
			case "last_name":
				val = c.LastName
			case "first_name":
				val = c.FirstName
			case "middle_name":
				val = c.MiddleName
			case "birth_date":
				val = c.BirthDate
			case "passport":
				val = fmt.Sprintf("%s %s", c.PassportSeries, c.PassportNumber)
			case "passport_series":
				val = c.PassportSeries
			case "passport_number":
				val = c.PassportNumber
			case "passport_type":
				val = c.PassportType
			case "tax_number":
				val = c.TaxNumber
			case "gender":
				val = c.GenderDisplay
			case "birth_place":
				val = c.BirthPlace
			case "phone":
				val = c.Phone
			case "email":
				val = c.Email
			case "notes":
				val = c.Notes
			case "address":
				val = c.ActiveAddress
			}
			f.SetCellValue(sheet, cell, val)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ImportFromExcel imports citizens from an Excel file
func (s *ImportExportService) ImportFromExcel(data []byte) (int, error) {
	reader := bytes.NewReader(data)
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return 0, fmt.Errorf("failed to open excel: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Citizens")
	if err != nil {
		// Try first sheet if "Citizens" doesn't exist
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return 0, fmt.Errorf("no sheets found")
		}
		rows, err = f.GetRows(sheets[0])
		if err != nil {
			return 0, fmt.Errorf("failed to get rows: %w", err)
		}
	}

	if len(rows) < 2 {
		return 0, nil // Header only or empty
	}

	count := 0
	for i, row := range rows {
		if i == 0 {
			continue // Skip header
		}
		if len(row) < 3 {
			continue // Mandatory: Last name, First name
		}

		input := &CitizenInput{
			LastName:   row[1],
			FirstName:  row[2],
			MiddleName: getRowValue(row, 3),
			BirthDate:  getRowValue(row, 4),
			// ... and so on for other fields if needed for full import
		}
		// Special case: if we want full import, we need to handle all columns
		// and potentially encryption if we are importing raw data.
		// For MVP, focus on basic fields.

		if len(row) >= 7 {
			input.PassportSeries = row[5]
			input.PassportNumber = row[6]
		}
		if len(row) >= 8 {
			input.PassportType = row[7]
		}
		if len(row) >= 9 {
			input.TaxNumber = row[8]
		}
		if len(row) >= 10 {
			input.Gender = row[9]
		}
		if len(row) >= 11 {
			input.BirthPlace = row[10]
		}
		if len(row) >= 12 {
			input.Phone = row[11]
		}

		_, err := s.citizenService.Create(input)
		if err == nil {
			count++
		}
	}

	return count, nil
}

func getRowValue(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}
