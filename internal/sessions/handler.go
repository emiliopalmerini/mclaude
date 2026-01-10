package sessions

import (
	"context"
	"net/http"
	"strconv"

	"claude-watcher/internal/database/sqlc"
	apperrors "claude-watcher/internal/shared/errors"
	"claude-watcher/internal/shared/middleware"
)

type Handler struct {
	repo     Repository
	pageSize int64
}

func NewHandler(repo Repository, pageSize int64) *Handler {
	return &Handler{repo: repo, pageSize: pageSize}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	offset := int64(page-1) * h.pageSize

	filters := ActiveFilters{
		Hostname:  q.Get("hostname"),
		Branch:    q.Get("branch"),
		Model:     q.Get("model"),
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
	}

	filterParams := sqlc.ListSessionsFilteredParams{
		Hostname:  nilIfEmpty(filters.Hostname),
		GitBranch: nilIfEmpty(filters.Branch),
		Model:     nilIfEmpty(filters.Model),
		StartDate: nilIfEmpty(filters.StartDate),
		EndDate:   nilIfEmpty(filters.EndDate),
		Limit:     h.pageSize,
		Offset:    offset,
	}

	sessions, err := h.repo.ListSessionsFiltered(ctx, filterParams)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	countParams := sqlc.CountSessionsFilteredParams{
		Hostname:  filterParams.Hostname,
		GitBranch: filterParams.GitBranch,
		Model:     filterParams.Model,
		StartDate: filterParams.StartDate,
		EndDate:   filterParams.EndDate,
	}
	count, err := h.repo.CountSessionsFiltered(ctx, countParams)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}
	totalPages := int((count + h.pageSize - 1) / h.pageSize)

	filterOptions, err := h.loadFilterOptions(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	data := SessionsData{
		Sessions:      sessions,
		Page:          page,
		TotalPages:    totalPages,
		FilterOptions: filterOptions,
		ActiveFilters: filters,
	}

	if middleware.IsHTMX(r) {
		SessionsTable(data).Render(ctx, w)
		return
	}

	SessionsList(data).Render(ctx, w)
}

func (h *Handler) loadFilterOptions(ctx context.Context) (FilterOptions, error) {
	hostnames, err := h.repo.GetDistinctHostnames(ctx)
	if err != nil {
		return FilterOptions{}, err
	}

	branchRows, err := h.repo.GetDistinctBranches(ctx)
	if err != nil {
		return FilterOptions{}, err
	}
	branches := make([]string, 0, len(branchRows))
	for _, b := range branchRows {
		if b.Valid {
			branches = append(branches, b.String)
		}
	}

	modelRows, err := h.repo.GetDistinctModels(ctx)
	if err != nil {
		return FilterOptions{}, err
	}
	models := make([]string, 0, len(modelRows))
	for _, m := range modelRows {
		if m.Valid {
			models = append(models, m.String)
		}
	}

	return FilterOptions{
		Hostnames: hostnames,
		Branches:  branches,
		Models:    models,
	}, nil
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
