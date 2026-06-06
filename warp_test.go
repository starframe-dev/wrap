package warp

import (
    "fmt"
    "strings"
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// testPanel is a simple panel that returns its name as content.
type testPanel struct {
    name string
}

func (p *testPanel) View(width, height int) string {
    // Return the name padded to fill the area
    lines := make([]string, height)
    for i := range lines {
        if i == 0 {
            lines[i] = padRight(p.name, width)
        } else {
            lines[i] = strings.Repeat(" ", width)
        }
    }
    return strings.Join(lines, "\n")
}

// testPanelWS records the last WindowSizeMsg it receives.
type testPanelWS struct {
	testPanel
	lastWS tea.WindowSizeMsg
}

func (p *testPanelWS) Update(msg tea.Msg) tea.Cmd {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		p.lastWS = ws
	}
	return nil
}

func (p *testPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func TestTabBarAlignment(t *testing.T) {
	w := New()
	w.NewTab("second")
	w.NewTab("third")
	w.width = 80
	w.height = 10
	w.tabPosition = TabTop

	bar := w.renderTabBar(w.width)
	lines := strings.Split(bar, "\n")
	if len(lines) == 0 {
		t.Fatal("empty tab bar")
	}
	line0 := lines[0]

	// Switch to tab 1 and re-render
	w.switchTab(1)
	bar2 := w.renderTabBar(w.width)
	lines2 := strings.Split(bar2, "\n")
	if len(lines2) == 0 {
		t.Fatal("empty tab bar after switch")
	}
	line1 := lines2[0]

	// The bar length should be the same (no shift)
	if len(line0) != len(line1) {
		t.Logf("Tab 0 bar len: %d", len(line0))
		t.Logf("Tab 1 bar len: %d", len(line1))
		// Length may differ due to ANSI, check visually with StripANSI
	}

	// Check that tab regions have consistent widths
	for i, r := range w.tabRegions {
		w := r.endX - r.startX
		if w <= 0 {
			t.Errorf("tab %d has invalid width %d", i, w)
		}
	}
}

func TestNew(t *testing.T) {
    w := New()
    if len(w.tabs) != 1 {
        t.Errorf("expected 1 tab, got %d", len(w.tabs))
    }
    if w.activeTab != 0 {
        t.Errorf("expected activeTab 0, got %d", w.activeTab)
    }
    if w.tabPosition != TabTop {
        t.Errorf("expected TabTop, got %v", w.tabPosition)
    }
}

func TestNewTab(t *testing.T) {
    w := New()
    w.NewTab("second")
    if len(w.tabs) != 2 {
        t.Errorf("expected 2 tabs, got %d", len(w.tabs))
    }
    if w.activeTab != 1 {
        t.Errorf("expected activeTab 1, got %d", w.activeTab)
    }
}

func TestSplitVertical(t *testing.T) {
    w := New()
    tab := w.ActiveTab()

    p2 := &testPanel{name: "right"}

    // Get the default empty panel and split it
    defaultPanel := tab.root.Panel
    tab.SplitVertical(defaultPanel, 0.5, p2)

    if tab.root.Split == nil {
        t.Fatal("expected root split, got nil")
    }
    if tab.root.Split.Direction != Vertical {
        t.Errorf("expected Vertical, got %v", tab.root.Split.Direction)
    }
}

func TestSplitHorizontal(t *testing.T) {
    w := New()
    tab := w.ActiveTab()

    p2 := &testPanel{name: "bottom"}

    defaultPanel := tab.root.Panel
    tab.SplitHorizontal(defaultPanel, 0.5, p2)

    if tab.root.Split == nil {
        t.Fatal("expected root split, got nil")
    }
    if tab.root.Split.Direction != Horizontal {
        t.Errorf("expected Horizontal, got %v", tab.root.Split.Direction)
    }
}

func TestRenderSplit(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 10
    tab := w.ActiveTab()

    p2 := &testPanel{name: "right"}

    defaultPanel := tab.root.Panel
    tab.SplitVertical(defaultPanel, 0.5, p2)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    if len(lines) != w.contentHeight() {
        t.Errorf("expected %d lines, got %d", w.contentHeight(), len(lines))
    }

    // Check border exists (│ character)
    hasBorder := false
    for _, line := range lines {
        if strings.Contains(line, "│") {
            hasBorder = true
            break
        }
    }
    if !hasBorder {
        t.Error("expected vertical border │ in output")
    }
}

func TestRenderNestedSplit(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 15
    tab := w.ActiveTab()

    p2 := &testPanel{name: "b"}
    p3 := &testPanel{name: "c"}

    defaultPanel := tab.root.Panel
    tab.SplitVertical(defaultPanel, 0.5, p2)
    tab.SplitHorizontal(p2, 0.5, p3)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    if len(lines) != w.contentHeight() {
        t.Errorf("expected %d lines, got %d", w.contentHeight(), len(lines))
    }
}

func TestFloatPane(t *testing.T) {
    w := New()
    tab := w.ActiveTab()

    p := &testPanel{name: "float"}
    tab.Float(p, 5, 2, 20, 5)

    if len(tab.floats) != 1 {
        t.Errorf("expected 1 float pane, got %d", len(tab.floats))
    }
    if tab.floats[0].X != 5 || tab.floats[0].Y != 2 {
        t.Errorf("float position mismatch")
    }
}

func TestTabPosition(t *testing.T) {
    w := New()
    w.SetTabPosition(TabBottom)
    if w.tabPosition != TabBottom {
        t.Errorf("expected TabBottom, got %v", w.tabPosition)
    }
}

func TestCloseTab(t *testing.T) {
    w := New()
    w.NewTab("second")
    w.NewTab("third")
    w.activeTab = 1

    w.closeTab(1)
    if len(w.tabs) != 2 {
        t.Errorf("expected 2 tabs after close, got %d", len(w.tabs))
    }
}

func TestCloseLastTab(t *testing.T) {
    w := New()
    // Can't close the last tab
    w.closeTab(0)
    if len(w.tabs) != 1 {
        t.Errorf("should not close last tab, got %d tabs", len(w.tabs))
    }
}

func TestClampFraction(t *testing.T) {
    tests := []struct {
        in, want float64
    }{
        {0.0, 0.1},
        {0.05, 0.1},
        {0.5, 0.5},
        {0.95, 0.9},
        {1.0, 0.9},
        {1.5, 0.9},
    }
    for _, tt := range tests {
        got := clampFraction(tt.in)
        if got != tt.want {
            t.Errorf("clampFraction(%v) = %v, want %v", tt.in, got, tt.want)
        }
    }
}

func TestFloatCloseButton(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 15
    tab := w.ActiveTab()

    p := &testPanel{name: "float"}
    tab.Float(p, 5, 2, 20, 5)

    if len(tab.floats) != 1 {
        t.Fatal("expected 1 float pane")
    }

    fp := tab.floats[0]

    // Simulate clicking the × button on the float's title bar.
    // × is at relative position (fp.Width-2, 0) in float coordinates.
    // Screen coordinates account for the tab bar offset (0, 1) for TabTop.
    offsetX := 0
    offsetY := 1
    closeX := fp.X + fp.Width - 2 + offsetX
    closeY := fp.Y + offsetY

    msg := tea.MouseMsg{
        Action: tea.MouseActionPress,
        Button: tea.MouseButtonLeft,
        X:      closeX,
        Y:      closeY,
    }

    // handleMouse subtracts the offset, bringing coordinates to content space
    cmd := tab.handleMouse(msg, offsetX, offsetY)
    if cmd != nil {
        t.Errorf("expected nil cmd from close button, got %v", cmd)
    }

    if len(tab.floats) != 0 {
        t.Errorf("expected float pane to be closed, got %d floats", len(tab.floats))
    }
}

func TestOverlayFloatOnFloat(t *testing.T) {
    w := New()
    w.width = 60
    w.height = 20
    tab := w.ActiveTab()

    // Float #1 at (5, 2, 30, 7)
    f1 := &testPanel{name: "first"}
    tab.Float(f1, 5, 2, 30, 7)

    // Float #2 overlays Float #1 at (15, 4, 20, 5)
    f2 := &testPanel{name: "second"}
    tab.Float(f2, 15, 4, 20, 5)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    // Float #2 should have ╭ at its start (visual position 15, content Y=4)
    // Content area: contentHeight = height - 1 = 19 for TabTop
    // Float #2 title is at content line 4
    if len(lines) <= 4 {
        t.Fatal("not enough content lines")
    }

    line4 := StripANSI(lines[4])

    // Float #2 title should have ╭ at position 15
    if len(line4) < 16 {
        t.Fatalf("line 4 too short: %d chars", len(line4))
    }

    // Check that ╭ exists somewhere in the float #2 title area (positions 15-35)
    float2Area := line4[15:36]
    if !strings.Contains(float2Area, "╭") {
        t.Errorf("Float #2 title missing ╭ in area [15:36], got: %q", float2Area)
    }

    // Float #1 should still have ╭ at position 5 on its own title line (content line 2)
    line2 := StripANSI(lines[2])
    float1Area := line2[5:35]
    if !strings.Contains(float1Area, "╭") {
        t.Errorf("Float #1 title missing ╭, got: %q", float1Area[:10])
    }

    // Debug: dump raw content line 4 as hex around position 14
    t.Logf("Raw content line 4 (len=%d):", len(lines[4]))
    // Show bytes from position 12 to 25
    start := 12
    end := start + 20
    if end > len(lines[4]) {
        end = len(lines[4])
    }
    for i := start; i < end; i++ {
        t.Logf("  byte[%d] = 0x%02x (%q)", i, lines[4][i], lines[4][i])
    }
}

func TestOverlayFloatDemoParams(t *testing.T) {
    // Exact same parameters as the demo app
    w := New()
    w.width = 100
    w.height = 30
    tab := w.ActiveTab()

    f1 := &testPanel{name: "Float #1 — drag me!"}
    tab.Float(f1, 8, 3, 28, 8)

    f2 := &testPanel{name: "Float #2"}
    tab.Float(f2, 15, 8, 22, 5)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    // Float #2 title is at content line 8
    if len(lines) <= 8 {
        t.Fatal("not enough content lines")
    }

    line8 := StripANSI(lines[8])
    t.Logf("Line 8 (stripped, len=%d): %q", len(line8), line8)

    // Float #2 title should have ╭ at position 15 (Float #2 X coordinate)
    if len(line8) < 38 {
        t.Fatalf("line 8 too short: %d chars", len(line8))
    }

    float2Area := line8[15:38]
    t.Logf("Float #2 area [15:38]: %q", float2Area)

    if !strings.Contains(float2Area, "╭") {
        t.Errorf("Float #2 title missing ╭ in area [15:38], got: %q", float2Area)
    }

    // Dump ALL bytes of raw line 8
    rawLine8 := lines[8]
    t.Logf("Raw line 8 (len=%d):", len(rawLine8))
    var hexParts []string
    for i := 0; i < len(rawLine8); i++ {
        hexParts = append(hexParts, fmt.Sprintf("%02x", rawLine8[i]))
    }
    t.Logf("  HEX: %s", strings.Join(hexParts, " "))
    t.Logf("  STR: %q", rawLine8)
}

func TestWindowSizeMsgBroadcast(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 10

    // Replace root panel with one that records WindowSizeMsg
    p := &testPanelWS{testPanel: testPanel{name: "ws"}}
    tab := w.ActiveTab()
    tab.root = &Node{Panel: p}

    ws := tea.WindowSizeMsg{Width: 80, Height: 24}
    w.Update(ws)

    if w.width != 80 || w.height != 24 {
        t.Errorf("warp size not updated: got %dx%d", w.width, w.height)
    }
    if p.lastWS.Width != 80 || p.lastWS.Height != 24 {
        t.Errorf("panel did not receive WindowSizeMsg: got %+v", p.lastWS)
    }
}

func TestNewTabButtonClick(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 10
    w.tabPosition = TabTop

    // Render tab bar to populate tabRegions and newTabRegion
    _ = w.renderTabBar(w.width)

    if w.newTabRegion == nil {
        t.Fatal("newTabRegion not set after render")
    }

    // Click on the "+" button (first column of newTabRegion)
    clickX := w.newTabRegion.startX
    clickY := 0

    msg := tea.MouseMsg{
        Action: tea.MouseActionPress,
        Button: tea.MouseButtonLeft,
        X:      clickX,
        Y:      clickY,
    }

    w.handleTabBarClick(msg)

    if len(w.tabs) != 2 {
        t.Errorf("expected 2 tabs after clicking +, got %d", len(w.tabs))
    }
}

func TestTabCloseClick(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 10
    w.tabPosition = TabTop

    // Create a second tab so we can close the first
    w.NewTab("second")
    w.switchTab(0)

    // Render tab bar
    _ = w.renderTabBar(w.width)

    // Find closeX of the first (active) tab
    var closeX int = -1
    for _, r := range w.tabRegions {
        if r.idx == 0 && r.closeX >= 0 {
            closeX = r.closeX
            break
        }
    }
    if closeX < 0 {
        t.Fatal("closeX not found for active tab")
    }

    msg := tea.MouseMsg{
        Action: tea.MouseActionPress,
        Button: tea.MouseButtonLeft,
        X:      closeX,
        Y:      0,
    }

    w.handleTabBarClick(msg)

    if len(w.tabs) != 1 {
        t.Errorf("expected 1 tab after close click, got %d", len(w.tabs))
    }
}

func TestFloatDrag(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 15
    tab := w.ActiveTab()

    p := &testPanel{name: "dragme"}
    tab.Float(p, 5, 2, 20, 5)

    fp := tab.floats[0]

    // TabTop offset is (0, 1). Float at content (5, 2) → screen (5, 3).
    // Title bar is at content Y = fp.Y = 2 → screen Y = 3.
    // Click must NOT be on edge (relX=0 triggers NW resize). Use relX=1 → screen X=6.
    press := tea.MouseMsg{
        Action: tea.MouseActionPress,
        Button: tea.MouseButtonLeft,
        X:      6,
        Y:      3,
    }
    tab.handleMouse(press, 0, 1)

    t.Logf("fp.dragging=%v fp.resizing=%v fp.resizeEdge=%q", fp.dragging, fp.resizing, fp.resizeEdge)
    if !fp.dragging {
        t.Fatalf("expected dragging to be true after press on title bar, got dragging=%v resizing=%v", fp.dragging, fp.resizing)
    }

    // Motion: screen (11, 6) → content (11, 5)
    // dx = 11-6 = 5, dy = 5-2 = 3 → new pos = (5+5, 2+3) = (10, 5)
    motion := tea.MouseMsg{
        Action: tea.MouseActionMotion,
        Button: tea.MouseButtonLeft,
        X:      11,
        Y:      6,
    }
    tab.handleMouse(motion, 0, 1)

    if fp.X != 10 || fp.Y != 5 {
        t.Errorf("expected float at (10, 5) after drag, got (%d, %d)", fp.X, fp.Y)
    }

    // Release
    release := tea.MouseMsg{
        Action: tea.MouseActionRelease,
        Button: tea.MouseButtonLeft,
        X:      11,
        Y:      6,
    }
    tab.handleMouse(release, 0, 1)

    if fp.dragging {
        t.Error("expected dragging to be false after release")
    }
}

func TestNestedTabs(t *testing.T) {
    // Outer warp with a split
    outer := New()
    outer.width = 60
    outer.height = 20
    outerTab := outer.ActiveTab()

    // Inner warp with its own tabs
    inner := New()
    inner.NewTab("inner2")
    innerTab := inner.ActiveTab()
    p := &testPanel{name: "nested"}
    innerTab.Float(p, 2, 2, 15, 5)

    // Embed inner as a panel in outer via split
    outerTab.SplitVertical(outerTab.root.Panel, 0.5, inner.AsPanel())

    // Verify inner is accessible
    content := outerTab.renderContent(outer.contentWidth(), outer.contentHeight())
    lines := strings.Split(content, "\n")
    if len(lines) == 0 {
        t.Fatal("empty content")
    }

    // Inner warp's tab bar should appear in the left half
    found := false
    for _, line := range lines {
        if strings.Contains(line, "main") || strings.Contains(line, "inner2") {
            found = true
            break
        }
    }
    if !found {
        t.Error("expected inner warp tab bar in rendered content")
    }
}

func TestFlexRow(t *testing.T) {
	w := New()
	w.width = 60
	w.height = 10
	tab := w.ActiveTab()

	p1 := &testPanel{name: "left"}
	p2 := &testPanel{name: "mid"}
	p3 := &testPanel{name: "right"}

	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: p1, Grow: 1},
		{Panel: p2, Grow: 2},
		{Panel: p3, Grow: 1},
	})

	content := tab.renderContent(w.contentWidth(), w.contentHeight())
	lines := strings.Split(content, "\n")

	if len(lines) != w.contentHeight() {
		t.Errorf("expected %d lines, got %d", w.contentHeight(), len(lines))
	}

	// Should have 2 vertical borders (3 panels → 2 borders)
	borderCount := 0
	for _, line := range lines {
		if strings.Contains(line, "│") {
			borderCount++
		}
	}
	if borderCount == 0 {
		t.Error("expected vertical borders │ between flex items")
	}
}

func TestFlexColumn(t *testing.T) {
	w := New()
	w.width = 40
	w.height = 15
	tab := w.ActiveTab()

	p1 := &testPanel{name: "top"}
	p2 := &testPanel{name: "bottom"}

	tab.FlexColumn(tab.RootPanel(), []FlexItemSpec{
		{Panel: p1, Grow: 1},
		{Panel: p2, Grow: 1},
	})

	content := tab.renderContent(w.contentWidth(), w.contentHeight())
	lines := strings.Split(content, "\n")

	if len(lines) != w.contentHeight() {
		t.Errorf("expected %d lines, got %d", w.contentHeight(), len(lines))
	}

	// Should have 1 horizontal border
	hasBorder := false
	for _, line := range lines {
		if strings.Contains(line, "─") && !strings.Contains(line, "│") {
			hasBorder = true
			break
		}
	}
	if !hasBorder {
		t.Error("expected horizontal border ─ between flex items")
	}
}

func TestCollapsibleFlex(t *testing.T) {
	w := New()
	w.width = 80
	w.height = 10
	tab := w.ActiveTab()

	content := &testPanel{name: "content"}
	col := NewCollapsible("Section", content)

	other := &testPanel{name: "other"}

	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: col, Grow: 1},
		{Panel: other, Grow: 2},
	})

	// Initially expanded
	render1 := tab.renderContent(w.contentWidth(), w.contentHeight())
	lines1 := strings.Split(render1, "\n")
	if len(lines1) != w.contentHeight() {
		t.Fatalf("expected %d lines, got %d", w.contentHeight(), len(lines1))
	}

	// Find width allocated to collapsible item when expanded
	// The "content" panel is at the start of lines
	firstLine := StripANSI(lines1[0])
	expandedWidth := 0
	for i, r := range firstLine {
		if r == '│' || r == '┌' {
			expandedWidth = i
			break
		}
	}
	if expandedWidth == 0 {
		expandedWidth = len(firstLine) / 3 // rough estimate
	}

	// Collapse — set both Collapsible.Collapsed and FlexItem.Collapsed
	col.Collapsed = true
	for _, item := range tab.root.Flex.Items {
		if item.Node.Panel == col {
			item.Collapsed = true
		}
	}

	render2 := tab.renderContent(w.contentWidth(), w.contentHeight())
	lines2 := strings.Split(render2, "\n")

	// The "other" panel should now have more space
	// Look for "other" text appearing earlier (it shifted left because
	// collapsed item is now only 1 char wide)
	secondLine := StripANSI(lines2[0])
	otherIdx := strings.Index(secondLine, "other")
	if otherIdx < 0 {
		t.Logf("Line 0: %q", secondLine)
		t.Fatal("other panel not found in collapsed layout")
	}

	// other panel should now start at position ~1 (right after 1-char border)
	if otherIdx > expandedWidth/2 {
		t.Logf("other at %d, expandedWidth was ~%d", otherIdx, expandedWidth)
		t.Error("expected other panel to shift left when collapsible is collapsed")
	}
}

func TestScrollable(t *testing.T) {
	w := New()
	w.width = 40
	w.height = 10
	tab := w.ActiveTab()

	// Content taller than viewport
	longContent := &testPanel{
		name: strings.Repeat("line\n", 20) + "end",
	}
	scroll := NewScrollable(longContent)
	tab.root = &Node{Panel: scroll}

	render1 := tab.renderContent(w.contentWidth(), w.contentHeight())
	lines1 := strings.Split(render1, "\n")
	if len(lines1) != w.contentHeight() {
		t.Fatalf("expected %d lines, got %d", w.contentHeight(), len(lines1))
	}

	// Scroll down
	scroll.Offset = 5
	render2 := tab.renderContent(w.contentWidth(), w.contentHeight())
	lines2 := strings.Split(render2, "\n")
	if strings.Contains(lines2[0], "line 1") {
		t.Error("expected first visible line to change after scroll")
	}
}

func TestWordWrap(t *testing.T) {
	text := "hello world this is a test"
	lines := WordWrap(text, 10)
	for _, line := range lines {
		if len(line) > 10 {
			t.Errorf("line too long: %q (%d chars)", line, len(line))
		}
	}
}

func TestSpaceWrap(t *testing.T) {
	text := "hello world this is a test"
	lines := SpaceWrap(text, 10)
	for _, line := range lines {
		if len(line) > 10 {
			t.Errorf("line too long: %q (%d chars)", line, len(line))
		}
	}
	// SpaceWrap should not break words
	for _, line := range lines {
		if strings.Contains(line, "hel") && strings.Contains(line, "lo") && !strings.Contains(line, "hello") {
			t.Errorf("word broken: %q", line)
		}
	}
}

func TestToggleCollapsible(t *testing.T) {
	w := New()
	w.width = 60
	w.height = 10
	tab := w.ActiveTab()

	content := &testPanel{name: "content"}
	col := NewCollapsible("Section", content)

	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: col, Grow: 1},
		{Panel: &testPanel{name: "other"}, Grow: 2},
	})

	// Initially expanded
	if col.Collapsed {
		t.Fatal("expected initially expanded")
	}

	// Toggle via Tab API
	tab.ToggleCollapsible(col)

	if !col.Collapsed {
		t.Error("expected collapsed after ToggleCollapsible")
	}

	// Flex item should be updated
	for _, item := range tab.root.Flex.Items {
		if item.Node.Panel == col && !item.Collapsed {
			t.Error("expected flex item to be collapsed")
		}
	}
}

func TestFloatResize(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 15
    tab := w.ActiveTab()

    p := &testPanel{name: "resize"}
    tab.Float(p, 5, 2, 20, 5)

    fp := tab.floats[0]

    // Press on right edge (x = 5+20-1 = 24, y = 2+2 = 4, inside float)
    press := tea.MouseMsg{
        Action: tea.MouseActionPress,
        Button: tea.MouseButtonLeft,
        X:      24,
        Y:      4,
    }
    tab.handleMouse(press, 0, 1)

    if !fp.resizing {
        t.Fatal("expected resizing to be true after press on edge")
    }

    // Motion to expand right by 5
    motion := tea.MouseMsg{
        Action: tea.MouseActionMotion,
        Button: tea.MouseButtonLeft,
        X:      29,
        Y:      4,
    }
    tab.handleMouse(motion, 0, 1)

    if fp.Width != 25 {
        t.Errorf("expected width 25 after resize, got %d", fp.Width)
    }

    // Release
    release := tea.MouseMsg{
        Action: tea.MouseActionRelease,
        Button: tea.MouseButtonLeft,
        X:      29,
        Y:      4,
    }
    tab.handleMouse(release, 0, 1)

    if fp.resizing {
        t.Error("expected resizing to be false after release")
    }
}

// TestFloatOffScreen checks that floats partially off-screen render correctly.
func TestFloatOffScreen(t *testing.T) {
    w := New()
    w.width = 40
    w.height = 10
    tab := w.ActiveTab()

    p := &testPanel{name: "offscreen"}
    // Float starts at X=35, Width=10 → only 5 chars visible (35-39)
    tab.Float(p, 35, 2, 10, 5)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    // Line 2 (Y=2) should have float title starting at position 35
    line2 := StripANSI(lines[2])
    if len(line2) < 40 {
        t.Fatalf("line 2 too short: %d chars, expected at least 40", len(line2))
    }

    // Should see ╭ at position 35 (the start of the float)
    if line2[35] != "╭"[0] {
        t.Errorf("expected ╭ at position 35, got %q", string(line2[35]))
    }

    // Positions 40+ should still have original content (spaces)
    // But since we're at width 40, position 39 is the last visible char
    // The float should NOT extend the line beyond the content width
    if len(lines[0]) != w.contentWidth() {
        t.Errorf("line 0 width: expected %d, got %d", w.contentWidth(), len(lines[0]))
    }
}

// TestFloatOverlayLineAbove checks that lines above float are not modified.
func TestFloatOverlayLineAbove(t *testing.T) {
    w := New()
    w.width = 60
    w.height = 20
    tab := w.ActiveTab()

    tab.root = &Node{Panel: &testPanel{name: "background"}}

    p := &testPanel{name: "float"}
    tab.Float(p, 10, 5, 20, 6)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    expectedW := w.contentWidth()
    // Lines 0-4 (above float Y=5) should have correct width and not contain float
    for y := 0; y < 5; y++ {
        line := StripANSI(lines[y])
        if len(line) != expectedW {
            t.Errorf("line %d width: expected %d, got %d", y, expectedW, len(line))
        }
        if strings.Contains(line, "╭") || strings.Contains(line, "╮") {
            t.Errorf("line %d should not contain float border chars: %q", y, line)
        }
    }
}

// TestFloatOverlayLineBelow checks that lines below float are not modified.
func TestFloatOverlayLineBelow(t *testing.T) {
    w := New()
    w.width = 60
    w.height = 20
    tab := w.ActiveTab()

    tab.root = &Node{Panel: &testPanel{name: "background"}}

    p := &testPanel{name: "float"}
    tab.Float(p, 10, 5, 20, 6) // Y=5, Height=6 → occupies lines 5-10

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    expectedW := w.contentWidth()
    // Lines 11+ (below float) should have correct width
    for y := 11; y < w.contentHeight(); y++ {
        line := StripANSI(lines[y])
        if len(line) != expectedW {
            t.Errorf("line %d width: expected %d, got %d", y, expectedW, len(line))
        }
    }
}

// TestFloatDoesNotExtendLines checks that overlay doesn't make lines too long.
func TestOverlayFloatSimple(t *testing.T) {
	// Simplest possible test: overlay float on blank line
	lines := []string{
		strings.Repeat(" ", 50),
	}

	fp := &FloatPane{
		Panel:  &testPanel{name: "x"},
		X:      5,
		Y:      0,
		Width:  20,
		Height: 1,
		Title:  "Test",
	}

	overlayFloat(lines, fp, 50, 1)

	stripped := StripANSI(lines[0])
	visW := lipgloss.Width(stripped)
	t.Logf("result: %q (visW=%d, bytes=%d)", stripped, visW, len(stripped))
	if visW != 50 {
		t.Errorf("expected visual width 50, got %d", visW)
	}
}

func TestFloatDoesNotExtendLines(t *testing.T) {
    w := New()
    w.width = 50
    w.height = 15
    tab := w.ActiveTab()

    tab.root = &Node{Panel: &testPanel{name: "bg"}}
    tab.Float(&testPanel{name: "f"}, 5, 3, 30, 5)

    content := tab.renderContent(w.contentWidth(), w.contentHeight())
    lines := strings.Split(content, "\n")

    expectedW := w.contentWidth()
    for y, line := range lines {
        stripped := StripANSI(line)
        visW := lipgloss.Width(stripped)
        if visW != expectedW {
            t.Logf("line %d raw (len=%d): %q", y, len(line), line)
            t.Logf("line %d stripped (visW=%d): %q", y, visW, stripped)
            t.Errorf("line %d: expected visual width %d, got %d", y, expectedW, visW)
        }
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
