package web

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

type exportSession struct {
	ID                    string  `json:"id"`
	ProjectID             string  `json:"project_id"`
	ExperimentID          string  `json:"experiment_id,omitempty"`
	Cwd                   string  `json:"cwd"`
	PermissionMode        string  `json:"permission_mode"`
	ExitReason            string  `json:"exit_reason"`
	StartedAt             string  `json:"started_at,omitempty"`
	EndedAt               string  `json:"ended_at,omitempty"`
	DurationSeconds       int64   `json:"duration_seconds,omitempty"`
	CreatedAt             string  `json:"created_at"`
	MessageCountUser      int64   `json:"message_count_user"`
	MessageCountAssistant int64   `json:"message_count_assistant"`
	TurnCount             int64   `json:"turn_count"`
	TokenInput            int64   `json:"token_input"`
	TokenOutput           int64   `json:"token_output"`
	TokenCacheRead        int64   `json:"token_cache_read"`
	TokenCacheWrite       int64   `json:"token_cache_write"`
	CostEstimateUsd       float64 `json:"cost_estimate_usd,omitempty"`
	ErrorCount            int64   `json:"error_count"`
}

func (s *Server) handleAPIExportSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	limit := 1000
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	opts := ports.ListSessionsOptions{Limit: limit}
	if expID := r.URL.Query().Get("experiment"); expID != "" {
		opts.ExperimentID = &expID
	}
	if projID := r.URL.Query().Get("project"); projID != "" {
		opts.ProjectID = &projID
	}

	sessions, err := s.sessionRepo.List(ctx, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	exportData := make([]exportSession, 0, len(sessions))
	for _, sess := range sessions {
		es := exportSession{
			ID:             sess.ID,
			ProjectID:      sess.ProjectID,
			Cwd:            sess.Cwd,
			PermissionMode: sess.PermissionMode,
			ExitReason:     sess.ExitReason,
			CreatedAt:      sess.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if sess.ExperimentID != nil {
			es.ExperimentID = *sess.ExperimentID
		}
		if sess.StartedAt != nil {
			es.StartedAt = sess.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if sess.EndedAt != nil {
			es.EndedAt = sess.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if sess.DurationSeconds != nil {
			es.DurationSeconds = *sess.DurationSeconds
		}

		if m, err := s.metricsRepo.GetBySessionID(ctx, sess.ID); err == nil && m != nil {
			es.MessageCountUser = m.MessageCountUser
			es.MessageCountAssistant = m.MessageCountAssistant
			es.TurnCount = m.TurnCount
			es.TokenInput = m.TokenInput
			es.TokenOutput = m.TokenOutput
			es.TokenCacheRead = m.TokenCacheRead
			es.TokenCacheWrite = m.TokenCacheWrite
			es.ErrorCount = m.ErrorCount
			if m.CostEstimateUSD != nil {
				es.CostEstimateUsd = *m.CostEstimateUSD
			}
		}

		exportData = append(exportData, es)
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=sessions.csv")

		writer := csv.NewWriter(w)
		defer writer.Flush()

		header := []string{
			"id", "project_id", "experiment_id", "cwd", "permission_mode", "exit_reason",
			"started_at", "ended_at", "duration_seconds", "created_at",
			"message_count_user", "message_count_assistant", "turn_count",
			"token_input", "token_output", "token_cache_read", "token_cache_write",
			"cost_estimate_usd", "error_count",
		}
		writer.Write(header)

		for _, es := range exportData {
			row := []string{
				es.ID, es.ProjectID, es.ExperimentID, es.Cwd, es.PermissionMode, es.ExitReason,
				es.StartedAt, es.EndedAt, fmt.Sprintf("%d", es.DurationSeconds), es.CreatedAt,
				fmt.Sprintf("%d", es.MessageCountUser), fmt.Sprintf("%d", es.MessageCountAssistant),
				fmt.Sprintf("%d", es.TurnCount), fmt.Sprintf("%d", es.TokenInput),
				fmt.Sprintf("%d", es.TokenOutput), fmt.Sprintf("%d", es.TokenCacheRead),
				fmt.Sprintf("%d", es.TokenCacheWrite), fmt.Sprintf("%.6f", es.CostEstimateUsd),
				fmt.Sprintf("%d", es.ErrorCount),
			}
			writer.Write(row)
		}

	default: // json
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=sessions.json")

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(exportData)
	}
}
