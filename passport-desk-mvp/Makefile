# Makefile для проєкту Passport Desk MVP

# Визначення змінних
APP_NAME=passport-desk-mvp
BUILD_DIR=build/bin
FRONTEND_DIR=frontend
DB_NAME=passport_desk.db

# Визначення шляхів до даних
ifeq ($(OS),Windows_NT)
	DATA_DIR=$(APPDATA)/$(APP_NAME)
	RM_CMD=rmdir /s /q
	RM_FILE=del /f /q
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		DATA_DIR=$(HOME)/.config/$(APP_NAME)
	endif
	ifeq ($(UNAME_S),Darwin)
		DATA_DIR=$(HOME)/Library/Application Support/$(APP_NAME)
	endif
	RM_CMD=rm -rf
	RM_FILE=rm -f
endif

.PHONY: help build dev clean clean-db clean-db-confirm clean-all clean-logs backup-db

# Допомога
help:
	@echo "=== Makefile для Passport Desk MVP ==="
	@echo ""
	@echo "Доступні команди:"
	@echo "  make build          - Збірка проєкту"
	@echo "  make dev            - Запуск в режимі розробки"
	@echo "  make clean          - Очищення build директорії"
	@echo "  make clean-db       - Видалення бази даних (з підтвердженням)"
	@echo "  make clean-db-force - Видалення бази даних БЕЗ підтвердження"
	@echo "  make clean-logs     - Видалення логів"
	@echo "  make clean-all      - Повне очищення (build + data)"
	@echo "  make backup-db      - Створення резервної копії БД"
	@echo "  make show-db-path   - Показати шлях до БД"
	@echo "  make show-db-info   - Показати інформацію про БД"

# Збірка
build:
	@echo "🔨 Збірка застосунку..."
	wails build

# Розробка
dev:
	@echo "🚀 Запуск у режимі розробки..."
	wails dev

# Очищення build
clean:
	@echo "🧹 Очищення build директорії..."
	$(RM_CMD) $(BUILD_DIR)
	@echo "✅ Build директорію очищено"

# Показати шлях до БД
show-db-path:
	@echo "📁 Шлях до бази даних:"
	@echo "   $(DATA_DIR)"
	@echo ""
	@echo "📄 Файли бази даних:"
	@echo "   $(DATA_DIR)/$(DB_NAME)"
	@echo "   $(DATA_DIR)/db.salt"

# Показати інформацію про БД
show-db-info:
	@echo "ℹ️  Інформація про базу даних:"
	@echo "   Розташування: $(DATA_DIR)"
ifeq ($(OS),Windows_NT)
	@if exist "$(DATA_DIR)\$(DB_NAME)" ( \
		echo    Статус: БД існує & \
		for %%A in ("$(DATA_DIR)\$(DB_NAME)") do echo    Розмір: %%~zA байт \
	) else ( \
		echo    Статус: БД не знайдено \
	)
	@if exist "$(DATA_DIR)\db.salt" ( \
		echo    Salt файл: Присутній \
	) else ( \
		echo    Salt файл: Відсутній \
	)
else
	@if [ -f "$(DATA_DIR)/$(DB_NAME)" ]; then \
		echo "   Статус: БД існує"; \
		echo "   Розмір: $$(du -h '$(DATA_DIR)/$(DB_NAME)' | cut -f1)"; \
	else \
		echo "   Статус: БД не знайдено"; \
	fi
	@if [ -f "$(DATA_DIR)/db.salt" ]; then \
		echo "   Salt файл: Присутній"; \
	else \
		echo "   Salt файл: Відсутній"; \
	fi
endif

# Резервна копія БД
backup-db:
	@echo "💾 Створення резервної копії бази даних..."
ifeq ($(OS),Windows_NT)
	@if not exist "$(DATA_DIR)\backups" mkdir "$(DATA_DIR)\backups"
	@if exist "$(DATA_DIR)\$(DB_NAME)" ( \
		copy "$(DATA_DIR)\$(DB_NAME)" "$(DATA_DIR)\backups\$(DB_NAME)-manual-backup-%date:~-4,4%%date:~-10,2%%date:~-7,2%-%time:~0,2%%time:~3,2%%time:~6,2%.db" & \
		copy "$(DATA_DIR)\db.salt" "$(DATA_DIR)\backups\db-manual-backup-%date:~-4,4%%date:~-10,2%%date:~-7,2%-%time:~0,2%%time:~3,2%%time:~6,2%.salt" & \
		echo ✅ Резервну копію створено \
	) else ( \
		echo ❌ База даних не знайдена \
	)
else
	@mkdir -p "$(DATA_DIR)/backups"
	@if [ -f "$(DATA_DIR)/$(DB_NAME)" ]; then \
		cp "$(DATA_DIR)/$(DB_NAME)" "$(DATA_DIR)/backups/$(DB_NAME)-manual-backup-$$(date +%Y%m%d-%H%M%S).db"; \
		cp "$(DATA_DIR)/db.salt" "$(DATA_DIR)/backups/db-manual-backup-$$(date +%Y%m%d-%H%M%S).salt"; \
		echo "✅ Резервну копію створено"; \
	else \
		echo "❌ База даних не знайдена"; \
	fi
endif

# Видалення БД з підтвердженням
clean-db: backup-db
	@echo ""
	@echo "⚠️  УВАГА: Ви збираєтесь видалити базу даних!"
	@echo "   Це видалить ВСІ дані про громадян, реєстрації та операторів."
	@echo "   Резервну копію вже створено."
	@echo ""
	@echo "   Для підтвердження введіть 'yes' або натисніть Ctrl+C для скасування:"
	@read -p "   > " CONFIRM; \
	if [ "$$CONFIRM" = "yes" ]; then \
		$(MAKE) clean-db-force; \
	else \
		echo "❌ Видалення скасовано"; \
	fi

# Видалення БД БЕЗ підтвердження (небезпечно!)
clean-db-force:
	@echo "🗑️  Видалення бази даних..."
ifeq ($(OS),Windows_NT)
	@if exist "$(DATA_DIR)\$(DB_NAME)" ( \
		del /f /q "$(DATA_DIR)\$(DB_NAME)" & \
		echo    ✅ $(DB_NAME) видалено \
	) else ( \
		echo    ℹ️  $(DB_NAME) не знайдено \
	)
	@if exist "$(DATA_DIR)\db.salt" ( \
		del /f /q "$(DATA_DIR)\db.salt" & \
		echo    ✅ db.salt видалено \
	) else ( \
		echo    ℹ️  db.salt не знайдено \
	)
else
	@if [ -f "$(DATA_DIR)/$(DB_NAME)" ]; then \
		rm -f "$(DATA_DIR)/$(DB_NAME)"; \
		echo "   ✅ $(DB_NAME) видалено"; \
	else \
		echo "   ℹ️  $(DB_NAME) не знайдено"; \
	fi
	@if [ -f "$(DATA_DIR)/db.salt" ]; then \
		rm -f "$(DATA_DIR)/db.salt"; \
		echo "   ✅ db.salt видалено"; \
	else \
		echo "   ℹ️  db.salt не знайдено"; \
	fi
endif
	@echo ""
	@echo "✅ База даних видалена. При наступному запуску буде створена нова."

# Видалення логів
clean-logs:
	@echo "🗑️  Видалення логів..."
ifeq ($(OS),Windows_NT)
	@if exist "$(DATA_DIR)\logs" ( \
		rmdir /s /q "$(DATA_DIR)\logs" & \
		echo ✅ Логи видалено \
	) else ( \
		echo ℹ️  Логи не знайдено \
	)
else
	@if [ -d "$(DATA_DIR)/logs" ]; then \
		rm -rf "$(DATA_DIR)/logs"; \
		echo "✅ Логи видалено"; \
	else \
		echo "ℹ️  Логи не знайдено"; \
	fi
endif

# Повне очищення (build + data)
clean-all: clean clean-db-force clean-logs
	@echo "🧹 Повне очищення завершено"
	@echo ""
	@echo "Видалено:"
	@echo "  ✅ Build директорію"
	@echo "  ✅ База даних"
	@echo "  ✅ Логи"

# Швидкий restart (видалення БД + запуск dev)
restart-dev: clean-db-force
	@echo ""
	@echo "🚀 Перезапуск у режимі розробки..."
	$(MAKE) dev

# Тестування (майбутнє)
test:
	@echo "🧪 Запуск тестів..."
	go test -v ./...

# Тестування з покриттям
test-coverage:
	@echo "🧪 Запуск тестів з покриттям..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Звіт про покриття збережено у coverage.html"

# Лінтинг
lint:
	@echo "🔍 Запуск лінтера..."
	golangci-lint run

# Форматування коду
fmt:
	@echo "✨ Форматування коду..."
	go fmt ./...
	cd $(FRONTEND_DIR) && npm run format

# Встановлення залежностей
install:
	@echo "📦 Встановлення залежностей..."
	go mod download
	cd $(FRONTEND_DIR) && npm install

# Оновлення залежностей
update:
	@echo "🔄 Оновлення залежностей..."
	go get -u ./...
	go mod tidy
	cd $(FRONTEND_DIR) && npm update