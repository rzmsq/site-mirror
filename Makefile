.PHONY: all build test clean run help

# Переменные
BINARY_NAME=site-mirror
BUILD_DIR=bin
GO=go
GOFLAGS=-v

# По умолчанию
all: test build

# Сборка проекта
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

# Запуск тестов
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

# Запуск тестов с покрытием
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -cover ./...

# Очистка
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GO) clean

# Запуск приложения
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Форматирование кода
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Проверка кода
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

# Справка
help:
	@echo "Available targets:"
	@echo "  all            - Run tests and build (default)"
	@echo "  build          - Build the application"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Remove build artifacts"
	@echo "  run            - Build and run the application"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Run go vet"
	@echo "  help           - Show this help message"
