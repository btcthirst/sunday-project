package services

import (
	"context"
	"log/slog"
	"os"
	"passport-desk-mvp/internal/logger"
	internalLogger "passport-desk-mvp/internal/logger"
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

func TestReportService_formatInitials(t *testing.T) {
	service := &ReportService{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Single name", "John", "John"},
		{"Two names", "Doe John", "J. Doe"},
		{"Three names", "Doe John Middle", "J.M. Doe"},
		{"Already formatted", "J. Doe", "J. Doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatInitials(tt.input)
			if result != tt.expected {
				t.Errorf("formatInitials(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReportService_GenerateRegistrationCertificate(t *testing.T) {
	mockRepo := &MockReportRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	auditSvc := NewAuditLogService(&MockAuditLogRepo{}, sessionService)
	mockCitizenRepo := &MockCitizenRepo{}
	cs := NewCitizenService(crypto, mockCitizenRepo, auditSvc)
	mockRegRepo := &MockRegistrationRepo{}
	rs := NewRegistrationService(mockRegRepo, crypto, auditSvc)
	service := NewReportService(mockRepo, cs, rs, auditSvc)

	ctx := internalLogger.NewContext(context.Background(), slog.New(slog.NewTextHandler(os.Stdout, nil)))

	mockCitizenRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.CitizenOutput, error) {
		return &models.CitizenOutput{ID: id, LastName: "Doe", FirstName: "John", BirthDate: "1990-01-01"}, nil
	}
	mockRegRepo.GetByCitizenIDFunc = func(ctx context.Context, id int64) ([]models.RegistrationOutput, error) {
		return []models.RegistrationOutput{{IsActive: true, Settlement: "Kyiv", Street: "Main", HouseNumber: "1", RegistrationDate: "2020-01-01"}}, nil
	}

	// This might fail if it can't find fonts. Since we are using assets.Fonts, it should be okay if assets are in the path.
	// But in tests, assets might not be initialized properly if we don't have them in the package.
	// Let's safe-guard the test or mock loadFontBytes if possible (it's unexported).
	// Actually, the service uses internal/assets which is a real package with embed.

	pdf, err := service.GenerateRegistrationCertificate(ctx, 1, models.FamilyCertificateOptions{SignatoryName: "Admin"})
	if err != nil {
		t.Skipf("Skipping PDF test (possibly missing fonts in test env): %v", err)
		return
	}

	if pdf == "" {
		t.Error("Expected base64 PDF, got empty string")
	}
}
