// Package preview provides a content preview component for the TUI.
package preview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State represents the current state of the preview.
type State int

const (
	StateNone     State = iota // No file selected
	StateLoading               // Loading file content
	StateText                  // Displaying text content
	StateBinary                // Displaying hex dump
	StateError                 // Error loading file
	StateDir                   // Directory selected (no preview)
	StateTooLarge              // File too large for preview
)

// MaxPreviewBytes is the maximum size of file content to preview.
// Files larger than this will show a "too large" message instead of loading.
const MaxPreviewBytes = 512 * 1024 // 512KB

// Model represents the preview component state.
type Model struct {
	viewport viewport.Model
	state    State
	path     string
	language string // Detected language for syntax highlighting
	errMsg   string
	width    int
	height   int
	focused  bool
	ready    bool
}

// New creates a new preview component.
func New() Model {
	return Model{
		state: StateNone,
	}
}

// SetSize updates the component dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Viewport height needs to account for: 2 border + 1 header + 1 separator
	vpWidth := width - 4 // Account for border padding
	vpHeight := max(height-6, 1)
	if !m.ready {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.ready = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
	}
}

// SetFocused sets the focus state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// Focused returns whether the component is focused.
func (m *Model) Focused() bool {
	return m.focused
}

// SetLoading shows the loading state for a path.
func (m *Model) SetLoading(path string) {
	m.state = StateLoading
	m.path = path
	m.language = ""
	m.errMsg = ""
	if m.ready {
		m.viewport.SetContent(fmt.Sprintf("Loading %s...", path))
		m.viewport.GotoTop()
	}
}

// SetContent sets the content to display.
func (m *Model) SetContent(path string, content []byte, isBinary bool) {
	m.path = path
	m.errMsg = ""

	if isBinary {
		m.state = StateBinary
		// Truncate for hex display
		displayContent := content
		truncated := false
		if len(displayContent) > MaxPreviewBytes {
			displayContent = displayContent[:MaxPreviewBytes]
			truncated = true
		}
		hexContent := FormatHex(displayContent)
		if truncated {
			hexContent += fmt.Sprintf("\n\n... (truncated, showing first %d bytes)", MaxPreviewBytes)
		}
		if m.ready {
			m.viewport.SetContent(hexContent)
			m.viewport.GotoTop()
		}
	} else {
		m.state = StateText
		displayContent := content
		truncated := false
		if len(content) > MaxPreviewBytes {
			displayContent = content[:MaxPreviewBytes]
			truncated = true
		}

		// Apply syntax highlighting if available
		var text string
		m.language = GetLanguage(path)
		if m.language != "" {
			text = Highlight(path, displayContent)
		} else {
			text = string(displayContent)
		}

		if truncated {
			text += fmt.Sprintf("\n\n... (truncated, showing first %d bytes)", MaxPreviewBytes)
		}

		if m.ready {
			// Wrap text to viewport width
			wrapped := m.wrapText(text)
			m.viewport.SetContent(wrapped)
			m.viewport.GotoTop()
		}
	}
}

// wrapText wraps text to fit the viewport width.
func (m *Model) wrapText(text string) string {
	if m.viewport.Width <= 0 {
		return text
	}
	return lipgloss.NewStyle().Width(m.viewport.Width).Render(text)
}

// SetError shows an error message.
func (m *Model) SetError(path string, err error) {
	m.state = StateError
	m.path = path
	m.language = ""
	m.errMsg = err.Error()
	if m.ready {
		m.viewport.SetContent(fmt.Sprintf("Error loading %s:\n\n%s", path, err.Error()))
		m.viewport.GotoTop()
	}
}

// SetDir shows directory selected state.
func (m *Model) SetDir(path string) {
	m.state = StateDir
	m.path = path
	m.language = ""
	m.errMsg = ""
	if m.ready {
		m.viewport.SetContent(fmt.Sprintf("Directory: %s\n\nPress Enter to browse contents", path))
		m.viewport.GotoTop()
	}
}

// SetTooLarge shows the file-too-large state.
func (m *Model) SetTooLarge(path string, size uint64) {
	m.state = StateTooLarge
	m.path = path
	m.language = ""
	m.errMsg = ""
	if m.ready {
		content := fmt.Sprintf(
			"File too large for preview\n\n"+
				"Path: %s\n"+
				"Size: %s\n"+
				"Limit: %s\n\n"+
				"Press 'c' to copy this file to local filesystem",
			path,
			formatBytes(size),
			formatBytes(MaxPreviewBytes),
		)
		m.viewport.SetContent(content)
		m.viewport.GotoTop()
	}
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

// SetNone clears the preview.
func (m *Model) SetNone() {
	m.state = StateNone
	m.path = ""
	m.language = ""
	m.errMsg = ""
	if m.ready {
		m.viewport.SetContent("No file selected")
		m.viewport.GotoTop()
	}
}

// Path returns the current path being previewed.
func (m *Model) Path() string {
	return m.path
}

// State returns the current state.
func (m *Model) State() State {
	return m.state
}

// ScrollPercent returns the scroll position as a percentage.
func (m *Model) ScrollPercent() float64 {
	return m.viewport.ScrollPercent()
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
	if !m.ready {
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the component.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Build header with scroll indicator for scrollable content
	var header string
	switch m.state {
	case StateNone:
		header = "Preview"
	case StateLoading:
		header = "Loading: " + m.path
	case StateText:
		header = m.buildHeader("Text", m.path)
	case StateBinary:
		header = m.buildHeader("Binary", m.path)
	case StateError:
		header = "Error: " + m.path
	case StateDir:
		header = "Directory: " + m.path
	case StateTooLarge:
		header = "Too Large: " + m.path
	}

	// Style based on focus
	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(header),
		strings.Repeat("─", m.width-4),
		m.viewport.View(),
	)

	return boxStyle.Render(content)
}

// buildHeader creates a header with optional language and scroll indicator.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) buildHeader(prefix, path string) string {
	base := prefix + ": " + path

	// Add detected language if available
	if m.language != "" {
		base += " (" + m.language + ")"
	}

	// Only show scroll info if content is scrollable
	if m.viewport.TotalLineCount() <= m.viewport.Height {
		return base
	}

	// Show scroll percentage
	percent := int(m.viewport.ScrollPercent() * 100)
	scrollInfo := fmt.Sprintf(" [%d%%]", percent)

	// Add hint when focused
	if m.focused {
		scrollInfo += " (↑↓ to scroll)"
	}

	return base + scrollInfo
}
