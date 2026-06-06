package warp

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func padRight(s string, w int) string {
	lw := lipgloss.Width(s)
	if lw < w {
		return s + strings.Repeat(" ", w-lw)
	}
	return s
}

// tabRegion describes a clickable area in the tab bar.
type tabRegion struct {
	idx     int
	startX  int
	endX    int
	closeX  int // X position of close button, -1 if none
}

// TabGroup is a Panel that renders a tab bar and switches between tabs.
// Use it as a component inside splits, flex layouts, or as the root panel.
type TabGroup struct {
	tabs      []*Tab
	activeTab int
	width     int
	height    int

	tabPosition      TabPosition
	tabRegions       []tabRegion
	newTabRegion     *tabRegion
	verticalTabWidth int
}

// NewTabGroup creates a TabGroup panel with one default tab.
func NewTabGroup(pos TabPosition) *TabGroup {
	tg := &TabGroup{tabPosition: pos}
	tg.NewTab("main")
	return tg
}

// NewTab creates a new tab and switches to it.
func (tg *TabGroup) NewTab(name string) *Tab {
	tab := newTab(name, tg)
	tg.tabs = append(tg.tabs, tab)
	tg.activeTab = len(tg.tabs) - 1
	return tab
}

// ActiveTab returns the currently active tab.
func (tg *TabGroup) ActiveTab() *Tab {
	if tg.activeTab < 0 || tg.activeTab >= len(tg.tabs) {
		return nil
	}
	return tg.tabs[tg.activeTab]
}

func (tg *TabGroup) closeTab(idx int) {
	if idx < 0 || idx >= len(tg.tabs) || len(tg.tabs) <= 1 {
		return
	}
	tg.tabs = append(tg.tabs[:idx], tg.tabs[idx+1:]...)
	if tg.activeTab >= len(tg.tabs) {
		tg.activeTab = len(tg.tabs) - 1
	}
}

func (tg *TabGroup) switchTab(idx int) {
	if idx >= 0 && idx < len(tg.tabs) {
		tg.activeTab = idx
	}
}

// NextTab switches to the next tab.
func (tg *TabGroup) NextTab() {
	if len(tg.tabs) > 1 {
		tg.activeTab = (tg.activeTab + 1) % len(tg.tabs)
	}
}

// PrevTab switches to the previous tab.
func (tg *TabGroup) PrevTab() {
	if len(tg.tabs) > 1 {
		tg.activeTab = (tg.activeTab - 1 + len(tg.tabs)) % len(tg.tabs)
	}
}

func (tg *TabGroup) contentWidth(totalW int) int {
	if tg.tabPosition == TabLeft || tg.tabPosition == TabRight {
		return totalW - tg.verticalTabWidth
	}
	return totalW
}

func (tg *TabGroup) contentHeight(totalH int) int {
	if tg.tabPosition == TabTop || tg.tabPosition == TabBottom {
		return totalH - 1
	}
	return totalH
}

func (tg *TabGroup) contentOffset() (int, int) {
	switch tg.tabPosition {
	case TabTop:
		return 0, 1
	case TabBottom:
		return 0, 0
	case TabLeft:
		return tg.verticalTabWidth, 0
	case TabRight:
		return 0, 0
	}
	return 0, 0
}

// View renders the tab bar + active tab content.
func (tg *TabGroup) View(w, h int) string {
	tg.width = w
	tg.height = h

	tab := tg.ActiveTab()
	if tab == nil {
		return strings.Repeat("\n", h)
	}

	cw := tg.contentWidth(w)
	ch := tg.contentHeight(h)

	switch tg.tabPosition {
	case TabTop:
		tabBar := tg.renderTabBar(w)
		content := tab.renderContent(cw, ch)
		return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)

	case TabBottom:
		content := tab.renderContent(cw, ch)
		tabBar := tg.renderTabBar(w)
		return lipgloss.JoinVertical(lipgloss.Left, content, tabBar)

	case TabLeft:
		tabBar := tg.renderTabBar(tg.verticalTabWidth)
		content := tab.renderContent(cw, ch)
		return lipgloss.JoinHorizontal(lipgloss.Top, tabBar, content)

	case TabRight:
		content := tab.renderContent(cw, ch)
		tabBar := tg.renderTabBar(tg.verticalTabWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, content, tabBar)
	}

	return ""
}

// Update handles keys, mouse, and window resize for the tab group.
func (tg *TabGroup) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return tg.handleKeyMsg(msg)
	case tea.MouseMsg:
		return tg.handleMouseMsg(msg)
	case tea.WindowSizeMsg:
		tg.width = msg.Width
		tg.height = msg.Height
		var cmds []tea.Cmd
		for _, tab := range tg.tabs {
			cmds = append(cmds, tab.broadcastMsg(msg)...)
		}
		return tea.Batch(cmds...)
	}
	// Forward to active tab's focused panel
	if tab := tg.ActiveTab(); tab != nil && tab.focused != nil {
		return tab.focused.Update(msg)
	}
	return nil
}

func (tg *TabGroup) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "q":
		return tea.Quit
	case "ctrl+tab":
		tg.NextTab()
		return nil
	case "ctrl+shift+tab":
		tg.PrevTab()
		return nil
	case "ctrl+w":
		tg.closeTab(tg.activeTab)
		return nil
	case "ctrl+t":
		tg.NewTab("tab")
		return nil
	}

	if tab := tg.ActiveTab(); tab != nil {
		return tab.handleKeys(msg)
	}
	return nil
}

func (tg *TabGroup) handleMouseMsg(msg tea.MouseMsg) tea.Cmd {
	if tg.isOnTabBar(msg.X, msg.Y) {
		return tg.handleTabBarClick(msg)
	}

	ox, oy := tg.contentOffset()
	cw := tg.contentWidth(tg.width)
	ch := tg.contentHeight(tg.height)
	if tab := tg.ActiveTab(); tab != nil {
		return tab.handleMouse(msg, ox, oy, cw, ch)
	}
	return nil
}

func (tg *TabGroup) isOnTabBar(x, y int) bool {
	switch tg.tabPosition {
	case TabTop:
		return y == 0
	case TabBottom:
		return y == tg.height-1
	case TabLeft:
		return x < tg.verticalTabWidth
	case TabRight:
		return x >= tg.width-tg.verticalTabWidth
	}
	return false
}

func (tg *TabGroup) handleTabBarClick(msg tea.MouseMsg) tea.Cmd {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return nil
	}

	x := msg.X
	y := msg.Y

	switch tg.tabPosition {
	case TabLeft, TabRight:
		row := y
		if row >= 0 && row < len(tg.tabRegions) {
			r := tg.tabRegions[row]
			if r.idx == -1 {
				tg.NewTab("tab")
				return nil
			}
			if r.closeX >= 0 && x >= r.closeX && x < r.endX {
				tg.closeTab(r.idx)
				return nil
			}
			tg.switchTab(r.idx)
		}
		return nil
	default:
		for _, r := range tg.tabRegions {
			if x >= r.startX && x < r.endX {
				if r.idx == -1 {
					tg.NewTab("tab")
					return nil
				}
				if r.closeX >= 0 && x >= r.closeX && x < r.endX {
					tg.closeTab(r.idx)
					return nil
				}
				tg.switchTab(r.idx)
				return nil
			}
		}
		if tg.newTabRegion != nil && x >= tg.newTabRegion.startX && x < tg.newTabRegion.endX {
			tg.NewTab("tab")
			return nil
		}
	}
	return nil
}

func (tg *TabGroup) renderTabBar(width int) string {
	if tg.tabPosition == TabLeft || tg.tabPosition == TabRight {
		return tg.renderVerticalTabBar(width)
	}
	return tg.renderHorizontalTabBar(width)
}

func (tg *TabGroup) renderHorizontalTabBar(width int) string {
	tg.tabRegions = nil
	activeIdx := tg.activeTab

	var parts []string
	col := 0
	for i, tab := range tg.tabs {
		name := tab.name
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		label := fmt.Sprintf(" %s ", name)
		if i == activeIdx {
			label = fmt.Sprintf("▎ %s ×", name)
		}

		labelW := lipgloss.Width(label)
		endX := col + labelW
		closeX := -1
		if i == activeIdx {
			closeX = endX - 1
		}
		tg.tabRegions = append(tg.tabRegions, tabRegion{
			idx: i, startX: col, endX: endX, closeX: closeX,
		})
		col += labelW

		style := inactiveTabStyle
		if i == activeIdx {
			style = activeTabStyle
		}
		parts = append(parts, style.Render(label))
	}

	newLabel := " + "
	newW := utf8.RuneCountInString(newLabel)
	tg.newTabRegion = &tabRegion{startX: col, endX: col + newW}
	col += newW
	parts = append(parts, newTabStyle.Render(newLabel))

	bar := tabBarStyle.Render(strings.Join(parts, ""))
	if padding := width - col; padding > 0 {
		bar += tabBarStyle.Render(strings.Repeat(" ", padding))
	}
	return bar
}

func (tg *TabGroup) renderVerticalTabBar(_ int) string {
	tabs := tg.tabs
	activeIdx := tg.activeTab
	tg.tabRegions = nil
	tg.newTabRegion = nil

	var lines []string
	maxW := 0
	for i, tab := range tabs {
		name := tab.name
		if len(name) > 15 {
			name = name[:12] + "..."
		}
		label := fmt.Sprintf(" %s ", name)
		if i == activeIdx {
			label = fmt.Sprintf("▎ %s ×", name)
		}

		labelW := lipgloss.Width(label)
		if labelW > maxW {
			maxW = labelW
		}

		tg.tabRegions = append(tg.tabRegions, tabRegion{
			idx: i, startX: 0, endX: labelW, closeX: labelW - 1,
		})

		style := inactiveTabStyle
		if i == activeIdx {
			style = activeTabStyle
		}
		lines = append(lines, style.Render(padRight(label, maxW)))
	}

	newLabel := " + "
	tg.tabRegions = append(tg.tabRegions, tabRegion{
		idx: -1, startX: 0, endX: lipgloss.Width(newLabel),
	})
	lines = append(lines, newTabStyle.Render(padRight(newLabel, maxW)))

	tg.verticalTabWidth = maxW
	return strings.Join(lines, "\n")
}
