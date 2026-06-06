package warp

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// tabRegion describes a clickable area in the tab bar.
type tabRegion struct {
    idx     int
    startX  int
    endX    int
    closeX  int // X position of close button, -1 if none
}

// Warp is the root layout engine. It implements tea.Model.
type Warp struct {
    tabs      []*Tab
    activeTab int
    width     int
    height    int

    tabPosition     TabPosition
    tabRegions      []tabRegion
    newTabRegion    *tabRegion
    verticalTabWidth int
}

// New creates a new Warp instance with one empty tab.
func New() *Warp {
    w := &Warp{
        tabPosition: TabTop,
    }
    w.NewTab("main")
    return w
}

// NewTab creates a new tab and switches to it.
// Returns the created Tab for further configuration.
func (w *Warp) NewTab(name string) *Tab {
    tab := newTab(name, w)
    w.tabs = append(w.tabs, tab)
    w.activeTab = len(w.tabs) - 1
    return tab
}

// SetTabPosition sets where the tab bar is rendered.
func (w *Warp) SetTabPosition(pos TabPosition) {
    w.tabPosition = pos
}

// ActiveTab returns the currently active tab.
func (w *Warp) ActiveTab() *Tab {
    if w.activeTab < 0 || w.activeTab >= len(w.tabs) {
        return nil
    }
    return w.tabs[w.activeTab]
}

func (w *Warp) closeTab(idx int) {
    if idx < 0 || idx >= len(w.tabs) || len(w.tabs) <= 1 {
        return
    }
    w.tabs = append(w.tabs[:idx], w.tabs[idx+1:]...)
    if w.activeTab >= len(w.tabs) {
        w.activeTab = len(w.tabs) - 1
    }
}

func (w *Warp) switchTab(idx int) {
    if idx >= 0 && idx < len(w.tabs) {
        w.activeTab = idx
    }
}

// NextTab switches to the next tab.
func (w *Warp) NextTab() {
    if len(w.tabs) > 1 {
        w.activeTab = (w.activeTab + 1) % len(w.tabs)
    }
}

// PrevTab switches to the previous tab.
func (w *Warp) PrevTab() {
    if len(w.tabs) > 1 {
        w.activeTab = (w.activeTab - 1 + len(w.tabs)) % len(w.tabs)
    }
}

func (w *Warp) contentWidth() int {
    if w.tabPosition == TabLeft || w.tabPosition == TabRight {
        return w.width - w.verticalTabWidth
    }
    return w.width
}

func (w *Warp) contentHeight() int {
    if w.tabPosition == TabTop || w.tabPosition == TabBottom {
        return w.height - 1 // Tab bar is 1 line
    }
    return w.height
}

// Init is the Bubbletea initialization.
func (w *Warp) Init() tea.Cmd {
    return nil
}

// Update handles Bubbletea messages.
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return w.handleKeyMsg(msg)

    case tea.WindowSizeMsg:
        w.width = msg.Width
        w.height = msg.Height
        var cmds []tea.Cmd
        for _, tab := range w.tabs {
            cmds = append(cmds, tab.broadcastMsg(msg)...)
        }
        return w, tea.Batch(cmds...)

    case tea.MouseMsg:
        return w.handleMouseMsg(msg)
    }

    // Forward other messages to active tab's focused panel
    if tab := w.ActiveTab(); tab != nil && tab.focused != nil {
        return w, tab.focused.Update(msg)
    }
    return w, nil
}

func (w *Warp) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    keyStr := msg.String()

    // Global shortcuts
    switch keyStr {
    case "ctrl+c", "q":
        return w, tea.Quit

    case "ctrl+tab":
        w.NextTab()
        return w, nil

    case "ctrl+shift+tab":
        w.PrevTab()
        return w, nil

    case "ctrl+w":
        w.closeTab(w.activeTab)
        return w, nil

    case "ctrl+t":
        w.NewTab("tab")
        return w, nil
    }

    // Forward to active tab
    if tab := w.ActiveTab(); tab != nil {
        return w, tab.handleKeys(msg)
    }
    return w, nil
}

func (w *Warp) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
    // Tab bar hit test
    if w.isOnTabBar(msg.X, msg.Y) {
        return w, w.handleTabBarClick(msg)
    }

    // Forward to active tab with offset
    ox, oy := w.contentOffset()
    if tab := w.ActiveTab(); tab != nil {
        return w, tab.handleMouse(msg, ox, oy)
    }
    return w, nil
}

func (w *Warp) isOnTabBar(x, y int) bool {
    switch w.tabPosition {
    case TabTop:
        return y == 0
    case TabBottom:
        return y == w.height-1
    case TabLeft:
        return x < w.verticalTabWidth
    case TabRight:
        return x >= w.width-w.verticalTabWidth
    }
    return false
}

func (w *Warp) contentOffset() (int, int) {
    switch w.tabPosition {
    case TabTop:
        return 0, 1
    case TabBottom:
        return 0, 0
    case TabLeft:
        return w.verticalTabWidth, 0
    case TabRight:
        return 0, 0
    }
    return 0, 0
}

func (w *Warp) handleTabBarClick(msg tea.MouseMsg) tea.Cmd {
    if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
        return nil
    }

    x := msg.X
    y := msg.Y

    // Adjust for bottom/right tab bars
    switch w.tabPosition {
    case TabLeft, TabRight:
        // For vertical tab bar, y is the row index
        row := y
        if row >= 0 && row < len(w.tabRegions) {
            r := w.tabRegions[row]
            if r.idx == -1 {
                w.NewTab("tab")
                return nil
            }
            if r.closeX >= 0 && x >= r.closeX && x < r.endX {
                w.closeTab(r.idx)
                return nil
            }
            w.switchTab(r.idx)
        }
        return nil
    default:
        // Horizontal tab bar
        for _, r := range w.tabRegions {
            if x >= r.startX && x < r.endX {
                if r.idx == -1 {
                    w.NewTab("tab")
                    return nil
                }
                if r.closeX >= 0 && x >= r.closeX && x < r.endX {
                    w.closeTab(r.idx)
                    return nil
                }
                w.switchTab(r.idx)
                return nil
            }
        }
        // Check the "+" new-tab button
        if w.newTabRegion != nil && x >= w.newTabRegion.startX && x < w.newTabRegion.endX {
            w.NewTab("tab")
            return nil
        }
    }
    return nil
}

// View renders the full TUI.
func (w *Warp) View() string {
    if w.width == 0 || w.height == 0 {
        return "Loading..."
    }

    tab := w.ActiveTab()
    if tab == nil {
        return "No tabs"
    }

    cw := w.contentWidth()
    ch := w.contentHeight()

    switch w.tabPosition {
    case TabTop:
        tabBar := w.renderTabBar(w.width)
        content := tab.renderContent(cw, ch)
        return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)

    case TabBottom:
        content := tab.renderContent(cw, ch)
        tabBar := w.renderTabBar(w.width)
        return lipgloss.JoinVertical(lipgloss.Left, content, tabBar)

    case TabLeft:
        tabBar := w.renderTabBar(w.verticalTabWidth)
        content := tab.renderContent(cw, ch)
        return lipgloss.JoinHorizontal(lipgloss.Top, tabBar, content)

    case TabRight:
        content := tab.renderContent(cw, ch)
        tabBar := w.renderTabBar(w.verticalTabWidth)
        return lipgloss.JoinHorizontal(lipgloss.Top, content, tabBar)
    }

    return ""
}

// AsPanel returns a Panel adapter for this Warp, enabling nested tabs.
// The inner Warp renders within the bounds of the outer panel.
// Note: mouse coordinates are terminal-relative; for nested warps
// use keyboard shortcuts (Ctrl+Tab, Ctrl+Shift+Tab) to switch inner tabs.
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
