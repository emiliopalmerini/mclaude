package dashboard

import "claude-watcher/internal/database/sqlc"

type DashboardData struct {
	Metrics           sqlc.GetDashboardMetricsRow
	Today             sqlc.GetTodayMetricsRow
	Week              sqlc.GetWeekMetricsRow
	CacheMetrics      sqlc.GetCacheMetricsRow
	TopProject        sqlc.GetTopProjectRow
	EfficiencyMetrics sqlc.GetEfficiencyMetricsRow
	TopTool           string
	CacheHitRate      float64
	UsageSinceLimit   sqlc.GetUsageSinceLastLimitRow
}
