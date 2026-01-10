package productivity

import "claude-watcher/internal/database/sqlc"

type ProductivityData struct {
	Efficiency  sqlc.GetEfficiencyMetricsRow
	DailyTrends []sqlc.GetEfficiencyMetricsDailyRow
	DayOfWeek   []sqlc.GetDayOfWeekDistributionRow
	HourOfDay   []sqlc.GetHourOfDayDistributionRow
	TopTools    []ToolCount
	Range       string
}

type ToolCount struct {
	Name  string
	Count int
}
