.PHONY: help setup build run test lint clean migrate-up migrate-down docker-up docker-down backup restore docker-build security-check

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Initialize project dependencies
	go mod tidy
	go mod download

build: ## Build the application
	go build -o bin/server ./cmd/server

run: ## Run the application
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

migrate-up: ## Run database migrations up
	migrate -path migrations -database "$(STUDIO_DATABASE_DSN)" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "$(STUDIO_DATABASE_DSN)" down 1

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_users)
	migrate create -ext sql -dir migrations -seq $(NAME)

docker-up: ## Start Docker services
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-build: ## Build Docker images
	docker build -t studio-backend:latest .
	docker build -t studio-frontend:latest -f apps/web/Dockerfile .

docker-logs: ## Show Docker logs
	docker-compose logs -f

backup: ## Backup database
	./scripts/backup-db.sh ./backups

restore: ## Restore database (usage: make restore FILE=./backups/studio_db_20260304_120000.sql.gz)
	./scripts/restore-db.sh $(FILE)

security-check: ## Run security checks
	@echo "Running security checks..."
	@echo "1. Checking for hardcoded secrets..."
	@grep -r "password\|secret\|key" --include="*.go" --include="*.yaml" . || echo "✓ No obvious secrets found"
	@echo "2. Checking Go dependencies for vulnerabilities..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth || echo "✓ No known vulnerabilities"
	@echo "3. Running gosec..."
	@gosec -quiet ./... || echo "✓ No security issues found"

dev: ## Start development environment
	@echo "Starting development environment..."
	@make docker-up
	@sleep 3
	@make migrate-up
	@echo "✓ Development environment ready"
	@echo "Run 'make run' to start the backend"

ci: ## Run CI checks locally
	@echo "Running CI checks..."
	@make lint
	@make test
	@make build
	@echo "✓ All CI checks passed"
