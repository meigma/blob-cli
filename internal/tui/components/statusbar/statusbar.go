// Package statusbar provides a status bar component for the TUI.
package statusbar

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// errorDuration is how long error messages are shown.
const errorDuration = 5 * time.Second

// Model represents the status bar component state.
type Model struct {
	ref        string
	path       string
	entryCount int
	message    string
	isError    bool
	messageExp time.Time // when message expires
	width      int

	// Selected file metadata
	selectedName  string
	selectedSize  uint64
	selectedTime  time.Time
	selectedIsDir bool
	hasSelection  bool
}

// New creates a new status bar component.
func New(ref string) Model {
	return Model{
		ref: ref,
	}
}

// SetRef updates the archive reference.
func (m *Model) SetRef(ref string) {
	m.ref = ref
}

// SetPath updates the current path display.
func (m *Model) SetPath(path string) {
	m.path = path
}

// SetEntryCount updates the entry count display.
func (m *Model) SetEntryCount(count int) {
	m.entryCount = count
}

// SetMessage sets a transient message.
func (m *Model) SetMessage(msg string) {
	m.message = msg
	m.isError = false
	m.messageExp = time.Now().Add(errorDuration)
}

// SetError sets a transient error message.
func (m *Model) SetError(err error) {
	m.message = err.Error()
	m.isError = true
	m.messageExp = time.Now().Add(errorDuration)
}

// SetWidth updates the status bar width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetSelectedFile updates the selected file metadata display.
func (m *Model) SetSelectedFile(name string, size uint64, modTime time.Time, isDir bool) {
	m.selectedName = name
	m.selectedSize = size
	m.selectedTime = modTime
	m.selectedIsDir = isDir
	m.hasSelection = true
}

// ClearSelection clears the selected file metadata.
func (m *Model) ClearSelection() {
	m.hasSelection = false
	m.selectedName = ""
}

// formatSelectionInfo formats the selected file/directory metadata.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) formatSelectionInfo(style lipgloss.Style) string {
	if m.selectedIsDir {
		return style.Render("directory")
	}

	// Format: "4.2 KB · Jan 15 10:30" or "4.2 KB · 2d ago"
	size := formatBytes(m.selectedSize)
	timeStr := formatTime(m.selectedTime)

	if timeStr == "" {
		return style.Render(size)
	}
	return style.Render(size + " · " + timeStr)
}

// ClearMessage clears any transient message if expired.
func (m *Model) ClearMessage() {
	if m.message != "" && time.Now().After(m.messageExp) {
		m.message = ""
		m.isError = false
	}
}

// ClearMessageMsg is sent to clear expired messages.
type ClearMessageMsg struct{}

// ScheduleClear returns a command to clear the message after the error duration.
//
//nolint:gocritic // hugeParam: value receiver matches tea.Cmd pattern
func (m Model) ScheduleClear() tea.Cmd {
	return tea.Tick(errorDuration, func(t time.Time) tea.Msg {
		return ClearMessageMsg{}
	})
}

// Init initializes the component.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if _, ok := msg.(ClearMessageMsg); ok {
		m.ClearMessage()
	}
	return m, nil
}

// View renders the component.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) View() string {
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.width)

	refStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("75")).
		Bold(true)

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	msgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Build left section: ref and path
	left := refStyle.Render(m.ref)
	if m.path != "" {
		left += " " + pathStyle.Render(m.path)
	}

	// Build middle section: message, file metadata, or entry count
	var middle string
	if m.message != "" && time.Now().Before(m.messageExp) {
		if m.isError {
			middle = errorStyle.Render(m.message)
		} else {
			middle = msgStyle.Render(m.message)
		}
	} else if m.hasSelection {
		middle = m.formatSelectionInfo(countStyle)
	} else if m.entryCount > 0 {
		middle = countStyle.Render(fmt.Sprintf("%d items", m.entryCount))
	}

	// Build right section: help hints
	right := helpStyle.Render("q:quit  c:copy  Tab:focus  ?:help")

	// Calculate spacing
	leftLen := lipgloss.Width(left)
	middleLen := lipgloss.Width(middle)
	rightLen := lipgloss.Width(right)

	// Distribute space
	totalContent := leftLen + middleLen + rightLen
	if totalContent >= m.width {
		// Content too long, just show left and right
		gap := max(m.width-leftLen-rightLen, 0)
		return barStyle.Render(left + spaces(gap) + right)
	}

	// Center the middle, with left and right at edges
	gapLeft := max((m.width-middleLen)/2-leftLen, 1)
	gapRight := max(m.width-leftLen-gapLeft-middleLen-rightLen, 1)

	return barStyle.Render(left + spaces(gapLeft) + middle + spaces(gapRight) + right)
}

// spaces returns n space characters.
func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// formatBytes formats a byte count in human-readable form.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatTime formats a time as relative (if recent) or absolute.
// Shows relative time for times within 7 days, absolute otherwise.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	now := time.Now()
	diff := now.Sub(t)

	// Use relative time for recent times (within 7 days)
	if diff >= 0 && diff < 7*24*time.Hour {
		return formatRelativeTime(diff)
	}

	// Use absolute time for older or future times
	// Same year: "Jan 15 10:30"
	// Different year: "Jan 15 2023"
	if t.Year() == now.Year() {
		return t.Format("Jan 2 15:04")
	}
	return t.Format("Jan 2 2006")
}

// formatRelativeTime formats a duration as a relative time string.
func formatRelativeTime(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}
