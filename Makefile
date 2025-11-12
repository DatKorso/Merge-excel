# Makefile для Excel Merger

# Переменные
APP_NAME = excel-merger
BUILD_DIR = build
CMD_DIR = cmd/excel-merger
LDFLAGS = -s -w
EXTLDFLAGS = -Wl,-w

# Цвета для вывода
CYAN = \033[0;36m
GREEN = \033[0;32m
NC = \033[0m # No Color

.PHONY: all build run clean test help package

all: build

# Сборка приложения (без предупреждений)
build:
	@echo "$(CYAN)Сборка $(APP_NAME)...$(NC)"
	@go build -ldflags="$(LDFLAGS) -extldflags '$(EXTLDFLAGS)'" -o $(APP_NAME) ./$(CMD_DIR)
	@echo "$(GREEN)✓ Сборка завершена: $(APP_NAME)$(NC)"

# Быстрая сборка для разработки (с предупреждениями)
build-dev:
	@echo "$(CYAN)Быстрая сборка для разработки...$(NC)"
	@go build -o $(APP_NAME) ./$(CMD_DIR)
	@echo "$(GREEN)✓ Сборка завершена$(NC)"

# Запуск приложения
run: build
	@echo "$(CYAN)Запуск $(APP_NAME)...$(NC)"
	@./$(APP_NAME)

# Тесты
test:
	@echo "$(CYAN)Запуск тестов...$(NC)"
	@go test -v ./...

# Тесты с покрытием
test-coverage:
	@echo "$(CYAN)Запуск тестов с анализом покрытия...$(NC)"
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Отчет о покрытии сохранен в coverage.html$(NC)"

# Форматирование кода
fmt:
	@echo "$(CYAN)Форматирование кода...$(NC)"
	@gofmt -w .
	@echo "$(GREEN)✓ Форматирование завершено$(NC)"

# Проверка кода
vet:
	@echo "$(CYAN)Проверка кода (go vet)...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Проверка завершена$(NC)"

# Линтинг (требует golangci-lint)
lint:
	@echo "$(CYAN)Линтинг кода...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
		echo "$(GREEN)✓ Линтинг завершен$(NC)"; \
	else \
		echo "golangci-lint не установлен. Установите: brew install golangci-lint"; \
	fi

# Простая сборка для текущей платформы в build/
build-release:
	@echo "$(CYAN)Сборка релиза для текущей платформы...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="$(LDFLAGS) -extldflags '$(EXTLDFLAGS)'" -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "$(GREEN)✓ Релиз собран: $(BUILD_DIR)/$(APP_NAME)$(NC)"

# Упаковка для macOS (с fyne package)
package-macos:
	@echo "$(CYAN)Упаковка для macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if [ -f ~/go/bin/fyne ]; then \
		cd $(CMD_DIR) && ~/go/bin/fyne package --os darwin --icon ../../assets/icon.png --release; \
		mv $(APP_NAME).app ../../$(BUILD_DIR)/; \
		echo "$(GREEN)✓ macOS app: $(BUILD_DIR)/$(APP_NAME).app$(NC)"; \
	else \
		echo "Fyne CLI не установлен. Установите: go install fyne.io/tools/cmd/fyne@latest"; \
		exit 1; \
	fi

# Упаковка для Windows (требует MinGW)
package-windows:
	@echo "$(CYAN)Упаковка для Windows (требует MinGW)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if [ -f ~/go/bin/fyne ] && command -v x86_64-w64-mingw32-gcc > /dev/null; then \
		cd $(CMD_DIR) && CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
			~/go/bin/fyne package --os windows --icon ../../assets/icon.png --release --app-id com.github.excel-merger; \
		mv $(APP_NAME).exe ../../$(BUILD_DIR)/$(APP_NAME)-win-x64.exe; \
		echo "$(GREEN)✓ Windows exe: $(BUILD_DIR)/$(APP_NAME)-win-x64.exe$(NC)"; \
	else \
		if [ ! -f ~/go/bin/fyne ]; then \
			echo "Fyne CLI не установлен. Установите: go install fyne.io/tools/cmd/fyne@latest"; \
		fi; \
		if ! command -v x86_64-w64-mingw32-gcc > /dev/null; then \
			echo "MinGW не установлен. Установите: brew install mingw-w64"; \
		fi; \
		exit 1; \
	fi

# Упаковка для Linux
package-linux:
	@echo "$(CYAN)Упаковка для Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if [ -f ~/go/bin/fyne ]; then \
		cd $(CMD_DIR) && GOOS=linux GOARCH=amd64 ~/go/bin/fyne package --os linux --icon ../../assets/icon.png --release 2>&1 | tee /tmp/fyne-linux.log; \
		if [ $$? -eq 0 ]; then \
			mv $(APP_NAME).tar.xz ../../$(BUILD_DIR)/$(APP_NAME)-linux-amd64.tar.xz 2>/dev/null || \
			mv $(APP_NAME) ../../$(BUILD_DIR)/$(APP_NAME)-linux-amd64 2>/dev/null || true; \
			echo "$(GREEN)✓ Linux: $(BUILD_DIR)/$(APP_NAME)-linux-amd64*$(NC)"; \
		else \
			echo "$(CYAN)⚠ Сборка для Linux пропущена (требуется Linux или Docker/fyne-cross)$(NC)"; \
		fi; \
	else \
		echo "Fyne CLI не установлен. Установите: go install fyne.io/tools/cmd/fyne@latest"; \
		exit 1; \
	fi

# Упаковка для всех платформ (macOS + Windows, Linux опционально)
package-all: package-macos package-windows
	@echo "$(GREEN)✓ Основные платформы собраны в директории $(BUILD_DIR)/$(NC)"
	@-make package-linux 2>/dev/null || echo "$(CYAN)ℹ Linux сборка пропущена (требуется нативная сборка на Linux)$(NC)"

# Очистка
clean:
	@echo "$(CYAN)Очистка...$(NC)"
	@rm -f $(APP_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf *.app
	@echo "$(GREEN)✓ Очистка завершена$(NC)"

# Установка зависимостей
deps:
	@echo "$(CYAN)Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Зависимости установлены$(NC)"

# Обновление зависимостей
deps-update:
	@echo "$(CYAN)Обновление зависимостей...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Зависимости обновлены$(NC)"

# Помощь
help:
	@echo "$(CYAN)Доступные команды:$(NC)"
	@echo ""
	@echo "  $(GREEN)make build$(NC)              - Сборка приложения (без предупреждений)"
	@echo "  $(GREEN)make build-dev$(NC)          - Быстрая сборка для разработки"
	@echo "  $(GREEN)make build-release$(NC)      - Сборка релиза в build/ для текущей платформы"
	@echo "  $(GREEN)make run$(NC)                - Сборка и запуск приложения"
	@echo "  $(GREEN)make test$(NC)               - Запуск тестов"
	@echo "  $(GREEN)make test-coverage$(NC)      - Тесты с анализом покрытия"
	@echo "  $(GREEN)make fmt$(NC)                - Форматирование кода"
	@echo "  $(GREEN)make vet$(NC)                - Проверка кода (go vet)"
	@echo "  $(GREEN)make lint$(NC)               - Линтинг кода (требует golangci-lint)"
	@echo "  $(GREEN)make package-macos$(NC)      - Упаковка .app для macOS (требует fyne CLI)"
	@echo "  $(GREEN)make package-windows$(NC)    - Упаковка .exe для Windows (требует fyne CLI + MinGW)"
	@echo "  $(GREEN)make package-linux$(NC)      - Упаковка для Linux (требует fyne CLI)"
	@echo "  $(GREEN)make package-all$(NC)        - Упаковка для всех платформ"
	@echo "  $(GREEN)make clean$(NC)              - Очистка сборочных файлов"
	@echo "  $(GREEN)make deps$(NC)               - Установка зависимостей"
	@echo "  $(GREEN)make deps-update$(NC)        - Обновление зависимостей"
	@echo "  $(GREEN)make help$(NC)               - Показать эту справку"
	@echo ""
	@echo "$(CYAN)Требования для кросс-компиляции:$(NC)"
	@echo "  - Fyne CLI: go install fyne.io/tools/cmd/fyne@latest"
	@echo "  - MinGW (для Windows): brew install mingw-w64"
	@echo ""
