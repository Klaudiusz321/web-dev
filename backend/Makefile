# Variables
APP_NAME=web-crawler-backend
DB_URL=root:password@tcp(localhost:3306)/webcrawler?charset=utf8mb4&parseTime=True&loc=Local

# Build commands
.PHONY: build
build:
	go build -o bin/$(APP_NAME) main.go

.PHONY: build-migrate
build-migrate:
	go build -o bin/migrate cmd/migrate/main.go

# Development commands
.PHONY: dev
dev:
	go run main.go

.PHONY: test
test:
	go test ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

# Database commands
.PHONY: migrate-up
migrate-up: build-migrate
	./bin/migrate -action=up

.PHONY: migrate-down
migrate-down: build-migrate
	./bin/migrate -action=down -steps=1

.PHONY: migrate-version
migrate-version: build-migrate
	./bin/migrate -action=version

.PHONY: migrate-reset
migrate-reset: build-migrate
	./bin/migrate -action=down -steps=10
	./bin/migrate -action=up

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 $(APP_NAME)

# Dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Clean
.PHONY: clean
clean:
	rm -rf bin/
	go clean

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  build-migrate - Build the migration tool"
	@echo "  dev           - Run in development mode"
	@echo "  test          - Run tests"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback one migration"
	@echo "  migrate-version - Show current migration version"
	@echo "  migrate-reset - Reset all migrations and reapply"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  clean         - Clean build artifacts" 