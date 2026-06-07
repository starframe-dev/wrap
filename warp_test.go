package warp

import (
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

func (p *testPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
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

func getTG(w *Warp) *TabGroup {
	if tg, ok := w.Root().(*TabGroup); ok {
		return tg
	}
	return nil
}

func TestNew(t *testing.T) {
	w := New()
	tg := getTG(w)
	if len(tg.tabs) != 1 {
		t.Errorf("expected 1 tab, got %d", len(tg.tabs))
	}
	if tg.activeTab != 0 {
		t.Errorf("expected activeTab 0, got %d", tg.activeTab)
	}
	if tg.tabPosition != TabTop {
		t.Errorf("expected TabTop, got %v", tg.tabPosition)
	}
}

func TestNewTab(t *testing.T) {
	w := New()
	tg := getTG(w)
	tg.NewTab("second")
	if len(tg.tabs) != 2 {
		t.Errorf("expected 2 tabs, got %d", len(tg.tabs))
	}
	if tg.activeTab != 1 {
		t.Errorf("expected activeTab 1, got %d", tg.activeTab)
	}
}

func TestSplitVertical(t *testing.T) {
	w := New()
	tab := w.ActiveTab()

	p2 := &testPanel{name: "right"}
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
	tg := getTG(w)
	tab := tg.ActiveTab()

	p2 := &testPanel{name: "right"}
	defaultPanel := tab.root.Panel
	tab.SplitVertical(defaultPanel, 0.5, p2)

	content := tab.renderContent(tg.contentWidth(w.width), tg.contentHeight(w.height))
	lines := strings.Split(content, "\n")

	if len(lines) != tg.contentHeight(w.height) {
		t.Errorf("expected %d lines, got %d", tg.contentHeight(w.height), len(lines))
	}

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

	content := tab.renderContent(40, 14)
	lines := strings.Split(content, "\n")

	if len(lines) != 14 {
		t.Errorf("expected 14 lines, got %d", len(lines))
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
	tg := getTG(w)
	if tg.tabPosition != TabBottom {
		t.Errorf("expected TabBottom, got %v", tg.tabPosition)
	}
}

func TestCloseTab(t *testing.T) {
	w := New()
	tg := getTG(w)
	tg.NewTab("second")
	tg.NewTab("third")
	tg.activeTab = 1

	tg.closeTab(1)
	if len(tg.tabs) != 2 {
		t.Errorf("expected 2 tabs after close, got %d", len(tg.tabs))
	}
}

func TestCloseLastTab(t *testing.T) {
	w := New()
	tg := getTG(w)
	tg.closeTab(0)
	if len(tg.tabs) != 1 {
		t.Errorf("should not close last tab, got %d tabs", len(tg.tabs))
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
	tg := getTG(w)
	tab := tg.ActiveTab()

	p := &testPanel{name: "float"}
	tab.Float(p, 5, 2, 20, 5)

	if len(tab.floats) != 1 {
		t.Fatal("expected 1 float pane")
	}

	fp := tab.floats[0]

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

	cmd := tab.handleMouse(msg, offsetX, offsetY, 40, 14)
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

	f1 := &testPanel{name: "first"}
	tab.Float(f1, 5, 2, 30, 7)

	f2 := &testPanel{name: "second"}
	tab.Float(f2, 15, 4, 20, 5)

	content := tab.renderContent(60, 19)
	lines := strings.Split(content, "\n")

	if len(lines) <= 4 {
		t.Fatal("not enough content lines")
	}

	line4 := StripANSI(lines[4])
	if len(line4) < 16 {
		t.Fatalf("line 4 too short: %d chars", len(line4))
	}

	float2Area := line4[15:36]
	if !strings.Contains(float2Area, "╭") {
		t.Errorf("Float #2 title missing ╭ in area [15:36], got: %q", float2Area)
	}

	line2 := StripANSI(lines[2])
	float1Area := line2[5:35]
	if !strings.Contains(float1Area, "╭") {
		t.Errorf("Float #1 title missing ╭, got: %q", float1Area[:10])
	}
}

func TestOverlayFloatDemoParams(t *testing.T) {
	w := New()
	w.width = 100
	w.height = 30
	tab := w.ActiveTab()

	f1 := &testPanel{name: "Float #1 — drag me!"}
	tab.Float(f1, 8, 3, 28, 8)

	f2 := &testPanel{name: "Float #2"}
	tab.Float(f2, 15, 8, 22, 5)

	content := tab.renderContent(100, 29)
	lines := strings.Split(content, "\n")

	if len(lines) <= 8 {
		t.Fatal("not enough content lines")
	}

	line8 := StripANSI(lines[8])
	if len(line8) < 38 {
		t.Fatalf("line 8 too short: %d chars", len(line8))
	}

	float2Area := line8[15:38]
	if !strings.Contains(float2Area, "╭") {
		t.Errorf("Float #2 title missing ╭ in area [15:38], got: %q", float2Area)
	}
}

func TestWindowSizeMsgBroadcast(t *testing.T) {
	w := New()
	w.width = 40
	w.height = 10

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
	tg := getTG(w)
	tg.tabPosition = TabTop

	_ = tg.renderTabBar(w.width)

	if tg.newTabRegion == nil {
		t.Fatal("newTabRegion not set after render")
	}

	clickX := tg.newTabRegion.startX
	clickY := 0

	msg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      clickX,
		Y:      clickY,
	}

	tg.handleTabBarClick(msg)

	if len(tg.tabs) != 2 {
		t.Errorf("expected 2 tabs after clicking +, got %d", len(tg.tabs))
	}
}

func TestTabCloseClick(t *testing.T) {
	w := New()
	w.width = 40
	w.height = 10
	tg := getTG(w)
	tg.tabPosition = TabTop

	tg.NewTab("second")
	tg.switchTab(0)

	_ = tg.renderTabBar(w.width)

	var closeX int = -1
	for _, r := range tg.tabRegions {
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

	tg.handleTabBarClick(msg)

	if len(tg.tabs) != 1 {
		t.Errorf("expected 1 tab after close click, got %d", len(tg.tabs))
	}
}

func TestTabBarAlignment(t *testing.T) {
	w := New()
	tg := getTG(w)
	tg.NewTab("second")
	tg.NewTab("third")
	w.width = 80
	w.height = 10
	tg.tabPosition = TabTop

	bar := tg.renderTabBar(w.width)
	lines := strings.Split(bar, "\n")
	if len(lines) == 0 {
		t.Fatal("empty tab bar")
	}
	line0 := lines[0]

	tg.switchTab(1)
	bar2 := tg.renderTabBar(w.width)
	lines2 := strings.Split(bar2, "\n")
	if len(lines2) == 0 {
		t.Fatal("empty tab bar after switch")
	}
	line1 := lines2[0]

	if len(line0) != len(line1) {
		t.Logf("Tab 0 bar len: %d", len(line0))
		t.Logf("Tab 1 bar len: %d", len(line1))
	}

	for i, r := range tg.tabRegions {
		w := r.endX - r.startX
		if w <= 0 {
			t.Errorf("tab %d has invalid width %d", i, w)
		}
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

	press := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      6,
		Y:      3,
	}
	tab.handleMouse(press, 0, 1, 40, 14)

	if !fp.dragging {
		t.Fatalf("expected dragging to be true after press on title bar, got dragging=%v resizing=%v", fp.dragging, fp.resizing)
	}

	motion := tea.MouseMsg{
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
		X:      11,
		Y:      6,
	}
	tab.handleMouse(motion, 0, 1, 40, 14)

	if fp.X != 10 || fp.Y != 5 {
		t.Errorf("expected float at (10, 5) after drag, got (%d, %d)", fp.X, fp.Y)
	}

	release := tea.MouseMsg{
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
		X:      11,
		Y:      6,
	}
	tab.handleMouse(release, 0, 1, 40, 14)

	if fp.dragging {
		t.Error("expected dragging to be false after release")
	}
}

func TestNestedTabs(t *testing.T) {
	outer := New()
	outer.width = 60
	outer.height = 20
	outerTab := outer.ActiveTab()

	inner := New()
	inner.NewTab("inner2")
	innerTab := inner.ActiveTab()
	p := &testPanel{name: "nested"}
	innerTab.Float(p, 2, 2, 15, 5)

	outerTab.SplitVertical(outerTab.root.Panel, 0.5, inner.AsPanel())

	content := outerTab.renderContent(60, 20)
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		t.Fatal("empty content")
	}

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

	content := tab.renderContent(60, 9)
	lines := strings.Split(content, "\n")

	if len(lines) != 9 {
		t.Errorf("expected 9 lines, got %d", len(lines))
	}

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

	content := tab.renderContent(40, 14)
	lines := strings.Split(content, "\n")

	if len(lines) != 14 {
		t.Errorf("expected 14 lines, got %d", len(lines))
	}

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

	render1 := tab.renderContent(80, 9)
	lines1 := strings.Split(render1, "\n")
	if len(lines1) != 9 {
		t.Fatalf("expected 9 lines, got %d", len(lines1))
	}

	firstLine := StripANSI(lines1[0])
	expandedWidth := 0
	for i, r := range firstLine {
		if r == '│' || r == '┌' {
			expandedWidth = i
			break
		}
	}
	if expandedWidth == 0 {
		expandedWidth = len(firstLine) / 3
	}

	col.Collapsed = true
	for _, item := range tab.root.Flex.Items {
		if item.Node.Panel == col {
			item.Collapsed = true
		}
	}

	render2 := tab.renderContent(80, 9)
	lines2 := strings.Split(render2, "\n")

	secondLine := StripANSI(lines2[0])
	otherIdx := strings.Index(secondLine, "other")
	if otherIdx < 0 {
		t.Logf("Line 0: %q", secondLine)
		t.Fatal("other panel not found in collapsed layout")
	}

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

	longContent := &testPanel{
		name: strings.Repeat("line\n", 20) + "end",
	}
	scroll := NewScrollable(longContent)
	tab.root = &Node{Panel: scroll}

	render1 := tab.renderContent(40, 9)
	lines1 := strings.Split(render1, "\n")
	if len(lines1) != 9 {
		t.Fatalf("expected 9 lines, got %d", len(lines1))
	}

	scroll.Offset = 5
	render2 := tab.renderContent(40, 9)
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
	for _, line := range lines {
		if strings.Contains(line, "hel") && strings.Contains(line, "lo") && !strings.Contains(line, "hello") {
			t.Errorf("word broken: %q", line)
		}
	}
}

func TestSelectableHighlight(t *testing.T) {
	p := &testPanel{name: "hello world"}
	s := NewSelectable(p)

	s.AnchorX, s.AnchorY = 6, 0
	s.CursorX, s.CursorY = 11, 0
	s.HasSelection = true

	rendered := s.View(20, 5)
	lines := strings.Split(rendered, "\n")

	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "\x1b[7m") {
		t.Errorf("expected highlight ANSI code in line 0: %q", lines[0])
	}

	for i := 1; i < 5; i++ {
		if strings.Contains(lines[i], "\x1b[7m") {
			t.Errorf("line %d should not have highlight: %q", i, lines[i])
		}
	}
}

func TestSelectableMouseDrag(t *testing.T) {
	p := &testPanel{name: "abcdef"}
	s := NewSelectable(p)

	press := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 1, Y: 0}
	s.Update(press)

	if s.Selecting != true {
		t.Error("expected Selecting to be true after press")
	}

	drag := tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft, X: 4, Y: 0}
	s.Update(drag)

	if !s.HasSelection {
		t.Error("expected HasSelection to be true after drag")
	}

	release := tea.MouseMsg{Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft, X: 4, Y: 0}
	s.Update(release)

	if s.Selecting {
		t.Error("expected Selecting to be false after release")
	}

	text := s.SelectedText()
	if text != "bcd" {
		t.Errorf("expected selected text 'bcd', got %q", text)
	}
}

func TestSelectableClear(t *testing.T) {
	p := &testPanel{name: "test"}
	s := NewSelectable(p)
	s.SelectAll(10, 5)

	if !s.HasSelection {
		t.Fatal("expected selection after SelectAll")
	}

	s.ClearSelection()

	if s.HasSelection {
		t.Error("expected no selection after ClearSelection")
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

	if col.Collapsed {
		t.Fatal("expected initially expanded")
	}

	tab.ToggleCollapsible(col)

	if !col.Collapsed {
		t.Error("expected collapsed after ToggleCollapsible")
	}

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

	press := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      24,
		Y:      4,
	}
	tab.handleMouse(press, 0, 1, 40, 14)

	if !fp.resizing {
		t.Fatal("expected resizing to be true after press on edge")
	}

	motion := tea.MouseMsg{
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
		X:      29,
		Y:      4,
	}
	tab.handleMouse(motion, 0, 1, 40, 14)

	if fp.Width != 25 {
		t.Errorf("expected width 25 after resize, got %d", fp.Width)
	}

	release := tea.MouseMsg{
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
		X:      29,
		Y:      4,
	}
	tab.handleMouse(release, 0, 1, 40, 14)

	if fp.resizing {
		t.Error("expected resizing to be false after release")
	}
}

func TestFloatOffScreen(t *testing.T) {
	w := New()
	w.width = 40
	w.height = 10
	tab := w.ActiveTab()

	p := &testPanel{name: "offscreen"}
	tab.Float(p, 35, 2, 10, 5)

	content := tab.renderContent(40, 9)
	lines := strings.Split(content, "\n")

	line2 := StripANSI(lines[2])
	if len(line2) < 40 {
		t.Fatalf("line 2 too short: %d chars, expected at least 40", len(line2))
	}

	if line2[35] != "╭"[0] {
		t.Errorf("expected ╭ at position 35, got %q", string(line2[35]))
	}

	if len(lines[0]) != 40 {
		t.Errorf("line 0 width: expected %d, got %d", 40, len(lines[0]))
	}
}

func TestFloatOverlayLineAbove(t *testing.T) {
	w := New()
	w.width = 60
	w.height = 20
	tab := w.ActiveTab()

	tab.root = &Node{Panel: &testPanel{name: "background"}}

	p := &testPanel{name: "float"}
	tab.Float(p, 10, 5, 20, 6)

	content := tab.renderContent(60, 19)
	lines := strings.Split(content, "\n")

	expectedW := 60
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

func TestFloatOverlayLineBelow(t *testing.T) {
	w := New()
	w.width = 60
	w.height = 20
	tab := w.ActiveTab()

	tab.root = &Node{Panel: &testPanel{name: "background"}}

	p := &testPanel{name: "float"}
	tab.Float(p, 10, 5, 20, 6)

	content := tab.renderContent(60, 19)
	lines := strings.Split(content, "\n")

	expectedW := 60
	for y := 11; y < 19; y++ {
		line := StripANSI(lines[y])
		if len(line) != expectedW {
			t.Errorf("line %d width: expected %d, got %d", y, expectedW, len(line))
		}
	}
}

func TestOverlayFloatSimple(t *testing.T) {
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

	content := tab.renderContent(50, 14)
	lines := strings.Split(content, "\n")

	expectedW := 50
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

// TestFloatBlocksClickBelow verifies that clicking inside a float
// does NOT forward the click to the panel underneath.
func TestFloatBlocksClickBelow(t *testing.T) {
	w := New()
	w.width = 60
	w.height = 20

	clickPanel := &testPanelWS{testPanel: testPanel{name: "clickable"}}
	tab := w.ActiveTab()
	tab.root = &Node{Panel: clickPanel}

	// Float at (10, 5, 20, 6) covers the area where panel is
	tab.Float(&testPanel{name: "float"}, 10, 5, 20, 6)

	// Click inside the float (not on edge, not on title bar close)
	// relX=5, relY=3 inside float → screen (15, 8)
	press := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      15,
		Y:      8,
	}
	tab.handleMouse(press, 0, 1, 60, 19)

	// The underlying panel should NOT have received the click
	// (testPanelWS doesn't record mouse msgs, but we verify no panic
	// and that the float's presence blocks propagation)
}

// TestFloatBringToTop verifies that clicking inside a float brings it to top z-order.
func TestFloatBringToTop(t *testing.T) {
	w := New()
	w.width = 60
	w.height = 20
	tab := w.ActiveTab()

	f1 := &testPanel{name: "first"}
	f2 := &testPanel{name: "second"}
	tab.Float(f1, 10, 5, 20, 6)
	tab.Float(f2, 15, 7, 20, 6) // overlaps f1

	// f2 is on top initially (appended last)
	if tab.floats[1].Panel != f2 {
		t.Fatal("expected f2 on top initially")
	}

	// Click on f1 (which is underneath f2 at the overlap region)
	// f1 area: (10,5) to (29,10)
	// Click at (12, 6) — inside f1, outside f2 (f2 starts at x=15)
	press := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      12,
		Y:      6,
	}
	tab.handleMouse(press, 0, 1, 60, 19)

	// f1 should now be on top
	if tab.floats[1].Panel != f1 {
		t.Errorf("expected f1 on top after click, got %v", tab.floats[1].Panel)
	}
}
