package theme

import "github.com/charmbracelet/lipgloss"

// Monochrome color palette inspired by Google Fonts Korean
// Clean, typography-focused aesthetic with high contrast
var (
	// Primary - Pure black and white for maximum contrast
	White = lipgloss.Color("#FFFFFF")
	Black = lipgloss.Color("#000000")

	// Gray scale - carefully graduated for hierarchy
	Gray100 = lipgloss.Color("#F5F5F5") // Lightest - backgrounds
	Gray200 = lipgloss.Color("#E5E5E5") // Borders, dividers
	Gray300 = lipgloss.Color("#D4D4D4") // Disabled states
	Gray400 = lipgloss.Color("#A3A3A3") // Muted text
	Gray500 = lipgloss.Color("#737373") // Secondary text
	Gray600 = lipgloss.Color("#525252") // Body text
	Gray700 = lipgloss.Color("#404040") // Strong text
	Gray800 = lipgloss.Color("#262626") // Headings
	Gray900 = lipgloss.Color("#171717") // Darkest

	// Semantic colors - muted to preserve monochrome feel
	Success = lipgloss.Color("#525252") // Use text differentiation instead
	Warning = lipgloss.Color("#525252")
	Error   = lipgloss.Color("#737373")
	Info    = lipgloss.Color("#525252")

	// Legacy aliases for backward compatibility
	LightGray = Gray400
	DimGray   = Gray500
	DarkGray  = Gray700

	// Accent - inverted for selection/active states
	Accent       = White
	AccentBg     = Gray900
	BrightPurple = White // Legacy alias
)
