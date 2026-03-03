.PHONY: help setup build run test lint clean migrate-up migrate-down docker-up docker-down

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

lint: ## Run linters
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/

migrate-up: ## Run database migrations up
	migrate -path migrations -database "$(STUDIO_DATABASE_DSN)" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "$(STUDIO_DATABASE_DSN)" down 1

docker-up: ## Start Docker services
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down
