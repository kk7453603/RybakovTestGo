.PHONY: help build run test clean docker-build docker-run docker-stop proto generate

# Переменные
GOOGLEAPIS_DIR = third_party/googleapis
APP_NAME=crypto-currency-service
DOCKER_IMAGE=$(APP_NAME):latest
PROTO_DIR=pkg/api/proto
GEN_DIR=pkg/api/gen
DOCS_DIR = docs

# Помощь
help: ## Показать справку
	@echo "Доступные команды:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\\n", $$1, $$2}' $(MAKEFILE_LIST)

# Сборка
build: ## Собрать приложение
	@echo "Сборка приложения..."
	@go build -o bin/server cmd/server/main.go
	@echo "✅ Приложение собрано успешно!"

# Запуск
run: ## Запустить приложение
	@echo "Запуск приложения..."
	@go run cmd/server/main.go

# Тесты
test: ## Запустить тесты
	@echo "Запуск тестов..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Тесты выполнены успешно!"

# Генерация protobuf
proto:
	@echo "Генерация protobuf кода..."
	@mkdir -p $(GEN_DIR)
	@mkdir -p $(DOCS_DIR)
	protoc \
		-I $(PROTO_DIR) \
		-I $(GOOGLEAPIS_DIR) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(DOCS_DIR) \
		--openapiv2_opt=logtostderr=true \
		$(PROTO_DIR)/*.proto
	@echo "✅ Protobuf код сгенерирован успешно!"

# Установка зависимостей для protobuf
install-proto-deps: ## Установить зависимости для protobuf
	@echo "Установка protobuf зависимостей..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Скачивание googleapis..."
	@mkdir -p $(GOOGLEAPIS_DIR)/google/api
	@curl -L -o $(GOOGLEAPIS_DIR)/google/api/annotations.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
	@curl -L -o $(GOOGLEAPIS_DIR)/google/api/http.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
	@curl -L -o $(GOOGLEAPIS_DIR)/google/api/field_behavior.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/field_behavior.proto
	@curl -L -o $(GOOGLEAPIS_DIR)/google/api/httpbody.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/httpbody.proto
	@echo "✅ Proto зависимости установлены!"


# Docker команды
docker-build: ## Собрать Docker образ
	@echo "Сборка Docker образа..."
	@docker build -f docker/Dockerfile -t $(DOCKER_IMAGE) .
	@echo "✅ Docker образ собран успешно!"

docker-run: ## Запустить с Docker Compose
	@echo "Запуск с Docker Compose..."
	@docker-compose -f docker/docker-compose.yml up -d
	@echo "✅ Сервисы запущены успешно!"
	@echo "📝 HTTP API: http://localhost:8080"
	@echo "🔌 gRPC: localhost:9090"
	@echo "📚 Swagger UI: http://localhost:8081"
	@echo "🗄️ Adminer: http://localhost:8082"

docker-stop: ## Остановить Docker Compose
	@echo "Остановка Docker Compose..."
	@docker-compose -f docker/docker-compose.yml down
	@echo "✅ Сервисы остановлены!"

docker-logs: ## Посмотреть логи
	@docker-compose -f docker/docker-compose.yml logs -f

# Очистка
clean: ## Очистить сгенерированные файлы
	@echo "Очистка..."
	@rm -rf bin/
	@rm -rf $(GEN_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Очистка завершена!"

# Форматирование кода
fmt: ## Форматировать код
	@echo "Форматирование кода..."
	@go fmt ./...
	@echo "✅ Код отформатирован!"

# Линтинг
lint: ## Запустить линтер
	@echo "Запуск линтера..."
	@golangci-lint run
	@echo "✅ Линтинг завершен!"

# Модули
mod-tidy: ## Обновить go.mod
	@echo "Обновление go.mod..."
	@go mod tidy
	@go mod verify
	@echo "✅ Модули обновлены!"

# Полная сборка
all: clean mod-tidy proto build ## Полная пересборка проекта
	@echo "✅ Проект собран полностью!"
