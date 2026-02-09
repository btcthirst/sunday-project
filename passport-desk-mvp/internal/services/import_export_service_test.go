package services

import (
	"context"
	"log/slog"
	"os"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/security"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestImportExportService_Export(t *testing.T) {
	mockRepo := &MockCitizenRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	auditService := NewAuditLogService(&MockAuditLogRepo{}, sessionService)
	cs := NewCitizenService(crypto, mockRepo, auditService)
	service := NewImportExportService(cs, auditService)

	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := logger.NewContext(context.Background(), l)

	mockRepo.GetAllFunc = func(ctx context.Context) ([]*models.CitizenOutput, error) {
		return []*models.CitizenOutput{
			{ID: 1, LastName: "Doe", FirstName: "John"},
		}, nil
	}

	data, err := service.ExportToExcel(ctx)
	if err != nil {
		t.Fatalf("ExportToExcel failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty excel data")
	}
}

func TestImportExportService_Import(t *testing.T) {
	mockRepo := &MockCitizenRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	auditService := NewAuditLogService(&MockAuditLogRepo{}, sessionService)
	cs := NewCitizenService(crypto, mockRepo, auditService)
	service := NewImportExportService(cs, auditService)

	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := logger.NewContext(context.Background(), l)

	// Create a dummy Excel file in memory
	f := excelize.NewFile()
	sheet := "Citizens"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "ID")
	f.SetCellValue(sheet, "B1", "Прізвище")
	f.SetCellValue(sheet, "C1", "Ім'я")
	f.SetCellValue(sheet, "D1", "По батькові")
	f.SetCellValue(sheet, "E1", "Дата народження")
	f.SetCellValue(sheet, "F1", "Серія")
	f.SetCellValue(sheet, "G1", "Номер")
	f.SetCellValue(sheet, "H1", "Тип")
	f.SetCellValue(sheet, "I1", "ІПН")
	f.SetCellValue(sheet, "J1", "Стать")

	f.SetCellValue(sheet, "A2", "1")
	f.SetCellValue(sheet, "B2", "Doe")
	f.SetCellValue(sheet, "C2", "John")
	f.SetCellValue(sheet, "D2", "Junior")
	f.SetCellValue(sheet, "E2", "1990-01-01")
	f.SetCellValue(sheet, "F2", "AA")
	f.SetCellValue(sheet, "G2", "123456")
	f.SetCellValue(sheet, "H2", "old")
	f.SetCellValue(sheet, "I2", "1234567890")
	f.SetCellValue(sheet, "J2", "M")

	buf, _ := f.WriteToBuffer()

	mockRepo.CreateFunc = func(ctx context.Context, citizen *models.CitizenInput) (int64, error) {
		return 1, nil
	}

	count, err := service.ImportFromExcel(ctx, buf.Bytes())
	if err != nil {
		t.Fatalf("ImportFromExcel failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 imported, got %d", count)
	}
}
