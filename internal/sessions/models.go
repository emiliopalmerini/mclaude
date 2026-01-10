package sessions

import (
	"net/url"

	"claude-watcher/internal/database/sqlc"
)

type FilterOptions struct {
	Hostnames []string
	Branches  []string
	Models    []string
}

type ActiveFilters struct {
	Hostname  string
	Branch    string
	Model     string
	StartDate string
	EndDate   string
}

func (f ActiveFilters) HasFilters() bool {
	return f.Hostname != "" || f.Branch != "" || f.Model != "" || f.StartDate != "" || f.EndDate != ""
}

func (f ActiveFilters) QueryString() string {
	params := url.Values{}
	if f.Hostname != "" {
		params.Set("hostname", f.Hostname)
	}
	if f.Branch != "" {
		params.Set("branch", f.Branch)
	}
	if f.Model != "" {
		params.Set("model", f.Model)
	}
	if f.StartDate != "" {
		params.Set("start_date", f.StartDate)
	}
	if f.EndDate != "" {
		params.Set("end_date", f.EndDate)
	}
	return params.Encode()
}

type SessionsData struct {
	Sessions      []sqlc.ListSessionsFilteredRow
	Page          int
	TotalPages    int
	FilterOptions FilterOptions
	ActiveFilters ActiveFilters
}
