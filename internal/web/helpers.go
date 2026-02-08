package web

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

// calculateSuccessRate computes a success rate from nullable success/failure counts.
// Returns nil if total is zero.
func calculateSuccessRate(successCount, failureCount sql.NullFloat64) *float64 {
	sc := int64(0)
	fc := int64(0)
	if successCount.Valid {
		sc = int64(successCount.Float64)
	}
	if failureCount.Valid {
		fc = int64(failureCount.Float64)
	}
	total := sc + fc
	if total > 0 {
		rate := float64(sc) / float64(total)
		return &rate
	}
	return nil
}

// formatChartDate converts a date interface{} (from sqlc) to a "YYYY-MM-DD" string.
func formatChartDate(date any) string {
	switch d := date.(type) {
	case time.Time:
		return d.Format("2006-01-02")
	case string:
		return d
	default:
		return fmt.Sprintf("%v", date)
	}
}

// buildSessionDetail constructs a templates.SessionDetail from a sqlc Session and optional metrics.
func buildSessionDetail(session sqlc.Session, metrics *sqlc.SessionMetric) templates.SessionDetail {
	detail := templates.SessionDetail{
		ID:             session.ID,
		ProjectID:      session.ProjectID,
		Cwd:            session.Cwd,
		PermissionMode: session.PermissionMode,
		ExitReason:     session.ExitReason,
		CreatedAt:      session.CreatedAt,
	}

	if session.ExperimentID.Valid {
		detail.ExperimentID = session.ExperimentID.String
	}
	if session.StartedAt.Valid {
		detail.StartedAt = session.StartedAt.String
	}
	if session.EndedAt.Valid {
		detail.EndedAt = session.EndedAt.String
	}
	if session.DurationSeconds.Valid {
		detail.DurationSeconds = session.DurationSeconds.Int64
	}

	if metrics != nil {
		detail.MessageCountUser = metrics.MessageCountUser
		detail.MessageCountAssistant = metrics.MessageCountAssistant
		detail.TurnCount = metrics.TurnCount
		detail.TokenInput = metrics.TokenInput
		detail.TokenOutput = metrics.TokenOutput
		detail.TokenCacheRead = metrics.TokenCacheRead
		detail.TokenCacheWrite = metrics.TokenCacheWrite
		detail.ErrorCount = metrics.ErrorCount
		if metrics.CostEstimateUsd.Valid {
			detail.CostEstimateUsd = metrics.CostEstimateUsd.Float64
		}
		if metrics.ModelID.Valid {
			detail.ModelID = metrics.ModelID.String
		}
	}

	return detail
}
