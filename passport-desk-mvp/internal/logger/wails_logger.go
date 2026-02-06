package logger

import (
	"log/slog"
	"os"
)

// internal/services/wails_logger.go
type WailsSlogAdapter struct {
	logger *slog.Logger
}

func NewWailsSlogAdapter(l *slog.Logger) *WailsSlogAdapter {
	return &WailsSlogAdapter{logger: l}
}

// Реалізуємо методи інтерфейсу logger.Logger із Wails
func (a *WailsSlogAdapter) Print(message string)   { a.logger.Info(message) }
func (a *WailsSlogAdapter) Trace(message string)   { a.logger.Debug(message, "level", "trace") }
func (a *WailsSlogAdapter) Debug(message string)   { a.logger.Debug(message) }
func (a *WailsSlogAdapter) Info(message string)    { a.logger.Info(message) }
func (a *WailsSlogAdapter) Warning(message string) { a.logger.Warn(message) }
func (a *WailsSlogAdapter) Error(message string)   { a.logger.Error(message) }
func (a *WailsSlogAdapter) Fatal(message string)   { a.logger.Error(message, "fatal", true); os.Exit(1) }
