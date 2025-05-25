include .env
export

# Development tools
REQUIRED_TOOLS := go goose sqlc
MIGRATIONS_DIR := ./db/migrations

.PHONY: docker-up docker-down build run migrate-up migrate-down sqlc dev-setup migrate-create

docker-up:
	docker-compose -f docker-compose/docker-compose.yaml up -d

docker-down:
	docker-compose -f docker-compose/docker-compose.yaml down

build:
	go build -o bin/app cmd/main.go

run: build
	./bin/app

migrate-up:
	@goose -dir db/migrations up

migrate-down:
	@goose -dir db/migrations down
migrate-create:
	@test $(name) || (echo "name argument is required. Usage: make migrate-create name=migration_name"; exit 1)
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

sqlc:
	sqlc generate -f db/sqlc/sqlc.yml

dev-setup:
	@echo "Installing required development tools..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go mod tidy


