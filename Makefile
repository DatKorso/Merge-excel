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

# Упаковка для macOS
package-macos:
	@echo "$(CYAN)Упаковка приложения для macOS...$(NC)"
	@if command -v fyne > /dev/null; then \
		fyne package -os darwin -icon assets/icon.png -name "Excel Merger"; \
		echo "$(GREEN)✓ Упаковка завершена: Excel Merger.app$(NC)"; \
	else \
		echo "Fyne CLI не установлен. Установите: go install fyne.io/fyne/v2/cmd/fyne@latest"; \
	fi

# Упаковка для всех платформ
package-all:
	@echo "$(CYAN)Упаковка для всех платформ...$(NC)"
	@mkdir -p $(BUILD_DIR)
	# macOS
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS) -extldflags '$(EXTLDFLAGS)'" -o $(BUILD_DIR)/$(APP_NAME)-macos-amd64 ./$(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-macos-arm64 ./$(CMD_DIR)
	# Windows
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./$(CMD_DIR)
	# Linux
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./$(CMD_DIR)
	@echo "$(GREEN)✓ Упаковка завершена в директории $(BUILD_DIR)/$(NC)"

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
	@echo "  $(GREEN)make build$(NC)          - Сборка приложения (без предупреждений)"
	@echo "  $(GREEN)make build-dev$(NC)      - Быстрая сборка для разработки"
	@echo "  $(GREEN)make run$(NC)            - Сборка и запуск приложения"
	@echo "  $(GREEN)make test$(NC)           - Запуск тестов"
	@echo "  $(GREEN)make test-coverage$(NC)  - Тесты с анализом покрытия"
	@echo "  $(GREEN)make fmt$(NC)            - Форматирование кода"
	@echo "  $(GREEN)make vet$(NC)            - Проверка кода (go vet)"
	@echo "  $(GREEN)make lint$(NC)           - Линтинг кода (требует golangci-lint)"
	@echo "  $(GREEN)make package-macos$(NC)  - Упаковка приложения для macOS"
	@echo "  $(GREEN)make package-all$(NC)    - Сборка для всех платформ"
	@echo "  $(GREEN)make clean$(NC)          - Очистка сборочных файлов"
	@echo "  $(GREEN)make deps$(NC)           - Установка зависимостей"
	@echo "  $(GREEN)make deps-update$(NC)    - Обновление зависимостей"
	@echo "  $(GREEN)make help$(NC)           - Показать эту справку"
	@echo ""
