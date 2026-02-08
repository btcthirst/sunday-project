package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Log *slog.Logger
)

// Config містить налаштування логування
type Config struct {
	// Рівень логування: debug, info, warn, error
	Level string

	// Шлях до директорії з логами
	LogDir string

	// Ім'я файлу логів
	LogFile string

	// Максимальний розмір файлу в мегабайтах (default: 10)
	MaxSize int

	// Максимальна кількість днів зберігання логів (default: 30)
	MaxAge int

	// Максимальна кількість backup файлів (default: 10)
	MaxBackups int

	// Стискати старі лог-файли (default: true)
	Compress bool

	// Логувати також в консоль (default: true для dev)
	Console bool

	// JSON формат (default: false для читабельності)
	JSON bool
}

// Init ініціалізує глобальний logger
func Init(cfg Config) error {
	// Встановити defaults
	if cfg.MaxSize == 0 {
		cfg.MaxSize = 10
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 30
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 10
	}

	// Створити директорію для логів якщо не існує
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return err
	}

	// Налаштувати lumberjack для ротації логів
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, cfg.LogFile),
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	// Визначити рівень логування
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Налаштувати handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Додавати інформацію про файл та рядок
	}

	// Вибрати writer(s)
	var writers []io.Writer
	writers = append(writers, logFile)

	if cfg.Console {
		writers = append(writers, os.Stdout)
	}

	multiWriter := io.MultiWriter(writers...)

	// Створити handler (JSON або Text)
	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(multiWriter, opts)
	} else {
		handler = slog.NewTextHandler(multiWriter, opts)
	}

	// Створити logger
	Log = slog.New(handler)

	// Встановити як default logger
	slog.SetDefault(Log)

	Log.Info("Logger initialized",
		slog.String("level", cfg.Level),
		slog.String("logfile", filepath.Join(cfg.LogDir, cfg.LogFile)),
		slog.Bool("console", cfg.Console),
	)

	return nil
}

// Close закриває logger (викликати при shutdown)
func Close() error {
	// slog не вимагає явного закриття, але якщо потрібно
	// можна додати додаткову логіку
	return nil
}

// WithContext створює logger з додатковими контекстними полями
func WithContext(fields ...slog.Attr) *slog.Logger {
	args := make([]any, 0, len(fields))
	for _, field := range fields {
		args = append(args, field)
	}
	return Log.With(args...)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return Log
	}
	if v := ctx.Value("logger"); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return Log
}

// NewContext adds logger to context
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}
