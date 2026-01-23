// Package copydialog provides a modal dialog for copying files.
package copydialog

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the copy dialog component state.
type Model struct {
	input      textinput.Model
	sourcePath string
	visible    bool
	width      int
	height     int
}

// New creates a new copy dialog component.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "destination path"
	ti.CharLimit = 256
	ti.Width = 40

	return Model{
		input: ti,
	}
}

// Show displays the dialog for copying a file.
func (m *Model) Show(sourcePath string) {
	m.sourcePath = sourcePath
	m.visible = true

	// Set default destination to current directory with source filename
	baseName := filepath.Base(sourcePath)
	m.input.SetValue(baseName)
	m.input.Focus()
	m.input.CursorEnd()
}

// Hide hides the dialog.
func (m *Model) Hide() {
	m.visible = false
	m.input.Blur()
}

// Visible returns whether the dialog is visible.
func (m *Model) Visible() bool {
	return m.visible
}

// SourcePath returns the source file path.
func (m *Model) SourcePath() string {
	return m.sourcePath
}

// Destination returns the entered destination path.
func (m *Model) Destination() string {
	return m.input.Value()
}

// SetSize updates the dialog dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Adjust input width based on dialog size
	inputWidth := min(width-10, 60)
	inputWidth = max(inputWidth, 20)
	m.input.Width = inputWidth
}

// Init initializes the component.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the component.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	dialogWidth := 50
	if m.width > 0 && m.width < dialogWidth+4 {
		dialogWidth = m.width - 4
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(dialogWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Copy File"),
		"",
		labelStyle.Render("Source: "+m.sourcePath),
		"",
		labelStyle.Render("Destination:"),
		m.input.View(),
		"",
		hintStyle.Render("Enter: confirm  Esc: cancel"),
	)

	return borderStyle.Render(content)
}
