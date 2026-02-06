package logger

import "log/slog"

// Debug логує повідомлення рівня Debug
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// Info логує повідомлення рівня Info
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Warn логує повідомлення рівня Warn
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Error логує повідомлення рівня Error
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

// Fatal логує помилку та викликає panic
func Fatal(msg string, args ...any) {
	Log.Error(msg, args...)
	panic(msg)
}

// WithFields створює logger з додатковими полями
func WithFields(fields map[string]any) *slog.Logger {
	args := make([]any, 0, len(fields))
	for k, v := range fields {
		args = append(args, slog.Any(k, v))
	}
	return Log.With(args...)
}
