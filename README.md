# mclaude

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

### Using Go

```bash
go install github.com/emiliopalmerini/mclaude/cmd/mclaude@latest
```

### Build from source

```bash
git clone https://github.com/emiliopalmerini/mclaude.git
cd mclaude
make build
```

### Using Nix

```bash
# Run directly
nix run github:emiliopalmerini/mclaude

# Install to profile
nix profile install github:emiliopalmerini/mclaude

# Build locally
nix build
```

## Environment Variables

```bash
# Database
export MCLAUDE_DATABASE_URL="libsql://your-database.turso.io"
export MCLAUDE_AUTH_TOKEN="your-auth-token"
```

## Quick Start

### 1. Run database migrations

```bash
mclaude migrate
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
            "command": "mclaude record"
          }
        ]
      }
    ]
  }
}
```

### 3. Create an experiment

```bash
mclaude experiment create "baseline" \
  --description "Normal usage patterns" \
  --hypothesis "Establish baseline metrics for comparison"
```

### 4. Use Claude Code normally

Sessions are automatically recorded when they end.

### 5. View your data

```bash
# Quick stats
mclaude stats

# Or start the web dashboard
mclaude serve
```

## CLI Commands

### Core

| Command                       | Description                                         |
| ----------------------------- | --------------------------------------------------- |
| `mclaude record`              | Hook handler - captures session data from stdin     |
| `mclaude migrate [n]`         | Run migrations (up to version n, or all if omitted) |
| `mclaude serve [--port 8080]` | Start web dashboard                                 |

### Experiments

```bash
# Create and activate an experiment
mclaude experiment create "minimal-prompts" \
  --description "Testing with shorter, more focused prompts" \
  --hypothesis "Shorter prompts reduce token usage without impacting quality"

# List experiments
mclaude experiment list

# Switch active experiment
mclaude experiment activate <name>
mclaude experiment deactivate <name>

# End an experiment (sets end date)
mclaude experiment end <name>

# Compare two experiments
mclaude experiment compare <exp1> <exp2>

# Delete an experiment
mclaude experiment delete <name>
```

### Stats & Sessions

```bash
# Summary stats
mclaude stats
mclaude stats --experiment "minimal-prompts"
mclaude stats --project <id>
mclaude stats --period week  # today, week, month, all

# List sessions
mclaude sessions list [--last 10]
```

### Cost Configuration

```bash
# List configured pricing
mclaude cost list

# Set model pricing (USD per 1M tokens)
mclaude cost set claude-sonnet-4-20250514 \
  --input 3.00 \
  --output 15.00 \
  --cache-read 0.30 \
  --cache-write 3.75

# Set default model for cost estimation
mclaude cost default claude-sonnet-4-20250514
```

### Cleanup

```bash
# Delete old sessions
mclaude cleanup --before 2024-01-01

# Delete by project or experiment
mclaude cleanup --project <id>
mclaude cleanup --experiment <name>

# Delete specific session
mclaude cleanup --session <id>

# Preview what would be deleted
mclaude cleanup --before 2024-01-01 --dry-run
```

### Export

```bash
mclaude export sessions --format json --output sessions.json
mclaude export sessions --format csv --output sessions.csv
```

## Web Dashboard

Start the dashboard:

```bash
mclaude serve --port 8080
```

Open http://localhost:8080 to view:

- **Dashboard**: Overview metrics, token usage charts, cost trends
- **Sessions**: Browse and filter sessions, view detailed breakdowns
- **Experiments**: Manage experiments, compare results side-by-side
- **Projects**: Aggregate stats by project
- **Settings**: Configure model pricing

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
~/.local/share/mclaude/transcripts/<session_id>.jsonl.gz
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
└── mclaude/         # CLI entry point

internal/
├── domain/                 # Domain entities
├── ports/                  # Repository interfaces
├── adapters/
│   ├── turso/              # Database implementations
│   ├── storage/            # Transcript storage
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
