package dashboard

import "claude-watcher/internal/database/sqlc"

type DashboardData struct {
	Metrics sqlc.GetDashboardMetricsRow
	Today   sqlc.GetTodayMetricsRow
	Week    sqlc.GetWeekMetricsRow
}
