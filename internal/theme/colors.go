package theme

import "github.com/charmbracelet/lipgloss"

// Dracula Theme official color palette
// Spec: https://spec.draculatheme.com/
var (
	// Background colors
	Background     = lipgloss.Color("#282A36")
	CurrentLine    = lipgloss.Color("#44475A")
	Selection      = lipgloss.Color("#44475A")
	DarkerBg       = lipgloss.Color("#21222C")

	// Foreground colors
	Foreground     = lipgloss.Color("#F8F8F2")
	Comment        = lipgloss.Color("#6272A4")
	Muted          = lipgloss.Color("#6272A4")

	// Accent colors
	Cyan           = lipgloss.Color("#8BE9FD")
	Green          = lipgloss.Color("#50FA7B")
	Orange         = lipgloss.Color("#FFB86C")
	Pink           = lipgloss.Color("#FF79C6")
	Purple         = lipgloss.Color("#BD93F9")
	Red            = lipgloss.Color("#FF5555")
	Yellow         = lipgloss.Color("#F1FA8C")
)
