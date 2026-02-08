package turso

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type ToolEventRepository struct {
	queries *sqlc.Queries
}

func NewToolEventRepository(db *sql.DB) *ToolEventRepository {
	return &ToolEventRepository{
		queries: sqlc.New(db),
	}
}

func (r *ToolEventRepository) ListBySessionID(ctx context.Context, sessionID string) ([]*domain.ToolEvent, error) {
	rows, err := r.queries.ListToolEventsBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tool events: %w", err)
	}

	events := make([]*domain.ToolEvent, len(rows))
	for i, row := range rows {
		events[i] = &domain.ToolEvent{
			ID:           row.ID,
			SessionID:    row.SessionID,
			ToolName:     row.ToolName,
			ToolUseID:    row.ToolUseID,
			ToolInput:    util.NullStringToPtr(row.ToolInput),
			ToolResponse: util.NullStringToPtr(row.ToolResponse),
			CapturedAt:   row.CapturedAt,
		}
	}
	return events, nil
}
