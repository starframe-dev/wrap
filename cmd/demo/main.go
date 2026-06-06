package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/wrap"
)

// --- demoPanel — basic clickable panel ---

type demoPanel struct {
	name  string
	count int
}

func (p *demoPanel) View(w, h int) string {
	lines := make([]string, h)
	for i := range lines {
		switch {
		case i == 0:
			lines[i] = padRight(p.name, w)
		case i == h-1:
			lines[i] = padRight(fmt.Sprintf("clicks: %d", p.count), w)
		default:
			lines[i] = strings.Repeat("·", w)
		}
	}
	return strings.Join(lines, "\n")
}

func (p *demoPanel) Update(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(tea.MouseMsg); ok {
		p.count++
	}
	return nil
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

// --- textPanel — scrollable wrapped text ---

type textPanel struct {
	lines []string
}

func newTextPanel(text string, wrapW int) *textPanel {
	wrapped := warp.WordWrap(text, wrapW)
	return &textPanel{lines: wrapped}
}

func (p *textPanel) View(w, h int) string {
	result := make([]string, h)
	for i := 0; i < h; i++ {
		if i < len(p.lines) {
			line := p.lines[i]
			if len(line) > w {
				line = line[:w]
			}
			result[i] = line + strings.Repeat(" ", w-len(line))
		} else {
			result[i] = strings.Repeat(" ", w)
		}
	}
	return strings.Join(result, "\n")
}

func (p *textPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// --- statusPanel — shows current state ---

type statusPanel struct {
	msg string
}

func (p *statusPanel) View(w, h int) string {
	line := padRight(p.msg, w)
	lines := make([]string, h)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func (p *statusPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (p *statusPanel) Set(msg string) {
	p.msg = msg
}

func main() {
	status := &statusPanel{msg: "Welcome to Warp demo! Click things. Use mouse wheel to scroll. Ctrl+T=new tab, Ctrl+W=close tab, Ctrl+Tab=next tab"}

	w := warp.New()
	w.SetTabPosition(warp.TabTop)

	// ═══════════════════════════════════════════════════
	// TAB 1: Masterpiece — all features in one tab
	// ═══════════════════════════════════════════════════
	tab1 := w.ActiveTab()

	// Middle area: 4-column flex with collapsible, scrollable, dropdown, terminal
	explorer := warp.NewCollapsible("Explorer", &demoPanel{name: "File Tree"})

	longText := "This is a demonstration of word wrapping in Warp. " +
		"The text should flow naturally and wrap at word boundaries without breaking mid-word. " +
		"You can scroll this panel with your mouse wheel or arrow keys. " +
		"Warp supports tabs, splits, floating panels, flex layouts, collapsible sections, " +
		"dropdown menus, context menus, and mouse-driven interactions. " +
		"Everything is rendered with beautiful Gruvbox Dark colors. " +
		"Try dragging the borders between panels to resize them. " +
		"Try clicking the float panels to drag them around. " +
		"Try collapsing the Explorer and Terminal panels. " +
		"The library is built on Bubbletea and Lipgloss for a modern TUI experience. " +
		"All interactions are fully mouse and keyboard driven. " +
		"Have fun exploring! This text is intentionally long to demonstrate scrolling."
	textContent := newTextPanel(longText, 40)
	selectableText := warp.NewSelectable(textContent)
	scroll := warp.NewScrollable(selectableText)

	dropdown := warp.NewDropdownMenu("Actions", []warp.DropdownItem{
		{Label: "New File"},
		{Label: "Open..."},
		{Label: "Save"},
		{Label: "Save As..."},
		{Label: "Exit"},
	})
	dropdown.OnSelect = func(idx int) {
		status.Set(fmt.Sprintf("Dropdown selected: %s", dropdown.Items[idx].Label))
	}

	terminal := warp.NewCollapsible("Terminal", &demoPanel{name: "Shell"})

	// Build flex row for the 3 middle columns
	midRow := warp.NewCollapsible("", nil)
	tab1.FlexRow(tab1.RootPanel(), []warp.FlexItemSpec{
		{Panel: midRow, Grow: 1},
	})
	// Replace the dummy with actual flex row
	tab1.FlexRow(midRow, []warp.FlexItemSpec{
		{Panel: explorer, Grow: 1},
		{Panel: scroll, Grow: 2},
		{Panel: dropdown, Grow: 1},
		{Panel: terminal, Grow: 1},
	})

	// Float panel: draggable, resizable, closable
	float1 := &demoPanel{name: "Float #1 — drag title bar!"}
	tab1.Float(float1, 8, 5, 26, 6)

	// ═══════════════════════════════════════════════════
	// TAB 2: Nested tabs demo
	// ═══════════════════════════════════════════════════
	tab2 := w.NewTab("nested")
	inner := warp.New()
	inner.NewTab("inner-code")
	inner.NewTab("inner-debug")
	inner.ActiveTab().Float(&demoPanel{name: "Nested Float"}, 3, 2, 18, 5)

	tab2.SplitVertical(tab2.RootPanel(), 0.4, inner.AsPanel())
	tab2.SplitVertical(tab2.RootPanel(), 0.5, &demoPanel{name: "Side Panel"})

	// ═══════════════════════════════════════════════════
	// TAB 3: Column layout with collapsible
	// ═══════════════════════════════════════════════════
	tab3 := w.NewTab("columns")
	output := warp.NewCollapsible("Build Output", &demoPanel{name: "Compiler output here"})
	bottom := &demoPanel{name: "Bottom Panel — click me!"}
	tab3.FlexColumn(tab3.RootPanel(), []warp.FlexItemSpec{
		{Panel: output, Grow: 1},
		{Panel: bottom, Grow: 1},
	})

	// ═══════════════════════════════════════════════════
	// TAB 4: Split panes demo
	// ═══════════════════════════════════════════════════
	tab4 := w.NewTab("splits")
	topRight := &demoPanel{name: "Top Right"}
	bottomPane := &demoPanel{name: "Bottom Pane"}

	tab4.SplitVertical(tab4.RootPanel(), 0.5, topRight)
	tab4.SplitHorizontal(topRight, 0.5, bottomPane)

	if err := w.Run(); err != nil {
		panic(err)
	}
}
