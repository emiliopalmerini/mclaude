package sessions

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"claude-watcher/internal/database/sqlc"
)

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockRepo       *MockRepository
		pageSize       int64
		expectedStatus int
	}{
		{
			name:        "success first page",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					if params.Offset != 0 {
						t.Errorf("expected offset 0, got %d", params.Offset)
					}
					if params.Limit != 20 {
						t.Errorf("expected limit 20, got %d", params.Limit)
					}
					return []sqlc.ListSessionsFilteredRow{
						{
							ID:        1,
							SessionID: "abc123",
							Hostname:  "localhost",
							Timestamp: "2024-01-01",
						},
					}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 100, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "success second page",
			queryParams: "?page=2",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					if params.Offset != 20 {
						t.Errorf("expected offset 20, got %d", params.Offset)
					}
					return []sqlc.ListSessionsFilteredRow{}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 100, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid page defaults to 1",
			queryParams: "?page=0",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					if params.Offset != 0 {
						t.Errorf("expected offset 0, got %d", params.Offset)
					}
					return []sqlc.ListSessionsFilteredRow{}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 0, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "negative page defaults to 1",
			queryParams: "?page=-5",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					if params.Offset != 0 {
						t.Errorf("expected offset 0, got %d", params.Offset)
					}
					return []sqlc.ListSessionsFilteredRow{}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 0, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "list sessions error",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					return nil, errors.New("database error")
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "count sessions error",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					return []sqlc.ListSessionsFilteredRow{}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 0, errors.New("database error")
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "custom page size",
			queryParams: "",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					if params.Limit != 10 {
						t.Errorf("expected limit 10, got %d", params.Limit)
					}
					return []sqlc.ListSessionsFilteredRow{}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 50, nil
				},
			},
			pageSize:       10,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "with hostname filter",
			queryParams: "?hostname=myhost",
			mockRepo: &MockRepository{
				ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
					if params.Hostname != "myhost" {
						t.Errorf("expected hostname filter 'myhost', got %v", params.Hostname)
					}
					return []sqlc.ListSessionsFilteredRow{}, nil
				},
				CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
					return 10, nil
				},
			},
			pageSize:       20,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mockRepo, tt.pageSize)

			req := httptest.NewRequest(http.MethodGet, "/sessions"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("List() status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHandler_List_Pagination(t *testing.T) {
	mockRepo := &MockRepository{
		ListSessionsFilteredFunc: func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
			return []sqlc.ListSessionsFilteredRow{
				{ID: 1, SessionID: "sess1", Hostname: "host1", Timestamp: "2024-01-01"},
			}, nil
		},
		CountSessionsFilteredFunc: func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
			return 45, nil
		},
	}

	handler := NewHandler(mockRepo, 20)

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestNewHandler_Sessions(t *testing.T) {
	mockRepo := &MockRepository{}
	pageSize := int64(25)
	handler := NewHandler(mockRepo, pageSize)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.repo != mockRepo {
		t.Error("NewHandler() did not set repository correctly")
	}
	if handler.pageSize != pageSize {
		t.Errorf("NewHandler() pageSize = %d, want %d", handler.pageSize, pageSize)
	}
}

func TestSessionsData(t *testing.T) {
	data := SessionsData{
		Sessions: []sqlc.ListSessionsFilteredRow{
			{
				ID:               1,
				SessionID:        "abc123",
				Hostname:         "localhost",
				Timestamp:        "2024-01-01",
				GitBranch:        sql.NullString{String: "main", Valid: true},
				DurationSeconds:  sql.NullInt64{Int64: 120, Valid: true},
				UserPrompts:      sql.NullInt64{Int64: 5, Valid: true},
				ToolCalls:        sql.NullInt64{Int64: 10, Valid: true},
				EstimatedCostUsd: sql.NullFloat64{Float64: 0.50, Valid: true},
			},
		},
		Page:       1,
		TotalPages: 5,
	}

	if len(data.Sessions) != 1 {
		t.Errorf("SessionsData.Sessions length = %d, want 1", len(data.Sessions))
	}
	if data.Page != 1 {
		t.Errorf("SessionsData.Page = %d, want 1", data.Page)
	}
	if data.TotalPages != 5 {
		t.Errorf("SessionsData.TotalPages = %d, want 5", data.TotalPages)
	}
}

func TestActiveFilters_HasFilters(t *testing.T) {
	tests := []struct {
		name     string
		filters  ActiveFilters
		expected bool
	}{
		{
			name:     "no filters",
			filters:  ActiveFilters{},
			expected: false,
		},
		{
			name:     "hostname filter",
			filters:  ActiveFilters{Hostname: "myhost"},
			expected: true,
		},
		{
			name:     "branch filter",
			filters:  ActiveFilters{Branch: "main"},
			expected: true,
		},
		{
			name:     "date range filter",
			filters:  ActiveFilters{StartDate: "2024-01-01", EndDate: "2024-01-31"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filters.HasFilters(); got != tt.expected {
				t.Errorf("HasFilters() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestActiveFilters_QueryString(t *testing.T) {
	tests := []struct {
		name     string
		filters  ActiveFilters
		contains []string
	}{
		{
			name:     "empty filters",
			filters:  ActiveFilters{},
			contains: []string{},
		},
		{
			name:     "hostname only",
			filters:  ActiveFilters{Hostname: "myhost"},
			contains: []string{"hostname=myhost"},
		},
		{
			name:     "multiple filters",
			filters:  ActiveFilters{Hostname: "myhost", Branch: "main"},
			contains: []string{"hostname=myhost", "branch=main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := tt.filters.QueryString()
			for _, c := range tt.contains {
				if len(c) > 0 && !containsString(qs, c) {
					t.Errorf("QueryString() = %q, want to contain %q", qs, c)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}
