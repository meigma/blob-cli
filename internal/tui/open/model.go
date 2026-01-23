// Package open provides the TUI for the blob open command.
package open

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/meigma/blob"

	"github.com/meigma/blob-cli/internal/tui/components/copydialog"
	"github.com/meigma/blob-cli/internal/tui/components/filetree"
	"github.com/meigma/blob-cli/internal/tui/components/preview"
	"github.com/meigma/blob-cli/internal/tui/components/statusbar"
)

// state represents the current TUI state.
type state int

const (
	stateLoading state = iota
	stateReady
	stateError
)

// focus indicates which pane has focus.
type focus int

const (
	focusTree focus = iota
	focusPreview
)

// LoadFunc is a function that loads the archive data.
// It's called asynchronously in Init().
type LoadFunc func() (*blob.IndexView, *blob.Archive, error)

// Model is the main TUI model for blob open.
type Model struct {
	// Loading state
	state   state
	loader  LoadFunc
	loadErr error
	spinner spinner.Model

	// Archive data (set after loading)
	ref     string
	index   *blob.IndexView
	archive *blob.Archive

	// Components (initialized after loading)
	tree       filetree.Model
	preview    preview.Model
	copyDialog copydialog.Model
	statusBar  statusbar.Model
	help       help.Model

	// State
	focus    focus
	showHelp bool
	styles   Styles

	// Dimensions
	width  int
	height int
	ready  bool
}

// New creates a new TUI model in loading state.
// The loader function will be called asynchronously to fetch the archive.
func New(ref string, loader LoadFunc) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		state:   stateLoading,
		ref:     ref,
		loader:  loader,
		spinner: s,
		styles:  DefaultStyles(),
	}
}
