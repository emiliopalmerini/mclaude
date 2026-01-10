package costs

import "claude-watcher/internal/database/sqlc"

type CostsData struct {
	Projects     []sqlc.GetProjectMetricsRow
	Models       []sqlc.GetModelEfficiencyRow
	CacheMetrics sqlc.GetCacheMetricsRow
	CacheDaily   []sqlc.GetCacheMetricsDailyRow
	TotalSavings float64
	CacheHitRate float64
	Range        string
}
