package open

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) View() string {
	switch m.state {
	case stateLoading:
		return m.viewLoading()
	case stateError:
		return m.viewError()
	case stateReady:
		return m.viewReady()
	}
	return ""
}

// viewLoading renders the loading screen.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) viewLoading() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	refStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	message := fmt.Sprintf("%s Loading %s...", m.spinner.View(), refStyle.Render(m.ref))

	// Center the message if we have dimensions
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			spinnerStyle.Render(message),
		)
	}

	return spinnerStyle.Render(message)
}

// viewError renders the error screen.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) viewError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	message := lipgloss.JoinVertical(lipgloss.Center,
		errorStyle.Render("Error loading archive"),
		"",
		m.loadErr.Error(),
		"",
		hintStyle.Render("Press q to quit"),
	)

	// Center the message if we have dimensions
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			message,
		)
	}

	return message
}

// viewReady renders the main browser interface.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) viewReady() string {
	if !m.ready {
		return "Initializing..."
	}

	// Build the main layout
	treeView := m.tree.View()
	previewView := m.preview.View()

	// Join tree and preview horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, treeView, previewView)

	// Add status bar at the bottom
	statusView := m.statusBar.View()
	fullView := lipgloss.JoinVertical(lipgloss.Left, mainContent, statusView)

	// Overlay copy dialog if visible
	if m.copyDialog.Visible() {
		fullView = m.overlayDialog(fullView)
	}

	// Overlay help if visible
	if m.showHelp {
		fullView = m.overlayHelp(fullView)
	}

	return fullView
}

// overlayDialog overlays the copy dialog centered on the screen.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) overlayDialog(_ string) string {
	dialog := m.copyDialog.View()

	// Create the overlay by placing dialog on top of background
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

// overlayHelp overlays the help panel centered on the screen.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) overlayHelp(_ string) string {
	// Build help content using the full help view
	helpContent := m.help.View(keys)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229")).
		MarginBottom(1)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Keyboard Shortcuts"),
		helpContent,
		hintStyle.Render("Press ? or Esc to close"),
	)

	dialog := boxStyle.Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}
