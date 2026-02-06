package logger

import (
	"log/slog"
	"runtime"
)

// LogError логує помилку з stack trace
func LogError(err error, msg string, args ...any) {
	if err == nil {
		return
	}

	// Отримати stack trace
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = fn.Name()
	}

	// Додати stack info до аргументів
	allArgs := append(args,
		slog.String("error", err.Error()),
		slog.String("file", file),
		slog.Int("line", line),
		slog.String("function", funcName),
	)

	Log.Error(msg, allArgs...)
}

// RecoverPanic перехоплює panic та логує
func RecoverPanic() {
	if r := recover(); r != nil {
		// Отримати stack trace
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		Log.Error("Panic recovered",
			slog.Any("panic", r),
			slog.String("stack_trace", stackTrace),
		)
	}
}
