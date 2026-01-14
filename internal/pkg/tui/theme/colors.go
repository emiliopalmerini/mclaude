package theme

import "github.com/charmbracelet/lipgloss"

// Color palette inspired by Claude Code
var (
	// Primary colors
	Purple       = lipgloss.Color("#A855F7")
	BrightPurple = lipgloss.Color("#C084FC")
	DarkPurple   = lipgloss.Color("#7C3AED")

	// Neutrals
	White     = lipgloss.Color("#FFFFFF")
	LightGray = lipgloss.Color("#9CA3AF")
	DimGray   = lipgloss.Color("#6B7280")
	DarkGray  = lipgloss.Color("#374151")
	Black     = lipgloss.Color("#111827")

	// Semantic colors
	Success = lipgloss.Color("#22C55E")
	Warning = lipgloss.Color("#F59E0B")
	Error   = lipgloss.Color("#EF4444")
	Info    = lipgloss.Color("#3B82F6")

	// Accent colors
	Cyan   = lipgloss.Color("#06B6D4")
	Orange = lipgloss.Color("#F97316")
)
