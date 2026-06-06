package warp

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Warp is the root Bubbletea model. It holds a root Panel and forwards
// all messages to it. By default New() creates a Warp with a TabGroup root.
type Warp struct {
	root   Panel
	width  int
	height int
}

// New creates a new Warp with a TabGroup root (one default tab).
func New() *Warp {
	tg := NewTabGroup(TabTop)
	return &Warp{root: tg}
}

// SetRoot replaces the root panel. Use this to install custom layouts
// (splits, flex, nested tab groups, etc.).
func (w *Warp) SetRoot(panel Panel) {
	w.root = panel
}

// Root returns the current root panel.
func (w *Warp) Root() Panel {
	return w.root
}

// tabGroup returns the root TabGroup (if any).
func (w *Warp) tabGroup() *TabGroup {
	if tg, ok := w.root.(*TabGroup); ok {
		return tg
	}
	return nil
}

// --- Convenience delegates (no-op if root is not a TabGroup) ---

// NewTab delegates to the root TabGroup.
func (w *Warp) NewTab(name string) *Tab {
	if tg := w.tabGroup(); tg != nil {
		return tg.NewTab(name)
	}
	return nil
}

// ActiveTab delegates to the root TabGroup.
func (w *Warp) ActiveTab() *Tab {
	if tg := w.tabGroup(); tg != nil {
		return tg.ActiveTab()
	}
	return nil
}

// SetTabPosition delegates to the root TabGroup.
func (w *Warp) SetTabPosition(pos TabPosition) {
	if tg := w.tabGroup(); tg != nil {
		 tg.tabPosition = pos
	}
}

// NextTab delegates to the root TabGroup.
func (w *Warp) NextTab() {
	if tg := w.tabGroup(); tg != nil {
		tg.NextTab()
	}
}

// PrevTab delegates to the root TabGroup.
func (w *Warp) PrevTab() {
	if tg := w.tabGroup(); tg != nil {
		tg.PrevTab()
	}
}

// Init is the Bubbletea initialization.
func (w *Warp) Init() tea.Cmd {
	return nil
}

// Update handles Bubbletea messages.
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return w, tea.Quit
		}
		if w.root != nil {
			return w, w.root.Update(msg)
		}

	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		if w.root != nil {
			return w, w.root.Update(msg)
		}

	case tea.MouseMsg:
		if w.root != nil {
			return w, w.root.Update(msg)
		}
	}

	if w.root != nil {
		return w, w.root.Update(msg)
	}
	return w, nil
}

// View renders the root panel.
func (w *Warp) View() string {
	if w.width == 0 || w.height == 0 {
		return "Loading..."
	}
	if w.root == nil {
		return "No root panel"
	}
	return w.root.View(w.width, w.height)
}

// AsPanel returns a Panel adapter for this Warp, enabling nested warps.
// Note: mouse coordinates are forwarded as-is; nested warps should
// rely on keyboard shortcuts (Ctrl+Tab, etc.) for inner tab switching.
func (w *Warp) AsPanel() Panel {
	return &warpPanel{warp: w}
}

// warpPanel adapts Warp to the Panel interface.
type warpPanel struct {
	warp *Warp
}

func (wp *warpPanel) View(width, height int) string {
	wp.warp.width = width
	wp.warp.height = height
	return wp.warp.View()
}

func (wp *warpPanel) Update(msg tea.Msg) tea.Cmd {
	_, cmd := wp.warp.Update(msg)
	return cmd
}

// Run starts the Bubbletea program.
func (w *Warp) Run() error {
	p := tea.NewProgram(
		w,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
