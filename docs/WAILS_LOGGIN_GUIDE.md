# Налаштування slog/lumberjack логування для Wails проєкту

## Вступ

Для production-ready Wails застосунку потрібна надійна система логування. Рекомендується використовувати:
- **slog** (стандартна бібліотека Go 1.21+) - структуроване логування
- **lumberjack** - ротація лог-файлів

## 1. Встановлення залежностей

```bash
# Lumberjack для ротації логів
go get gopkg.in/natefinch/lumberjack.v2
```

## 2. Структура логування

### 2.1 Створіть пакет logger

**Файл:** `internal/logger/logger.go`

```go
package logger

import (
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
		Level: level,
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
	return Log.With(slog.Group("context", fields...))
}
```

### 2.2 Wrapper функції для зручності

**Файл:** `internal/logger/helpers.go`

```go
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
func WithFields(fields map[string]interface{}) *slog.Logger {
	attrs := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return Log.With(attrs...)
}
```

## 3. Інтеграція з Wails

### 3.1 Оновіть app.go

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"passport-desk-mvp/internal/logger"
	// ... інші імпорти
)

type App struct {
	ctx context.Context
	// ... інші поля
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Отримати директорію для даних
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	dataDir := filepath.Join(configDir, "passport-desk-mvp")

	// Ініціалізувати logger
	logConfig := logger.Config{
		Level:      getLogLevel(),
		LogDir:     filepath.Join(dataDir, "logs"),
		LogFile:    "passport-desk.log",
		MaxSize:    10,  // 10 MB
		MaxAge:     30,  // 30 днів
		MaxBackups: 10,  // 10 backup файлів
		Compress:   true,
		Console:    isDevelopment(),
		JSON:       false, // Text формат для читабельності
	}

	if err := logger.Init(logConfig); err != nil {
		// Fallback на стандартний log якщо не вдалось ініціалізувати
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.Info("Application started",
		slog.String("version", "1.0.0"),
		slog.String("data_dir", dataDir),
	)

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		logger.Error("Failed to create data directory",
			slog.String("error", err.Error()),
		)
	}

	// ... решта ініціалізації
}

func (a *App) shutdown(ctx context.Context) {
	logger.Info("Application shutting down")
	
	// Закрити ресурси
	if a.db != nil {
		a.db.Close()
	}
	
	logger.Close()
}

// getLogLevel визначає рівень логування з environment або config
func getLogLevel() string {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		return level
	}
	if isDevelopment() {
		return "debug"
	}
	return "info"
}

// isDevelopment перевіряє чи це dev режим
func isDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development" || env == "dev" || env == ""
}
```

## 4. Використання в коді

### 4.1 У Service Layer

```go
package services

import (
	"log/slog"
	"passport-desk-mvp/internal/logger"
)

func (s *CitizenService) Create(input *CitizenInput) (*CitizenOutput, error) {
	logger.Info("Creating new citizen",
		slog.String("last_name", input.LastName),
		slog.String("first_name", input.FirstName),
	)

	// Бізнес логіка
	citizen, err := s.createCitizen(input)
	if err != nil {
		logger.Error("Failed to create citizen",
			slog.String("error", err.Error()),
			slog.String("last_name", input.LastName),
		)
		return nil, err
	}

	logger.Info("Citizen created successfully",
		slog.Int64("citizen_id", citizen.ID),
	)

	return citizen, nil
}
```

### 4.2 З контекстними полями

```go
// Створити logger для конкретної операції
opLogger := logger.WithFields(map[string]interface{}{
	"operation": "citizen_registration",
	"operator_id": operatorID,
	"citizen_id": citizenID,
})

opLogger.Info("Starting registration process")

// Виконати операцію
if err := registerCitizen(); err != nil {
	opLogger.Error("Registration failed",
		slog.String("error", err.Error()),
	)
}

opLogger.Info("Registration completed")
```

### 4.3 З структурованими даними

```go
logger.Info("Database query executed",
	slog.Group("query",
		slog.String("sql", "SELECT * FROM citizens WHERE id = ?"),
		slog.Int64("id", 123),
		slog.Duration("duration", duration),
	),
	slog.Group("result",
		slog.Int("rows", 1),
		slog.Bool("cached", false),
	),
)
```

## 5. Логування помилок з stack trace

### 5.1 Error wrapper

**Файл:** `internal/logger/errors.go`

```go
package logger

import (
	"fmt"
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
```

### 5.2 Використання

```go
func (a *App) CreateCitizen(input services.CitizenInput) (*services.CitizenOutput, error) {
	defer logger.RecoverPanic()

	a.UpdateActivity()

	citizen, err := a.citizenService.Create(&input)
	if err != nil {
		logger.LogError(err, "Failed to create citizen",
			slog.String("last_name", input.LastName),
			slog.String("operator", a.currentOperator.Username),
		)
		return nil, err
	}

	return citizen, nil
}
```

## 6. Ротація логів на практиці

### 6.1 Приклад конфігурації

```go
// Для production
productionConfig := logger.Config{
	Level:      "info",
	LogDir:     "/var/log/passport-desk",
	LogFile:    "application.log",
	MaxSize:    50,    // 50 MB
	MaxAge:     90,    // 90 днів
	MaxBackups: 20,    // 20 файлів
	Compress:   true,  // Стискати старі файли
	Console:    false, // Не логувати в консоль
	JSON:       true,  // JSON для парсингу
}

// Для development
devConfig := logger.Config{
	Level:      "debug",
	LogDir:     "./logs",
	LogFile:    "dev.log",
	MaxSize:    10,
	MaxAge:     7,
	MaxBackups: 3,
	Compress:   false,
	Console:    true, // Дублювати в консоль
	JSON:       false, // Text для читабельності
}
```

### 6.2 Структура лог-файлів

```
logs/
├── passport-desk.log              # Поточний лог
├── passport-desk-2024-02-01.log.gz  # Старий лог (стиснутий)
├── passport-desk-2024-02-02.log.gz
└── passport-desk-2024-02-03.log.gz
```

## 7. Моніторинг та аналіз логів

### 7.1 Пошук помилок

```bash
# Знайти всі ERROR записи
grep "level=ERROR" logs/passport-desk.log

# Знайти помилки за останню годину
grep "level=ERROR" logs/passport-desk.log | grep "$(date -d '1 hour ago' '+%Y-%m-%d %H')"

# Статистика по рівням
grep -o "level=[A-Z]*" logs/passport-desk.log | sort | uniq -c
```

### 7.2 Viewing tool для Windows/Linux

**Windows:**
```powershell
# PowerShell
Get-Content -Path "logs\passport-desk.log" -Tail 100 -Wait
```

**Linux:**
```bash
# Tail з follow
tail -f logs/passport-desk.log

# С фільтрацією
tail -f logs/passport-desk.log | grep ERROR
```

## 8. Додаткові можливості

### 8.1 Різні logger для різних компонентів

```go
// У app.go
var (
	AppLogger      *slog.Logger
	DatabaseLogger *slog.Logger
	SecurityLogger *slog.Logger
)

func initLoggers(baseLogger *slog.Logger) {
	AppLogger = baseLogger.With(slog.String("component", "app"))
	DatabaseLogger = baseLogger.With(slog.String("component", "database"))
	SecurityLogger = baseLogger.With(slog.String("component", "security"))
}

// Використання
DatabaseLogger.Info("Query executed",
	slog.String("sql", query),
	slog.Duration("duration", duration),
)
```

### 8.2 Умовне логування

```go
// Логувати тільки повільні запити
func (s *Service) Query(sql string) error {
	start := time.Now()
	err := s.db.Exec(sql)
	duration := time.Since(start)

	// Логувати тільки якщо > 100ms
	if duration > 100*time.Millisecond {
		logger.Warn("Slow query detected",
			slog.String("sql", sql),
			slog.Duration("duration", duration),
		)
	}

	return err
}
```

### 8.3 Метрики через логи

```go
type Metrics struct {
	mu sync.Mutex
	counters map[string]int64
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

func (m *Metrics) LogPeriodically(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for name, count := range m.counters {
			logger.Info("Metric",
				slog.String("name", name),
				slog.Int64("count", count),
			)
			m.counters[name] = 0 // Reset
		}
		m.mu.Unlock()
	}
}
```

## 9. Оновлення go.mod

```go
module passport-desk-mvp

go 1.24.0

require (
	// ... існуючі залежності
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
```

## 10. Best Practices

### ✅ DO

1. **Використовуйте структуровані поля**
   ```go
   logger.Info("User logged in",
       slog.String("username", username),
       slog.String("ip", ipAddress),
   )
   ```

2. **Групуйте пов'язані поля**
   ```go
   logger.Info("Request completed",
       slog.Group("request",
           slog.String("method", "POST"),
           slog.String("path", "/api/citizens"),
       ),
       slog.Group("response",
           slog.Int("status", 200),
           slog.Duration("duration", duration),
       ),
   )
   ```

3. **Логуйте на правильному рівні**
   - DEBUG: Детальна інформація для розробки
   - INFO: Важливі події в нормальній роботі
   - WARN: Потенційні проблеми
   - ERROR: Помилки що потребують уваги

### ❌ DON'T

1. **НЕ логуйте чутливі дані**
   ```go
   // ❌ ПОГАНО
   logger.Info("Login attempt", slog.String("password", password))
   
   // ✅ ДОБРЕ
   logger.Info("Login attempt", slog.String("username", username))
   ```

2. **НЕ логуйте занадто багато в циклах**
   ```go
   // ❌ ПОГАНО
   for i := 0; i < 10000; i++ {
       logger.Debug("Processing item", slog.Int("index", i))
   }
   
   // ✅ ДОБРЕ
   logger.Info("Processing batch",
       slog.Int("count", 10000),
       slog.Int("batch_size", batchSize),
   )
   ```

3. **НЕ ігноруйте помилки ініціалізації**
   ```go
   // ❌ ПОГАНО
   logger.Init(config) // Ігнорується помилка
   
   // ✅ ДОБРЕ
   if err := logger.Init(config); err != nil {
       panic("Failed to initialize logger: " + err.Error())
   }
   ```

## 11. Тестування логування

```go
package logger_test

import (
	"bytes"
	"log/slog"
	"testing"
	
	"passport-desk-mvp/internal/logger"
)

func TestLogger(t *testing.T) {
	// Створити buffer для перехоплення логів
	var buf bytes.Buffer
	
	// Створити test logger
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	
	logger.Log = testLogger
	
	// Тест
	logger.Info("Test message", slog.String("key", "value"))
	
	// Перевірка
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Test message")) {
		t.Errorf("Expected log output to contain 'Test message', got: %s", output)
	}
}
```

## Висновок

Ця конфігурація забезпечує:
- ✅ Автоматичну ротацію логів
- ✅ Структуроване логування
- ✅ Різні рівні логування
- ✅ Контекстну інформацію
- ✅ Stack traces для помилок
- ✅ Production-ready setup