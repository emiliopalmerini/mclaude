.PHONY: all fmt vet build run test clean generate sqlc templ migrate-up migrate-down migrate-version migrate-create

# Database connection string for migrations
# Override with: make migrate-up TURSO_URL=libsql://... TURSO_TOKEN=...
TURSO_URL ?= $(TURSO_DATABASE_URL)
TURSO_TOKEN ?= $(TURSO_AUTH_TOKEN)
MIGRATE_DB_URL = $(TURSO_URL)?authToken=$(TURSO_TOKEN)

all: generate fmt vet test build

generate: sqlc templ

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

sqlc:
	sqlc generate

templ:
	templ generate

build: vet sqlc templ
	go build -o claude-watcher ./cmd

run: build
	./claude-watcher

test: vet
	go test -v ./...

clean:
	rm -f claude-watcher
	go clean ./...

# Database migrations using golang-migrate
# Install: go install -tags 'libsql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate-up:
	migrate -database "$(MIGRATE_DB_URL)" -path migrations up

migrate-down:
	migrate -database "$(MIGRATE_DB_URL)" -path migrations down

migrate-version:
	migrate -database "$(MIGRATE_DB_URL)" -path migrations version

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name
