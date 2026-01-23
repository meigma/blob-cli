package open

import "github.com/charmbracelet/bubbles/key"

// keyMap defines the key bindings for the TUI.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Tab    key.Binding
	Copy   key.Binding
	Quit   key.Binding
	Escape key.Binding
	Help   key.Binding
}

// keys is the default key mapping.
var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "backspace"),
		key.WithHelp("←/⌫", "parent dir"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "enter/preview"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "enter/confirm"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy file"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel/quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// ShortHelp returns key bindings for the short help view.
//
//nolint:gocritic // hugeParam: value receiver required by help.KeyMap interface
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Tab, k.Copy, k.Quit}
}

// FullHelp returns key bindings for the full help view.
//
//nolint:gocritic // hugeParam: value receiver required by help.KeyMap interface
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.Copy, k.Quit, k.Help},
	}
}
