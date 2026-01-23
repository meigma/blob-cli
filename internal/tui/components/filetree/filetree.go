// Package filetree provides a file browser component for the TUI.
package filetree

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/meigma/blob"

	"github.com/meigma/blob-cli/internal/archive"
)

// Model represents the file tree component state.
type Model struct {
	index      *blob.IndexView
	currentDir string
	entries    []*archive.DirEntry
	cursor     int
	offset     int // scroll offset
	width      int
	height     int
	focused    bool
	history    []historyEntry // navigation history for Back
}

// historyEntry stores state for navigation history.
type historyEntry struct {
	dir    string
	cursor int
	offset int
}

// New creates a new file tree component.
func New(index *blob.IndexView) Model {
	m := Model{
		index:   index,
		history: make([]historyEntry, 0),
	}
	m.loadDir("")
	return m
}

// SetSize updates the component dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.adjustScroll()
}

// SetFocused sets the focus state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// Focused returns whether the component is focused.
func (m *Model) Focused() bool {
	return m.focused
}

// Selected returns the currently selected entry, or nil if none.
func (m *Model) Selected() *archive.DirEntry {
	if len(m.entries) == 0 {
		return nil
	}
	return m.entries[m.cursor]
}

// CurrentDir returns the current directory path.
func (m *Model) CurrentDir() string {
	return m.currentDir
}

// EntryCount returns the number of entries in the current directory.
func (m *Model) EntryCount() int {
	return len(m.entries)
}

// CursorUp moves the cursor up one item.
func (m *Model) CursorUp() {
	if m.cursor > 0 {
		m.cursor--
		m.adjustScroll()
	}
}

// CursorDown moves the cursor down one item.
func (m *Model) CursorDown() {
	if m.cursor < len(m.entries)-1 {
		m.cursor++
		m.adjustScroll()
	}
}

// Enter enters the selected directory or returns the selected file.
// Returns true if a directory was entered, false if a file was selected.
func (m *Model) Enter() bool {
	selected := m.Selected()
	if selected == nil {
		return false
	}

	if selected.IsDir {
		// Save current state to history
		m.history = append(m.history, historyEntry{
			dir:    m.currentDir,
			cursor: m.cursor,
			offset: m.offset,
		})
		m.loadDir(selected.Path)
		return true
	}
	return false
}

// Back goes to the parent directory.
// Returns true if navigation occurred.
func (m *Model) Back() bool {
	// First try history
	if len(m.history) > 0 {
		h := m.history[len(m.history)-1]
		m.history = m.history[:len(m.history)-1]
		m.loadDir(h.dir)
		m.cursor = h.cursor
		m.offset = h.offset
		m.adjustScroll()
		return true
	}

	// Otherwise try parent directory
	if m.currentDir == "" {
		return false
	}

	parent := parentPath(m.currentDir)
	m.loadDir(parent)
	return true
}

// loadDir loads entries for a directory.
func (m *Model) loadDir(dir string) {
	m.currentDir = dir
	m.cursor = 0
	m.offset = 0

	entries, err := archive.ListDir(m.index, dir)
	if err != nil {
		m.entries = nil
		return
	}

	// Sort directories first
	archive.SortDirsFirst(entries)
	m.entries = entries
}

// adjustScroll ensures the cursor is visible within the viewport.
func (m *Model) adjustScroll() {
	visibleLines := m.visibleLines()
	if visibleLines <= 0 {
		return
	}

	// Scroll up if cursor is above viewport
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// Scroll down if cursor is below viewport
	if m.cursor >= m.offset+visibleLines {
		m.offset = m.cursor - visibleLines + 1
	}
}

// visibleLines returns the number of visible lines in the viewport.
func (m *Model) visibleLines() int {
	// Account for: 2 border lines + 1 header + 1 separator
	return m.height - 6
}

// parentPath returns the parent directory path.
func parentPath(p string) string {
	if p == "" {
		return ""
	}
	idx := strings.LastIndex(p, "/")
	if idx == -1 {
		return ""
	}
	return p[:idx]
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
	return m, nil
}

// viewStyles holds the styles used for rendering.
type viewStyles struct {
	header   lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	dir      lipgloss.Style
	box      lipgloss.Style
}

// newViewStyles creates styles based on focus state.
func newViewStyles(focused bool, width, height int) viewStyles {
	borderColor := lipgloss.Color("240")
	if focused {
		borderColor = lipgloss.Color("62")
	}

	return viewStyles{
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Bold(true),
		normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		dir: lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")),
		box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Width(width - 2).
			Height(height - 2),
	}
}

// formatEntry formats a single entry for display.
func (m *Model) formatEntry(entry *archive.DirEntry, index int, styles *viewStyles) string {
	name := entry.Name
	if entry.IsDir {
		name += "/"
	}

	var line string
	switch {
	case index == m.cursor && m.focused:
		line = styles.selected.Render("> " + name)
	case index == m.cursor:
		line = styles.normal.Render("> " + name)
	case entry.IsDir:
		line = styles.dir.Render("  " + name)
	default:
		line = styles.normal.Render("  " + name)
	}

	// Truncate if too wide
	maxWidth := m.width - 6
	if len(line) > maxWidth && maxWidth > 3 {
		line = line[:maxWidth-3] + "..."
	}

	return line
}

// View renders the component.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) View() string {
	pathDisplay := "/" + m.currentDir
	if m.currentDir == "" {
		pathDisplay = "/"
	}

	styles := newViewStyles(m.focused, m.width, m.height)

	// Build entry list
	var lines []string
	visibleLines := m.visibleLines()
	for i := m.offset; i < len(m.entries) && i < m.offset+visibleLines; i++ {
		lines = append(lines, m.formatEntry(m.entries[i], i, &styles))
	}

	// Pad with empty lines if needed
	for len(lines) < visibleLines {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		styles.header.Render(pathDisplay),
		strings.Repeat("─", m.width-4),
		strings.Join(lines, "\n"),
	)

	return styles.box.Render(content)
}
