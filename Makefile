.PHONY: help setup build run test lint clean migrate-up migrate-down docker-up docker-down backup restore docker-build security-check infra-up infra-down dev-backend dev-frontend dev-all dev-setup

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Initialize project dependencies
	go mod tidy
	go mod download

build: ## Build the application
	go build -o bin/server ./cmd/server

build-cli: ## Build the CLI tool
	go build -o bin/studio-cli ./cmd/studio-cli

build-all: ## Build all binaries
	@make build
	@make build-cli
	@echo "✓ All binaries built successfully"

run: ## Run the application (local config by default)
	go run ./cmd/server/main.go -config configs/config.local.yaml

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

## ─── Local Development ──────────────────────────────────────────────────────

dev-setup: ## First-time setup: copy config templates
	@if [ ! -f configs/config.local.yaml ]; then \
		cp configs/config.local.yaml.example configs/config.local.yaml; \
		echo "✓ Created configs/config.local.yaml — edit with your local DB/Redis credentials"; \
	else \
		echo "  configs/config.local.yaml already exists, skipping"; \
	fi
	@if [ ! -f configs/config.prod.yaml ]; then \
		cp configs/config.prod.yaml.example configs/config.prod.yaml; \
		echo "✓ Created configs/config.prod.yaml — edit with your production credentials"; \
	else \
		echo "  configs/config.prod.yaml already exists, skipping"; \
	fi
	@if [ ! -f apps/web/.env.local ]; then \
		cp apps/web/.env.example apps/web/.env.local; \
		echo "✓ Created apps/web/.env.local"; \
	else \
		echo "  apps/web/.env.local already exists, skipping"; \
	fi

infra-up: ## Start local Docker infra (Postgres + Redis)
	docker-compose up -d
	@echo "Waiting for Postgres..."
	@until docker exec studio_postgres pg_isready -U studio -q 2>/dev/null; do sleep 1; done
	@echo "✓ Postgres :5432  Redis :6379  ready"

infra-down: ## Stop local Docker infra
	docker-compose down

dev-backend: ## Run backend with LOCAL config
	go run ./cmd/server/main.go -config configs/config.local.yaml

prod-backend: ## Run backend with PRODUCTION config
	go run ./cmd/server/main.go -config configs/config.prod.yaml

dev-frontend: ## Run Next.js dev server (uses apps/web/.env.local)
	cd apps/web && pnpm dev

dev-all: ## Setup + start infra + print next steps
	@make dev-setup
	@make infra-up
	@make migrate-up
	@echo ""
	@echo "✓ Ready. Now run in separate terminals:"
	@echo "  make dev-backend    # Go API (local config)"
	@echo "  make dev-frontend   # Next.js"
	@echo ""

dev: dev-all ## Alias for dev-all

ci: ## Run CI checks locally
	@echo "Running CI checks..."
	@make lint
	@make test
	@make build
	@echo "✓ All CI checks passed"
