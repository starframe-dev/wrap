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

// --- statusPanel ---

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
	status := &statusPanel{msg: "Welcome! TabGroup is just a Panel — you can put it anywhere. Ctrl+T=new tab, Ctrl+W=close, Ctrl+Tab=next"}

	w := warp.New()
	w.SetTabPosition(warp.TabTop)

	// ═══════════════════════════════════════════════════
	// TAB 1: Flex + TabGroup INSIDE flex (tabs are local!)
	// ═══════════════════════════════════════════════════
	tab1 := w.ActiveTab()

	// A TabGroup as a LOCAL component inside the flex layout
	localTabs := warp.NewTabGroup(warp.TabLeft)
	localTabs.NewTab("local-a")
	localTabs.NewTab("local-b")
	// Put some content in local tabs
	localTabs.ActiveTab().SplitVertical(localTabs.ActiveTab().RootPanel(), 0.5, &demoPanel{name: "Local Right"})

	explorer := warp.NewCollapsible("Explorer", &demoPanel{name: "File Tree"})

	longText := "This demonstrates that TabGroup is just a Panel component. " +
		"You can embed it inside splits, flex layouts, or anywhere else. " +
		"It is NOT global — each TabGroup manages its own set of tabs independently. " +
		"The root warp has its own tabs, and this local TabGroup has its own."
	textContent := newTextPanel(longText, 40)
	selectableText := warp.NewSelectable(textContent)
	scroll := warp.NewScrollable(selectableText)

	dropdown := warp.NewDropdownMenu("Actions", []warp.DropdownItem{
		{Label: "New File"},
		{Label: "Open..."},
		{Label: "Save"},
		{Label: "Exit"},
	})
	dropdown.OnSelect = func(idx int) {
		status.Set(fmt.Sprintf("Dropdown: %s", dropdown.Items[idx].Label))
	}

	terminal := warp.NewCollapsible("Terminal", &demoPanel{name: "Shell"})

	// Flex row: local TabGroup | Explorer | Scrollable | Dropdown | Terminal
	tab1.FlexRow(tab1.RootPanel(), []warp.FlexItemSpec{
		{Panel: localTabs, Grow: 2},     // ← TabGroup inside flex!
		{Panel: explorer, Grow: 1},
		{Panel: scroll, Grow: 2},
		{Panel: dropdown, Grow: 1},
		{Panel: terminal, Grow: 1},
	})

	// Overlapping floats
	tab1.Float(&demoPanel{name: "Float #1"}, 10, 4, 24, 6)
	tab1.Float(&demoPanel{name: "Float #2"}, 22, 7, 20, 5)
	tab1.Float(&demoPanel{name: "Float #3"}, 45, 2, 18, 4)

	// ═══════════════════════════════════════════════════
	// TAB 2: TabGroup with TabBottom inside a split
	// ═══════════════════════════════════════════════════
	tab2 := w.NewTab("bottom-tabs")

	bottomTabs := warp.NewTabGroup(warp.TabBottom)
	bottomTabs.NewTab("btm-1")
	bottomTabs.NewTab("btm-2")
	bottomTabs.ActiveTab().SplitHorizontal(bottomTabs.ActiveTab().RootPanel(), 0.5, &demoPanel{name: "Bottom Content"})

	tab2.SplitVertical(tab2.RootPanel(), 0.5, bottomTabs) // ← TabGroup as split child
	tab2.SplitVertical(tab2.RootPanel(), 0.5, &demoPanel{name: "Side"})

	// ═══════════════════════════════════════════════════
	// TAB 3: TabGroup with TabRight inside flex column
	// ═══════════════════════════════════════════════════
	tab3 := w.NewTab("right-tabs")

	rightTabs := warp.NewTabGroup(warp.TabRight)
	rightTabs.NewTab("rt-1")
	rightTabs.NewTab("rt-2")
	rightTabs.NewTab("rt-3")
	rightTabs.ActiveTab().Float(&demoPanel{name: "Right Float"}, 3, 2, 16, 4)

	tab3.FlexColumn(tab3.RootPanel(), []warp.FlexItemSpec{
		{Panel: rightTabs, Grow: 1}, // ← TabGroup in column
		{Panel: &demoPanel{name: "Below"}, Grow: 1},
	})

	// ═══════════════════════════════════════════════════
	// TAB 4: Classic splits (no tabs inside)
	// ═══════════════════════════════════════════════════
	tab4 := w.NewTab("splits")
	topRight := &demoPanel{name: "Top Right"}
	tab4.SplitVertical(tab4.RootPanel(), 0.5, topRight)
	tab4.SplitHorizontal(topRight, 0.5, &demoPanel{name: "Bottom"})

	if err := w.Run(); err != nil {
		panic(err)
	}
}
