# claude-watcher

A personal analytics and experimentation platform for Claude Code usage.

Track sessions, measure token consumption, estimate costs, and run experiments to optimize your Claude Code workflow.

## Features

- **Session Tracking**: Automatically capture session data via Claude Code hooks
- **Transcript Parsing**: Extract detailed metrics from session transcripts (tokens, tools, files, commands)
- **Experiments**: A/B test different usage styles and compare results
- **Cost Estimation**: Track spending with configurable model pricing
- **Web Dashboard**: Visualize metrics, compare experiments, and analyze usage patterns
- **CLI Tools**: Quick stats, session management, and data export

## Prerequisites

- Go 1.22+
- A Turso database (or local libsql)
- Claude Code with hooks enabled

## Installation

### Build from source

```bash
make build
```

### Using Nix

```bash
nix build
```

## Environment Variables

```bash
export CLAUDE_WATCHER_DATABASE_URL="libsql://your-database.turso.io"
export CLAUDE_WATCHER_AUTH_TOKEN="your-auth-token"
```

## Quick Start

### 1. Run database migrations

```bash
claude-watcher migrate
```

### 2. Configure Claude Code hook

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "claude-watcher record"
          }
        ]
      }
    ]
  }
}
```

### 3. Create an experiment

```bash
claude-watcher experiment create "baseline" \
  --description "Normal usage patterns" \
  --hypothesis "Establish baseline metrics for comparison"
```

### 4. Use Claude Code normally

Sessions are automatically recorded when they end.

### 5. View your data

```bash
# Quick stats
claude-watcher stats

# Or start the web dashboard
claude-watcher serve
```

## CLI Commands

### Core

| Command | Description |
|---------|-------------|
| `claude-watcher record` | Hook handler - captures session data from stdin |
| `claude-watcher migrate [n]` | Run migrations (up to version n, or all if omitted) |
| `claude-watcher serve [--port 8080]` | Start web dashboard |

### Experiments

```bash
# Create and activate an experiment
claude-watcher experiment create "minimal-prompts" \
  --description "Testing with shorter, more focused prompts" \
  --hypothesis "Shorter prompts reduce token usage without impacting quality"

# List experiments
claude-watcher experiment list

# Switch active experiment
claude-watcher experiment activate <name>
claude-watcher experiment deactivate <name>

# End an experiment (sets end date)
claude-watcher experiment end <name>

# Compare two experiments
claude-watcher experiment compare <exp1> <exp2>

# Delete an experiment
claude-watcher experiment delete <name>
```

### Stats & Sessions

```bash
# Summary stats
claude-watcher stats
claude-watcher stats --experiment "minimal-prompts"
claude-watcher stats --project <id>
claude-watcher stats --period week  # today, week, month, all

# List sessions
claude-watcher sessions list [--last 10]
```

### Cost Configuration

```bash
# List configured pricing
claude-watcher cost list

# Set model pricing (USD per 1M tokens)
claude-watcher cost set claude-sonnet-4-20250514 \
  --input 3.00 \
  --output 15.00 \
  --cache-read 0.30 \
  --cache-write 3.75

# Set default model for cost estimation
claude-watcher cost default claude-sonnet-4-20250514
```

### Cleanup

```bash
# Delete old sessions
claude-watcher cleanup --before 2024-01-01

# Delete by project or experiment
claude-watcher cleanup --project <id>
claude-watcher cleanup --experiment <name>

# Delete specific session
claude-watcher cleanup --session <id>

# Preview what would be deleted
claude-watcher cleanup --before 2024-01-01 --dry-run
```

### Export

```bash
claude-watcher export sessions --format json --output sessions.json
claude-watcher export sessions --format csv --output sessions.csv
```

## Web Dashboard

Start the dashboard:

```bash
claude-watcher serve --port 8080
```

Open http://localhost:8080 to view:

- **Dashboard**: Overview metrics, token usage charts, cost trends
- **Sessions**: Browse and filter sessions, view detailed breakdowns
- **Experiments**: Manage experiments, compare results side-by-side
- **Projects**: Aggregate stats by project
- **Settings**: Configure model pricing, manage active experiment

## Data Storage

### Database (Turso)

All metrics are stored in a Turso database with normalized tables:

- `sessions` - Core session data
- `session_metrics` - Token counts, costs
- `session_tools` - Tool usage per session
- `session_files` - File operations per session
- `session_commands` - Bash commands executed
- `experiments` - Experiment definitions
- `projects` - Project aggregations
- `model_pricing` - Cost configuration

### Transcripts

Session transcripts are copied and compressed to:

```
~/.local/share/claude-watcher/transcripts/<session_id>.jsonl.gz
```

## Development

```bash
# Generate sqlc code
make sqlc

# Generate templ templates
make templ

# Run tests
make test

# Format code
make fmt

# Build everything
make build

# Clean build artifacts
make clean
```

## Architecture

```
cmd/
└── claude-watcher/         # CLI entry point

internal/
├── domain/                 # Domain entities
├── ports/                  # Repository interfaces
├── adapters/
│   ├── turso/              # Database implementations
│   └── storage/            # Transcript storage
├── parser/                 # Transcript JSONL parser
├── cli/                    # Cobra commands
└── web/
    ├── handlers/           # HTTP handlers
    ├── templates/          # templ templates
    └── static/             # JS/CSS assets

migrations/                 # SQL migrations (embedded)
```

## License

MIT
