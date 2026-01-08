# claude-watcher

A small dashboard to check Claude Code usage.

## Prerequisites

- Go 1.25+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI with libsql support
- A Turso database

## Setup

### Install golang-migrate with libsql support

```bash
go install -tags 'libsql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Environment Variables

Set the following environment variables:

```bash
export TURSO_DATABASE_URL="libsql://your-database.turso.io"
export TURSO_AUTH_TOKEN="your-auth-token"
```

## Database Migrations

Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate). Migration files are located in the `migrations/` directory.

### Run migrations

```bash
# Apply all pending migrations
make migrate-up

# Or manually with migrate CLI
migrate -database "libsql://your-database.turso.io?authToken=your-token" -path migrations up
```

### Rollback migrations

```bash
# Rollback the last migration
make migrate-down

# Rollback all migrations
migrate -database "libsql://your-database.turso.io?authToken=your-token" -path migrations down
```

### Check migration version

```bash
make migrate-version
```

### Create a new migration

```bash
make migrate-create
# Enter migration name when prompted
```

## Build and Run

```bash
# Build the application
make build

# Run the application
make run

# Or build and run in one step
./claude-watcher
```

## Development

```bash
# Generate sqlc and templ code
make generate

# Run tests
make test

# Format code
make fmt
```
