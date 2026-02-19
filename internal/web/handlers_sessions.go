package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read filter params
	experimentFilter := r.URL.Query().Get("experiment")
	projectFilter := r.URL.Query().Get("project")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	// List sessions with joined metrics (single query, no N+1)
	opts := ports.ListSessionsOptions{Limit: limit}
	if experimentFilter != "" {
		opts.ExperimentID = &experimentFilter
	}
	if projectFilter != "" {
		opts.ProjectID = &projectFilter
	}

	items, _ := s.sessionRepo.ListWithMetrics(ctx, opts)

	// Build project and experiment name lookup maps
	projectNameMap := make(map[string]string)
	if projects, err := s.projectRepo.List(ctx); err == nil {
		for _, p := range projects {
			projectNameMap[p.ID] = p.Name
		}
	}
	experimentNameMap := make(map[string]string)
	if experiments, err := s.experimentRepo.List(ctx); err == nil {
		for _, e := range experiments {
			experimentNameMap[e.ID] = e.Name
		}
	}

	var maxTokens int64
	sessionList := make([]templates.SessionSummary, 0, len(items))
	for _, item := range items {
		summary := templates.SessionSummary{
			ID:            item.ID,
			ProjectID:     item.ProjectID,
			ProjectName:   projectNameMap[item.ProjectID],
			CreatedAt:     item.CreatedAt,
			ExitReason:    item.ExitReason,
			Turns:         item.TurnCount,
			Tokens:        item.TotalTokens,
			SubagentCount: item.SubagentCount,
		}
		if item.ExperimentID != nil {
			summary.ExperimentID = *item.ExperimentID
			summary.ExperimentName = experimentNameMap[*item.ExperimentID]
		}
		if item.Cost != nil {
			summary.Cost = *item.Cost
		}
		if item.ModelID != nil {
			summary.Model = *item.ModelID
		}
		if item.Duration != nil {
			summary.Duration = *item.Duration
		}

		if summary.Tokens > maxTokens {
			maxTokens = summary.Tokens
		}

		sessionList = append(sessionList, summary)
	}

	// Populate filter dropdowns
	pageData := templates.SessionsPageData{
		Sessions:         sessionList,
		FilterExperiment: experimentFilter,
		FilterProject:    projectFilter,
		FilterLimit:      limit,
		MaxTokens:        maxTokens,
	}

	for id, name := range experimentNameMap {
		pageData.Experiments = append(pageData.Experiments, templates.FilterOption{ID: id, Name: name})
	}
	for id, name := range projectNameMap {
		pageData.Projects = append(pageData.Projects, templates.FilterOption{ID: id, Name: name})
	}

	// HTMX partial: render only the session table
	if r.Header.Get("HX-Request") != "" {
		templates.SessionTable(pageData).Render(ctx, w)
		return
	}

	templates.SessionsPage(pageData).Render(ctx, w)
}

func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	session, err := queries.GetSessionByID(ctx, id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var metrics *sqlc.SessionMetric
	if m, err := queries.GetSessionMetricsBySessionID(ctx, id); err == nil {
		metrics = &m
	}
	detail := buildSessionDetail(session, metrics)

	// Get tools
	tools, _ := queries.ListSessionToolsBySessionID(ctx, id)
	for _, t := range tools {
		detail.Tools = append(detail.Tools, templates.ToolUsage{
			Name:  t.ToolName,
			Count: t.InvocationCount,
		})
	}

	// Get files
	files, _ := queries.ListSessionFilesBySessionID(ctx, id)
	for _, f := range files {
		detail.Files = append(detail.Files, templates.FileOperation{
			Path:      f.FilePath,
			Operation: f.Operation,
			Count:     f.OperationCount,
		})
	}

	// Get sub-agents (aggregated by type+kind)
	subagentStats, _ := queries.GetSubagentStatsBySession(ctx, id)
	for _, sa := range subagentStats {
		usage := templates.SubagentUsage{
			AgentType: sa.AgentType,
			AgentKind: sa.AgentKind,
			Count:     sa.InvocationCount,
			Tokens:    util.ToInt64(sa.TotalTokens),
			Cost:      util.ToFloat64(sa.TotalCost),
		}
		detail.Subagents = append(detail.Subagents, usage)
	}

	// Get tool events
	toolEvents, _ := queries.ListToolEventsBySessionID(ctx, id)
	for _, te := range toolEvents {
		view := templates.ToolEventView{
			ToolName:   te.ToolName,
			ToolUseID:  te.ToolUseID,
			CapturedAt: te.CapturedAt,
		}
		if te.ToolInput.Valid {
			view.ToolInput = te.ToolInput.String
		}
		if te.ToolResponse.Valid {
			view.ToolResponse = te.ToolResponse.String
		}
		detail.ToolEvents = append(detail.ToolEvents, view)
	}

	templates.SessionDetailPage(detail).Render(ctx, w)
}

func (s *Server) handleAPIDeleteSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	// Delete session from database
	if err := s.sessionRepo.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/sessions")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPICleanupSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	beforeDate := r.FormValue("before_date")
	projectFilter := r.FormValue("project")
	experimentFilter := r.FormValue("experiment")

	if beforeDate == "" && projectFilter == "" && experimentFilter == "" {
		http.Error(w, "At least one filter is required", http.StatusBadRequest)
		return
	}

	var sessionsToDelete []domain.TranscriptPathInfo

	if beforeDate != "" {
		parsed, err := time.Parse("2006-01-02", beforeDate)
		if err != nil {
			http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		sessionsToDelete, _ = s.sessionRepo.GetTranscriptPathsBefore(ctx, parsed.Format(time.RFC3339))
	} else if projectFilter != "" {
		sessionsToDelete, _ = s.sessionRepo.GetTranscriptPathsByProject(ctx, projectFilter)
	} else if experimentFilter != "" {
		sessionsToDelete, _ = s.sessionRepo.GetTranscriptPathsByExperiment(ctx, experimentFilter)
	}

	for _, sess := range sessionsToDelete {
		_ = s.sessionRepo.Delete(ctx, sess.ID)
	}

	w.Header().Set("HX-Redirect", "/sessions")
	w.WriteHeader(http.StatusOK)
}
