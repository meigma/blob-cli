package open

import "github.com/charmbracelet/lipgloss"

// Styles holds the TUI styles.
type Styles struct {
	// Focus indicator colors
	FocusedBorder   lipgloss.Color
	UnfocusedBorder lipgloss.Color

	// Text styles
	Title    lipgloss.Style
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Dir      lipgloss.Style
	Error    lipgloss.Style
	Hint     lipgloss.Style
}

// DefaultStyles returns the default style configuration.
func DefaultStyles() Styles {
	return Styles{
		FocusedBorder:   lipgloss.Color("62"),
		UnfocusedBorder: lipgloss.Color("240"),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Bold(true),

		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		Dir: lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}
