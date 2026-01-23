package open

import (
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/meigma/blob-cli/internal/tui/components/copydialog"
	"github.com/meigma/blob-cli/internal/tui/components/filetree"
	"github.com/meigma/blob-cli/internal/tui/components/preview"
	"github.com/meigma/blob-cli/internal/tui/components/statusbar"
	"github.com/meigma/blob-cli/internal/tui/detect"
)

// Init initializes the model.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) Init() tea.Cmd {
	// Start spinner and kick off archive loading
	return tea.Batch(
		m.spinner.Tick,
		m.loadArchive(),
	)
}

// loadArchive returns a command that loads the archive asynchronously.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) loadArchive() tea.Cmd {
	loader := m.loader
	return func() tea.Msg {
		index, archive, err := loader()
		if err != nil {
			return ArchiveErrorMsg{Err: err}
		}
		return ArchiveLoadedMsg{Index: index, Archive: archive}
	}
}

// Update handles messages and updates the model.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == stateReady {
			return m.handleResize(msg)
		}
		return m, nil

	case tea.KeyMsg:
		// Allow quitting with 'q' in any state
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		// Escape is handled per-state (may close dialogs/help first)
	}

	// Route to state-specific handler
	switch m.state {
	case stateLoading:
		return m.updateLoading(msg)
	case stateError:
		return m.updateError(msg)
	case stateReady:
		return m.updateReady(msg)
	}

	return m, nil
}

// updateLoading handles messages during the loading state.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keys.Escape) {
			return m, tea.Quit
		}

	case ArchiveLoadedMsg:
		// Initialize components now that we have the archive
		m.state = stateReady
		m.index = msg.Index
		m.archive = msg.Archive
		m.tree = filetree.New(msg.Index)
		m.preview = preview.New()
		m.copyDialog = copydialog.New()
		m.statusBar = statusbar.New(m.ref)
		m.help = help.New()

		// Set initial focus
		m.tree.SetFocused(true)
		m.preview.SetFocused(false)

		// Apply dimensions if we already received a WindowSizeMsg
		var cmds []tea.Cmd
		if m.width > 0 && m.height > 0 {
			resized, cmd := m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			if resizedModel, ok := resized.(Model); ok {
				m = resizedModel
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		// Initialize child components
		cmds = append(cmds,
			m.tree.Init(),
			m.preview.Init(),
			m.copyDialog.Init(),
			m.statusBar.Init(),
		)

		// Load initial preview and selection status for the first selected item
		m.updateSelectionStatus()
		if cmd := m.loadSelectedPreview(); cmd != nil {
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case ArchiveErrorMsg:
		m.state = stateError
		m.loadErr = msg.Err
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateError handles messages during the error state.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) updateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, keys.Escape) {
			return m, tea.Quit
		}
	}
	return m, nil
}

// updateReady handles messages when the archive is loaded.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) updateReady(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle copy dialog if visible
		if m.copyDialog.Visible() {
			return m.handleCopyDialogKeys(msg)
		}
		return m.handleKeys(msg)

	case FileContentMsg:
		m.preview.SetContent(msg.Path, msg.Content, msg.IsBinary)
		return m, nil

	case FileErrorMsg:
		m.preview.SetError(msg.Path, msg.Err)
		m.statusBar.SetError(msg.Err)
		return m, m.statusBar.ScheduleClear()

	case CopyCompleteMsg:
		m.copyDialog.Hide()
		m.statusBar.SetMessage("Copied to " + msg.DestPath)
		return m, m.statusBar.ScheduleClear()

	case CopyErrorMsg:
		m.copyDialog.Hide()
		m.statusBar.SetError(msg.Err)
		return m, m.statusBar.ScheduleClear()

	case statusbar.ClearMessageMsg:
		m.statusBar, _ = m.statusBar.Update(msg)
		return m, nil
	}

	// Forward messages to copy dialog if visible
	if m.copyDialog.Visible() {
		var cmd tea.Cmd
		m.copyDialog, cmd = m.copyDialog.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Forward to focused component
	if m.focus == focusPreview {
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleResize handles window resize events when in ready state.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true

	// Calculate layout (40% tree, 60% preview)
	treeWidth := m.width * 40 / 100
	previewWidth := m.width - treeWidth

	// Height: full height minus status bar (1 line)
	contentHeight := m.height - 1

	m.tree.SetSize(treeWidth, contentHeight)
	m.preview.SetSize(previewWidth, contentHeight)
	m.copyDialog.SetSize(m.width, m.height)
	m.statusBar.SetWidth(m.width)

	// Update status bar with entry count
	m.statusBar.SetPath(m.tree.CurrentDir())
	m.statusBar.SetEntryCount(m.tree.EntryCount())

	return m, nil
}

// handleKeys handles key presses in normal mode.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys that work in any focus state
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Escape):
		// If help is showing, close it; otherwise quit
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp
		return m, nil

	case key.Matches(msg, keys.Tab):
		return m.toggleFocus(), nil

	case key.Matches(msg, keys.Copy):
		return m.startCopy()
	}

	// Focus-specific handling
	if m.focus == focusTree {
		return m.handleTreeKeys(msg)
	}
	return m.handlePreviewKeys(msg)
}

// handleTreeKeys handles key presses when the tree is focused.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) handleTreeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		m.tree.CursorUp()
		m.updateSelectionStatus()
		return m, m.loadSelectedPreview()

	case key.Matches(msg, keys.Down):
		m.tree.CursorDown()
		m.updateSelectionStatus()
		return m, m.loadSelectedPreview()

	case key.Matches(msg, keys.Left):
		if m.tree.Back() {
			m.updateStatusBar()
			m.updateSelectionStatus()
			return m, m.loadSelectedPreview()
		}
		return m, nil

	case key.Matches(msg, keys.Right), key.Matches(msg, keys.Enter):
		if m.tree.Enter() {
			// Entered a directory
			m.updateStatusBar()
			m.updateSelectionStatus()
			return m, m.loadSelectedPreview()
		}
		// Selected a file - load preview
		m.updateSelectionStatus()
		return m, m.loadSelectedPreview()
	}

	return m, nil
}

// handlePreviewKeys handles key presses when the preview is focused.
// Most keys are forwarded to the viewport for scrolling.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Forward all navigation keys to the viewport
	// The viewport handles: up/down/j/k, page up/down, g/G (home/end), ctrl+u/d
	var cmd tea.Cmd
	m.preview, cmd = m.preview.Update(msg)
	return m, cmd
}

// handleCopyDialogKeys handles key presses in copy dialog mode.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) handleCopyDialogKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.copyDialog.Hide()
		return m, nil

	case key.Matches(msg, keys.Enter):
		return m.executeCopy()
	}

	// Forward other keys to the text input
	var cmd tea.Cmd
	m.copyDialog, cmd = m.copyDialog.Update(msg)
	return m, cmd
}

// toggleFocus switches focus between tree and preview.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) toggleFocus() Model {
	if m.focus == focusTree {
		m.focus = focusPreview
		m.tree.SetFocused(false)
		m.preview.SetFocused(true)
	} else {
		m.focus = focusTree
		m.tree.SetFocused(true)
		m.preview.SetFocused(false)
	}
	return m
}

// updateSelectionStatus updates the status bar with the currently selected file's metadata.
func (m *Model) updateSelectionStatus() {
	selected := m.tree.Selected()
	if selected == nil {
		m.statusBar.ClearSelection()
		return
	}
	m.statusBar.SetSelectedFile(selected.Name, selected.Size, selected.ModTime, selected.IsDir)
}

// loadSelectedPreview loads the preview for the currently selected item.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) loadSelectedPreview() tea.Cmd {
	selected := m.tree.Selected()
	if selected == nil {
		m.preview.SetNone()
		return nil
	}

	if selected.IsDir {
		m.preview.SetDir(selected.Path)
		return nil
	}

	// Check file size before loading to prevent memory issues
	if selected.Size > preview.MaxPreviewBytes {
		m.preview.SetTooLarge(selected.Path, selected.Size)
		return nil
	}

	// Load file content asynchronously
	m.preview.SetLoading(selected.Path)
	path := selected.Path
	archive := m.archive

	return func() tea.Msg {
		content, err := archive.ReadFile(path)
		if err != nil {
			return FileErrorMsg{Path: path, Err: err}
		}
		isBinary := detect.IsBinary(content)
		return FileContentMsg{Path: path, Content: content, IsBinary: isBinary}
	}
}

// updateStatusBar updates the status bar with current state.
func (m *Model) updateStatusBar() {
	m.statusBar.SetPath(m.tree.CurrentDir())
	m.statusBar.SetEntryCount(m.tree.EntryCount())
}

// startCopy initiates the copy dialog for the selected file.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) startCopy() (tea.Model, tea.Cmd) {
	selected := m.tree.Selected()
	if selected == nil || selected.IsDir {
		m.statusBar.SetMessage("Select a file to copy")
		return m, m.statusBar.ScheduleClear()
	}

	m.copyDialog.Show(selected.Path)
	return m, nil
}

// executeCopy performs the file copy operation.
//
//nolint:gocritic // hugeParam: consistent with tea.Model pattern
func (m Model) executeCopy() (tea.Model, tea.Cmd) {
	sourcePath := m.copyDialog.SourcePath()
	destPath := m.copyDialog.Destination()

	if destPath == "" {
		m.statusBar.SetMessage("Destination path required")
		return m, m.statusBar.ScheduleClear()
	}

	archive := m.archive

	return m, func() tea.Msg {
		content, err := archive.ReadFile(sourcePath)
		if err != nil {
			return CopyErrorMsg{SourcePath: sourcePath, DestPath: destPath, Err: err}
		}

		if err := os.WriteFile(destPath, content, 0o600); err != nil {
			return CopyErrorMsg{SourcePath: sourcePath, DestPath: destPath, Err: err}
		}

		return CopyCompleteMsg{SourcePath: sourcePath, DestPath: destPath}
	}
}
