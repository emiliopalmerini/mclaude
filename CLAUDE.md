# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

mclaude is a personal analytics and experimentation platform for Claude Code usage. It captures session data via hooks, parses transcripts for detailed metrics, and provides a web dashboard for visualization and analysis.

**Primary use case**: Run month-long experiments with different Claude usage styles, then compare token usage, efficacy, efficiency, and cost metrics.

## Tech Stack

| Component         | Technology               |
| ----------------- | ------------------------ |
| Language          | Go 1.22+                 |
| CLI               | Cobra                    |
| Database          | Turso (libsql)           |
| SQL               | sqlc (type-safe queries) |
| Migrations        | go-migrate (embedded)    |
| Web Templates     | templ                    |
| Web Interactivity | HTMX + Alpine.js         |
| Charts            | Apache ECharts           |

## Environment Variables

```bash
# Database
MCLAUDE_DATABASE_URL  # libsql connection URL
MCLAUDE_AUTH_TOKEN    # Turso auth token

# OpenTelemetry (optional)
MCLAUDE_OTEL_ENABLED      # Enable OTEL export (true/false)
MCLAUDE_OTEL_ENDPOINT     # OTEL Collector endpoint (e.g., localhost:4317)
MCLAUDE_OTEL_INSECURE     # Use insecure connection (true/false)

# Prometheus (optional)
MCLAUDE_PROMETHEUS_ENABLED  # Enable Prometheus queries (true/false)
MCLAUDE_PROMETHEUS_URL      # Prometheus URL (e.g., http://localhost:9090)
```

## Build Commands

```bash
make build      # Build the mclaude binary
make test       # Run tests
make sqlc       # Regenerate sqlc code
make templ      # Regenerate templ templates
make fmt        # Format code
make clean      # Remove build artifacts
```

## Architecture

The codebase follows hexagonal architecture (ports & adapters):

```
cmd/
└── mclaude/             # Single CLI entry point (Cobra)

internal/
├── domain/                     # Domain entities (no dependencies)
│   ├── session.go              # Session, SessionMetrics
│   ├── experiment.go           # Experiment
│   ├── project.go              # Project
│   ├── pricing.go              # ModelPricing
│   └── usage.go                # PlanConfig, UsageSummary, plan presets
│
├── ports/                      # Interfaces (defined by domain)
│   ├── session_repository.go
│   ├── experiment_repository.go
│   ├── project_repository.go
│   ├── pricing_repository.go
│   ├── usage_repository.go
│   ├── transcript_storage.go
│   ├── metrics_exporter.go     # OTEL metrics export interface
│   └── prometheus_client.go    # Prometheus query interface
│
├── adapters/                   # Interface implementations
│   ├── turso/                  # Database adapters
│   │   ├── db.go               # Connection setup
│   │   ├── session_repository.go
│   │   ├── experiment_repository.go
│   │   ├── project_repository.go
│   │   ├── pricing_repository.go
│   │   └── usage_repository.go
│   ├── storage/
│   │   └── transcript_storage.go   # XDG + gzip storage
│   ├── otel/                   # OpenTelemetry adapters
│   │   ├── exporter.go         # OTEL SDK metrics exporter
│   │   ├── noop.go             # NoOp for graceful degradation
│   │   └── config.go           # Environment config
│   └── prometheus/             # Prometheus adapters
│       ├── client.go           # HTTP query client
│       ├── noop.go             # NoOp for graceful degradation
│       └── config.go           # Environment config
│
├── parser/                     # Transcript parsing (pure logic)
│   └── transcript.go           # JSONL parser, metric extraction
│
├── cli/                        # Cobra commands
│   ├── root.go
│   ├── record.go               # Hook handler
│   ├── migrate.go              # Database migrations
│   ├── serve.go                # Web server
│   ├── experiment.go           # Experiment CRUD
│   ├── stats.go                # Quick statistics
│   ├── sessions.go             # Session listing
│   ├── cost.go                 # Pricing config
│   ├── limits.go               # Usage limit tracking
│   ├── cleanup.go              # Data deletion
│   └── export.go               # JSON/CSV export
│
└── web/                        # Web dashboard
    ├── server.go               # HTTP server setup
    ├── routes.go               # Route definitions
    ├── handlers/               # HTTP handlers
    │   ├── dashboard.go
    │   ├── experiments.go
    │   ├── sessions.go
    │   ├── projects.go
    │   └── settings.go
    ├── templates/              # templ templates
    │   ├── layouts/
    │   │   └── base.templ
    │   ├── pages/
    │   │   ├── dashboard.templ
    │   │   ├── experiments.templ
    │   │   ├── sessions.templ
    │   │   ├── projects.templ
    │   │   └── settings.templ
    │   └── components/
    │       ├── navbar.templ
    │       ├── stats_card.templ
    │       └── chart.templ
    └── static/
        ├── css/
        └── js/                 # htmx, alpine, echarts

migrations/                     # SQL files (go:embed)
├── 001_create_experiments.up.sql
├── 001_create_experiments.down.sql
├── ...
└── embed.go

sqlc/
├── sqlc.yaml
├── queries/                    # SQL query definitions
│   ├── sessions.sql
│   ├── experiments.sql
│   ├── projects.sql
│   └── pricing.sql
└── generated/                  # DO NOT EDIT - sqlc output
```

## Database Schema

9 normalized tables:

1. **experiments** - Experiment definitions (name, description, hypothesis, dates, is_active)
2. **projects** - Project aggregations (id=SHA256 of cwd, path, name)
3. **sessions** - Core session data (links to project and experiment)
4. **session_metrics** - Token counts, cost estimates per session
5. **session_tools** - Tool usage breakdown per session
6. **session_files** - File operations per session
7. **session_commands** - Bash commands executed per session
8. **model_pricing** - Configurable model pricing for cost estimation
9. **plan_config** - Usage limit tracking (plan type, 5-hour and weekly window limits)

## Key Patterns

### Adding a New CLI Command

1. Create `internal/cli/<command>.go`
2. Implement Cobra command with `RunE` function
3. Register in `internal/cli/root.go` via `rootCmd.AddCommand()`
4. Inject dependencies (repositories) from root command setup

### Adding a New Database Query

1. Add SQL to `sqlc/queries/<domain>.sql` with sqlc annotations:
   ```sql
   -- name: GetSessionByID :one
   SELECT * FROM sessions WHERE id = ?;
   ```
2. Run `make sqlc` to regenerate `sqlc/generated/`
3. Use generated functions in repository adapters

### Adding a Web Page

1. Create handler in `internal/web/handlers/<page>.go`
2. Create templ template in `internal/web/templates/pages/<page>.templ`
3. Run `make templ` to generate Go code
4. Register route in `internal/web/routes.go`

### Transcript Parsing

The `record` command receives JSON from stdin (Claude Code hook), then:

1. Parses the hook input (session_id, transcript_path, cwd, permission_mode, reason)
2. Reads and parses the JSONL transcript file
3. Extracts metrics: timestamps, token counts, tool usage, file ops, commands
4. Copies transcript to XDG storage with gzip compression
5. Saves all data to database (session + metrics + tools + files + commands)

### Domain Rules

- **Active experiment**: Only one experiment can be active at a time
- **Project ID**: SHA256 hash of the `cwd` path (deterministic)
- **Cost estimation**: Uses default model pricing if no model specified
- **Cleanup cascade**: Deleting sessions also removes transcript files from storage
- **Usage windows**: Dual-window tracking (5-hour and 7-day) with fixed start times that auto-reset when expired
- **Plan presets**: Three tiers (pro, max_5x, max_20x) with estimated limits; users can learn actual limits

### Observability Integration

- **OTEL Export**: After each session is saved to the database, enriched metrics are exported to OTEL Collector (if configured)
- **Prometheus Queries**: `limits` command and web dashboard can query Prometheus for real-time usage data
- **Graceful Degradation**: If OTEL/Prometheus unavailable, system continues with local data only
- **Data Sources**: Dashboard and CLI show data source indicator (prometheus vs local)

## Testing Strategy

- **Unit tests**: Domain logic, parser, cost calculations
- **Integration tests**: Repository adapters with test database
- **No mocks for ports**: Use real implementations with test fixtures

## Code Style

- Follow Go conventions (gofmt, golint)
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Table-driven tests
- No global state - dependency injection via constructors
- Context propagation for cancellation

## Performance Considerations

- `record` command must be fast (runs on every session end)
- Use single INSERT statements, no transactions for simple writes
- Transcript parsing is I/O bound - read file once, extract all metrics
- Web dashboard queries should use appropriate indexes

## Common Tasks

### Reset database

```bash
mclaude migrate 0    # Migrate down to version 0
mclaude migrate      # Migrate up to latest
```

### Test record command locally

```bash
echo '{"session_id":"test123","transcript_path":"/path/to/transcript.jsonl","cwd":"/project","permission_mode":"default","reason":"exit"}' | mclaude record
```

### Debug transcript parsing

```bash
cat /path/to/transcript.jsonl | head -20  # Inspect structure
```
