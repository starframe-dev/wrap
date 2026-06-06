package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Collapsible is a panel that can be collapsed to a single-line title bar.
type Collapsible struct {
	Title     string
	Collapsed bool
	Content   Panel
}

// NewCollapsible creates a new collapsible panel.
func NewCollapsible(title string, content Panel) *Collapsible {
	return &Collapsible{
		Title:   title,
		Content: content,
	}
}

// View renders the collapsible panel.
// When collapsed, returns a single-line title bar.
func (c *Collapsible) View(w, h int) string {
	if c.Collapsed {
		return c.renderCollapsed(w)
	}
	if c.Content != nil {
		return c.Content.View(w, h)
	}
	return ""
}

// Update forwards messages to the inner content panel.
func (c *Collapsible) Update(msg tea.Msg) tea.Cmd {
	if c.Content != nil {
		return c.Content.Update(msg)
	}
	return nil
}

// Toggle switches between collapsed and expanded states.
func (c *Collapsible) Toggle() {
	c.Collapsed = !c.Collapsed
}

// renderCollapsed renders the title bar for a collapsed panel.
func (c *Collapsible) renderCollapsed(w int) string {
	if w <= 0 {
		return ""
	}

	indicator := "▶"
	if !c.Collapsed {
		indicator = "▼"
	}

	title := c.Title
	reserve := 5 // indicator + spaces + corners
	if len(title) > w-reserve {
		maxLen := w - reserve - 3
		if maxLen < 0 {
			maxLen = 0
		}
		if maxLen == 0 {
			title = ""
		} else {
			title = title[:maxLen] + "..."
		}
	}

	padding := w - len(title) - reserve
	if padding < 0 {
		padding = 0
	}

	return collapsibleStyle.Render("┌"+indicator+" "+title) +
		collapsibleBorderStyle.Render(strings.Repeat("─", padding)+"┐")
}
