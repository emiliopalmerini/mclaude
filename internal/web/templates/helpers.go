package templates

import (
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/emiliopalmerini/mclaude/internal/util"
)

func formatTokens(n int64) string {
	return util.FormatTokensInt(n)
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func formatDateTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2, 15:04")
}

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2, 2006")
}

func formatDateShort(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2")
}

func formatCost(c float64) string {
	return fmt.Sprintf("$%.2f", c)
}

func formatCostPrecise(c float64) string {
	return fmt.Sprintf("$%.4f", c)
}

func formatInt(n int64) string {
	return fmt.Sprintf("%d", n)
}

func colSpan(n int) string {
	return fmt.Sprintf("%d", n)
}

func formatRating(r float64) string {
	return fmt.Sprintf("%.1f ★", r)
}

func formatPercent(p float64) string {
	return fmt.Sprintf("%.0f%%", p*100)
}

func formatTokensFloat(n float64) string {
	return util.FormatTokens(n)
}

func formatUsagePercent(p float64) string {
	return fmt.Sprintf("%.0f%%", p)
}

func buildSessionsExportURL(format, experiment string, limit int) templ.SafeURL {
	url := fmt.Sprintf("/api/export/sessions?format=%s&limit=%d", format, limit)
	if experiment != "" {
		url += "&experiment=" + experiment
	}
	return templ.SafeURL(url)
}

func buildDashboardURL(period, experiment, project string) templ.SafeURL {
	url := "/?"
	params := []string{}
	if period != "" {
		params = append(params, "period="+period)
	}
	if experiment != "" {
		params = append(params, "experiment="+experiment)
	}
	if project != "" {
		params = append(params, "project="+project)
	}
	for i, p := range params {
		if i > 0 {
			url += "&"
		}
		url += p
	}
	return templ.SafeURL(url)
}

func tokenBarWidth(tokens, maxTokens int64) string {
	if maxTokens == 0 {
		return "width: 0%"
	}
	pct := float64(tokens) / float64(maxTokens) * 100
	return fmt.Sprintf("width: %.0f%%", pct)
}

func shortModelName(modelID string) string {
	// Strip common prefixes for compact display
	// e.g. "claude-opus-4-5-20250929" -> "opus-4-5"
	// e.g. "claude-sonnet-4-5-20250929" -> "sonnet-4-5"
	// e.g. "claude-haiku-4-5-20251001" -> "haiku-4-5"
	parts := modelID
	if len(parts) > 7 && parts[:7] == "claude-" {
		parts = parts[7:]
	}
	// Strip trailing date suffix (e.g. "-20250929")
	if len(parts) > 9 && parts[len(parts)-9] == '-' {
		candidate := parts[len(parts)-8:]
		allDigits := true
		for _, c := range candidate {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			parts = parts[:len(parts)-9]
		}
	}
	return parts
}

func planDisplayName(planType string) string {
	switch planType {
	case "pro":
		return "Pro"
	case "max_5x":
		return "Max 5x"
	case "max_20x":
		return "Max 20x"
	default:
		return planType
	}
}
