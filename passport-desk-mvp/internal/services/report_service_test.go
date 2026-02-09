package services

import (
	"context"
	"log/slog"
	"os"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/security"
	"testing"
)

func TestReportService_GetStats(t *testing.T) {
	mockRepo := &MockReportRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	auditService := NewAuditLogService(&MockAuditLogRepo{}, sessionService)
	cs := NewCitizenService(crypto, &MockCitizenRepo{}, auditService)
	rs := NewRegistrationService(&MockRegistrationRepo{}, crypto, auditService)
	service := NewReportService(mockRepo, cs, rs, auditService)
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := logger.NewContext(context.Background(), l)

	mockRepo.GetStatsFunc = func(ctx context.Context) (*models.StatsOutput, error) {
		return &models.StatsOutput{
			TotalCitizens: 100,
		}, nil
	}

	stats, err := service.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalCitizens != 100 {
		t.Errorf("Expected 100 citizens, got %d", stats.TotalCitizens)
	}
}

func TestReportService_ExportRegisteredCitizens(t *testing.T) {
	mockRepo := &MockReportRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	auditService := NewAuditLogService(&MockAuditLogRepo{}, sessionService)
	cs := NewCitizenService(crypto, &MockCitizenRepo{}, auditService)
	rs := NewRegistrationService(&MockRegistrationRepo{}, crypto, auditService)
	service := NewReportService(mockRepo, cs, rs, auditService)
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := logger.NewContext(context.Background(), l)

	mockRepo.ExportRegisteredCitizensFunc = func(ctx context.Context, from, to string) ([]models.CitizenOutput, error) {
		return []models.CitizenOutput{
			{ID: 1, LastName: "Doe", FirstName: "John"},
		}, nil
	}

	// Returns base64 string
	data, err := service.ExportRegisteredCitizens(ctx, "2023-01-01", "2023-12-31")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if data == "" {
		t.Error("Expected base64 data, got empty string")
	}
}
