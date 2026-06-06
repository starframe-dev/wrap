package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// DropdownItem is a single item in a dropdown menu.
type DropdownItem struct {
	Label    string
	Selected bool
}

// DropdownMenu is a panel that shows a button and an expandable list of items.
type DropdownMenu struct {
	Label    string
	Items    []DropdownItem
	Open     bool
	Hovered  int // index of hovered item, -1 if none
	OnSelect func(idx int)
}

// NewDropdownMenu creates a dropdown menu.
func NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu {
	return &DropdownMenu{
		Label:   label,
		Items:   items,
		Hovered: -1,
	}
}

// View renders the dropdown button or the open menu.
func (d *DropdownMenu) View(w, h int) string {
	if !d.Open {
		return d.renderButton(w)
	}
	return d.renderMenu(w, h)
}

func (d *DropdownMenu) renderButton(w int) string {
	label := d.Label + " ▼"
	if len(label) > w {
		label = label[:w-1] + "…"
	}
	return dropdownButtonStyle.Render(padRight(label, w))
}

func (d *DropdownMenu) renderMenu(w, h int) string {
	menuH := len(d.Items) + 1 // button + items
	if menuH > h {
		menuH = h
	}

	lines := make([]string, menuH)
	lines[0] = dropdownButtonStyle.Render(padRight(d.Label+" ▲", w))

	for i := 0; i < len(d.Items) && i+1 < menuH; i++ {
		item := d.Items[i]
		prefix := "  "
		if item.Selected {
			prefix = "✓ "
		}
		label := prefix + item.Label
		if len(label) > w {
			label = label[:w-1] + "…"
		}
		style := dropdownItemStyle
		if i == d.Hovered {
			style = dropdownItemHoverStyle
		}
		if item.Selected {
			style = dropdownItemSelectedStyle
		}
		lines[i+1] = style.Render(padRight(label, w))
	}
	return strings.Join(lines, "\n")
}

// Update handles mouse and keyboard for the dropdown.
func (d *DropdownMenu) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return nil
		}
		if !d.Open {
			// Click on button opens menu
			if msg.Y == 0 {
				d.Open = true
				d.Hovered = -1
			}
			return nil
		}
		// Click inside menu
		if msg.Y == 0 {
			// Click on button closes menu
			d.Open = false
			return nil
		}
		idx := msg.Y - 1
		if idx >= 0 && idx < len(d.Items) {
			d.selectItem(idx)
		}
	case tea.KeyMsg:
		if !d.Open {
			return nil
		}
		switch msg.String() {
		case "up":
			if d.Hovered > 0 {
				d.Hovered--
			}
		case "down":
			if d.Hovered < len(d.Items)-1 {
				d.Hovered++
			}
		case "enter":
			if d.Hovered >= 0 {
				d.selectItem(d.Hovered)
			}
		case "esc":
			d.Open = false
		}
	}
	return nil
}

func (d *DropdownMenu) selectItem(idx int) {
	for i := range d.Items {
		d.Items[i].Selected = false
	}
	d.Items[idx].Selected = true
	d.Open = false
	if d.OnSelect != nil {
		d.OnSelect(idx)
	}
}

// Close closes the dropdown menu.
func (d *DropdownMenu) Close() {
	d.Open = false
}
