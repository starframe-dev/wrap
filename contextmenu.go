package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ContextMenuItem is a single item in a context menu.
type ContextMenuItem struct {
	Label    string
	Shortcut string // optional shortcut display
	Action   func()
}

// ContextMenu is a popup menu rendered as a floating panel.
type ContextMenu struct {
	Items   []ContextMenuItem
	Hovered int
	Width   int
}

// NewContextMenu creates a new context menu.
func NewContextMenu(items []ContextMenuItem) *ContextMenu {
	cm := &ContextMenu{
		Items:   items,
		Hovered: -1,
	}
	// Auto-compute width based on longest label
	maxW := 10
	for _, item := range items {
		w := lipgloss.Width(item.Label)
		if item.Shortcut != "" {
			w += 2 + lipgloss.Width(item.Shortcut)
		}
		if w > maxW {
			maxW = w
		}
	}
	cm.Width = maxW + 4 // padding
	return cm
}

// View renders the context menu.
func (cm *ContextMenu) View(w, h int) string {
	menuH := len(cm.Items) + 2 // top/bottom border
	if menuH > h {
		menuH = h
	}

	lines := make([]string, menuH)
	lines[0] = contextMenuBorderStyle.Render("┌" + strings.Repeat("─", cm.Width-2) + "┐")

	for i := 0; i < len(cm.Items) && i+1 < menuH-1; i++ {
		item := cm.Items[i]
		label := " " + item.Label
		if item.Shortcut != "" {
			gap := cm.Width - 2 - lipgloss.Width(label) - lipgloss.Width(item.Shortcut)
			if gap < 1 {
				gap = 1
			}
			label += strings.Repeat(" ", gap) + item.Shortcut
		}
		label = padRight(label, cm.Width-2)

		style := contextMenuItemStyle
		if i == cm.Hovered {
			style = contextMenuItemHoverStyle
		}
		lines[i+1] = style.Render("│") + style.Render(label) + style.Render("│")
	}

	lines[menuH-1] = contextMenuBorderStyle.Render("└" + strings.Repeat("─", cm.Width-2) + "┘")
	return strings.Join(lines, "\n")
}

// Update handles mouse and keyboard for the context menu.
func (cm *ContextMenu) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionMotion {
			// Hover tracking: Y is relative to menu top
			idx := msg.Y - 1
			if idx >= 0 && idx < len(cm.Items) {
				cm.Hovered = idx
			} else {
				cm.Hovered = -1
			}
			return nil
		}
		if msg.Action == tea.MouseActionPress {
			idx := msg.Y - 1
			if idx >= 0 && idx < len(cm.Items) {
				cm.activate(idx)
			}
			return nil
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if cm.Hovered > 0 {
				cm.Hovered--
			} else {
				cm.Hovered = len(cm.Items) - 1
			}
		case "down":
			if cm.Hovered < len(cm.Items)-1 {
				cm.Hovered++
			} else {
				cm.Hovered = 0
			}
		case "enter":
			if cm.Hovered >= 0 {
				cm.activate(cm.Hovered)
			}
		case "esc":
			// Close menu — handled by parent float
		}
	}
	return nil
}

func (cm *ContextMenu) activate(idx int) {
	if idx >= 0 && idx < len(cm.Items) && cm.Items[idx].Action != nil {
		cm.Items[idx].Action()
	}
}

// ShowContextMenu is a helper that creates a context menu float at the given position.
func (t *Tab) ShowContextMenu(items []ContextMenuItem, x, y int) *FloatPane {
	cm := NewContextMenu(items)
	fp := &FloatPane{
		Panel:  cm,
		X:      x,
		Y:      y,
		Width:  cm.Width,
		Height: len(items) + 2,
		Title:  "",
	}
	t.floats = append(t.floats, fp)
	return fp
}
